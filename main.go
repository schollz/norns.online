package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/schollz/logger"
	"github.com/schollz/norns.online/src/norns"
	"github.com/schollz/norns.online/src/server"
	"github.com/shirou/gopsutil/v3/process"
)

var config = flag.String("config", "", "config file to use")
var debugMode = flag.Bool("debug", false, "debug mode")
var relayMode = flag.Bool("relay", false, "run relay")
var serverMode = flag.Bool("server", false, "run server")
var nornsOnlineHost = flag.String("host", "https://norns.online", "host to connect to")
var matronHost = flag.String("matron", "localhost:5555", "matron host to connect to")
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

	if *relayMode && !*serverMode {
		err = server.Run()
	} else if *serverMode {
		logger.Debug("serverMode")
		if *relayMode {
			go func() {
				server.Run()
			}()
			time.Sleep(1 * time.Second)
		}
		n, err := norns.New(*config, pid, *nornsOnlineHost, *matronHost)
		if err == nil {
			logger.Debug("running norns")
			err = n.Run()
		}
	}
	if err != nil {
		logger.Error(err)
	}
}
