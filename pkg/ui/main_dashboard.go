package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"

	"frp-cli-ui/internal/service"
)

// dashboardTickMsg 为Dashboard特定的时钟消息类型
type dashboardTickMsg time.Time

// MainDashboard 主控制面板
type MainDashboard struct {
	layout      *AppLayout
	width       int
	height      int
	activeTab   int
	tabRegistry *TabRegistry
	manager     *service.Manager
	apiClient   *service.APIClient
	statusInfo  struct {
		ServerStatus  string
		ClientStatus  string
		ActiveProxies int
		TotalTraffic  string
		LastUpdate    time.Time
	}
	showConfirmQuit bool
	ready           bool
}

// NewMainDashboard 创建新的主控制面板
func NewMainDashboard() *MainDashboard {
	return NewMainDashboardWithSize(0, 0)
}

// NewMainDashboardWithSize 创建指定大小的主控制面板
func NewMainDashboardWithSize(width, height int) *MainDashboard {
	// 设置字符宽度计算
	runewidth.DefaultCondition.EastAsianWidth = false

	// 初始化管理器和API客户端
	manager := service.NewManager()
	apiClient := service.NewAPIClient("http://127.0.0.1:7500", "admin", "admin")

	// 创建标签页注册中心
	tabRegistry := NewTabRegistry()

	// 注册标准标签页
	tabRegistry.Register(NewDashboardTab(apiClient))
	tabRegistry.Register(NewConfigTab())

	// 创建设置标签页并设置状态回调
	settingsTab := NewSettingsTab()
	tabRegistry.Register(settingsTab)

	// 注册新的示例标签页 - 这里演示如何添加新的标签页
	tabRegistry.Register(NewStatsTab())

	dashboard := &MainDashboard{
		width:       width,
		height:      height,
		tabRegistry: tabRegistry,
		activeTab:   0,
		statusInfo: struct {
			ServerStatus  string
			ClientStatus  string
			ActiveProxies int
			TotalTraffic  string
			LastUpdate    time.Time
		}{
			ServerStatus:  "已停止",
			ClientStatus:  "未连接",
			ActiveProxies: 0,
			TotalTraffic:  "0B",
			LastUpdate:    time.Now(),
		},
		manager:   manager,
		apiClient: apiClient,
		ready:     false,
	}

	// 设置状态更新回调
	settingsTab.SetStatusCallback(func(serverStatus, clientStatus string) {
		dashboard.statusInfo.ServerStatus = serverStatus
		dashboard.statusInfo.ClientStatus = clientStatus
	})

	return dashboard
}

// Init 初始化
func (m *MainDashboard) Init() tea.Cmd {
	var cmds []tea.Cmd

	// 初始化所有标签页
	for _, tab := range m.tabRegistry.GetTabs() {
		if cmd := tab.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// 添加主仪表板的时钟
	cmds = append(cmds, tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return dashboardTickMsg(t)
	}))

	return tea.Batch(cmds...)
}

// Update 更新状态
func (m *MainDashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// 初始化或更新AppLayout
		if m.layout == nil {
			m.layout = NewAppLayout(m.width, m.height)
		} else {
			m.layout.SetSize(m.width, m.height)
		}

		// 更新所有标签页大小
		for _, tab := range m.tabRegistry.GetTabs() {
			tab.SetSize(m.width, m.height)
		}

	case tea.KeyMsg:
		// 处理确认退出对话框
		if m.showConfirmQuit {
			switch msg.String() {
			case "y", "Y", "enter":
				return m, tea.Quit
			case "n", "N", "esc":
				m.showConfirmQuit = false
				return m, nil
			}
			return m, nil
		}

		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			m.showConfirmQuit = true
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			m.activeTab = (m.activeTab + 1) % len(m.tabRegistry.GetTabs())
			// 更新焦点状态
			m.updateFocus()
			return m, tea.ClearScreen

		case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab"))):
			m.activeTab = (m.activeTab - 1 + len(m.tabRegistry.GetTabs())) % len(m.tabRegistry.GetTabs())
			// 更新焦点状态
			m.updateFocus()
			return m, tea.ClearScreen

		case key.Matches(msg, key.NewBinding(key.WithKeys("s"))):
			// 启动服务端
			if m.manager != nil {
				configPath := "examples/frps.yaml"
				_ = m.manager.StartServer(configPath)
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+s"))):
			// 停止服务端
			if m.manager != nil {
				_ = m.manager.StopServer()
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
			// 启动客户端
			if m.manager != nil {
				configPath := "examples/frpc.yaml"
				_ = m.manager.StartClient(configPath)
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+d"))):
			// 停止客户端
			if m.manager != nil {
				_ = m.manager.StopClient()
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+z"))):
			// 处理 Ctrl+Z 挂起
			return m, func() tea.Msg { return tea.Suspend() }
		}

	case tea.SuspendMsg:
		// 程序即将挂起，可以在这里做一些清理工作
		return m, nil

	case tea.ResumeMsg:
		// 程序从挂起状态恢复，重新绘制屏幕
		return m, tea.ClearScreen

	case dashboardTickMsg:
		m.statusInfo.LastUpdate = time.Time(msg)

		// 更新服务状态 - 现在主要通过SettingsTab的回调更新
		if m.manager != nil {
			// 检查服务器状态
			serverRunning := m.checkServerStatus()
			if serverRunning {
				if m.statusInfo.ServerStatus != "运行中" {
					m.statusInfo.ServerStatus = "运行中"
				}
			} else {
				if m.statusInfo.ServerStatus != "已停止" {
					m.statusInfo.ServerStatus = "已停止"
				}
			}

			// 检查客户端状态
			clientRunning := m.checkClientStatus()
			if clientRunning {
				if m.statusInfo.ClientStatus != "已连接" {
					m.statusInfo.ClientStatus = "已连接"
				}
			} else {
				if m.statusInfo.ClientStatus != "未连接" {
					m.statusInfo.ClientStatus = "未连接"
				}
			}
		}

		// 更新代理状态
		if m.apiClient != nil && m.statusInfo.ServerStatus == "运行中" {
			// 获取代理列表 - 这里模拟实现
			// 实际项目中应该调用真实的方法
			var proxies []ProxyStatus

			// 模拟获取代理列表
			proxies = m.getProxyList()
			m.statusInfo.ActiveProxies = len(proxies)

			// 更新DashboardTab中的代理列表
			if tab, ok := m.tabRegistry.GetTabByIndex(0).(*DashboardTab); ok {
				tab.UpdateProxyList(proxies)
			}

			// 获取总流量
			if serverInfo, err := m.apiClient.GetServerInfo(); err == nil {
				totalTraffic := serverInfo.TotalTrafficIn + serverInfo.TotalTrafficOut
				m.statusInfo.TotalTraffic = service.FormatTraffic(totalTraffic)
			}
		}

		cmds = append(cmds, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return dashboardTickMsg(t)
		}))
	}

	// 更新当前活动的标签页
	if m.ready && m.activeTab < len(m.tabRegistry.GetTabs()) {
		activeTab := m.tabRegistry.GetTabByIndex(m.activeTab)
		updatedTab, cmd := activeTab.Update(msg)

		// 更新注册中心中的标签页
		tabs := m.tabRegistry.GetTabs()
		tabs[m.activeTab] = updatedTab

		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// 以下是模拟方法，实际项目中应替换为真实实现

// checkServerStatus 模拟检查服务器状态
func (m *MainDashboard) checkServerStatus() bool {
	// 这里仅用于演示，返回随机状态
	// 实际应该调用 m.manager 的相关方法
	return false
}

// checkClientStatus 模拟检查客户端状态
func (m *MainDashboard) checkClientStatus() bool {
	// 这里仅用于演示，返回随机状态
	// 实际应该调用 m.manager 的相关方法
	return false
}

// getProxyList 模拟获取代理列表
func (m *MainDashboard) getProxyList() []ProxyStatus {
	// 这里仅用于演示，返回模拟数据
	// 实际应该调用 m.apiClient 的相关方法
	return []ProxyStatus{
		{
			Name:       "示例代理",
			Type:       "tcp",
			LocalAddr:  "127.0.0.1:8080",
			RemotePort: "8080",
			Status:     "在线",
		},
	}
}

// updateFocus 更新标签页焦点状态
func (m *MainDashboard) updateFocus() {
	for i, tab := range m.tabRegistry.GetTabs() {
		if tab.Focusable() {
			tab.Focus(i == m.activeTab)
		}
	}
}

// View 渲染视图
func (m *MainDashboard) View() string {
	if !m.ready || m.layout == nil {
		return "正在初始化...\n\n按 Ctrl+C 退出"
	}

	// 显示确认退出对话框
	if m.showConfirmQuit {
		dialogContent := `确认退出

您确定要退出 FRP 管理工具吗？

[Y] 是的，退出  [N] 取消

按 Y 或 Enter 确认退出，按 N 或 ESC 取消`

		return m.layout.RenderDialog(dialogContent, DefaultDialogOptions())
	}

	// 使用AppLayout渲染主界面
	m.layout.UpdateConfig(func(config *AppLayoutConfig) {
		config.Title = "FRP 内网穿透管理工具"
		config.Tabs = m.tabRegistry.GetTabTitles()
		config.ActiveTab = m.activeTab
		config.StatusText = fmt.Sprintf(
			"Server: %s | Client: %s | Active Proxies: %d | Total Traffic: %s | Last Update: %s",
			m.statusInfo.ServerStatus,
			m.statusInfo.ClientStatus,
			m.statusInfo.ActiveProxies,
			m.statusInfo.TotalTraffic,
			m.statusInfo.LastUpdate.Format(time.DateTime),
		)
		config.HelpText = "Tab: 切换标签 | q: 退出"

		// 获取当前活动标签页的内容
		if m.activeTab < len(m.tabRegistry.GetTabs()) {
			activeTab := m.tabRegistry.GetTabByIndex(m.activeTab)
			config.MainContent = activeTab.View(m.width, m.height)
		}
	})

	return m.layout.Render()
}
