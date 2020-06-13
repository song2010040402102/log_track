package main

import (
	"config"
	"github.com/astaxie/beego/logs"
	"os"
	"os/signal"
	"plat"
	"stat"
	"syscall"
)

func initLog() {
	logs.SetLogFuncCallDepth(3)
	logs.EnableFuncCallDepth(true)
	logs.SetLevel(logs.LevelDebug)
}

func startServer() bool {
	argc := len(os.Args)
	if argc < 2 {
		return false
	}
	switch os.Args[1] {
	case "stat":
		stat.StartServer()
	case "plat":
		plat.StartServer()
	default:
		return false
	}
	return true
}

func closeServer() {
	switch os.Args[1] {
	case "stat":
		stat.CloseServer()
	case "plat":
		plat.CloseServer()
	}
}

func main() {
	initLog()
	config.Init()
	logs.Notice("log_track start...")
	if !startServer() {
		return
	}
	var sig os.Signal
	c := make(chan os.Signal, 1)
	for {
		signal.Notify(c)
		sig = <-c
		if sig != syscall.SIGPIPE {
			break
		}
	}
	closeServer()
	logs.Notice("log_track terminate with sig:", sig)
}
