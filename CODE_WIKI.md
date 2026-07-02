# 启SSH (0qssh) - Code Wiki

> 版本: 0.3.1 | 许可证: CC BY-NC-SA 4.0 | 官网: ssh.0qpz.com

---

## 目录

1. [项目概述](#1-项目概述)
2. [技术栈与依赖](#2-技术栈与依赖)
3. [项目架构](#3-项目架构)
4. [后端模块详解](#4-后端模块详解)
   - 4.1 [入口与初始化 (main.go)](#41-入口与初始化)
   - 4.2 [SSH 模块 (ssh/)](#42-ssh-模块)
   - 4.3 [AI 模块 (ai/)](#43-ai-模块)
   - 4.4 [Cloud 模块 (cloud/)](#44-cloud-模块)
5. [前端模块详解](#5-前端模块详解)
   - 5.1 [入口与路由](#51-入口与路由)
   - 5.2 [状态管理 (Stores)](#52-状态管理-stores)
   - 5.3 [视图层 (Views)](#53-视图层-views)
   - 5.4 [工具函数 (Utils)](#54-工具函数-utils)
6. [前后端通信机制](#6-前后端通信机制)
7. [数据持久化](#7-数据持久化)
8. [构建与运行](#8-构建与运行)

---

## 1. 项目概述

**启SSH** 是一个中文 SSH 远程连接工具，定位为便携、好看、开源的桌面应用。核心设计理念是"好看第一，好用第二"，面向中文用户群体。

### 核心功能

| 功能 | 说明 |
|------|------|
| SSH 连接管理 | 保存/编辑/快速连接，支持密码和密钥认证 |
| 多标签终端 | 基于 xterm.js 的终端，支持结构化/经典两种视图模式 |
| SFTP 文件管理 | 远程文件浏览、上传、下载、编辑、权限管理 |
| AI 助手 | 集成 OpenAI 兼容 API，支持工具调用（执行命令、获取系统信息等） |
| 端口转发 | 本地/远程端口转发，带流量统计 |
| 防火墙管理 | 支持 iptables / firewalld / ufw |
| 进程守护 | 基于 systemd 的服务管理 |
| 批量命令 | 多终端并行执行命令 |
| 系统监控 | CPU、内存、磁盘、网络实时监控 |
| 云端同步 | 私有云同步连接配置（WebSocket + TLS） |
| 系统托盘 | 最小化到托盘，关闭窗口不退出 |

---

## 2. 技术栈与依赖

### 后端 (Go)

| 依赖 | 版本 | 用途 |
|------|------|------|
| `wailsapp/wails/v3` | v3.0.0-alpha.74 | 桌面应用框架，Go + WebView |
| `golang.org/x/crypto` | v0.47.0 | SSH 协议实现 |
| `pkg/sftp` | v1.13.6 | SFTP 文件传输协议 |
| `gorilla/websocket` | v1.5.3 | WebSocket 通信（云端同步） |
| `go-git/go-git/v5` | v5.16.4 | Git 操作（间接依赖） |

### 前端 (Vue 3)

| 依赖 | 用途 |
|------|------|
| `vue` ^3.2 | 响应式 UI 框架 |
| `vue-router` ^4.6 | 前端路由 |
| `pinia` ^3.0 | 状态管理 |
| `xterm` ^5.3 | 终端模拟器 |
| `xterm-addon-fit/search/serialize/web-links` | xterm 插件 |
| `dockview-vue` ^6.0 | 可拖拽面板布局 |
| `codemirror` ^6.0 | 代码编辑器 |
| `monaco-editor` ^0.55 | 高级代码编辑器 |
| `@wailsio/runtime` | Wails 前端运行时（事件、绑定调用） |
| `marked` ^18.0 | Markdown 渲染 |
| `highlight.js` ^11.11 | 代码高亮 |
| `jszip` ^3.10 | ZIP 压缩/解压 |
| `splitpanes` ^4.0 | 可调整大小的分割面板 |
| `@vueuse/core` ^14.3 | Vue 组合式工具集 |

### 构建工具

| 工具 | 用途 |
|------|------|
| `vite` ^5.0 | 前端构建 |
| `Taskfile` | 任务运行器（跨平台构建） |
| `wails3` CLI | Wails 开发/构建命令 |

---

## 3. 项目架构

### 目录结构

```
/workspace
├── main.go                  # 应用入口，初始化所有服务
├── go.mod / go.sum          # Go 模块定义
├── Taskfile.yml             # 构建任务定义
├── ssh/                     # SSH 核心模块（后端）
│   ├── service.go           # SSH 主服务（连接管理、命令执行、目录上传）
│   ├── client.go            # SSH 客户端（底层连接、Shell 会话）
│   ├── config_service.go    # 应用配置服务
│   ├── storage.go           # 连接信息持久化（AES-256-GCM 加密）
│   ├── group_manager.go     # 连接分组管理
│   ├── window_manager.go    # SSH 窗口生命周期管理
│   ├── sftp_service.go      # SFTP 文件操作
│   ├── port_forward_service.go  # 端口转发
│   ├── firewall_service.go  # 防火墙管理
│   ├── monitor_service.go   # 系统监控
│   ├── process_guardian_service.go  # 进程守护
│   ├── cloud_service.go     # 云端同步服务（前端调用入口）
│   └── greetservice.go      # 问候服务（测试用）
├── ai/                      # AI 助手模块（后端）
│   ├── service.go           # AI 服务核心（消息处理、工具调用、API 通信）
│   ├── config.go            # AI 配置管理（OpenAI 兼容格式）
│   ├── storage.go           # 聊天历史持久化
│   ├── tools.go             # 工具定义（execute_command, get_server_info 等）
│   └── logger.go            # AI 操作日志
├── cloud/                   # 云端同步模块（后端）
│   ├── main.go              # 云端服务器入口
│   ├── client/client.go     # 云端客户端（WebSocket + TLS）
│   └── server/              # 云端服务器
│       ├── server.go        # 设备注册、心跳、同步
│       ├── websocket.go     # WebSocket TLS 服务器
│       ├── config.go        # 服务器配置
│       └── crypto/crypto.go # 加密、令牌验证、请求签名
├── frontend/                # 前端 (Vue 3)
│   ├── src/
│   │   ├── main.js          # Vue 应用入口
│   │   ├── App.vue          # 根组件
│   │   ├── router/          # 路由配置
│   │   ├── stores/          # Pinia 状态管理
│   │   ├── views/           # 页面视图
│   │   │   ├── Home/        # 首页（连接管理、设置、云端）
│   │   │   └── SSH/         # SSH 终端页面
│   │   │       ├── layout/  # SSH 布局（标签栏、工具栏）
│   │   │       └── panels/  # 功能面板
│   │   │           ├── Terminal/     # 终端面板
│   │   │           ├── FileManager/  # 文件管理面板
│   │   │           ├── AIChatPanel.vue    # AI 聊天面板
│   │   │           ├── MonitorPanel.vue   # 系统监控面板
│   │   │           ├── PortForwardPanel.vue  # 端口转发面板
│   │   │           ├── FirewallPanel.vue     # 防火墙面板
│   │   │           ├── ProcessGuardPanel.vue # 进程守护面板
│   │   │           ├── LogsPanel.vue         # 日志面板
│   │   │           └── BatchCommandPanel.vue # 批量命令面板
│   │   ├── utils/           # 工具函数
│   │   ├── components/      # 通用组件
│   │   └── styles/          # 全局样式
│   ├── bindings/            # Wails 自动生成的 JS 绑定
│   ├── package.json
│   └── vite.config.js
├── build/                   # 构建配置
│   ├── windows/             # Windows 构建（NSIS 安装器、MSIX）
│   ├── darwin/              # macOS 构建
│   ├── linux/               # Linux 构建（AppImage、nfpm）
│   ├── docker/              # Docker 构建（交叉编译、服务器模式）
│   ├── android/             # Android 构建
│   └── ios/                 # iOS 构建
└── web/                     # Web 模式入口
```

### 整体架构图

```
┌──────────────────────────────────────────────────────────┐
│                     Wails v3 桌面应用                      │
│                                                          │
│  ┌──────────────────────┐    ┌─────────────────────────┐ │
│  │    前端 (Vue 3)       │    │    后端 (Go)             │ │
│  │                      │    │                         │ │
│  │  ┌────────────────┐  │    │  ┌───────────────────┐  │ │
│  │  │   Home 视图     │  │    │  │   SSHService      │  │ │
│  │  │  - 连接管理     │◄─┼────┼─►│  - 连接管理        │  │ │
│  │  │  - 设置        │  │ W  │  │  - 命令执行        │  │ │
│  │  │  - 云端面板     │  │ a  │  │  - Shell 会话      │  │ │
│  │  └────────────────┘  │ i  │  └───────┬───────────┘  │ │
│  │                      │ l  │          │              │ │
│  │  ┌────────────────┐  │ s  │  ┌───────▼───────────┐  │ │
│  │  │   SSH 视图      │  │    │  │   SSHClient       │  │ │
│  │  │  - Dockview布局 │◄─┼────┼─►│  - SSH连接        │  │ │
│  │  │  - 终端面板     │  │ B  │  │  - Shell会话      │  │ │
│  │  │  - 文件管理     │  │ i  │  │  - SFTP           │  │ │
│  │  │  - AI聊天       │  │ n  │  └───────────────────┘  │ │
│  │  │  - 端口转发     │  │ d  │                         │ │
│  │  │  - 防火墙       │  │ i  │  ┌───────────────────┐  │ │
│  │  │  - 监控        │  │ n  │  │   AIService        │  │ │
│  │  │  - 进程守护     │  │ g  │  │  - OpenAI兼容API  │  │ │
│  │  └────────────────┘  │ s  │  │  - 工具调用        │  │ │
│  │                      │    │  │  - 审批流程        │  │ │
│  └──────────────────────┘    │  └───────────────────┘  │ │
│                              │                         │ │
│                              │  ┌───────────────────┐  │ │
│                              │  │   CloudService    │  │ │
│                              │  │  - WebSocket+TLS  │  │ │
│                              │  │  - 设备注册/心跳   │  │ │
│                              │  │  - 数据同步        │  │ │
│                              │  └───────────────────┘  │ │
│                              └─────────────────────────┘ │
└──────────────────────────────────────────────────────────┘
```

---

## 4. 后端模块详解

### 4.1 入口与初始化

**文件**: [main.go](file:///workspace/main.go)

应用入口负责创建和组装所有服务，注册到 Wails 框架。

#### 初始化流程

```
1. 创建 SSHService → 设置 WindowManager → 设置 App
2. 创建 AIService → 设置 App → 初始化依赖 (SSHService, SSHCommandRunner, ServerInfoProvider)
3. 创建 ConfigService → 设置 App
4. 创建 PortForwardService(依赖 SSHService)
5. 创建 FirewallService(依赖 SSHService)
6. 创建 ProcessGuardianService(依赖 SSHService)
7. 创建 CloudService(依赖 ConfigService)
8. 注册所有服务到 Wails
9. 创建主窗口 → 恢复窗口位置
10. 初始化系统托盘
11. 启动后台 goroutine:
    - 每秒发射 "time" 事件
    - 每 500ms 广播 SSH 分组状态
    - 每 20ms 轮询 Shell 输出并发射 "ssh:terminal-output" 事件
```

#### 关键常量

```go
const (
    AppVersion = "0.3.1"
    AppName    = "启SSH"
)
```

#### 关键函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `calculateWindowSize` | `(app *application.App) (int, int)` | 根据屏幕尺寸计算窗口大小（85% 屏幕占比，4 档位选择） |

---

### 4.2 SSH 模块

#### 4.2.1 SSHService — SSH 主服务

**文件**: [ssh/service.go](file:///workspace/ssh/service.go)

SSH 模块的核心服务，作为 Wails Service 暴露给前端，管理所有 SSH 连接的生命周期。

```go
type SSHService struct {
    mu            sync.RWMutex
    clients       map[string]*SSHClient   // connID → SSH客户端
    storage       *StorageManager         // 连接存储
    windowManager *WindowManager          // 窗口管理
    groupManager  *GroupManager           // 分组管理
    app           *application.App        // Wails 应用实例
    shellSessions map[string]bool         // Shell 会话追踪
    directoryUploadCancelMu sync.Mutex
    directoryUploadCancelled bool          // 目录上传取消标志
}
```

**关键方法（前端可调用）**:

| 方法 | 签名 | 说明 |
|------|------|------|
| `TestConnection` | `(config *SSHConfig) error` | 测试 SSH 连接（连接后立即断开） |
| `Connect` | `(connID string, config *SSHConfig) error` | 使用指定 ID 连接 |
| `CreateAndConnect` | `(config *SSHConfig) (string, error)` | 创建新连接并连接（返回 connID） |
| `CreateAndConnectWithGroup` | `(config *SSHConfig, groupID string) (map[string]string, error)` | 指定分组创建连接 |
| `Disconnect` | `(connID string) error` | 断开连接 |
| `CloseGroup` | `(groupID string) error` | 关闭分组内所有连接 |
| `RunCommand` | `(connID, command string) (string, error)` | 执行远程命令 |
| `StartShellSession` | `(connID string) (string, error)` | 启动交互式 Shell |
| `StartShellSessionByID` | `(connID string, sessionID string) error` | 指定 ID 启动 Shell |
| `WriteToShell` | `(connID, data string) error` | 向 Shell 写入数据 |
| `WriteToTerminalByID` | `(connID, sessionID, data string) error` | 向指定终端写入 |
| `ReadFromShellSession` | `(connID, sessionID string, buf []byte) (int, error)` | 从 Shell 会话读取 |
| `GetAllConnections` | `() []*ConnectionInfo` | 获取所有连接信息 |
| `GetConnectionList` | `() []string` | 获取所有连接 ID |
| `GetSessionIDs` | `(connID string) []string` | 获取连接的所有会话 ID |
| `IsShellSessionActive` | `(connID, sessionID string) bool` | 检查会话是否活跃 |
| `StartHealthCheck` | `()` | 启动健康检查（5 秒间隔 ping） |
| `GetServerKey` | `(connID string) string` | 获取服务器标识（host:port） |
| `CancelDirectoryUpload` | `()` | 取消目录上传 |
| `ClearWindowPositions` | `()` | 清除窗口位置记忆 |

**辅助类型**:

```go
type FriendlyError struct {
    Message string
}
```

`formatSSHError()` 将底层 SSH 错误转换为中文友好提示，覆盖认证失败、连接超时、连接拒绝、主机不可达、密钥错误等场景。

---

#### 4.2.2 SSHClient — SSH 客户端

**文件**: [ssh/client.go](file:///workspace/ssh/client.go)

封装底层 SSH 连接，管理 Shell 会话和 SFTP。

```go
type SSHConfig struct {
    Name       string `json:"name"`
    Host       string `json:"host"`
    Port       int    `json:"port"`
    Username   string `json:"username"`
    Password   string `json:"password,omitempty"`
    KeyPath    string `json:"keyPath,omitempty"`
    PrivateKey string `json:"privateKey,omitempty"`
    Timeout    int    `json:"timeout,omitempty"`
}

type SSHClient struct {
    config      *SSHConfig
    client      *ssh.Client
    sftpClient  *sftp.Client
    isConnected bool
    closing     bool
    sessions    map[string]*shellSession  // sessionID → session
    sessMu      sync.RWMutex
    lastPingTime time.Time
    latency      int64
    searchCancelMap map[string]bool
    searchMutex     sync.RWMutex
    onDisconnect func(connID string)
    connID       string
}

type shellSession struct {
    session   *ssh.Session
    stdin     io.WriteCloser
    stdout    io.Reader
    readBuf   []byte
    bufMutex  sync.Mutex
    closing   int32  // atomic: 0=active, 1=closing
}

type CommandResult struct {
    Stdout   string `json:"stdout"`
    Stderr   string `json:"stderr"`
    ExitCode int    `json:"exitCode"`
    Success  bool   `json:"success"`
}
```

**关键方法**:

| 方法 | 说明 |
|------|------|
| `Connect()` | 建立 SSH 连接（支持密码/密钥认证，优先密钥） |
| `Close()` | 关闭连接和所有会话 |
| `ExecuteCommand(cmd string)` | 执行命令并返回结果 |
| `StartShell()` | 启动交互式 Shell（pty 模式） |
| `WriteToShell(data string)` | 向 Shell stdin 写入 |
| `ReadFromShell(buf []byte)` | 从 Shell stdout 读取 |
| `InitSFTP()` | 初始化 SFTP 客户端 |
| `Ping()` | 发送 keepalive 请求 |
| `IsConnected()` | 检查连接状态 |
| `SetDisconnectCallback()` | 设置断线回调 |

---

#### 4.2.3 ConfigService — 配置服务

**文件**: [ssh/config_service.go](file:///workspace/ssh/config_service.go)

管理应用全局配置，持久化到 JSON 文件。

```go
type AppConfig struct {
    Terminal  TerminalConfig  `json:"terminal"`
    UI        UIConfig        `json:"ui"`
    Cloud     CloudConfig     `json:"cloud"`
    Shortcuts ShortcutsConfig `json:"shortcuts"`
    Advanced  AdvancedConfig  `json:"advanced"`
}

type TerminalConfig struct {
    DefaultType       string `json:"defaultType"`       // structured | classic
    AutoSwitchClassic bool   `json:"autoSwitchClassic"`
    SwitchMode        string `json:"switchMode"`        // prompt | auto | inline
    FontSize          int    `json:"fontSize"`
    CommandSendMode   string `json:"commandSendMode"`   // enter | button
    CodeHighlight     bool   `json:"codeHighlight"`
}

type UIConfig struct {
    AutoTray         bool   `json:"autoTray"`
    RememberPosition bool   `json:"rememberPosition"`
    AutoShowHome     bool   `json:"autoShowHome"`
    Theme            string `json:"theme"`             // dark | light
}

type CloudConfig struct {
    Enabled      bool   `json:"enabled"`
    ServerURL    string `json:"serverUrl"`
    Token        string `json:"token"`
    SyncInterval int    `json:"syncInterval"`
    AutoSyncTo   bool   `json:"autoSyncTo"`
    AutoSyncFrom bool   `json:"autoSyncFrom"`
}

type ShortcutsConfig struct {
    Enabled       bool `json:"enabled"`
    SwitchTab     bool `json:"switchTab"`
    SaveGroup     bool `json:"saveGroup"`
    CloudUpload   bool `json:"cloudUpload"`
    CloudDownload bool `json:"cloudDownload"`
}

type AdvancedConfig struct {
    GroupBehavior string `json:"groupBehavior"`  // join_default | new_window | prompt
}
```

**关键方法**:

| 方法 | 说明 |
|------|------|
| `GetConfig()` | 获取当前配置 |
| `SaveConfig(config *AppConfig)` | 保存配置 |
| `UpdateTerminalConfig(tc TerminalConfig)` | 更新终端配置 |
| `UpdateUIConfig(uc UIConfig)` | 更新 UI 配置 |
| `UpdateCloudConfig(cc CloudConfig)` | 更新云端配置 |

---

#### 4.2.4 StorageManager — 连接存储

**文件**: [ssh/storage.go](file:///workspace/ssh/storage.go)

管理 SSH 连接信息的持久化，使用 AES-256-GCM 加密敏感数据。

```go
type ConnectionInfo struct {
    ID         string `json:"id"`
    Name       string `json:"name"`
    Host       string `json:"host"`
    Port       int    `json:"port"`
    Username   string `json:"username"`
    Password   string `json:"password,omitempty"`
    KeyPath    string `json:"keyPath,omitempty"`
    PrivateKey string `json:"privateKey,omitempty"`
    Status     string `json:"status"`
    Saved      bool   `json:"saved"`
    GroupID    string `json:"group_id,omitempty"`
}

type StorageManager struct {
    mu                sync.RWMutex
    connections       map[string]*ConnectionInfo
    dataFile          string           // 缓存文件路径
    permanentDataFile string           // 永久保存文件路径
    encryptKey        []byte           // AES-256 加密密钥
}
```

**存储策略**:
- 双文件机制：缓存文件（运行时）+ 永久文件（持久化）
- 密码等敏感字段使用 AES-256-GCM 加密存储
- 数据目录：可执行文件同级的 `data/` 目录

**关键方法**:

| 方法 | 说明 |
|------|------|
| `AddConnection(info *ConnectionInfo)` | 添加连接 |
| `GetConnection(id string)` | 获取连接信息 |
| `UpdateConnection(info *ConnectionInfo)` | 更新连接 |
| `DeleteConnection(id string)` | 删除连接 |
| `GetAllConnections()` | 获取所有连接 |
| `UpdateConnectionStatus(id, status string)` | 更新连接状态 |
| `DeleteFromCache(id string)` | 从缓存中删除（保留永久记录） |

---

#### 4.2.5 GroupManager — 分组管理

**文件**: [ssh/group_manager.go](file:///workspace/ssh/group_manager.go)

管理 SSH 连接的分组，每个分组对应一个独立的 SSH 窗口。

```go
type SSHGroup struct {
    ID        string   `json:"id"`
    Name      string   `json:"name"`
    ConnIDs   []string `json:"conn_ids"`
    WindowID  string   `json:"window_id"`
    IsDefault bool     `json:"is_default"`
}

type GroupManager struct {
    mu     sync.RWMutex
    groups map[string]*SSHGroup
    nextID int
}
```

**关键方法**:

| 方法 | 说明 |
|------|------|
| `CreateGroup(name string)` | 创建新分组 |
| `CreateGroupWithName(groupID, name string)` | 使用指定 ID 创建分组 |
| `GetGroup(groupID string)` | 获取分组 |
| `AddConnectionToGroup(groupID, connID string)` | 添加连接到分组 |
| `RemoveConnectionFromGroup(groupID, connID string)` | 从分组移除连接 |
| `GetAllGroups()` | 获取所有分组 |
| `DeleteGroup(groupID string)` | 删除分组 |

---

#### 4.2.6 WindowManager — 窗口管理

**文件**: [ssh/window_manager.go](file:///workspace/ssh/window_manager.go)

管理 SSH 分组窗口的创建、位置记忆和生命周期。

```go
type WindowPosition struct {
    X, Y, Width, Height int
}

type WindowManager struct {
    app            *application.App
    configService  *ConfigService
    windows        map[string]*application.WebviewWindow  // groupID → Window
    onGroupClose   func(groupID string)
    positions      map[string]*WindowPosition
    positionsFile  string
}
```

**关键方法**:

| 方法 | 说明 |
|------|------|
| `CreateSSHWindow(groupID, groupName, activeConnID string)` | 创建或聚焦 SSH 窗口 |
| `CloseWindow(groupID string)` | 关闭窗口并清理 |
| `SaveMainWindowPosition(window)` | 保存主窗口位置 |
| `RestoreMainWindowPosition(window)` | 恢复主窗口位置 |
| `HideAllWindows()` / `ShowAllWindows()` | 隐藏/显示所有窗口 |
| `GetWindowCount()` | 获取打开的窗口数量 |
| `ClearPositions()` | 清除所有窗口位置记忆 |

**窗口位置策略**:
- 每 3 秒检查一次位置变化，有变化才写盘
- 窗口关闭时最终保存一次
- 位置超出屏幕范围时不恢复
- 屏幕尺寸档位：1920x1080 / 1600x1000 / 1400x900 / 1200x800（取不超过 85% 屏幕的最大档位）

---

#### 4.2.7 SFTPService — SFTP 文件操作

**文件**: [ssh/sftp_service.go](file:///workspace/ssh/sftp_service.go)

基于 `pkg/sftp` 的文件操作服务，挂载在 `SSHClient` 上。

```go
type FileInfo struct {
    Name    string    `json:"name"`
    Path    string    `json:"path"`
    Size    int64     `json:"size"`
    Mode    string    `json:"mode"`
    ModTime time.Time `json:"modTime"`
    IsDir   bool      `json:"isDir"`
    Owner   string    `json:"owner"`
    Group   string    `json:"group"`
}
```

**关键方法**:

| 方法 | 说明 |
|------|------|
| `InitSFTP()` | 初始化 SFTP 客户端 |
| `UploadFile(localPath, remotePath)` | 上传文件 |
| `DownloadFile(remotePath, localPath)` | 下载文件 |
| `ListDirectory(remotePath)` | 列出目录内容 |
| `CreateDirectory(remotePath)` | 创建目录 |
| `DeleteFile(remotePath)` | 删除文件 |
| `RenameFile(oldPath, newPath)` | 重命名文件 |
| `GetFileInfo(remotePath)` | 获取文件信息 |
| `ChangePermissions(remotePath, mode)` | 修改文件权限 |

---

#### 4.2.8 PortForwardService — 端口转发

**文件**: [ssh/port_forward_service.go](file:///workspace/ssh/port_forward_service.go)

管理本地/远程端口转发，带流量统计。

```go
type PortForward struct {
    ID         string `json:"id"`
    ConnID     string `json:"connId"`
    Type       string `json:"type"`       // "local" / "remote"
    BindAddr   string `json:"bindAddr"`
    BindPort   int    `json:"bindPort"`
    RemoteHost string `json:"remoteHost"`
    RemotePort int    `json:"remotePort"`
    Status     string `json:"status"`     // "running" / "stopped" / "error"
    ActiveConns int64 `json:"activeConns"`
    TotalConns  int64 `json:"totalConns"`
    BytesSent   int64 `json:"bytesSent"`
    BytesRecv   int64 `json:"bytesRecv"`
}
```

**关键方法**:

| 方法 | 说明 |
|------|------|
| `CreateForward(connID, fwdType, bindAddr string, bindPort int, remoteHost string, remotePort int)` | 创建端口转发 |
| `StopForward(id string)` | 停止转发 |
| `ListForwards(connID string)` | 列出所有转发 |
| `StopAllForwards(connID string)` | 停止所有转发 |

---

#### 4.2.9 FirewallService — 防火墙管理

**文件**: [ssh/firewall_service.go](file:///workspace/ssh/firewall_service.go)

远程防火墙规则管理，自动检测防火墙类型。

```go
type FirewallRule struct {
    Index    int    `json:"index"`
    Chain    string `json:"chain"`
    Target   string `json:"target"`
    Protocol string `json:"protocol"`
    Source   string `json:"source"`
    Dest     string `json:"dest"`
    Port     string `json:"port"`
    Comment  string `json:"comment"`
    Raw      string `json:"raw"`
}

type FirewallInfo struct {
    Type      string         `json:"type"`      // iptables / firewalld / ufw / unknown
    Status    string         `json:"status"`    // active / inactive / unknown
    Rules     []FirewallRule `json:"rules"`
    RawOutput string         `json:"rawOutput"`
    Chains    []string       `json:"chains"`
}
```

**检测优先级**: ufw > firewalld > iptables

**关键方法**:

| 方法 | 说明 |
|------|------|
| `GetFirewallInfo(connID string)` | 获取防火墙信息 |
| `AddRule(connID, chain, target, protocol, source, dest, port string)` | 添加规则 |
| `DeleteRule(connID string, index int)` | 删除规则 |
| `ToggleFirewall(connID string, enable bool)` | 启用/禁用防火墙 |

---

#### 4.2.10 MonitorService — 系统监控

**文件**: [ssh/monitor_service.go](file:///workspace/ssh/monitor_service.go)

远程服务器系统资源监控。

```go
type SystemStats struct {
    Timestamp int64        `json:"timestamp"`
    Uptime    string       `json:"uptime"`
    CPU       CPUStats     `json:"cpu"`
    Memory    MemoryStats  `json:"memory"`
    Disk      DiskStats    `json:"disk"`
    Network   NetworkStats `json:"network"`
}

type CPUStats struct {
    UsagePercent float64   `json:"usagePercent"`
    Cores        int       `json:"cores"`
    PerCPUUsage  []float64 `json:"perCpuUsage"`
    LoadAvg      LoadAvg   `json:"loadAvg"`
}

type MemoryStats struct {
    Total       uint64  `json:"total"`
    Used        uint64  `json:"used"`
    Free        uint64  `json:"free"`
    UsedPercent float64 `json:"usedPercent"`
    Cached      uint64  `json:"cached"`
    SwapTotal   uint64  `json:"swapTotal"`
    SwapUsed    uint64  `json:"swapUsed"`
}
```

**关键方法**:

| 方法 | 说明 |
|------|------|
| `GetSystemStats(connID string)` | 获取系统统计 |
| `GetProcessList(connID string)` | 获取进程列表 |
| `KillProcess(connID string, pid int)` | 终止进程 |

---

#### 4.2.11 ProcessGuardianService — 进程守护

**文件**: [ssh/process_guardian_service.go](file:///workspace/ssh/process_guardian_service.go)

基于 systemd 的进程守护管理。

```go
type GuardianProcess struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Command     string `json:"command"`
    WorkDir     string `json:"workDir"`
    Status      string `json:"status"`  // running / stopped / failed / unknown
    PID         int    `json:"pid"`
    AutoRestart bool   `json:"autoRestart"`
    LogPath     string `json:"logPath"`
    Restarts    int    `json:"restarts"`
}
```

**检测优先级**: systemd > OpenRC > sysvinit

---

#### 4.2.12 CloudService — 云端同步

**文件**: [ssh/cloud_service.go](file:///workspace/ssh/cloud_service.go)

作为 Wails Service 暴露给前端，封装 `cloud/client` 包的功能。

```go
type CloudService struct {
    mu      sync.RWMutex
    app     *application.App
    client  *client.CloudClient
    config  *ConfigService
    running bool
    stopCh  chan struct{}
}
```

**关键方法**:

| 方法 | 说明 |
|------|------|
| `Connect(serverAddr, token string)` | 连接云端（TLS + 设备注册 + 心跳） |
| `Disconnect()` | 断开连接 |
| `IsConnected()` | 检查连接状态 |
| `PullSync()` | 拉取同步数据 |
| `PushSync(connections)` | 推送同步数据 |

---

### 4.3 AI 模块

#### 4.3.1 AIService — AI 服务核心

**文件**: [ai/service.go](file:///workspace/ai/service.go)

AI 助手的核心服务，支持 OpenAI 兼容 API 的对话和工具调用。

```go
type AIService struct {
    mu       sync.RWMutex
    app      *application.App
    config   *ConfigManager
    store    *ChatStore
    logger   *AILogger
    sessions map[string]*session    // serverKey → 共享会话
    conns    map[string]*connState  // connId → 独立状态
    approvals map[string]chan bool  // callID → approve/deny
    results   map[string]chan string // callID → result
    serverInfo ServerInfoProvider   // 外部依赖：从 connID 解析 serverKey
    sshRunner  SSHCommandRunner     // 外部依赖：执行 SSH 命令
}

type session struct {
    mu       sync.Mutex
    messages []map[string]interface{}
}

type connState struct {
    mu         sync.Mutex
    processing bool
    cancel     chan struct{}
    queue      []*ChatRequest
}
```

**接口依赖**:

```go
type ServerInfoProvider interface {
    GetServerKey(connID string) string
}

type SSHCommandRunner interface {
    RunCommand(connID, command string) (string, error)
}
```

这两个接口在 `main.go` 中通过 `ai.InitDeps()` 注入，均由 `SSHService` 实现。

**关键方法**:

| 方法 | 说明 |
|------|------|
| `SendMessage(req ChatRequest)` | 发送消息（入口，带队列和去重） |
| `SubmitToolResult(callID, result string)` | 提交工具调用结果 |
| `GetChatHistory(connID string)` | 获取聊天历史 |
| `ClearHistory(connID string)` | 清除历史 |
| `GetAIConfig()` / `SaveAIConfig(config)` | 获取/保存 AI 配置 |

**工具调用流程**:

```
1. 用户发送消息 → SendMessage()
2. 调用 OpenAI 兼容 API（流式响应）
3. API 返回工具调用请求 → 发送 "ai:tool-approval" 事件到前端
4. 前端弹出确认框 → 用户审批
5. 审批通过 → 发送 "ai:execute-tool" 事件到前端
6. 前端通过终端执行命令 → 收集结果
7. 结果通过 "ai:submit-tool-result" 事件回传后端
8. 后端将结果追加到上下文 → 继续调用 API
```

---

#### 4.3.2 AI 配置

**文件**: [ai/config.go](file:///workspace/ai/config.go)

```go
type AIConfig struct {
    APIEndpoint  string  `json:"api_endpoint"`    // OpenAI 兼容端点
    APIKey       string  `json:"api_key"`
    Model        string  `json:"model"`
    Timeout      int     `json:"timeout"`         // 默认 120 秒
    Temperature  float64 `json:"temperature"`     // 0.0-2.0，默认 1.0
    MaxTokens    int     `json:"max_tokens"`      // 默认 4096
    TopP         float64 `json:"top_p"`           // 默认 0.95
    SystemPrompt string  `json:"system_prompt"`   // 系统提示词
}
```

默认系统提示词定义了 AI 为"启SSH AI"角色，严格限制工具调用场景（仅用户明确要求执行命令时才调用工具）。

---

#### 4.3.3 AI 工具定义

**文件**: [ai/tools.go](file:///workspace/ai/tools.go)

定义了 OpenAI function calling 格式的工具：

| 工具名 | 参数 | 说明 |
|--------|------|------|
| `execute_command` | `command`, `targetTerminal?` | 执行单条 Shell 命令 |
| `get_server_info` | 无 | 获取服务器基本信息 |
| `get_system_status` | 无 | 获取服务器实时状态 |
| `open_terminal` | 无 | 打开新终端面板 |
| `close_terminal` | 无 | 已废弃 |
| `multi_execute` | `commands[]` | 并行执行多条命令 |

---

#### 4.3.4 聊天存储

**文件**: [ai/storage.go](file:///workspace/ai/storage.go)

```go
type Project struct {
    ID, Name, Host string
    Port           int
    Username       string
    CreatedAt, UpdatedAt time.Time
}

type Session struct {
    ID, ProjectID, Title string
    CreatedAt, UpdatedAt time.Time
}

type ChatStore struct {
    baseDir  string
    projects map[string]*Project
    sessions map[string]*ChatSession
}
```

存储结构: `data/ai/projects/` + `data/ai/sessions/` + `data/ai/history/`

---

### 4.4 Cloud 模块

#### 4.4.1 CloudClient — 云端客户端

**文件**: [cloud/client/client.go](file:///workspace/cloud/client/client.go)

```go
type CloudClient struct {
    mu         sync.RWMutex
    serverAddr string
    token      string
    deviceID   string
    conn       *websocket.Conn
    connected  bool
}

type SyncConnection struct {
    ID, Name, Host string
    Port           int
    Username       string
    Password       string
    KeyPath        string
    Source         string
    UpdatedAt      time.Time
}

type SyncData struct {
    Connections []SyncConnection
    UpdatedAt   time.Time
    DeviceID    string
}
```

**关键方法**:

| 方法 | 说明 |
|------|------|
| `Connect()` | 通过 WebSocket + TLS 连接服务器 |
| `Disconnect()` | 断开连接 |
| `Register(name, host, os, version)` | 注册设备 |
| `Heartbeat()` | 发送心跳（30 秒间隔） |
| `PullSync()` | 拉取同步数据 |
| `PushSync(data)` | 推送同步数据 |

---

#### 4.4.2 CloudServer — 云端服务器

**文件**: [cloud/server/server.go](file:///workspace/cloud/server/server.go)

独立的云端同步服务器，管理设备注册、心跳检测和数据同步。

#### 4.4.3 Crypto — 加密模块

**文件**: [cloud/server/crypto/crypto.go](file:///workspace/cloud/server/crypto/crypto.go)

提供 AES-256-GCM 加密、令牌验证（防暴力破解）、请求签名验证等安全功能。

---

## 5. 前端模块详解

### 5.1 入口与路由

**入口**: [frontend/src/main.js](file:///workspace/frontend/src/main.js)

```javascript
// 创建 Vue 应用，加载 Pinia、路由
const app = createApp(App)
app.use(pinia)
app.use(router)
app.mount('#app')
```

**路由**: [frontend/src/router/index.js](file:///workspace/frontend/src/router/index.js)

```
/               → 重定向到 /home
/home           → HomeLayout（首页布局）
  /home/new     → NewConnection（新建连接）
  /home/settings → Settings（设置）
  /home/cloud   → CloudPanel（私有云端）
/ssh            → SSHLayout（SSH 布局）
  /ssh/terminal → Terminal（终端）
```

路由使用 Hash 模式 (`createWebHashHistory`)，每个 SSH 窗口通过 URL 参数 `?group=xxx&activeConn=xxx` 传递分组信息。

---

### 5.2 状态管理 (Stores)

| Store | 文件 | 职责 |
|-------|------|------|
| `sshConnections` | [stores/sshConnections.js](file:///workspace/frontend/src/stores/sshConnections.js) | SSH 连接列表管理，监听 `ssh:connections-updated` 事件同步状态 |
| `sshTabs` | [stores/sshTabs.js](file:///workspace/frontend/src/stores/sshTabs.js) | SSH 标签页管理（添加/关闭/激活/刷新） |
| `sshLayout` | [stores/sshLayout.js](file:///workspace/frontend/src/stores/sshLayout.js) | Dockview 面板布局配置 |
| `terminalSessions` | [stores/terminalSessions.js](file:///workspace/frontend/src/stores/terminalSessions.js) | 终端会话管理 |
| `config` | [stores/config.js](file:///workspace/frontend/src/stores/config.js) | 应用配置状态 |
| `layout` | [stores/layout.js](file:///workspace/frontend/src/stores/layout.js) | 布局状态 |
| `aiToolLog` | [stores/aiToolLog.js](file:///workspace/frontend/src/stores/aiToolLog.js) | AI 工具调用日志 |
| `commandHistory` | [stores/commandHistory.js](file:///workspace/frontend/src/stores/commandHistory.js) | 命令历史记录 |

**sshConnections Store 关键逻辑**:
- `updateConnections()` 使用指纹对比（id:status:name:host:port:saved:group_id）避免无效更新
- 300ms 防抖处理，减少高频事件导致的重复渲染

**sshTabs Store 关键逻辑**:
- `refreshTabs()` 采用增量更新策略：只添加缺失标签，不删除用户手动关闭的标签
- 同名标签自动添加 ID 后 8 位作为标识

---

### 5.3 视图层 (Views)

#### Home 视图

```
HomeLayout
├── AppLayout          # 顶部导航栏
├── SidebarPanel       # 左侧连接列表（搜索、已保存/缓存分组）
└── 内容区域
    ├── NewConnection  # 新建连接表单
    ├── Settings       # 设置页面
    └── CloudPanel     # 云端同步面板
```

#### SSH 视图

```
SSHLayout
├── TopBookmarkBar     # 顶部标签栏
└── TabContent         # 标签内容区
    └── DockviewLayout # Dockview 可拖拽面板布局
        ├── StructuredTerminalPanel  # 结构化终端（默认）
        ├── FileManagerPanel        # 文件管理
        ├── AIChatPanel             # AI 聊天
        ├── MonitorPanel            # 系统监控
        ├── PortForwardPanel        # 端口转发
        ├── FirewallPanel           # 防火墙
        ├── ProcessGuardPanel       # 进程守护
        ├── LogsPanel               # 日志
        └── BatchCommandPanel       # 批量命令
```

#### StructuredTerminalPanel — 结构化终端

**文件**: [views/SSH/panels/Terminal/StructuredTerminalPanel.vue](file:///workspace/frontend/src/views/SSH/panels/Terminal/StructuredTerminalPanel.vue)

核心终端组件，集成 xterm.js，支持两种视图模式：

| 模式 | 说明 |
|------|------|
| **结构化** | 基于 Shell 提示符检测，将输出分割为命令块（Command Block），显示命令、输出、状态、耗时 |
| **经典** | 传统终端视图，直接显示原始输出 |

**组合式函数 (Composables)**:

| 函数 | 文件 | 职责 |
|------|------|------|
| `useTerminal` | composables/useTerminal.js | xterm.js 实例管理、插件加载、尺寸调整 |
| `useBlockManager` | composables/useBlockManager.js | 命令块分割（基于提示符检测）、状态追踪 |
| `useShellSession` | composables/useShellSession.js | Shell 会话管理（启动/写入/读取/关闭） |
| `useCommandCompletion` | composables/useCommandCompletion.js | 命令自动补全 |
| `useCommandHistory` | composables/useCommandHistory.js | 命令历史记录 |
| `useInteractiveMode` | composables/useInteractiveMode.js | 交互式模式检测与切换 |
| `useRecording` | composables/useRecording.js | 终端录制 |
| `useBlockManager` | composables/useBlockManager.js | 块管理（命令边界检测） |

**BlockManager 核心逻辑**:

```
终端输出流 → 提示符检测（PROMPT_RE: /[\$#]\s*$/）→ 命令边界分割 → Command Block
  ├── type: COMMAND | SYSTEM
  ├── status: RUNNING → SUCCESS | FAILED | CANCELLED
  ├── command: 执行的命令
  ├── content: 命令输出
  └── exitCode / duration
```

#### FileManagerPanel — 文件管理

**文件**: [views/SSH/panels/FileManager/FileManagerPanel.vue](file:///workspace/frontend/src/views/SSH/panels/FileManager/FileManagerPanel.vue)

**子组件**:

| 组件 | 职责 |
|------|------|
| `FileToolbar` | 工具栏（导航、上传、搜索、批量操作） |
| `FileTable` | 文件列表表格 |
| `ActionMenu` | 右键/三点操作菜单 |
| `FilePreview` | 文件预览 |
| `CodeEditor` | 代码编辑器（CodeMirror / Monaco） |
| `PermissionEditor` | 权限编辑器 |

**组合式函数**:

| 函数 | 职责 |
|------|------|
| `useFileOperations` | 文件操作（复制、粘贴、删除、重命名、压缩） |
| `useNavigation` | 目录导航 |
| `usePermissions` | 权限管理 |
| `useUploadTasks` | 上传任务管理 |

#### AIChatPanel — AI 聊天

**文件**: [views/SSH/panels/AIChatPanel.vue](file:///workspace/frontend/src/views/SSH/panels/AIChatPanel.vue)

与后端 AIService 通信，处理消息发送、工具调用审批和结果返回。

---

### 5.4 工具函数 (Utils)

| 文件 | 职责 |
|------|------|
| [aiToolExecutor.js](file:///workspace/frontend/src/utils/aiToolExecutor.js) | AI 工具执行器：管理 AI 终端、监听 `ai:execute-tool` 事件、通过终端执行命令 |
| [commandRunner.js](file:///workspace/frontend/src/utils/commandRunner.js) | AI 命令执行器：通过 AI 终端 session 发送命令并收集输出 |
| [sshEvents.js](file:///workspace/frontend/src/utils/sshEvents.js) | SSH 事件管理：监听分组更新、终端输出等事件 |
| [commandAnalyzer.js](file:///workspace/frontend/src/utils/commandAnalyzer.js) | 命令分析 |
| [commandSecurityAnalyzer.js](file:///workspace/frontend/src/utils/commandSecurityAnalyzer.js) | 命令安全分析 |
| [commandDescriptions.js](file:///workspace/frontend/src/utils/commandDescriptions.js) | 命令描述映射 |
| [cloudClient.js](file:///workspace/frontend/src/utils/cloudClient.js) | 云端客户端工具 |
| [confirm.js](file:///workspace/frontend/src/utils/confirm.js) | 确认对话框工具 |
| [logger.js](file:///workspace/frontend/src/utils/logger.js) | 日志工具 |
| [message.js](file:///workspace/frontend/src/utils/message.js) | 消息提示工具 |

**aiToolExecutor.js 核心逻辑**:

```
1. 监听 "ai:execute-tool" 事件
2. 解析工具调用（execute_command / multi_execute / get_server_info / get_system_status / open_terminal）
3. 确保有可用的 AI 终端（自动创建或复用）
4. 通过终端写入命令
5. 监听终端输出，检测提示符判断命令完成
6. 收集输出结果，通过 "ai:submit-tool-result" 事件回传
```

---

## 6. 前后端通信机制

### Wails 绑定调用

前端通过 Wails 自动生成的 JS 绑定调用后端服务方法：

```javascript
// 示例：调用 SSHService 的方法
import * as SSHService from '../../bindings/changeme/ssh/sshservice.js'
await SSHService.CreateAndConnect(config)
await SSHService.RunCommand(connID, command)
```

绑定文件位于 `frontend/bindings/` 目录，由 Wails 构建时自动生成。

### 事件系统

前后端通过 Wails 事件系统进行异步通信：

| 事件名 | 方向 | 数据 | 说明 |
|--------|------|------|------|
| `time` | 后端→前端 | `string` (RFC1123 时间) | 每秒广播当前时间 |
| `ssh:connections-updated` | 后端→前端 | `{connections, timestamp}` | 全局连接状态同步（500ms） |
| `ssh:group-updated` | 后端→前端 | `{groupID, action, connections}` | 分组状态同步（500ms） |
| `ssh:terminal-output` | 后端→前端 | `{connID, sessionID, data}` | 终端输出数据（20ms 轮询） |
| `ssh:tray-hide` | 前端→后端 | 无 | 请求隐藏主窗口到托盘 |
| `ai:tool-approval` | 后端→前端 | `{callID, toolName, args, description}` | AI 工具调用审批请求 |
| `ai:execute-tool` | 后端→前端 | `{callID, toolName, args}` | AI 工具执行请求 |
| `ai:submit-tool-result` | 前端→后端 | `{callId, result}` | AI 工具执行结果 |
| `ai:ai-terminal-ready` | 前端→前端 | `{sessionId, connId}` | AI 终端就绪通知 |
| `cloud:status` | 后端→前端 | `{connected, error?}` | 云端连接状态变化 |

### 数据流概览

```
┌──────────┐   Wails Bindings   ┌──────────┐
│  前端     │ ──────────────────► │  后端     │
│  (Vue 3) │ ◄────────────────── │  (Go)    │
└──────────┘   Wails Events      └──────────┘
     │                                    │
     │  用户操作                           │  SSH 连接
     │  ├─ 新建连接 ──► CreateAndConnect   │  ├─ SSHClient.Connect()
     │  ├─ 执行命令 ──► RunCommand         │  ├─ SSHClient.ExecuteCommand()
     │  ├─ 文件操作 ──► SFTP 方法          │  ├─ SSHClient.SFTP 操作
     │  └─ AI 对话 ──► SendMessage         │  └─ AIService → API
     │                                    │
     │  事件监听                           │  事件发射
     │  ├─ ssh:terminal-output ◄── 20ms轮询│
     │  ├─ ssh:connections-updated ◄── 500ms│
     │  └─ ai:tool-approval ◄── 工具调用   │
```

---

## 7. 数据持久化

### 数据目录结构

```
data/
├── config.json              # 应用配置（ConfigService）
├── window_positions.json    # 窗口位置记忆（WindowManager）
├── connections_cache.json   # 连接缓存（StorageManager，运行时）
├── connections_saved.json   # 连接永久保存（StorageManager，加密）
├── ai/
│   ├── ai_config.json       # AI 配置
│   ├── projects/            # AI 项目数据
│   ├── sessions/            # AI 会话数据
│   └── history/             # AI 聊天历史
└── logs/
    └── ai.log               # AI 操作日志
```

### 加密方案

- **连接密码**: AES-256-GCM 加密，密钥硬编码在程序中
- **云端通信**: WebSocket + TLS（客户端默认跳过证书验证 `InsecureSkipVerify: true`）
- **云端服务器**: AES-256-GCM 加密 + 令牌验证 + 请求签名

---

## 8. 构建与运行

### 开发环境

```bash
# 1. 安装前端依赖
cd frontend && npm install

# 2. 开发模式运行
wails3 dev

# 3. 构建生产版本
wails3 package
```

### 构建任务 (Taskfile)

| 任务 | 命令 | 说明 |
|------|------|------|
| 开发模式 | `task dev` | 启动开发服务器 |
| 构建 | `task build` | 构建当前平台 |
| 打包 | `task package` | 打包生产版本 |
| 服务器模式构建 | `task build:server` | 构建 HTTP 服务器模式（无 GUI） |
| Docker 构建 | `task build:docker` | 构建 Docker 镜像 |

### 跨平台构建

| 平台 | 构建目录 | 说明 |
|------|----------|------|
| Windows | `build/windows/` | NSIS 安装器、MSIX 打包 |
| macOS | `build/darwin/` | .icns 图标、plist 配置 |
| Linux | `build/linux/` | AppImage、nfpm（deb/rpm） |
| Docker | `build/docker/` | 交叉编译、服务器模式 |

### 系统要求

- **运行**: Windows 10+（主要支持），macOS/Linux 理论支持
- **开发**: Go 1.25+、Node.js、Wails v3 CLI
- **下载**: [GitHub Releases](https://github.com/nanxiangxi/0qssh/releases) 提供 `0qssh-amd64.exe` 和 `0qssh-arm64.exe`
