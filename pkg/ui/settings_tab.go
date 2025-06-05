package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"frp-cli-ui/internal/installer"
	"frp-cli-ui/internal/service"
)

// settingsTickMsg 设置标签页时钟消息类型
type settingsTickMsg time.Time

// installStatusMsg 安装状态消息
type installStatusMsg struct {
	status *installer.InstallStatus
	err    error
}

// installProgressMsg 安装进度消息
type installProgressMsg struct {
	message string
	done    bool
	err     error
}

// serviceStatusMsg 服务状态消息
type serviceStatusMsg struct {
	serverStatus string
	clientStatus string
}

// logUpdateMsg 日志更新消息
type logUpdateMsg struct {
	serverLogs []string
	clientLogs []string
}

// StatusUpdateCallback 状态更新回调函数类型
type StatusUpdateCallback func(serverStatus, clientStatus string)

// SettingsTab 设置标签页 - 简化版本
type SettingsTab struct {
	BaseTab
	installer       *installer.Installer
	manager         *service.Manager
	installStatus   *installer.InstallStatus
	isInstalling    bool
	installProgress string
	serverStatus    string
	clientStatus    string
	statusCallback  StatusUpdateCallback
	serverLogs      []string
	clientLogs      []string
	maxLogLines     int
	// 简化状态检查
	lastServerRunning bool
	lastClientRunning bool
}

// NewSettingsTab 创建设置标签页 - 简化版本
func NewSettingsTab() *SettingsTab {
	baseTab := NewBaseTab("设置")
	baseTab.focusable = true

	st := &SettingsTab{
		BaseTab:           baseTab,
		installer:         installer.NewInstaller(""),
		manager:           service.NewManager(),
		serverStatus:      "已停止",
		clientStatus:      "未连接",
		serverLogs:        []string{},
		clientLogs:        []string{},
		maxLogLines:       20,
		lastServerRunning: false,
		lastClientRunning: false,
	}

	// 添加初始的调试日志来测试日志系统
	st.serverLogs = append(st.serverLogs, "[15:04:05] [INFO] 日志系统已初始化")
	st.clientLogs = append(st.clientLogs, "[15:04:05] [INFO] 等待客户端启动...")

	return st
}

// SetStatusCallback 设置状态更新回调
func (st *SettingsTab) SetStatusCallback(callback StatusUpdateCallback) {
	st.statusCallback = callback
}

// Init 初始化 - 简化日志系统
func (st *SettingsTab) Init() tea.Cmd {
	// 同步检查安装状态
	status, err := st.installer.CheckInstallation()
	if err == nil {
		st.installStatus = status
	} else {
		st.installProgress = fmt.Sprintf("检查安装状态失败: %v", err)
	}

	return tea.Batch(
		st.checkServiceStatus(), // 检查服务状态
		// 延迟500ms后再开始自动刷新
		tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
			return settingsTickMsg(t)
		}),
	)
}

// startAutoRefresh 启动自动刷新
func (st *SettingsTab) startAutoRefresh() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return settingsTickMsg(t)
	})
}

// Update 更新状态 - 清理版本
func (st *SettingsTab) Update(msg tea.Msg) (Tab, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		st.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		if st.focused {
			switch msg.String() {
			case "i":
				// 安装 FRP
				if st.installStatus != nil && !st.installStatus.IsInstalled && !st.isInstalling {
					return st, st.installFRP()
				}
			case "u":
				// 更新 FRP
				if st.installStatus != nil && st.installStatus.IsInstalled && st.installStatus.NeedsUpdate && !st.isInstalling {
					return st, st.updateFRP()
				}
			case "ctrl+u":
				// 卸载 FRP
				if st.installStatus != nil && st.installStatus.IsInstalled && !st.isInstalling {
					return st, st.uninstallFRP()
				}
			case "s":
				// 启动服务端 - 简化条件，优先检查服务状态
				if st.serverStatus == "已停止" {
					return st, tea.Batch(st.startServer(), st.updateLogs())
				}
			case "ctrl+s":
				// 停止服务端 - 不管是否是自己启动的都尝试停止
				if st.serverStatus == "运行中" {
					return st, st.stopServer()
				}
			case "c":
				// 启动客户端 - 简化条件，优先检查服务状态
				if st.clientStatus == "未连接" {
					return st, tea.Batch(st.startClient(), st.updateLogs())
				}
			case "ctrl+x":
				// 停止客户端 - 不管是否是自己启动的都尝试停止
				if st.clientStatus == "已连接" || st.clientStatus == "连接中" {
					return st, st.stopClient()
				}
			case "r":
				// 手动刷新安装状态
				return st, st.refreshInstallStatus()
			}
		}

	case settingsTickMsg:
		// 自动刷新状态和日志
		cmds = append(cmds,
			st.checkServiceStatus(),
			st.updateLogs(),       // 更新日志
			st.startAutoRefresh(), // 继续下一次自动刷新
		)

	case installStatusMsg:
		st.isInstalling = false // 检查完成
		st.installStatus = msg.status
		if msg.err != nil {
			st.installProgress = fmt.Sprintf("检查安装状态失败: %v", msg.err)
		} else {
			st.installProgress = "" // 清除之前的错误信息
		}

	case installProgressMsg:
		if msg.done {
			st.isInstalling = false
			if msg.err != nil {
				st.installProgress = fmt.Sprintf("操作失败: %v", msg.err)
				// 如果是启动失败，立即检查服务状态
				if strings.Contains(msg.message, "启动") {
					cmds = append(cmds, st.checkServiceStatus())
				}
			} else {
				st.installProgress = msg.message
				// 安装完成后同步检查状态
				status, err := st.installer.CheckInstallation()
				if err == nil {
					st.installStatus = status
				}
			}
		} else {
			st.installProgress = msg.message
		}

	case serviceStatusMsg:
		st.serverStatus = msg.serverStatus
		st.clientStatus = msg.clientStatus
		// 通知主界面更新状态
		if st.statusCallback != nil {
			st.statusCallback(st.serverStatus, st.clientStatus)
		}
		// 当服务状态改变时，立即更新一次日志
		cmds = append(cmds, st.updateLogs())

	case logUpdateMsg:
		st.serverLogs = msg.serverLogs
		st.clientLogs = msg.clientLogs

	case dashboardTickMsg:
		// 处理来自主仪表板的时钟消息
		if st.focused {
			cmds = append(cmds, st.checkServiceStatus())
		}
	}

	return st, tea.Batch(cmds...)
}

// View 渲染视图
func (st *SettingsTab) View(width int, height int) string {
	contentWidth := width - 12
	if contentWidth < 40 {
		contentWidth = 40
	}

	// 计算左右分屏的宽度，确保总宽度匹配
	leftWidth := (contentWidth - 4) / 2
	rightWidth := contentWidth - leftWidth - 4

	// 左侧内容样式
	leftStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(leftWidth)

	// 右侧日志样式
	rightStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(rightWidth)

	// 构建左侧内容
	leftContent := st.renderLeftContent()

	// 构建右侧日志内容，传递实际内容宽度
	rightContent := st.renderRightLogs(rightWidth - 2) // 减去padding

	// 横向组合内容
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftStyle.Render(leftContent),
		rightStyle.Render(rightContent),
	)
}

// renderLeftContent 渲染左侧内容
func (st *SettingsTab) renderLeftContent() string {
	var content string

	// FRP 安装状态部分
	content += st.renderFRPStatus()
	content += "\n\n"

	// FRP 服务控制部分
	content += st.renderServiceControl()
	content += "\n\n"

	// 应用配置部分
	content += st.renderAppConfig()
	content += "\n\n"

	// 操作提示部分（放在左侧内容底部）
	content += st.renderHorizontalHelp()

	return content
}

// renderRightLogs 渲染右侧日志内容 - 使用简单emoji避免宽度问题
func (st *SettingsTab) renderRightLogs(width int) string {
	var content string

	// 标题
	content += lipgloss.NewStyle().Bold(true).Render("📋 实时日志") + "\n\n"

	// 服务端日志区域
	content += lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render("🎯 服务端日志:") + "\n" // 使用🎯替代🖥️
	if len(st.serverLogs) == 0 {
		content += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("暂无日志 (状态: "+st.serverStatus+")") + "\n"
	} else {
		// 显示最新的日志
		for _, log := range st.serverLogs {
			// 根据日志级别设置颜色
			logColor := "250"
			if strings.Contains(log, "[ERROR]") {
				logColor = "196" // 红色
			} else if strings.Contains(log, "[WARN]") {
				logColor = "226" // 黄色
			} else if strings.Contains(log, "[INFO]") {
				logColor = "46" // 绿色
			} else if strings.Contains(log, "[DEBUG]") {
				logColor = "240" // 暗灰色
			}
			content += lipgloss.NewStyle().Foreground(lipgloss.Color(logColor)).Render("• "+log) + "\n"
		}
	}

	// 添加空行撑满上半部分
	for i := 0; i < 3; i++ {
		content += "\n"
	}

	// 分割线，使用实际宽度
	separator := strings.Repeat("─", width)
	content += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(separator) + "\n\n"

	// 客户端日志区域
	content += lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Render("💻 客户端日志:") + "\n"
	if len(st.clientLogs) == 0 {
		content += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("暂无日志 (状态: "+st.clientStatus+")") + "\n"
	} else {
		// 显示最新的日志
		for _, log := range st.clientLogs {
			// 根据日志级别设置颜色
			logColor := "250"
			if strings.Contains(log, "[ERROR]") {
				logColor = "196" // 红色
			} else if strings.Contains(log, "[WARN]") {
				logColor = "226" // 黄色
			} else if strings.Contains(log, "[INFO]") {
				logColor = "81" // 蓝色
			} else if strings.Contains(log, "[DEBUG]") {
				logColor = "240" // 暗灰色
			}
			content += lipgloss.NewStyle().Foreground(lipgloss.Color(logColor)).Render("• "+log) + "\n"
		}
	}

	// 添加空行撑满下半部分
	for i := 0; i < 3; i++ {
		content += "\n"
	}

	return content
}

// renderFRPStatus 渲染FRP安装状态 - 使用简单emoji避免宽度问题
func (st *SettingsTab) renderFRPStatus() string {
	statusStyle := lipgloss.NewStyle().Bold(true)

	var status string
	status += statusStyle.Render("🔧 FRP 安装状态") + "\n\n"

	if st.installStatus == nil {
		status += "正在检查安装状态..."
		return status
	}

	if st.installStatus.IsInstalled {
		status += fmt.Sprintf("✅ 已安装 (版本: %s)\n", st.installStatus.Version)
		status += fmt.Sprintf("📁 安装目录: %s\n", st.installStatus.InstallDir)
		status += fmt.Sprintf("🎯 服务端: %s\n", st.installStatus.FrpsPath) // 使用🎯替代🖥️避免宽度问题
		status += fmt.Sprintf("💻 客户端: %s\n", st.installStatus.FrpcPath)

		if st.installStatus.NeedsUpdate {
			status += fmt.Sprintf("🔄 有新版本可用: %s\n", st.installStatus.LatestVersion)
		} else {
			status += "✨ 已是最新版本\n"
		}
	} else {
		status += "❌ 未安装\n"
		status += fmt.Sprintf("📁 将安装到: %s\n", st.installer.GetInstallDir())
		status += fmt.Sprintf("📦 最新版本: %s\n", st.installer.GetVersion())
	}

	// 显示安装进度或状态
	if st.isInstalling {
		status += "\n🔄 " + st.installProgress
	} else if st.installProgress != "" {
		status += "\n" + st.installProgress
	}

	return status
}

// renderServiceControl 渲染服务控制部分 - 使用简单emoji避免宽度问题
func (st *SettingsTab) renderServiceControl() string {
	controlStyle := lipgloss.NewStyle().Bold(true)

	var control string
	control += controlStyle.Render("🚀 FRP 服务控制") + "\n\n"

	// 服务端状态
	serverStatusColor := "240"
	if st.serverStatus == "运行中" {
		serverStatusColor = "46" // 绿色
	} else if st.serverStatus == "启动中" {
		serverStatusColor = "226" // 黄色
	}
	serverStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(serverStatusColor))
	control += fmt.Sprintf("🎯 服务端状态: %s\n", serverStyle.Render(st.serverStatus)) // 使用🎯替代🖥️

	// 客户端状态
	clientStatusColor := "240"
	if st.clientStatus == "已连接" {
		clientStatusColor = "46" // 绿色
	} else if st.clientStatus == "连接中" {
		clientStatusColor = "226" // 黄色
	}
	clientStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(clientStatusColor))
	control += fmt.Sprintf("💻 客户端状态: %s\n", clientStyle.Render(st.clientStatus))

	control += "\n📂 配置文件:\n"
	control += "• 服务端: examples/frps.yaml\n"
	control += "• 客户端: examples/frpc.yaml\n"

	return control
}

// renderAppConfig 渲染应用配置信息
func (st *SettingsTab) renderAppConfig() string {
	configStyle := lipgloss.NewStyle().Bold(true)

	config := configStyle.Render("🔧 应用配置") + "\n\n"
	config += "🎨 主题颜色: 紫色 (#7D56F4)\n"
	config += "🔄 自动刷新: 启用 (2秒)\n"
	config += "🌐 服务端地址: 127.0.0.1:7500\n"
	config += "👤 管理员用户: admin\n"
	config += "📂 配置文件路径: examples/\n"

	return config
}

// renderHorizontalHelp 渲染横向操作提示 - 去掉边框，避免闪烁
func (st *SettingsTab) renderHorizontalHelp() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 1)

	var helpItems []string

	// 根据状态动态显示可用操作
	if st.installStatus == nil {
		helpItems = append(helpItems, "r: 刷新状态")
	} else if !st.installStatus.IsInstalled {
		helpItems = append(helpItems, "i: 安装FRP", "r: 刷新状态")
	} else {
		if st.installStatus.NeedsUpdate {
			helpItems = append(helpItems, "u: 更新FRP")
		}
		helpItems = append(helpItems, "Ctrl+U: 卸载FRP", "r: 刷新状态")

		// 服务控制操作
		if st.serverStatus == "已停止" {
			helpItems = append(helpItems, "s: 启动服务端")
		} else if st.serverStatus == "运行中" {
			helpItems = append(helpItems, "Ctrl+S: 停止服务端")
		}

		if st.clientStatus == "未连接" {
			helpItems = append(helpItems, "c: 启动客户端")
		} else if st.clientStatus == "已连接" || st.clientStatus == "连接中" {
			helpItems = append(helpItems, "Ctrl+X: 停止客户端")
		}
	}

	// 添加自动刷新提示
	helpItems = append(helpItems, "⚡ 自动刷新: 2秒")

	return helpStyle.Render("💡 " + strings.Join(helpItems, " • "))
}

// checkServiceStatus 检查服务状态 - 优化避免频繁切换
func (st *SettingsTab) checkServiceStatus() tea.Cmd {
	return func() tea.Msg {
		var serverStatus, clientStatus string

		// 检查服务端状态
		serverProcessStatus := st.manager.GetServerStatus()
		currentServerRunning := serverProcessStatus.IsRunning

		// 只有状态真正改变时才更新状态
		if currentServerRunning != st.lastServerRunning {
			st.lastServerRunning = currentServerRunning
			if currentServerRunning {
				serverStatus = "运行中"
			} else {
				serverStatus = "已停止"
			}
		} else {
			serverStatus = st.serverStatus // 保持当前状态
		}

		// 检查客户端状态
		clientProcessStatus := st.manager.GetClientStatus()
		currentClientRunning := clientProcessStatus.IsRunning

		// 客户端状态逻辑：如果进程不存在且当前状态是"连接中"，则改为"未连接"
		if currentClientRunning != st.lastClientRunning {
			st.lastClientRunning = currentClientRunning
			if currentClientRunning {
				clientStatus = "已连接"
			} else {
				clientStatus = "未连接"
			}
		} else {
			// 特殊处理：如果客户端进程已经不存在，但状态仍是"连接中"，则更新为"未连接"
			if !currentClientRunning && st.clientStatus == "连接中" {
				clientStatus = "未连接"
			} else {
				clientStatus = st.clientStatus // 保持当前状态
			}
		}

		// 只有状态真正改变时才发送更新消息
		if serverStatus != st.serverStatus || clientStatus != st.clientStatus {
			return serviceStatusMsg{
				serverStatus: serverStatus,
				clientStatus: clientStatus,
			}
		}

		// 状态未改变，返回nil避免不必要的重绘
		return nil
	}
}

// startServer 启动服务端
func (st *SettingsTab) startServer() tea.Cmd {
	return func() tea.Msg {
		err := st.manager.StartServer("examples/frps.yaml")
		if err != nil {
			return installProgressMsg{
				message: fmt.Sprintf("启动服务端失败: %v", err),
				done:    true,
				err:     err,
			}
		}
		// 先更新状态
		return serviceStatusMsg{
			serverStatus: "启动中",
			clientStatus: st.clientStatus,
		}
	}
}

// stopServer 停止服务端
func (st *SettingsTab) stopServer() tea.Cmd {
	return func() tea.Msg {
		err := st.manager.StopServer()
		if err != nil {
			return installProgressMsg{
				message: fmt.Sprintf("停止服务端失败: %v", err),
				done:    true,
				err:     err,
			}
		}
		// 先更新状态
		return serviceStatusMsg{
			serverStatus: "已停止",
			clientStatus: st.clientStatus,
		}
	}
}

// startClient 启动客户端
func (st *SettingsTab) startClient() tea.Cmd {
	return func() tea.Msg {
		err := st.manager.StartClient("examples/frpc.yaml")
		if err != nil {
			return installProgressMsg{
				message: fmt.Sprintf("启动客户端失败: %v", err),
				done:    true,
				err:     err,
			}
		}
		// 先更新状态为连接中
		return serviceStatusMsg{
			serverStatus: st.serverStatus,
			clientStatus: "连接中",
		}
	}
}

// stopClient 停止客户端
func (st *SettingsTab) stopClient() tea.Cmd {
	return func() tea.Msg {
		err := st.manager.StopClient()
		if err != nil {
			return installProgressMsg{
				message: fmt.Sprintf("停止客户端失败: %v", err),
				done:    true,
				err:     err,
			}
		}
		// 先更新状态
		return serviceStatusMsg{
			serverStatus: st.serverStatus,
			clientStatus: "未连接",
		}
	}
}

// installFRP 安装FRP
func (st *SettingsTab) installFRP() tea.Cmd {
	st.isInstalling = true
	st.installProgress = "正在下载 FRP..."

	return func() tea.Msg {
		err := st.installer.InstallFRP()
		if err != nil {
			return installProgressMsg{
				message: "",
				done:    true,
				err:     err,
			}
		}
		return installProgressMsg{
			message: "✅ FRP 安装成功！",
			done:    true,
			err:     nil,
		}
	}
}

// updateFRP 更新FRP
func (st *SettingsTab) updateFRP() tea.Cmd {
	st.isInstalling = true
	st.installProgress = "正在更新 FRP..."

	return func() tea.Msg {
		err := st.installer.UpdateFRP()
		if err != nil {
			return installProgressMsg{
				message: "",
				done:    true,
				err:     err,
			}
		}
		return installProgressMsg{
			message: "✅ FRP 更新成功！",
			done:    true,
			err:     nil,
		}
	}
}

// uninstallFRP 卸载FRP
func (st *SettingsTab) uninstallFRP() tea.Cmd {
	st.isInstalling = true
	st.installProgress = "正在卸载 FRP..."

	return func() tea.Msg {
		err := st.installer.Uninstall()
		if err != nil {
			return installProgressMsg{
				message: "",
				done:    true,
				err:     err,
			}
		}
		return installProgressMsg{
			message: "✅ FRP 卸载成功！",
			done:    true,
			err:     nil,
		}
	}
}

// refreshInstallStatus 手动刷新安装状态
func (st *SettingsTab) refreshInstallStatus() tea.Cmd {
	return func() tea.Msg {
		status, err := st.installer.CheckInstallation()
		if err != nil {
			return installStatusMsg{
				status: nil,
				err:    err,
			}
		}
		return installStatusMsg{
			status: status,
			err:    nil,
		}
	}
}

// updateLogs 更新日志 - 从manager日志通道收集
func (st *SettingsTab) updateLogs() tea.Cmd {
	return func() tea.Msg {
		// 从service manager获取日志通道
		logChan := st.manager.GetLogChannel()

		var newServerLogs, newClientLogs []string
		hasNewLogs := false

		// 非阻塞读取所有可用的新日志
		for {
			select {
			case logMsg := <-logChan:
				hasNewLogs = true
				// 格式化日志消息，包含日志级别信息
				formattedLog := fmt.Sprintf("[%s] [%s] %s",
					logMsg.Timestamp.Format("15:04:05"),
					logMsg.Level,
					logMsg.Message)

				// 根据来源分类
				if logMsg.Source == "server" {
					newServerLogs = append(newServerLogs, formattedLog)
				} else if logMsg.Source == "client" {
					newClientLogs = append(newClientLogs, formattedLog)
				}
			default:
				// 没有更多日志时退出
				goto done
			}
		}

	done:
		// 只有当有新日志时才发送更新消息
		if !hasNewLogs {
			return nil
		}

		// 合并新日志到现有日志
		allServerLogs := append(st.serverLogs, newServerLogs...)
		allClientLogs := append(st.clientLogs, newClientLogs...)

		// 限制日志行数，保留最新的日志
		if len(allServerLogs) > st.maxLogLines {
			allServerLogs = allServerLogs[len(allServerLogs)-st.maxLogLines:]
		}
		if len(allClientLogs) > st.maxLogLines {
			allClientLogs = allClientLogs[len(allClientLogs)-st.maxLogLines:]
		}

		return logUpdateMsg{
			serverLogs: allServerLogs,
			clientLogs: allClientLogs,
		}
	}
}
