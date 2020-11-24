package main

import (
	"encoding/base64"
	"io/ioutil"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/schollz/logger"
	"norns.online/src/models"
)

func main() {
	err := sendAsNorns()
	if err != nil {
		panic(err)
	}
}

func sendAsNorns() (err error) {
	// pretend to be a norns connecting to a room and sending audio data
	// wsURL := url.URL{Scheme: "ws", Host: "192.168.0.3:8098", Path: "/ws"}
	wsURL := url.URL{Scheme: "wss", Host: "norns.online", Path: "/ws"}
	logger.Debugf("connecting to %s", wsURL.String())
	ws, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		return
	}
	// initially tell the server who i am
	ws.WriteJSON(models.Message{
		Name:  "infinitedigits2", // the norns goes by its name in its group
		Group: "infinitedigits2", // a norns designates a group by its name
		Room:  "AA",              // tells it which audio group it wants to be in
	})

	b, err := ioutil.ReadFile("rach.mp3")
	if err != nil {
		return
	}
	audiodata := base64.StdEncoding.EncodeToString(b)

	// go func() {
	// 	for {
	// 		m := models.Message{}
	// 		ws.ReadJSON(&m)
	// 	}
	// }()
	for {
		logger.Debug("sending audio data")
		err = ws.WriteJSON(models.Message{
			Audio: audiodata,
		})
		if err != nil {
			return
		}
		time.Sleep(3144 * time.Millisecond) // length of this sample
	}
	return
}
