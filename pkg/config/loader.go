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
				Name:       "ssh",
				Type:       "tcp",
				LocalIP:    "127.0.0.1",
				LocalPort:  22,
				RemotePort: 6000,
			},
		},
	}
}

// AddProxy 添加代理配置
func (l *Loader) AddProxy(proxy ProxyConfig) error {
	if l.config == nil {
		return fmt.Errorf("配置尚未加载")
	}

	for _, existingProxy := range l.config.Proxies {
		if existingProxy.Name == proxy.Name {
			return fmt.Errorf("代理名称 '%s' 已存在", proxy.Name)
		}
	}

	l.config.Proxies = append(l.config.Proxies, proxy)
	return nil
}

// RemoveProxy 移除代理配置
func (l *Loader) RemoveProxy(name string) error {
	if l.config == nil {
		return fmt.Errorf("配置尚未加载")
	}

	for i, proxy := range l.config.Proxies {
		if proxy.Name == name {
			l.config.Proxies = append(l.config.Proxies[:i], l.config.Proxies[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("未找到名称为 '%s' 的代理", name)
}

// UpdateProxy 更新代理配置
func (l *Loader) UpdateProxy(name string, newProxy ProxyConfig) error {
	if l.config == nil {
		return fmt.Errorf("配置尚未加载")
	}

	for i, proxy := range l.config.Proxies {
		if proxy.Name == name {
			newProxy.Name = name
			l.config.Proxies[i] = newProxy
			return nil
		}
	}

	return fmt.Errorf("未找到名称为 '%s' 的代理", name)
}

// GetProxy 获取代理配置
func (l *Loader) GetProxy(name string) (*ProxyConfig, error) {
	if l.config == nil {
		return nil, fmt.Errorf("配置尚未加载")
	}

	for _, proxy := range l.config.Proxies {
		if proxy.Name == name {
			return &proxy, nil
		}
	}

	return nil, fmt.Errorf("未找到名称为 '%s' 的代理", name)
}

// ListProxies 列出所有代理
func (l *Loader) ListProxies() []ProxyConfig {
	if l.config == nil {
		return []ProxyConfig{}
	}
	return l.config.Proxies
}

// Backup 备份配置文件
func (l *Loader) Backup() error {
	backupPath := l.configPath + ".backup." + time.Now().Format("20060102_150405")

	originalData, err := os.ReadFile(l.configPath)
	if err != nil {
		return fmt.Errorf("读取原配置文件失败: %w", err)
	}

	if err := os.WriteFile(backupPath, originalData, 0644); err != nil {
		return fmt.Errorf("创建备份文件失败: %w", err)
	}

	return nil
}

// Restore 恢复配置文件
func (l *Loader) Restore() error {
	backupPattern := l.configPath + ".backup.*"
	matches, err := filepath.Glob(backupPattern)
	if err != nil {
		return fmt.Errorf("查找备份文件失败: %w", err)
	}

	if len(matches) == 0 {
		return fmt.Errorf("未找到备份文件")
	}

	latestBackup := matches[len(matches)-1]
	backupData, err := os.ReadFile(latestBackup)
	if err != nil {
		return fmt.Errorf("读取备份文件失败: %w", err)
	}

	if err := os.WriteFile(l.configPath, backupData, 0644); err != nil {
		return fmt.Errorf("恢复配置文件失败: %w", err)
	}

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
		return nil, fmt.Errorf("导入文件不存在: %s", filePath)
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取导入文件失败: %w", err)
	}

	// 解析 YAML
	var config Config
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("解析导入文件失败: %w", err)
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

	if source.ServerAddr != "" {
		merged.ServerAddr = source.ServerAddr
	}
	if source.ServerPort != 0 {
		merged.ServerPort = source.ServerPort
	}
	if source.Token != "" {
		merged.Token = source.Token
	}

	if source.BindPort != 0 {
		merged.BindPort = source.BindPort
	}
	if source.BindUDPPort != 0 {
		merged.BindUDPPort = source.BindUDPPort
	}
	if source.KCPBindPort != 0 {
		merged.KCPBindPort = source.KCPBindPort
	}

	if source.WebServer.Port != 0 {
		merged.WebServer.Port = source.WebServer.Port
	}
	if source.WebServer.User != "" {
		merged.WebServer.User = source.WebServer.User
	}
	if source.WebServer.Password != "" {
		merged.WebServer.Password = source.WebServer.Password
	}

	if source.Log.Level != "" {
		merged.Log.Level = source.Log.Level
	}
	if source.Log.To != "" {
		merged.Log.To = source.Log.To
	}

	proxyMap := make(map[string]bool)
	for _, proxy := range merged.Proxies {
		proxyMap[proxy.Name] = true
	}

	for _, proxy := range source.Proxies {
		if !proxyMap[proxy.Name] {
			merged.Proxies = append(merged.Proxies, proxy)
		}
	}

	visitorMap := make(map[string]bool)
	for _, visitor := range merged.Visitors {
		visitorMap[visitor.Name] = true
	}

	for _, visitor := range source.Visitors {
		if !visitorMap[visitor.Name] {
			merged.Visitors = append(merged.Visitors, visitor)
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
	_, err := l.ImportFromFile(filePath)
	return err
}
