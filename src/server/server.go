package server

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/schollz/logger"
	"norns.online/src/models"
)

var sockets map[string]Client
var mutex sync.Mutex
var wsmutex sync.Mutex

type Client struct {
	Group string
	Room  string
	conn  *websocket.Conn
}

func Run() (err error) {
	sockets = make(map[string]Client)
	port := 8098
	log.Infof("listening on :%d", port)
	http.HandleFunc("/", handler)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
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
		if strings.HasPrefix(r.URL.Path, "/static/") {
			http.FileServer(http.Dir(".")).ServeHTTP(w, r)
		} else {
			http.ServeFile(w, r, "static/index.html")
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

func handleWebsocket(w http.ResponseWriter, r *http.Request) (err error) {
	c, errUpgrade := wsupgrader.Upgrade(w, r, nil)
	if errUpgrade != nil {
		return errUpgrade
	}
	defer c.Close()
	var m models.Message
	err = c.ReadJSON(&m)
	if err != nil {
		log.Error(err)
		return
	}

	log.Debugf("initial m: %+v", m)
	name := m.Name
	if name == "" || (m.Group == "" && m.Room == "") {
		return
	}

	mutex.Lock()
	sockets[name] = Client{
		Group: m.Group,
		Room:  m.Room,
		conn:  c,
	}
	log.Debugf("have %d sockets", len(sockets))
	mutex.Unlock()

	defer func() {
		mutex.Lock()
		delete(sockets, name)
		mutex.Unlock()

	}()

	for {
		m = models.Message{}
		err = c.ReadJSON(&m)
		if err != nil {
			break
		}
		mutex.Lock()

		// send out audio data / img data to browser
		for name2, client := range sockets {
			if name == name2 {
				// never send back to self
				continue
			}
			sendData := false
			if m.Room == client.Room && m.Audio != "" && client.Room != "" {
				sendData = true
			}
			if m.Group == client.Group && client.Group != "" {
				sendData = true
			}
			if m.Recipient == name2 {
				sendData = true
			}
			if sendData {
				log.Debugf("sending data from %s to %s", name, name2)
				go func(name2 string, c2 *websocket.Conn, m models.Message) {
					wsmutex.Lock()
					err := c2.WriteJSON(m)
					wsmutex.Unlock()
					if err != nil {
						mutex.Lock()
						delete(sockets, name2)
						mutex.Unlock()
					}
				}(name2, client.conn, m)
			}
		}
		mutex.Unlock()
	}
	return
}
