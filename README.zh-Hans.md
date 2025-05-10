[![version](https://img.shields.io/badge/version-1.0.0-blue.svg)]()
[![license](https://img.shields.io/github/license/ihugang/AutoShutdown)]()
[![platform](https://img.shields.io/badge/platform-Windows-lightgrey)]()
[![language](https://img.shields.io/badge/language-golang-orange)]()
> 🌐 [View this README in English](./README.md)


🖥️ AutoShutdown 是一个适用于 Windows 的自动关机/休眠工具，支持定时操作与远程控制（TCP/UDP）。

## 功能特点
- ⏰ 定时自动关机或休眠
- 🧠 后台运行，占用资源小
- 🌐 支持 TCP/UDP 远程控制指令
- 🔒 可用于家长控制孩子使用电脑时间

## 使用场景
- 控制孩子电脑使用时长
- 节能降耗，定时关机
- 局域网远程关机

## 快速开始
1.	克隆仓库
```bash
    git clone https://github.com/ihugang/AutoShutdown.git
    cd AutoShutdown
```
2.	使用 Visual Studio 打开项目并编译。
3.	配置定时规则和远程端口。

## TCP/UDP 远程控制（预览功能）
**默认端口：** 9527（可配置）

**支持命令：**
- shutdown → 执行关机
- sleep → 执行休眠

注意：使用前请检查防火墙设置是否允许对应端口通信。

⸻

## License

MIT License
Copyright (c) 2025
