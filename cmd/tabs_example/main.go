package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"frp-cli-ui/pkg/ui"
)

// 自定义标签页
type CustomTab struct {
	ui.BaseTab
	content string
}

// 更新标签页状态
func (ct *CustomTab) Update(msg tea.Msg) (ui.Tab, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ct.SetSize(msg.Width, msg.Height)
	}
	return ct, nil
}

// 渲染标签页内容
func (ct *CustomTab) View(width int, height int) string {
	contentWidth := width - 12
	if contentWidth < 20 {
		contentWidth = 20
	}

	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(2).
		Width(contentWidth)

	return contentStyle.Render(ct.content)
}

// 初始化标签页
func (ct *CustomTab) Init() tea.Cmd {
	return nil
}

// 示例程序主入口
type TabsApp struct {
	registry   *ui.TabRegistry
	activeTab  int
	width      int
	height     int
	layout     *ui.AppLayout
	showDialog bool
}

// 初始化程序
func (a *TabsApp) Init() tea.Cmd {
	return nil
}

// 更新程序状态
func (a *TabsApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

		// 初始化布局
		if a.layout == nil {
			a.layout = ui.NewAppLayout(a.width, a.height)
		} else {
			a.layout.SetSize(a.width, a.height)
		}

		// 更新所有标签页大小
		for _, tab := range a.registry.GetTabs() {
			tab.SetSize(a.width, a.height)
		}

	case tea.KeyMsg:
		if a.showDialog {
			switch msg.String() {
			case "y", "Y", "enter":
				return a, tea.Quit
			case "n", "N", "esc":
				a.showDialog = false
				return a, nil
			}
			return a, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			a.showDialog = true
			return a, nil
		case "tab":
			a.activeTab = (a.activeTab + 1) % len(a.registry.GetTabs())
			a.updateFocus()
			return a, nil
		case "shift+tab":
			a.activeTab = (a.activeTab - 1 + len(a.registry.GetTabs())) % len(a.registry.GetTabs())
			a.updateFocus()
			return a, nil
		}
	}

	// 更新当前活动标签页
	if a.activeTab < len(a.registry.GetTabs()) {
		activeTab := a.registry.GetTabByIndex(a.activeTab)
		updatedTab, cmd := activeTab.Update(msg)

		// 更新注册表中的标签页
		tabs := a.registry.GetTabs()
		tabs[a.activeTab] = updatedTab

		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return a, tea.Batch(cmds...)
}

// 渲染视图
func (a *TabsApp) View() string {
	if a.layout == nil {
		return "正在初始化..."
	}

	// 显示确认对话框
	if a.showDialog {
		dialogContent := `确认退出

您确定要退出示例应用吗？

[Y] 是的，退出
[N] 取消

按 Y 确认退出，按 N 或 ESC 取消`

		return a.layout.RenderDialog(dialogContent, ui.DefaultDialogOptions())
	}

	// 配置布局
	a.layout.UpdateConfig(func(config *ui.AppLayoutConfig) {
		config.Title = "示例应用 - 可插拔标签页演示"
		config.Tabs = a.registry.GetTabTitles()
		config.ActiveTab = a.activeTab
		config.StatusText = fmt.Sprintf("当前标签: %s | 标签总数: %d | 演示如何添加新标签页",
			a.registry.GetTabTitles()[a.activeTab],
			len(a.registry.GetTabs()),
		)
		config.HelpText = "Tab: 切换标签 | Shift+Tab: 反向切换 | q: 退出"

		// 获取当前活动标签页内容
		if a.activeTab < len(a.registry.GetTabs()) {
			activeTab := a.registry.GetTabByIndex(a.activeTab)
			config.MainContent = activeTab.View(a.width, a.height)
		}
	})

	return a.layout.Render()
}

// 更新焦点状态
func (a *TabsApp) updateFocus() {
	for i, tab := range a.registry.GetTabs() {
		if tab.Focusable() {
			tab.Focus(i == a.activeTab)
		}
	}
}

// 创建标签页
func createTab(title, content string) ui.Tab {
	tab := &CustomTab{
		BaseTab: ui.NewBaseTab(title),
		content: content,
	}
	return tab
}

// 主函数
func main() {
	fmt.Println("启动可插拔标签页演示程序...")

	// 创建标签页注册表
	registry := ui.NewTabRegistry()

	// 注册各种标签页
	registry.Register(createTab("首页", `
欢迎使用可插拔标签页系统！

这是一个演示如何使用 Tab 接口的示例应用。

主要特性：
• 🎨 基于接口的标签页系统
• 📱 支持动态添加和移除标签页
• 🔧 符合开闭原则的设计
• 🚀 易于扩展

示例代码演示了如何：
- 创建自定义标签页
- 注册到标签页系统
- 动态添加新标签页
- 处理焦点管理
`))

	registry.Register(createTab("数据", `
数据管理页面

这里可以显示各种数据内容：

📊 统计信息：
- 用户数量: 1,234
- 活跃会话: 56
- 数据传输: 2.3GB
- 系统负载: 45%

📈 实时监控：
- CPU 使用率: 23%
- 内存使用率: 67%
- 磁盘使用率: 34%
- 网络流量: 125MB/s

🔄 最近活动：
- 2024-01-15 14:30 - 用户登录
- 2024-01-15 14:25 - 数据同步完成
- 2024-01-15 14:20 - 系统备份开始
`))

	registry.Register(createTab("设置", `
系统设置

⚙️ 应用配置：
- 主题颜色: 紫色 (#7D56F4)
- 语言设置: 中文
- 自动保存: 启用
- 通知提醒: 启用

🎨 界面设置：
- 显示标题: ✓
- 显示标签页: ✓
- 显示状态栏: ✓
- 显示帮助信息: ✓

🔧 高级选项：
- 调试模式: 关闭
- 日志级别: INFO
- 缓存大小: 100MB
- 连接超时: 30秒
`))

	registry.Register(createTab("关于", `
关于此应用

📦 应用信息：
- 名称: 可插拔标签页示例
- 版本: v1.0.0
- 构建时间: 2024-01-15
- Go 版本: 1.21+

🛠️ 技术栈：
- Bubble Tea: TUI 框架
- Lip Gloss: 样式库
- 自定义 Tab 接口系统

👨‍💻 开发者：
- 基于 FRP-CLI-UI 项目提炼
- 遵循六大设计原则
- 支持动态扩展

📄 许可证：
- MIT License
- 开源免费使用
`))

	// 设置字符宽度计算
	runewidth.DefaultCondition.EastAsianWidth = false
	// 创建应用
	app := &TabsApp{
		registry:  registry,
		activeTab: 0,
	}

	// 运行程序
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("程序已退出")
}
