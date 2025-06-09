package ui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"frp-cli-ui/pkg/config"
)

// ConfigFormType 配置表单类型
type ConfigFormType int

const (
	ServerConfigForm ConfigFormType = iota
	ClientConfigForm
	ProxyConfigForm
	VisitorConfigForm
)

// ConfigFormModel 配置表单模型
type ConfigFormModel struct {
	form          *huh.Form
	formType      ConfigFormType
	config        *config.Config
	proxyConfig   *config.ProxyConfig
	visitorConfig *config.VisitorConfig
	completed     bool
	err           error
	// 添加表单数据绑定字段
	formData map[string]*string
}

// NewServerConfigForm 创建服务端配置表单
func NewServerConfigForm(cfg *config.Config) *ConfigFormModel {
	if cfg == nil {
		cfg = config.CreateDefaultServerConfig()
	}

	// 创建表单数据绑定
	formData := make(map[string]*string)
	formData["bindPort"] = new(string)
	formData["webPort"] = new(string)
	formData["webAddr"] = new(string)
	formData["webUser"] = new(string)
	formData["webPassword"] = new(string)
	formData["logTo"] = new(string)
	formData["logLevel"] = new(string)
	formData["token"] = new(string)

	// 初始化表单数据
	if cfg.BindPort > 0 {
		*formData["bindPort"] = strconv.Itoa(cfg.BindPort)
	}
	if cfg.WebServer.Port > 0 {
		*formData["webPort"] = strconv.Itoa(cfg.WebServer.Port)
	}
	*formData["webAddr"] = cfg.WebServer.Addr
	*formData["webUser"] = cfg.WebServer.User
	*formData["webPassword"] = cfg.WebServer.Password
	*formData["logTo"] = cfg.Log.To
	*formData["logLevel"] = cfg.Log.Level
	*formData["token"] = cfg.Token

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("服务端监听端口").
				Description("FRP 服务端监听端口，客户端通过此端口连接").
				Placeholder("7000").
				Value(formData["bindPort"]),

			huh.NewInput().
				Title("认证令牌 (可选)").
				Description("客户端连接时使用的认证令牌，留空表示不需要认证").
				Placeholder("your_secure_token_here").
				Value(formData["token"]),

			huh.NewInput().
				Title("Web 管理界面地址").
				Description("Web 管理界面监听地址").
				Placeholder("127.0.0.1").
				Value(formData["webAddr"]),

			huh.NewInput().
				Title("Web 管理界面端口").
				Description("Web 管理界面监听端口").
				Placeholder("7500").
				Value(formData["webPort"]).
				Validate(func(str string) error {
					if str == "" {
						return nil // Web 端口可以为空
					}
					port, err := strconv.Atoi(str)
					if err != nil {
						return fmt.Errorf("端口必须是数字")
					}
					if port < 1 || port > 65535 {
						return fmt.Errorf("端口必须在 1-65535 范围内")
					}
					return nil
				}),

			huh.NewInput().
				Title("Web 管理用户名").
				Description("Web 管理界面登录用户名").
				Placeholder("admin").
				Value(formData["webUser"]),

			huh.NewInput().
				Title("Web 管理密码").
				Description("Web 管理界面登录密码").
				Placeholder("admin").
				Value(formData["webPassword"]).
				EchoMode(huh.EchoModePassword),
			huh.NewSelect[string]().
				Title("日志输出位置").
				Description("选择日志输出的位置").
				Options(
					huh.NewOption("控制台", "console"),
					huh.NewOption("文件", "file"),
				).
				Value(formData["logTo"]),

			huh.NewSelect[string]().
				Title("日志级别").
				Description("选择日志记录级别").
				Options(
					huh.NewOption("Trace", "trace"),
					huh.NewOption("Debug", "debug"),
					huh.NewOption("Info", "info"),
					huh.NewOption("Warn", "warn"),
					huh.NewOption("Error", "error"),
				).
				Value(formData["logLevel"]),
		).Title("📄 日志配置"),
	)

	// 表单创建完成，配置更新在 Update 方法中处理

	return &ConfigFormModel{
		form:     form,
		formType: ServerConfigForm,
		config:   cfg,
		formData: formData,
	}
}

// NewClientConfigForm 创建客户端配置表单
func NewClientConfigForm(cfg *config.Config) *ConfigFormModel {
	if cfg == nil {
		cfg = config.CreateDefaultClientConfig()
	}

	var serverAddr, serverPort, token string
	var logTo, logLevel string

	serverAddr = cfg.ServerAddr
	if cfg.ServerPort > 0 {
		serverPort = strconv.Itoa(cfg.ServerPort)
	}
	token = cfg.Token
	logTo = cfg.Log.To
	logLevel = cfg.Log.Level

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("服务器地址").
				Description("FRP 服务端的 IP 地址或域名").
				Placeholder("如: 123.456.789.123 或 your-server.com (本地测试填 127.0.0.1)").
				Value(&serverAddr).
				Validate(func(str string) error {
					if strings.TrimSpace(str) == "" {
						return fmt.Errorf("服务器地址不能为空")
					}
					// 简单的IP或域名格式检查
					str = strings.TrimSpace(str)
					if str == "localhost" || str == "127.0.0.1" {
						return nil // 本地地址总是有效的
					}
					// 检查是否包含非法字符
					if strings.Contains(str, " ") {
						return fmt.Errorf("服务器地址不能包含空格")
					}
					return nil
				}),

			huh.NewInput().
				Title("服务器端口").
				Description("FRP 服务端监听端口 (默认: 7000)").
				Placeholder("7000").
				Value(&serverPort).
				Validate(func(str string) error {
					// 如果为空，设置默认值
					if str == "" {
						serverPort = "7000"
						return nil
					}
					port, err := strconv.Atoi(str)
					if err != nil {
						return fmt.Errorf("端口必须是数字")
					}
					if port < 1 || port > 65535 {
						return fmt.Errorf("端口必须在 1-65535 范围内")
					}
					return nil
				}),

			huh.NewInput().
				Title("认证令牌 (可选)").
				Description("服务端设置的认证令牌，需与服务端一致。如果服务端未设置可留空").
				Placeholder("留空表示无认证").
				Value(&token),
		).Title("🔧 服务器连接配置"),

		huh.NewGroup(
			huh.NewSelect[string]().
				Title("日志输出位置").
				Description("选择日志输出的位置").
				Options(
					huh.NewOption("控制台", "console"),
					huh.NewOption("文件", "file"),
				).
				Value(&logTo),

			huh.NewSelect[string]().
				Title("日志级别").
				Description("选择日志记录级别").
				Options(
					huh.NewOption("Trace", "trace"),
					huh.NewOption("Debug", "debug"),
					huh.NewOption("Info", "info"),
					huh.NewOption("Warn", "warn"),
					huh.NewOption("Error", "error"),
				).
				Value(&logLevel),
		).Title("📄 日志配置"),
	)

	// 表单创建完成，配置更新在 Update 方法中处理

	return &ConfigFormModel{
		form:     form,
		formType: ClientConfigForm,
		config:   cfg,
	}
}

// NewProxyConfigForm 创建代理配置表单
func NewProxyConfigForm(proxy *config.ProxyConfig) *ConfigFormModel {
	if proxy == nil {
		proxy = &config.ProxyConfig{
			Type:    "tcp",
			LocalIP: "127.0.0.1",
		}
	}

	var name, proxyType, localIP, localPort, remotePort string
	var customDomains, secretKey string

	name = proxy.Name
	proxyType = proxy.Type
	localIP = proxy.LocalIP
	if proxy.LocalPort > 0 {
		localPort = strconv.Itoa(proxy.LocalPort)
	}
	if proxy.RemotePort > 0 {
		remotePort = strconv.Itoa(proxy.RemotePort)
	}
	customDomains = strings.Join(proxy.CustomDomains, ",")
	secretKey = proxy.SecretKey

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("代理名称").
				Description("代理的唯一标识名称 (建议使用有意义的名称，如: web-server, ssh-tunnel)").
				Placeholder("web-server").
				Value(&name).
				Validate(func(str string) error {
					str = strings.TrimSpace(str)
					if str == "" {
						return fmt.Errorf("代理名称不能为空")
					}
					// 检查名称格式
					if strings.Contains(str, " ") {
						return fmt.Errorf("代理名称不能包含空格，建议使用连字符")
					}
					if len(str) < 2 {
						return fmt.Errorf("代理名称至少需要2个字符")
					}
					return nil
				}),

			huh.NewSelect[string]().
				Title("代理类型").
				Description("选择代理协议类型 (TCP最常用，HTTP用于网站)").
				Options(
					huh.NewOption("TCP - 通用端口转发 (推荐)", "tcp"),
					huh.NewOption("HTTP - 网站代理", "http"),
					huh.NewOption("HTTPS - 安全网站代理", "https"),
					huh.NewOption("UDP - UDP协议转发", "udp"),
					huh.NewOption("STCP - 安全TCP (需要密钥)", "stcp"),
					huh.NewOption("SUDP - 安全UDP (需要密钥)", "sudp"),
					huh.NewOption("XTCP - 点对点TCP (需要密钥)", "xtcp"),
				).
				Value(&proxyType),

			huh.NewInput().
				Title("本地 IP 地址").
				Description("要代理的本地服务的 IP 地址").
				Placeholder("127.0.0.1").
				Value(&localIP),

			huh.NewInput().
				Title("本地端口").
				Description("要代理的本地服务端口 (如: 22=SSH, 80=HTTP, 3389=RDP, 8080=Web服务)").
				Placeholder("8080").
				Value(&localPort).
				Validate(func(str string) error {
					if str == "" {
						return fmt.Errorf("本地端口不能为空")
					}
					port, err := strconv.Atoi(str)
					if err != nil {
						return fmt.Errorf("端口必须是数字")
					}
					if port < 1 || port > 65535 {
						return fmt.Errorf("端口必须在 1-65535 范围内")
					}
					// 提供常用端口的友好提示
					commonPorts := map[int]string{
						22:   "SSH",
						80:   "HTTP",
						443:  "HTTPS",
						3389: "远程桌面",
						5432: "PostgreSQL",
						3306: "MySQL",
						6379: "Redis",
						8080: "Web服务",
					}
					if service, exists := commonPorts[port]; exists {
						// 这里可以添加提示，但huh库的验证函数只能返回错误
						_ = service // 避免未使用变量警告
					}
					return nil
				}),
		).Title("🔧 基本代理配置"),

		// TCP/UDP 特有配置
		huh.NewGroup(
			huh.NewInput().
				Title("远程端口").
				Description("服务端监听的公网端口 (仅TCP/UDP类型需要)").
				Placeholder("6000").
				Value(&remotePort).
				Validate(func(str string) error {
					if proxyType != "tcp" && proxyType != "udp" {
						return nil // 非 TCP/UDP 类型不需要验证
					}
					if str == "" {
						return fmt.Errorf("TCP/UDP 代理需要设置远程端口")
					}
					port, err := strconv.Atoi(str)
					if err != nil {
						return fmt.Errorf("端口必须是数字")
					}
					if port < 1 || port > 65535 {
						return fmt.Errorf("端口必须在 1-65535 范围内")
					}
					return nil
				}),
		).Title("🌐 TCP/UDP 配置").
			WithHideFunc(func() bool {
				return proxyType != "tcp" && proxyType != "udp"
			}),

		// HTTP/HTTPS 特有配置
		huh.NewGroup(
			huh.NewInput().
				Title("自定义域名").
				Description("绑定的域名，多个域名用逗号分隔 (仅HTTP/HTTPS类型需要)").
				Placeholder("example.com,www.example.com").
				Value(&customDomains).
				Validate(func(str string) error {
					if proxyType != "http" && proxyType != "https" {
						return nil // 非 HTTP/HTTPS 类型不需要验证
					}
					if strings.TrimSpace(str) == "" {
						return fmt.Errorf("HTTP/HTTPS 代理需要设置自定义域名")
					}
					return nil
				}),
		).Title("🌐 HTTP/HTTPS 配置").
			WithHideFunc(func() bool {
				return proxyType != "http" && proxyType != "https"
			}),

		// STCP/SUDP/XTCP 特有配置
		huh.NewGroup(
			huh.NewInput().
				Title("密钥").
				Description("用于安全连接的密钥 (仅STCP/SUDP/XTCP类型需要)").
				Placeholder("your_secret_key").
				Value(&secretKey).
				Validate(func(str string) error {
					if proxyType != "stcp" && proxyType != "sudp" && proxyType != "xtcp" {
						return nil // 非加密类型不需要验证
					}
					if strings.TrimSpace(str) == "" {
						return fmt.Errorf("加密代理需要设置密钥")
					}
					if len(str) < 6 {
						return fmt.Errorf("密钥长度至少6个字符")
					}
					return nil
				}),
		).Title("🔒 加密代理配置").
			WithHideFunc(func() bool {
				return proxyType != "stcp" && proxyType != "sudp" && proxyType != "xtcp"
			}),
	)

	// 表单创建完成，配置更新在 Update 方法中处理

	return &ConfigFormModel{
		form:        form,
		formType:    ProxyConfigForm,
		proxyConfig: proxy,
	}
}

// NewVisitorConfigForm 创建访问者配置表单
func NewVisitorConfigForm(visitor *config.VisitorConfig) *ConfigFormModel {
	if visitor == nil {
		visitor = &config.VisitorConfig{
			Type:     "stcp",
			BindAddr: "127.0.0.1",
		}
	}

	var name, visitorType, serverName, secretKey, bindAddr, bindPort string

	name = visitor.Name
	visitorType = visitor.Type
	serverName = visitor.ServerName
	secretKey = visitor.SecretKey
	bindAddr = visitor.BindAddr
	if visitor.BindPort > 0 {
		bindPort = strconv.Itoa(visitor.BindPort)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("访问者名称").
				Description("访问者的唯一标识名称").
				Placeholder("my-visitor").
				Value(&name).
				Validate(func(str string) error {
					if strings.TrimSpace(str) == "" {
						return fmt.Errorf("访问者名称不能为空")
					}
					return nil
				}),

			huh.NewSelect[string]().
				Title("访问者类型").
				Description("选择访问者类型").
				Options(
					huh.NewOption("STCP", "stcp"),
					huh.NewOption("SUDP", "sudp"),
					huh.NewOption("XTCP", "xtcp"),
				).
				Value(&visitorType),

			huh.NewInput().
				Title("服务器名称").
				Description("要访问的代理服务器名称").
				Placeholder("secret_ssh").
				Value(&serverName).
				Validate(func(str string) error {
					if strings.TrimSpace(str) == "" {
						return fmt.Errorf("服务器名称不能为空")
					}
					return nil
				}),

			huh.NewInput().
				Title("密钥").
				Description("与代理服务器相同的密钥").
				Placeholder("your_secret_key").
				Value(&secretKey).
				Validate(func(str string) error {
					if strings.TrimSpace(str) == "" {
						return fmt.Errorf("密钥不能为空")
					}
					if len(str) < 6 {
						return fmt.Errorf("密钥长度至少6个字符")
					}
					return nil
				}),
		).Title("🔧 基本访问者配置"),

		huh.NewGroup(
			huh.NewInput().
				Title("绑定地址").
				Description("本地绑定的 IP 地址").
				Placeholder("127.0.0.1").
				Value(&bindAddr),

			huh.NewInput().
				Title("绑定端口").
				Description("本地监听端口").
				Placeholder("9000").
				Value(&bindPort).
				Validate(func(str string) error {
					if str == "" {
						return fmt.Errorf("绑定端口不能为空")
					}
					port, err := strconv.Atoi(str)
					if err != nil {
						return fmt.Errorf("端口必须是数字")
					}
					if port < 1 || port > 65535 {
						return fmt.Errorf("端口必须在 1-65535 范围内")
					}
					return nil
				}),
		).Title("🌐 连接配置"),
	)

	// 表单创建完成，配置更新在 Update 方法中处理

	return &ConfigFormModel{
		form:          form,
		formType:      VisitorConfigForm,
		visitorConfig: visitor,
	}
}

// Init 初始化表单
func (m *ConfigFormModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update 更新表单状态
func (m *ConfigFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.form.State == huh.StateCompleted {
				return m, tea.Quit
			}
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		if m.form.State == huh.StateCompleted && !m.completed {
			m.completed = true
			// 表单完成时更新配置
			m.updateConfigFromForm()
		}
	}

	return m, cmd
}

// updateConfigFromForm 从表单更新配置
func (m *ConfigFormModel) updateConfigFromForm() {
	if m.config == nil || m.formData == nil {
		return
	}

	switch m.formType {
	case ServerConfigForm:
		// 更新服务端配置
		if bindPort := *m.formData["bindPort"]; bindPort != "" {
			if port, err := strconv.Atoi(bindPort); err == nil {
				m.config.BindPort = port
			}
		}
		m.config.Token = *m.formData["token"]
		m.config.WebServer.Addr = *m.formData["webAddr"]
		if webPort := *m.formData["webPort"]; webPort != "" {
			if port, err := strconv.Atoi(webPort); err == nil {
				m.config.WebServer.Port = port
			}
		}
		m.config.WebServer.User = *m.formData["webUser"]
		m.config.WebServer.Password = *m.formData["webPassword"]
		m.config.Log.To = *m.formData["logTo"]
		m.config.Log.Level = *m.formData["logLevel"]

	case ClientConfigForm:
		// 更新客户端配置
		m.config.ServerAddr = *m.formData["serverAddr"]
		if serverPort := *m.formData["serverPort"]; serverPort != "" {
			if port, err := strconv.Atoi(serverPort); err == nil {
				m.config.ServerPort = port
			}
		}
		m.config.Token = *m.formData["token"]
		m.config.Log.To = *m.formData["logTo"]
		m.config.Log.Level = *m.formData["logLevel"]

	case ProxyConfigForm:
		// 更新代理配置
		if m.proxyConfig == nil {
			return
		}
		m.proxyConfig.Name = *m.formData["name"]
		m.proxyConfig.Type = *m.formData["proxyType"]
		m.proxyConfig.LocalIP = *m.formData["localIP"]
		if localPort := *m.formData["localPort"]; localPort != "" {
			if port, err := strconv.Atoi(localPort); err == nil {
				m.proxyConfig.LocalPort = port
			}
		}
		if remotePort := *m.formData["remotePort"]; remotePort != "" {
			if port, err := strconv.Atoi(remotePort); err == nil {
				m.proxyConfig.RemotePort = port
			}
		}
		if customDomains := *m.formData["customDomains"]; customDomains != "" {
			m.proxyConfig.CustomDomains = strings.Split(customDomains, ",")
		}
		m.proxyConfig.SecretKey = *m.formData["secretKey"]

	case VisitorConfigForm:
		// 更新访问者配置
		if m.visitorConfig == nil {
			return
		}
		m.visitorConfig.Name = *m.formData["name"]
		m.visitorConfig.Type = *m.formData["visitorType"]
		m.visitorConfig.ServerName = *m.formData["serverName"]
		m.visitorConfig.SecretKey = *m.formData["secretKey"]
		m.visitorConfig.BindAddr = *m.formData["bindAddr"]
		if bindPort := *m.formData["bindPort"]; bindPort != "" {
			if port, err := strconv.Atoi(bindPort); err == nil {
				m.visitorConfig.BindPort = port
			}
		}
	}
}

// View 渲染表单视图
func (m *ConfigFormModel) View() string {
	if m.completed {
		var title string
		switch m.formType {
		case ServerConfigForm:
			title = "服务端配置已完成"
		case ClientConfigForm:
			title = "客户端配置已完成"
		case ProxyConfigForm:
			title = "代理配置已完成"
		case VisitorConfigForm:
			title = "访问者配置已完成"
		}
		return fmt.Sprintf("\n✅ %s\n\n按 ESC 返回\n", title)
	}

	return m.form.View()
}

// IsCompleted 检查表单是否完成
func (m *ConfigFormModel) IsCompleted() bool {
	return m.completed
}

// GetConfig 获取配置
func (m *ConfigFormModel) GetConfig() *config.Config {
	return m.config
}

// GetProxyConfig 获取代理配置
func (m *ConfigFormModel) GetProxyConfig() *config.ProxyConfig {
	return m.proxyConfig
}

// GetVisitorConfig 获取访问者配置
func (m *ConfigFormModel) GetVisitorConfig() *config.VisitorConfig {
	return m.visitorConfig
}

// GetError 获取错误
func (m *ConfigFormModel) GetError() error {
	return m.err
}
