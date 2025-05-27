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
	templates []ConfigTemplate
}

// NewTemplateManager 创建新的模板管理器
func NewTemplateManager() *TemplateManager {
	return &TemplateManager{
		templates: getBuiltinTemplates(),
	}
}

// GetTemplates 获取所有模板
func (tm *TemplateManager) GetTemplates() []ConfigTemplate {
	return tm.templates
}

// GetTemplatesByType 根据类型获取模板
func (tm *TemplateManager) GetTemplatesByType(templateType string) []ConfigTemplate {
	var filtered []ConfigTemplate
	for _, template := range tm.templates {
		if template.Type == templateType {
			filtered = append(filtered, template)
		}
	}
	return filtered
}

// GetTemplate 根据名称获取模板
func (tm *TemplateManager) GetTemplate(name string) *ConfigTemplate {
	for _, template := range tm.templates {
		if template.Name == name {
			return &template
		}
	}
	return nil
}

// AddTemplate 添加自定义模板
func (tm *TemplateManager) AddTemplate(template ConfigTemplate) {
	template.CreatedAt = time.Now()
	tm.templates = append(tm.templates, template)
}

// getBuiltinTemplates 获取内置模板
func getBuiltinTemplates() []ConfigTemplate {
	return []ConfigTemplate{
		{
			Name:        "基础服务端",
			Description: "基础 FRP 服务端配置，包含 Web 管理界面",
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
			Description: "带认证令牌的安全 FRP 服务端配置",
			Type:        "server",
			Config: &Config{
				BindPort: 7000,
				Token:    "your_secure_token_here",
				WebServer: WebServerConfig{
					Port:     7500,
					User:     "admin",
					Password: "secure_password",
				},
				Log: LogConfig{
					To:    "console",
					Level: "warn",
				},
			},
			CreatedAt: time.Now(),
		},
		{
			Name:        "Web 服务代理",
			Description: "HTTP/HTTPS Web 服务代理模板",
			Type:        "client",
			Config: &Config{
				ServerAddr: "frp.example.com",
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
						LocalPort:     8080,
						CustomDomains: []string{"www.example.com"},
					},
				},
			},
			CreatedAt: time.Now(),
		},
		{
			Name:        "SSH 隧道",
			Description: "SSH 服务代理模板",
			Type:        "client",
			Config: &Config{
				ServerAddr: "frp.example.com",
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
			Name:        "多服务代理",
			Description: "包含多个服务的客户端配置模板",
			Type:        "client",
			Config: &Config{
				ServerAddr: "frp.example.com",
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
						LocalPort:     8080,
						CustomDomains: []string{"web.example.com"},
					},
					{
						Name:       "ssh",
						Type:       "tcp",
						LocalIP:    "127.0.0.1",
						LocalPort:  22,
						RemotePort: 6000,
					},
					{
						Name:       "database",
						Type:       "tcp",
						LocalIP:    "127.0.0.1",
						LocalPort:  3306,
						RemotePort: 6001,
					},
				},
			},
			CreatedAt: time.Now(),
		},
		{
			Name:        "HTTPS 安全代理",
			Description: "HTTPS 服务代理模板",
			Type:        "client",
			Config: &Config{
				ServerAddr: "frp.example.com",
				ServerPort: 7000,
				Token:      "your_secure_token_here",
				Log: LogConfig{
					To:    "console",
					Level: "info",
				},
				Proxies: []ProxyConfig{
					{
						Name:          "https_web",
						Type:          "https",
						LocalIP:       "127.0.0.1",
						LocalPort:     8443,
						CustomDomains: []string{"secure.example.com"},
					},
				},
			},
			CreatedAt: time.Now(),
		},
		{
			Name:        "内网穿透",
			Description: "STCP 内网穿透模板",
			Type:        "client",
			Config: &Config{
				ServerAddr: "frp.example.com",
				ServerPort: 7000,
				Log: LogConfig{
					To:    "console",
					Level: "info",
				},
				Proxies: []ProxyConfig{
					{
						Name:      "secret_ssh",
						Type:      "stcp",
						LocalIP:   "127.0.0.1",
						LocalPort: 22,
						SecretKey: "abcdefg",
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
func (tm *TemplateManager) ApplyTemplate(templateName string, targetConfig *Config) error {
	template := tm.GetTemplate(templateName)
	if template == nil {
		return fmt.Errorf("模板 '%s' 不存在", templateName)
	}

	// 深拷贝模板配置
	*targetConfig = *template.Config

	// 复制代理配置
	if len(template.Config.Proxies) > 0 {
		targetConfig.Proxies = make([]ProxyConfig, len(template.Config.Proxies))
		copy(targetConfig.Proxies, template.Config.Proxies)
	}

	// 复制访问者配置
	if len(template.Config.Visitors) > 0 {
		targetConfig.Visitors = make([]VisitorConfig, len(template.Config.Visitors))
		copy(targetConfig.Visitors, template.Config.Visitors)
	}

	return nil
}

// MergeTemplate 合并模板到现有配置
func (tm *TemplateManager) MergeTemplate(templateName string, targetConfig *Config) error {
	template := tm.GetTemplate(templateName)
	if template == nil {
		return fmt.Errorf("模板 '%s' 不存在", templateName)
	}

	// 合并基本配置（只有目标配置为空时才设置）
	if targetConfig.ServerAddr == "" && template.Config.ServerAddr != "" {
		targetConfig.ServerAddr = template.Config.ServerAddr
	}
	if targetConfig.ServerPort == 0 && template.Config.ServerPort != 0 {
		targetConfig.ServerPort = template.Config.ServerPort
	}
	if targetConfig.BindPort == 0 && template.Config.BindPort != 0 {
		targetConfig.BindPort = template.Config.BindPort
	}
	if targetConfig.Token == "" && template.Config.Token != "" {
		targetConfig.Token = template.Config.Token
	}

	// 合并 Web 服务器配置
	if targetConfig.WebServer.Port == 0 && template.Config.WebServer.Port != 0 {
		targetConfig.WebServer.Port = template.Config.WebServer.Port
	}
	if targetConfig.WebServer.User == "" && template.Config.WebServer.User != "" {
		targetConfig.WebServer.User = template.Config.WebServer.User
	}
	if targetConfig.WebServer.Password == "" && template.Config.WebServer.Password != "" {
		targetConfig.WebServer.Password = template.Config.WebServer.Password
	}

	// 合并日志配置
	if targetConfig.Log.Level == "" && template.Config.Log.Level != "" {
		targetConfig.Log.Level = template.Config.Log.Level
	}
	if targetConfig.Log.To == "" && template.Config.Log.To != "" {
		targetConfig.Log.To = template.Config.Log.To
	}

	// 添加代理配置（避免重复）
	existingProxyNames := make(map[string]bool)
	for _, proxy := range targetConfig.Proxies {
		existingProxyNames[proxy.Name] = true
	}

	for _, templateProxy := range template.Config.Proxies {
		if !existingProxyNames[templateProxy.Name] {
			targetConfig.Proxies = append(targetConfig.Proxies, templateProxy)
		}
	}

	// 添加访问者配置（避免重复）
	existingVisitorNames := make(map[string]bool)
	for _, visitor := range targetConfig.Visitors {
		existingVisitorNames[visitor.Name] = true
	}

	for _, templateVisitor := range template.Config.Visitors {
		if !existingVisitorNames[templateVisitor.Name] {
			targetConfig.Visitors = append(targetConfig.Visitors, templateVisitor)
		}
	}

	return nil
}

// SaveTemplate 保存当前配置为模板
func (tm *TemplateManager) SaveTemplate(name, description, templateType string, config *Config) {
	template := ConfigTemplate{
		Name:        name,
		Description: description,
		Type:        templateType,
		Config:      config,
		CreatedAt:   time.Now(),
	}

	tm.AddTemplate(template)
}

// DeleteTemplate 删除模板
func (tm *TemplateManager) DeleteTemplate(name string) bool {
	for i, template := range tm.templates {
		if template.Name == name {
			tm.templates = append(tm.templates[:i], tm.templates[i+1:]...)
			return true
		}
	}
	return false
}

// ExportTemplate 导出模板为配置文件
func (tm *TemplateManager) ExportTemplate(templateName, filePath string) error {
	template := tm.GetTemplate(templateName)
	if template == nil {
		return fmt.Errorf("模板 '%s' 不存在", templateName)
	}

	loader := NewLoader(filePath)
	return loader.Save(template.Config)
}

// ImportTemplate 从配置文件导入模板
func (tm *TemplateManager) ImportTemplate(name, description, templateType, filePath string) error {
	loader := NewLoader(filePath)
	config, err := loader.Load()
	if err != nil {
		return fmt.Errorf("导入模板失败: %w", err)
	}

	template := ConfigTemplate{
		Name:        name,
		Description: description,
		Type:        templateType,
		Config:      config,
		CreatedAt:   time.Now(),
	}

	tm.AddTemplate(template)
	return nil
}
