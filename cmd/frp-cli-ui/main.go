package main

import (
	"log"
	"os"

	"frp-cli-ui/internal/installer"
	"frp-cli-ui/pkg/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)

func main() {
	// 设置字符宽度计算
	runewidth.DefaultCondition.EastAsianWidth = false

	// 检查 FRP 安装状态（可选操作）
	inst := installer.NewInstaller("")
	_, _ = inst.CheckInstallation() // 忽略错误，仅作为检查

	// 使用新架构创建主控制面板
	initialModel := ui.NewMainDashboard()

	// 初始化 TUI 程序，Bubble Tea 默认已支持 Ctrl+Z 挂起和信号处理
	p := tea.NewProgram(
		initialModel,
		tea.WithAltScreen(),
	)

	// 启动 TUI
	if _, err := p.Run(); err != nil {
		log.Printf("FRP CLI UI 启动失败: %v", err)
		os.Exit(1)
	}
}
