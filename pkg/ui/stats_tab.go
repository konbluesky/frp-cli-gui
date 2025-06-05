package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StatsTab 统计信息标签页 - 演示如何添加新的标签页
type StatsTab struct {
	BaseTab
	startTime time.Time
	uptime    time.Duration
}

// NewStatsTab 创建统计信息标签页
func NewStatsTab() *StatsTab {
	baseTab := NewBaseTab("统计信息")

	return &StatsTab{
		BaseTab:   baseTab,
		startTime: time.Now(),
		uptime:    0,
	}
}

// Init 初始化
func (st *StatsTab) Init() tea.Cmd {
	return nil
}

// Update 更新状态
func (st *StatsTab) Update(msg tea.Msg) (Tab, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		st.SetSize(msg.Width, msg.Height)
	case dashboardTickMsg:
		st.uptime = time.Since(st.startTime)
	}

	return st, nil
}

// View 渲染视图
func (st *StatsTab) View(width int, height int) string {
	contentWidth := width - 12
	if contentWidth < 20 {
		contentWidth = 20
	}

	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(2).
		Width(contentWidth)

	// 格式化运行时间
	days := int(st.uptime.Hours()) / 24
	hours := int(st.uptime.Hours()) % 24
	minutes := int(st.uptime.Minutes()) % 60
	seconds := int(st.uptime.Seconds()) % 60

	uptimeStr := fmt.Sprintf("%d天 %d小时 %d分钟 %d秒", days, hours, minutes, seconds)

	content := fmt.Sprintf(`系统统计信息

📊 运行统计：
- 启动时间: %s
- 运行时长: %s
- 内存使用: 24.5MB
- CPU使用率: 1.2%%

📈 流量统计：
- 总发送: 1.5GB
- 总接收: 3.2GB
- 连接数: 12
- 峰值带宽: 2.5MB/s

🔍 请求统计：
- 总请求数: 1,245
- 成功率: 99.8%%
- 平均响应时间: 45ms
- 最大响应时间: 350ms`,
		st.startTime.Format("2006-01-02 15:04:05"),
		uptimeStr)

	return contentStyle.Render(content)
}
