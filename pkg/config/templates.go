package config

import (
	"fmt"
	"time"
)

// ConfigTemplate 配置模板
type ConfigTemplate struct {
	Name        string
	Description string
	Type        string // "server" or "client"
	Config      *Config
	CreatedAt   time.Time
}

// TemplateManager 模板管理器
type TemplateManager struct {
	templates map[string]*ConfigTemplate
}

// NewTemplateManager 创建新的模板管理器
func NewTemplateManager() *TemplateManager {
	tm := &TemplateManager{
		templates: make(map[string]*ConfigTemplate),
	}

	for _, template := range getBuiltinTemplates() {
		tm.templates[template.Name] = template
	}

	return tm
}

// GetTemplates 获取所有模板
func (tm *TemplateManager) GetTemplates() []*ConfigTemplate {
	templates := make([]*ConfigTemplate, 0, len(tm.templates))
	for _, template := range tm.templates {
		templates = append(templates, template)
	}
	return templates
}

// GetTemplatesByType 根据类型获取模板
func (tm *TemplateManager) GetTemplatesByType(configType string) []*ConfigTemplate {
	var templates []*ConfigTemplate
	for _, template := range tm.templates {
		if template.Type == configType {
			templates = append(templates, template)
		}
	}
	return templates
}

// GetTemplate 根据名称获取模板
func (tm *TemplateManager) GetTemplate(name string) (*ConfigTemplate, error) {
	template, exists := tm.templates[name]
	if !exists {
		return nil, fmt.Errorf("模板不存在: %s", name)
	}
	return template, nil
}

// AddTemplate 添加自定义模板
func (tm *TemplateManager) AddTemplate(template *ConfigTemplate) error {
	if template.Name == "" {
		return fmt.Errorf("模板名称不能为空")
	}
	tm.templates[template.Name] = template
	return nil
}

// getBuiltinTemplates 获取内置模板
func getBuiltinTemplates() []*ConfigTemplate {
	return []*ConfigTemplate{
		{
			Name:        "基础服务端",
			Description: "基本的FRP服务端配置",
			Type:        "server",
			Config: &Config{
				BindPort: 7000,
				WebServer: WebServerConfig{
					Port:     7500,
					User:     "admin",
					Password: "admin",
				},
				Log: LogConfig{
					To:    "console",
					Level: "info",
				},
			},
			CreatedAt: time.Now(),
		},
		{
			Name:        "安全服务端",
			Description: "带认证的安全服务端配置",
			Type:        "server",
			Config: &Config{
				BindPort: 7000,
				Token:    "your_token_here",
				WebServer: WebServerConfig{
					Port:     7500,
					User:     "admin",
					Password: "secure_password",
				},
				Log: LogConfig{
					To:    "file",
					Level: "info",
				},
			},
			CreatedAt: time.Now(),
		},
		{
			Name:        "SSH隧道客户端",
			Description: "SSH端口转发客户端配置",
			Type:        "client",
			Config: &Config{
				ServerAddr: "your_server_ip",
				ServerPort: 7000,
				Log: LogConfig{
					To:    "console",
					Level: "info",
				},
				Proxies: []ProxyConfig{
					{
						Name:       "ssh",
						Type:       "tcp",
						LocalIP:    "127.0.0.1",
						LocalPort:  22,
						RemotePort: 6000,
					},
				},
			},
			CreatedAt: time.Now(),
		},
		{
			Name:        "Web服务客户端",
			Description: "HTTP/HTTPS服务客户端配置",
			Type:        "client",
			Config: &Config{
				ServerAddr: "your_server_ip",
				ServerPort: 7000,
				Log: LogConfig{
					To:    "console",
					Level: "info",
				},
				Proxies: []ProxyConfig{
					{
						Name:          "web",
						Type:          "http",
						LocalIP:       "127.0.0.1",
						LocalPort:     80,
						CustomDomains: []string{"yourdomain.com"},
					},
				},
			},
			CreatedAt: time.Now(),
		},
		{
			Name:        "远程桌面客户端",
			Description: "RDP/VNC远程桌面客户端配置",
			Type:        "client",
			Config: &Config{
				ServerAddr: "your_server_ip",
				ServerPort: 7000,
				Log: LogConfig{
					To:    "console",
					Level: "info",
				},
				Proxies: []ProxyConfig{
					{
						Name:       "rdp",
						Type:       "tcp",
						LocalIP:    "127.0.0.1",
						LocalPort:  3389,
						RemotePort: 3389,
					},
				},
			},
			CreatedAt: time.Now(),
		},
		{
			Name:        "数据库客户端",
			Description: "数据库端口转发客户端配置",
			Type:        "client",
			Config: &Config{
				ServerAddr: "your_server_ip",
				ServerPort: 7000,
				Log: LogConfig{
					To:    "console",
					Level: "info",
				},
				Proxies: []ProxyConfig{
					{
						Name:       "mysql",
						Type:       "tcp",
						LocalIP:    "127.0.0.1",
						LocalPort:  3306,
						RemotePort: 3306,
					},
					{
						Name:       "postgres",
						Type:       "tcp",
						LocalIP:    "127.0.0.1",
						LocalPort:  5432,
						RemotePort: 5432,
					},
				},
			},
			CreatedAt: time.Now(),
		},
		{
			Name:        "游戏服务器客户端",
			Description: "游戏服务器端口转发配置",
			Type:        "client",
			Config: &Config{
				ServerAddr: "your_server_ip",
				ServerPort: 7000,
				Log: LogConfig{
					To:    "console",
					Level: "info",
				},
				Proxies: []ProxyConfig{
					{
						Name:       "minecraft",
						Type:       "tcp",
						LocalIP:    "127.0.0.1",
						LocalPort:  25565,
						RemotePort: 25565,
					},
					{
						Name:       "cs-server",
						Type:       "udp",
						LocalIP:    "127.0.0.1",
						LocalPort:  27015,
						RemotePort: 27015,
					},
				},
			},
			CreatedAt: time.Now(),
		},
		{
			Name:        "安全内网穿透",
			Description: "使用STCP的安全内网穿透配置",
			Type:        "client",
			Config: &Config{
				ServerAddr: "your_server_ip",
				ServerPort: 7000,
				Token:      "your_token_here",
				Log: LogConfig{
					To:    "console",
					Level: "info",
				},
				Proxies: []ProxyConfig{
					{
						Name:      "secret_ssh",
						Type:      "stcp",
						SecretKey: "abcdefg",
						LocalIP:   "127.0.0.1",
						LocalPort: 22,
					},
				},
				Visitors: []VisitorConfig{
					{
						Name:       "secret_ssh_visitor",
						Type:       "stcp",
						ServerName: "secret_ssh",
						SecretKey:  "abcdefg",
						BindAddr:   "127.0.0.1",
						BindPort:   6000,
					},
				},
			},
			CreatedAt: time.Now(),
		},
	}
}

// ApplyTemplate 应用模板到配置
func (tm *TemplateManager) ApplyTemplate(templateName string) (*Config, error) {
	template, err := tm.GetTemplate(templateName)
	if err != nil {
		return nil, err
	}

	config := *template.Config

	proxies := make([]ProxyConfig, len(template.Config.Proxies))
	copy(proxies, template.Config.Proxies)
	config.Proxies = proxies

	visitors := make([]VisitorConfig, len(template.Config.Visitors))
	copy(visitors, template.Config.Visitors)
	config.Visitors = visitors

	return &config, nil
}

// MergeTemplate 合并模板到现有配置
func (tm *TemplateManager) MergeTemplate(target *Config, templateName string) (*Config, error) {
	template, err := tm.GetTemplate(templateName)
	if err != nil {
		return nil, err
	}

	if target == nil {
		return tm.ApplyTemplate(templateName)
	}

	merged := *target

	if merged.ServerAddr == "" && template.Config.ServerAddr != "" {
		merged.ServerAddr = template.Config.ServerAddr
	}
	if merged.ServerPort == 0 && template.Config.ServerPort != 0 {
		merged.ServerPort = template.Config.ServerPort
	}
	if merged.Token == "" && template.Config.Token != "" {
		merged.Token = template.Config.Token
	}
	if merged.BindPort == 0 && template.Config.BindPort != 0 {
		merged.BindPort = template.Config.BindPort
	}

	if merged.WebServer.Port == 0 && template.Config.WebServer.Port != 0 {
		merged.WebServer = template.Config.WebServer
	}

	if merged.Log.Level == "" && template.Config.Log.Level != "" {
		merged.Log = template.Config.Log
	}

	proxyNames := make(map[string]bool)
	for _, proxy := range merged.Proxies {
		proxyNames[proxy.Name] = true
	}

	for _, proxy := range template.Config.Proxies {
		if !proxyNames[proxy.Name] {
			merged.Proxies = append(merged.Proxies, proxy)
		}
	}

	visitorNames := make(map[string]bool)
	for _, visitor := range merged.Visitors {
		visitorNames[visitor.Name] = true
	}

	for _, visitor := range template.Config.Visitors {
		if !visitorNames[visitor.Name] {
			merged.Visitors = append(merged.Visitors, visitor)
		}
	}

	return &merged, nil
}

// SaveTemplate 保存当前配置为模板
func (tm *TemplateManager) SaveTemplate(name, description, configType string, config *Config) error {
	if name == "" {
		return fmt.Errorf("模板名称不能为空")
	}
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	template := &ConfigTemplate{
		Name:        name,
		Description: description,
		Type:        configType,
		Config:      config,
		CreatedAt:   time.Now(),
	}

	return tm.AddTemplate(template)
}

// DeleteTemplate 删除模板
func (tm *TemplateManager) DeleteTemplate(name string) error {
	if _, exists := tm.templates[name]; !exists {
		return fmt.Errorf("模板不存在: %s", name)
	}
	delete(tm.templates, name)
	return nil
}

// ExportTemplate 导出模板为配置文件
func (tm *TemplateManager) ExportTemplate(name, filePath string) error {
	template, err := tm.GetTemplate(name)
	if err != nil {
		return err
	}

	loader := NewLoader(filePath)
	return loader.ExportToFile(template.Config, filePath)
}

// ImportTemplate 从配置文件导入模板
func (tm *TemplateManager) ImportTemplate(name, description, filePath string) error {
	loader := NewLoader(filePath)
	config, err := loader.ImportFromFile(filePath)
	if err != nil {
		return err
	}

	configType := detectConfigType(config)
	return tm.SaveTemplate(name, description, configType, config)
}
