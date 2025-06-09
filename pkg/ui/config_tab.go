package ui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"

	"frp-cli-ui/pkg/config"
)

// ConfigTabState 配置标签页状态
type ConfigTabState int

const (
	ConfigTabMenu ConfigTabState = iota
	ConfigTabServerForm
	ConfigTabClientForm
	ConfigTabProxyForm
	ConfigTabVisitorForm
	ConfigTabPreview
)

// ConfigTab 配置管理标签页
type ConfigTab struct {
	BaseTab
	state            ConfigTabState
	currentForm      *ConfigFormModel
	serverConfig     *config.Config
	clientConfig     *config.Config
	currentProxy     *config.ProxyConfig
	currentVisitor   *config.VisitorConfig
	loader           *config.Loader
	menuItems        []string
	selectedItem     int
	focusOnForm      bool // 新增：标记焦点是否在表单上
	filePicker       *FilePicker
	serverConfigPath string
	clientConfigPath string
}

// NewConfigTab 创建配置管理标签页
func NewConfigTab() *ConfigTab {
	baseTab := NewBaseTab("配置管理")
	baseTab.focusable = true

	return &ConfigTab{
		BaseTab:          baseTab,
		state:            ConfigTabMenu,
		menuItems:        []string{"🎯 服务端配置", "💻 客户端配置", "🔗 添加代理", "👥 添加访问者", "📁 选择配置文件", "👀 预览配置", "💾 保存配置"},
		selectedItem:     0,
		focusOnForm:      false,
		serverConfigPath: config.GetDefaultServerConfigPath(),
		clientConfigPath: config.GetDefaultClientConfigPath(),
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
		if ct.currentForm != nil {
			form, cmd := ct.currentForm.Update(msg)
			if f, ok := form.(*ConfigFormModel); ok {
				ct.currentForm = f
			}
			return ct, cmd
		}

	case tea.KeyMsg:
		if !ct.focused {
			return ct, nil
		}

		// 如果文件选择器可见，优先处理文件选择器事件
		if ct.filePicker != nil && ct.filePicker.IsVisible() {
			cmd := ct.filePicker.Update(msg)
			return ct, cmd
		}

		// 根据焦点位置处理键盘事件
		if ct.focusOnForm && ct.currentForm != nil {
			// 表单有焦点时，优先处理表单内的Tab/Shift+Tab
			switch msg.String() {
			case "esc":
				// ESC 用于退出表单编辑
				ct.focusOnForm = false
				return ct, nil
			case "ctrl+tab":
				// Ctrl+Tab 用于切换到菜单焦点
				ct.focusOnForm = false
				return ct, nil
			default:
				// 其他所有键盘事件（包括tab/shift+tab）传递给表单处理
				form, cmd := ct.currentForm.Update(msg)
				if f, ok := form.(*ConfigFormModel); ok {
					ct.currentForm = f
				}
				return ct, cmd
			}
		} else {
			// 菜单有焦点时的全局快捷键处理
			switch msg.String() {
			case "esc":
				// ESC 用于清除选择
				if ct.state != ConfigTabMenu {
					ct.state = ConfigTabMenu
					ct.currentForm = nil
					ct.focusOnForm = false
					return ct, nil
				}
			case "tab", "ctrl+tab":
				// Tab 用于切换到表单焦点
				if ct.currentForm != nil {
					ct.focusOnForm = true
					return ct, nil
				}
			}
			// 菜单有焦点时，处理菜单导航
			switch msg.String() {
			case "up", "k":
				if ct.selectedItem > 0 {
					ct.selectedItem--
				}
			case "down", "j":
				if ct.selectedItem < len(ct.menuItems)-1 {
					ct.selectedItem++
				}
			case "enter", " ":
				return ct.handleMenuSelection()
			}
		}

	default:
		// 处理文件选择器结果
		if result, ok := GetFilePickerResult(msg); ok {
			return ct.handleFilePickerResult(result)
		}

		// 表单模式下，将所有其他消息传递给表单处理
		if ct.currentForm != nil {
			form, cmd := ct.currentForm.Update(msg)
			if f, ok := form.(*ConfigFormModel); ok {
				ct.currentForm = f
			}
			return ct, cmd
		}
	}

	return ct, nil
}

// handleMenuSelection 处理菜单选择
func (ct *ConfigTab) handleMenuSelection() (Tab, tea.Cmd) {
	switch ct.selectedItem {
	case 0: // 🎯 服务端配置
		return ct.handleServerConfig()

	case 1: // 💻 客户端配置
		return ct.handleClientConfig()

	case 2: // 🔗 添加代理
		return ct.handleAddProxy()

	case 3: // 👥 添加访问者
		return ct.handleAddVisitor()

	case 4: // 📁 选择配置文件
		return ct.handleChangeConfigFile()

	case 5: // 👀️ 预览配置
		return ct.handlePreviewConfig()

	case 6: // 💾 保存配置
		return ct.handleSaveAllConfigs()
	}

	return ct, nil
}

// handleServerConfig 处理服务端配置
func (ct *ConfigTab) handleServerConfig() (Tab, tea.Cmd) {
	if ct.serverConfig == nil {
		ct.serverConfig = config.CreateDefaultServerConfig()
	}
	ct.currentForm = NewServerConfigForm(ct.serverConfig)
	ct.state = ConfigTabServerForm
	ct.focusOnForm = true
	return ct, ct.currentForm.Init()
}

// handleClientConfig 处理客户端配置
func (ct *ConfigTab) handleClientConfig() (Tab, tea.Cmd) {
	if ct.clientConfig == nil {
		ct.clientConfig = config.CreateDefaultClientConfig()
	}
	ct.currentForm = NewClientConfigForm(ct.clientConfig)
	ct.state = ConfigTabClientForm
	ct.focusOnForm = true
	return ct, ct.currentForm.Init()
}

// handleAddProxy 处理添加代理
func (ct *ConfigTab) handleAddProxy() (Tab, tea.Cmd) {
	ct.currentProxy = &config.ProxyConfig{
		Type:    "tcp",
		LocalIP: "127.0.0.1",
	}
	ct.currentForm = NewProxyConfigForm(ct.currentProxy)
	ct.state = ConfigTabProxyForm
	ct.focusOnForm = true
	return ct, ct.currentForm.Init()
}

// handleAddVisitor 处理添加访问者
func (ct *ConfigTab) handleAddVisitor() (Tab, tea.Cmd) {
	ct.currentVisitor = &config.VisitorConfig{
		Type:     "stcp",
		BindAddr: "127.0.0.1",
	}
	ct.currentForm = NewVisitorConfigForm(ct.currentVisitor)
	ct.state = ConfigTabVisitorForm
	ct.focusOnForm = true
	return ct, ct.currentForm.Init()
}

// handleChangeConfigFile 处理更换配置文件
func (ct *ConfigTab) handleChangeConfigFile() (Tab, tea.Cmd) {
	// 显示配置文件选择菜单
	ct.state = ConfigTabMenu
	// 这里可以扩展为一个子菜单，让用户选择是更换服务端还是客户端配置文件
	ct.filePicker = NewFilePicker("选择配置文件", FilePickerModeFile)
	ct.filePicker.SetExtensions([]string{".yaml", ".yml", ".toml", ".ini"})
	ct.filePicker.SetStartPath(config.GetDefaultWorkDir())
	ct.filePicker.SetSize(ct.width, ct.height)
	return ct, ct.filePicker.Show()
}

// handlePreviewConfig 处理预览配置
func (ct *ConfigTab) handlePreviewConfig() (Tab, tea.Cmd) {
	ct.state = ConfigTabPreview
	ct.focusOnForm = false
	return ct, nil
}

// handleSaveAllConfigs 处理保存所有配置
func (ct *ConfigTab) handleSaveAllConfigs() (Tab, tea.Cmd) {
	// 自动保存到当前设置的配置文件路径
	if ct.serverConfig != nil {
		loader := config.NewLoader(ct.serverConfigPath)
		if err := loader.Save(ct.serverConfig); err == nil {
			// 保存成功
		}
	}

	if ct.clientConfig != nil {
		loader := config.NewLoader(ct.clientConfigPath)
		if err := loader.Save(ct.clientConfig); err == nil {
			// 保存成功
		}
	}

	return ct, nil
}

// handleFilePickerResult 处理文件选择器结果
func (ct *ConfigTab) handleFilePickerResult(result FilePickerResult) (Tab, tea.Cmd) {
	if !result.Selected {
		return ct, nil
	}

	// 根据当前选择的菜单项确定是服务端还是客户端配置文件
	switch ct.selectedItem {
	case 4: // 选择服务端配置文件
		ct.serverConfigPath = result.Path
		// 自动加载选择的服务端配置
		if loader := config.NewLoader(result.Path); loader != nil {
			if cfg, err := loader.Load(); err == nil {
				ct.serverConfig = cfg
			}
		}

	case 5: // 选择客户端配置文件
		ct.clientConfigPath = result.Path
		// 自动加载选择的客户端配置
		if loader := config.NewLoader(result.Path); loader != nil {
			if cfg, err := loader.Load(); err == nil {
				ct.clientConfig = cfg
			}
		}
	}

	return ct, nil
}

// loadConfigFile 加载配置文件
func (ct *ConfigTab) loadConfigFile() (Tab, tea.Cmd) {
	// 使用当前设置的配置文件路径
	if _, err := os.Stat(ct.serverConfigPath); err == nil {
		loader := config.NewLoader(ct.serverConfigPath)
		if cfg, err := loader.Load(); err == nil {
			ct.serverConfig = cfg
		}
	}

	if _, err := os.Stat(ct.clientConfigPath); err == nil {
		loader := config.NewLoader(ct.clientConfigPath)
		if cfg, err := loader.Load(); err == nil {
			ct.clientConfig = cfg
		}
	}

	return ct, nil
}

// saveConfigFile 保存配置文件
func (ct *ConfigTab) saveConfigFile() (Tab, tea.Cmd) {
	configDir := "configs"
	os.MkdirAll(configDir, 0755)

	if ct.serverConfig != nil {
		loader := config.NewLoader(filepath.Join(configDir, "frps.yaml"))
		loader.Save(ct.serverConfig)
	}

	if ct.clientConfig != nil {
		loader := config.NewLoader(filepath.Join(configDir, "frpc.yaml"))
		loader.Save(ct.clientConfig)
	}

	return ct, nil
}

// IsInFormMode 检查是否处于表单编辑模式
func (ct *ConfigTab) IsInFormMode() bool {
	return ct.focusOnForm && ct.currentForm != nil
}

// View 渲染视图 - 新的左右分栏布局
func (ct *ConfigTab) View(width int, height int) string {
	// 如果文件选择器可见，显示文件选择器
	if ct.filePicker != nil && ct.filePicker.IsVisible() {
		return ct.filePicker.View()
	}

	contentWidth := width - 12
	if contentWidth < 60 {
		contentWidth = 60
	}

	// 计算左右分栏宽度
	leftWidth := contentWidth / 3
	rightWidth := contentWidth - leftWidth - 4

	// 确保最小宽度
	if leftWidth < 25 {
		leftWidth = 25
		rightWidth = contentWidth - leftWidth - 4
	}

	// 计算可用高度，为标题栏、状态栏等预留空间
	availableHeight := height - 6
	if availableHeight < 10 {
		availableHeight = 10
	}

	// 左侧菜单样式
	leftStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(leftWidth).
		MaxHeight(availableHeight)

	// 右侧内容样式
	rightStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(rightWidth).
		MaxHeight(availableHeight)

	// 如果表单有焦点，高亮右侧边框
	if ct.focusOnForm {
		rightStyle = rightStyle.BorderForeground(lipgloss.Color("#7D56F4"))
	}

	// 如果菜单有焦点，高亮左侧边框
	if !ct.focusOnForm {
		leftStyle = leftStyle.BorderForeground(lipgloss.Color("#7D56F4"))
	}

	// 渲染左侧菜单
	leftContent := ct.renderLeftMenu()

	// 渲染右侧内容
	rightContent := ct.renderRightContent(rightWidth - 2)

	// 横向组合
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftStyle.Render(leftContent),
		rightStyle.Render(rightContent),
	)
}

// renderLeftMenu 渲染左侧菜单
func (ct *ConfigTab) renderLeftMenu() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Padding(0, 0, 1, 0)

	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(0, 1)

	normalStyle := lipgloss.NewStyle().
		Padding(0, 1)

	var content string
	content += titleStyle.Render("📁 配置类型") + "\n"

	for i, item := range ct.menuItems {
		style := normalStyle
		prefix := "  "
		if i == ct.selectedItem {
			style = selectedStyle
			prefix = "▶ "
		}
		content += fmt.Sprintf("%s%s\n", prefix, style.Render(item))
	}

	content += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("当前配置文件:") + "\n"

	// 显示配置文件路径（完整路径）
	if _, err := os.Stat(ct.serverConfigPath); err == nil {
		content += fmt.Sprintf("📄 服务端: %s\n", ct.serverConfigPath)
	} else {
		content += fmt.Sprintf("❌ 服务端: %s (不存在)\n", ct.serverConfigPath)
	}

	if _, err := os.Stat(ct.clientConfigPath); err == nil {
		content += fmt.Sprintf("📄 客户端: %s\n", ct.clientConfigPath)
	} else {
		content += fmt.Sprintf("❌ 客户端: %s (不存在)\n", ct.clientConfigPath)
	}

	content += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("配置状态:") + "\n"

	// 显示配置状态
	if ct.serverConfig != nil {
		content += fmt.Sprintf("✓ 服务端: 端口 %d\n", ct.serverConfig.BindPort)
	} else {
		content += "✗ 服务端: 未加载\n"
	}

	if ct.clientConfig != nil {
		content += fmt.Sprintf("✓ 客户端: %s:%d\n", ct.clientConfig.ServerAddr, ct.clientConfig.ServerPort)
		if len(ct.clientConfig.Proxies) > 0 {
			content += fmt.Sprintf("  └ 代理: %d个\n", len(ct.clientConfig.Proxies))
		}
	} else {
		content += "✗ 客户端: 未加载\n"
	}

	content += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("操作提示:") + "\n"
	content += "↑/↓ 选择菜单\n"
	content += "Enter 确认选择\n"
	content += "Tab 激活表单\n"
	content += "ESC 退出表单"

	return content
}

// renderRightContent 渲染右侧内容
func (ct *ConfigTab) renderRightContent(width int) string {
	if ct.currentForm != nil {
		// 显示表单
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Padding(0, 0, 1, 0)

		var title string
		switch ct.state {
		case ConfigTabServerForm:
			title = "🎯 服务端"
		case ConfigTabClientForm:
			title = "💻 客户端"
		case ConfigTabProxyForm:
			title = "🔗 代理"
		case ConfigTabVisitorForm:
			title = "👥 访问者"
		case ConfigTabPreview:
			title = "👁️ 配置预览"
		}

		content := titleStyle.Render(title) + "\n\n"

		if ct.state == ConfigTabPreview {
			// 显示配置预览
			content += ct.renderConfigPreview()
		} else {
			// 显示表单
			content += ct.currentForm.View()

			// 添加表单操作提示
			if ct.focusOnForm {
				content += "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("表单操作: Tab/Shift+Tab 切换字段 | ESC 退出编辑 | Ctrl+Tab 回到菜单")
			} else {
				content += "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("按 Tab 键激活表单编辑")
			}
		}

		return content
	}

	// 显示欢迎信息
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Padding(0, 0, 1, 0)

	content := titleStyle.Render("📋 FRP 配置管理") + "\n\n"

	// 显示当前配置状态
	content += lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Render("📊 配置状态") + "\n\n"

	if ct.serverConfig != nil {
		content += fmt.Sprintf("✓ 服务端: 端口 %d", ct.serverConfig.BindPort)
		if ct.serverConfig.Token != "" {
			content += " (已设置认证)"
		}
		content += "\n"
	} else {
		content += "○ 服务端: 未配置\n"
	}

	if ct.clientConfig != nil {
		content += fmt.Sprintf("✓ 客户端: %s:%d", ct.clientConfig.ServerAddr, ct.clientConfig.ServerPort)
		if len(ct.clientConfig.Proxies) > 0 {
			content += fmt.Sprintf(" (%d个代理)", len(ct.clientConfig.Proxies))
		}
		content += "\n"
	} else {
		content += "○ 客户端: 未配置\n"
	}

	content += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Render("📚 功能说明") + "\n\n"
	content += "• 🎯 服务端配置: 配置FRP服务端参数\n"
	content += "• 💻 客户端配置: 配置客户端连接信息\n"
	content += "• 🔗 添加代理: 添加端口转发规则\n"
	content += "• 👥 添加访问者: 添加P2P连接配置\n"
	content += "• 📁 选择配置文件: 选择不同的配置文件\n"
	content += "• 👀 预览配置: 查看当前配置的YAML内容\n"
	content += "• 💾 保存配置: 保存当前配置到文件\n\n"

	content += lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true).Render("💡 操作提示") + "\n\n"
	content += "• 修改配置后需要手动保存\n"
	content += "• 代理配置属于客户端配置的一部分\n"
	content += "• 可以同时配置多个代理规则"

	return content
}

// renderConfigPreview 渲染配置预览
func (ct *ConfigTab) renderConfigPreview() string {
	var content string

	// 服务端配置预览
	content += lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("46")).Render("🎯 服务端配置文件内容:") + "\n\n"

	if ct.serverConfig != nil {
		data, err := yaml.Marshal(ct.serverConfig)
		if err == nil {
			content += lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")).
				Padding(1).
				Render(string(data)) + "\n\n"
		} else {
			content += lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("错误: "+err.Error()) + "\n\n"
		}
	} else {
		content += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("服务端配置为空") + "\n\n"
	}

	// 客户端配置预览
	content += lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81")).Render("💻 客户端配置文件内容:") + "\n\n"

	if ct.clientConfig != nil {
		data, err := yaml.Marshal(ct.clientConfig)
		if err == nil {
			content += lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")).
				Padding(1).
				Render(string(data)) + "\n\n"
		} else {
			content += lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("错误: "+err.Error()) + "\n\n"
		}
	} else {
		content += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("客户端配置为空") + "\n\n"
	}

	content += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("按 ESC 返回菜单")

	return content
}
