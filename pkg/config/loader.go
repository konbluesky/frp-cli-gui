package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config FRP 配置结构
type Config struct {
	// 通用配置
	ServerAddr string `yaml:"serverAddr,omitempty"`
	ServerPort int    `yaml:"serverPort,omitempty"`
	Token      string `yaml:"token,omitempty"`

	// 服务端配置
	BindPort      int    `yaml:"bindPort,omitempty"`
	BindUDPPort   int    `yaml:"bindUDPPort,omitempty"`
	KCPBindPort   int    `yaml:"kcpBindPort,omitempty"`
	ProxyBindAddr string `yaml:"proxyBindAddr,omitempty"`

	// Web 服务器配置
	WebServer WebServerConfig `yaml:"webServer,omitempty"`

	// 日志配置
	Log LogConfig `yaml:"log,omitempty"`

	// 客户端代理配置
	Proxies []ProxyConfig `yaml:"proxies,omitempty"`

	// 访问者配置
	Visitors []VisitorConfig `yaml:"visitors,omitempty"`
}

// WebServerConfig Web 服务器配置
type WebServerConfig struct {
	Addr        string `yaml:"addr,omitempty"`
	Port        int    `yaml:"port,omitempty"`
	User        string `yaml:"user,omitempty"`
	Password    string `yaml:"password,omitempty"`
	AssetsDir   string `yaml:"assetsDir,omitempty"`
	PProfEnable bool   `yaml:"pprofEnable,omitempty"`
}

// LogConfig 日志配置
type LogConfig struct {
	To                string `yaml:"to,omitempty"`
	Level             string `yaml:"level,omitempty"`
	MaxLogFile        int    `yaml:"maxLogFile,omitempty"`
	DisablePrintColor bool   `yaml:"disablePrintColor,omitempty"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Name      string `yaml:"name"`
	Type      string `yaml:"type"`
	LocalIP   string `yaml:"localIP,omitempty"`
	LocalPort int    `yaml:"localPort,omitempty"`

	// TCP/UDP 代理配置
	RemotePort int `yaml:"remotePort,omitempty"`

	// HTTP/HTTPS 代理配置
	CustomDomains     []string `yaml:"customDomains,omitempty"`
	Subdomain         string   `yaml:"subdomain,omitempty"`
	Locations         []string `yaml:"locations,omitempty"`
	HTTPUser          string   `yaml:"httpUser,omitempty"`
	HTTPPwd           string   `yaml:"httpPwd,omitempty"`
	HostHeaderRewrite string   `yaml:"hostHeaderRewrite,omitempty"`

	// STCP/SUDP/XTCP 代理配置
	SecretKey  string `yaml:"secretKey,omitempty"`
	Role       string `yaml:"role,omitempty"`
	ServerName string `yaml:"serverName,omitempty"`

	// 插件配置
	Plugin       string            `yaml:"plugin,omitempty"`
	PluginParams map[string]string `yaml:"pluginParams,omitempty"`

	// 负载均衡配置
	Group    string `yaml:"group,omitempty"`
	GroupKey string `yaml:"groupKey,omitempty"`

	// 健康检查配置
	HealthCheck HealthCheckConfig `yaml:"healthCheck,omitempty"`

	// 带宽限制
	BandwidthLimit string `yaml:"bandwidthLimit,omitempty"`

	// 其他配置
	UseEncryption  bool `yaml:"useEncryption,omitempty"`
	UseCompression bool `yaml:"useCompression,omitempty"`
}

// VisitorConfig 访问者配置
type VisitorConfig struct {
	Name       string `yaml:"name"`
	Type       string `yaml:"type"`
	ServerName string `yaml:"serverName"`
	SecretKey  string `yaml:"secretKey"`
	BindAddr   string `yaml:"bindAddr,omitempty"`
	BindPort   int    `yaml:"bindPort"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Type        string   `yaml:"type,omitempty"`
	TimeoutS    int      `yaml:"timeoutS,omitempty"`
	MaxFailed   int      `yaml:"maxFailed,omitempty"`
	IntervalS   int      `yaml:"intervalS,omitempty"`
	Path        string   `yaml:"path,omitempty"`
	HTTPHeaders []string `yaml:"httpHeaders,omitempty"`
}

// Loader 配置加载器
type Loader struct {
	configPath string
	config     *Config
}

// NewLoader 创建新的配置加载器
func NewLoader(configPath string) *Loader {
	return &Loader{
		configPath: configPath,
	}
}

// Load 加载配置文件
func (l *Loader) Load() (*Config, error) {
	// 检查文件是否存在
	if _, err := os.Stat(l.configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", l.configPath)
	}

	// 读取文件内容
	file, err := os.Open(l.configPath)
	if err != nil {
		return nil, fmt.Errorf("打开配置文件失败: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析 YAML
	var config Config
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	l.config = &config
	return &config, nil
}

// Save 保存配置文件
func (l *Loader) Save(config *Config) error {
	// 创建目录（如果不存在）
	dir := filepath.Dir(l.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 序列化为 YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(l.configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	l.config = config
	return nil
}

// GetConfig 获取当前配置
func (l *Loader) GetConfig() *Config {
	return l.config
}

// CreateDefaultServerConfig 创建默认服务端配置
func CreateDefaultServerConfig() *Config {
	return &Config{
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
	}
}

// CreateDefaultClientConfig 创建默认客户端配置
func CreateDefaultClientConfig() *Config {
	return &Config{
		ServerAddr: "127.0.0.1",
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
			{
				Name:       "ssh",
				Type:       "tcp",
				LocalIP:    "127.0.0.1",
				LocalPort:  22,
				RemotePort: 2222,
			},
		},
	}
}

// AddProxy 添加代理配置
func (l *Loader) AddProxy(proxy ProxyConfig) error {
	if l.config == nil {
		return fmt.Errorf("配置未加载")
	}

	// 检查代理名称是否已存在
	for _, p := range l.config.Proxies {
		if p.Name == proxy.Name {
			return fmt.Errorf("代理名称 '%s' 已存在", proxy.Name)
		}
	}

	l.config.Proxies = append(l.config.Proxies, proxy)
	return nil
}

// RemoveProxy 移除代理配置
func (l *Loader) RemoveProxy(name string) error {
	if l.config == nil {
		return fmt.Errorf("配置未加载")
	}

	for i, proxy := range l.config.Proxies {
		if proxy.Name == name {
			l.config.Proxies = append(l.config.Proxies[:i], l.config.Proxies[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("代理 '%s' 不存在", name)
}

// UpdateProxy 更新代理配置
func (l *Loader) UpdateProxy(name string, newProxy ProxyConfig) error {
	if l.config == nil {
		return fmt.Errorf("配置未加载")
	}

	for i, proxy := range l.config.Proxies {
		if proxy.Name == name {
			newProxy.Name = name // 保持名称不变
			l.config.Proxies[i] = newProxy
			return nil
		}
	}

	return fmt.Errorf("代理 '%s' 不存在", name)
}

// GetProxy 获取代理配置
func (l *Loader) GetProxy(name string) (*ProxyConfig, error) {
	if l.config == nil {
		return nil, fmt.Errorf("配置未加载")
	}

	for _, proxy := range l.config.Proxies {
		if proxy.Name == name {
			return &proxy, nil
		}
	}

	return nil, fmt.Errorf("代理 '%s' 不存在", name)
}

// ListProxies 列出所有代理
func (l *Loader) ListProxies() []ProxyConfig {
	if l.config == nil {
		return nil
	}
	return l.config.Proxies
}

// Backup 备份配置文件
func (l *Loader) Backup() error {
	if _, err := os.Stat(l.configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在，无法备份")
	}

	backupPath := l.configPath + ".backup"

	// 读取原文件
	content, err := os.ReadFile(l.configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 写入备份文件
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return fmt.Errorf("创建备份文件失败: %w", err)
	}

	return nil
}

// Restore 恢复配置文件
func (l *Loader) Restore() error {
	backupPath := l.configPath + ".backup"

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("备份文件不存在")
	}

	// 读取备份文件
	content, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("读取备份文件失败: %w", err)
	}

	// 写入原文件
	if err := os.WriteFile(l.configPath, content, 0644); err != nil {
		return fmt.Errorf("恢复配置文件失败: %w", err)
	}

	// 重新加载配置
	_, err = l.Load()
	return err
}

// GetConfigPath 获取配置文件路径
func (l *Loader) GetConfigPath() string {
	return l.configPath
}

// SetConfigPath 设置配置文件路径
func (l *Loader) SetConfigPath(path string) {
	l.configPath = path
}

// ExportToFile 导出配置到指定文件
func (l *Loader) ExportToFile(config *Config, filePath string) error {
	// 创建目录（如果不存在）
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建导出目录失败: %w", err)
	}

	// 序列化为 YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 添加配置文件头部注释
	header := fmt.Sprintf("# FRP 配置文件\n# 导出时间: %s\n# 配置类型: %s\n\n",
		time.Now().Format("2006-01-02 15:04:05"),
		detectConfigType(config))

	finalData := append([]byte(header), data...)

	// 写入文件
	if err := os.WriteFile(filePath, finalData, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// ImportFromFile 从指定文件导入配置
func (l *Loader) ImportFromFile(filePath string) (*Config, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", filePath)
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析 YAML
	var config Config
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// MergeConfig 合并两个配置
func (l *Loader) MergeConfig(target *Config, source *Config) *Config {
	if target == nil {
		return source
	}
	if source == nil {
		return target
	}

	merged := *target

	// 合并基本配置（只有目标配置为空时才设置）
	if merged.ServerAddr == "" && source.ServerAddr != "" {
		merged.ServerAddr = source.ServerAddr
	}
	if merged.ServerPort == 0 && source.ServerPort != 0 {
		merged.ServerPort = source.ServerPort
	}
	if merged.BindPort == 0 && source.BindPort != 0 {
		merged.BindPort = source.BindPort
	}
	if merged.Token == "" && source.Token != "" {
		merged.Token = source.Token
	}

	// 合并 Web 服务器配置
	if merged.WebServer.Port == 0 && source.WebServer.Port != 0 {
		merged.WebServer.Port = source.WebServer.Port
	}
	if merged.WebServer.User == "" && source.WebServer.User != "" {
		merged.WebServer.User = source.WebServer.User
	}
	if merged.WebServer.Password == "" && source.WebServer.Password != "" {
		merged.WebServer.Password = source.WebServer.Password
	}

	// 合并日志配置
	if merged.Log.Level == "" && source.Log.Level != "" {
		merged.Log.Level = source.Log.Level
	}
	if merged.Log.To == "" && source.Log.To != "" {
		merged.Log.To = source.Log.To
	}

	// 合并代理配置（避免重复）
	existingProxyNames := make(map[string]bool)
	for _, proxy := range merged.Proxies {
		existingProxyNames[proxy.Name] = true
	}

	for _, sourceProxy := range source.Proxies {
		if !existingProxyNames[sourceProxy.Name] {
			merged.Proxies = append(merged.Proxies, sourceProxy)
		}
	}

	// 合并访问者配置（避免重复）
	existingVisitorNames := make(map[string]bool)
	for _, visitor := range merged.Visitors {
		existingVisitorNames[visitor.Name] = true
	}

	for _, sourceVisitor := range source.Visitors {
		if !existingVisitorNames[sourceVisitor.Name] {
			merged.Visitors = append(merged.Visitors, sourceVisitor)
		}
	}

	return &merged
}

// detectConfigType 检测配置类型
func detectConfigType(config *Config) string {
	if config.BindPort > 0 || config.WebServer.Port > 0 {
		return "server"
	}
	if config.ServerAddr != "" || len(config.Proxies) > 0 {
		return "client"
	}
	return "unknown"
}

// ValidateConfigFile 验证配置文件格式
func (l *Loader) ValidateConfigFile(filePath string) error {
	// 检查文件扩展名
	ext := filepath.Ext(filePath)
	if ext != ".yaml" && ext != ".yml" {
		return fmt.Errorf("不支持的文件格式: %s，仅支持 .yaml 和 .yml", ext)
	}

	// 尝试解析文件
	_, err := l.ImportFromFile(filePath)
	return err
}
