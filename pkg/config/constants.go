package config

import (
	"os"
	"path/filepath"
)

const (
	AppName    = "FRP Manager"
	AppVersion = "1.0.0"
)

// GetDefaultWorkDir 获取默认工作目录
func GetDefaultWorkDir() string {
	// 优先使用用户主目录下的 .frp-manager 目录
	if homeDir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(homeDir, ".frp-manager")
	}
	// 如果获取不到主目录，使用当前目录下的 .frp-manager
	return ".frp-manager"
}
