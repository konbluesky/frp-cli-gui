# FRP CLI UI

一个基于 Bubble Tea 的 FRP (Fast Reverse Proxy) 命令行图形界面管理工具，支持自动检查和安装 FRP。

## 功能特性

### 🎯 核心功能
- **自动安装** - 启动时自动检查 FRP 安装状态，未安装时提供一键安装
- **实时监控面板** - 显示 FRP 服务状态、连接数、流量统计
- **智能配置管理** - 可视化编辑 FRP 服务端和客户端配置，支持模板和验证
- **日志查看** - 实时日志显示，支持过滤和搜索
- **进程管理** - 启动、停止、重启 FRP 服务

### 🎨 界面特性
- **现代化 TUI** - 基于 Bubble Tea 框架的美观界面
- **响应式布局** - 自适应终端窗口大小
- **多标签页** - 仪表盘、配置管理、日志查看、设置
- **键盘导航** - 完整的键盘快捷键支持

### 🔧 技术特性
- **智能安装** - 自动检测系统架构，下载对应版本的 FRP
- **配置验证** - 智能配置校验和错误提示
- **配置模板** - 内置多种常用配置模板
- **配置比较** - 支持配置文件差异比较和智能合并
- **进程监控** - 实时监控 FRP 进程状态
- **API 集成** - 调用 FRP 服务端监控 API
- **配置备份** - 自动配置备份和恢复

## 项目结构

```
frp-cli-ui/
├── go.mod              # Go 模块依赖
├── main.go             # 程序入口
├── README.md           # 项目文档
├── CONFIG_MANAGEMENT.md # 配置管理详细文档
├── tui/                # 用户界面组件
│   ├── dashboard.go    # 主控面板 (实时监控)
│   ├── config_editor.go # 配置编辑器 (多状态界面)
│   ├── logs_view.go    # 带过滤的日志窗口
│   └── installer_view.go # FRP 安装界面
├── service/            # FRP 控制
│   ├── manager.go      # 进程启停管理
│   ├── frp_api.go      # 调用 frps 监控 API
│   └── installer.go    # FRP 安装管理器
├── config/             # 配置处理
│   ├── loader.go       # 配置加载器 (支持导入导出)
│   ├── validator.go    # 配置验证器 (详细验证)
│   └── templates.go    # 配置模板管理器
├── test/               # 测试文件
│   ├── config_management_test.go # 配置管理测试
│   ├── local_test.go   # 本地功能测试
│   ├── test_config_ui.sh # 配置UI测试脚本
│   ├── test_frp.sh     # FRP功能测试脚本
│   ├── quick_test.sh   # 快速验证脚本
│   ├── demo.sh         # 演示脚本
│   ├── test_web_server.py # 测试Web服务器
│   ├── test_client.toml # 测试客户端配置
│   └── test_server.toml # 测试服务端配置
└── examples/           # 示例配置
    ├── frps.yaml       # 服务端配置示例
    └── frpc.yaml       # 客户端配置示例
```

## 快速开始

### 安装依赖

```bash
go mod tidy
```

### 运行程序

```bash
go run main.go
```

程序启动时会自动检查 FRP 安装状态：
- 如果未安装，会显示安装界面，支持一键安装
- 如果已安装，直接进入主界面
- 如果有新版本，会提示更新

### 编译程序

```bash
# 编译当前平台
go build -o frp-cli-ui

# 交叉编译 Linux
GOOS=linux GOARCH=amd64 go build -o frp-cli-ui-linux

# 交叉编译 Windows
GOOS=windows GOARCH=amd64 go build -o frp-cli-ui.exe
```

## 使用说明

### 首次运行

1. **自动检查** - 程序启动时自动检查 FRP 安装状态
2. **一键安装** - 如果未安装，按 Enter 键开始自动下载和安装
3. **版本管理** - 自动检测最新版本，支持更新提醒

### FRP 安装

- **安装位置**: 默认安装到 `~/.frp/` 目录
- **支持平台**: Linux、macOS、Windows
- **支持架构**: amd64、arm64、386、arm
- **版本管理**: 自动下载最新稳定版本 (当前: v0.52.3)

### 快捷键

#### 全局快捷键
- **Tab** - 切换标签页或控件间跳转
- **Shift+Tab** - 反向切换标签页或控件间跳转
- **Ctrl+Tab** - 状态间切换（配置管理中）
- **Enter** - 确认选择
- **X** - 返回上级模块
- **Esc** - 取消/返回
- **Q** 或 **Ctrl+C** - 退出程序

#### 配置管理快捷键

进入配置管理界面：
1. 使用 Tab 键切换到 "配置管理" 标签
2. 按 Enter 或 c 键进入配置编辑器

配置编辑器功能：
- **多状态界面**: 基本配置、代理管理、模板、导入导出、验证
- **智能导航**: 
  - Tab/Shift+Tab: 控件间跳转
  - Ctrl+Tab/Ctrl+Shift+Tab: 状态间切换
  - 方向键 ↑↓←→: 精确控件跳转
- **实时预览**: 配置更改实时显示YAML预览
- **配置验证**: 实时验证配置正确性，详细错误提示
- **模板系统**: 内置7种常用配置模板，一键应用
- **导入导出**: 支持配置文件的导入导出和智能合并
- **历史记录**: 支持撤销操作，最多保存50条历史

配置编辑器快捷键：
- **Ctrl+S**: 保存配置
- **Ctrl+L**: 加载配置
- **Ctrl+E**: 导出配置
- **Ctrl+I**: 导入配置
- **Ctrl+V**: 验证配置
- **Ctrl+R**: 重置配置
- **Ctrl+Z**: 撤销更改
- **Ctrl+H**: 显示/隐藏帮助
- **X**: 返回主界面

代理管理快捷键（客户端模式）：
- **N**: 新建代理
- **Enter**: 编辑选中代理
- **D**: 删除选中代理

详细使用说明请查看 [CONFIG_MANAGEMENT.md](CONFIG_MANAGEMENT.md)

#### 日志查看快捷键

- **F** 或 **Ctrl+F**: 进入过滤模式
- **A** 或 **Ctrl+A**: 切换自动滚动
- **Ctrl+L**: 清空日志
- **↑↓**: 滚动日志
- **X** 或 **Esc**: 返回主界面

### 仪表盘功能

- 实时显示 FRP 服务状态
- 代理列表和状态监控
- 流量统计和连接数
- 服务器和客户端状态

### 日志查看

- 实时日志流显示
- 日志级别过滤
- 关键词搜索
- 自动滚动控制

## 测试

### 运行配置管理测试

```bash
# 运行配置管理功能测试
./test/test_config_ui.sh

# 运行单元测试
cd test && go test -v config_management_test.go
```

### 运行完整功能测试

```bash
# 运行完整FRP功能测试
./test/test_frp.sh

# 运行快速验证
./test/quick_test.sh

# 运行本地功能验证
go run test/local_test.go
```

## 配置文件

### 服务端配置示例

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

### 客户端配置示例

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

### 第一阶段 ✅
- [x] 基础 TUI 框架搭建
- [x] 主控面板界面
- [x] 配置编辑器
- [x] 日志查看器
- [x] FRP 自动安装功能

### 第二阶段 ✅
- [x] 智能配置管理系统
- [x] 配置模板和验证
- [x] 配置导入导出功能
- [x] 配置比较和合并
- [x] 历史记录和撤销

### 第三阶段 🚧
- [ ] 进程管理功能完善
- [ ] FRP API 集成
- [ ] 错误处理优化
- [ ] 性能优化

### 第四阶段 📋
- [ ] 多语言支持
- [ ] 主题系统
- [ ] 插件架构
- [ ] 高级配置功能

### 第五阶段 🔮
- [ ] Web 界面
- [ ] 远程管理
- [ ] 集群支持
- [ ] 监控告警

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

## 致谢

- [FRP](https://github.com/fatedier/frp) - 优秀的内网穿透工具
- [Charm](https://charm.sh/) - 提供优秀的 TUI 开发工具
- 所有贡献者和用户的支持

## 联系方式

如有问题或建议，请通过以下方式联系：

- 提交 Issue
- 发起 Discussion
- 邮件联系

---

**享受使用 FRP CLI UI！** 🚀 