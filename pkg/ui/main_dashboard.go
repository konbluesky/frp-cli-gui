package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"

	"frp-cli-ui/internal/service"
	constants "frp-cli-ui/pkg/config"
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
	lastProxyUpdate time.Time // 记录上次代理状态更新时间
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

	// 创建设置标签页并设置状态回调，共享同一个Manager实例
	settingsTab := NewSettingsTab()
	settingsTab.SetManager(manager) // 设置共享的Manager实例
	tabRegistry.Register(settingsTab)

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

	// 立即发送一次时钟消息，触发初始状态检查
	cmds = append(cmds, func() tea.Msg {
		return dashboardTickMsg(time.Now())
	})

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

		// 更新服务状态 - 主要依赖API检测，避免Manager进程检测的干扰
		previousServerStatus := m.statusInfo.ServerStatus
		previousClientStatus := m.statusInfo.ClientStatus

		// 检查服务器状态（仅依赖API可达性，避免进程检测的不稳定性）
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

		// 客户端状态直接从设置模块同步（避免在主仪表盘中进行复杂检测）
		// 这样状态更新来源唯一，避免冲突

		// 检测状态变化，如果有变化则立即更新代理状态
		statusChanged := (previousServerStatus != m.statusInfo.ServerStatus) ||
			(previousClientStatus != m.statusInfo.ClientStatus)

		// 更新代理状态和流量信息 - 每3秒更新一次，或状态变化时立即更新
		currentTime := time.Time(msg)
		shouldUpdateProxy := m.lastProxyUpdate.IsZero() ||
			statusChanged ||
			currentTime.Sub(m.lastProxyUpdate) >= 3*time.Second

		// 特殊情况：如果服务刚启动，多给几次机会获取数据（因为API可能需要时间启动）
		if statusChanged && m.statusInfo.ServerStatus == "运行中" && previousServerStatus == "已停止" {
			shouldUpdateProxy = true
		}

		// 如果服务器状态是运行中但还没有代理数据，强制更新
		if m.statusInfo.ServerStatus == "运行中" && m.statusInfo.ActiveProxies == 0 &&
			currentTime.Sub(m.lastProxyUpdate) >= 1*time.Second {
			shouldUpdateProxy = true
		}

		if m.apiClient != nil && shouldUpdateProxy {
			// 无论服务器状态如何，都尝试更新代理列表
			// 如果服务器未运行，getProxyList会返回空列表，这是正确的行为
			proxies := m.getProxyList()
			m.statusInfo.ActiveProxies = len(proxies)

			// 更新DashboardTab中的代理列表
			if tab, ok := m.tabRegistry.GetTabByIndex(0).(*DashboardTab); ok {
				tab.UpdateProxyList(proxies)
			}

			// 只有在服务器运行时才获取流量信息
			if m.statusInfo.ServerStatus == "运行中" {
				if serverInfo, err := m.apiClient.GetServerInfo(); err == nil {
					totalTraffic := serverInfo.TotalTrafficIn + serverInfo.TotalTrafficOut
					m.statusInfo.TotalTraffic = service.FormatTraffic(totalTraffic)
				} else {
					// 如果获取失败，保持上一次的流量显示或显示错误状态
					if m.statusInfo.TotalTraffic == "" {
						m.statusInfo.TotalTraffic = "N/A"
					}
				}
			} else {
				// 服务器未运行时重置流量信息
				m.statusInfo.TotalTraffic = "0B"
			}

			m.lastProxyUpdate = currentTime
		} else if m.statusInfo.ServerStatus != "运行中" {
			// 如果服务器未运行，清空代理列表和流量信息
			m.statusInfo.ActiveProxies = 0
			m.statusInfo.TotalTraffic = "0B"

			// 清空DashboardTab中的代理列表
			if tab, ok := m.tabRegistry.GetTabByIndex(0).(*DashboardTab); ok {
				tab.UpdateProxyList([]ProxyStatus{})
			}

			// 重置更新时间
			m.lastProxyUpdate = time.Time{}
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

// 以下是真实的服务状态检查和代理获取方法

// checkServerStatus 检查服务器状态
func (m *MainDashboard) checkServerStatus() bool {
	if m.manager == nil {
		return false
	}

	// 检查API是否可达（更准确地反映服务器是否真正可用）
	var isAPIReachable bool
	if m.apiClient != nil {
		isAPIReachable = m.apiClient.IsServerReachable()
	}

	// 修正逻辑：只有API可达时才认为服务器真正可用
	// 这样可以检测到通过Manager启动的服务和外部启动的服务
	return isAPIReachable
}

// 注意：客户端状态检查已移除，现在完全依赖设置模块的状态回调
// 这样避免了多个地方检测状态造成的冲突和混乱

// getProxyList 获取真实的代理列表
func (m *MainDashboard) getProxyList() []ProxyStatus {
	if m.apiClient == nil {
		return []ProxyStatus{}
	}

	// 尝试从API获取代理列表
	proxies, err := m.apiClient.GetProxyList()
	if err != nil {
		// API调用失败，可能是服务器还没完全启动，返回空列表但不影响状态显示
		// 可以在这里添加调试信息（如果需要的话）
		return []ProxyStatus{}
	}

	// 转换为UI层的ProxyStatus格式
	var result []ProxyStatus
	for _, proxy := range proxies {
		status := ProxyStatus{
			Name:   proxy.Name,
			Type:   proxy.Conf.Type,
			Status: proxy.Status,
		}

		// 构建本地地址（API不提供localPort信息）
		if proxy.Conf.LocalIP != "" {
			status.LocalAddr = proxy.Conf.LocalIP
		} else {
			status.LocalAddr = "N/A"
		}

		// 构建远程端口信息
		if proxy.Conf.RemotePort > 0 {
			status.RemotePort = fmt.Sprintf("%d", proxy.Conf.RemotePort)
		} else {
			status.RemotePort = "N/A"
		}

		result = append(result, status)
	}

	return result
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
		config.Title = constants.AppName + " " + constants.AppVersion
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
