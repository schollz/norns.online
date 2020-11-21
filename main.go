package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gorilla/websocket"
	"github.com/schollz/logger"
	"github.com/shirou/gopsutil/v3/process"
	"gopkg.in/natefinch/lumberjack.v2"
)

var RELAY_ADDRESS = "http://duct.schollz.com/norns.online."

var config = flag.String("config", "", "config file to use")

func main() {
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
	logger.SetOutput(&lumberjack.Logger{
		Filename:   "/tmp/norns.online.log",
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28,    //days
		Compress:   false, // disabled by default
	})
	logger.SetLevel("debug")
	flag.Parse()

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
	client         *http.Client
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

	// setup dialer for fast DNS resolution
	var (
		dnsResolverIP        = "1.1.1.1:53" // Google DNS resolver.
		dnsResolverProto     = "udp"        // Protocol to use for the DNS resolver
		dnsResolverTimeoutMs = 5000         // Timeout (ms) for the DNS resolver (optional)
	)

	dialer := &net.Dialer{
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Duration(dnsResolverTimeoutMs) * time.Millisecond,
				}
				return d.DialContext(ctx, dnsResolverProto, dnsResolverIP)
			},
		},
	}

	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, addr)
	}

	http.DefaultTransport.(*http.Transport).DialContext = dialContext
	n.client = &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialContext,
			MaxIdleConnsPerHost:   3,
			MaxIdleConns:          100,
			IdleConnTimeout:       9000 * time.Second,
			TLSHandshakeTimeout:   1000 * time.Second,
			ExpectContinueTimeout: 3000 * time.Second,
		},
		Timeout: 30000 * time.Second,
	}
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

// Run forever
func (n *NornsOnline) Run() (err error) {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: "192.168.0.82:5555", Path: "/"}
	logger.Infof("connecting to %s", u.String())
	var cstDialer = websocket.Dialer{
		Subprotocols:     []string{"bus.sp.nanomsg.org"},
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 3 * time.Second,
	}

	c, _, err := cstDialer.Dial(u.String(), nil)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		pings := 0
		for {
			response, err := func() (response string, err error) {
				resp, err := n.client.Get(RELAY_ADDRESS + n.Name)
				if err != nil {
					return
				}
				defer resp.Body.Close()
				data, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return
				}
				response = string(data)
				return
			}()
			if err != nil {
				logger.Error(err)
				continue
			}
			logger.Debugf("got command: '%s'", response)
			cmd, err := n.processMessage(response)
			if err != nil {
				logger.Error(err)
				continue
			}
			if cmd == "" {
				continue
			}
			n.Lock()
			logger.Debugf("running command: '%s'", cmd)
			err = c.WriteMessage(websocket.TextMessage, []byte(cmd+"\n"))
			if err != nil {
				logger.Error(err)
				continue
			}
			pings++
			if pings%20 == 0 && n.KeepAwake {
				err = c.WriteMessage(websocket.TextMessage, []byte(`screen.ping()`+"\n"))
				if err != nil {
					logger.Error(err)
					continue
				}
			}
			n.Unlock()
		}
	}()

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
			updated, _ := n.Load()
			if updated {
				ticker.Stop()
				if n.FrameRate == 1 {
					ticker = time.NewTicker(1 * time.Second)
				} else {
					ticker = time.NewTicker(time.Duration(1000/n.FrameRate) * time.Millisecond)
				}
			}
		case _ = <-ticker.C:
			n.Lock()
			err = c.WriteMessage(websocket.TextMessage, []byte(`_norns.screen_export_png("/tmp/screenshot.png")`+"\n"))
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
			err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
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

type Payload struct {
	Img    string `json:"img"`
	Twitch bool   `json:"twitch"`
}

func (n *NornsOnline) updateClient() (err error) {
	// open dumped image
	src, err := imaging.Open("/tmp/screenshot.png")
	if err != nil {
		return
	}

	// Resize the cropped image to width = 200px preserving the aspect ratio.
	src = imaging.Resize(src, 550, 0, imaging.NearestNeighbor)
	src = imaging.AdjustGamma(src, 1.25)
	err = imaging.Save(src, "/tmp/screenshot2.png")
	if err != nil {
		return
	}

	b, err := ioutil.ReadFile("/tmp/screenshot2.png")
	if err != nil {
		return
	}
	base64data := base64.StdEncoding.EncodeToString(b)

	payload := Payload{
		Img:    base64data,
		Twitch: n.AllowTwitch,
	}
	payloadbytes, err := json.Marshal(payload)
	if err != nil {
		return
	}
	payloadbase64 := base64.StdEncoding.EncodeToString(payloadbytes)
	tsent := time.Now()
	req, err := http.NewRequest("POST", RELAY_ADDRESS+n.Name+".png?pubsub=true", bytes.NewBufferString(payloadbase64))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := n.client.Do(req)
	if err != nil {
		return
	}
	logger.Debugf("sent data in %s", time.Since(tsent))
	defer resp.Body.Close()
	return
}

type Message struct {
	Kind string
	N    int
	Z    int
	Fast bool
}

// processMessage only lets certain k inds of messages through
func (n *NornsOnline) processMessage(s string) (cmd string, err error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		logger.Error(err)
		return
	}
	var m Message
	err = json.Unmarshal(b, &m)
	if err != nil {
		logger.Error(err)
		return
	}

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
