package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gorilla/websocket"
	"github.com/schollz/logger"
	"github.com/shirou/gopsutil/v3/process"
	wav "github.com/youpy/go-wav"
)

var RELAY_ADDRESS = "http://duct.schollz.com/norns.online."

var config = flag.String("config", "", "config file to use")
var debugMode = flag.Bool("debug", false, "debug mode")

func main() {
	filename, _ := FindChangingFile("/home/we/dust/audio/tape")
	SpliceEndOfWavFile(filename, "/tmp/test.wav",12)
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
	ioutil.WriteFile("/tmp/norns.online.kill", []byte(`#!/bin/bash
kill -9 `+fmt.Sprint(pid)+`
rm -- "$0"
`), 0777)

	fmt.Printf("%d\n", pid)

	// setup logger
	flag.Parse()
	logger.SetLevel("error")
	if *debugMode {
		logger.SetLevel("debug")
	}

	if *config == "" {
		logger.Error("need config, use --config")
		os.Exit(1)
	}
	n, err := New(*config)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	err = n.Run()
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}

type NornsOnline struct {
	Name        string `json:"name"`
	AllowMenu   bool   `json:"allowmenu"`
	AllowEncs   bool   `json:"allowencs"`
	AllowKeys   bool   `json:"allowkeys"`
	AllowTwitch bool   `json:"allowtwitch"`
	KeepAwake   bool   `json:"keepawake"`
	FrameRate   int    `json:"framerate"`

	configFile     string
	configFileHash []byte
	active         bool
	inMenu         bool
	norns          *websocket.Conn
	ws             *websocket.Conn

	sync.Mutex
}

// New returns a new instance
func New(configFile string) (n *NornsOnline, err error) {
	n = new(NornsOnline)
	n.configFile = configFile
	n.AllowEncs = true
	n.AllowKeys = true
	n.KeepAwake = false
	n.FrameRate = 4
	_, err = n.Load()
	go n.connectToWebsockets()
	return
}

// Load will update the configuration if config file changes
func (n *NornsOnline) Load() (updated bool, err error) {
	currentHash, err := MD5HashFile(n.configFile)
	if err != nil {
		return
	}
	if bytes.Equal(n.configFileHash, currentHash) {
		return
	}
	b, err := ioutil.ReadFile(n.configFile)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &n)
	if err != nil {
		return
	}
	n.configFileHash = currentHash
	logger.Debugf("loaded: %+v", n)
	updated = true
	return
}

func (n *NornsOnline) connectToWebsockets() (err error) {
	for {
		if n.ws != nil {
			n.ws.Close()
			time.Sleep(500 * time.Millisecond)
		}
		//wsURL := url.URL{Scheme: "ws", Host: "192.168.0.3:8098", Path: "/ws"}
		wsURL := url.URL{Scheme: "wss", Host: "norns.online", Path: "/ws"}
		logger.Debugf("connecting to %s as %s", wsURL, n.Name)
		n.ws, _, err = websocket.DefaultDialer.Dial(wsURL.String(), nil)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		n.ws.WriteJSON(Message{
			Group: n.Name,
			Name:  "norns", // specify that i am the norns
		})
		pings := 0
		for {
			var m Message
			err = n.ws.ReadJSON(&m)
			if err != nil {
				logger.Debug(err)
				return
			}
			logger.Debugf("got message: %+v", m)

			cmd, err := n.processMessage(m)
			if err != nil {
				continue
			}
			n.Lock()
			logger.Debugf("running command: '%s'", cmd)
			err = n.norns.WriteMessage(websocket.TextMessage, []byte(cmd+"\n"))
			if err != nil {
				logger.Error(err)
				continue
			}
			pings++
			if pings%20 == 0 && n.KeepAwake {
				err = n.norns.WriteMessage(websocket.TextMessage, []byte(`screen.ping()`+"\n"))
				if err != nil {
					logger.Error(err)
					continue
				}
			}
			n.Unlock()

		}
	}
	return
}

// Run forever
func (n *NornsOnline) Run() (err error) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// bind to internal address
	u := url.URL{Scheme: "ws", Host: "localhost:5555", Path: "/"}
	logger.Infof("connecting to %s", u.String())
	var cstDialer = websocket.Dialer{
		Subprotocols:     []string{"bus.sp.nanomsg.org"},
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 3 * time.Second,
	}

	n.norns, _, err = cstDialer.Dial(u.String(), nil)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	defer n.norns.Close()

	done := make(chan struct{})

	logger.Info("connected")
	ticker := time.NewTicker(1 * time.Second)
	if n.FrameRate > 1 {
		ticker = time.NewTicker(time.Duration(1000/n.FrameRate) * time.Millisecond)
	}
	logger.Debugf("ticker: %+v", ticker)
	defer ticker.Stop()
	ticker2 := time.NewTicker(1000 * time.Millisecond)
	defer ticker2.Stop()
	for {
		select {
		case <-done:
			return
		case _ = <-ticker2.C:
			currentName := n.Name
			updated, _ := n.Load()
			if updated {
				ticker.Stop()
				if n.FrameRate == 1 {
					ticker = time.NewTicker(1 * time.Second)
				} else {
					ticker = time.NewTicker(time.Duration(1000/n.FrameRate) * time.Millisecond)
				}
				if n.Name != currentName {
					// restart websockets with new name
					n.ws.Close()
					go n.connectToWebsockets()
				}
			}
		case _ = <-ticker.C:
			n.Lock()
			err = n.norns.WriteMessage(websocket.TextMessage, []byte(`_norns.screen_export_png("/dev/shm/screenshot.png")`+"\n"))
			if err != nil {
				logger.Debugf("write: %w", err)
				return
			}
			n.Unlock()
			time.Sleep(10 * time.Millisecond)

			go n.updateClient()
		case <-interrupt:
			logger.Info("interrupt - quitting gracefully")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err = n.norns.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				logger.Info("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
	return
}

func (n *NornsOnline) updateClient() (err error) {
	// open dumped image
	src, err := imaging.Open("/dev/shm/screenshot.png")
	if err != nil {
		return
	}

	// Resize the cropped image to width = 200px preserving the aspect ratio.
	src = imaging.Resize(src, 550, 0, imaging.NearestNeighbor)
	src = imaging.AdjustGamma(src, 1.25)
	err = imaging.Save(src, "/dev/shm/screenshot2.png")
	if err != nil {
		return
	}

	b, err := ioutil.ReadFile("/dev/shm/screenshot2.png")
	if err != nil {
		return
	}
	base64data := base64.StdEncoding.EncodeToString(b)

	tsent := time.Now()
	if n.ws != nil {
		n.ws.WriteJSON(Message{
			Img:    base64data,
			Twitch: n.AllowTwitch,
		})
	}
	logger.Tracef("sent data in %s", time.Since(tsent))
	return
}

type Message struct {
	Name      string `json:"name,omitempty"`
	Group     string `json:"group,omitempty"`
	Recipient string `json:"recipient,omitempty"`

	Img    string `json:"img,omitempty"`
	Kind   string `json:"kind,omitempty"`
	N      int    `json:"n"`
	Z      int    `json:"z"`
	Fast   bool   `json:"fast,omitempty"`
	Twitch bool   `json:"twitch"`
}

// processMessage only lets certain k inds of messages through
func (n *NornsOnline) processMessage(m Message) (cmd string, err error) {
	if m.Kind == "enc" {
		if n.AllowEncs {
			cmd = fmt.Sprintf("enc(%d,%d)", sanitizeIndex(m.N), sanitizeEnc(m.Z))
		} else {
			logger.Debug("encs disabled")
		}
	} else if m.Kind == "key" {
		if n.AllowKeys {
			cmd = fmt.Sprintf("key(%d,%d)", sanitizeIndex(m.N), sanitizeKey(m.Z))
			if m.Fast && m.N == 1 && n.AllowMenu {
				n.inMenu = !n.inMenu
				if n.inMenu {
					cmd = "set_mode(true)"
				} else {
					cmd = "_menu.set_mode(false)"
				}
			}
		} else {
			logger.Debug("keys disabled")
		}
	}
	if n.inMenu {
		cmd = "_menu." + cmd
	}

	return
}

func sanitizeIndex(v int) int {
	if v < 1 {
		return 1
	} else if v > 3 {
		return 3
	}
	return v
}

func sanitizeEnc(v int) int {
	if v < -3 {
		return -3
	} else if v > 3 {
		return 3
	}
	return v
}
func sanitizeKey(v int) int {
	if v < 0 {
		return 0
	} else if v > 1 {
		return 1
	}
	return v
}

// MD5HashFile returns MD5 hash
func MD5HashFile(fname string) (hash256 []byte, err error) {
	f, err := os.Open(fname)
	if err != nil {
		return
	}
	defer f.Close()

	h := md5.New()
	if _, err = io.Copy(h, f); err != nil {
		return
	}

	hash256 = h.Sum(nil)
	return
}

// FindChangingFile returns the name of the file that's changing
// (the one that's being recorded)
func FindChangingFile(folder string) (filename string, err error) {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return
	}
	if len(files) == 0 {
		return
	}
	sort.Slice(files[:], func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})
	filename = files[0].Name()
	return
}

// SpliceEndOfWavFile will splice out the last part of a file
func SpliceEndOfWavFile(filename string, outfilename string, minFromEnd float64) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	reader := wav.NewReader(f)
	fm, err := reader.Format()
	if err != nil {
		return
	}
	fmt.Printf("fm: %+v", fm)
	start := uint32(48000 * 1)
	end := uint32(start + 48000*3)
	_, err = reader.ReadSamples(start)
	if err != nil {
		return
	}
	fmt.Println(end)

	outfile, err := os.Create(outfilename)
	if err != nil {
		return
	}
	defer outfile.Close()

	writer := wav.NewWriter(outfile, end-start, fm.NumChannels, fm.SampleRate, fm.BitsPerSample)
	samples, err := reader.ReadSamples(end - start)
	if err != nil {
		return
	}
	err = writer.WriteSamples(samples)
	return
}
