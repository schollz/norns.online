package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/schollz/logger"
)

var addr = flag.String("addr", "192.168.0.82:5555", "http service address")
var mu sync.Mutex

type State struct {
	Input Input
	Menu  bool
}

type Input struct {
	Encs []Enc `json:"encs"`
	Keys []Key `json:"keys"`
}
type Enc struct {
	N int `json:"n"`
	Z int `json:"z"`
}
type Key struct {
	N int  `json:"n"`
	Z int  `json:"z"`
	F bool `json:"f"`
}

var state State

func main() {
	logger.SetLevel("debug")
	flag.Parse()
	log.SetFlags(0)

	state = State{Input: Input{}}
	for i := 0; i < 3; i++ {
		state.Input.Keys[i] = Key{}
		state.Input.Encs[i] = Enc{}
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	log.Printf("connecting to %s", u.String())
	var cstDialer = websocket.Dialer{
		Subprotocols:     []string{"bus.sp.nanomsg.org"},
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 3 * time.Second,
	}

	c, _, err := cstDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				logger.Error("read:", err)
				continue
			}
			logger.Tracef("recv: '%s'", message)
		}
	}()

	go func() {
		for {
			response, err := func() (response string, err error) {
				resp, err := http.Get("http://duct.schollz.com/b")
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
			cmd, err := processMessage(response)
			if err != nil {
				logger.Error(err)
				continue
			}
			if cmd == "" {
				continue
			}
			mu.Lock()
			err = c.WriteMessage(websocket.TextMessage, []byte(`_norns.system_cmd_lua("`+cmd+`")`+"\n"))
			mu.Unlock()
			if err != nil {
				logger.Error(err)
				continue
			}
		}
	}()

	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	fmt.Println("ceonnected")
	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			logger.Tracef("writing message at %s\n", t)
			// err := c.WriteMessage(websocket.TextMessage, []byte(`_norns.system_cmd_lua("enc(2,1)")`+"\n"))
			mu.Lock()
			err := c.WriteMessage(websocket.TextMessage, []byte(`_norns.screen_export_png("/tmp/screenshot.png")`+"\n"))
			mu.Unlock()
			if err != nil {
				log.Println("write:", err)
				return
			}
			cmd := exec.Command("convert", strings.Fields(`/tmp/screenshot.png -gamma 1.25 -filter point -resize 400% -gravity center -background black -extent 120% /tmp/screenshot2.png`)...)
			_, err = cmd.Output()
			if err != nil {
				logger.Error(err)
				return
			}
			postImage()
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
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
}

func postImage() (err error) {
	f, err := os.Open("/tmp/screenshot2.png")
	if err != nil {
		return
	}
	defer f.Close()
	req, err := http.NewRequest("POST", "https://duct.schollz.com/a.png?pubsub=true", f)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	return
}

func processMessage(s string) (cmd string, err error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		logger.Error(err)
		return
	}
	var m Input
	err = json.Unmarshal(b, &m)
	if err != nil {
		logger.Error(err)
		return
	}

	for i, k := range m.Keys {
		k.Z = sanitizeKey(k.Z)
		if k.F != state.Input.Keys[i].F && i == 0 && k.F {
			// TODO toggle menu
		}
		state.Input.Keys[i].F = k.F
		if k.Z != state.Input.Keys[i].Z {
			state.Input.Keys[i].Z = k.Z
			cmd += fmt.Sprintf("key(%d,%d) ", i, k.Z)
		}
	}
	for i, k := range m.Encs {
		k.Z = sanitizeEnc(k.Z)
		if k.Z != state.Input.Encs[i].Z {
			state.Input.Encs[i].Z = k.Z
			cmd += fmt.Sprintf("enc(%d,%d) ", i, k.Z)
		}
	}
	return
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
