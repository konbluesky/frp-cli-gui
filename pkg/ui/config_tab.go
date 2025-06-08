package ui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
)

// ConfigTab 配置管理标签页
type ConfigTab struct {
	BaseTab
	state          ConfigTabState
	currentForm    *ConfigFormModel
	serverConfig   *config.Config
	clientConfig   *config.Config
	currentProxy   *config.ProxyConfig
	currentVisitor *config.VisitorConfig
	loader         *config.Loader
	menuItems      []string
	selectedItem   int
}

// NewConfigTab 创建配置管理标签页
func NewConfigTab() *ConfigTab {
	baseTab := NewBaseTab("配置管理")
	baseTab.focusable = true

	return &ConfigTab{
		BaseTab:      baseTab,
		state:        ConfigTabMenu,
		menuItems:    []string{"服务端配置", "客户端配置", "添加代理", "添加访问者", "加载配置文件", "保存配置文件"},
		selectedItem: 0,
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

		if ct.state != ConfigTabMenu && ct.currentForm != nil {
			// 表单模式 - 优先处理表单事件
			switch msg.String() {
			case "esc":
				// ESC 始终用于退出表单，不管表单是否完成
				ct.state = ConfigTabMenu
				ct.currentForm = nil
				return ct, nil
			}
			// 其他所有键盘事件都传递给表单处理，包括 Tab/Shift+Tab
			form, cmd := ct.currentForm.Update(msg)
			if f, ok := form.(*ConfigFormModel); ok {
				ct.currentForm = f
			}
			return ct, cmd
		}

		// 菜单模式
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

	default:
		// 表单模式下，将所有其他消息（如huh.nextFieldMsg等）传递给表单处理
		if ct.state != ConfigTabMenu && ct.currentForm != nil {
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
	case 0: // 服务端配置
		if ct.serverConfig == nil {
			ct.serverConfig = config.CreateDefaultServerConfig()
		}
		ct.currentForm = NewServerConfigForm(ct.serverConfig)
		ct.state = ConfigTabServerForm
		return ct, ct.currentForm.Init()

	case 1: // 客户端配置
		if ct.clientConfig == nil {
			ct.clientConfig = config.CreateDefaultClientConfig()
		}
		ct.currentForm = NewClientConfigForm(ct.clientConfig)
		ct.state = ConfigTabClientForm
		return ct, ct.currentForm.Init()

	case 2: // 添加代理
		ct.currentProxy = &config.ProxyConfig{
			Type:    "tcp",
			LocalIP: "127.0.0.1",
		}
		ct.currentForm = NewProxyConfigForm(ct.currentProxy)
		ct.state = ConfigTabProxyForm
		return ct, ct.currentForm.Init()

	case 3: // 添加访问者
		ct.currentVisitor = &config.VisitorConfig{
			Type:     "stcp",
			BindAddr: "127.0.0.1",
		}
		ct.currentForm = NewVisitorConfigForm(ct.currentVisitor)
		ct.state = ConfigTabVisitorForm
		return ct, ct.currentForm.Init()

	case 4: // 加载配置文件
		return ct.loadConfigFile()

	case 5: // 保存配置文件
		return ct.saveConfigFile()
	}

	return ct, nil
}

// loadConfigFile 加载配置文件
func (ct *ConfigTab) loadConfigFile() (Tab, tea.Cmd) {
	// 这里可以实现文件选择逻辑，暂时使用示例文件
	serverPath := "examples/frps.yaml"
	clientPath := "examples/frpc.yaml"

	if _, err := os.Stat(serverPath); err == nil {
		loader := config.NewLoader(serverPath)
		if cfg, err := loader.Load(); err == nil {
			ct.serverConfig = cfg
		}
	}

	if _, err := os.Stat(clientPath); err == nil {
		loader := config.NewLoader(clientPath)
		if cfg, err := loader.Load(); err == nil {
			ct.clientConfig = cfg
		}
	}

	return ct, nil
}

// saveConfigFile 保存配置文件
func (ct *ConfigTab) saveConfigFile() (Tab, tea.Cmd) {
	// 创建配置目录
	configDir := "configs"
	os.MkdirAll(configDir, 0755)

	// 保存服务端配置
	if ct.serverConfig != nil {
		loader := config.NewLoader(filepath.Join(configDir, "frps.yaml"))
		loader.Save(ct.serverConfig)
	}

	// 保存客户端配置
	if ct.clientConfig != nil {
		loader := config.NewLoader(filepath.Join(configDir, "frpc.yaml"))
		loader.Save(ct.clientConfig)
	}

	return ct, nil
}

// IsInFormMode 检查是否处于表单编辑模式
func (ct *ConfigTab) IsInFormMode() bool {
	return ct.state != ConfigTabMenu
}

// View 渲染视图
func (ct *ConfigTab) View(width int, height int) string {
	if ct.state != ConfigTabMenu && ct.currentForm != nil {
		// 表单视图
		return ct.currentForm.View()
	}

	// 菜单视图
	contentWidth := width - 12
	if contentWidth < 40 {
		contentWidth = 40
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Padding(1, 0)

	menuStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(contentWidth)

	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(0, 1)

	normalStyle := lipgloss.NewStyle().
		Padding(0, 1)

	var content string
	content += titleStyle.Render("📁 FRP 配置管理") + "\n\n"

	content += "请选择配置类型:\n\n"

	for i, item := range ct.menuItems {
		style := normalStyle
		prefix := "  "
		if i == ct.selectedItem {
			style = selectedStyle
			prefix = "▶ "
		}
		content += fmt.Sprintf("%s%s\n", prefix, style.Render(item))
	}

	content += "\n"

	// 显示当前配置状态
	content += "当前配置状态:\n"
	if ct.serverConfig != nil {
		content += fmt.Sprintf("✓ 服务端配置: 端口 %d\n", ct.serverConfig.BindPort)
	} else {
		content += "✗ 服务端配置: 未设置\n"
	}

	if ct.clientConfig != nil {
		content += fmt.Sprintf("✓ 客户端配置: %s:%d\n", ct.clientConfig.ServerAddr, ct.clientConfig.ServerPort)
		content += fmt.Sprintf("  └ 代理数量: %d\n", len(ct.clientConfig.Proxies))
	} else {
		content += "✗ 客户端配置: 未设置\n"
	}

	content += "\n操作提示:\n"
	content += "↑/↓ 选择菜单项\n"
	content += "Enter 确认选择\n"
	content += "ESC 返回菜单 (在表单中)"

	return menuStyle.Render(content)
}
