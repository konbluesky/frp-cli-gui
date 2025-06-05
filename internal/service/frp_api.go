package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIClient FRP API 客户端
type APIClient struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

// ProxyInfo 代理信息（匹配FRP实际API响应）
type ProxyInfo struct {
	Name            string    `json:"name"`
	Conf            ProxyConf `json:"conf"`
	ClientVersion   string    `json:"clientVersion"`
	TodayTrafficIn  int64     `json:"todayTrafficIn"`
	TodayTrafficOut int64     `json:"todayTrafficOut"`
	CurConns        int       `json:"curConns"`
	LastStartTime   string    `json:"lastStartTime"`
	LastCloseTime   string    `json:"lastCloseTime"`
	Status          string    `json:"status"`
}

// ProxyConf 代理配置信息
type ProxyConf struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	LocalIP    string `json:"localIP"`
	RemotePort int    `json:"remotePort"`
	// FRP API中有很多额外字段，但我们主要需要这些
	Transport    map[string]string      `json:"transport"`
	LoadBalancer map[string]string      `json:"loadBalancer"`
	HealthCheck  map[string]interface{} `json:"healthCheck"`
	Plugin       map[string]interface{} `json:"plugin"`
}

// ServerInfo 服务器信息
type ServerInfo struct {
	Version               string         `json:"version"`
	BindPort              int            `json:"bind_port"`
	BindUDPPort           int            `json:"bind_udp_port"`
	VhostHTTPPort         int            `json:"vhost_http_port"`
	VhostHTTPSPort        int            `json:"vhost_https_port"`
	TCPMuxHTTPConnectPort int            `json:"tcpmux_httpconnect_port"`
	KCPBindPort           int            `json:"kcp_bind_port"`
	SubdomainHost         string         `json:"subdomain_host"`
	MaxPoolCount          int            `json:"max_pool_count"`
	MaxPortsPerClient     int            `json:"max_ports_per_client"`
	HeartBeatTimeout      int            `json:"heart_beat_timeout"`
	AllowPortsStr         string         `json:"allow_ports_str"`
	TotalTrafficIn        int64          `json:"total_traffic_in"`
	TotalTrafficOut       int64          `json:"total_traffic_out"`
	CurConns              int            `json:"cur_conns"`
	ClientCounts          int            `json:"client_counts"`
	ProxyTypeCounts       map[string]int `json:"proxy_type_counts"`
}

// ClientInfo 客户端信息
type ClientInfo struct {
	Version              string `json:"version"`
	Hostname             string `json:"hostname"`
	OS                   string `json:"os"`
	Arch                 string `json:"arch"`
	User                 string `json:"user"`
	Privilege            string `json:"privilege"`
	RunID                string `json:"run_id"`
	PoolCount            int    `json:"pool_count"`
	ProxyNum             int    `json:"proxy_num"`
	ConnectServerLocalIP string `json:"connect_server_local_ip"`
	LastStartTime        string `json:"last_start_time"`
	LastCloseTime        string `json:"last_close_time"`
}

// TrafficInfo 流量信息
type TrafficInfo struct {
	Name       string `json:"name"`
	TrafficIn  int64  `json:"traffic_in"`
	TrafficOut int64  `json:"traffic_out"`
}

// NewAPIClient 创建新的 API 客户端
func NewAPIClient(baseURL, username, password string) *APIClient {
	return &APIClient{
		baseURL:  baseURL,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// makeRequest 发送 HTTP 请求
func (c *APIClient) makeRequest(endpoint string) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 添加基本认证
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 请求失败，状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// GetServerInfo 获取服务器信息
func (c *APIClient) GetServerInfo() (*ServerInfo, error) {
	data, err := c.makeRequest("/api/serverinfo")
	if err != nil {
		return nil, fmt.Errorf("获取服务器信息失败: %w", err)
	}

	var serverInfo ServerInfo
	if err := json.Unmarshal(data, &serverInfo); err != nil {
		return nil, fmt.Errorf("解析服务器信息失败: %w", err)
	}

	return &serverInfo, nil
}

// GetProxyList 获取所有类型的代理列表
func (c *APIClient) GetProxyList() ([]ProxyInfo, error) {
	// FRP API需要按类型分别查询，常见的代理类型包括：
	proxyTypes := []string{"tcp", "http", "https", "stcp", "sudp", "udp", "xtcp"}
	var allProxies []ProxyInfo

	for _, proxyType := range proxyTypes {
		proxies, err := c.getProxyListByType(proxyType)
		if err != nil {
			// 如果某个类型查询失败，记录但不中断整个查询
			continue
		}
		allProxies = append(allProxies, proxies...)
	}

	return allProxies, nil
}

// getProxyListByType 按类型获取代理列表
func (c *APIClient) getProxyListByType(proxyType string) ([]ProxyInfo, error) {
	endpoint := fmt.Sprintf("/api/proxy/%s", proxyType)
	data, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("获取%s类型代理失败: %w", proxyType, err)
	}

	var response struct {
		Proxies []ProxyInfo `json:"proxies"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("解析%s类型代理失败: %w", proxyType, err)
	}

	return response.Proxies, nil
}

// GetProxyInfo 获取特定代理信息
func (c *APIClient) GetProxyInfo(name string) (*ProxyInfo, error) {
	endpoint := fmt.Sprintf("/api/proxy/%s", name)
	data, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("获取代理信息失败: %w", err)
	}

	var proxyInfo ProxyInfo
	if err := json.Unmarshal(data, &proxyInfo); err != nil {
		return nil, fmt.Errorf("解析代理信息失败: %w", err)
	}

	return &proxyInfo, nil
}

// GetClientList 获取客户端列表
func (c *APIClient) GetClientList() ([]ClientInfo, error) {
	data, err := c.makeRequest("/api/client")
	if err != nil {
		return nil, fmt.Errorf("获取客户端列表失败: %w", err)
	}

	var response struct {
		Clients []ClientInfo `json:"clients"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("解析客户端列表失败: %w", err)
	}

	return response.Clients, nil
}

// GetTrafficInfo 获取流量信息
func (c *APIClient) GetTrafficInfo() ([]TrafficInfo, error) {
	data, err := c.makeRequest("/api/traffic")
	if err != nil {
		return nil, fmt.Errorf("获取流量信息失败: %w", err)
	}

	var response struct {
		Traffic []TrafficInfo `json:"traffic"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("解析流量信息失败: %w", err)
	}

	return response.Traffic, nil
}

// CloseProxy 关闭代理
func (c *APIClient) CloseProxy(name string) error {
	url := fmt.Sprintf("%s/api/proxy/%s", c.baseURL, name)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 添加基本认证
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("关闭代理失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// ReloadConfig 重新加载配置
func (c *APIClient) ReloadConfig() error {
	url := fmt.Sprintf("%s/api/reload", c.baseURL)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 添加基本认证
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("重新加载配置失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// IsServerReachable 检查服务器是否可达
func (c *APIClient) IsServerReachable() bool {
	_, err := c.GetServerInfo()
	return err == nil
}

// GetConnectionStats 获取连接统计信息
func (c *APIClient) GetConnectionStats() (map[string]interface{}, error) {
	data, err := c.makeRequest("/api/status")
	if err != nil {
		return nil, fmt.Errorf("获取连接统计失败: %w", err)
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("解析连接统计失败: %w", err)
	}

	return stats, nil
}

// FormatTraffic 格式化流量显示
func FormatTraffic(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetProxyStatus 获取代理状态摘要
func (c *APIClient) GetProxyStatus() (map[string]int, error) {
	proxies, err := c.GetProxyList()
	if err != nil {
		return nil, err
	}

	status := map[string]int{
		"running": 0,
		"stopped": 0,
		"error":   0,
	}

	for _, proxy := range proxies {
		switch proxy.Status {
		case "running":
			status["running"]++
		case "stopped":
			status["stopped"]++
		default:
			status["error"]++
		}
	}

	return status, nil
}
