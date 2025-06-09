package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FilePickerMode 文件选择器模式
type FilePickerMode int

const (
	FilePickerModeFile FilePickerMode = iota // 选择文件
	FilePickerModeDir                        // 选择目录
	FilePickerModeBoth                       // 文件和目录都可选择
)

// FileItem 文件项
type FileItem struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	ModTime string
}

// FilePickerResult 文件选择结果
type FilePickerResult struct {
	Selected bool
	Path     string
	IsDir    bool
}

// filePickerResultMsg 文件选择结果消息
type filePickerResultMsg FilePickerResult

// FilePicker 文件选择器组件
type FilePicker struct {
	title       string
	mode        FilePickerMode
	currentPath string
	items       []FileItem
	selectedIdx int
	width       int
	height      int
	visible     bool
	showHidden  bool
	extensions  []string // 允许的文件扩展名（为空表示所有文件）
}

// NewFilePicker 创建文件选择器
func NewFilePicker(title string, mode FilePickerMode) *FilePicker {
	currentDir, _ := os.Getwd()

	fp := &FilePicker{
		title:       title,
		mode:        mode,
		currentPath: currentDir,
		selectedIdx: 0,
		visible:     false,
		showHidden:  false,
	}

	fp.loadDirectory()
	return fp
}

// SetExtensions 设置允许的文件扩展名
func (fp *FilePicker) SetExtensions(extensions []string) {
	fp.extensions = extensions
	fp.loadDirectory()
}

// Show 显示文件选择器
func (fp *FilePicker) Show() tea.Cmd {
	fp.visible = true
	fp.loadDirectory()
	return nil
}

// Hide 隐藏文件选择器
func (fp *FilePicker) Hide() {
	fp.visible = false
}

// IsVisible 检查是否可见
func (fp *FilePicker) IsVisible() bool {
	return fp.visible
}

// SetSize 设置大小
func (fp *FilePicker) SetSize(width, height int) {
	fp.width = width
	fp.height = height
}

// SetStartPath 设置起始路径
func (fp *FilePicker) SetStartPath(path string) {
	if path != "" {
		fp.currentPath = path
		fp.loadDirectory()
		fp.selectedIdx = 0
	}
}

// loadDirectory 加载目录内容
func (fp *FilePicker) loadDirectory() {
	fp.items = []FileItem{}

	// 添加上级目录项（除非在根目录）
	if fp.currentPath != "/" && fp.currentPath != "" {
		parent := filepath.Dir(fp.currentPath)
		if parent != fp.currentPath {
			fp.items = append(fp.items, FileItem{
				Name:  "..",
				Path:  parent,
				IsDir: true,
			})
		}
	}

	// 读取当前目录
	entries, err := os.ReadDir(fp.currentPath)
	if err != nil {
		return
	}

	var dirs, files []FileItem

	for _, entry := range entries {
		// 跳过隐藏文件（除非显示隐藏文件）
		if !fp.showHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		item := FileItem{
			Name:    entry.Name(),
			Path:    filepath.Join(fp.currentPath, entry.Name()),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format("2006-01-02 15:04"),
		}

		if entry.IsDir() {
			dirs = append(dirs, item)
		} else {
			// 检查文件扩展名
			if len(fp.extensions) > 0 {
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				allowed := false
				for _, allowedExt := range fp.extensions {
					if ext == strings.ToLower(allowedExt) {
						allowed = true
						break
					}
				}
				if !allowed {
					continue
				}
			}
			files = append(files, item)
		}
	}

	// 排序
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Name < dirs[j].Name
	})
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	// 合并目录和文件
	fp.items = append(fp.items, dirs...)
	if fp.mode != FilePickerModeDir {
		fp.items = append(fp.items, files...)
	}

	// 重置选择索引
	if fp.selectedIdx >= len(fp.items) {
		fp.selectedIdx = 0
	}
}

// Update 更新状态
func (fp *FilePicker) Update(msg tea.Msg) tea.Cmd {
	if !fp.visible {
		return nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		fp.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// 取消选择
			fp.Hide()
			return func() tea.Msg {
				return filePickerResultMsg{Selected: false}
			}

		case "up", "k":
			if fp.selectedIdx > 0 {
				fp.selectedIdx--
			}

		case "down", "j":
			if fp.selectedIdx < len(fp.items)-1 {
				fp.selectedIdx++
			}

		case "enter", " ":
			if fp.selectedIdx < len(fp.items) {
				item := fp.items[fp.selectedIdx]

				if item.IsDir {
					if item.Name == ".." {
						// 返回上级目录
						fp.currentPath = item.Path
					} else {
						// 进入子目录
						fp.currentPath = item.Path
					}
					fp.loadDirectory()
					fp.selectedIdx = 0
				} else {
					// 选择文件
					fp.Hide()
					return func() tea.Msg {
						return filePickerResultMsg{
							Selected: true,
							Path:     item.Path,
							IsDir:    false,
						}
					}
				}
			}

		case "ctrl+d":
			// 选择当前目录（仅在目录模式下）
			if fp.mode == FilePickerModeDir || fp.mode == FilePickerModeBoth {
				fp.Hide()
				return func() tea.Msg {
					return filePickerResultMsg{
						Selected: true,
						Path:     fp.currentPath,
						IsDir:    true,
					}
				}
			}

		case "ctrl+h":
			// 切换显示隐藏文件
			fp.showHidden = !fp.showHidden
			fp.loadDirectory()

		case "home":
			// 回到用户主目录
			if homeDir, err := os.UserHomeDir(); err == nil {
				fp.currentPath = homeDir
				fp.loadDirectory()
				fp.selectedIdx = 0
			}
		}
	}

	return nil
}

// View 渲染视图
func (fp *FilePicker) View() string {
	if !fp.visible {
		return ""
	}

	// 计算对话框大小
	dialogWidth := fp.width * 3 / 4
	if dialogWidth < 60 {
		dialogWidth = 60
	}
	if dialogWidth > 100 {
		dialogWidth = 100
	}

	dialogHeight := fp.height * 3 / 4
	if dialogHeight < 20 {
		dialogHeight = 20
	}
	if dialogHeight > 30 {
		dialogHeight = 30
	}

	// 对话框样式
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1).
		Width(dialogWidth).
		Height(dialogHeight).
		Background(lipgloss.Color("#1e1e1e"))

	// 标题样式
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Padding(0, 0, 1, 0)

	// 路径样式
	pathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 0, 1, 0)

	// 选中项样式
	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(0, 1)

	// 普通项样式
	normalStyle := lipgloss.NewStyle().
		Padding(0, 1)

	// 目录样式
	dirStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Padding(0, 1)

	// 构建内容
	var content strings.Builder

	// 标题
	content.WriteString(titleStyle.Render(fp.title))
	content.WriteString("\n")

	// 当前路径
	content.WriteString(pathStyle.Render("📁 " + fp.currentPath))
	content.WriteString("\n")

	// 文件列表
	listHeight := dialogHeight - 8 // 减去标题、路径、帮助等占用的行数
	startIdx := 0
	endIdx := len(fp.items)

	// 如果列表太长，只显示当前选择项周围的内容
	if len(fp.items) > listHeight {
		startIdx = fp.selectedIdx - listHeight/2
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + listHeight
		if endIdx > len(fp.items) {
			endIdx = len(fp.items)
			startIdx = endIdx - listHeight
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	for i := startIdx; i < endIdx; i++ {
		item := fp.items[i]

		var icon, name string
		if item.IsDir {
			if item.Name == ".." {
				icon = "↑"
				name = item.Name
			} else {
				icon = "📁"
				name = item.Name
			}
		} else {
			icon = "📄"
			name = item.Name
		}

		line := fmt.Sprintf("%s %s", icon, name)

		if i == fp.selectedIdx {
			content.WriteString(selectedStyle.Render("▶ " + line))
		} else if item.IsDir {
			content.WriteString(dirStyle.Render("  " + line))
		} else {
			content.WriteString(normalStyle.Render("  " + line))
		}
		content.WriteString("\n")
	}

	// 添加空行填充
	for i := endIdx - startIdx; i < listHeight; i++ {
		content.WriteString("\n")
	}

	// 帮助信息
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(1, 0, 0, 0)

	var helpText string
	switch fp.mode {
	case FilePickerModeFile:
		helpText = "↑/↓ 导航 | Enter 选择文件/进入目录 | ESC 取消"
	case FilePickerModeDir:
		helpText = "↑/↓ 导航 | Enter 进入目录 | Ctrl+D 选择当前目录 | ESC 取消"
	case FilePickerModeBoth:
		helpText = "↑/↓ 导航 | Enter 选择/进入 | Ctrl+D 选择当前目录 | ESC 取消"
	}
	helpText += " | Ctrl+H 显示隐藏文件 | Home 回到主目录"

	content.WriteString(helpStyle.Render(helpText))

	// 居中显示对话框
	return lipgloss.Place(
		fp.width, fp.height,
		lipgloss.Center, lipgloss.Center,
		dialogStyle.Render(content.String()),
	)
}

// GetResult 获取选择结果（用于处理filePickerResultMsg）
func GetFilePickerResult(msg tea.Msg) (FilePickerResult, bool) {
	if result, ok := msg.(filePickerResultMsg); ok {
		return FilePickerResult(result), true
	}
	return FilePickerResult{}, false
}
