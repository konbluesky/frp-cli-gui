package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// AppLayoutConfig 应用布局配置
type AppLayoutConfig struct {
	// 主题颜色
	PrimaryColor   string // 主色调，默认 "#7D56F4"
	SecondaryColor string // 次要色调，默认 "#57"
	BorderColor    string // 边框颜色，默认 "240"
	HelpColor      string // 帮助文本颜色，默认 "240"
	StatusColor    string // 状态文本颜色，默认 "240"

	// 布局设置
	ShowTitle     bool // 是否显示标题
	ShowTabs      bool // 是否显示标签页
	ShowBottomBar bool // 是否显示底部栏（包含帮助和状态信息）

	// 自定义内容
	Title       string   // 应用标题
	Tabs        []string // 标签页列表
	ActiveTab   int      // 当前活跃标签
	StatusText  string   // 状态栏文本（显示在底部右侧）
	HelpText    string   // 帮助文本（显示在底部左侧）
	MainContent string   // 主内容区域
}

// AppLayout 通用应用布局渲染器
type AppLayout struct {
	width  int
	height int
	config AppLayoutConfig
}

// NewAppLayout 创建新的应用布局
func NewAppLayout(width, height int) *AppLayout {
	return &AppLayout{
		width:  width,
		height: height,
		config: AppLayoutConfig{
			PrimaryColor:   "#7D56F4",
			SecondaryColor: "#57",
			BorderColor:    "240",
			HelpColor:      "240",
			StatusColor:    "240",
			ShowTitle:      true,
			ShowTabs:       true,
			ShowBottomBar:  true,
		},
	}
}

// SetSize 设置布局尺寸
func (al *AppLayout) SetSize(width, height int) {
	al.width = width
	al.height = height
}

// UpdateConfig 更新部分配置
func (al *AppLayout) UpdateConfig(updater func(*AppLayoutConfig)) {
	updater(&al.config)
}

// Render 渲染完整的应用布局
func (al *AppLayout) Render() string {
	if al.width == 0 {
		return "正在加载..."
	}

	// 创建样式
	styles := al.createStyles()

	// 构建各个组件
	var components []string

	// 标题
	if al.config.ShowTitle && al.config.Title != "" {
		title := styles.title.Render(al.config.Title)
		components = append(components, title, "")
	}

	// 标签页
	if al.config.ShowTabs && len(al.config.Tabs) > 0 {
		tabsRow := al.renderTabs(styles)
		components = append(components, tabsRow, "")
	}

	// 主内容
	if al.config.MainContent != "" {
		components = append(components, al.config.MainContent)
	}

	// 组合内容
	innerContent := lipgloss.JoinVertical(lipgloss.Left, components...)

	// 处理底部栏和对齐
	var finalContent string
	if al.config.ShowBottomBar && (al.config.HelpText != "" || al.config.StatusText != "") {
		bottomBar := al.renderBottomBar(styles)
		finalContent = al.alignBottomBarToBottom(innerContent, bottomBar)
	} else {
		finalContent = innerContent
	}

	// 应用整体边框
	return styles.appBorder.Render(finalContent)
}

// createStyles 创建所有样式
func (al *AppLayout) createStyles() appStyles {
	return appStyles{
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color(al.config.PrimaryColor)).
			Padding(1, 1).
			Width(al.width - 4).
			Align(lipgloss.Center),

		tab: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(al.config.BorderColor)).
			Padding(0, 5),

		activeTab: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(al.config.SecondaryColor)).
			Foreground(lipgloss.Color(al.config.SecondaryColor)).
			Padding(0, 5),

		help: lipgloss.NewStyle().
			Foreground(lipgloss.Color(al.config.HelpColor)),

		status: lipgloss.NewStyle().
			Foreground(lipgloss.Color(al.config.StatusColor)),

		appBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(al.config.PrimaryColor)).
			Padding(1),
	}
}

// appStyles 应用样式集合
type appStyles struct {
	title     lipgloss.Style
	tab       lipgloss.Style
	activeTab lipgloss.Style
	help      lipgloss.Style
	status    lipgloss.Style
	appBorder lipgloss.Style
}

// renderTabs 渲染标签页
func (al *AppLayout) renderTabs(styles appStyles) string {
	var tabs []string
	for i, tab := range al.config.Tabs {
		if i == al.config.ActiveTab {
			tabs = append(tabs, styles.activeTab.Render(tab))
		} else {
			tabs = append(tabs, styles.tab.Render(tab))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

// renderBottomBar 渲染底部栏（左侧帮助信息，右侧状态信息）
func (al *AppLayout) renderBottomBar(styles appStyles) string {
	// 计算可用宽度（减去边框和内边距）
	availableWidth := al.width - 4

	var leftContent, rightContent string

	// 左侧帮助信息
	if al.config.HelpText != "" {
		leftContent = styles.help.Render(al.config.HelpText)
	}

	// 右侧状态信息
	if al.config.StatusText != "" {
		rightContent = styles.status.Render(al.config.StatusText)
	}

	// 如果只有一侧有内容，直接返回
	if leftContent == "" && rightContent == "" {
		return ""
	}
	if leftContent == "" {
		return lipgloss.NewStyle().Width(availableWidth).Align(lipgloss.Right).Render(rightContent)
	}
	if rightContent == "" {
		return lipgloss.NewStyle().Width(availableWidth).Align(lipgloss.Left).Render(leftContent)
	}

	// 计算左右内容的实际宽度
	leftWidth := lipgloss.Width(leftContent)
	rightWidth := lipgloss.Width(rightContent)

	// 计算中间需要的空白宽度
	middleWidth := availableWidth - leftWidth - rightWidth
	if middleWidth < 0 {
		middleWidth = 0
	}

	// 创建中间的空白填充
	middle := strings.Repeat(" ", middleWidth)

	// 组合左侧、中间空白、右侧
	return leftContent + middle + rightContent
}

// alignBottomBarToBottom 将底部栏对齐到底部
func (al *AppLayout) alignBottomBarToBottom(innerContent, bottomBar string) string {
	contentHeight := lipgloss.Height(innerContent)
	verticalPaddingAndBorder := 4 // appBorderStyle 的垂直边距
	remainingHeight := al.height - contentHeight - lipgloss.Height(bottomBar) - verticalPaddingAndBorder

	if remainingHeight < 0 {
		remainingHeight = 0
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		innerContent,
		lipgloss.PlaceVertical(remainingHeight, lipgloss.Bottom, bottomBar),
	)
}

// RenderDialog 渲染居中对话框
func (al *AppLayout) RenderDialog(content string, options DialogOptions) string {
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(options.BorderColor)).
		Background(lipgloss.Color(options.BackgroundColor)).
		Foreground(lipgloss.Color(options.ForegroundColor)).
		Padding(1, 2).
		Width(options.Width).
		Align(lipgloss.Center)

	dialog := dialogStyle.Render(content)

	if al.width > 0 && al.height > 0 {
		return lipgloss.Place(al.width, al.height,
			lipgloss.Center, lipgloss.Center,
			dialog,
			lipgloss.WithWhitespaceBackground(lipgloss.Color(options.OverlayColor)))
	}

	return dialog
}

// DialogOptions 对话框选项
type DialogOptions struct {
	Width           int
	BorderColor     string
	BackgroundColor string
	ForegroundColor string
	OverlayColor    string
}

// DefaultDialogOptions 默认对话框选项
func DefaultDialogOptions() DialogOptions {
	return DialogOptions{
		Width:           50,
		BorderColor:     "#FF6B6B",
		BackgroundColor: "#1A1A1A",
		ForegroundColor: "#FFFFFF",
		OverlayColor:    "235",
	}
}

// ContentRenderer 内容渲染器接口
type ContentRenderer interface {
	RenderContent(width, height int) string
}

// SimpleContentRenderer 简单内容渲染器
type SimpleContentRenderer struct {
	content string
	style   lipgloss.Style
}

// NewSimpleContentRenderer 创建简单内容渲染器
func NewSimpleContentRenderer(content string) *SimpleContentRenderer {
	return &SimpleContentRenderer{
		content: content,
		style: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1),
	}
}

// SetStyle 设置样式
func (scr *SimpleContentRenderer) SetStyle(style lipgloss.Style) {
	scr.style = style
}

// RenderContent 渲染内容
func (scr *SimpleContentRenderer) RenderContent(width, height int) string {
	return scr.style.Width(width - 8).Render(scr.content)
}
