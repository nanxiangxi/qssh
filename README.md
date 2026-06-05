# 0qssh

一个中文的 SSH 远程连接工具，开源、好看、便携。

## 关于

我缺乏一个好看的 SSH 工具给自己用，所以自己写了一个。

设计初衷很简单：**好看第一，好用第二**。不追求极致性能，只追求视觉上的舒适。

**官方网站**：[ssh.0qpz.com](ssh.0qpz.com)

## 特点

- 中文界面，开箱即用
- 暗色主题，好看优先
- 便携式，双击即用，无需安装
- 多标签终端管理
- SFTP 文件管理
- 内置 AI 助手
- 端口转发（本地/远程）
- 防火墙管理（iptables / firewalld / ufw）
- 进程守护（systemd 服务管理）
- 批量命令执行
- 连接管理（保存/编辑/快速连接）

## 系统要求

- Windows 10 或更高版本

## 下载

从 [Releases](https://github.com/nanxiangxi/0qssh/releases) 页面下载：

- `0qssh-amd64.exe` — 适用于大多数 Windows 电脑
- `0qssh-arm64.exe` — 适用于 ARM 架构设备

下载后直接双击运行，无需安装。

## 平台支持

| 平台 | 状态 |
|---|---|
| Windows 10+ | 主要支持，持续更新 |
| macOS | 理论支持，不发布、不测试 |
| Linux | 理论支持，不发布、不测试 |

## 更新

更新速率取决于我的个人需求。我在日常使用这个软件，会按照个人喜好和使用需求来更新。

有功能需求或 bug 反馈，欢迎提交 issue。

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

本项目基于 **CC BY-NC-SA 4.0** 协议开源。

**你可以：**
- 自由使用、复制、分发本软件
- 修改、二次开发
- 在相同协议下分发修改后的版本

**你必须：**
- 保留原作者署名
- 修改后的作品必须以相同协议开源

**你不能：**
- 将本软件用于商业用途
- 将本软件闭源分发

详情请参阅 [CC BY-NC-SA 4.0](https://creativecommons.org/licenses/by-nc-sa/4.0/deed.zh-hans)
