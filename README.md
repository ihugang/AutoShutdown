[![version](https://img.shields.io/badge/version-1.0.0-blue.svg)]()
[![license](https://img.shields.io/github/license/ihugang/AutoShutdown)]()
[![platform](https://img.shields.io/badge/platform-Windows-lightgrey)]()
[![language](https://img.shields.io/badge/language-golang-orange)]()
> 🌐 [简体中文文档 / 中文版说明](./README.zh-Hans.md)
# AutoShutdown

> 🖥️ A Windows utility for scheduled and remotely controlled shutdown/sleep.  
> 📅 Supports timer-based control and now includes TCP/UDP command support for remote management.

## Features

- ⏰ Schedule automatic shutdown or sleep
- 🧠 Lightweight and runs quietly in background
- 🌐 Remote command support via **TCP/UDP**
- 🔒 Useful for parental control and personal PC automation

## Use Cases

- Enforcing screen time for kids
- Power-saving automation
- Remote PC shutdown in home network

## Getting Started

1. Clone the repo
   ```bash
   git clone https://github.com/ihugang/AutoShutdown.git
   cd AutoShutdown
   ```
2.	Open the project in Visual Studio and build the solution.
3.	Configure your schedule and remote port.

TCP/UDP Remote Control (Preview)
	•	Port: Default 9527 (configurable)
	•	Commands:
	•	shutdown → triggers system shutdown
	•	sleep → triggers system sleep

Note: Firewall rules may need to be adjusted.

## License

MIT License
Copyright (c) 2025

⸻

Made with ❤️ by Hu Gang