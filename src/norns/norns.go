package norns

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gorilla/websocket"
	"github.com/schollz/logger"
	"github.com/schollz/norns.online/src/models"
	"github.com/schollz/norns.online/src/utils"
)

type Norns struct {
	Name        string `json:"name"`
	Room        string `json:"room"` // designates where it wants to receive audio\
	AllowMenu   bool   `json:"allowmenu"`
	AllowEncs   bool   `json:"allowencs"`
	AllowKeys   bool   `json:"allowkeys"`
	AllowTwitch bool   `json:"allowtwitch"`
	AllowRoom   bool   `json:"allowroom"`
	SendAudio   bool   `json:"sendaudio"`
	KeepAwake   bool   `json:"keepawake"`
	FrameRate   int    `json:"framerate"`
	PacketSize  int    `json:"packetsize"`
	BufferTime  int    `json:"buffertime"`
	RoomSize    int    `json:"roomsize"`
	RoomVolume  int    `json:"roomvolume"`

	srcbkg         image.Image
	configFile     string
	configFileHash []byte
	active         bool
	inMenu         bool
	norns          *websocket.Conn
	ws             *websocket.Conn
	incomingAudio  chan string

	mpvs           map[string]string // map sender to filename
	timeSinceAudio time.Time

	streamPosition int

	sync.Mutex
}

// New returns a new instance
func New(configFile string, pid int32) (n *Norns, err error) {
	if configFile == "" {
		err = fmt.Errorf("need config file!")
		return
	}
	// write scripts
	ioutil.WriteFile("/dev/shm/norns.online.kill.sh", []byte(`#!/bin/bash
# this script is meant to kill norns.online tasks
kill -9 `+fmt.Sprint(pid)+`
pkill jack_capture
pkill mpv
rm -rf /dev/shm/norns.online*.flac
rm -rf /dev/shm/jack_capture*.flac
rm -rf /dev/shm/norns.online*.ogg
rm -rf /dev/shm/jack_capture*.ogg
rm -rf /dev/shm/norns.online.mpv*
rm -rf /dev/shm/norns.online.screenshot*
rm -rf /dev/shm/norns.online.*sh
rm -- "$0"
	`), 0777)

	n = new(Norns)
	n.mpvs = make(map[string]string)
	n.timeSinceAudio = time.Now()
	n.configFile = configFile
	n.AllowEncs = true
	n.AllowKeys = true
	n.KeepAwake = false
	n.FrameRate = 4
	n.srcbkg, err = imaging.Open("/home/we/dust/code/norns.online/static/img/background.png")
	n.incomingAudio = make(chan string, 300)
	if err != nil {
		logger.Error(err)
		return
	}
	_, err = n.Load()
	if err != nil {
		logger.Error(err)
		return
	}
	if n.PacketSize < 1 {
		n.PacketSize = 1
	}
	if n.BufferTime < 1 {
		n.BufferTime = 1000
	}
	if n.RoomSize < 1 {
		n.RoomSize = 1
	}
	if n.RoomVolume == 0 {
		n.RoomVolume = 80
	}

	startsh := `#!/bin/bash
cd /dev/shm
rm -rf /dev/shm/norns.online*.flac
rm -rf /dev/shm/jack_capture*.flac
rm -rf /dev/shm/norns.online*.ogg
rm -rf /dev/shm/jack_capture*.ogg
rm -rf /dev/sh/norns.online.mpv*
chmod +x /home/we/dust/code/norns.online/jack_capture
/home/we/dust/code/norns.online/jack_capture -f flac --port crone:output_1 --port crone:output_2 --recording-time 36000 -Rf ` + fmt.Sprint(n.PacketSize*48000) + ` -z 4 &
`
	if n.Room != "" && n.AllowRoom {
		startsh += `	
# launch playback server
`
		for i := 0; i < n.RoomSize; i++ {
			startsh += `
mkfifo /dev/shm/norns.online.mpv` + fmt.Sprint(i) + `
sleep 0.1
mpv --merge-files=yes --gapless-audio=yes --no-video --jack-port="system:playback_(1|2)" --input-file=/dev/shm/norns.online.mpv` + fmt.Sprint(i) + ` --idle &
sleep 0.6
echo "set_property volume ` + fmt.Sprint(n.RoomVolume) + `" > /dev/shm/norns.online.mpv` + fmt.Sprint(i) + `
sleep 0.3
`
		}

	}
	ioutil.WriteFile("/dev/shm/norns.online.start.sh", []byte(startsh), 0777)

	go n.connectToWebsockets()
	return
}

// Load will update the configuration if config file changes
func (n *Norns) Load() (updated bool, err error) {
	currentHash, err := utils.MD5HashFile(n.configFile)
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

func (n *Norns) connectToWebsockets() (err error) {
	for {
		if n.ws != nil {
			// close connection and try reconnecting
			n.ws.Close()
			n.ws = nil
			time.Sleep(500 * time.Millisecond)
		}
		// wsURL := url.URL{Scheme: "ws", Host: "192.168.0.3:8098", Path: "/ws"}
		wsURL := url.URL{Scheme: "wss", Host: "norns.online", Path: "/ws"}
		logger.Debugf("connecting to %s as %s", wsURL.String(), n.Name)
		n.ws, _, err = websocket.DefaultDialer.Dial(wsURL.String(), nil)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		// initially tell the server who i am
		n.Lock()
		n.ws.WriteJSON(models.Message{
			Name:  n.Name, // the norns goes by its name in its group
			Group: n.Name, // a norns designates a group by its name
			Room:  n.Room, // tells it which audio group it wants to be in
		})
		n.Unlock()
		pings := 0
		for {
			var m models.Message
			logger.Debug("waiting for message")
			err = n.ws.ReadJSON(&m)
			if err != nil {
				logger.Debug(err)
				break
			}
			if m.Audio == "" {
				logger.Debugf("got models.Message from %s: %+v", m.Sender, m)
			} else {
				logger.Debugf("got message with %d bytes of audio from %s", len(m.Audio), m.Sender)
			}

			cmd, err := n.processMessage(m)
			if err != nil {
				continue
			}
			if cmd == "" {
				continue
			}
			logger.Debugf("running command: '%s'", cmd)
			n.Lock()
			err = n.norns.WriteMessage(websocket.TextMessage, []byte(cmd+"\n"))
			n.Unlock()
			if err != nil {
				logger.Error(err)
				continue
			}
			pings++
			if pings%20 == 0 && n.KeepAwake {
				n.Lock()
				err = n.norns.WriteMessage(websocket.TextMessage, []byte(`screen.ping()`+"\n"))
				n.Unlock()
				if err != nil {
					logger.Error(err)
					continue
				}
			}

		}
	}
	return
}

// Run forever
func (n *Norns) Run() (err error) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	if n.SendAudio {
		logger.Debug("sending audio")
		cmd := exec.Command("/dev/shm/norns.online.start.sh")
		if err = cmd.Start(); err != nil {
			return
		}
		go n.Stream() // cleans up captured files

		defer func() {
			logger.Debug("killing jack capture")
			// Kill it just in case it iddn't get killed
			if err = cmd.Process.Kill(); err != nil {
				logger.Error("failed to kill process: ", err)
			}
			cmd = exec.Command("pkill", "jack_capture")
			cmd.Start()
			cmd = exec.Command("pkill", "mpv")
			cmd.Start()
			logger.Info("killed")
		}()
	}

	go func() {
		for {
			// process incoming audio
			cmds := strings.Split(<-n.incomingAudio, "|")
			if len(cmds) != 2 {
				continue
			}
			logger.Debugf("timeSince: %d ms", int(time.Since(n.timeSinceAudio).Seconds()*1000))

			// TODO make this deviation user adjustable instead of just 1.5
			if time.Since(n.timeSinceAudio).Seconds() > float64(n.PacketSize)*1.5 {
				// buffer for packet size
				logger.Debugf("buffering for %d ms", n.BufferTime)
				time.Sleep(time.Duration(n.BufferTime) * time.Millisecond)
			}
			n.timeSinceAudio = time.Now()

			errCmd := sendCommandToMPV(cmds[0], cmds[1])
			if errCmd != nil {
				logger.Error(errCmd)
			}
			time.Sleep(100 * time.Millisecond)
		}

	}()

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
	if n.FrameRate == 1 {
		ticker = time.NewTicker(1 * time.Second)
	} else if n.FrameRate == 0 {
		ticker = time.NewTicker(3 * time.Minute)
	} else {
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
				ticker = nil
				if n.FrameRate == 1 {
					ticker = time.NewTicker(1 * time.Second)
				} else if n.FrameRate == 0 {
					ticker = time.NewTicker(3 * time.Minute)
				} else {
					ticker = time.NewTicker(time.Duration(1000/n.FrameRate) * time.Millisecond)
				}
				if n.Name != currentName {
					// restart websockets with new name
					n.ws.Close()
					go n.connectToWebsockets()
				}
				// update the volume
				if n.Room != "" && n.AllowRoom {
					for i := 0; i < n.RoomSize; i++ {
						sendCommandToMPV("set_property volume "+fmt.Sprint(n.RoomVolume), "/dev/shm/norns.online.mpv"+fmt.Sprint(i))
					}
				}
			}
		case _ = <-ticker.C:
			n.Lock()
			err = n.norns.WriteMessage(websocket.TextMessage, []byte(`_norns.screen_export_png("/dev/shm/norns.online.screenshot.png")`+"\n"))
			n.Unlock()
			if err != nil {
				logger.Debugf("write: %w", err)
				return
			}
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
			logger.Debug("returning")
			return
		}
	}
	logger.Debug("exiting")
	return
}

func (n *Norns) updateClient() (err error) {
	// open dumped image
	src, err := imaging.Open("/dev/shm/norns.online.screenshot.png")
	if err != nil {
		return
	}

	src = imaging.Resize(src, 510, 0, imaging.NearestNeighbor) // full width is 550, padding is added
	src = imaging.AdjustGamma(src, 1.25)
	src = imaging.OverlayCenter(n.srcbkg, src, 1)
	err = imaging.Save(src, "/dev/shm/norns.online.screenshot2.png")
	if err != nil {
		return
	}

	b, err := ioutil.ReadFile("/dev/shm/norns.online.screenshot2.png")
	if err != nil {
		return
	}
	base64data := base64.StdEncoding.EncodeToString(b)

	tsent := time.Now()
	if n.ws != nil {
		n.Lock()
		n.ws.WriteJSON(models.Message{
			Img:    base64data,
			Twitch: n.AllowTwitch,
		})
		n.Unlock()
	}
	logger.Tracef("sent data in %s", time.Since(tsent))
	return
}

// processmodels.Message only lets certain k inds of models.Messages through
func (n *Norns) processMessage(m models.Message) (cmd string, err error) {
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
	} else if m.Audio != "" {
		// got some audio!
		go func(sender, audio string) {
			errF := n.processAudio(sender, audio)
			if errF != nil {
				logger.Error(errF)
			}
		}(m.Sender, m.Audio)
	}
	if n.inMenu {
		cmd = "_menu." + cmd
	}

	return
}

func (n *Norns) processAudio(sender, audioData string) (err error) {
	// first write the audio to a file
	audioBytes, err := base64.StdEncoding.DecodeString(audioData)
	if err != nil {
		return
	}
	audioFile, err := ioutil.TempFile("/dev/shm", "norns.online.incoming.*.flac")
	if err != nil {
		return
	}
	filename := audioFile.Name()
	audioFile.Write(audioBytes)
	audioFile.Close()

	defer func() {
		go func() {
			// remove file after 6 seconds
			time.Sleep(6000 * time.Millisecond)
			os.Remove(filename)
		}()
	}()

	// figure out which mpv to use
	if _, ok := n.mpvs[sender]; !ok {
		if len(n.mpvs) == n.RoomSize {
			err = fmt.Errorf("can't support any more in room")
			return
		}
		n.mpvs[sender] = fmt.Sprintf("/dev/shm/norns.online.mpv%d", len(n.mpvs))
	}

	n.incomingAudio <- "loadfile " + filename + " append-play|" + n.mpvs[sender]
	logger.Debug("audio processed!")

	return
}

func sendCommandToMPV(command, fifofile string) (err error) {
	bashFile := "/dev/shm/norns.online.input" + utils.RandString(5) + ".sh"
	defer os.Remove(bashFile)
	bashData := `#!/bin/bash
echo "` + command + `" > ` + fifofile + `
`
	err = ioutil.WriteFile(bashFile, []byte(bashData), 0777)
	if err != nil {
		return
	}

	logger.Debugf("sending command %s to %s", command, fifofile)
	cmd := exec.Command(bashFile)
	err = cmd.Run()
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

// FindChangingFile returns the name of the file that's changing
// (the one that's being recorded)
func (n *Norns) Stream() (filename string, err error) {
	currentFile := make(chan string, 1)
	go func() {
		for {
			// clean up jack captured files
			files, _ := filepath.Glob("/dev/shm/jack_capture*.flac")
			if len(files) > 1 {
				currentFile <- files[0]
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	for {
		fname := <-currentFile // current file is a flac file
		// // convert audio from flac
		// fname, err = utils.ConvertAudio(fname)
		// if err != nil {
		// 	logger.Error(err)
		// 	os.Remove(fname)
		// 	continue
		// }
		logger.Debugf("processing %s", fname)
		b, errb := ioutil.ReadFile(fname)
		if errb != nil {
			return
		}
		os.Remove(fname)
		if n.ws != nil {
			audiodata := base64.StdEncoding.EncodeToString(b)
			logger.Debugf("sending %d bytes of data", len(audiodata))
			n.Lock()
			n.ws.WriteJSON(models.Message{
				Audio: audiodata,
			})
			n.Unlock()
		}
	}

	return
}
