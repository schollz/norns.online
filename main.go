package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/schollz/logger"
	"github.com/shirou/gopsutil/v3/process"
	"norns.online/src/norns"
	"norns.online/src/server"
)

var RELAY_ADDRESS = "http://duct.schollz.com/norns.online."

var config = flag.String("config", "", "config file to use")
var debugMode = flag.Bool("debug", false, "debug mode")
var relayMode = flag.Bool("relay", false, "run relay")

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
	if numRunning > 1 {
		fmt.Println("already running")
		os.Exit(1)
	}

	fmt.Printf("%d\n", pid)

	// setup logger
	flag.Parse()
	logger.SetLevel("info")
	if *debugMode {
		logger.SetLevel("debug")
	}

	if *relayMode {
		err = runRelay()
	} else {
		err = runNornsServer(pid)
	}
	if err != nil {
		logger.Error(err)
	}
}

func runRelay() (err error) {
	err = server.Run()
	return
}

func runNornsServer(pid int32) (err error) {
	ioutil.WriteFile("/tmp/norns.online.kill", []byte(`#!/bin/bash
kill -9 `+fmt.Sprint(pid)+`
pkill jack_capture
rm -rf /dev/shm/jack*.flac
rm -- "$0"
`), 0777)
	ioutil.WriteFile("/dev/shm/jack_capture.sh", []byte(`#!/bin/bash
cd /dev/shm
rm -rf /dev/shm/*.wav
rm -rf /dev/shm/*.flac
chmod +x /home/we/dust/code/norns.online/jack_capture
/home/we/dust/code/norns.online/jack_capture -f flac --port system:playback_1 --port system:playback_2 --recording-time 36000 -Rf 96000 -z 4
`), 0777)
	if *config == "" {
		logger.Error("need config, use --config")
		os.Exit(1)
	}
	n, err := norns.New(*config)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	err = n.Run()
	return
}
