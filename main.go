package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gorilla/websocket"
	"github.com/schollz/logger"
)

var addr = flag.String("addr", "192.168.0.82:5555", "http service address")
var mu sync.Mutex
var inMenu bool

func main() {
	logger.SetLevel("debug")
	flag.Parse()
	log.SetFlags(0)

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

	// go func() {
	// 	defer close(done)
	// 	for {
	// 		_, message, err := c.ReadMessage()
	// 		if err != nil {
	// 			logger.Error("read:", err)
	// 			continue
	// 		}
	// 		logger.Tracef("recv: '%s'", message)
	// 	}
	// }()

	go func() {
		client := &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 1,
			},
			Timeout: 30000 * time.Second,
		}
		for {
			response, err := func() (response string, err error) {
				resp, err := client.Get("http://duct.schollz.com/b")
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
			logger.Debugf("running command: '%s'", cmd)
			err = c.WriteMessage(websocket.TextMessage, []byte(`_norns.system_cmd_lua("`+cmd+`")`+"\n"))
			mu.Unlock()
			if err != nil {
				logger.Error(err)
				continue
			}
		}
	}()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	fmt.Println("ceonnected")
	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			logger.Tracef("writing message at %s\n", t)
			mu.Lock()
			err := c.WriteMessage(websocket.TextMessage, []byte(`_norns.screen_export_png("/tmp/screenshot.png")`+"\n"))
			mu.Unlock()
			if err != nil {
				log.Println("write:", err)
				return
			}
			time.Sleep(10 * time.Millisecond)

			err = postImage()
			if err != nil {
				logger.Errorf("image: %+w", err)
				continue
			}
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
	src, err := imaging.Open("/tmp/screenshot.png")
	if err != nil {
		logger.Error("failed to open image: %v", err)
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
	req, err := http.NewRequest("POST", "https://duct.schollz.com/a.png?pubsub=true", bytes.NewBufferString(base64data))
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

type Message struct {
	Kind string
	N    int
	Z    int
	Fast bool
}

func processMessage(s string) (cmd string, err error) {
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
		cmd = fmt.Sprintf("enc(%d,%d)", sanitizeIndex(m.N), sanitizeEnc(m.Z))
	} else if m.Kind == "key" {
		cmd = fmt.Sprintf("key(%d,%d)", sanitizeIndex(m.N), sanitizeKey(m.Z))
		if m.Fast && m.N == 1 {
			inMenu = !inMenu
			if inMenu {
				cmd = "set_mode(true)"
			} else {
				cmd = "_menu.set_mode(false)"
			}
		}
	}
	if inMenu {
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
	if inMenu {
		if v < -1 {
			return -1
		} else if v > 1 {
			return 1
		}
	}
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
