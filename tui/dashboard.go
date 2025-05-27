package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"frp-cli-ui/service"
)

// Dashboard 主控面板模型
type Dashboard struct {
	width         int
	height        int
	activeTab     int
	tabs          []string
	table         table.Model
	statusInfo    StatusInfo
	configEditor  *ConfigEditor
	showingConfig bool
	showingLogs   bool
	logsView      *LogsView
	manager       *service.Manager
	apiClient     *service.APIClient
	logMessages   []service.LogMessage // 缓存日志消息
}

// StatusInfo FRP 状态信息
type StatusInfo struct {
	ServerStatus  string
	ClientStatus  string
	ActiveProxies int
	TotalTraffic  string
	LastUpdate    time.Time
}

// NewDashboard 创建新的主控面板
func NewDashboard() Dashboard {
	return NewDashboardWithSize(0, 0)
}

// NewDashboardWithSize 创建指定大小的主控面板
func NewDashboardWithSize(width, height int) Dashboard {
	// 初始化表格
	columns := []table.Column{
		{Title: "代理名称", Width: 15},
		{Title: "类型", Width: 8},
		{Title: "本地地址", Width: 20},
		{Title: "远程端口", Width: 10},
		{Title: "状态", Width: 8},
	}

	rows := []table.Row{
		{"web-server", "http", "127.0.0.1:8080", "80", "运行中"},
		{"ssh-tunnel", "tcp", "127.0.0.1:22", "2222", "运行中"},
		{"database", "tcp", "127.0.0.1:3306", "3306", "停止"},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	// 如果提供了宽度，设置表格宽度
	if width > 0 {
		t.SetWidth(width - 6)
	}

	// 初始化管理器和API客户端
	manager := service.NewManager()
	apiClient := service.NewAPIClient("http://127.0.0.1:7500", "admin", "admin")

	return Dashboard{
		width:  width,
		height: height,
		tabs:   []string{"仪表盘", "配置管理", "日志查看", "设置"},
		table:  t,
		statusInfo: StatusInfo{
			ServerStatus:  "运行中",
			ClientStatus:  "已连接",
			ActiveProxies: 2,
			TotalTraffic:  "1.2GB",
			LastUpdate:    time.Now(),
		},
		showingConfig: false,
		manager:       manager,
		apiClient:     apiClient,
		logMessages:   make([]service.LogMessage, 0),
	}
}

// Init 初始化
func (m Dashboard) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type tickMsg time.Time

// Update 更新状态
func (m Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// 如果正在显示配置编辑器，将消息传递给它
	if m.showingConfig && m.configEditor != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" {
				m.showingConfig = false
				return m, nil
			}
		}

		updatedEditor, editorCmd := m.configEditor.Update(msg)
		if editor, ok := updatedEditor.(ConfigEditor); ok {
			*m.configEditor = editor
		}
		return m, editorCmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.configEditor != nil {
			updatedEditor, _ := m.configEditor.Update(msg)
			if editor, ok := updatedEditor.(ConfigEditor); ok {
				*m.configEditor = editor
			}
		}
		// 更新表格宽度以适应新的窗口大小和边框
		m.table.SetWidth(m.width - 6) // 与 renderDashboard 中的设置保持一致
		// m.table.SetHeight(...) // 如果需要，也调整高度

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
			// 强制重绘界面以避免UI混乱
			return m, tea.ClearScreen

		case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab"))):
			m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
			// 强制重绘界面以避免UI混乱
			return m, tea.ClearScreen

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			// 处理选中项操作
			if m.activeTab == 1 { // 配置管理
				m.showConfigEditor()
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
			// 快捷键直接进入配置管理
			if m.activeTab == 1 {
				m.showConfigEditor()
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("s"))):
			// s 键启动服务端
			if m.manager != nil {
				// 这里应该有配置文件路径，暂时使用默认路径
				configPath := "examples/frps.yaml"
				if err := m.manager.StartServer(configPath); err != nil {
					// 可以添加错误显示逻辑
					_ = err
				}
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+s"))):
			// Ctrl+s 停止服务端
			if m.manager != nil {
				if err := m.manager.StopServer(); err != nil {
					// 可以添加错误显示逻辑
					_ = err
				}
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
			// d 键启动客户端
			if m.manager != nil {
				// 这里应该有配置文件路径，暂时使用默认路径
				configPath := "examples/frpc.yaml"
				if err := m.manager.StartClient(configPath); err != nil {
					// 可以添加错误显示逻辑
					_ = err
				}
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+d"))):
			// Ctrl+d 停止客户端
			if m.manager != nil {
				if err := m.manager.StopClient(); err != nil {
					// 可以添加错误显示逻辑
					_ = err
				}
			}
		}

	case tickMsg:
		// 更新状态信息
		m.statusInfo.LastUpdate = time.Time(msg)

		// 从管理器获取实时状态
		if m.manager != nil {
			serverStatus := m.manager.GetServerStatus()
			clientStatus := m.manager.GetClientStatus()

			// 更新服务器状态
			if serverStatus.IsRunning {
				m.statusInfo.ServerStatus = fmt.Sprintf("运行中 (PID: %d)", serverStatus.PID)
			} else {
				m.statusInfo.ServerStatus = "已停止"
			}

			// 更新客户端状态
			if clientStatus.IsRunning {
				m.statusInfo.ClientStatus = fmt.Sprintf("已连接 (PID: %d)", clientStatus.PID)
			} else {
				m.statusInfo.ClientStatus = "未连接"
			}
		}

		// 从API客户端获取代理信息
		if m.apiClient != nil && m.apiClient.IsServerReachable() {
			if proxies, err := m.apiClient.GetProxyList(); err == nil {
				m.statusInfo.ActiveProxies = len(proxies)

				// 更新表格数据
				var rows []table.Row
				for _, proxy := range proxies {
					status := "运行中"
					if proxy.Status != "online" {
						status = "停止"
					}
					rows = append(rows, table.Row{
						proxy.Name,
						proxy.Type,
						proxy.LocalAddr,
						proxy.RemoteAddr,
						status,
					})
				}
				m.table.SetRows(rows)
			}

			// 获取服务器信息更新流量统计
			if serverInfo, err := m.apiClient.GetServerInfo(); err == nil {
				totalTraffic := serverInfo.TotalTrafficIn + serverInfo.TotalTrafficOut
				m.statusInfo.TotalTraffic = service.FormatTraffic(totalTraffic)
			}
		}

		// 收集日志消息
		if m.manager != nil {
			logChan := m.manager.GetLogChannel()
			// 非阻塞地读取所有可用的日志消息
			for {
				select {
				case logMsg := <-logChan:
					m.logMessages = append(m.logMessages, logMsg)
					// 只保留最近的100条日志
					if len(m.logMessages) > 100 {
						m.logMessages = m.logMessages[1:]
					}
				default:
					// 没有更多日志，退出循环
					goto continueUpdate
				}
			}
		}

	continueUpdate:
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// showConfigEditor 显示配置编辑器
func (m *Dashboard) showConfigEditor() {
	if m.configEditor == nil {
		// 创建客户端配置编辑器作为默认
		editor := NewConfigEditor("client")
		m.configEditor = &editor
	}
	m.showingConfig = true
}

// View 渲染视图
func (m Dashboard) View() string {
	if m.width == 0 {
		return "正在加载..."
	}

	// 如果正在显示配置编辑器，直接返回编辑器视图
	if m.showingConfig && m.configEditor != nil {
		return m.configEditor.View()
	}

	// 样式定义
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width - 4). // 减去 appBorderStyle 的 padding(2) + border(2)
		Align(lipgloss.Center)

	tabStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 5)

	activeTabStyle := tabStyle.Copy().
		BorderForeground(lipgloss.Color("57")).
		Foreground(lipgloss.Color("57"))

	statusStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(m.width - 8) // 减去 appBorderStyle 的 padding(2) + border(2) + statusStyle 自身的 padding(2) + border(2)

	// 整个应用的边框样式
	appBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1)

	// 标题
	title := titleStyle.Render("FRP 内网穿透管理工具")

	// 标签页
	var tabs []string
	for i, tab := range m.tabs {
		if i == m.activeTab {
			tabs = append(tabs, activeTabStyle.Render(tab))
		} else {
			tabs = append(tabs, tabStyle.Render(tab))
		}
	}
	tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	// 状态信息
	statusContent := fmt.Sprintf(
		"服务器状态: %s | 客户端状态: %s | 活跃代理: %d | 总流量: %s | 更新时间: %s",
		m.statusInfo.ServerStatus,
		m.statusInfo.ClientStatus,
		m.statusInfo.ActiveProxies,
		m.statusInfo.TotalTraffic,
		m.statusInfo.LastUpdate.Format("15:04:05"),
	)
	statusBar := statusStyle.Render(statusContent)

	// 主内容区域
	var content string
	switch m.activeTab {
	case 0: // 仪表盘
		content = m.renderDashboard()
	case 1: // 配置管理
		content = m.renderConfigManagement()
	case 2: // 日志查看
		content = m.renderLogsView()
	case 3: // 设置
		content = "设置功能开发中..."
	}

	// 帮助信息
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Tab: 切换标签 | Enter: 选择 | c: 配置管理 | s: 启动服务端 | Ctrl+s: 停止服务端 | d: 启动客户端 | Ctrl+d: 停止客户端 | q: 退出")

	// 组合内容
	innerContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"", // 标题和标签之间的间隔
		tabsRow,
		"",
		statusBar,
		"",
		content,
	)

	// 计算内容高度，以便将帮助信息推到底部
	contentHeight := lipgloss.Height(innerContent)
	// appBorderStyle 的垂直 padding 是 1+1=2，Border 也是 1+1=2 （近似）
	// help 的高度 lipgloss.Height(help)
	// 我们需要计算 innerContent 和 help 之间的空白区域
	verticalPaddingAndBorder := 4 // 近似值，根据实际效果调整
	remainingHeight := m.height - contentHeight - lipgloss.Height(help) - verticalPaddingAndBorder
	if remainingHeight < 0 {
		remainingHeight = 0
	}

	// 最终组合，添加帮助信息和底部填充
	finalViewContent := lipgloss.JoinVertical(
		lipgloss.Left,
		innerContent,
		lipgloss.PlaceVertical(remainingHeight, lipgloss.Bottom, help),
	)

	// 应用整体边框
	return appBorderStyle.Render(finalViewContent)
}

// renderDashboard 渲染仪表盘内容
func (m Dashboard) renderDashboard() string {
	// 确保 m.table.View() 在调用前，其尺寸已根据 m.width/m.height 更新
	// table 宽度应为可用内容宽度，即 m.width - 4 (appBorderStyle padding + border)
	// 表格本身可能有自己的边框/内边距，需要进一步调整
	m.table.SetWidth(m.width - 6) // 调整为 -4 (app border) -2 (table margin)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render("代理状态"),
		"",
		m.table.View(),
	)
}

// renderConfigManagement 渲染配置管理内容
func (m Dashboard) renderConfigManagement() string {
	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(2).
		Width(m.width - 10) // 减去 appBorderStyle(4) + contentStyle 自身的 padding(4) + border(2)

	content := `配置管理功能

可用操作:
• Enter 或 c - 进入配置编辑器
• 支持服务端和客户端配置
• 内置配置模板
• 配置验证和导入导出
• 历史记录和撤销功能

按 Enter 或 c 键开始配置管理...`

	return contentStyle.Render(content)
}

// renderLogsView 渲染日志查看内容
func (m Dashboard) renderLogsView() string {
	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(m.width - 8).   // 减去 appBorderStyle(4) + contentStyle 自身的 padding(2) + border(2)
		Height(m.height - 12) // 为其他UI元素留空间

	if m.manager == nil {
		return contentStyle.Render("管理器未初始化")
	}

	if len(m.logMessages) == 0 {
		return contentStyle.Render("暂无日志信息\n\n提示：启动 FRP 服务后将显示实时日志")
	}

	// 格式化日志消息
	var logLines []string
	// 显示最近的30条日志
	startIdx := 0
	if len(m.logMessages) > 30 {
		startIdx = len(m.logMessages) - 30
	}

	for i := startIdx; i < len(m.logMessages); i++ {
		logMsg := m.logMessages[i]
		logLine := fmt.Sprintf("[%s] [%s] [%s] %s",
			logMsg.Timestamp.Format("15:04:05"),
			logMsg.Source,
			logMsg.Level,
			logMsg.Message,
		)
		logLines = append(logLines, logLine)
	}

	// 将日志行连接成字符串
	logContent := strings.Join(logLines, "\n")

	return contentStyle.Render(fmt.Sprintf("实时日志 (最近 %d 条):\n\n%s", len(logLines), logContent))
}
