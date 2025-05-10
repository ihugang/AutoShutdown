//go:build windows
// +build windows

// i18n.go - Internationalization support for AutoShutdown
package main

import (
	"fmt"
)

// Language strings map
var langStrings = map[string]map[string]string{
	"en": {
		// General
		"app_name":                "AutoShutdown",
		"developed_by":            "Developed by Hu Gang",
		"version_info":            "%s version: %s (%s)",
		
		// Operation modes
		"mode_shutdown":           "Shutdown",
		"mode_hibernate":          "Hibernate",
		"mode_reboot":             "Reboot",
		"mode_logoff":             "Logoff",
		
		// Status messages
		"current_status":          "Current time range: %02d:%02d - %02d:%02d | Operation mode: %s | Version: %s",
		"scheduled_shutdown":      "Scheduled %s at %s",
		"executing_operation":     "Executing %s operation...",
		"operation_successful":    "%s operation successful",
		"operation_failed":        "%s command failed: %v",
		"hibernate_failed":        "Hibernate failed, trying shutdown...",
		"unknown_command":         "Unknown command. Use 'help' to see available commands",
		"please_specify_language": "Please specify language code (en or zh-Hans)",
		"invalid_time_type":       "Invalid time type. Please use 'start' or 'end'",
		"system_status":           "System status: Running\nCurrent time: %s\nTime range: %02d:%02d - %02d:%02d\nOperation mode: %s",
		
		// Command responses
		"enter_command":           "Please enter a command",
		"enter_start_time":        "Please enter start time (format HH:MM), e.g. 22:00",
		"enter_end_time":          "Please enter end time (format HH:MM), e.g. 06:00",
		"invalid_time_format":     "Invalid time format. Please use HH:MM format.",
		"time_set_success":        "%s time set to %02d:%02d",
		"invalid_mode":            "Invalid mode. Available modes: shutdown, hibernate, reboot, logoff",
		"mode_set_success":        "Operation mode set to: %s",
		
		// Menu
		"welcome_title":           "===== AutoShutdown Remote Control =====",
		"menu_item":               "%d. %s",
		"menu_shutdown":           "Shutdown computer",
		"menu_hibernate":          "Hibernate computer",
		"menu_reboot":             "Restart computer",
		"menu_logoff":             "Log off current user",
		"menu_status":             "View system status",
		"menu_set_start_time":     "Set start time",
		"menu_set_end_time":       "Set end time",
		"menu_set_mode":           "Set operation mode",
		"menu_language":           "Change language",
		"menu_help":               "Show help",
		"menu_exit":               "Exit",
		"menu_prompt":             "Enter option number: ",
		
		// Help text
		"help_text": `Available commands:
- shutdown: Shutdown computer
- hibernate: Hibernate computer
- reboot: Restart computer
- logoff: Log off current user
- setmode [mode]: Set operation mode (shutdown/hibernate/reboot/logoff)
- status: View system status
- settime start HH:MM: Set start time
- settime end HH:MM: Set end time
- language [code]: Change language (en/zh-Hans)
- version: Show version information
- help: Show help information`,

		// Language
		"language_changed":        "Language changed to: %s",
		"language_name_en":        "English",
		"language_name_zh-Hans":   "Simplified Chinese",
		"invalid_language":        "Invalid language code. Available languages: en, zh-Hans",
		
		// Log messages
		"log_tcp_server_started":  "TCP remote control server started, listening on port %s",
		"log_udp_server_started":  "UDP remote control server started, listening on port %s",
		"log_accept_failed":       "Failed to accept connection: %v",
		"log_new_tcp_connection":  "New TCP connection from: %s",
		"log_command_read_failed": "Failed to read command: %v",
		"log_udp_addr_failed":     "Failed to resolve UDP address: %v",
		"log_udp_listen_failed":   "Failed to start UDP server: %v",
		"log_udp_read_failed":     "Failed to read UDP data: %v",
		"log_udp_command":         "Received UDP command from %s: %s",
		"log_service_installed":   "Service installed successfully",
		"log_service_removed":     "Service removed successfully",
		"log_service_stopped":     "Service stopped successfully",
		"log_service_started":     "Service started successfully",
		"log_hibernate_failed":    "Hibernate command failed: %v",
	},
	"zh-Hans": {
		// 通用
		"app_name":                "自动关机",
		"developed_by":            "由 Hu Gang 开发",
		"version_info":            "%s 版本: %s (%s)",
		
		// 操作模式
		"mode_shutdown":           "关机",
		"mode_hibernate":          "休眠",
		"mode_reboot":             "重启",
		"mode_logoff":             "注销",
		
		// 状态消息
		"current_status":          "当前时间范围: %02d:%02d - %02d:%02d | 操作模式: %s | 版本: %s",
		"scheduled_shutdown":      "计划在 %s %s",
		"executing_operation":     "正在执行%s操作...",
		"operation_successful":    "%s操作成功",
		"operation_failed":        "%s命令失败: %v",
		"hibernate_failed":        "休眠失败，尝试关机...",
		"unknown_command":         "未知命令，请使用 'help' 查看可用命令",
		"please_specify_language": "请指定语言代码（en 或 zh-Hans）",
		"invalid_time_type":       "时间类型无效。请使用 'start' 或 'end'",
		"system_status":           "系统状态: 正常运行\n当前时间: %s\n时间范围: %02d:%02d - %02d:%02d\n当前操作模式: %s",
		
		// 命令响应
		"enter_command":           "请输入命令",
		"enter_start_time":        "请输入开始时间（格式为 HH:MM），例如 22:00",
		"enter_end_time":          "请输入结束时间（格式为 HH:MM），例如 06:00",
		"invalid_time_format":     "时间格式无效。请使用 HH:MM 格式。",
		"time_set_success":        "%s时间设置为 %02d:%02d",
		"invalid_mode":            "无效的模式。可用模式: shutdown(关机), hibernate(休眠), reboot(重启), logoff(注销)",
		"mode_set_success":        "操作模式设置为: %s",
		
		// 菜单
		"welcome_title":           "===== 自动关机远程控制 =====",
		"menu_item":               "%d. %s",
		"menu_shutdown":           "关闭计算机",
		"menu_hibernate":          "休眠计算机",
		"menu_reboot":             "重启计算机",
		"menu_logoff":             "注销当前用户",
		"menu_status":             "查看系统状态",
		"menu_set_start_time":     "设置开始时间",
		"menu_set_end_time":       "设置结束时间",
		"menu_set_mode":           "设置操作模式",
		"menu_language":           "更改语言",
		"menu_help":               "显示帮助",
		"menu_exit":               "退出",
		"menu_prompt":             "请输入选项编号: ",
		
		// 帮助文本
		"help_text": `可用命令:
- shutdown: 关闭计算机
- hibernate: 休眠计算机
- reboot: 重启计算机
- logoff: 注销当前用户
- setmode [mode]: 设置操作模式 (shutdown/hibernate/reboot/logoff)
- status: 查看系统状态
- settime start HH:MM: 设置开始时间
- settime end HH:MM: 设置结束时间
- language [code]: 更改语言 (en/zh-Hans)
- version: 显示版本信息
- help: 显示帮助信息`,

		// 语言
		"language_changed":        "语言已更改为: %s",
		"language_name_en":        "英文",
		"language_name_zh-Hans":   "简体中文",
		"invalid_language":        "无效的语言代码。可用语言: en, zh-Hans",
		
		// 日志消息
		"log_tcp_server_started":  "TCP远程控制服务器已启动，监听端口 %s",
		"log_udp_server_started":  "UDP远程控制服务器已启动，监听端口 %s",
		"log_accept_failed":       "接受连接失败: %v",
		"log_new_tcp_connection":  "新的TCP连接来自: %s",
		"log_command_read_failed": "读取命令失败: %v",
		"log_udp_addr_failed":     "解析UDP地址失败: %v",
		"log_udp_listen_failed":   "UDP服务器启动失败: %v",
		"log_udp_read_failed":     "读取UDP数据失败: %v",
		"log_udp_command":         "收到来自 %s 的UDP命令: %s",
		"log_service_installed":   "服务安装成功",
		"log_service_removed":     "服务卸载成功",
		"log_service_stopped":     "服务停止成功",
		"log_service_started":     "服务启动成功",
		"log_hibernate_failed":    "休眠命令失败: %v",
	},
}

// Current language
var currentLang = "en"

// Get localized string
func T(key string, args ...interface{}) string {
	if str, ok := langStrings[currentLang][key]; ok {
		if len(args) > 0 {
			return fmt.Sprintf(str, args...)
		}
		return str
	}
	
	// Fallback to English
	if str, ok := langStrings["en"][key]; ok {
		if len(args) > 0 {
			return fmt.Sprintf(str, args...)
		}
		return str
	}
	
	// Return key if not found
	return key
}

// Change language
func SetLanguage(lang string) bool {
	if _, ok := langStrings[lang]; ok {
		currentLang = lang
		return true
	}
	return false
}

// Get language name
func GetLanguageName(lang string) string {
	key := "language_name_" + lang
	return T(key)
}
