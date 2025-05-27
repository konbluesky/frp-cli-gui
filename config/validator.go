package config

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// Validator 配置验证器
type Validator struct{}

// NewValidator 创建新的验证器
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateConfig 验证完整配置
func (v *Validator) ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	// 验证服务端配置
	if err := v.validateServerConfig(config); err != nil {
		return fmt.Errorf("服务端配置错误: %w", err)
	}

	// 验证客户端配置
	if err := v.validateClientConfig(config); err != nil {
		return fmt.Errorf("客户端配置错误: %w", err)
	}

	// 验证代理配置
	if err := v.validateProxies(config.Proxies); err != nil {
		return fmt.Errorf("代理配置错误: %w", err)
	}

	// 验证访问者配置
	if err := v.validateVisitors(config.Visitors); err != nil {
		return fmt.Errorf("访问者配置错误: %w", err)
	}

	return nil
}

// ValidateConfigDetailed 详细验证配置，返回所有错误
func (v *Validator) ValidateConfigDetailed(config *Config) []string {
	var errors []string

	if config == nil {
		return []string{"配置不能为空"}
	}

	// 验证服务端配置
	if serverErrors := v.validateServerConfigDetailed(config); len(serverErrors) > 0 {
		errors = append(errors, serverErrors...)
	}

	// 验证客户端配置
	if clientErrors := v.validateClientConfigDetailed(config); len(clientErrors) > 0 {
		errors = append(errors, clientErrors...)
	}

	// 验证代理配置
	if proxyErrors := v.validateProxiesDetailed(config.Proxies); len(proxyErrors) > 0 {
		errors = append(errors, proxyErrors...)
	}

	// 验证访问者配置
	if visitorErrors := v.validateVisitorsDetailed(config.Visitors); len(visitorErrors) > 0 {
		errors = append(errors, visitorErrors...)
	}

	return errors
}

// validateServerConfig 验证服务端配置
func (v *Validator) validateServerConfig(config *Config) error {
	// 验证绑定端口
	if config.BindPort != 0 {
		if err := v.validatePort(config.BindPort); err != nil {
			return fmt.Errorf("绑定端口无效: %w", err)
		}
	}

	// 验证 UDP 端口
	if config.BindUDPPort != 0 {
		if err := v.validatePort(config.BindUDPPort); err != nil {
			return fmt.Errorf("UDP 绑定端口无效: %w", err)
		}
	}

	// 验证 KCP 端口
	if config.KCPBindPort != 0 {
		if err := v.validatePort(config.KCPBindPort); err != nil {
			return fmt.Errorf("KCP 绑定端口无效: %w", err)
		}
	}

	// 验证 Web 服务器配置
	if config.WebServer.Port != 0 {
		if err := v.validatePort(config.WebServer.Port); err != nil {
			return fmt.Errorf("Web 服务器端口无效: %w", err)
		}
	}

	// 验证 Web 服务器地址
	if config.WebServer.Addr != "" {
		if err := v.validateAddress(config.WebServer.Addr); err != nil {
			return fmt.Errorf("Web 服务器地址无效: %w", err)
		}
	}

	return nil
}

// validateServerConfigDetailed 详细验证服务端配置
func (v *Validator) validateServerConfigDetailed(config *Config) []string {
	var errors []string

	// 验证绑定端口
	if config.BindPort != 0 {
		if err := v.validatePort(config.BindPort); err != nil {
			errors = append(errors, fmt.Sprintf("绑定端口无效: %v", err))
		}
	}

	// 验证 UDP 端口
	if config.BindUDPPort != 0 {
		if err := v.validatePort(config.BindUDPPort); err != nil {
			errors = append(errors, fmt.Sprintf("UDP 绑定端口无效: %v", err))
		}
	}

	// 验证 KCP 端口
	if config.KCPBindPort != 0 {
		if err := v.validatePort(config.KCPBindPort); err != nil {
			errors = append(errors, fmt.Sprintf("KCP 绑定端口无效: %v", err))
		}
	}

	// 验证 Web 服务器配置
	if config.WebServer.Port != 0 {
		if err := v.validatePort(config.WebServer.Port); err != nil {
			errors = append(errors, fmt.Sprintf("Web 服务器端口无效: %v", err))
		}
	}

	// 验证 Web 服务器地址
	if config.WebServer.Addr != "" {
		if err := v.validateAddress(config.WebServer.Addr); err != nil {
			errors = append(errors, fmt.Sprintf("Web 服务器地址无效: %v", err))
		}
	}

	return errors
}

// validateClientConfig 验证客户端配置
func (v *Validator) validateClientConfig(config *Config) error {
	// 验证服务器地址
	if config.ServerAddr != "" {
		if err := v.validateAddress(config.ServerAddr); err != nil {
			return fmt.Errorf("服务器地址无效: %w", err)
		}
	}

	// 验证服务器端口
	if config.ServerPort != 0 {
		if err := v.validatePort(config.ServerPort); err != nil {
			return fmt.Errorf("服务器端口无效: %w", err)
		}
	}

	return nil
}

// validateClientConfigDetailed 详细验证客户端配置
func (v *Validator) validateClientConfigDetailed(config *Config) []string {
	var errors []string

	// 验证服务器地址
	if config.ServerAddr != "" {
		if err := v.validateAddress(config.ServerAddr); err != nil {
			errors = append(errors, fmt.Sprintf("服务器地址无效: %v", err))
		}
	}

	// 验证服务器端口
	if config.ServerPort != 0 {
		if err := v.validatePort(config.ServerPort); err != nil {
			errors = append(errors, fmt.Sprintf("服务器端口无效: %v", err))
		}
	}

	return errors
}

// validateProxies 验证代理配置
func (v *Validator) validateProxies(proxies []ProxyConfig) error {
	names := make(map[string]bool)

	for i, proxy := range proxies {
		// 验证代理名称
		if err := v.validateProxyName(proxy.Name); err != nil {
			return fmt.Errorf("代理 %d 名称无效: %w", i+1, err)
		}

		// 检查名称重复
		if names[proxy.Name] {
			return fmt.Errorf("代理名称 '%s' 重复", proxy.Name)
		}
		names[proxy.Name] = true

		// 验证代理类型
		if err := v.validateProxyType(proxy.Type); err != nil {
			return fmt.Errorf("代理 '%s' 类型无效: %w", proxy.Name, err)
		}

		// 验证本地地址
		if proxy.LocalIP != "" {
			if err := v.validateAddress(proxy.LocalIP); err != nil {
				return fmt.Errorf("代理 '%s' 本地地址无效: %w", proxy.Name, err)
			}
		}

		// 验证本地端口
		if proxy.LocalPort != 0 {
			if err := v.validatePort(proxy.LocalPort); err != nil {
				return fmt.Errorf("代理 '%s' 本地端口无效: %w", proxy.Name, err)
			}
		}

		// 根据类型验证特定配置
		if err := v.validateProxyByType(proxy); err != nil {
			return fmt.Errorf("代理 '%s' 配置错误: %w", proxy.Name, err)
		}
	}

	return nil
}

// validateProxiesDetailed 详细验证代理配置
func (v *Validator) validateProxiesDetailed(proxies []ProxyConfig) []string {
	var errors []string
	names := make(map[string]bool)

	for i, proxy := range proxies {
		// 验证代理名称
		if err := v.validateProxyName(proxy.Name); err != nil {
			errors = append(errors, fmt.Sprintf("代理 %d 名称无效: %v", i+1, err))
		}

		// 检查名称重复
		if names[proxy.Name] {
			errors = append(errors, fmt.Sprintf("代理名称 '%s' 重复", proxy.Name))
		}
		names[proxy.Name] = true

		// 验证代理类型
		if err := v.validateProxyType(proxy.Type); err != nil {
			errors = append(errors, fmt.Sprintf("代理 '%s' 类型无效: %v", proxy.Name, err))
		}

		// 验证本地地址
		if proxy.LocalIP != "" {
			if err := v.validateAddress(proxy.LocalIP); err != nil {
				errors = append(errors, fmt.Sprintf("代理 '%s' 本地地址无效: %v", proxy.Name, err))
			}
		}

		// 验证本地端口
		if proxy.LocalPort != 0 {
			if err := v.validatePort(proxy.LocalPort); err != nil {
				errors = append(errors, fmt.Sprintf("代理 '%s' 本地端口无效: %v", proxy.Name, err))
			}
		}

		// 根据类型验证特定配置
		if err := v.validateProxyByType(proxy); err != nil {
			errors = append(errors, fmt.Sprintf("代理 '%s' 配置错误: %v", proxy.Name, err))
		}
	}

	return errors
}

// validateVisitors 验证访问者配置
func (v *Validator) validateVisitors(visitors []VisitorConfig) error {
	names := make(map[string]bool)

	for i, visitor := range visitors {
		// 验证访问者名称
		if err := v.validateProxyName(visitor.Name); err != nil {
			return fmt.Errorf("访问者 %d 名称无效: %w", i+1, err)
		}

		// 检查名称重复
		if names[visitor.Name] {
			return fmt.Errorf("访问者名称 '%s' 重复", visitor.Name)
		}
		names[visitor.Name] = true

		// 验证访问者类型
		if err := v.validateVisitorType(visitor.Type); err != nil {
			return fmt.Errorf("访问者 '%s' 类型无效: %w", visitor.Name, err)
		}

		// 验证绑定端口
		if err := v.validatePort(visitor.BindPort); err != nil {
			return fmt.Errorf("访问者 '%s' 绑定端口无效: %w", visitor.Name, err)
		}

		// 验证绑定地址
		if visitor.BindAddr != "" {
			if err := v.validateAddress(visitor.BindAddr); err != nil {
				return fmt.Errorf("访问者 '%s' 绑定地址无效: %w", visitor.Name, err)
			}
		}
	}

	return nil
}

// validateVisitorsDetailed 详细验证访问者配置
func (v *Validator) validateVisitorsDetailed(visitors []VisitorConfig) []string {
	var errors []string
	names := make(map[string]bool)

	for i, visitor := range visitors {
		// 验证访问者名称
		if err := v.validateProxyName(visitor.Name); err != nil {
			errors = append(errors, fmt.Sprintf("访问者 %d 名称无效: %v", i+1, err))
		}

		// 检查名称重复
		if names[visitor.Name] {
			errors = append(errors, fmt.Sprintf("访问者名称 '%s' 重复", visitor.Name))
		}
		names[visitor.Name] = true

		// 验证访问者类型
		if err := v.validateVisitorType(visitor.Type); err != nil {
			errors = append(errors, fmt.Sprintf("访问者 '%s' 类型无效: %v", visitor.Name, err))
		}

		// 验证绑定端口
		if err := v.validatePort(visitor.BindPort); err != nil {
			errors = append(errors, fmt.Sprintf("访问者 '%s' 绑定端口无效: %v", visitor.Name, err))
		}

		// 验证绑定地址
		if visitor.BindAddr != "" {
			if err := v.validateAddress(visitor.BindAddr); err != nil {
				errors = append(errors, fmt.Sprintf("访问者 '%s' 绑定地址无效: %v", visitor.Name, err))
			}
		}
	}

	return errors
}

// validateProxyByType 根据类型验证代理配置
func (v *Validator) validateProxyByType(proxy ProxyConfig) error {
	switch proxy.Type {
	case "tcp", "udp":
		return v.validateTCPUDPProxy(proxy)
	case "http", "https":
		return v.validateHTTPProxy(proxy)
	case "stcp", "sudp", "xtcp":
		return v.validateSecretProxy(proxy)
	default:
		return fmt.Errorf("不支持的代理类型: %s", proxy.Type)
	}
}

// validateTCPUDPProxy 验证 TCP/UDP 代理
func (v *Validator) validateTCPUDPProxy(proxy ProxyConfig) error {
	if proxy.RemotePort == 0 {
		return fmt.Errorf("TCP/UDP 代理必须指定远程端口")
	}

	if err := v.validatePort(proxy.RemotePort); err != nil {
		return fmt.Errorf("远程端口无效: %w", err)
	}

	return nil
}

// validateHTTPProxy 验证 HTTP 代理
func (v *Validator) validateHTTPProxy(proxy ProxyConfig) error {
	// HTTP 代理必须有自定义域名或子域名
	if len(proxy.CustomDomains) == 0 && proxy.Subdomain == "" {
		return fmt.Errorf("HTTP 代理必须指定自定义域名或子域名")
	}

	// 验证自定义域名
	for _, domain := range proxy.CustomDomains {
		if err := v.validateDomain(domain); err != nil {
			return fmt.Errorf("自定义域名 '%s' 无效: %w", domain, err)
		}
	}

	// 验证子域名
	if proxy.Subdomain != "" {
		if err := v.validateSubdomain(proxy.Subdomain); err != nil {
			return fmt.Errorf("子域名无效: %w", err)
		}
	}

	return nil
}

// validateSecretProxy 验证加密代理
func (v *Validator) validateSecretProxy(proxy ProxyConfig) error {
	if proxy.SecretKey == "" {
		return fmt.Errorf("加密代理必须指定密钥")
	}

	if proxy.Type == "xtcp" && proxy.Role == "" {
		return fmt.Errorf("XTCP 代理必须指定角色 (server/visitor)")
	}

	if proxy.Role != "" && proxy.Role != "server" && proxy.Role != "visitor" {
		return fmt.Errorf("角色必须是 'server' 或 'visitor'")
	}

	return nil
}

// validatePort 验证端口号
func (v *Validator) validatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("端口号必须在 1-65535 范围内，当前值: %d", port)
	}
	return nil
}

// validateAddress 验证地址
func (v *Validator) validateAddress(addr string) error {
	// 检查是否为有效的 IP 地址
	if ip := net.ParseIP(addr); ip != nil {
		return nil
	}

	// 检查是否为有效的域名
	if err := v.validateDomain(addr); err == nil {
		return nil
	}

	// 检查是否为 IP:Port 格式
	if host, portStr, err := net.SplitHostPort(addr); err == nil {
		if port, err := strconv.Atoi(portStr); err == nil {
			if err := v.validatePort(port); err != nil {
				return err
			}
			return v.validateAddress(host)
		}
	}

	return fmt.Errorf("无效的地址格式: %s", addr)
}

// validateDomain 验证域名
func (v *Validator) validateDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("域名不能为空")
	}

	// 域名长度限制
	if len(domain) > 253 {
		return fmt.Errorf("域名长度不能超过 253 个字符")
	}

	// 域名格式验证
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	if !domainRegex.MatchString(domain) {
		return fmt.Errorf("域名格式无效: %s", domain)
	}

	return nil
}

// validateSubdomain 验证子域名
func (v *Validator) validateSubdomain(subdomain string) error {
	if subdomain == "" {
		return fmt.Errorf("子域名不能为空")
	}

	// 子域名只能包含字母、数字和连字符
	subdomainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?$`)
	if !subdomainRegex.MatchString(subdomain) {
		return fmt.Errorf("子域名格式无效: %s", subdomain)
	}

	return nil
}

// validateProxyName 验证代理名称
func (v *Validator) validateProxyName(name string) error {
	if name == "" {
		return fmt.Errorf("代理名称不能为空")
	}

	// 代理名称只能包含字母、数字、下划线和连字符
	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
	if !nameRegex.MatchString(name) {
		return fmt.Errorf("代理名称只能包含字母、数字、下划线和连字符: %s", name)
	}

	// 长度限制
	if len(name) > 50 {
		return fmt.Errorf("代理名称长度不能超过 50 个字符")
	}

	return nil
}

// validateProxyType 验证代理类型
func (v *Validator) validateProxyType(proxyType string) error {
	validTypes := []string{"tcp", "udp", "http", "https", "stcp", "sudp", "xtcp"}

	for _, validType := range validTypes {
		if proxyType == validType {
			return nil
		}
	}

	return fmt.Errorf("不支持的代理类型: %s，支持的类型: %s",
		proxyType, strings.Join(validTypes, ", "))
}

// validateVisitorType 验证访问者类型
func (v *Validator) validateVisitorType(visitorType string) error {
	validTypes := []string{"stcp", "sudp", "xtcp"}

	for _, validType := range validTypes {
		if visitorType == validType {
			return nil
		}
	}

	return fmt.Errorf("不支持的访问者类型: %s，支持的类型: %s",
		visitorType, strings.Join(validTypes, ", "))
}

// ValidateProxyConfig 验证单个代理配置
func (v *Validator) ValidateProxyConfig(proxy ProxyConfig) error {
	// 验证代理名称
	if err := v.validateProxyName(proxy.Name); err != nil {
		return fmt.Errorf("代理名称无效: %w", err)
	}

	// 验证代理类型
	if err := v.validateProxyType(proxy.Type); err != nil {
		return fmt.Errorf("代理类型无效: %w", err)
	}

	// 验证本地地址
	if proxy.LocalIP != "" {
		if err := v.validateAddress(proxy.LocalIP); err != nil {
			return fmt.Errorf("本地地址无效: %w", err)
		}
	}

	// 验证本地端口
	if proxy.LocalPort != 0 {
		if err := v.validatePort(proxy.LocalPort); err != nil {
			return fmt.Errorf("本地端口无效: %w", err)
		}
	}

	// 根据类型验证特定配置
	return v.validateProxyByType(proxy)
}

// GetValidationSummary 获取验证摘要
func (v *Validator) GetValidationSummary(config *Config) map[string][]string {
	summary := map[string][]string{
		"errors":   {},
		"warnings": {},
		"info":     {},
	}

	// 检查基本配置
	if config.ServerAddr == "" && len(config.Proxies) > 0 {
		summary["warnings"] = append(summary["warnings"], "客户端配置中未指定服务器地址")
	}

	if config.Token == "" {
		summary["warnings"] = append(summary["warnings"], "未设置认证令牌，建议设置以提高安全性")
	}

	// 检查代理配置
	if len(config.Proxies) == 0 {
		summary["info"] = append(summary["info"], "未配置任何代理")
	} else {
		summary["info"] = append(summary["info"], fmt.Sprintf("配置了 %d 个代理", len(config.Proxies)))
	}

	// 检查端口冲突
	ports := make(map[int][]string)
	for _, proxy := range config.Proxies {
		if proxy.RemotePort != 0 {
			ports[proxy.RemotePort] = append(ports[proxy.RemotePort], proxy.Name)
		}
	}

	for port, names := range ports {
		if len(names) > 1 {
			summary["errors"] = append(summary["errors"],
				fmt.Sprintf("端口 %d 被多个代理使用: %s", port, strings.Join(names, ", ")))
		}
	}

	return summary
}

// CompareConfigs 比较两个配置的差异
func (v *Validator) CompareConfigs(config1, config2 *Config) map[string][]string {
	differences := map[string][]string{
		"basic":    {},
		"proxies":  {},
		"visitors": {},
		"web":      {},
		"log":      {},
	}

	if config1 == nil || config2 == nil {
		differences["basic"] = append(differences["basic"], "其中一个配置为空")
		return differences
	}

	// 比较基本配置
	if config1.ServerAddr != config2.ServerAddr {
		differences["basic"] = append(differences["basic"],
			fmt.Sprintf("服务器地址: '%s' vs '%s'", config1.ServerAddr, config2.ServerAddr))
	}
	if config1.ServerPort != config2.ServerPort {
		differences["basic"] = append(differences["basic"],
			fmt.Sprintf("服务器端口: %d vs %d", config1.ServerPort, config2.ServerPort))
	}
	if config1.BindPort != config2.BindPort {
		differences["basic"] = append(differences["basic"],
			fmt.Sprintf("绑定端口: %d vs %d", config1.BindPort, config2.BindPort))
	}
	if config1.Token != config2.Token {
		differences["basic"] = append(differences["basic"], "认证令牌不同")
	}

	// 比较 Web 服务器配置
	if config1.WebServer.Port != config2.WebServer.Port {
		differences["web"] = append(differences["web"],
			fmt.Sprintf("Web 端口: %d vs %d", config1.WebServer.Port, config2.WebServer.Port))
	}
	if config1.WebServer.User != config2.WebServer.User {
		differences["web"] = append(differences["web"],
			fmt.Sprintf("Web 用户名: '%s' vs '%s'", config1.WebServer.User, config2.WebServer.User))
	}
	if config1.WebServer.Password != config2.WebServer.Password {
		differences["web"] = append(differences["web"], "Web 密码不同")
	}

	// 比较日志配置
	if config1.Log.Level != config2.Log.Level {
		differences["log"] = append(differences["log"],
			fmt.Sprintf("日志级别: '%s' vs '%s'", config1.Log.Level, config2.Log.Level))
	}
	if config1.Log.To != config2.Log.To {
		differences["log"] = append(differences["log"],
			fmt.Sprintf("日志输出: '%s' vs '%s'", config1.Log.To, config2.Log.To))
	}

	// 比较代理配置
	proxy1Map := make(map[string]ProxyConfig)
	proxy2Map := make(map[string]ProxyConfig)

	for _, proxy := range config1.Proxies {
		proxy1Map[proxy.Name] = proxy
	}
	for _, proxy := range config2.Proxies {
		proxy2Map[proxy.Name] = proxy
	}

	// 检查配置1中存在但配置2中不存在的代理
	for name := range proxy1Map {
		if _, exists := proxy2Map[name]; !exists {
			differences["proxies"] = append(differences["proxies"],
				fmt.Sprintf("代理 '%s' 仅在配置1中存在", name))
		}
	}

	// 检查配置2中存在但配置1中不存在的代理
	for name := range proxy2Map {
		if _, exists := proxy1Map[name]; !exists {
			differences["proxies"] = append(differences["proxies"],
				fmt.Sprintf("代理 '%s' 仅在配置2中存在", name))
		}
	}

	// 检查同名代理的差异
	for name, proxy1 := range proxy1Map {
		if proxy2, exists := proxy2Map[name]; exists {
			if proxy1.Type != proxy2.Type {
				differences["proxies"] = append(differences["proxies"],
					fmt.Sprintf("代理 '%s' 类型不同: %s vs %s", name, proxy1.Type, proxy2.Type))
			}
			if proxy1.LocalPort != proxy2.LocalPort {
				differences["proxies"] = append(differences["proxies"],
					fmt.Sprintf("代理 '%s' 本地端口不同: %d vs %d", name, proxy1.LocalPort, proxy2.LocalPort))
			}
			if proxy1.RemotePort != proxy2.RemotePort {
				differences["proxies"] = append(differences["proxies"],
					fmt.Sprintf("代理 '%s' 远程端口不同: %d vs %d", name, proxy1.RemotePort, proxy2.RemotePort))
			}
		}
	}

	// 比较访问者配置
	visitor1Map := make(map[string]VisitorConfig)
	visitor2Map := make(map[string]VisitorConfig)

	for _, visitor := range config1.Visitors {
		visitor1Map[visitor.Name] = visitor
	}
	for _, visitor := range config2.Visitors {
		visitor2Map[visitor.Name] = visitor
	}

	// 检查访问者差异
	for name := range visitor1Map {
		if _, exists := visitor2Map[name]; !exists {
			differences["visitors"] = append(differences["visitors"],
				fmt.Sprintf("访问者 '%s' 仅在配置1中存在", name))
		}
	}

	for name := range visitor2Map {
		if _, exists := visitor1Map[name]; !exists {
			differences["visitors"] = append(differences["visitors"],
				fmt.Sprintf("访问者 '%s' 仅在配置2中存在", name))
		}
	}

	return differences
}

// GetConfigSummary 获取配置摘要
func (v *Validator) GetConfigSummary(config *Config) map[string]interface{} {
	summary := map[string]interface{}{
		"type":         "unknown",
		"proxies":      len(config.Proxies),
		"visitors":     len(config.Visitors),
		"hasToken":     config.Token != "",
		"hasWebServer": config.WebServer.Port > 0,
		"logLevel":     config.Log.Level,
	}

	// 检测配置类型
	if config.BindPort > 0 || config.WebServer.Port > 0 {
		summary["type"] = "server"
		summary["bindPort"] = config.BindPort
		summary["webPort"] = config.WebServer.Port
	} else if config.ServerAddr != "" || len(config.Proxies) > 0 {
		summary["type"] = "client"
		summary["serverAddr"] = config.ServerAddr
		summary["serverPort"] = config.ServerPort
	}

	// 代理类型统计
	proxyTypes := make(map[string]int)
	for _, proxy := range config.Proxies {
		proxyTypes[proxy.Type]++
	}
	summary["proxyTypes"] = proxyTypes

	return summary
}
