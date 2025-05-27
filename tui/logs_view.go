package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"frp-cli-ui/service"
)

// LogsView 日志查看器模型
type LogsView struct {
	width        int
	height       int
	viewport     viewport.Model
	filter       textinput.Model
	logs         []LogEntry
	filteredLogs []LogEntry
	autoScroll   bool
	manager      *service.Manager
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Source    string // "server" 或 "client"
}

// NewLogsView 创建新的日志查看器
func NewLogsView() LogsView {
	// 初始化视口
	vp := viewport.New(80, 20)
	vp.YPosition = 0

	// 初始化过滤器
	filter := textinput.New()
	filter.Placeholder = "输入关键词过滤日志..."
	filter.CharLimit = 100
	filter.Width = 50

	// 初始化空日志列表
	logs := []LogEntry{}

	lv := LogsView{
		viewport:     vp,
		filter:       filter,
		logs:         logs,
		filteredLogs: logs,
		autoScroll:   true,
		manager:      nil, // 稍后设置
	}

	lv.updateViewport()
	return lv
}

// NewLogsViewWithManager 创建带管理器的日志查看器
func NewLogsViewWithManager(manager *service.Manager) LogsView {
	lv := NewLogsView()
	lv.manager = manager
	return lv
}

// SetManager 设置管理器
func (m *LogsView) SetManager(manager *service.Manager) {
	m.manager = manager
}

// Init 初始化
func (m LogsView) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
	)
}

// Update 更新状态
func (m LogsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 8 // 为标题、过滤器和帮助信息留空间

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "q":
			// q 键退出应用
			return m, tea.Quit

		case "x":
			// x 键返回上级（这会被 Dashboard 捕获）
			return m, tea.Quit

		case "ctrl+f", "f":
			// 切换到过滤器
			m.filter.Focus()
			return m, textinput.Blink

		case "esc":
			// 退出过滤器焦点，如果没有焦点则返回上级
			if m.filter.Focused() {
				m.filter.Blur()
			} else {
				// 如果过滤器没有焦点，esc 键也可以返回上级
				return m, tea.Quit
			}

		case "ctrl+a", "a":
			// 切换自动滚动
			m.autoScroll = !m.autoScroll

		case "ctrl+l":
			// 清空日志
			m.logs = []LogEntry{}
			m.filteredLogs = []LogEntry{}
			m.updateViewport()

		case "enter":
			if m.filter.Focused() {
				// 应用过滤器
				m.applyFilter()
				m.filter.Blur()
			}
		}

	case tickMsg:
		// 从管理器获取真实日志
		if m.manager != nil {
			m.collectLogsFromManager()
		}

		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}

	// 更新过滤器
	if m.filter.Focused() {
		m.filter, cmd = m.filter.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		// 更新视口
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// applyFilter 应用过滤器
func (m *LogsView) applyFilter() {
	filterText := strings.ToLower(m.filter.Value())

	if filterText == "" {
		m.filteredLogs = m.logs
	} else {
		m.filteredLogs = []LogEntry{}
		for _, log := range m.logs {
			if strings.Contains(strings.ToLower(log.Message), filterText) ||
				strings.Contains(strings.ToLower(log.Level), filterText) ||
				strings.Contains(strings.ToLower(log.Source), filterText) {
				m.filteredLogs = append(m.filteredLogs, log)
			}
		}
	}

	m.updateViewport()
}

// updateViewport 更新视口内容
func (m *LogsView) updateViewport() {
	var content strings.Builder

	for _, log := range m.filteredLogs {
		// 根据日志级别设置颜色
		var levelStyle lipgloss.Style
		switch log.Level {
		case "ERROR":
			levelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		case "WARN":
			levelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
		case "INFO":
			levelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
		default:
			levelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		}

		// 根据来源设置样式
		var sourceStyle lipgloss.Style
		if log.Source == "server" {
			sourceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
		} else {
			sourceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("213"))
		}

		// 格式化日志行
		timestamp := log.Timestamp.Format("15:04:05")
		level := levelStyle.Render(fmt.Sprintf("[%s]", log.Level))
		source := sourceStyle.Render(fmt.Sprintf("[%s]", log.Source))

		content.WriteString(fmt.Sprintf("%s %s %s %s\n",
			timestamp, level, source, log.Message))
	}

	m.viewport.SetContent(content.String())

	// 自动滚动到底部
	if m.autoScroll {
		m.viewport.GotoBottom()
	}
}

// View 渲染视图
func (m LogsView) View() string {
	// 如果没有设置窗口大小，使用默认值
	if m.width == 0 {
		m.width = 80
		m.height = 24
		m.viewport.Width = m.width - 4
		m.viewport.Height = m.height - 8
	}

	var b strings.Builder

	// 标题
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	title := fmt.Sprintf("FRP 日志查看器 (共 %d 条，显示 %d 条)",
		len(m.logs), len(m.filteredLogs))
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	// 过滤器
	filterStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(60)

	if m.filter.Focused() {
		filterStyle = filterStyle.BorderForeground(lipgloss.Color("57"))
	}

	b.WriteString("过滤器:")
	b.WriteString("\n")
	b.WriteString(filterStyle.Render(m.filter.View()))
	b.WriteString("\n\n")

	// 日志内容
	logStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(m.width - 4).
		Height(m.height - 10)

	// 如果没有日志，显示提示信息
	viewportContent := m.viewport.View()
	if len(m.filteredLogs) == 0 {
		if m.manager == nil {
			viewportContent = "未连接到日志管理器\n请确保FRP服务正在运行..."
		} else {
			viewportContent = "暂无日志数据\n等待FRP服务产生日志..."
		}
	}

	b.WriteString(logStyle.Render(viewportContent))
	b.WriteString("\n")

	// 状态栏
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	autoScrollStatus := "关闭"
	if m.autoScroll {
		autoScrollStatus = "开启"
	}

	status := fmt.Sprintf("自动滚动: %s | 滚动位置: %d%%",
		autoScrollStatus, int(m.viewport.ScrollPercent()*100))
	b.WriteString(statusStyle.Render(status))
	b.WriteString("\n")

	// 帮助信息
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("F/Ctrl+F: 过滤 | A/Ctrl+A: 切换自动滚动 | Ctrl+L: 清空日志 | ↑↓: 滚动 | X/Esc: 返回 | Q: 退出")

	b.WriteString(help)

	return b.String()
}

// collectLogsFromManager 从管理器收集日志
func (m *LogsView) collectLogsFromManager() {
	if m.manager == nil {
		return
	}

	// 从管理器的日志通道读取新日志
	logChan := m.manager.GetLogChannel()

	// 非阻塞读取所有可用日志
	for {
		select {
		case logMsg := <-logChan:
			// 转换为LogEntry格式
			entry := LogEntry{
				Timestamp: logMsg.Timestamp,
				Level:     logMsg.Level,
				Message:   logMsg.Message,
				Source:    logMsg.Source,
			}

			// 添加到日志列表
			m.logs = append(m.logs, entry)

			// 限制日志数量，避免内存过度使用
			if len(m.logs) > 1000 {
				// 保留最新的800条日志
				m.logs = m.logs[len(m.logs)-800:]
			}

		default:
			// 没有更多日志，退出循环
			goto done
		}
	}

done:
	// 如果有新日志，重新应用过滤器
	if len(m.logs) > 0 {
		m.applyFilter()
	}
}
