# 0qssh

一个中文的 SSH 远程连接工具，开源、好看、便携。

## 关于

我缺乏一个好看的 SSH 工具给自己用，所以我自己写了一个。

这个软件的设计初衷很简单：**好看第一，好用第二**。不追求极致性能，只追求视觉上的舒适。

## 特点

- 开源免费
- 界面好看，暗色主题
- 便携式，无需安装
- 中文界面
- 支持多连接管理
- 内置 AI 助手
- 端口转发、防火墙管理、进程守护
- 批量命令执行

## 系统要求

- Windows 10 或更高版本

## 下载

从 [Releases](https://github.com/nanxiangxi/0qssh/releases) 页面下载最新版本的可执行文件。

提供两个版本：
- `0qssh-amd64.exe` — 适用于大多数 Windows 电脑
- `0qssh-arm64.exe` — 适用于 ARM 架构的 Windows 设备

下载后直接双击运行，无需安装。

## 平台支持

| 平台 | 状态 |
|---|---|
| Windows 10+ | 主要支持，持续更新 |
| macOS | 理论支持，不发布、不测试 |
| Linux | 理论支持，不发布、不测试 |

## 更新

更新速率取决于我的个人需求。我在日常使用这个软件，所以会按照我的个人喜好和使用需求来更新。

如果你有功能需求或 bug 反馈，欢迎提交 issue。

## 开发

```bash
# 安装依赖
cd frontend && npm install

# 开发模式
wails3 dev

# 构建便携版（amd64 + arm64）
wails3 package
```

## 技术栈

- 后端：Go + Wails v3
- 前端：Vue 3 + xterm.js + dockview-vue
- SSH：golang.org/x/crypto/ssh

## 许可证

MIT License
