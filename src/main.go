//go:build windows
// +build windows

// AutoShutdown - Windows系统专用的自动关机/休眠工具
// 支持定时操作与远程控制（TCP/UDP）
// 仅适用于Windows操作系统
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/CodyGuo/win"
	"github.com/kardianos/service"
)

var (
	arg                  string
	tcpPort              string
	udpPort              string
	remoteControlEnabled bool
	operationMode        string = "hibernate" // 默认操作模式: shutdown(关机), hibernate(休眠), reboot(重启), logoff(注销)

	// 自动关机时间设置
	shutdownStartHour   int        = 22 // 开始时间（小时）
	shutdownStartMinute int        = 0  // 开始时间（分钟）
	shutdownEndHour     int        = 23 // 结束时间（小时）
	shutdownEndMinute   int        = 59 // 结束时间（分钟）
	shutdownMutex       sync.Mutex      // 用于保护时间设置的互斥锁
)

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	// 启动远程控制服务器
	if remoteControlEnabled {
		go startTCPServer()
		go startUDPServer()
	}

	// 启动自动关机功能
	doIt()
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func init() {
	flag.StringVar(&arg, "uFlags", "hibernate", "shutdown hibernate logoff reboot")
	flag.StringVar(&tcpPort, "tcp", "2200", "TCP port for remote control")
	flag.StringVar(&udpPort, "udp", "2200", "UDP port for remote control")
	flag.BoolVar(&remoteControlEnabled, "remote", true, "Enable remote control")
	flag.StringVar(&operationMode, "mode", "hibernate", "Operation mode: shutdown, hibernate, reboot, logoff")
}

func main() {
	svcConfig := &service.Config{
		Name:        "EarlySleepService",                          //服务显示名称
		DisplayName: "EarlySleep",                                 //服务名称
		Description: "If u are a student,u should sleep earlier.", //服务描述
		Option: service.KeyValue{
			"StartType": "automatic",
		},
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		if os.Args[1] == "install" {
			s.Install()
			fmt.Println("服务安装成功")
			// s.Start()
			// fmt.Println("服务启动成功")
			return
		}

		if os.Args[1] == "remove" {
			s.Uninstall()
			fmt.Println("服务卸载成功")
			return
		} else if os.Args[1] == "stop" {
			s.Stop()
			fmt.Println("服务停止成功")
			return
		} else if os.Args[1] == "start" {
			s.Start()
			fmt.Println("服务启动成功")
			return
		} else if os.Args[1] == "status" {
			s.Status()
			return
		}
	}

	err = s.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

}

func doIt() {
	for {
		now := time.Now()
		hour := now.Hour()
		minute := now.Minute()

		// 获取当前的关机时间设置
		shutdownMutex.Lock()
		startHour := shutdownStartHour
		startMinute := shutdownStartMinute
		endHour := shutdownEndHour
		endMinute := shutdownEndMinute
		currentMode := operationMode
		shutdownMutex.Unlock()

		// 检查当前时间是否在关机时间范围内
		inShutdownPeriod := false

		// 如果开始时间小于结束时间，表示在同一天内
		if startHour < endHour || (startHour == endHour && startMinute <= endMinute) {
			// 检查当前时间是否在范围内
			if (hour > startHour || (hour == startHour && minute >= startMinute)) &&
				(hour < endHour || (hour == endHour && minute < endMinute)) {
				inShutdownPeriod = true
			}
		} else { // 开始时间大于结束时间，跨天时间范围（如晚上22点到次日早上6点）
			if (hour > startHour || (hour == startHour && minute >= startMinute)) ||
				(hour < endHour || (hour == endHour && minute < endMinute)) {
				inShutdownPeriod = true
			}
		}

		if inShutdownPeriod {
			log.Printf("当前时间 %02d:%02d，在时间范围内（%02d:%02d-%02d:%02d），执行%s操作\n",
				hour, minute, startHour, startMinute, endHour, endMinute, getOperationName(currentMode))

			// 根据当前模式执行相应操作
			performOperation(currentMode)
		}

		// 每分钟检查一次
		time.Sleep(1 * time.Minute)
	}
}

func shutdown() {
	log.Println("执行关机操作")
	getPrivileges()
	ExitWindowsEx(EWX_SHUTDOWN, 0)
}

func hibernate() {
	log.Println("执行休眠操作")

	// 使用系统命令执行休眠
	cmd := exec.Command("rundll32.exe", "powrprof.dll,SetSuspendState", "0,1,0")
	err := cmd.Run()
	if err != nil {
		log.Printf("休眠命令执行失败: %v\n", err)

		// 如果休眠失败，尝试使用关机
		log.Println("尝试关机...")
		getPrivileges()
		ExitWindowsEx(EWX_SHUTDOWN, 0)
	}
}

func reboot() {
	log.Println("执行重启操作")
	getPrivileges()
	ExitWindowsEx(EWX_REBOOT, 0)
}

func logoff() {
	log.Println("执行注销操作")
	getPrivileges()
	ExitWindowsEx(EWX_LOGOFF, 0)
}

// 根据操作模式执行相应操作
func performOperation(mode string) {
	switch mode {
	case "shutdown":
		shutdown()
	case "hibernate":
		hibernate()
	case "reboot":
		reboot()
	case "logoff":
		logoff()
	default:
		// 默认使用休眠模式
		hibernate()
	}
}

// 获取操作模式的中文名称
func getOperationName(mode string) string {
	switch mode {
	case "shutdown":
		return "关机"
	case "hibernate":
		return "休眠"
	case "reboot":
		return "重启"
	case "logoff":
		return "注销"
	default:
		return "休眠"
	}
}

func getPrivileges() {
	var hToken HANDLE
	var tkp TOKEN_PRIVILEGES

	OpenProcessToken(GetCurrentProcess(), TOKEN_ADJUST_PRIVILEGES|TOKEN_QUERY, &hToken)
	LookupPrivilegeValueA(nil, StringToBytePtr(SE_SHUTDOWN_NAME), &tkp.Privileges[0].Luid)
	tkp.PrivilegeCount = 1
	tkp.Privileges[0].Attributes = SE_PRIVILEGE_ENABLED
	AdjustTokenPrivileges(hToken, false, &tkp, 0, nil, nil)
}

// 启动TCP服务器进行远程控制
func startTCPServer() {
	addr := fmt.Sprintf(":%s", tcpPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("TCP服务器启动失败: %v\n", err)
		return
	}
	defer listener.Close()

	log.Printf("TCP远程控制服务器已启动，监听端口 %s\n", tcpPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("接受连接失败: %v\n", err)
			continue
		}

		go handleTCPConnection(conn)
	}
}

// 处理TCP连接
func handleTCPConnection(conn net.Conn) {
	defer conn.Close()

	log.Printf("新的TCP连接来自: %s\n", conn.RemoteAddr().String())
	
	// 显示欢迎信息和交互菜单
	showWelcomeMenu(conn)

	reader := bufio.NewReader(conn)
	
	// 用于跟踪当前状态
	var waitingForStartTime bool = false
	var waitingForEndTime bool = false
	
	for {
		// 显示命令提示符
		if !waitingForStartTime && !waitingForEndTime {
			conn.Write([]byte("\n请输入命令或菜单选项 [1-9]: "))
		}
		
		// 读取用户输入
		cmd, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("读取命令失败: %v\n", err)
			break
		}

		cmd = strings.TrimSpace(cmd)
		
		// 如果正在等待时间输入
		if waitingForStartTime {
			waitingForStartTime = false
			response := processCommand("settime start " + cmd)
			conn.Write([]byte("\n" + response + "\n"))
			conn.Write([]byte("\n按回车返回菜单..."))
			reader.ReadString('\n')
			showWelcomeMenu(conn)
			continue
		} else if waitingForEndTime {
			waitingForEndTime = false
			response := processCommand("settime end " + cmd)
			conn.Write([]byte("\n" + response + "\n"))
			conn.Write([]byte("\n按回车返回菜单..."))
			reader.ReadString('\n')
			showWelcomeMenu(conn)
			continue
		}
		
		// 处理菜单选项
		if len(cmd) == 1 && cmd >= "1" && cmd <= "9" {
			cmd = getCommandFromMenuOption(cmd)
		}
		
		// 如果用户输入"menu"，显示菜单
		if cmd == "menu" {
			showWelcomeMenu(conn)
			continue
		}
		
		// 处理特殊菜单命令
		if cmd == "settime_start_menu" {
			conn.Write([]byte("\n请输入开始时间（格式为 HH:MM），例如 22:00: "))
			waitingForStartTime = true
			continue
		} else if cmd == "settime_end_menu" {
			conn.Write([]byte("\n请输入结束时间（格式为 HH:MM），例如 06:00: "))
			waitingForEndTime = true
			continue
		}
		
		// 处理命令并返回响应
		response := processCommand(cmd)
		conn.Write([]byte("\n" + response + "\n"))
		
		// 如果是状态命令或帮助命令，显示菜单
		if cmd == "status" || cmd == "help" {
			conn.Write([]byte("\n按回车继续..."))
			reader.ReadString('\n')
			showWelcomeMenu(conn)
		}
	}
}

// 显示欢迎菜单
func showWelcomeMenu(conn net.Conn) {
	// 获取当前状态
	shutdownMutex.Lock()
	status := fmt.Sprintf("当前时间范围: %02d:%02d - %02d:%02d | 操作模式: %s", 
		shutdownStartHour, shutdownStartMinute, shutdownEndHour, shutdownEndMinute,
		getOperationName(operationMode))
	shutdownMutex.Unlock()
	
	// 构建菜单
	menu := fmt.Sprintf(`
╔═══════════════════════════════════════════════════════════════╗
║                 AutoShutdown 远程控制系统                  ║
║                                                              ║
║  %s  ║
╠═══════════════════════════════════════════════════════════════╣
║  [1] 查看系统状态      [2] 立即休眠      [3] 立即关机  ║
║  [4] 立即重启          [5] 立即注销      [6] 设置休眠模式  ║
║  [7] 设置关机模式      [8] 设置开始时间  [9] 设置结束时间  ║
║                                                              ║
║  输入命令或菜单选项编号，输入 'menu' 再次显示此菜单      ║
║  输入 'help' 查看全部可用命令                            ║
╚═══════════════════════════════════════════════════════════════╝
`, status)
	
	// 发送菜单到客户端
	conn.Write([]byte(menu))
}

// 根据菜单选项获取命令
func getCommandFromMenuOption(option string) string {
	switch option {
	case "1":
		return "status"
	case "2":
		return "hibernate"
	case "3":
		return "shutdown"
	case "4":
		return "reboot"
	case "5":
		return "logoff"
	case "6":
		return "setmode hibernate"
	case "7":
		return "setmode shutdown"
	case "8":
		return "settime_start_menu"
	case "9":
		return "settime_end_menu"
	default:
		return "help"
	}
}

// 启动UDP服务器进行远程控制
func startUDPServer() {
	addr := fmt.Sprintf(":%s", udpPort)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Printf("解析UDP地址失败: %v\n", err)
		return
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Printf("UDP服务器启动失败: %v\n", err)
		return
	}
	defer conn.Close()

	log.Printf("UDP远程控制服务器已启动，监听端口 %s\n", udpPort)

	buf := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("读取UDP数据失败: %v\n", err)
			continue
		}

		cmd := strings.TrimSpace(string(buf[:n]))
		log.Printf("收到来自 %s 的UDP命令: %s\n", addr.String(), cmd)

		response := processCommand(cmd)
		conn.WriteToUDP([]byte(response), addr)
	}
}

// 处理远程命令
func processCommand(cmd string) string {
	// 分割命令和参数
	parts := strings.Fields(strings.ToLower(cmd))
	if len(parts) == 0 {
		return "请输入命令"
	}

	mainCmd := parts[0]
	switch mainCmd {
	case "settime_start_menu":
		return "请输入开始时间（格式为 HH:MM），例如 22:00"
		
	case "settime_end_menu":
		return "请输入结束时间（格式为 HH:MM），例如 06:00"
	case "shutdown":
		go shutdown()
		return "正在执行关机操作"

	case "hibernate":
		go hibernate()
		return "正在执行休眠操作"

	case "reboot":
		go reboot()
		return "正在执行重启操作"

	case "logoff":
		go logoff()
		return "正在执行注销操作"

	case "setmode":
		if len(parts) != 2 {
			return "格式错误，正确的格式是: setmode [shutdown|hibernate|reboot|logoff]"
		}

		mode := parts[1]
		if mode != "shutdown" && mode != "hibernate" && mode != "reboot" && mode != "logoff" {
			return "无效的操作模式，必须是: shutdown, hibernate, reboot 或 logoff"
		}

		shutdownMutex.Lock()
		operationMode = mode
		shutdownMutex.Unlock()
		return fmt.Sprintf("操作模式已设置为: %s", getOperationName(mode))

	case "status":
		shutdownMutex.Lock()
		status := fmt.Sprintf("系统状态: 正常运行\n当前时间: %s\n时间范围: %02d:%02d - %02d:%02d\n当前操作模式: %s",
			time.Now().Format("2006-01-02 15:04:05"),
			shutdownStartHour, shutdownStartMinute, shutdownEndHour, shutdownEndMinute,
			getOperationName(operationMode))
		shutdownMutex.Unlock()
		return status

	case "help":
		return `可用命令:
- shutdown: 关闭计算机
- hibernate: 休眠计算机
- reboot: 重启计算机
- logoff: 注销当前用户
- setmode [mode]: 设置操作模式 (shutdown/hibernate/reboot/logoff)
- status: 查看系统状态
- settime start HH:MM: 设置开始时间
- settime end HH:MM: 设置结束时间
- help: 显示帮助信息`

	case "settime":
		if len(parts) != 3 {
			return "格式错误，正确的格式是: settime [start|end] HH:MM"
		}

		timeType := parts[1]  // start 或 end
		timeValue := parts[2] // HH:MM

		// 解析时间格式
		timeParts := strings.Split(timeValue, ":")
		if len(timeParts) != 2 {
			return "时间格式错误，应为 HH:MM"
		}

		hour, err := strconv.Atoi(timeParts[0])
		if err != nil || hour < 0 || hour > 23 {
			return "小时格式错误，应为 0-23 之间的数字"
		}

		minute, err := strconv.Atoi(timeParts[1])
		if err != nil || minute < 0 || minute > 59 {
			return "分钟格式错误，应为 0-59 之间的数字"
		}

		// 设置时间
		shutdownMutex.Lock()
		if timeType == "start" {
			shutdownStartHour = hour
			shutdownStartMinute = minute
			shutdownMutex.Unlock()
			return fmt.Sprintf("关机开始时间已设置为 %02d:%02d", hour, minute)
		} else if timeType == "end" {
			shutdownEndHour = hour
			shutdownEndMinute = minute
			shutdownMutex.Unlock()
			return fmt.Sprintf("关机结束时间已设置为 %02d:%02d", hour, minute)
		} else {
			shutdownMutex.Unlock()
			return "时间类型错误，应为 'start' 或 'end'"
		}

	default:
		return "未知命令，请使用 'help' 查看可用命令"
	}
}
