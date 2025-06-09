package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"frp-cli-ui/internal/service"
)

// ProxyStatus 代理状态
type ProxyStatus struct {
	Name            string
	Type            string
	LocalAddr       string
	RemotePort      string
	Status          string
	CurConns        int
	TodayTrafficIn  int64
	TodayTrafficOut int64
	ClientVersion   string
	LastStartTime   string
}

// DashboardTab 仪表盘标签页
type DashboardTab struct {
	BaseTab
	table     table.Model
	apiClient *service.APIClient
}

// NewDashboardTab 创建仪表盘标签页
func NewDashboardTab(apiClient *service.APIClient) *DashboardTab {
	// 初始化表格
	columns := []table.Column{
		{Title: "代理名称", Width: 12},
		{Title: "类型", Width: 6},
		{Title: "本地地址", Width: 16},
		{Title: "远程端口", Width: 8},
		{Title: "状态", Width: 8},
		{Title: "连接数", Width: 6},
		{Title: "今日上行", Width: 10},
		{Title: "今日下行", Width: 10},
		{Title: "启动时间", Width: 16},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(10),
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

	baseTab := NewBaseTab("仪表盘")
	baseTab.focusable = true

	return &DashboardTab{
		BaseTab:   baseTab,
		table:     t,
		apiClient: apiClient,
	}
}

// Init 初始化
func (dt *DashboardTab) Init() tea.Cmd {
	return nil
}

// Update 更新状态
func (dt *DashboardTab) Update(msg tea.Msg) (Tab, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		dt.SetSize(msg.Width, msg.Height)
		if dt.width > 20 {
			dt.table.SetWidth(dt.width - 12)
		}
	}

	dt.table, cmd = dt.table.Update(msg)
	return dt, cmd
}

// View 渲染视图
func (dt *DashboardTab) View(width int, height int) string {
	// 动态调整表格宽度，确保适应屏幕
	tableWidth := width - 20 // 为边框和内边距留空间
	if tableWidth < 100 {
		tableWidth = 100 // 最小宽度
	}
	dt.table.SetWidth(tableWidth)

	// 标题样式
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Padding(0, 0, 1, 0)

	// 计算信息卡片宽度，考虑边框、内边距和间距
	// 每个卡片需要：边框(2) + 内边距(2) + 外边距(2) = 6个字符的额外空间
	availableWidth := width - 8            // 为整体布局留边距
	cardWidth := (availableWidth - 24) / 4 // 4个卡片，每个卡片6个字符额外空间
	if cardWidth < 16 {
		cardWidth = 16 // 确保最小宽度
	}

	// 信息卡片样式
	infoCardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Margin(0, 1, 1, 0).
		Width(cardWidth)

	// 创建信息卡片
	serverCard := infoCardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("🎯 服务端"),
			"状态: 运行中",
			"端口: 7000",
		),
	)

	clientCard := infoCardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("💻 客户端"),
			"状态: 已连接",
			fmt.Sprintf("代理: %d 个", len(dt.table.Rows())),
		),
	)

	trafficCard := infoCardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("📈 流量"),
			"上行: 1.2MB",
			"下行: 3.4MB",
		),
	)

	uptimeCard := infoCardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("⏰ 运行时间"),
			"服务端: 2h 15m",
			"客户端: 1h 45m",
		),
	)

	// 水平排列信息卡片
	infoCards := lipgloss.JoinHorizontal(lipgloss.Top, serverCard, clientCard, trafficCard, uptimeCard)

	// 表格标题
	tableTitle := titleStyle.Render("📋 代理状态详情")

	// 表格容器样式
	tableContainerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Margin(1, 0, 0, 0)

	tableContainer := tableContainerStyle.Render(dt.table.View())

	// 如果没有代理，显示提示信息
	var tableContent string
	if len(dt.table.Rows()) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true).
			Align(lipgloss.Center).
			Width(width - 20).
			Padding(2)

		emptyMessage := emptyStyle.Render("暂无活跃代理\n\n请在配置管理中添加代理配置，或启动 FRP 客户端")
		tableContent = tableContainerStyle.Render(emptyMessage)
	} else {
		tableContent = tableContainer
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		infoCards,
		"",
		tableTitle,
		tableContent,
	)
}

// UpdateProxyList 更新代理列表
func (dt *DashboardTab) UpdateProxyList(proxies []ProxyStatus) {
	rows := make([]table.Row, len(proxies))

	for i, proxy := range proxies {
		// 格式化流量显示
		trafficIn := formatTraffic(proxy.TodayTrafficIn)
		trafficOut := formatTraffic(proxy.TodayTrafficOut)

		// 格式化启动时间
		startTime := formatTime(proxy.LastStartTime)

		rows[i] = table.Row{
			proxy.Name,
			proxy.Type,
			proxy.LocalAddr,
			proxy.RemotePort,
			proxy.Status,
			fmt.Sprintf("%d", proxy.CurConns),
			trafficIn,
			trafficOut,
			startTime,
		}
	}

	dt.table.SetRows(rows)
}

// formatTraffic 格式化流量显示
func formatTraffic(bytes int64) string {
	if bytes == 0 {
		return "0B"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f%s", float64(bytes)/float64(div), units[exp])
}

// formatTime 格式化时间显示
func formatTime(timeStr string) string {
	if timeStr == "" {
		return "-"
	}

	// 如果时间字符串太长，只显示日期和时间部分
	if len(timeStr) > 16 {
		return timeStr[:16]
	}

	return timeStr
}
