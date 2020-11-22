package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/schollz/logger"
)

var sockets map[string]map[*websocket.Conn]string
var mutex sync.Mutex

func main() {
	log.SetLevel("debug")
	sockets = make(map[string]map[*websocket.Conn]string)
	port := 8098
	log.Infof("listening on :%d", port)
	http.HandleFunc("/", handler)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	t := time.Now().UTC()
	err := handle(w, r)
	if err != nil {
		log.Error(err)
	}
	log.Infof("%v %v %v %s\n", r.RemoteAddr, r.Method, r.URL.Path, time.Since(t))
}

func handle(w http.ResponseWriter, r *http.Request) (err error) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	// very special paths
	if r.URL.Path == "/ws" {
		err = handleWebsocket(w, r)
		log.Infof("ws: %w", err)
		if err != nil {
			err = nil
		}
	} else {
		if strings.HasPrefix(r.URL.Path, "/js/") || strings.HasPrefix(r.URL.Path, "/img/") {
			http.FileServer(http.Dir(".")).ServeHTTP(w, r)
		} else {
			http.ServeFile(w, r, "index.html")
		}
	}

	return
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
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
	MP3    string `json:"mp3"`
}

func handleWebsocket(w http.ResponseWriter, r *http.Request) (err error) {
	c, errUpgrade := wsupgrader.Upgrade(w, r, nil)
	if errUpgrade != nil {
		return errUpgrade
	}
	defer c.Close()
	var m Message
	err = c.ReadJSON(&m)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("initial m: %+v", m)
	group := m.Group
	if group == "" {
		return
	}
	name := m.Name
	if name == "" {
		name = RandString(5)
	}
	mutex.Lock()
	if _, ok := sockets[group]; !ok {
		sockets[group] = make(map[*websocket.Conn]string)
	}
	sockets[group][c] = name
	log.Debugf("%d connected in %s", len(sockets[group]), group)
	mutex.Unlock()

	defer func() {
		mutex.Lock()
		delete(sockets[group], c)
		if len(sockets[group]) == 0 {
			delete(sockets, group)
		}
		log.Debugf("%d connected in %s", len(sockets[group]), group)
		mutex.Unlock()
	}()

	for {
		m = Message{}
		err = c.ReadJSON(&m)
		if err != nil {
			break
		}
		// send out to others
		mutex.Lock()
		for c2, k := range sockets[group] {
			if k == sockets[group][c] {
				// never send to self
				continue
			}
			if k != m.Recipient && m.Recipient != "" {
				// skip unless its recipient
				continue
			}
			go func(c2 *websocket.Conn, m Message) {
				err := c2.WriteJSON(m)
				if err != nil {
					mutex.Lock()
					delete(sockets[group], c2)
					mutex.Unlock()
				}
			}(c2, m)
		}
		mutex.Unlock()
	}
	return
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandString(n int) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
