[![version](https://img.shields.io/badge/version-1.0.0-blue.svg)]()
[![license](https://img.shields.io/github/license/ihugang/AutoShutdown)]()
[![platform](https://img.shields.io/badge/platform-Windows(x64/ARM64)-lightgrey)]()
[![language](https://img.shields.io/badge/language-golang-orange)]()
![visitors](https://visitor-badge.laobi.icu/badge?page_id=ihugang.AutoShutdown)
> 🌐 [View this README in English](./README.md)


🖥️ AutoShutdown 是一个适用于 Windows 的自动关机/休眠工具，支持定时操作与远程控制（TCP/UDP）。

## 功能特点
- ⭐ 定时自动关机或休眠，支持随机延迟
- 🧠 后台运行，占用资源小
- 🌐 支持 TCP/UDP 远程控制指令
- 🔒 可用于家长控制孩子使用电脑时间
- ⚠️ 关机/休眠前显示警告对话框，可配置提前时间

## 使用场景
- 控制孩子电脑使用时长
- 节能降耗，定时关机
- 局域网远程关机

## 随机延迟机制

为了使自动关机/休眠操作更加自然，AutoShutdown 采用了随机延迟机制：

- 当进入设定的时间范围后，系统不会立即执行关机/休眠
- 而是会随机选择接下来 1-10 分钟内的任意时间点
- 这种随机性可以避免用户预测确切的关机时间
- 同时也给予用户一定的缓冲时间来保存工作

例如，如果设置关机时间为 22:00，系统会在 22:00 到 22:10 之间的随机时间点执行关机或休眠操作。

## 关机前警告功能

AutoShutdown 可以在执行关机或休眠操作前显示警告对话框：

- 给用户提供保存工作的机会
- 可配置警告提前时间（默认：操作前 5 分钟）
- 可通过命令行或远程命令启用/禁用
- 用户可以选择继续操作或取消操作

## 快速开始

### 1. 克隆仓库
```bash
git clone https://github.com/ihugang/AutoShutdown.git
cd AutoShutdown
```

### 2. 编译项目

> **注意：** 由于本项目使用了 Windows 特有的 API，强烈建议在 Windows 环境下进行编译。在 macOS 或 Linux 等非 Windows 系统上交叉编译可能会遇到依赖问题。

#### 在 Windows 上编译（推荐）

```bash
# 安装依赖包
go mod tidy

# 编译 x64 版本
go build -o AutoShutdown-amd64.exe ./src

# 编译 ARM64 版本
set GOARCH=arm64
go build -o AutoShutdown-arm64.exe ./src
```

#### 在非 Windows 系统上交叉编译（可能需要额外配置）

```bash
# 安装依赖包
go mod tidy

# 编译 Windows x64 版本
GOOS=windows GOARCH=amd64 go build -tags windows -o AutoShutdown-amd64.exe ./src

# 编译 Windows ARM64 版本
GOOS=windows GOARCH=arm64 go build -tags windows -o AutoShutdown-arm64.exe ./src
```

#### 推荐的编译环境

- Windows 10/11 + Go 1.18 或更高版本
- Visual Studio Code + Go 插件

### 3. 配置和运行

配置定时规则、远程端口和警告设置，然后运行程序。

#### 命令行选项

##### 全部可用参数

| 参数 | 说明 | 默认值 |
|----------|---------|--------|
| `-mode` | 操作模式: shutdown(关机), hibernate(休眠), reboot(重启), logoff(注销) | `hibernate` |
| `-tcp` | TCP远程控制端口 | `2200` |
| `-udp` | UDP远程控制端口 | `2200` |
| `-remote` | 是否启用远程控制 | `true` |
| `-warning` | 是否显示关机/休眠前警告 | `true` |
| `-warning-time` | 关机/休眠前警告提前时间(分钟) | `5` |
| `-start-hour` | 开始时间(小时, 0-23) | `22` |
| `-start-minute` | 开始时间(分钟, 0-59) | `0` |
| `-end-hour` | 结束时间(小时, 0-23) | `23` |
| `-end-minute` | 结束时间(分钟, 0-59) | `59` |
| `-start-time` | 开始时间(HH:MM格式), 会覆盖 start-hour 和 start-minute | - |
| `-end-time` | 结束时间(HH:MM格式), 会覆盖 end-hour 和 end-minute | - |
| `-lang` | 语言: en(英文), zh-Hans(简体中文) | `en` |
| `-version` | 显示版本信息 | `false` |

##### 使用示例

```bash
# 使用默认设置运行
AutoShutdown.exe

# 禁用警告对话框
AutoShutdown.exe -warning=false

# 将警告时间改为关机前10分钟
AutoShutdown.exe -warning-time=10

# 设置开始时间和结束时间（单独设置小时和分钟）
AutoShutdown.exe -start-hour=21 -start-minute=30 -end-hour=6 -end-minute=30

# 使用HH:MM格式设置时间范围
AutoShutdown.exe -start-time=21:30 -end-time=06:30

# 完整配置示例
AutoShutdown.exe -mode=hibernate -tcp=2200 -udp=2200 -remote=true -warning=true -warning-time=5 -start-time=22:00 -end-time=06:00 -lang=zh-Hans
```

##### 安装为服务的完整示例

```bash
# 安装为 Windows 服务，使用自定义设置
AutoShutdown.exe -mode=hibernate -warning=true -warning-time=10 -start-time=22:30 -end-time=06:30 -lang=zh-Hans install

# 启动服务
AutoShutdown.exe start
```

## TCP/UDP 远程控制

### 端口配置

- **默认 TCP/UDP 端口**: 2200（可配置）
- **命令行选项**:
  ```
  AutoShutdown.exe -tcp=2200 -udp=2200
  ```
- **安全性**: 请确保调整防火墙规则以允许这些端口通信

### 已安装服务的端口修改

如果已将 AutoShutdown 安装为 Windows 服务，需要按以下步骤更改端口：

1. **停止服务**：
   ```
   AutoShutdown.exe stop
   ```

2. **卸载服务**：
   ```
   AutoShutdown.exe remove
   ```

3. **使用新端口重新安装**：
   ```
   AutoShutdown.exe -tcp=8080 -udp=8080 install
   ```

4. **启动服务**：
   ```
   AutoShutdown.exe start
   ```

注意：更改端口后，请确保相应调整防火墙规则。

### 连接方法

#### TCP 连接（交互式菜单）

```bash
# Windows
telnet <目标IP> 2200

# macOS（无内置telnet）
nc <目标IP> 2200

# Linux
telnet <目标IP> 2200
# 或
nc <目标IP> 2200
```

#### UDP 命令

```bash
# Windows (PowerShell)
$endpoint = New-Object System.Net.IPEndPoint([System.Net.IPAddress]::Parse("<目标IP>"), 2200)
$client = New-Object System.Net.Sockets.UdpClient
$bytes = [System.Text.Encoding]::ASCII.GetBytes("hibernate")
$client.Send($bytes, $bytes.Length, $endpoint)
$client.Close()

# macOS/Linux
echo "hibernate" | nc -u <目标IP> 2200
```

### 可用命令

- `shutdown`: 关机
- `hibernate`: 休眠（默认操作）
- `reboot`: 重启计算机
- `logoff`: 注销当前用户
- `status`: 查看系统状态
- `setmode <mode>`: 设置操作模式（shutdown, hibernate, reboot, logoff）
- `settime start HH:MM`: 设置开始时间
- `settime end HH:MM`: 设置结束时间
- `setwarning on [minutes]`: 启用关机警告（可选指定分钟数）
- `setwarning off`: 禁用关机警告
- `help`: 显示帮助信息
- `menu`: 显示交互式菜单（仅TCP模式）

⸻

## License

MIT License
Copyright (c) 2025
