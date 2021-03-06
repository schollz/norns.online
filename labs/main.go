package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/url"
	"os/exec"
	"time"

	"github.com/gorilla/websocket"
	"github.com/schollz/logger"
	"github.com/schollz/norns.online/src/models"
)

func main() {
	var err error
	// err = splitAudio("philip.ogg", 60, 2)
	// if err != nil {
	// 	panic(err)
	// }
	err = sendAsNorns()
	if err != nil {
		panic(err)
	}
}

func splitAudio(fname string, splits int, seconds int) (err error) {
	curSeconds := 8
	for i := 0; i < splits; i++ {
		fmt.Println("ffmpeg", "-i", fname, "-ss", fmt.Sprint(curSeconds), "-t", fmt.Sprint(seconds), fmt.Sprintf("pp/%s.%d.opus", fname, i))
		cmd := exec.Command("ffmpeg", "-y", "-i", fname, "-ss", fmt.Sprint(curSeconds), "-t", fmt.Sprint(seconds), fmt.Sprintf("pp/%s.%d.opus", fname, i))
		err = cmd.Run()
		if err != nil {
			return
		}
		curSeconds += seconds
	}
	return
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
		Room:  "llllllll",        // tells it which audio group it wants to be in
	})

	// go func() {
	// 	for {
	// 		m := models.Message{}
	// 		ws.ReadJSON(&m)
	// 	}
	// }()
	for {
		for i := 0; i < 60; i++ {
			tt := time.Now()
			var b []byte
			b, err = ioutil.ReadFile(fmt.Sprintf("pp/philip.ogg.%d.flac", i))
			if err != nil {
				return
			}
			audiodata := base64.StdEncoding.EncodeToString(b)
			logger.Debug("sending audio data")
			err = ws.WriteJSON(models.Message{
				Audio: audiodata,
			})
			if err != nil {
				return
			}
			t2 := int(2000 - time.Since(tt).Seconds()*1000)
			fmt.Println(t2)
			time.Sleep(time.Duration(t2) * time.Millisecond) // length of this sample
		}

	}
	return
}
