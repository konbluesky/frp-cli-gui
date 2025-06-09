package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// InitializeWorkspace 初始化工作空间
func InitializeWorkspace() error {
	workDir := GetDefaultWorkDir()

	// 创建工作目录
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("创建工作目录失败: %w", err)
	}

	// 创建配置文件目录
	configDir := filepath.Join(workDir, "configs")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 创建日志目录
	logDir := filepath.Join(workDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 创建默认配置文件
	serverConfigPath := filepath.Join(configDir, "frps.toml")
	clientConfigPath := filepath.Join(configDir, "frpc.toml")

	// 如果配置文件不存在，创建默认配置文件
	if _, err := os.Stat(serverConfigPath); os.IsNotExist(err) {
		if err := os.WriteFile(serverConfigPath, []byte(DefaultServerConfigTemplate), 0644); err != nil {
			return fmt.Errorf("创建默认服务端配置文件失败: %w", err)
		}
	}

	if _, err := os.Stat(clientConfigPath); os.IsNotExist(err) {
		if err := os.WriteFile(clientConfigPath, []byte(DefaultClientConfigTemplate), 0644); err != nil {
			return fmt.Errorf("创建默认客户端配置文件失败: %w", err)
		}
	}

	return nil
}

// GetDefaultServerConfigPath 获取默认服务端配置文件路径
func GetDefaultServerConfigPath() string {
	return filepath.Join(GetDefaultWorkDir(), "configs", "frps.toml")
}

// GetDefaultClientConfigPath 获取默认客户端配置文件路径
func GetDefaultClientConfigPath() string {
	return filepath.Join(GetDefaultWorkDir(), "configs", "frpc.toml")
}

// EnsureWorkspaceExists 确保工作空间存在
func EnsureWorkspaceExists() error {
	workDir := GetDefaultWorkDir()
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		return InitializeWorkspace()
	}
	return nil
}
