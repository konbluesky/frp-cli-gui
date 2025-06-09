# FRP CLI UI

一个基于 Bubble Tea 的 FRP (Fast Reverse Proxy) 命令行图形界面管理工具，支持自动检查和安装 FRP。

![FRP CLI UI](./cli_snapshot.gif)

## 功能特性

### 🎯 核心功能
- **自动安装** - 启动时自动检查 FRP 安装状态，未安装时提供一键安装
- **实时监控面板** - 显示 FRP 服务状态、连接数、流量统计
- **智能配置管理** - 可视化编辑 FRP 服务端和客户端配置，支持模板和验证
- **进程管理** - 启动、停止、重启 FRP 服务
- **文件管理** - 内置文件选择器，支持配置文件的选择和管理

### 🎨 界面特性
- **现代化 TUI** - 基于 Bubble Tea 框架的美观界面
- **响应式布局** - 自适应终端窗口大小，智能处理Emoji字符宽度
- **多标签页架构** - 可插拔的标签页系统，支持动态扩展
- **键盘导航** - 完整的键盘快捷键支持
- **双栏布局** - 左侧菜单，右侧详细信息的直观布局

### 🔧 技术特性
- **智能安装** - 自动检测系统架构，下载对应版本的 FRP (v0.52.3)
- **配置验证** - 实时配置校验和错误提示
- **配置模板** - 内置多种常用配置模板，一键应用
- **进程监控** - 实时监控 FRP 进程状态和日志
- **API 集成** - 调用 FRP 服务端监控 API
- **Unicode处理** - 正确处理Emoji字符显示宽度，避免界面错位

## 项目结构

```
frp-cli-ui/
├── cmd/
│   ├── frp-cli-ui/         # 主程序入口
│   │   └── main.go
│   └── tabs_example/       # 标签页示例
├── pkg/
│   ├── ui/                 # 用户界面组件
│   │   ├── main_dashboard.go    # 主控面板
│   │   ├── dashboard_tab.go     # 仪表板标签页
│   │   ├── config_tab.go        # 配置管理标签页
│   │   ├── settings_tab.go      # 设置标签页
│   │   ├── config_form.go       # 配置表单组件
│   │   ├── file_picker.go       # 文件选择器
│   │   ├── app_layout.go        # 应用布局管理器
│   │   └── tab.go              # 标签页基础接口
│   └── config/             # 配置处理
│       ├── loader.go       # 配置加载器
│       ├── validator.go    # 配置验证器
│       ├── templates.go    # 配置模板管理器
│       ├── constants.go    # 常量定义
│       └── init.go         # 工作空间初始化
├── internal/
│   ├── installer/          # FRP 安装管理
│   │   └── installer.go
│   └── service/            # FRP 服务管理
│       ├── manager.go      # 进程管理
│       ├── api_client.go   # API 客户端
│       └── test_runner.go  # 测试运行器
├── examples/               # 示例程序
│   ├── config-form/        # 配置表单示例
│   ├── config-test/        # 配置测试
│   ├── keyboard-demo/      # 键盘交互演示
│   └── file-picker/        # 文件选择器演示
├── configs/                # 默认配置文件
├── build/                  # 构建输出目录
├── docs/                   # 文档
├── Makefile               # 构建脚本
├── go.mod                 # Go 模块依赖
└── README.md              # 项目文档
```

## 快速开始

### 环境要求
- Go 1.23.0 或更高版本
- 支持的操作系统：Linux、macOS、Windows
- 终端支持：推荐使用现代终端（iTerm2、Windows Terminal等）

### 安装依赖

```bash
# 安装项目依赖
make deps

# 或手动安装
go mod tidy
```

### 运行程序

```bash
# 开发模式运行
make run

# 或直接运行
go run ./cmd/frp-cli-ui
```

程序启动时会自动：
- 检查 FRP 安装状态
- 初始化工作空间 (~/.frp-manager/)
- 创建默认配置文件

### 构建程序

```bash
git clone https://github.com/konbluesky/frp-cli-ui.git
cd frp-cli-ui

# 构建当前平台
make build

# 构建所有平台
make build-all

# 安装到 GOPATH/bin
make install
```

构建后的二进制文件位于 `build/` 目录下。

## 使用说明

### 主界面功能

#### 📊 仪表板
- 实时显示 FRP 服务状态
- 代理列表和连接信息
- 流量统计和性能监控
- 服务器健康状态检查

#### 📝 配置管理
**左右分栏设计**：
- **左侧菜单**：配置类型选择、文件路径显示、操作提示
- **右侧内容**：表单编辑区域、配置预览

**配置功能**：
- 🎯 服务端配置：端口、认证、日志等设置
- 💻 客户端配置：服务器连接、代理列表管理
- 🔗 添加代理：TCP、HTTP、HTTPS、UDP代理配置
- 👥 添加访问者：P2P连接配置
- 📁 选择配置文件：通过文件选择器更换配置文件
- 👀 预览配置：实时查看YAML格式配置内容
- 💾 保存配置：一键保存到指定路径

#### ⚙️ 设置
- **FRP 安装管理**：检查、安装、更新、卸载
- **服务控制**：启动/停止服务端和客户端
- **实时日志**：查看服务运行日志
- **系统状态**：显示进程信息和资源使用

### 快捷键说明

#### 全局快捷键
- **Tab** - 切换标签页
- **Shift+Tab** - 反向切换标签页
- **Q** 或 **Ctrl+C** - 退出程序
- **Ctrl+Z** - 挂起程序

#### 配置管理快捷键
- **Tab/Shift+Tab** - 在菜单和表单间切换焦点
- **↑/↓** - 菜单导航
- **Enter** - 确认选择/进入编辑
- **ESC** - 退出表单编辑

#### 文件选择器快捷键
- **↑/↓** - 文件导航
- **Enter** - 选择文件/进入目录
- **Ctrl+D** - 选择当前目录
- **Ctrl+H** - 显示/隐藏隐藏文件
- **Home** - 回到主目录
- **ESC** - 取消选择

#### 设置页面快捷键
- **I** - 安装 FRP
- **U** - 更新 FRP  
- **Ctrl+U** - 卸载 FRP
- **S** - 启动服务端
- **Ctrl+S** - 停止服务端
- **C** - 启动客户端
- **Ctrl+X** - 停止客户端
- **R** - 刷新状态

### FRP 安装

- **安装位置**: 默认安装到 `~/.frp-manager/` 目录
- **支持平台**: Linux、macOS、Windows
- **支持架构**: amd64、arm64、386、arm
- **版本管理**: 自动下载最新稳定版本 (当前: v0.52.3)

## 开发指南

### 运行示例

```bash
# 查看可用示例
make run-example

# 运行配置表单示例
go run ./examples/config-form/main.go server

# 运行文件选择器示例  
go run ./examples/file-picker/main.go

# 运行键盘交互演示
go run ./examples/keyboard-demo/main.go

# 运行标签页示例
go run ./cmd/tabs_example/main.go
```

### 测试

```bash
# 运行所有测试
make test

# 测试覆盖率
make test-coverage

# 代码格式化
make fmt

# 代码检查
make lint
```

### 架构特点

#### 可插拔标签页系统
- 基于 `Tab` 接口的标签页架构
- 支持动态注册和管理标签页
- 统一的焦点管理和事件处理

#### Unicode字符处理
- 使用 `go-runewidth` 库正确计算字符显示宽度
- 解决Emoji字符导致的界面错位问题
- 支持终端兼容性检测和降级处理

#### 智能布局管理
- 响应式设计，适配不同终端尺寸
- 左右分栏布局，信息展示更直观
- 统一的样式管理和主题支持

## 配置文件

### 服务端配置示例 (frps.yaml)

```yaml
bindPort: 7000
token: "your-secret-token"
webServer:
  port: 7500
  user: "admin"
  password: "admin"
log:
  to: "console"
  level: "info"
```

### 客户端配置示例 (frpc.yaml)

```yaml
serverAddr: "your-server.com"
serverPort: 7000
token: "your-secret-token"
log:
  to: "console"
  level: "info"

proxies:
  - name: "web"
    type: "http"
    localIP: "127.0.0.1"
    localPort: 8080
    customDomains: ["www.example.com"]
    
  - name: "ssh"
    type: "tcp"
    localIP: "127.0.0.1"
    localPort: 22
    remotePort: 2222
```

## 开发计划

### 已完成 ✅
- [x] 基础 TUI 框架搭建 (Bubble Tea)
- [x] 可插拔标签页系统
- [x] 主控面板和仪表板
- [x] 智能配置管理系统
- [x] 文件选择器组件
- [x] FRP 自动安装功能
- [x] 进程管理和监控
- [x] 响应式布局设计

### 进行中 🚧
- [ ] 日志查看功能完善
- [ ] API 集成优化
- [ ] 错误处理改进
- [ ] 性能优化

### 计划中 📋
- [ ] 配置导入导出功能
- [ ] 配置模板扩展
- [ ] 多语言支持
- [ ] 主题系统
- [ ] 插件架构

## 贡献

欢迎提交 Issue 和 Pull Request！

### 开发环境设置
1. Fork 本项目
2. 创建功能分支
3. 提交更改
4. 创建 Pull Request

### 代码风格
- 使用 `gofmt` 格式化代码
- 遵循 Go 编程规范
- 添加必要的注释和文档

## 许可证

MIT License

## 致谢

- [FRP](https://github.com/fatedier/frp) - 优秀的内网穿透工具
- [Charm](https://charm.sh/) - 提供优秀的 TUI 开发工具库
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - 强大的 TUI 框架
- 所有贡献者和用户的支持

## 联系方式

如有问题或建议，请通过以下方式联系：

- 提交 Issue
- 发起 Discussion
- 邮件联系

---

**享受使用 FRP CLI UI！** 🚀 