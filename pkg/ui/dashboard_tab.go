package ui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"frp-cli-ui/internal/service"
)

// ProxyStatus 代理状态
type ProxyStatus struct {
	Name       string
	Type       string
	LocalAddr  string
	RemotePort string
	Status     string
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
		{Title: "代理名称", Width: 15},
		{Title: "类型", Width: 8},
		{Title: "本地地址", Width: 20},
		{Title: "远程端口", Width: 10},
		{Title: "状态", Width: 8},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
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
	// 安全地设置表格宽度
	if width > 20 {
		dt.table.SetWidth(width - 12)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render("代理状态"),
		"",
		dt.table.View(),
	)
}

// UpdateProxyList 更新代理列表
func (dt *DashboardTab) UpdateProxyList(proxies []ProxyStatus) {
	rows := make([]table.Row, len(proxies))

	for i, proxy := range proxies {
		rows[i] = table.Row{
			proxy.Name,
			proxy.Type,
			proxy.LocalAddr,
			proxy.RemotePort,
			proxy.Status,
		}
	}

	dt.table.SetRows(rows)
}
