package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"frp-cli-ui/config"
)

// ConfigEditor 配置编辑器模型
type ConfigEditor struct {
	width  int
	height int
	mode   string // "server", "client", "template", "import"
	state  ConfigEditorState

	// 基本输入
	inputs     []textinput.Model
	textarea   textarea.Model
	focusIndex int

	// 代理管理
	proxyList    list.Model
	proxyEditor  ProxyEditor
	editingProxy bool

	// 配置管理
	config          *config.Config
	loader          *config.Loader
	validator       *config.Validator
	templateManager *config.TemplateManager

	// 模板和历史
	history []ConfigHistory

	// 状态信息
	message     string
	messageType MessageType
	showHelp    bool

	// 文件操作
	fileDialog FileDialog
	showDialog bool
}

// ConfigEditorState 编辑器状态
type ConfigEditorState int

const (
	StateBasicConfig ConfigEditorState = iota
	StateProxyList
	StateProxyEdit
	StateTemplates
	StateImportExport
	StateValidation
)

// MessageType 消息类型
type MessageType int

const (
	MessageInfo MessageType = iota
	MessageSuccess
	MessageWarning
	MessageError
)

// ConfigTemplate 配置模板
type ConfigTemplate struct {
	Name        string
	Description string
	Type        string // "server" or "client"
	Config      *config.Config
	CreatedAt   time.Time
}

// ConfigHistory 配置历史
type ConfigHistory struct {
	Timestamp time.Time
	Action    string
	Config    *config.Config
	Message   string
}

// ProxyEditor 代理编辑器
type ProxyEditor struct {
	inputs       []textinput.Model
	focusIndex   int
	proxyType    string
	isNew        bool
	originalName string
}

// FileDialog 文件对话框
type FileDialog struct {
	input       textinput.Model
	action      string // "save", "load"
	defaultPath string
}

// NewConfigEditor 创建新的配置编辑器
func NewConfigEditor(mode string) ConfigEditor {
	// 创建基本输入框
	inputs := createBasicInputs()

	// 创建代理列表
	proxyItems := []list.Item{}
	proxyList := list.New(proxyItems, list.NewDefaultDelegate(), 0, 0)
	proxyList.Title = "代理配置列表"
	proxyList.SetShowStatusBar(false)
	proxyList.SetFilteringEnabled(true)

	// 创建文本区域
	ta := textarea.New()
	ta.Placeholder = "配置预览将显示在这里..."
	ta.SetWidth(80)
	ta.SetHeight(15)

	// 创建配置加载器
	configPath := getDefaultConfigPath(mode)
	loader := config.NewLoader(configPath)
	validator := config.NewValidator()
	templateManager := config.NewTemplateManager()

	// 加载现有配置或创建默认配置
	cfg, err := loader.Load()
	if err != nil {
		if mode == "server" {
			cfg = config.CreateDefaultServerConfig()
		} else {
			cfg = config.CreateDefaultClientConfig()
		}
	}

	editor := ConfigEditor{
		mode:            mode,
		state:           StateBasicConfig,
		inputs:          inputs,
		textarea:        ta,
		focusIndex:      0,
		proxyList:       proxyList,
		proxyEditor:     NewProxyEditor(),
		config:          cfg,
		loader:          loader,
		validator:       validator,
		templateManager: templateManager,
		history:         []ConfigHistory{},
		messageType:     MessageInfo,
		fileDialog:      NewFileDialog(),
	}

	// 填充现有配置值
	editor.populateInputs()
	editor.updateProxyList()
	editor.generatePreview()

	return editor
}

// createBasicInputs 创建基本输入框
func createBasicInputs() []textinput.Model {
	inputs := make([]textinput.Model, 8)

	// 服务器地址
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "服务器地址 (例: frp.example.com)"
	inputs[0].Focus()
	inputs[0].CharLimit = 100
	inputs[0].Width = 40

	// 服务器端口
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "服务器端口 (例: 7000)"
	inputs[1].CharLimit = 5
	inputs[1].Width = 20

	// 认证令牌
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "认证令牌 (可选)"
	inputs[2].EchoMode = textinput.EchoPassword
	inputs[2].CharLimit = 100
	inputs[2].Width = 40

	// 绑定端口 (服务端)
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "绑定端口 (例: 7000)"
	inputs[3].CharLimit = 5
	inputs[3].Width = 20

	// Web 服务器端口
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "Web 端口 (例: 7500)"
	inputs[4].CharLimit = 5
	inputs[4].Width = 20

	// Web 用户名
	inputs[5] = textinput.New()
	inputs[5].Placeholder = "Web 用户名"
	inputs[5].CharLimit = 50
	inputs[5].Width = 30

	// Web 密码
	inputs[6] = textinput.New()
	inputs[6].Placeholder = "Web 密码"
	inputs[6].EchoMode = textinput.EchoPassword
	inputs[6].CharLimit = 50
	inputs[6].Width = 30

	// 日志级别
	inputs[7] = textinput.New()
	inputs[7].Placeholder = "日志级别 (info/debug/warn/error)"
	inputs[7].CharLimit = 10
	inputs[7].Width = 25

	return inputs
}

// NewProxyEditor 创建新的代理编辑器
func NewProxyEditor() ProxyEditor {
	inputs := make([]textinput.Model, 8)

	// 代理名称
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "代理名称"
	inputs[0].Focus()
	inputs[0].CharLimit = 50
	inputs[0].Width = 30

	// 代理类型
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "类型 (tcp/udp/http/https/stcp/sudp/xtcp)"
	inputs[1].CharLimit = 10
	inputs[1].Width = 35

	// 本地 IP
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "本地 IP (例: 127.0.0.1)"
	inputs[2].CharLimit = 15
	inputs[2].Width = 25

	// 本地端口
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "本地端口"
	inputs[3].CharLimit = 5
	inputs[3].Width = 15

	// 远程端口
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "远程端口"
	inputs[4].CharLimit = 5
	inputs[4].Width = 15

	// 自定义域名
	inputs[5] = textinput.New()
	inputs[5].Placeholder = "自定义域名 (用逗号分隔)"
	inputs[5].CharLimit = 200
	inputs[5].Width = 50

	// 子域名
	inputs[6] = textinput.New()
	inputs[6].Placeholder = "子域名"
	inputs[6].CharLimit = 50
	inputs[6].Width = 30

	// 密钥 (STCP/SUDP/XTCP)
	inputs[7] = textinput.New()
	inputs[7].Placeholder = "密钥 (STCP/SUDP/XTCP)"
	inputs[7].EchoMode = textinput.EchoPassword
	inputs[7].CharLimit = 100
	inputs[7].Width = 40

	return ProxyEditor{
		inputs:     inputs,
		focusIndex: 0,
		proxyType:  "tcp",
		isNew:      true,
	}
}

// NewFileDialog 创建新的文件对话框
func NewFileDialog() FileDialog {
	input := textinput.New()
	input.Placeholder = "输入文件路径..."
	input.CharLimit = 200
	input.Width = 60

	return FileDialog{
		input: input,
	}
}

// Init 初始化
func (m ConfigEditor) Init() tea.Cmd {
	return textinput.Blink
}

// Update 更新状态
func (m ConfigEditor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.proxyList.SetWidth(msg.Width - 4)
		m.proxyList.SetHeight(msg.Height - 10)

	case tea.KeyMsg:
		// 全局快捷键
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if !m.showDialog && !m.editingProxy {
				return m, tea.Quit
			}
		case "x":
			// x 键返回上级（如果不在对话框或编辑状态）
			if !m.showDialog && !m.editingProxy {
				// 这里应该返回到 Dashboard，但由于架构限制，我们使用 tea.Quit
				// 在实际使用中，这会被 Dashboard 的 Update 方法捕获
				return m, tea.Quit
			}
		case "ctrl+h":
			m.showHelp = !m.showHelp
			return m, nil
		}

		// 处理文件对话框
		if m.showDialog {
			return m.updateFileDialog(msg)
		}

		// 处理代理编辑
		if m.editingProxy {
			return m.updateProxyEditor(msg)
		}

		// 处理主界面
		return m.updateMainInterface(msg)
	}

	// 更新子组件
	if !m.showDialog && !m.editingProxy {
		switch m.state {
		case StateProxyList:
			m.proxyList, cmd = m.proxyList.Update(msg)
			cmds = append(cmds, cmd)
		default:
			cmd = m.updateInputs(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// updateMainInterface 更新主界面
func (m ConfigEditor) updateMainInterface(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		// 在同一状态内的控件间跳转
		if m.state == StateBasicConfig || m.state == StateImportExport || m.state == StateValidation {
			return m.navigateInputs("down")
		}
		// 否则切换状态
		return m.nextState(), nil
	case "shift+tab":
		// 在同一状态内的控件间反向跳转
		if m.state == StateBasicConfig || m.state == StateImportExport || m.state == StateValidation {
			return m.navigateInputs("up")
		}
		// 否则反向切换状态
		return m.prevState(), nil
	case "ctrl+tab":
		// Ctrl+Tab 专门用于状态切换
		return m.nextState(), nil
	case "ctrl+shift+tab":
		// Ctrl+Shift+Tab 专门用于反向状态切换
		return m.prevState(), nil
	case "ctrl+s":
		return m.saveConfig()
	case "ctrl+l":
		return m.loadConfig()
	case "ctrl+e":
		return m.exportConfig()
	case "ctrl+i":
		return m.importConfig()
	case "ctrl+t":
		return m.showTemplates()
	case "ctrl+v":
		return m.validateConfig()
	case "ctrl+r":
		return m.resetConfig()
	case "ctrl+z":
		return m.undoLastChange()
	case "ctrl+m":
		if m.state == StateTemplates {
			return m.mergeTemplate()
		}
	case "ctrl+n":
		if m.state == StateTemplates {
			return m.newTemplate()
		}
	case "ctrl+d":
		if m.state == StateTemplates {
			return m.deleteTemplate()
		}
	case "enter":
		if m.state == StateProxyList {
			return m.editSelectedProxy()
		}
	case "n":
		if m.state == StateProxyList {
			return m.newProxy()
		}
	case "d":
		if m.state == StateProxyList {
			return m.deleteSelectedProxy()
		}
	case "up", "down":
		if m.state != StateProxyList {
			return m.navigateInputs(msg.String())
		}
	case "left", "right":
		// 左右键也可以用于控件跳转
		if m.state != StateProxyList {
			direction := "up"
			if msg.String() == "right" {
				direction = "down"
			}
			return m.navigateInputs(direction)
		}
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		if m.state == StateTemplates {
			return m.applyTemplate(msg.String())
		}
	}

	return m, nil
}

// updateProxyEditor 更新代理编辑器
func (m ConfigEditor) updateProxyEditor(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "x":
		// Esc 或 x 键返回
		m.editingProxy = false
		return m, nil
	case "ctrl+s":
		return m.saveProxy()
	case "tab", "down", "right":
		// Tab、下键、右键：下一个输入框
		return m.navigateProxyInputs("down")
	case "shift+tab", "up", "left":
		// Shift+Tab、上键、左键：上一个输入框
		return m.navigateProxyInputs("up")
	}

	// 更新当前输入框
	if m.proxyEditor.focusIndex < len(m.proxyEditor.inputs) {
		var cmd tea.Cmd
		m.proxyEditor.inputs[m.proxyEditor.focusIndex], cmd = m.proxyEditor.inputs[m.proxyEditor.focusIndex].Update(msg)
		return m, cmd
	}

	return m, nil
}

// updateFileDialog 更新文件对话框
func (m ConfigEditor) updateFileDialog(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "x":
		// Esc 或 x 键取消对话框
		m.showDialog = false
		return m, nil
	case "enter":
		return m.executeFileAction()
	}

	var cmd tea.Cmd
	m.fileDialog.input, cmd = m.fileDialog.input.Update(msg)
	return m, cmd
}

// 状态切换方法
func (m ConfigEditor) nextState() ConfigEditor {
	switch m.state {
	case StateBasicConfig:
		if m.mode == "client" {
			m.state = StateProxyList
		} else {
			m.state = StateTemplates
		}
	case StateProxyList:
		m.state = StateTemplates
	case StateTemplates:
		m.state = StateImportExport
	case StateImportExport:
		m.state = StateValidation
	case StateValidation:
		m.state = StateBasicConfig
	}
	return m
}

func (m ConfigEditor) prevState() ConfigEditor {
	switch m.state {
	case StateBasicConfig:
		m.state = StateValidation
	case StateProxyList:
		m.state = StateBasicConfig
	case StateTemplates:
		if m.mode == "client" {
			m.state = StateProxyList
		} else {
			m.state = StateBasicConfig
		}
	case StateImportExport:
		m.state = StateTemplates
	case StateValidation:
		m.state = StateImportExport
	}
	return m
}

// 配置操作方法
func (m ConfigEditor) saveConfig() (ConfigEditor, tea.Cmd) {
	// 从输入框更新配置
	m.updateConfigFromInputs()

	// 验证配置
	if err := m.validator.ValidateConfig(m.config); err != nil {
		m.message = fmt.Sprintf("配置验证失败: %v", err)
		m.messageType = MessageError
		return m, nil
	}

	// 保存配置
	if err := m.loader.Save(m.config); err != nil {
		m.message = fmt.Sprintf("保存失败: %v", err)
		m.messageType = MessageError
		return m, nil
	}

	// 添加到历史记录
	m.addToHistory("保存配置", "配置已成功保存")

	m.message = "配置已保存"
	m.messageType = MessageSuccess
	m.generatePreview()

	return m, nil
}

func (m ConfigEditor) loadConfig() (ConfigEditor, tea.Cmd) {
	m.fileDialog.action = "load"
	m.fileDialog.defaultPath = m.loader.GetConfigPath()
	m.fileDialog.input.SetValue(m.fileDialog.defaultPath)
	m.fileDialog.input.Focus()
	m.showDialog = true
	return m, nil
}

func (m ConfigEditor) exportConfig() (ConfigEditor, tea.Cmd) {
	m.fileDialog.action = "export"
	m.fileDialog.defaultPath = fmt.Sprintf("frp_%s_config_%s.yaml", m.mode, time.Now().Format("20060102_150405"))
	m.fileDialog.input.SetValue(m.fileDialog.defaultPath)
	m.fileDialog.input.Focus()
	m.showDialog = true
	return m, nil
}

func (m ConfigEditor) importConfig() (ConfigEditor, tea.Cmd) {
	m.fileDialog.action = "import"
	m.fileDialog.defaultPath = ""
	m.fileDialog.input.SetValue("")
	m.fileDialog.input.Focus()
	m.showDialog = true
	return m, nil
}

func (m ConfigEditor) validateConfig() (ConfigEditor, tea.Cmd) {
	m.updateConfigFromInputs()

	err := m.validator.ValidateConfigDetailed(m.config)
	if len(err) == 0 {
		m.message = "✅ 配置验证通过"
		m.messageType = MessageSuccess
	} else {
		m.message = "❌ 验证失败"
		m.messageType = MessageError
	}

	m.state = StateValidation
	return m, nil
}

func (m ConfigEditor) resetConfig() (ConfigEditor, tea.Cmd) {
	if m.mode == "server" {
		m.config = config.CreateDefaultServerConfig()
	} else {
		m.config = config.CreateDefaultClientConfig()
	}

	m.populateInputs()
	m.updateProxyList()
	m.generatePreview()

	m.message = "配置已重置为默认值"
	m.messageType = MessageInfo

	return m, nil
}

// 代理管理方法
func (m ConfigEditor) newProxy() (ConfigEditor, tea.Cmd) {
	m.proxyEditor = NewProxyEditor()
	m.proxyEditor.isNew = true
	m.editingProxy = true
	return m, nil
}

func (m ConfigEditor) editSelectedProxy() (ConfigEditor, tea.Cmd) {
	if item := m.proxyList.SelectedItem(); item != nil {
		if proxyItem, ok := item.(ProxyListItem); ok {
			m.proxyEditor = m.createProxyEditorFromConfig(proxyItem.proxy)
			m.proxyEditor.isNew = false
			m.proxyEditor.originalName = proxyItem.proxy.Name
			m.editingProxy = true
		}
	}
	return m, nil
}

func (m ConfigEditor) deleteSelectedProxy() (ConfigEditor, tea.Cmd) {
	if item := m.proxyList.SelectedItem(); item != nil {
		if proxyItem, ok := item.(ProxyListItem); ok {
			// 从配置中删除代理
			newProxies := []config.ProxyConfig{}
			for _, proxy := range m.config.Proxies {
				if proxy.Name != proxyItem.proxy.Name {
					newProxies = append(newProxies, proxy)
				}
			}
			m.config.Proxies = newProxies

			m.updateProxyList()
			m.generatePreview()

			m.message = fmt.Sprintf("已删除代理: %s", proxyItem.proxy.Name)
			m.messageType = MessageSuccess
		}
	}
	return m, nil
}

func (m ConfigEditor) saveProxy() (ConfigEditor, tea.Cmd) {
	// 从输入框创建代理配置
	proxy := m.createProxyFromEditor()

	// 验证代理配置
	if err := m.validator.ValidateProxyConfig(proxy); err != nil {
		m.message = fmt.Sprintf("代理验证失败: %v", err)
		m.messageType = MessageError
		return m, nil
	}

	if m.proxyEditor.isNew {
		// 添加新代理
		m.config.Proxies = append(m.config.Proxies, proxy)
		m.message = fmt.Sprintf("已添加代理: %s", proxy.Name)
	} else {
		// 更新现有代理
		for i, p := range m.config.Proxies {
			if p.Name == m.proxyEditor.originalName {
				m.config.Proxies[i] = proxy
				break
			}
		}
		m.message = fmt.Sprintf("已更新代理: %s", proxy.Name)
	}

	m.messageType = MessageSuccess
	m.editingProxy = false
	m.updateProxyList()
	m.generatePreview()

	return m, nil
}

// 辅助方法
func (m ConfigEditor) populateInputs() {
	if m.config == nil {
		return
	}

	m.inputs[0].SetValue(m.config.ServerAddr)
	m.inputs[1].SetValue(strconv.Itoa(m.config.ServerPort))
	m.inputs[2].SetValue(m.config.Token)
	m.inputs[3].SetValue(strconv.Itoa(m.config.BindPort))
	m.inputs[4].SetValue(strconv.Itoa(m.config.WebServer.Port))
	m.inputs[5].SetValue(m.config.WebServer.User)
	m.inputs[6].SetValue(m.config.WebServer.Password)
	m.inputs[7].SetValue(m.config.Log.Level)
}

func (m ConfigEditor) updateConfigFromInputs() {
	if m.config == nil {
		if m.mode == "server" {
			m.config = config.CreateDefaultServerConfig()
		} else {
			m.config = config.CreateDefaultClientConfig()
		}
	}

	m.config.ServerAddr = m.inputs[0].Value()
	if port, err := strconv.Atoi(m.inputs[1].Value()); err == nil {
		m.config.ServerPort = port
	}
	m.config.Token = m.inputs[2].Value()
	if bindPort, err := strconv.Atoi(m.inputs[3].Value()); err == nil {
		m.config.BindPort = bindPort
	}
	if webPort, err := strconv.Atoi(m.inputs[4].Value()); err == nil {
		m.config.WebServer.Port = webPort
	}
	m.config.WebServer.User = m.inputs[5].Value()
	m.config.WebServer.Password = m.inputs[6].Value()
	m.config.Log.Level = m.inputs[7].Value()
}

func (m ConfigEditor) generatePreview() {
	if m.config == nil {
		return
	}

	var preview strings.Builder

	if m.mode == "server" {
		preview.WriteString("# FRP 服务端配置\n")
		if m.config.BindPort > 0 {
			preview.WriteString(fmt.Sprintf("bindPort = %d\n", m.config.BindPort))
		}
		if m.config.Token != "" {
			preview.WriteString(fmt.Sprintf("token = \"%s\"\n", m.config.Token))
		}
		if m.config.WebServer.Port > 0 {
			preview.WriteString(fmt.Sprintf("\n[webServer]\n"))
			preview.WriteString(fmt.Sprintf("port = %d\n", m.config.WebServer.Port))
			if m.config.WebServer.User != "" {
				preview.WriteString(fmt.Sprintf("user = \"%s\"\n", m.config.WebServer.User))
			}
			if m.config.WebServer.Password != "" {
				preview.WriteString(fmt.Sprintf("password = \"%s\"\n", m.config.WebServer.Password))
			}
		}
	} else {
		preview.WriteString("# FRP 客户端配置\n")
		if m.config.ServerAddr != "" {
			preview.WriteString(fmt.Sprintf("serverAddr = \"%s\"\n", m.config.ServerAddr))
		}
		if m.config.ServerPort > 0 {
			preview.WriteString(fmt.Sprintf("serverPort = %d\n", m.config.ServerPort))
		}
		if m.config.Token != "" {
			preview.WriteString(fmt.Sprintf("token = \"%s\"\n", m.config.Token))
		}

		// 添加代理配置
		for _, proxy := range m.config.Proxies {
			preview.WriteString(fmt.Sprintf("\n[[proxies]]\n"))
			preview.WriteString(fmt.Sprintf("name = \"%s\"\n", proxy.Name))
			preview.WriteString(fmt.Sprintf("type = \"%s\"\n", proxy.Type))
			if proxy.LocalIP != "" {
				preview.WriteString(fmt.Sprintf("localIP = \"%s\"\n", proxy.LocalIP))
			}
			if proxy.LocalPort > 0 {
				preview.WriteString(fmt.Sprintf("localPort = %d\n", proxy.LocalPort))
			}
			if proxy.RemotePort > 0 {
				preview.WriteString(fmt.Sprintf("remotePort = %d\n", proxy.RemotePort))
			}
			if len(proxy.CustomDomains) > 0 {
				preview.WriteString(fmt.Sprintf("customDomains = %v\n", proxy.CustomDomains))
			}
		}
	}

	if m.config.Log.Level != "" {
		preview.WriteString(fmt.Sprintf("\n[log]\n"))
		preview.WriteString(fmt.Sprintf("level = \"%s\"\n", m.config.Log.Level))
		preview.WriteString("to = \"console\"\n")
	}

	m.textarea.SetValue(preview.String())
}

// 更多辅助方法...
func (m ConfigEditor) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	// 实时更新预览
	m.updateConfigFromInputs()
	m.generatePreview()

	return tea.Batch(cmds...)
}

func (m ConfigEditor) navigateInputs(direction string) (ConfigEditor, tea.Cmd) {
	if direction == "up" {
		m.focusIndex--
		if m.focusIndex < 0 {
			m.focusIndex = len(m.inputs) - 1
		}
	} else {
		m.focusIndex++
		if m.focusIndex >= len(m.inputs) {
			m.focusIndex = 0
		}
	}

	// 更新焦点
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}

	return m, nil
}

// ProxyListItem 代理列表项
type ProxyListItem struct {
	proxy config.ProxyConfig
}

func (i ProxyListItem) FilterValue() string { return i.proxy.Name }
func (i ProxyListItem) Title() string       { return i.proxy.Name }
func (i ProxyListItem) Description() string {
	return fmt.Sprintf("%s: %s:%d -> %d", i.proxy.Type, i.proxy.LocalIP, i.proxy.LocalPort, i.proxy.RemotePort)
}

func (m ConfigEditor) updateProxyList() {
	items := []list.Item{}
	for _, proxy := range m.config.Proxies {
		items = append(items, ProxyListItem{proxy: proxy})
	}
	m.proxyList.SetItems(items)
}

// 获取默认配置路径
func getDefaultConfigPath(mode string) string {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".frp")
	os.MkdirAll(configDir, 0755)

	if mode == "server" {
		return filepath.Join(configDir, "frps.yaml")
	}
	return filepath.Join(configDir, "frpc.yaml")
}

// 加载配置模板
func loadConfigTemplates() []ConfigTemplate {
	// 使用模板管理器获取模板
	tm := config.NewTemplateManager()
	templates := tm.GetTemplates()

	// 转换为 TUI 模板格式
	var tuiTemplates []ConfigTemplate
	for _, template := range templates {
		tuiTemplates = append(tuiTemplates, ConfigTemplate{
			Name:        template.Name,
			Description: template.Description,
			Type:        template.Type,
			Config:      template.Config,
			CreatedAt:   template.CreatedAt,
		})
	}

	return tuiTemplates
}

// 添加到历史记录
func (m ConfigEditor) addToHistory(action, message string) {
	history := ConfigHistory{
		Timestamp: time.Now(),
		Action:    action,
		Config:    m.config,
		Message:   message,
	}

	m.history = append(m.history, history)

	// 保持历史记录数量限制
	if len(m.history) > 50 {
		m.history = m.history[1:]
	}
}

// 其他必要的辅助方法...
func (m ConfigEditor) createProxyEditorFromConfig(proxy config.ProxyConfig) ProxyEditor {
	editor := NewProxyEditor()
	editor.inputs[0].SetValue(proxy.Name)
	editor.inputs[1].SetValue(proxy.Type)
	editor.inputs[2].SetValue(proxy.LocalIP)
	editor.inputs[3].SetValue(strconv.Itoa(proxy.LocalPort))
	editor.inputs[4].SetValue(strconv.Itoa(proxy.RemotePort))
	if len(proxy.CustomDomains) > 0 {
		editor.inputs[5].SetValue(strings.Join(proxy.CustomDomains, ","))
	}
	editor.inputs[6].SetValue(proxy.Subdomain)
	editor.inputs[7].SetValue(proxy.SecretKey)
	return editor
}

func (m ConfigEditor) createProxyFromEditor() config.ProxyConfig {
	proxy := config.ProxyConfig{
		Name:    m.proxyEditor.inputs[0].Value(),
		Type:    m.proxyEditor.inputs[1].Value(),
		LocalIP: m.proxyEditor.inputs[2].Value(),
	}

	if localPort, err := strconv.Atoi(m.proxyEditor.inputs[3].Value()); err == nil {
		proxy.LocalPort = localPort
	}
	if remotePort, err := strconv.Atoi(m.proxyEditor.inputs[4].Value()); err == nil {
		proxy.RemotePort = remotePort
	}

	if domains := m.proxyEditor.inputs[5].Value(); domains != "" {
		proxy.CustomDomains = strings.Split(domains, ",")
		for i := range proxy.CustomDomains {
			proxy.CustomDomains[i] = strings.TrimSpace(proxy.CustomDomains[i])
		}
	}

	proxy.Subdomain = m.proxyEditor.inputs[6].Value()
	proxy.SecretKey = m.proxyEditor.inputs[7].Value()

	return proxy
}

// 继续实现其他方法...
func (m ConfigEditor) navigateProxyInputs(direction string) (ConfigEditor, tea.Cmd) {
	if direction == "up" {
		m.proxyEditor.focusIndex--
		if m.proxyEditor.focusIndex < 0 {
			m.proxyEditor.focusIndex = len(m.proxyEditor.inputs) - 1
		}
	} else {
		m.proxyEditor.focusIndex++
		if m.proxyEditor.focusIndex >= len(m.proxyEditor.inputs) {
			m.proxyEditor.focusIndex = 0
		}
	}

	// 更新焦点
	for i := range m.proxyEditor.inputs {
		if i == m.proxyEditor.focusIndex {
			m.proxyEditor.inputs[i].Focus()
		} else {
			m.proxyEditor.inputs[i].Blur()
		}
	}

	return m, nil
}

// executeFileAction 执行文件操作
func (m ConfigEditor) executeFileAction() (ConfigEditor, tea.Cmd) {
	path := m.fileDialog.input.Value()
	if path == "" {
		m.message = "请输入文件路径"
		m.messageType = MessageError
		return m, nil
	}

	switch m.fileDialog.action {
	case "load":
		if cfg, err := m.loader.ImportFromFile(path); err != nil {
			m.message = fmt.Sprintf("加载失败: %v", err)
			m.messageType = MessageError
		} else {
			m.config = cfg
			m.loader.SetConfigPath(path)
			m.populateInputs()
			m.updateProxyList()
			m.generatePreview()
			m.message = "配置加载成功"
			m.messageType = MessageSuccess
			m.addToHistory("加载配置", fmt.Sprintf("从 %s 加载配置", path))
		}

	case "export":
		// 确保路径有正确的扩展名
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			path += ".yaml"
		}

		m.updateConfigFromInputs()
		if err := m.loader.ExportToFile(m.config, path); err != nil {
			m.message = fmt.Sprintf("导出失败: %v", err)
			m.messageType = MessageError
		} else {
			m.message = fmt.Sprintf("配置已导出到: %s", path)
			m.messageType = MessageSuccess
			m.addToHistory("导出配置", fmt.Sprintf("导出配置到 %s", path))
		}

	case "import":
		// 验证文件格式
		if err := m.loader.ValidateConfigFile(path); err != nil {
			m.message = fmt.Sprintf("文件验证失败: %v", err)
			m.messageType = MessageError
			break
		}

		if cfg, err := m.loader.ImportFromFile(path); err != nil {
			m.message = fmt.Sprintf("导入失败: %v", err)
			m.messageType = MessageError
		} else {
			// 合并配置而不是替换
			m.config = m.loader.MergeConfig(m.config, cfg)
			m.populateInputs()
			m.updateProxyList()
			m.generatePreview()
			m.message = fmt.Sprintf("配置已从 %s 导入并合并", path)
			m.messageType = MessageSuccess
			m.addToHistory("导入配置", fmt.Sprintf("从 %s 导入并合并配置", path))
		}
	}

	m.showDialog = false
	return m, nil
}

func (m ConfigEditor) showTemplates() (ConfigEditor, tea.Cmd) {
	m.state = StateTemplates
	return m, nil
}

func (m ConfigEditor) undoLastChange() (ConfigEditor, tea.Cmd) {
	if len(m.history) > 0 {
		lastHistory := m.history[len(m.history)-1]
		m.config = lastHistory.Config
		m.history = m.history[:len(m.history)-1]

		m.populateInputs()
		m.updateProxyList()
		m.generatePreview()

		m.message = "已撤销上次更改"
		m.messageType = MessageInfo
	} else {
		m.message = "没有可撤销的更改"
		m.messageType = MessageWarning
	}

	return m, nil
}

// View 渲染视图
func (m ConfigEditor) View() string {
	var b strings.Builder

	// 处理文件对话框
	if m.showDialog {
		return m.renderFileDialog()
	}

	// 处理代理编辑
	if m.editingProxy {
		return m.renderProxyEditor()
	}

	// 主界面标题
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width)

	title := fmt.Sprintf("FRP %s配置管理", map[string]string{"server": "服务端", "client": "客户端"}[m.mode])
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	// 状态标签
	stateStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Bold(true)

	stateNames := map[ConfigEditorState]string{
		StateBasicConfig:  "基本配置",
		StateProxyList:    "代理管理",
		StateTemplates:    "配置模板",
		StateImportExport: "导入导出",
		StateValidation:   "配置验证",
	}

	b.WriteString(stateStyle.Render(fmt.Sprintf("当前: %s", stateNames[m.state])))
	b.WriteString("\n\n")

	// 根据状态渲染不同内容
	switch m.state {
	case StateBasicConfig:
		b.WriteString(m.renderBasicConfig())
	case StateProxyList:
		b.WriteString(m.renderProxyList())
	case StateTemplates:
		b.WriteString(m.renderTemplates())
	case StateImportExport:
		b.WriteString(m.renderImportExport())
	case StateValidation:
		b.WriteString(m.renderValidation())
	}

	// 消息显示
	if m.message != "" {
		b.WriteString("\n")
		b.WriteString(m.renderMessage())
	}

	// 帮助信息
	b.WriteString("\n")
	b.WriteString(m.renderHelp())

	return b.String()
}

// renderBasicConfig 渲染基本配置界面
func (m ConfigEditor) renderBasicConfig() string {
	var b strings.Builder

	// 输入表单
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(50)

	focusedStyle := inputStyle.Copy().
		BorderForeground(lipgloss.Color("57"))

	labels := []string{
		"服务器地址:",
		"服务器端口:",
		"认证令牌:",
		"绑定端口:",
		"Web 端口:",
		"Web 用户名:",
		"Web 密码:",
		"日志级别:",
	}

	// 根据模式显示不同的字段
	visibleFields := []int{}
	if m.mode == "server" {
		visibleFields = []int{2, 3, 4, 5, 6, 7} // 服务端字段
	} else {
		visibleFields = []int{0, 1, 2, 7} // 客户端字段
	}

	for _, i := range visibleFields {
		if i >= len(m.inputs) || i >= len(labels) {
			continue
		}

		b.WriteString(labels[i])
		b.WriteString("\n")

		if i == m.focusIndex {
			b.WriteString(focusedStyle.Render(m.inputs[i].View()))
		} else {
			b.WriteString(inputStyle.Render(m.inputs[i].View()))
		}
		b.WriteString("\n\n")
	}

	// 配置预览
	previewStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(min(80, m.width-4)).
		Height(15)

	b.WriteString("配置预览:")
	b.WriteString("\n")
	b.WriteString(previewStyle.Render(m.textarea.View()))

	return b.String()
}

// renderProxyList 渲染代理列表界面
func (m ConfigEditor) renderProxyList() string {
	var b strings.Builder

	// 代理列表标题
	listTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#04B575"))

	b.WriteString(listTitleStyle.Render("代理配置列表"))
	b.WriteString(fmt.Sprintf(" (共 %d 个)", len(m.config.Proxies)))
	b.WriteString("\n\n")

	// 代理列表
	if len(m.config.Proxies) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		b.WriteString(emptyStyle.Render("暂无代理配置，按 'n' 添加新代理"))
	} else {
		b.WriteString(m.proxyList.View())
	}

	return b.String()
}

// renderTemplates 渲染模板界面
func (m ConfigEditor) renderTemplates() string {
	var b strings.Builder

	templateStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(70)

	b.WriteString("配置模板\n\n")

	templates := m.templateManager.GetTemplatesByType(m.mode)

	if len(templates) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		b.WriteString(emptyStyle.Render("暂无可用模板"))
		return b.String()
	}

	for i, template := range templates {
		content := fmt.Sprintf("%d. %s\n   %s\n   创建时间: %s\n   按数字键应用模板",
			i+1,
			template.Name,
			template.Description,
			template.CreatedAt.Format("2006-01-02 15:04"))

		// 这里可以添加选择逻辑
		b.WriteString(templateStyle.Render(content))
		b.WriteString("\n\n")
	}

	// 添加操作说明
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	help := "操作说明:\n• 1-9: 应用对应模板\n• Ctrl+M: 合并模板\n• Ctrl+N: 新建模板\n• Ctrl+D: 删除模板"
	b.WriteString(helpStyle.Render(help))

	return b.String()
}

// renderImportExport 渲染导入导出界面
func (m ConfigEditor) renderImportExport() string {
	var b strings.Builder

	optionStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(50)

	b.WriteString("导入导出选项\n\n")

	options := []string{
		"Ctrl+L: 加载配置文件",
		"Ctrl+E: 导出当前配置",
		"Ctrl+I: 导入配置文件",
		"Ctrl+S: 保存当前配置",
	}

	for _, option := range options {
		b.WriteString(optionStyle.Render(option))
		b.WriteString("\n\n")
	}

	return b.String()
}

// renderValidation 渲染验证界面
func (m ConfigEditor) renderValidation() string {
	var b strings.Builder

	// 验证配置
	m.updateConfigFromInputs()
	errors := m.validator.ValidateConfigDetailed(m.config)

	validationStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Width(min(80, m.width-4))

	if len(errors) == 0 {
		successStyle := validationStyle.Copy().
			BorderForeground(lipgloss.Color("46"))

		b.WriteString(successStyle.Render("✅ 配置验证通过\n\n所有配置项都符合要求，可以安全使用。"))
	} else {
		errorStyle := validationStyle.Copy().
			BorderForeground(lipgloss.Color("196"))

		content := "❌ 配置验证失败\n\n发现以下问题:\n"
		for i, err := range errors {
			content += fmt.Sprintf("%d. %s\n", i+1, err)
		}

		b.WriteString(errorStyle.Render(content))
	}

	return b.String()
}

// renderProxyEditor 渲染代理编辑界面
func (m ConfigEditor) renderProxyEditor() string {
	var b strings.Builder

	// 标题
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width)

	title := "新建代理"
	if !m.proxyEditor.isNew {
		title = fmt.Sprintf("编辑代理: %s", m.proxyEditor.originalName)
	}

	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	// 输入表单
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(50)

	focusedStyle := inputStyle.Copy().
		BorderForeground(lipgloss.Color("57"))

	labels := []string{
		"代理名称:",
		"代理类型:",
		"本地 IP:",
		"本地端口:",
		"远程端口:",
		"自定义域名:",
		"子域名:",
		"密钥:",
	}

	for i := range m.proxyEditor.inputs {
		b.WriteString(labels[i])
		b.WriteString("\n")

		if i == m.proxyEditor.focusIndex {
			b.WriteString(focusedStyle.Render(m.proxyEditor.inputs[i].View()))
		} else {
			b.WriteString(inputStyle.Render(m.proxyEditor.inputs[i].View()))
		}
		b.WriteString("\n\n")
	}

	// 帮助信息
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	help := "Ctrl+S: 保存 | Tab/↑↓←→: 切换字段 | X/Esc: 返回"
	b.WriteString(helpStyle.Render(help))

	return b.String()
}

// renderFileDialog 渲染文件对话框
func (m ConfigEditor) renderFileDialog() string {
	var b strings.Builder

	// 对话框样式
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("57")).
		Padding(1).
		Width(70).
		Align(lipgloss.Center)

	// 标题
	titleMap := map[string]string{
		"load":   "加载配置文件",
		"export": "导出配置文件",
		"import": "导入配置文件",
	}

	title := titleMap[m.fileDialog.action]
	content := fmt.Sprintf("%s\n\n%s\n\n%s",
		title,
		m.fileDialog.input.View(),
		"Enter: 确认 | Esc: 取消")

	b.WriteString("\n\n")
	b.WriteString(dialogStyle.Render(content))

	return b.String()
}

// renderMessage 渲染消息
func (m ConfigEditor) renderMessage() string {
	var style lipgloss.Style

	switch m.messageType {
	case MessageSuccess:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)
	case MessageError:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
	case MessageWarning:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)
	default:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))
	}

	return style.Render(m.message)
}

// renderHelp 渲染帮助信息
func (m ConfigEditor) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	if m.showHelp {
		help := `
快捷键帮助:
• Tab/Shift+Tab: 控件间跳转
• Ctrl+Tab/Ctrl+Shift+Tab: 状态切换
• ↑↓←→: 控件间跳转
• Ctrl+S: 保存配置
• Ctrl+L: 加载配置
• Ctrl+E: 导出配置
• Ctrl+I: 导入配置
• Ctrl+V: 验证配置
• Ctrl+R: 重置配置
• Ctrl+Z: 撤销更改
• Ctrl+H: 显示/隐藏帮助
• X: 返回上级
• Q: 退出

代理管理 (客户端模式):
• N: 新建代理
• Enter: 编辑选中代理
• D: 删除选中代理
`
		return helpStyle.Render(help)
	}

	basicHelp := "Tab: 控件跳转 | Ctrl+Tab: 状态切换 | Ctrl+S: 保存 | Ctrl+H: 帮助 | X: 返回 | Q: 退出"
	return helpStyle.Render(basicHelp)
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 模板操作方法
func (m ConfigEditor) applyTemplate(templateIndex string) (ConfigEditor, tea.Cmd) {
	templates := m.templateManager.GetTemplatesByType(m.mode)

	index := int(templateIndex[0] - '1') // 转换为 0-based 索引
	if index < 0 || index >= len(templates) {
		m.message = "无效的模板编号"
		m.messageType = MessageError
		return m, nil
	}

	template := templates[index]

	// 应用模板
	if err := m.templateManager.ApplyTemplate(template.Name, m.config); err != nil {
		m.message = fmt.Sprintf("应用模板失败: %v", err)
		m.messageType = MessageError
		return m, nil
	}

	// 更新界面
	m.populateInputs()
	m.updateProxyList()
	m.generatePreview()

	m.message = fmt.Sprintf("已应用模板: %s", template.Name)
	m.messageType = MessageSuccess

	// 添加到历史记录
	m.addToHistory("应用模板", fmt.Sprintf("应用了模板: %s", template.Name))

	return m, nil
}

func (m ConfigEditor) mergeTemplate() (ConfigEditor, tea.Cmd) {
	// 这里可以实现模板合并逻辑
	// 为简化，暂时显示消息
	m.message = "模板合并功能开发中..."
	m.messageType = MessageInfo
	return m, nil
}

func (m ConfigEditor) newTemplate() (ConfigEditor, tea.Cmd) {
	// 这里可以实现新建模板逻辑
	// 为简化，暂时显示消息
	m.message = "新建模板功能开发中..."
	m.messageType = MessageInfo
	return m, nil
}

func (m ConfigEditor) deleteTemplate() (ConfigEditor, tea.Cmd) {
	// 这里可以实现删除模板逻辑
	// 为简化，暂时显示消息
	m.message = "删除模板功能开发中..."
	m.messageType = MessageInfo
	return m, nil
}
