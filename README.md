[![version](https://img.shields.io/badge/version-1.0.0-blue.svg)]()
[![license](https://img.shields.io/github/license/ihugang/AutoShutdown)]()
[![platform](https://img.shields.io/badge/platform-Windows(x64/ARM64)-lightgrey)]()
[![language](https://img.shields.io/badge/language-golang-orange)]()
> 🌐 [简体中文文档 / 中文版说明](./README.zh-Hans.md)
# AutoShutdown

> 🖥️ A Windows utility for scheduled and remotely controlled shutdown/sleep.  
> 📅 Supports timer-based control and now includes TCP/UDP command support for remote management.

## Features

- ⏰ Schedule automatic shutdown or sleep with random delay
- 🧠 Lightweight and runs quietly in background
- 🌐 Remote command support via **TCP/UDP**
- 🔒 Useful for parental control and personal PC automation

## Use Cases

- Enforcing screen time for kids
- Power-saving automation
- Remote PC shutdown in home network

## Random Delay Mechanism

To make automatic shutdown/hibernate operations more natural, AutoShutdown implements a random delay mechanism:

- When entering the specified time range, the system will not immediately execute shutdown/hibernate
- Instead, it randomly selects a time point within the next 1-10 minutes
- This randomness prevents users from predicting the exact shutdown time
- It also provides users with a buffer period to save their work

For example, if the shutdown time is set to 22:00, the system will execute the shutdown or hibernate operation at a random time between 22:00 and 22:10.

## Getting Started

### 1. Clone the Repository
```bash
git clone https://github.com/ihugang/AutoShutdown.git
cd AutoShutdown
```

### 2. Build the Project

> **Note:** Since this project uses Windows-specific APIs, it is strongly recommended to build it in a Windows environment. Cross-compiling on non-Windows systems like macOS or Linux may encounter dependency issues.

#### Building on Windows (Recommended)

```bash
# Install dependencies
go mod tidy

# Build for x64
go build -o AutoShutdown-amd64.exe ./src

# Build for ARM64
set GOARCH=arm64
go build -o AutoShutdown-arm64.exe ./src
```

#### Cross-compiling on Non-Windows Systems (May require additional configuration)

```bash
# Install dependencies
go mod tidy

# Build for Windows x64
GOOS=windows GOARCH=amd64 go build -tags windows -o AutoShutdown-amd64.exe ./src

# Build for Windows ARM64
GOOS=windows GOARCH=arm64 go build -tags windows -o AutoShutdown-arm64.exe ./src
```

#### Recommended Build Environment

- Windows 10/11 with Go 1.18 or higher
- Visual Studio Code with Go extension

### 3. Configure and Run

Configure your schedule and remote port, then run the program.

## TCP/UDP Remote Control

### Port Configuration

- **Default TCP/UDP Port**: 2200 (configurable)
- **Command Line Options**:
  ```
  AutoShutdown.exe -tcp=2200 -udp=2200
  ```
- **Security**: Make sure to adjust firewall rules to allow these ports

### Changing Ports for Installed Service

If you have already installed AutoShutdown as a Windows service, follow these steps to change the ports:

1. **Stop the service**:
   ```
   AutoShutdown.exe stop
   ```

2. **Remove the service**:
   ```
   AutoShutdown.exe remove
   ```

3. **Reinstall with new ports**:
   ```
   AutoShutdown.exe -tcp=8080 -udp=8080 install
   ```

4. **Start the service**:
   ```
   AutoShutdown.exe start
   ```

Note: After changing ports, make sure to update your firewall rules accordingly.

### Connection Methods

#### TCP Connection (Interactive Menu)

```bash
# Windows
telnet <target-ip> 2200

# macOS (no built-in telnet)
nc <target-ip> 2200

# Linux
telnet <target-ip> 2200
# or
nc <target-ip> 2200
```

#### UDP Commands

```bash
# Windows (PowerShell)
$endpoint = New-Object System.Net.IPEndPoint([System.Net.IPAddress]::Parse("<target-ip>"), 2200)
$client = New-Object System.Net.Sockets.UdpClient
$bytes = [System.Text.Encoding]::ASCII.GetBytes("hibernate")
$client.Send($bytes, $bytes.Length, $endpoint)
$client.Close()

# macOS/Linux
echo "hibernate" | nc -u <target-ip> 2200
```

### Available Commands

- `shutdown`: Shutdown the computer
- `hibernate`: Hibernate the computer (default action)
- `reboot`: Restart the computer
- `logoff`: Log off the current user
- `status`: View system status
- `setmode <mode>`: Set operation mode (shutdown, hibernate, reboot, logoff)
- `settime start HH:MM`: Set start time
- `settime end HH:MM`: Set end time
- `help`: Show help information
- `menu`: Show interactive menu (TCP only)

## License

MIT License
Copyright (c) 2025

⸻

Made with ❤️ by Hu Gang