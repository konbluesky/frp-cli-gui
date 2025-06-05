package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfigTab 配置管理标签页
type ConfigTab struct {
	BaseTab
}

// NewConfigTab 创建配置管理标签页
func NewConfigTab() *ConfigTab {
	baseTab := NewBaseTab("配置管理")

	return &ConfigTab{
		BaseTab: baseTab,
	}
}

// Init 初始化
func (ct *ConfigTab) Init() tea.Cmd {
	return nil
}

// Update 更新状态
func (ct *ConfigTab) Update(msg tea.Msg) (Tab, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ct.SetSize(msg.Width, msg.Height)
	}

	return ct, nil
}

// View 渲染视图
func (ct *ConfigTab) View(width int, height int) string {
	contentWidth := width - 12
	if contentWidth < 20 {
		contentWidth = 20
	}

	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(2).
		Width(contentWidth)

	content := `配置管理

暂未实现，将基于 AppLayout 重新开发

可用功能:
• 服务端配置编辑
• 客户端配置编辑
• 配置验证
• 配置导入导出`

	return contentStyle.Render(content)
}
