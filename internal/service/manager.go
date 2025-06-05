package service

import (
	"bufio"
	"context"
	"fmt"
	"frp-cli-ui/internal/installer"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Manager FRP 进程管理器
type Manager struct {
	mu           sync.RWMutex
	serverCmd    *exec.Cmd
	clientCmd    *exec.Cmd
	serverCancel context.CancelFunc
	clientCancel context.CancelFunc
	logChan      chan LogMessage
	isRunning    bool
}

// LogMessage 日志消息
type LogMessage struct {
	Timestamp time.Time
	Level     string
	Message   string
	Source    string // "server" 或 "client"
}

// ProcessStatus 进程状态
type ProcessStatus struct {
	IsRunning bool
	PID       int
	StartTime time.Time
	CPU       float64
	Memory    uint64
}

// NewManager 创建新的进程管理器
func NewManager() *Manager {
	return &Manager{
		logChan: make(chan LogMessage, 1000),
	}
}

// StartServer 启动 FRP 服务端
func (m *Manager) StartServer(configPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.serverCmd != nil && m.serverCmd.Process != nil {
		return fmt.Errorf("FRP 服务端已在运行")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.serverCancel = cancel

	frpsPath, err := m.findFRPExecutable("frps")
	if err != nil {
		return fmt.Errorf("找不到 frps 可执行文件: %w", err)
	}

	m.serverCmd = exec.CommandContext(ctx, frpsPath, "-c", configPath)

	stdout, err := m.serverCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建输出管道失败: %w", err)
	}

	stderr, err := m.serverCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建错误管道失败: %w", err)
	}

	if err := m.serverCmd.Start(); err != nil {
		return fmt.Errorf("启动 FRP 服务端失败: %w", err)
	}

	go m.collectLogs(stdout, "server", "INFO")
	go m.collectLogs(stderr, "server", "ERROR")
	go m.monitorProcess(m.serverCmd, "server")

	m.isRunning = true
	m.logChan <- LogMessage{
		Timestamp: time.Now(),
		Level:     "INFO",
		Message:   fmt.Sprintf("FRP 服务端启动成功 (PID: %d)", m.serverCmd.Process.Pid),
		Source:    "server",
	}

	return nil
}

// StartClient 启动 FRP 客户端
func (m *Manager) StartClient(configPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.clientCmd != nil && m.clientCmd.Process != nil {
		return fmt.Errorf("FRP 客户端已在运行")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.clientCancel = cancel

	frpcPath, err := m.findFRPExecutable("frpc")
	if err != nil {
		return fmt.Errorf("找不到 frpc 可执行文件: %w", err)
	}

	m.clientCmd = exec.CommandContext(ctx, frpcPath, "-c", configPath)

	stdout, err := m.clientCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建输出管道失败: %w", err)
	}

	stderr, err := m.clientCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建错误管道失败: %w", err)
	}

	if err := m.clientCmd.Start(); err != nil {
		return fmt.Errorf("启动 FRP 客户端失败: %w", err)
	}

	go m.collectLogs(stdout, "client", "INFO")
	go m.collectLogs(stderr, "client", "ERROR")
	go m.monitorProcess(m.clientCmd, "client")

	m.logChan <- LogMessage{
		Timestamp: time.Now(),
		Level:     "INFO",
		Message:   fmt.Sprintf("FRP 客户端启动成功 (PID: %d)", m.clientCmd.Process.Pid),
		Source:    "client",
	}

	return nil
}

// StopServer 停止 FRP 服务端 - 支持停止外部启动的进程
func (m *Manager) StopServer() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var stoppedPID int

	if m.serverCmd != nil && m.serverCmd.Process != nil {
		process := m.serverCmd.Process
		stoppedPID = process.Pid

		if m.serverCancel != nil {
			m.serverCancel()
			m.serverCancel = nil
		}

		if err := process.Signal(syscall.SIGTERM); err != nil {
			if killErr := process.Kill(); killErr != nil {
				return fmt.Errorf("强制停止进程失败: %w", killErr)
			}
		}

		m.serverCmd.Wait()
		m.serverCmd = nil
		m.isRunning = false
	} else {
		if pid := m.findFRPProcess("frps"); pid > 0 {
			stoppedPID = pid
			if err := m.killProcessByPID(pid); err != nil {
				return fmt.Errorf("停止外部FRP服务端失败: %w", err)
			}
		}
	}

	if stoppedPID > 0 {
		m.logChan <- LogMessage{
			Timestamp: time.Now(),
			Level:     "INFO",
			Message:   fmt.Sprintf("FRP 服务端已停止 (PID: %d)", stoppedPID),
			Source:    "server",
		}
	}

	return nil
}

// StopClient 停止 FRP 客户端 - 支持停止外部启动的进程
func (m *Manager) StopClient() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 首先尝试停止自己管理的进程
	if m.clientCmd != nil && m.clientCmd.Process != nil {
		process := m.clientCmd.Process
		cmd := m.clientCmd

		// 取消上下文
		if m.clientCancel != nil {
			m.clientCancel()
			m.clientCancel = nil
		}

		// 优雅关闭
		if err := process.Signal(syscall.SIGTERM); err != nil {
			// 如果优雅关闭失败，强制杀死进程
			if killErr := process.Kill(); killErr != nil {
				return fmt.Errorf("强制停止进程失败: %w", killErr)
			}
		}

		// 清理引用
		m.clientCmd = nil

		// 在后台等待进程结束，但不阻塞当前操作
		go func() {
			cmd.Wait()
		}()

		m.logChan <- LogMessage{
			Timestamp: time.Now(),
			Level:     "INFO",
			Message:   "FRP 客户端已停止",
			Source:    "client",
		}

		return nil
	}

	// 如果没有自己管理的进程，尝试查找并停止外部进程
	if pid := m.findFRPProcess("frpc"); pid > 0 {
		if err := m.killProcessByPID(pid); err != nil {
			return fmt.Errorf("停止外部 FRP 客户端进程失败: %w", err)
		}

		m.logChan <- LogMessage{
			Timestamp: time.Now(),
			Level:     "INFO",
			Message:   fmt.Sprintf("外部 FRP 客户端进程已停止 (PID: %d)", pid),
			Source:    "client",
		}

		return nil
	}

	return fmt.Errorf("没有找到运行中的 FRP 客户端进程")
}

// killProcessByPID 根据PID停止进程
func (m *Manager) killProcessByPID(pid int) error {
	switch runtime.GOOS {
	case "windows":
		// Windows 使用 taskkill
		cmd := exec.Command("taskkill", "/F", "/PID", fmt.Sprintf("%d", pid))
		return cmd.Run()
	case "darwin", "linux":
		// macOS 和 Linux 使用 kill
		cmd := exec.Command("kill", "-TERM", fmt.Sprintf("%d", pid))
		if err := cmd.Run(); err != nil {
			// 如果 SIGTERM 失败，尝试 SIGKILL
			cmd = exec.Command("kill", "-KILL", fmt.Sprintf("%d", pid))
			return cmd.Run()
		}
		return nil
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// GetServerStatus 获取服务端状态 - 仅检查自己管理的进程
func (m *Manager) GetServerStatus() ProcessStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 只检查自己管理的进程，避免受外部进程干扰
	if m.serverCmd != nil && m.serverCmd.Process != nil {
		return ProcessStatus{
			IsRunning: true,
			PID:       m.serverCmd.Process.Pid,
			StartTime: time.Now(), // 这里应该记录实际启动时间
		}
	}

	return ProcessStatus{IsRunning: false}
}

// GetClientStatus 获取客户端状态 - 仅检查自己管理的进程
func (m *Manager) GetClientStatus() ProcessStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 只检查自己管理的进程，避免受外部进程干扰
	if m.clientCmd != nil && m.clientCmd.Process != nil {
		return ProcessStatus{
			IsRunning: true,
			PID:       m.clientCmd.Process.Pid,
			StartTime: time.Now(), // 这里应该记录实际启动时间
		}
	}

	return ProcessStatus{IsRunning: false}
}

// GetLogChannel 获取日志通道
func (m *Manager) GetLogChannel() <-chan LogMessage {
	return m.logChan
}

// findFRPExecutable 查找 FRP 可执行文件
func (m *Manager) findFRPExecutable(name string) (string, error) {
	// 首先尝试使用安装器查找
	inst := installer.NewInstaller("")
	status, err := inst.CheckInstallation()
	if err == nil && status.IsInstalled {
		if name == "frps" && status.FrpsPath != "" {
			return status.FrpsPath, nil
		}
		if name == "frpc" && status.FrpcPath != "" {
			return status.FrpcPath, nil
		}
	}

	// 然后在 PATH 中查找
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}

	// 在当前目录查找
	currentDir, _ := os.Getwd()
	localPath := filepath.Join(currentDir, name)
	if runtime.GOOS == "windows" {
		localPath += ".exe"
	}
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	// 在常见位置查找
	commonPaths := []string{
		"/usr/local/bin/" + name,
		"/usr/bin/" + name,
		"/opt/frp/" + name,
		"./bin/" + name,
	}

	if runtime.GOOS == "windows" {
		for i, path := range commonPaths {
			commonPaths[i] = path + ".exe"
		}
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("找不到 %s 可执行文件", name)
}

// collectLogs 收集进程日志
func (m *Manager) collectLogs(reader io.Reader, source, level string) {
	// 只有INFO级别的日志收集器在结束时发送停止消息，避免重复
	defer func() {
		if level == "INFO" {
			// 当日志收集结束时，发送一条信息
			select {
			case m.logChan <- LogMessage{
				Timestamp: time.Now(),
				Level:     "DEBUG",
				Message:   fmt.Sprintf("%s 日志收集已停止", source),
				Source:    source,
			}:
			default:
				// 如果通道已关闭，忽略错误
			}
		}
	}()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			// 尝试发送日志到通道，如果通道已关闭则退出
			select {
			case m.logChan <- LogMessage{
				Timestamp: time.Now(),
				Level:     level,
				Message:   line,
				Source:    source,
			}:
			default:
				// 通道可能已满或关闭，退出
				return
			}
		}
	}

	// 检查scanner错误
	if err := scanner.Err(); err != nil && err != io.EOF {
		select {
		case m.logChan <- LogMessage{
			Timestamp: time.Now(),
			Level:     "ERROR",
			Message:   fmt.Sprintf("日志扫描错误: %v", err),
			Source:    source,
		}:
		default:
			// 如果通道已关闭，忽略错误
		}
	}
}

// monitorProcess 监控进程状态
func (m *Manager) monitorProcess(cmd *exec.Cmd, source string) {
	err := cmd.Wait()

	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查命令是否还存在（可能已被清理）
	var shouldLog bool
	if source == "server" && m.serverCmd == cmd {
		m.serverCmd = nil
		shouldLog = true
	} else if source == "client" && m.clientCmd == cmd {
		m.clientCmd = nil
		shouldLog = true
	}

	// 只有当命令仍然有效时才记录退出信息
	if shouldLog {
		if err != nil {
			// 检查是否是被取消的上下文（正常停止）
			if strings.Contains(err.Error(), "signal: terminated") ||
				strings.Contains(err.Error(), "context canceled") {
				m.logChan <- LogMessage{
					Timestamp: time.Now(),
					Level:     "INFO",
					Message:   fmt.Sprintf("%s 进程已正常停止", source),
					Source:    source,
				}
			} else {
				m.logChan <- LogMessage{
					Timestamp: time.Now(),
					Level:     "ERROR",
					Message:   fmt.Sprintf("进程异常退出: %v", err),
					Source:    source,
				}
			}
		} else {
			m.logChan <- LogMessage{
				Timestamp: time.Now(),
				Level:     "INFO",
				Message:   fmt.Sprintf("%s 进程正常退出", source),
				Source:    source,
			}
		}
	}
}

// Restart 重启服务
func (m *Manager) Restart(service, configPath string) error {
	switch service {
	case "server":
		if err := m.StopServer(); err != nil {
			return fmt.Errorf("停止服务端失败: %w", err)
		}
		// 等待进程完全停止
		time.Sleep(2 * time.Second)
		return m.StartServer(configPath)
	case "client":
		if err := m.StopClient(); err != nil {
			return fmt.Errorf("停止客户端失败: %w", err)
		}
		// 等待进程完全停止
		time.Sleep(2 * time.Second)
		return m.StartClient(configPath)
	default:
		return fmt.Errorf("未知的服务类型: %s", service)
	}
}

// Close 关闭管理器
func (m *Manager) Close() error {
	var errs []error

	if err := m.StopServer(); err != nil {
		errs = append(errs, err)
	}

	if err := m.StopClient(); err != nil {
		errs = append(errs, err)
	}

	close(m.logChan)

	if len(errs) > 0 {
		return fmt.Errorf("关闭时发生错误: %v", errs)
	}

	return nil
}

// findFRPProcess 查找系统中运行的 FRP 进程
func (m *Manager) findFRPProcess(processName string) int {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Windows 使用 tasklist
		cmd = exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s.exe", processName), "/FO", "CSV", "/NH")
	case "darwin", "linux":
		// macOS 和 Linux 使用 pgrep
		cmd = exec.Command("pgrep", "-f", processName)
	default:
		return 0
	}

	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return 0
	}

	switch runtime.GOOS {
	case "windows":
		// 解析 Windows tasklist 输出
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, processName) {
				fields := strings.Split(line, ",")
				if len(fields) >= 2 {
					pidStr := strings.Trim(fields[1], "\"")
					if pid, err := strconv.Atoi(pidStr); err == nil {
						return pid
					}
				}
			}
		}
	case "darwin", "linux":
		// 解析 pgrep 输出，返回第一个找到的PID
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				if pid, err := strconv.Atoi(line); err == nil {
					return pid
				}
			}
		}
	}

	return 0
}
