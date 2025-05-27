package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"frp-cli-ui/service"
)

// InstallerView 安装界面
type InstallerView struct {
	width     int
	height    int
	installer *service.Installer
	status    *service.InstallStatus
	progress  progress.Model
	spinner   spinner.Model
	state     InstallerState
	message   string
	error     error
}

// InstallerState 安装状态
type InstallerState int

const (
	StateChecking InstallerState = iota
	StateNotInstalled
	StateInstalled
	StateInstalling
	StateInstallSuccess
	StateInstallError
	StateUpdating
	StateUpdateSuccess
	StateUpdateError
)

// installerMsg 安装消息类型
type installerMsg struct {
	state   InstallerState
	message string
	error   error
	status  *service.InstallStatus
}

// NewInstallerView 创建新的安装界面
func NewInstallerView() InstallerView {
	// 初始化进度条
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 60

	// 初始化加载动画
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return InstallerView{
		installer: service.NewInstaller(""),
		progress:  prog,
		spinner:   s,
		state:     StateChecking,
		message:   "正在检查 FRP 安装状态...",
	}
}

// Init 初始化
func (m InstallerView) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.checkInstallation,
	)
}

// checkInstallation 检查安装状态
func (m InstallerView) checkInstallation() tea.Msg {
	status, err := m.installer.CheckInstallation()
	if err != nil {
		return installerMsg{
			state:   StateInstallError,
			message: "检查安装状态失败",
			error:   err,
		}
	}

	if status.IsInstalled {
		if status.NeedsUpdate {
			return installerMsg{
				state:   StateInstalled,
				message: fmt.Sprintf("FRP 已安装 (版本 %s)，有新版本 %s 可用", status.Version, status.LatestVersion),
				status:  status,
			}
		} else {
			return installerMsg{
				state:   StateInstalled,
				message: fmt.Sprintf("FRP 已安装 (版本 %s)", status.Version),
				status:  status,
			}
		}
	} else {
		return installerMsg{
			state:   StateNotInstalled,
			message: "FRP 未安装",
			status:  status,
		}
	}
}

// installFRP 安装 FRP
func (m InstallerView) installFRP() tea.Msg {
	err := m.installer.InstallFRP()
	if err != nil {
		return installerMsg{
			state:   StateInstallError,
			message: "安装失败",
			error:   err,
		}
	}

	// 重新检查安装状态
	status, err := m.installer.CheckInstallation()
	if err != nil {
		return installerMsg{
			state:   StateInstallError,
			message: "验证安装失败",
			error:   err,
		}
	}

	return installerMsg{
		state:   StateInstallSuccess,
		message: fmt.Sprintf("FRP 安装成功！版本: %s", status.Version),
		status:  status,
	}
}

// updateFRP 更新 FRP
func (m InstallerView) updateFRP() tea.Msg {
	err := m.installer.UpdateFRP()
	if err != nil {
		return installerMsg{
			state:   StateUpdateError,
			message: "更新失败",
			error:   err,
		}
	}

	// 重新检查安装状态
	status, err := m.installer.CheckInstallation()
	if err != nil {
		return installerMsg{
			state:   StateUpdateError,
			message: "验证更新失败",
			error:   err,
		}
	}

	return installerMsg{
		state:   StateUpdateSuccess,
		message: fmt.Sprintf("FRP 更新成功！版本: %s", status.Version),
		status:  status,
	}
}

// Update 更新状态
func (m InstallerView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 20

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter", " ":
			switch m.state {
			case StateNotInstalled:
				m.state = StateInstalling
				m.message = "正在下载和安装 FRP..."
				return m, m.installFRP

			case StateInstalled:
				if m.status != nil && m.status.NeedsUpdate {
					m.state = StateUpdating
					m.message = "正在更新 FRP..."
					return m, m.updateFRP
				} else {
					// FRP已安装且是最新版本，直接进入主界面
					dashboard := NewDashboardWithSize(m.width, m.height)
					return dashboard, tea.Batch(
						dashboard.Init(),
						func() tea.Msg {
							return tea.WindowSizeMsg{Width: m.width, Height: m.height}
						},
					)
				}

			case StateInstallSuccess, StateUpdateSuccess:
				// 安装/更新成功，可以继续到主界面
				dashboard := NewDashboardWithSize(m.width, m.height)
				// 发送WindowSizeMsg确保Dashboard正确初始化
				return dashboard, tea.Batch(
					dashboard.Init(),
					func() tea.Msg {
						return tea.WindowSizeMsg{Width: m.width, Height: m.height}
					},
				)

			case StateInstallError, StateUpdateError:
				// 重新检查
				m.state = StateChecking
				m.message = "正在重新检查..."
				return m, m.checkInstallation
			}

		case "r":
			// 重新检查
			m.state = StateChecking
			m.message = "正在重新检查..."
			return m, m.checkInstallation
		}

	case installerMsg:
		m.state = msg.state
		m.message = msg.message
		m.error = msg.error
		m.status = msg.status

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View 渲染视图
func (m InstallerView) View() string {
	if m.width == 0 {
		return "正在加载..."
	}

	// 样式定义
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width - 4). // 减去 appBorderStyle 的 padding(2) + border(2)
		Align(lipgloss.Center)

	statusStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(2).
		Width(m.width - 10) // 减去 appBorderStyle(4) + statusStyle 自身的 padding(4) + border(2)

	// 整个应用的边框样式
	appBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1)

	// 帮助信息样式
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	// 标题
	title := titleStyle.Render("FRP 内网穿透管理工具 - 安装管理器")

	var statusContent strings.Builder

	// 根据状态显示不同内容
	switch m.state {
	case StateChecking:
		statusContent.WriteString(fmt.Sprintf("%s %s", m.spinner.View(), m.message))

	case StateNotInstalled:
		statusContent.WriteString("❌ FRP 未安装\n\n")
		statusContent.WriteString("FRP (Fast Reverse Proxy) 是一个高性能的反向代理应用，\n")
		statusContent.WriteString("可以帮助您将内网服务暴露到公网。\n\n")
		statusContent.WriteString("安装信息:\n")
		statusContent.WriteString(fmt.Sprintf("• 版本: %s\n", m.installer.GetVersion()))
		statusContent.WriteString(fmt.Sprintf("• 安装目录: %s\n", m.installer.GetInstallDir()))

	case StateInstalled:
		if m.status.NeedsUpdate {
			statusContent.WriteString("🔄 FRP 已安装，有新版本可用\n\n")
		} else {
			statusContent.WriteString("✅ FRP 已安装\n\n")
		}
		statusContent.WriteString(fmt.Sprintf("• 当前版本: %s\n", m.status.Version))
		if m.status.NeedsUpdate {
			statusContent.WriteString(fmt.Sprintf("• 最新版本: %s\n", m.status.LatestVersion))
		}
		statusContent.WriteString(fmt.Sprintf("• 安装目录: %s\n", m.status.InstallDir))
		statusContent.WriteString(fmt.Sprintf("• frps 路径: %s\n", m.status.FrpsPath))
		statusContent.WriteString(fmt.Sprintf("• frpc 路径: %s\n", m.status.FrpcPath))

	case StateInstalling:
		statusContent.WriteString(fmt.Sprintf("%s %s\n\n", m.spinner.View(), m.message))
		statusContent.WriteString("正在执行以下步骤:\n")
		statusContent.WriteString("1. 下载 FRP 安装包\n")
		statusContent.WriteString("2. 解压安装文件\n")
		statusContent.WriteString("3. 设置执行权限\n")
		statusContent.WriteString("4. 验证安装\n\n")
		statusContent.WriteString("请稍候，这可能需要几分钟时间...")

	case StateUpdating:
		statusContent.WriteString(fmt.Sprintf("%s %s\n\n", m.spinner.View(), m.message))
		statusContent.WriteString("正在执行以下步骤:\n")
		statusContent.WriteString("1. 备份当前版本\n")
		statusContent.WriteString("2. 下载新版本\n")
		statusContent.WriteString("3. 安装新版本\n")
		statusContent.WriteString("4. 验证更新\n\n")
		statusContent.WriteString("请稍候...")

	case StateInstallSuccess:
		statusContent.WriteString("🎉 FRP 安装成功！\n\n")
		statusContent.WriteString(m.message + "\n\n")
		statusContent.WriteString("您现在可以开始使用 FRP 进行内网穿透了。")

	case StateUpdateSuccess:
		statusContent.WriteString("🎉 FRP 更新成功！\n\n")
		statusContent.WriteString(m.message + "\n\n")
		statusContent.WriteString("您现在可以使用最新版本的 FRP 了。")

	case StateInstallError, StateUpdateError:
		statusContent.WriteString("❌ 操作失败\n\n")
		statusContent.WriteString(fmt.Sprintf("错误信息: %s\n", m.message))
		if m.error != nil {
			statusContent.WriteString(fmt.Sprintf("详细错误: %v\n", m.error))
		}
		statusContent.WriteString("\n请检查网络连接或重试。")
	}

	// 状态区域
	statusBar := statusStyle.Render(statusContent.String())

	// 操作提示
	var helpText string
	switch m.state {
	case StateChecking, StateInstalling, StateUpdating:
		helpText = "请稍候..."
	case StateNotInstalled:
		helpText = "按 Enter 或空格键开始安装 | R: 重新检查 | Q: 退出"
	case StateInstalled:
		if m.status != nil && m.status.NeedsUpdate {
			helpText = "按 Enter 或空格键更新到最新版本 | R: 重新检查 | Q: 退出"
		} else {
			helpText = "按 Enter 或空格键继续到主界面 | R: 重新检查 | Q: 退出"
		}
	case StateInstallSuccess, StateUpdateSuccess:
		helpText = "按 Enter 或空格键继续到主界面 | Q: 退出"
	case StateInstallError, StateUpdateError:
		helpText = "按 Enter 或空格键重试 | R: 重新检查 | Q: 退出"
	}

	help := helpStyle.Render(helpText)

	// 组合内容
	innerContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"", // 标题和内容之间的间隔
		statusBar,
	)

	// 计算内容高度，以便将帮助信息推到底部
	contentHeight := lipgloss.Height(innerContent)
	verticalPaddingAndBorder := 4 // 近似值，根据实际效果调整
	remainingHeight := m.height - contentHeight - lipgloss.Height(help) - verticalPaddingAndBorder
	if remainingHeight < 0 {
		remainingHeight = 0
	}

	// 最终组合，添加帮助信息和底部填充
	finalViewContent := lipgloss.JoinVertical(
		lipgloss.Left,
		innerContent,
		lipgloss.PlaceVertical(remainingHeight, lipgloss.Bottom, help),
	)

	// 应用整体边框
	return appBorderStyle.Render(finalViewContent)
}

// GetInstaller 获取安装器实例
func (m InstallerView) GetInstaller() *service.Installer {
	return m.installer
}

// GetStatus 获取安装状态
func (m InstallerView) GetStatus() *service.InstallStatus {
	return m.status
}
