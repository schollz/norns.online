package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/schollz/logger"
	"github.com/schollz/norns.online/src/norns"
	"github.com/schollz/norns.online/src/server"
	"github.com/shirou/gopsutil/v3/process"
)

var config = flag.String("config", "", "config file to use")
var debugMode = flag.Bool("debug", false, "debug mode")
var relayMode = flag.Bool("relay", false, "run relay")
var nornsOnlineHost = flag.String("host", "https://norns.online", "host to connect to")
var forceRun = flag.Bool("force", false, "force running")

func main() {
	// logger.SetOutput(&lumberjack.Logger{
	// 	Filename:   "/dev/shm/norns.online.log",
	// 	MaxSize:    1, // megabytes
	// 	MaxBackups: 3,
	// 	MaxAge:     28,    //days
	// 	Compress:   false, // disabled by default
	// })

	// first make sure its not already running an instance
	processes, err := process.Processes()
	if err != nil {
		panic(err)
	}
	numRunning := 0
	pid := int32(0)
	for _, process := range processes {
		name, _ := process.Name()
		if name == "norns.online" {
			numRunning++
			if pid == 0 {
				pid = process.Pid
			}
		}
	}

	fmt.Printf("%d\n", pid)

	// setup logger
	flag.Parse()
	logger.SetLevel("info")
	if *debugMode {
		logger.SetLevel("debug")
	}

	if *forceRun == false {
		if numRunning > 1 {
			fmt.Println("already running")
			os.Exit(1)
		}
	}

	if *relayMode {
		err = server.Run()
	} else {
		n, err := norns.New(*config, pid, *nornsOnlineHost)
		if err == nil {
			err = n.Run()
		}
	}
	if err != nil {
		logger.Error(err)
	}
}
