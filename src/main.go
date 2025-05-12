//go:build windows
// +build windows

// AutoShutdown - Automatic shutdown/hibernate tool for Windows
// Supports scheduled operations and remote control (TCP/UDP)
// Only for Windows operating system
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
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

const (
	// Version information
	VERSION     = "1.00"
	VERSION_DATE = "2025-05-11"
)

var (
	arg                  string
	tcpPort              string
	udpPort              string
	remoteControlEnabled bool
	showVersion          bool
	language             string
	operationMode        string = "hibernate" // Default operation mode: shutdown, hibernate, reboot, logoff
	showWarning          bool   = true       // Whether to show warning before shutdown/hibernate
	warningMinutes       int    = 5          // Minutes to warn before shutdown/hibernate
	debugMode            bool   = false      // Debug mode for detailed logging
	logFile              string = ""         // Log file path for debug mode
	warningShown         bool   = false      // 跟踪是否已显示过警告对话框

	// Automatic shutdown time settings
	shutdownStartHour   int        = 22 // Start time (hour)
	shutdownStartMinute int        = 0  // Start time (minute)
	shutdownEndHour     int        = 23 // End time (hour)
	shutdownEndMinute   int        = 59 // End time (minute)
	shutdownMutex       sync.Mutex      // Mutex for protecting time settings
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

// 用于解析时间字符串的变量
var startTimeStr string
var endTimeStr string

func init() {
	// Initialize random number generator
	rand.Seed(time.Now().UnixNano())
	
	// Parse command line flags
	flag.StringVar(&arg, "uFlags", "hibernate", "shutdown hibernate logoff reboot")
	flag.StringVar(&tcpPort, "tcp", "2200", "TCP port for remote control")
	flag.StringVar(&udpPort, "udp", "2200", "UDP port for remote control")
	flag.BoolVar(&remoteControlEnabled, "remote", true, "Enable remote control")
	flag.StringVar(&operationMode, "mode", "hibernate", "Operation mode: shutdown, hibernate, reboot, logoff")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.StringVar(&language, "lang", "en", "Language (en, zh-Hans)")
	flag.BoolVar(&showWarning, "warning", true, "Show warning before shutdown/hibernate")
	flag.IntVar(&warningMinutes, "warning-time", 5, "Minutes to warn before shutdown/hibernate")
	
	// Time range settings
	flag.IntVar(&shutdownStartHour, "start-hour", 22, "Start hour (0-23)")
	flag.IntVar(&shutdownStartMinute, "start-minute", 0, "Start minute (0-59)")
	flag.IntVar(&shutdownEndHour, "end-hour", 23, "End hour (0-23)")
	flag.IntVar(&shutdownEndMinute, "end-minute", 59, "End minute (0-59)")
	
	// Alternative time format
	flag.StringVar(&startTimeStr, "start-time", "", "Start time in HH:MM format (e.g. 22:00)")
	flag.StringVar(&endTimeStr, "end-time", "", "End time in HH:MM format (e.g. 23:59)")
	
	// Debug mode settings
	flag.BoolVar(&debugMode, "debug", false, "Enable debug mode with detailed logging")
	flag.StringVar(&logFile, "log-file", "AutoShutdown.log", "Log file path for debug mode")
}

func main() {
	// Parse command line arguments
	flag.Parse()
	
	// 处理时间字符串格式
	if startTimeStr != "" {
		parts := strings.Split(startTimeStr, ":")
		if len(parts) == 2 {
			if h, err := strconv.Atoi(parts[0]); err == nil && h >= 0 && h <= 23 {
				shutdownStartHour = h
			}
			if m, err := strconv.Atoi(parts[1]); err == nil && m >= 0 && m <= 59 {
				shutdownStartMinute = m
			}
		}
	}
	
	if endTimeStr != "" {
		parts := strings.Split(endTimeStr, ":")
		if len(parts) == 2 {
			if h, err := strconv.Atoi(parts[0]); err == nil && h >= 0 && h <= 23 {
				shutdownEndHour = h
			}
			if m, err := strconv.Atoi(parts[1]); err == nil && m >= 0 && m <= 59 {
				shutdownEndMinute = m
			}
		}
	}
	
	// Set language
	if language != "" {
		SetLanguage(language)
	}
	
	// 设置调试日志
	if debugMode {
		// 配置日志输出到文件
		logWriter, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("无法打开日志文件: %v\n", err)
			os.Exit(1)
		}
		
		// 设置日志输出到文件和控制台
		multiWriter := io.MultiWriter(logWriter, os.Stdout)
		log.SetOutput(multiWriter)
		
		// 设置日志格式，包含时间戳
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
		
		log.Println("===== 调试模式已启用 =====")
		log.Printf("版本: %s (%s)", VERSION, VERSION_DATE)
		log.Printf("操作模式: %s", operationMode)
		log.Printf("时间范围: %02d:%02d - %02d:%02d", shutdownStartHour, shutdownStartMinute, shutdownEndHour, shutdownEndMinute)
		log.Printf("警告设置: 启用=%v, 提前时间=%d分钟", showWarning, warningMinutes)
		log.Printf("远程控制: 启用=%v, TCP端口=%s, UDP端口=%s", remoteControlEnabled, tcpPort, udpPort)
		log.Printf("语言: %s", language)
		log.Printf("日志文件: %s", logFile)
		log.Println("==============================")
	}
	
	// Show version information
	if showVersion {
		fmt.Printf(T("version_info", T("app_name"), VERSION, VERSION_DATE) + "\n")
		fmt.Println(T("developed_by"))
		os.Exit(0)
	}
	
	svcConfig := &service.Config{
		Name:        "EarlySleepService",                          // Service display name
		DisplayName: "EarlySleep",                                 // Service name
		Description: "If u are a student,u should sleep earlier.", // Service description
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
	// 记录上次检测到进入时间范围的时间
	var lastEnteredPeriod time.Time
	// 记录是否已经计划了一次随机关机
	var shutdownScheduled bool = false
	// 记录计划的关机时间
	var scheduledShutdownTime time.Time
	
	// 调试模式下记录初始化信息
	if debugMode {
		log.Println("[DEBUG] doIt函数已启动，开始监控时间范围")
	}
	
	for {
		now := time.Now()
		hour := now.Hour()
		minute := now.Minute()
		second := now.Second()

		// 获取当前的关机时间设置
		shutdownMutex.Lock()
		startHour := shutdownStartHour
		startMinute := shutdownStartMinute
		endHour := shutdownEndHour
		endMinute := shutdownEndMinute
		currentMode := operationMode
		shutdownMutex.Unlock()
		
		// 调试模式下每分钟记录一次当前状态
		if debugMode && second == 0 {
			log.Printf("[DEBUG] 当前时间: %02d:%02d:%02d", hour, minute, second)
			log.Printf("[DEBUG] 时间范围: %02d:%02d - %02d:%02d", startHour, startMinute, endHour, endMinute)
			log.Printf("[DEBUG] 操作模式: %s", currentMode)
			log.Printf("[DEBUG] 警告设置: 启用=%v, 提前时间=%d分钟", showWarning, warningMinutes)
			if shutdownScheduled {
				log.Printf("[DEBUG] 已计划关机时间: %s", scheduledShutdownTime.Format("15:04:05"))
				log.Printf("[DEBUG] 距离计划关机还有: %v", scheduledShutdownTime.Sub(now))
			} else {
				log.Printf("[DEBUG] 尚未计划关机时间")
			}
		}

		// 检查当前时间是否在关机时间范围内
		inShutdownPeriod := false

		// 如果开始时间小于结束时间，表示在同一天内
		if startHour < endHour || (startHour == endHour && startMinute <= endMinute) {
			// 检查当前时间是否在范围内
			if (hour > startHour || (hour == startHour && minute >= startMinute)) &&
				(hour < endHour || (hour == endHour && minute < endMinute)) {
				inShutdownPeriod = true
				if debugMode {
					log.Printf("[DEBUG] 当前时间在范围内 (同一天内时间范围)")
				}
			} else if debugMode {
				log.Printf("[DEBUG] 当前时间不在范围内 (同一天内时间范围)")
				log.Printf("[DEBUG] 检查条件: (%d > %d || (%d == %d && %d >= %d)) && (%d < %d || (%d == %d && %d < %d))", 
					hour, startHour, hour, startHour, minute, startMinute, hour, endHour, hour, endHour, minute, endMinute)
			}
		} else { // 开始时间大于结束时间，跨天时间范围（如晚上22点到次日早上6点）
			if (hour > startHour || (hour == startHour && minute >= startMinute)) ||
				(hour < endHour || (hour == endHour && minute < endMinute)) {
				inShutdownPeriod = true
				if debugMode {
					log.Printf("[DEBUG] 当前时间在范围内 (跨天时间范围)")
				}
			} else if debugMode {
				log.Printf("[DEBUG] 当前时间不在范围内 (跨天时间范围)")
				log.Printf("[DEBUG] 检查条件: (%d > %d || (%d == %d && %d >= %d)) || (%d < %d || (%d == %d && %d < %d))", 
					hour, startHour, hour, startHour, minute, startMinute, hour, endHour, hour, endHour, minute, endMinute)
			}
		}
		
		// 调试模式下记录时间范围检查结果
		if debugMode && second == 0 {
			log.Printf("[DEBUG] 时间范围检查结果: inShutdownPeriod=%v", inShutdownPeriod)
		}

		// 如果刚进入时间范围，计算随机关机时间
		if inShutdownPeriod {
			// 如果是新进入时间范围，或者上次进入已经超过12小时（防止时钟调整等异常情况）
			if lastEnteredPeriod.IsZero() || now.Sub(lastEnteredPeriod) > 12*time.Hour {
				lastEnteredPeriod = now
				shutdownScheduled = false
				if debugMode {
					log.Printf("[DEBUG] 新进入时间范围或重置状态")
				}
			}
			
			// 如果还没有计划关机时间，则计算一个随机时间
			if !shutdownScheduled {
				// 生成一个0-10分钟内的随机延迟
				randomMinutes := rand.Intn(10) + 1 // 1-10分钟
				randomSeconds := rand.Intn(60)     // 0-59秒
				delay := time.Duration(randomMinutes)*time.Minute + time.Duration(randomSeconds)*time.Second
				
				// 计算关机时间
				scheduledShutdownTime = now.Add(delay)
				shutdownScheduled = true
				
				log.Printf("当前时间 %02d:%02d，在时间范围内（%02d:%02d-%02d:%02d）\n", 
					hour, minute, startHour, startMinute, endHour, endMinute)
				log.Printf("已计划在 %s 执行%s操作（随机延迟%d分%d秒）\n",
					scheduledShutdownTime.Format("15:04:05"), getOperationName(currentMode), randomMinutes, randomSeconds)
				
				if debugMode {
					log.Printf("[DEBUG] 计算了新的关机时间: %s", scheduledShutdownTime.Format("15:04:05"))
					log.Printf("[DEBUG] 随机延迟: %d分%d秒", randomMinutes, randomSeconds)
					if showWarning {
						// 计算警告时间
						warningTime := scheduledShutdownTime.Add(-time.Duration(warningMinutes) * time.Minute)
						log.Printf("[DEBUG] 警告将在 %s 显示（提前%d分钟）", 
							warningTime.Format("15:04:05"), warningMinutes)
					} else {
						log.Printf("[DEBUG] 警告功能已禁用")
					}
				}
			}
			
			// 如果已经到了计划的关机时间，执行关机
			if shutdownScheduled {
				// 计算警告时间
				warningTime := scheduledShutdownTime.Add(-time.Duration(warningMinutes) * time.Minute)
				
				// 使用包级变量跟踪警告状态
				// warningShown变量已在包级声明
				
				// 如果启用了警告并且当前时间已过警告时间但还未到关机时间
				if showWarning && now.After(warningTime) && now.Before(scheduledShutdownTime) && !warningShown {
					if debugMode {
						log.Printf("[DEBUG] 当前时间 %s 已过警告时间 %s，准备显示警告", 
							now.Format("15:04:05"), warningTime.Format("15:04:05"))
					}
					
					// 计算实际剩余时间（分钟）
					remainMinutes := int(scheduledShutdownTime.Sub(now).Minutes())
					// 如果剩余时间小于1分钟，至少显示1分钟
					if remainMinutes < 1 {
						remainMinutes = 1
					}
					if debugMode {
						log.Printf("[DEBUG] 实际剩余时间: %d分钟", remainMinutes)
					}
					
					// 显示警告对话框，传入实际剩余时间
					warningResult := showWarningDialog(operationMode, remainMinutes)
					warningShown = true
					
					if debugMode {
						log.Printf("[DEBUG] 警告对话框结果: %v", warningResult)
					}
					
					// 如果用户取消了操作
					if !warningResult {
						log.Printf(T("shutdown_cancelled", getOperationName(operationMode)))
						shutdownScheduled = false
						lastEnteredPeriod = time.Time{} // 重置为零值
						return
					}
				}
				
				// 如果已经到了计划的关机时间
				if now.After(scheduledShutdownTime) {
					// 重置警告标志，为下一次关机做准备
					warningShown = false
					log.Printf("当前时间 %02d:%02d，已到计划的时间，执行%s操作\n",
						hour, minute, getOperationName(currentMode))
					
					if debugMode {
						log.Printf("[DEBUG] 准备执行%s操作", getOperationName(currentMode))
					}
					
					// 执行操作并重置状态
					performOperation(currentMode)
					shutdownScheduled = false
					lastEnteredPeriod = time.Time{} // 重置为零值
				}
			}
		} else {
			// 如果不在时间范围内，重置状态
			if shutdownScheduled && debugMode {
				log.Printf("[DEBUG] 不在时间范围内，重置关机计划")
			}
			shutdownScheduled = false
			lastEnteredPeriod = time.Time{} // 重置为零值
		}

		// 每10秒检查一次，以获得更精确的计时
		time.Sleep(10 * time.Second)
	}
}

func shutdown() {
	log.Println(T("executing_operation", T("mode_shutdown")))
	getPrivileges()
	ExitWindowsEx(EWX_SHUTDOWN, 0)
}

func hibernate() {
	log.Println(T("executing_operation", T("mode_hibernate")))
	// Use system command to execute hibernate
	cmd := exec.Command("rundll32.exe", "powrprof.dll,SetSuspendState", "0,1,0")
	err := cmd.Run()
	if err != nil {
		log.Printf(T("log_hibernate_failed", err))

		// If hibernate fails, try to shutdown
		log.Println(T("hibernate_failed"))
		getPrivileges()
		ExitWindowsEx(EWX_SHUTDOWN, 0)
	}
}

func reboot() {
	log.Println(T("executing_operation", T("mode_reboot")))
	getPrivileges()
	ExitWindowsEx(EWX_REBOOT, 0)
}

func logoff() {
	log.Println(T("executing_operation", T("mode_logoff")))
	getPrivileges()
	ExitWindowsEx(EWX_LOGOFF, 0)
}

// 根据操作模式执行相应操作
func performOperation(mode string) {
	// 如果启用了警告，则显示警告对话框
	if showWarning && warningMinutes > 0 {
		if debugMode {
			log.Printf("[DEBUG] 显示关机前警告对话框，操作模式: %s, 提前时间: %d分钟", 
				getOperationName(mode), warningMinutes)
		}
		
		// 显示警告对话框
		warningResult := showWarningDialog(mode, warningMinutes)
		
		if debugMode {
			log.Printf("[DEBUG] 警告对话框结果: %v (真=继续, 假=取消)", warningResult)
		}
		
		if !warningResult {
			// 用户取消了操作
			log.Printf(T("shutdown_cancelled", getOperationName(mode)))
			return
		}
	} else if debugMode {
		log.Printf("[DEBUG] 跳过警告对话框，警告功能已禁用或提前时间为0")
	}

	if debugMode {
		log.Printf("[DEBUG] 准备执行操作: %s", getOperationName(mode))
	}

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
		// 默认使用休眠
		hibernate()
	}
}

// Get localized operation mode name
func getOperationName(mode string) string {
	switch mode {
	case "shutdown":
		return T("mode_shutdown")
	case "hibernate":
		return T("mode_hibernate")
	case "reboot":
		return T("mode_reboot")
	case "logoff":
		return T("mode_logoff")
	default:
		return mode
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

// Start TCP server for remote control
func startTCPServer() {
	addr := fmt.Sprintf(":%s", tcpPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("TCP server start failed: %v\n", err)
		return
	}
	defer listener.Close()

	log.Printf(T("log_tcp_server_started", tcpPort))

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf(T("log_accept_failed", err))
			continue
		}

		go handleTCPConnection(conn)
	}
}

// Handle TCP connection
func handleTCPConnection(conn net.Conn) {
	defer conn.Close()

	log.Printf(T("log_new_tcp_connection", conn.RemoteAddr().String()))
	
	// Show welcome message and interactive menu
	showWelcomeMenu(conn)

	reader := bufio.NewReader(conn)
	
	// Track current state
	var waitingForStartTime bool = false
	var waitingForEndTime bool = false
	
	for {
		// 显示命令提示符
		if !waitingForStartTime && !waitingForEndTime {
			conn.Write([]byte("\n请输入命令或菜单选项 [1-9]: "))
		}
		
		// Read user input
		cmd, err := reader.ReadString('\n')
		if err != nil {
			log.Printf(T("log_command_read_failed", err))
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

// Show welcome menu
func showWelcomeMenu(conn net.Conn) {
	// 构建菜单
	menu := T("welcome_title") + "\n\n"
	menu += fmt.Sprintf(T("menu_item"), 1, T("menu_status")) + "\n"
	menu += fmt.Sprintf(T("menu_item"), 2, T("menu_hibernate")) + "\n"
	menu += fmt.Sprintf(T("menu_item"), 3, T("menu_shutdown")) + "\n"
	menu += fmt.Sprintf(T("menu_item"), 4, T("menu_reboot")) + "\n"
	menu += fmt.Sprintf(T("menu_item"), 5, T("menu_logoff")) + "\n"
	menu += fmt.Sprintf(T("menu_item"), 6, T("menu_set_mode") + " (Hibernate)") + "\n"
	menu += fmt.Sprintf(T("menu_item"), 7, T("menu_set_mode") + " (Shutdown)") + "\n"
	menu += fmt.Sprintf(T("menu_item"), 8, T("menu_set_start_time")) + "\n"
	menu += fmt.Sprintf(T("menu_item"), 9, T("menu_set_end_time")) + "\n"
	menu += fmt.Sprintf(T("menu_item"), 10, "启用关机警告") + "\n"
	menu += fmt.Sprintf(T("menu_item"), 11, "禁用关机警告") + "\n"
	menu += "\n"
	menu += T("menu_prompt")
	
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

// Start UDP server for remote control
func startUDPServer() {
	addr := fmt.Sprintf(":%s", udpPort)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Printf(T("log_udp_addr_failed", err))
		return
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Printf(T("log_udp_listen_failed", err))
		return
	}
	defer conn.Close()

	log.Printf(T("log_udp_server_started", udpPort))

	buf := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf(T("log_udp_read_failed", err))
			continue
		}

		cmd := strings.TrimSpace(string(buf[:n]))
		log.Printf(T("log_udp_command", addr.String(), cmd))

		response := processCommand(cmd)
		conn.WriteToUDP([]byte(response), addr)
	}
}

// Show warning dialog, return true if user confirms to continue
func showWarningDialog(mode string, minutes int) bool {
	// Create warning message
	message := T("shutdown_warning", minutes, getOperationName(mode))
	title := T("shutdown_warning_title", getOperationName(mode))

	if debugMode {
		log.Printf("[DEBUG] 准备显示警告对话框")
		log.Printf("[DEBUG] 标题: %s", title)
		log.Printf("[DEBUG] 消息: %s", message)
	}

	// 使用简单的MessageBox显示警告对话框
	// 这样可以避免中文字符在PowerShell脚本中的编码问题
	powershellCmd := fmt.Sprintf(
		"Add-Type -AssemblyName System.Windows.Forms; $result = [System.Windows.Forms.MessageBox]::Show('%s', '%s', 'OK', 'Warning'); if ($result -eq 'OK') { exit 0 } else { exit 1 }",
		message, title)

	if debugMode {
		log.Printf("[DEBUG] 使用MessageBox显示警告对话框")
		log.Printf("[DEBUG] PowerShell命令: %s", powershellCmd)
	}

	cmd := exec.Command("powershell", "-Command", powershellCmd)
	
	// 捕获命令输出以便调试
	if debugMode {
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		log.Printf("[DEBUG] 警告对话框命令执行结果: %v", err == nil)
		if stdout.Len() > 0 {
			log.Printf("[DEBUG] 命令标准输出: %s", stdout.String())
		}
		if stderr.Len() > 0 {
			log.Printf("[DEBUG] 命令错误输出: %s", stderr.String())
		}
		return err == nil
	} else {
		err := cmd.Run()
		return err == nil
	}
}

// Process remote commands
func processCommand(cmd string) string {
	// Split command and parameters
	parts := strings.Fields(strings.ToLower(cmd))
	if len(parts) == 0 {
		return T("enter_command")
	}

	mainCmd := parts[0]
	switch mainCmd {
	case "version":
		return T("version_info", T("app_name"), VERSION, VERSION_DATE)
		
	case "settime_start_menu":
		return T("enter_start_time")
		
	case "settime_end_menu":
		return T("enter_end_time")
	case "shutdown":
		go shutdown()
		return T("operation_successful", T("mode_shutdown"))

	case "hibernate":
		log.Println(T("executing_operation", T("mode_hibernate")))
		hibernate()
		return T("operation_successful", T("mode_hibernate"))

	case "reboot":
		go reboot()
		return T("operation_successful", T("mode_reboot"))

	case "logoff":
		go logoff()
		return T("operation_successful", T("mode_logoff"))

	case "setmode":
		if len(parts) < 2 {
			return T("invalid_mode")
		}
		
		newMode := parts[1]
		if newMode != "shutdown" && newMode != "hibernate" && newMode != "reboot" && newMode != "logoff" {
			return T("invalid_mode")
		}
		
		operationMode = newMode
		return T("mode_set_success", getOperationName(newMode))

	case "status":
		shutdownMutex.Lock()
		status := T("current_status", 
			shutdownStartHour, shutdownStartMinute, shutdownEndHour, shutdownEndMinute,
			getOperationName(operationMode), VERSION)
		shutdownMutex.Unlock()
		return status

	case "help":
		return T("help_text")
		
	case "setwarning":
		if len(parts) < 2 {
			return "用法: setwarning on/off [minutes]\n例如: setwarning on 5"
		}
		
		switch parts[1] {
		case "on":
			showWarning = true
			// 如果指定了分钟数
			if len(parts) >= 3 {
				if mins, err := strconv.Atoi(parts[2]); err == nil && mins > 0 {
					warningMinutes = mins
					return fmt.Sprintf("警告已启用，提前%d分钟显示", warningMinutes)
				}
			}
			return fmt.Sprintf("警告已启用，提前%d分钟显示", warningMinutes)
			
		case "off":
			showWarning = false
			return "警告已禁用"
			
		default:
			return "用法: setwarning on/off [minutes]\n例如: setwarning on 5"
		}

	case "settime":
		if len(parts) < 3 {
			return T("invalid_time_format")
		}
		
		timeType := parts[1] // start or end
		timeValue := parts[2] // HH:MM
		
		// Parse time
		timeParts := strings.Split(timeValue, ":")
		if len(timeParts) != 2 {
			return T("invalid_time_format")
		}
		
		hour, err1 := strconv.Atoi(timeParts[0])
		minute, err2 := strconv.Atoi(timeParts[1])
		
		if err1 != nil || err2 != nil || hour < 0 || hour > 23 || minute < 0 || minute > 59 {
			return T("invalid_time_format")
		}
		
		// Set time
		shutdownMutex.Lock()
		defer shutdownMutex.Unlock()
		
		if timeType == "start" {
			shutdownStartHour = hour
			shutdownStartMinute = minute
			return T("time_set_success", T("menu_set_start_time"), hour, minute)
		} else if timeType == "end" {
			shutdownEndHour = hour
			shutdownEndMinute = minute
			return T("time_set_success", T("menu_set_end_time"), hour, minute)
		} else {
			return T("invalid_time_type")
		}

	case "language":
		if len(parts) < 2 {
			return T("please_specify_language")
		}
		
		langCode := parts[1]
		if langCode != "en" && langCode != "zh-Hans" {
			return T("invalid_language")
		}
		
		SetLanguage(langCode)
		
		// Use the new language to respond
		return T("language_changed", GetLanguageName(langCode))

	default:
		return T("unknown_command")
	}
}
