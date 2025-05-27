package main

import (
	"fmt"
	"log"
	"os"

	"frp-cli-ui/service"
	"frp-cli-ui/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)

func main() {
	// 将东亚模糊宽度字符视为窄字符
	// 这通常能解决在某些终端或 locale 设置下，
	// 表格线等字符因宽度计算不一致导致的错位问题。
	runewidth.DefaultCondition.EastAsianWidth = false

	// 检查 FRP 安装状态
	installer := service.NewInstaller("")
	status, err := installer.CheckInstallation()

	var initialModel tea.Model

	if err != nil {
		log.Printf("检查 FRP 安装状态时出错: %v", err)
		// 出错时显示安装界面
		initialModel = tui.NewInstallerView()
	} else if !status.IsInstalled {
		// FRP 未安装，显示安装界面
		initialModel = tui.NewInstallerView()
	} else {
		// FRP 已安装，直接进入主界面
		initialModel = tui.NewDashboard()
	}

	// 初始化 TUI 程序
	p := tea.NewProgram(
		initialModel,
		tea.WithAltScreen(),
		// tea.WithMouseCellMotion(),
		// tea.WithMouseAllMotion(),
		// tea.WithInputTTY(),
	)

	// 启动 TUI
	if _, err := p.Run(); err != nil {
		log.Fatal(fmt.Errorf("启动 FRP CLI UI 失败: %w", err))
		os.Exit(1)
	}
}
