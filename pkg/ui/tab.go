package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Tab 定义标签页接口
type Tab interface {
	// Title 返回标签页标题
	Title() string

	// Init 初始化标签页
	Init() tea.Cmd

	// Update 更新标签页状态
	Update(msg tea.Msg) (Tab, tea.Cmd)

	// View 渲染标签页内容
	View(width int, height int) string

	// SetSize 设置标签页大小
	SetSize(width int, height int)

	// Focusable 是否可接收焦点
	Focusable() bool

	// Focused 当前是否获得焦点
	Focused() bool

	// Focus 设置焦点状态
	Focus(focused bool)
}

// BaseTab 提供Tab接口的基本实现
type BaseTab struct {
	title     string
	width     int
	height    int
	focused   bool
	focusable bool
}

// NewBaseTab 创建一个新的基本标签页
func NewBaseTab(title string) BaseTab {
	return BaseTab{
		title:     title,
		focused:   false,
		focusable: false,
	}
}

// Title 获取标签页标题
func (b BaseTab) Title() string {
	return b.title
}

// SetSize 设置标签页大小
func (b *BaseTab) SetSize(width int, height int) {
	b.width = width
	b.height = height
}

// Focusable 返回标签是否可接收焦点
func (b BaseTab) Focusable() bool {
	return b.focusable
}

// Focused 返回标签是否获得焦点
func (b BaseTab) Focused() bool {
	return b.focused
}

// Focus 设置焦点状态
func (b *BaseTab) Focus(focused bool) {
	b.focused = focused
}

// TabRegistry 标签页注册中心
type TabRegistry struct {
	tabs []Tab
}

// NewTabRegistry 创建一个新的标签页注册中心
func NewTabRegistry() *TabRegistry {
	return &TabRegistry{
		tabs: make([]Tab, 0),
	}
}

// Register 注册一个新的标签页
func (tr *TabRegistry) Register(tab Tab) {
	tr.tabs = append(tr.tabs, tab)
}

// GetTabs 获取所有已注册的标签页
func (tr *TabRegistry) GetTabs() []Tab {
	return tr.tabs
}

// GetTabTitles 获取所有标签页的标题
func (tr *TabRegistry) GetTabTitles() []string {
	titles := make([]string, len(tr.tabs))
	for i, tab := range tr.tabs {
		titles[i] = tab.Title()
	}
	return titles
}

// GetTabByIndex 根据索引获取标签页
func (tr *TabRegistry) GetTabByIndex(index int) Tab {
	if index < 0 || index >= len(tr.tabs) {
		return nil
	}
	return tr.tabs[index]
}
