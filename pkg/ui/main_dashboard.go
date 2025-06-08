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
	runewidth.DefaultCondition.EastAsianWidth = false

	manager := service.NewManager()
	apiClient := service.NewAPIClient("http://127.0.0.1:7500", "admin", "admin")

	tabRegistry := NewTabRegistry()
	tabRegistry.Register(NewDashboardTab(apiClient))
	tabRegistry.Register(NewConfigTab())

	settingsTab := NewSettingsTab()
	settingsTab.SetManager(manager)
	tabRegistry.Register(settingsTab)

	dashboard := &MainDashboard{
		tabRegistry: tabRegistry,
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
	}

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
	cmds = append(cmds,
		tea.Tick(time.Second, func(t time.Time) tea.Msg { return dashboardTickMsg(t) }),
		func() tea.Msg { return dashboardTickMsg(time.Now()) },
	)

	return tea.Batch(cmds...)
}

// Update 更新状态
func (m *MainDashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height, m.ready = msg.Width, msg.Height, true

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
			}
			return m, nil
		}

		// 检查当前标签页是否需要独占键盘输入
		shouldInterceptKeys := m.shouldInterceptKeysForCurrentTab()

		// 如果当前标签页不需要独占输入，处理全局快捷键
		if !shouldInterceptKeys {
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
					_ = m.manager.StartServer("examples/frps.yaml")
				}

			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+s"))):
				// 停止服务端
				if m.manager != nil {
					_ = m.manager.StopServer()
				}

			case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
				// 启动客户端
				if m.manager != nil {
					_ = m.manager.StartClient("examples/frpc.yaml")
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
		}

	case tea.SuspendMsg:
		// 程序即将挂起，可以在这里做一些清理工作
		return m, nil

	case tea.ResumeMsg:
		// 程序从挂起状态恢复，重新绘制屏幕
		return m, tea.ClearScreen

	case dashboardTickMsg:
		m.updateStatus(time.Time(msg))
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
	return m.manager != nil && m.apiClient != nil && m.apiClient.IsServerReachable()
}

// getProxyList 获取真实的代理列表
func (m *MainDashboard) getProxyList() []ProxyStatus {
	if m.apiClient == nil {
		return []ProxyStatus{}
	}

	proxies, err := m.apiClient.GetProxyList()
	if err != nil {
		return []ProxyStatus{}
	}

	result := make([]ProxyStatus, len(proxies))
	for i, proxy := range proxies {
		result[i] = ProxyStatus{
			Name:   proxy.Name,
			Type:   proxy.Conf.Type,
			Status: proxy.Status,
		}

		if proxy.Conf.LocalIP != "" {
			result[i].LocalAddr = proxy.Conf.LocalIP
		} else {
			result[i].LocalAddr = "N/A"
		}

		if proxy.Conf.RemotePort > 0 {
			result[i].RemotePort = fmt.Sprintf("%d", proxy.Conf.RemotePort)
		} else {
			result[i].RemotePort = "N/A"
		}
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

// shouldInterceptKeysForCurrentTab 检查当前标签页是否需要独占键盘输入
func (m *MainDashboard) shouldInterceptKeysForCurrentTab() bool {
	if !m.ready || m.activeTab >= len(m.tabRegistry.GetTabs()) {
		return false
	}

	activeTab := m.tabRegistry.GetTabByIndex(m.activeTab)

	// 检查是否为配置标签页且处于表单编辑模式
	if configTab, ok := activeTab.(*ConfigTab); ok {
		return configTab.IsInFormMode()
	}

	// 可以扩展其他需要独占键盘输入的标签页类型
	return false
}

func (m *MainDashboard) updateStatus(currentTime time.Time) {
	m.statusInfo.LastUpdate = currentTime

	previousServerStatus := m.statusInfo.ServerStatus
	previousClientStatus := m.statusInfo.ClientStatus

	// 更新服务器状态
	if m.checkServerStatus() {
		m.statusInfo.ServerStatus = "运行中"
	} else {
		m.statusInfo.ServerStatus = "已停止"
	}

	statusChanged := (previousServerStatus != m.statusInfo.ServerStatus) ||
		(previousClientStatus != m.statusInfo.ClientStatus)

	shouldUpdateProxy := m.lastProxyUpdate.IsZero() ||
		statusChanged ||
		currentTime.Sub(m.lastProxyUpdate) >= 3*time.Second ||
		(m.statusInfo.ServerStatus == "运行中" && m.statusInfo.ActiveProxies == 0 &&
			currentTime.Sub(m.lastProxyUpdate) >= 1*time.Second)

	if m.apiClient != nil && shouldUpdateProxy {
		m.updateProxyInfo()
		m.lastProxyUpdate = currentTime
	} else if m.statusInfo.ServerStatus != "运行中" {
		m.resetProxyInfo()
	}
}

func (m *MainDashboard) updateProxyInfo() {
	proxies := m.getProxyList()
	m.statusInfo.ActiveProxies = len(proxies)

	if tab, ok := m.tabRegistry.GetTabByIndex(0).(*DashboardTab); ok {
		tab.UpdateProxyList(proxies)
	}

	if m.statusInfo.ServerStatus == "运行中" {
		if serverInfo, err := m.apiClient.GetServerInfo(); err == nil {
			totalTraffic := serverInfo.TotalTrafficIn + serverInfo.TotalTrafficOut
			m.statusInfo.TotalTraffic = service.FormatTraffic(totalTraffic)
		} else if m.statusInfo.TotalTraffic == "" {
			m.statusInfo.TotalTraffic = "N/A"
		}
	} else {
		m.statusInfo.TotalTraffic = "0B"
	}
}

func (m *MainDashboard) resetProxyInfo() {
	m.statusInfo.ActiveProxies = 0
	m.statusInfo.TotalTraffic = "0B"

	if tab, ok := m.tabRegistry.GetTabByIndex(0).(*DashboardTab); ok {
		tab.UpdateProxyList([]ProxyStatus{})
	}

	m.lastProxyUpdate = time.Time{}
}
