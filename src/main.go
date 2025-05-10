package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	. "github.com/CodyGuo/win"
	"github.com/kardianos/service"
)

var (
	arg string
)

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	// 代码写在这儿
	doIt()
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func init() {
	flag.StringVar(&arg, "uFlags", "shutdown", "shutdown logoff reboot")
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
	for true {
		now := time.Now()
		hour := now.Hour()
		//minute := now.Minute()
		//if (hour == 22 && minute >= 20) || hour >= 23 {
		if hour%3 == 0 {
			shutdown()
		}
		time.Sleep(10 * time.Second)
	}
}

func shutdown() {
	getPrivileges()
	ExitWindowsEx(EWX_SHUTDOWN, 0)
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
