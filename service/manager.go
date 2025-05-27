package service

import (
	"bufio"
	"context"
	"fmt"
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

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 创建上下文用于取消
	ctx, cancel := context.WithCancel(context.Background())
	m.serverCancel = cancel

	// 查找 frps 可执行文件
	frpsPath, err := m.findFRPExecutable("frps")
	if err != nil {
		return fmt.Errorf("找不到 frps 可执行文件: %w", err)
	}

	// 创建命令
	m.serverCmd = exec.CommandContext(ctx, frpsPath, "-c", configPath)

	// 设置输出管道
	stdout, err := m.serverCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建输出管道失败: %w", err)
	}

	stderr, err := m.serverCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建错误管道失败: %w", err)
	}

	// 启动进程
	if err := m.serverCmd.Start(); err != nil {
		return fmt.Errorf("启动 FRP 服务端失败: %w", err)
	}

	// 启动日志收集
	go m.collectLogs(stdout, "server", "INFO")
	go m.collectLogs(stderr, "server", "ERROR")

	// 监控进程状态
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

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 创建上下文用于取消
	ctx, cancel := context.WithCancel(context.Background())
	m.clientCancel = cancel

	// 查找 frpc 可执行文件
	frpcPath, err := m.findFRPExecutable("frpc")
	if err != nil {
		return fmt.Errorf("找不到 frpc 可执行文件: %w", err)
	}

	// 创建命令
	m.clientCmd = exec.CommandContext(ctx, frpcPath, "-c", configPath)

	// 设置输出管道
	stdout, err := m.clientCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建输出管道失败: %w", err)
	}

	stderr, err := m.clientCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建错误管道失败: %w", err)
	}

	// 启动进程
	if err := m.clientCmd.Start(); err != nil {
		return fmt.Errorf("启动 FRP 客户端失败: %w", err)
	}

	// 启动日志收集
	go m.collectLogs(stdout, "client", "INFO")
	go m.collectLogs(stderr, "client", "ERROR")

	// 监控进程状态
	go m.monitorProcess(m.clientCmd, "client")

	m.logChan <- LogMessage{
		Timestamp: time.Now(),
		Level:     "INFO",
		Message:   fmt.Sprintf("FRP 客户端启动成功 (PID: %d)", m.clientCmd.Process.Pid),
		Source:    "client",
	}

	return nil
}

// StopServer 停止 FRP 服务端
func (m *Manager) StopServer() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.serverCmd == nil || m.serverCmd.Process == nil {
		return fmt.Errorf("FRP 服务端未运行")
	}

	// 取消上下文
	if m.serverCancel != nil {
		m.serverCancel()
	}

	// 优雅关闭
	if err := m.serverCmd.Process.Signal(syscall.SIGTERM); err != nil {
		// 如果优雅关闭失败，强制杀死进程
		if killErr := m.serverCmd.Process.Kill(); killErr != nil {
			return fmt.Errorf("强制停止进程失败: %w", killErr)
		}
	}

	// 等待进程结束
	go func() {
		m.serverCmd.Wait()
		m.mu.Lock()
		m.serverCmd = nil
		m.mu.Unlock()
	}()

	m.logChan <- LogMessage{
		Timestamp: time.Now(),
		Level:     "INFO",
		Message:   "FRP 服务端已停止",
		Source:    "server",
	}

	return nil
}

// StopClient 停止 FRP 客户端
func (m *Manager) StopClient() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.clientCmd == nil || m.clientCmd.Process == nil {
		return fmt.Errorf("FRP 客户端未运行")
	}

	// 取消上下文
	if m.clientCancel != nil {
		m.clientCancel()
	}

	// 优雅关闭
	if err := m.clientCmd.Process.Signal(syscall.SIGTERM); err != nil {
		// 如果优雅关闭失败，强制杀死进程
		if killErr := m.clientCmd.Process.Kill(); killErr != nil {
			return fmt.Errorf("强制停止进程失败: %w", killErr)
		}
	}

	// 等待进程结束
	go func() {
		m.clientCmd.Wait()
		m.mu.Lock()
		m.clientCmd = nil
		m.mu.Unlock()
	}()

	m.logChan <- LogMessage{
		Timestamp: time.Now(),
		Level:     "INFO",
		Message:   "FRP 客户端已停止",
		Source:    "client",
	}

	return nil
}

// GetServerStatus 获取服务端状态
func (m *Manager) GetServerStatus() ProcessStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 首先检查自己管理的进程
	if m.serverCmd != nil && m.serverCmd.Process != nil {
		return ProcessStatus{
			IsRunning: true,
			PID:       m.serverCmd.Process.Pid,
			StartTime: time.Now(), // 这里应该记录实际启动时间
		}
	}

	// 检查系统中是否有运行的 frps 进程
	if pid := m.findFRPProcess("frps"); pid > 0 {
		return ProcessStatus{
			IsRunning: true,
			PID:       pid,
			StartTime: time.Now(),
		}
	}

	return ProcessStatus{IsRunning: false}
}

// GetClientStatus 获取客户端状态
func (m *Manager) GetClientStatus() ProcessStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 首先检查自己管理的进程
	if m.clientCmd != nil && m.clientCmd.Process != nil {
		return ProcessStatus{
			IsRunning: true,
			PID:       m.clientCmd.Process.Pid,
			StartTime: time.Now(), // 这里应该记录实际启动时间
		}
	}

	// 检查系统中是否有运行的 frpc 进程
	if pid := m.findFRPProcess("frpc"); pid > 0 {
		return ProcessStatus{
			IsRunning: true,
			PID:       pid,
			StartTime: time.Now(),
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
	installer := NewInstaller("")
	status, err := installer.CheckInstallation()
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
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			m.logChan <- LogMessage{
				Timestamp: time.Now(),
				Level:     level,
				Message:   line,
				Source:    source,
			}
		}
	}
}

// monitorProcess 监控进程状态
func (m *Manager) monitorProcess(cmd *exec.Cmd, source string) {
	err := cmd.Wait()

	m.mu.Lock()
	defer m.mu.Unlock()

	if err != nil {
		m.logChan <- LogMessage{
			Timestamp: time.Now(),
			Level:     "ERROR",
			Message:   fmt.Sprintf("进程异常退出: %v", err),
			Source:    source,
		}
	} else {
		m.logChan <- LogMessage{
			Timestamp: time.Now(),
			Level:     "INFO",
			Message:   "进程正常退出",
			Source:    source,
		}
	}

	// 清理命令引用
	if source == "server" {
		m.serverCmd = nil
	} else {
		m.clientCmd = nil
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
		// 解析 pgrep 输出
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
