package config

import (
	"fmt"
	"net"
	"regexp"
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
	if config == nil {
		return []string{"配置不能为空"}
	}

	var errors []string
	errors = append(errors, v.validateServerConfigDetailed(config)...)
	errors = append(errors, v.validateClientConfigDetailed(config)...)
	errors = append(errors, v.validateProxiesDetailed(config.Proxies)...)
	errors = append(errors, v.validateVisitorsDetailed(config.Visitors)...)

	return errors
}

// validateServerConfig 验证服务端配置
func (v *Validator) validateServerConfig(config *Config) error {
	ports := map[string]int{
		"绑定端口":     config.BindPort,
		"UDP端口":    config.BindUDPPort,
		"KCP端口":    config.KCPBindPort,
		"Web服务器端口": config.WebServer.Port,
	}

	for name, port := range ports {
		if port != 0 {
			if err := v.validatePort(port); err != nil {
				return fmt.Errorf("%s无效: %w", name, err)
			}
		}
	}

	if config.WebServer.Addr != "" {
		if err := v.validateAddress(config.WebServer.Addr); err != nil {
			return fmt.Errorf("Web服务器地址无效: %w", err)
		}
	}

	return nil
}

// validateServerConfigDetailed 详细验证服务端配置
func (v *Validator) validateServerConfigDetailed(config *Config) []string {
	var errors []string

	ports := map[string]int{
		"绑定端口":     config.BindPort,
		"UDP端口":    config.BindUDPPort,
		"KCP端口":    config.KCPBindPort,
		"Web服务器端口": config.WebServer.Port,
	}

	for name, port := range ports {
		if port != 0 {
			if err := v.validatePort(port); err != nil {
				errors = append(errors, fmt.Sprintf("%s无效: %v", name, err))
			}
		}
	}

	if config.WebServer.Addr != "" {
		if err := v.validateAddress(config.WebServer.Addr); err != nil {
			errors = append(errors, fmt.Sprintf("Web服务器地址无效: %v", err))
		}
	}

	return errors
}

// validateClientConfig 验证客户端配置
func (v *Validator) validateClientConfig(config *Config) error {
	if config.ServerAddr != "" {
		if err := v.validateAddress(config.ServerAddr); err != nil {
			return fmt.Errorf("服务器地址无效: %w", err)
		}
	}

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

	if config.ServerAddr != "" {
		if err := v.validateAddress(config.ServerAddr); err != nil {
			errors = append(errors, fmt.Sprintf("服务器地址无效: %v", err))
		}
	}

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
		if err := v.validateProxyName(proxy.Name); err != nil {
			return fmt.Errorf("代理 %d 名称无效: %w", i+1, err)
		}

		if names[proxy.Name] {
			return fmt.Errorf("代理名称 '%s' 重复", proxy.Name)
		}
		names[proxy.Name] = true

		if err := v.validateProxyType(proxy.Type); err != nil {
			return fmt.Errorf("代理 '%s' 类型无效: %w", proxy.Name, err)
		}

		if proxy.LocalIP != "" {
			if err := v.validateAddress(proxy.LocalIP); err != nil {
				return fmt.Errorf("代理 '%s' 本地地址无效: %w", proxy.Name, err)
			}
		}

		if proxy.LocalPort != 0 {
			if err := v.validatePort(proxy.LocalPort); err != nil {
				return fmt.Errorf("代理 '%s' 本地端口无效: %w", proxy.Name, err)
			}
		}

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
		if err := v.validateProxyName(proxy.Name); err != nil {
			errors = append(errors, fmt.Sprintf("代理 %d 名称无效: %v", i+1, err))
		}

		if names[proxy.Name] {
			errors = append(errors, fmt.Sprintf("代理名称 '%s' 重复", proxy.Name))
		}
		names[proxy.Name] = true

		if err := v.validateProxyType(proxy.Type); err != nil {
			errors = append(errors, fmt.Sprintf("代理 '%s' 类型无效: %v", proxy.Name, err))
		}

		if proxy.LocalIP != "" {
			if err := v.validateAddress(proxy.LocalIP); err != nil {
				errors = append(errors, fmt.Sprintf("代理 '%s' 本地地址无效: %v", proxy.Name, err))
			}
		}

		if proxy.LocalPort != 0 {
			if err := v.validatePort(proxy.LocalPort); err != nil {
				errors = append(errors, fmt.Sprintf("代理 '%s' 本地端口无效: %v", proxy.Name, err))
			}
		}

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
		if visitor.Name == "" {
			return fmt.Errorf("访问者 %d 名称不能为空", i+1)
		}

		if names[visitor.Name] {
			return fmt.Errorf("访问者名称 '%s' 重复", visitor.Name)
		}
		names[visitor.Name] = true

		if err := v.validateVisitorType(visitor.Type); err != nil {
			return fmt.Errorf("访问者 '%s' 类型无效: %w", visitor.Name, err)
		}

		if visitor.BindPort != 0 {
			if err := v.validatePort(visitor.BindPort); err != nil {
				return fmt.Errorf("访问者 '%s' 绑定端口无效: %w", visitor.Name, err)
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
		if visitor.Name == "" {
			errors = append(errors, fmt.Sprintf("访问者 %d 名称不能为空", i+1))
		}

		if names[visitor.Name] {
			errors = append(errors, fmt.Sprintf("访问者名称 '%s' 重复", visitor.Name))
		}
		names[visitor.Name] = true

		if err := v.validateVisitorType(visitor.Type); err != nil {
			errors = append(errors, fmt.Sprintf("访问者 '%s' 类型无效: %v", visitor.Name, err))
		}

		if visitor.BindPort != 0 {
			if err := v.validatePort(visitor.BindPort); err != nil {
				errors = append(errors, fmt.Sprintf("访问者 '%s' 绑定端口无效: %v", visitor.Name, err))
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
	}
	return nil
}

// validateTCPUDPProxy 验证 TCP/UDP 代理
func (v *Validator) validateTCPUDPProxy(proxy ProxyConfig) error {
	if proxy.RemotePort == 0 {
		return fmt.Errorf("远程端口不能为空")
	}
	return v.validatePort(proxy.RemotePort)
}

// validateHTTPProxy 验证 HTTP 代理
func (v *Validator) validateHTTPProxy(proxy ProxyConfig) error {
	if len(proxy.CustomDomains) == 0 && proxy.Subdomain == "" {
		return fmt.Errorf("HTTP代理必须设置自定义域名或子域名")
	}

	for _, domain := range proxy.CustomDomains {
		if err := v.validateDomain(domain); err != nil {
			return fmt.Errorf("自定义域名无效: %w", err)
		}
	}

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
		return fmt.Errorf("密钥不能为空")
	}
	if len(proxy.SecretKey) < 8 {
		return fmt.Errorf("密钥长度不能少于8位")
	}
	return nil
}

// validatePort 验证端口号
func (v *Validator) validatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("端口必须在 1-65535 范围内")
	}
	return nil
}

// validateAddress 验证地址
func (v *Validator) validateAddress(addr string) error {
	if addr == "" {
		return fmt.Errorf("地址不能为空")
	}

	if net.ParseIP(addr) != nil {
		return nil
	}

	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9.-]+$`, addr); !matched {
		return fmt.Errorf("地址格式无效")
	}

	parts := strings.Split(addr, ".")
	if len(parts) < 2 {
		return fmt.Errorf("域名格式无效")
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 {
			return fmt.Errorf("域名部分长度无效")
		}
	}

	return nil
}

// validateDomain 验证域名
func (v *Validator) validateDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("域名不能为空")
	}

	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, domain); !matched {
		return fmt.Errorf("域名格式无效")
	}

	parts := strings.Split(domain, ".")
	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 {
			return fmt.Errorf("域名部分长度无效")
		}
	}

	return nil
}

// validateSubdomain 验证子域名
func (v *Validator) validateSubdomain(subdomain string) error {
	if subdomain == "" {
		return fmt.Errorf("子域名不能为空")
	}

	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9-]+$`, subdomain); !matched {
		return fmt.Errorf("子域名只能包含字母、数字和连字符")
	}

	if len(subdomain) > 63 {
		return fmt.Errorf("子域名长度不能超过63个字符")
	}

	return nil
}

// validateProxyName 验证代理名称
func (v *Validator) validateProxyName(name string) error {
	if name == "" {
		return fmt.Errorf("代理名称不能为空")
	}

	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name); !matched {
		return fmt.Errorf("代理名称只能包含字母、数字、下划线和连字符")
	}

	if len(name) > 50 {
		return fmt.Errorf("代理名称长度不能超过50个字符")
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
	return fmt.Errorf("无效的代理类型: %s", proxyType)
}

// validateVisitorType 验证访问者类型
func (v *Validator) validateVisitorType(visitorType string) error {
	validTypes := []string{"stcp", "sudp", "xtcp"}
	for _, validType := range validTypes {
		if visitorType == validType {
			return nil
		}
	}
	return fmt.Errorf("无效的访问者类型: %s", visitorType)
}

// ValidateProxyConfig 验证单个代理配置
func (v *Validator) ValidateProxyConfig(proxy ProxyConfig) error {
	if err := v.validateProxyName(proxy.Name); err != nil {
		return err
	}

	if err := v.validateProxyType(proxy.Type); err != nil {
		return err
	}

	if proxy.LocalIP != "" {
		if err := v.validateAddress(proxy.LocalIP); err != nil {
			return fmt.Errorf("本地地址无效: %w", err)
		}
	}

	if proxy.LocalPort != 0 {
		if err := v.validatePort(proxy.LocalPort); err != nil {
			return fmt.Errorf("本地端口无效: %w", err)
		}
	}

	return v.validateProxyByType(proxy)
}

// GetValidationSummary 获取验证摘要
func (v *Validator) GetValidationSummary(config *Config) map[string][]string {
	summary := make(map[string][]string)

	if config == nil {
		summary["error"] = []string{"配置不能为空"}
		return summary
	}

	if serverErrors := v.validateServerConfigDetailed(config); len(serverErrors) > 0 {
		summary["server"] = serverErrors
	}

	if clientErrors := v.validateClientConfigDetailed(config); len(clientErrors) > 0 {
		summary["client"] = clientErrors
	}

	if proxyErrors := v.validateProxiesDetailed(config.Proxies); len(proxyErrors) > 0 {
		summary["proxies"] = proxyErrors
	}

	if visitorErrors := v.validateVisitorsDetailed(config.Visitors); len(visitorErrors) > 0 {
		summary["visitors"] = visitorErrors
	}

	return summary
}

// CompareConfigs 比较两个配置的差异
func (v *Validator) CompareConfigs(config1, config2 *Config) map[string][]string {
	differences := make(map[string][]string)

	if config1 == nil || config2 == nil {
		differences["error"] = []string{"配置不能为空"}
		return differences
	}

	if config1.ServerAddr != config2.ServerAddr {
		differences["server"] = append(differences["server"],
			fmt.Sprintf("服务器地址: %s -> %s", config1.ServerAddr, config2.ServerAddr))
	}

	if config1.ServerPort != config2.ServerPort {
		differences["server"] = append(differences["server"],
			fmt.Sprintf("服务器端口: %d -> %d", config1.ServerPort, config2.ServerPort))
	}

	if config1.BindPort != config2.BindPort {
		differences["server"] = append(differences["server"],
			fmt.Sprintf("绑定端口: %d -> %d", config1.BindPort, config2.BindPort))
	}

	if len(config1.Proxies) != len(config2.Proxies) {
		differences["proxies"] = append(differences["proxies"],
			fmt.Sprintf("代理数量: %d -> %d", len(config1.Proxies), len(config2.Proxies)))
	}

	proxyMap1 := make(map[string]ProxyConfig)
	for _, proxy := range config1.Proxies {
		proxyMap1[proxy.Name] = proxy
	}

	for _, proxy2 := range config2.Proxies {
		if proxy1, exists := proxyMap1[proxy2.Name]; exists {
			if proxy1.Type != proxy2.Type {
				differences["proxies"] = append(differences["proxies"],
					fmt.Sprintf("代理 %s 类型: %s -> %s", proxy2.Name, proxy1.Type, proxy2.Type))
			}
			if proxy1.LocalPort != proxy2.LocalPort {
				differences["proxies"] = append(differences["proxies"],
					fmt.Sprintf("代理 %s 本地端口: %d -> %d", proxy2.Name, proxy1.LocalPort, proxy2.LocalPort))
			}
		} else {
			differences["proxies"] = append(differences["proxies"],
				fmt.Sprintf("新增代理: %s", proxy2.Name))
		}
	}

	return differences
}

// GetConfigSummary 获取配置摘要
func (v *Validator) GetConfigSummary(config *Config) map[string]interface{} {
	summary := make(map[string]interface{})

	if config == nil {
		summary["error"] = "配置为空"
		return summary
	}

	summary["server_addr"] = config.ServerAddr
	summary["server_port"] = config.ServerPort
	summary["bind_port"] = config.BindPort
	summary["proxy_count"] = len(config.Proxies)
	summary["visitor_count"] = len(config.Visitors)

	proxyTypes := make(map[string]int)
	for _, proxy := range config.Proxies {
		proxyTypes[proxy.Type]++
	}
	summary["proxy_types"] = proxyTypes

	return summary
}
