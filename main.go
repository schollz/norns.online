package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hypebeast/go-osc/osc"
	log "github.com/schollz/logger"
)

var windowName string
var windowID string
var port uint
var oscClient *osc.Client

func main() {
	flag.StringVar(&windowName, "window-name", "", "X Window Name to display selected by name.")
	flag.StringVar(&windowID, "window-id", "", "X Window ID to display selected by name.")
	flag.UintVar(&port, "port", 8889, "Port to serve on.")
	flag.Parse()
	var err error
	oscClient = osc.NewClient("localhost", 10111)

	err = Run()
	if err != nil {
		log.Error(err)
	}
}

func Run() (err error) {
	log.Infof("listening on :%d", port)
	http.HandleFunc("/", handler)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	t := time.Now().UTC()
	err := handle(w, r)
	if err != nil {
		log.Error(err)
		w.Write([]byte(err.Error()))
	}
	log.Infof("%v %v %v %s\n", r.RemoteAddr, r.Method, r.URL.Path, time.Since(t))
}

func handle(w http.ResponseWriter, r *http.Request) (err error) {

	// very special paths
	if strings.HasPrefix(r.URL.Path, "/screen.png") {
		err = displayToPNG(context.Background(), windowName, windowID, w)
	} else if r.URL.Path == "/ws" {
		err = handleWebsocket(w, r)
		log.Debugf("ws: %w", err)
		log.Debug(err.Error())
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

type Message struct {
	Kind string `json:"kind"`
	N    int    `json:"n"`
	Z    int    `json:"z"`
	Fast bool   `json:"fast"`
}

func handleWebsocket(w http.ResponseWriter, r *http.Request) (err error) {
	c, errUpgrade := wsupgrader.Upgrade(w, r, nil)
	if errUpgrade != nil {
		return errUpgrade
	}
	defer c.Close()

	for {
		var p Message
		err := c.ReadJSON(&p)
		if err != nil {
			log.Debug("read:", err)
			break
		}
		log.Debugf("recv: %v", p)
		if p.Kind == "key" {
			msg := osc.NewMessage("/remote/key")
			msg.Append(int32(p.N))
			msg.Append(int32(p.Z))
			oscClient.Send(msg)
		} else if p.Kind == "enc" {
			msg := osc.NewMessage("/remote/enc")
			msg.Append(int32(p.N))
			msg.Append(int32(p.Z))
			oscClient.Send(msg)
		}
	}
	return
}

// from https://github.com/winder/norns-dev/blob/will/enable-screen/norns-test-dummy/oled-server.go
// displayToPNG uses xwd and convert to display a screenshot of the matron display.
func displayToPNG(ctx context.Context, windowName, windowID string, w io.Writer) error {
	// xwd needs -id or -name to run non-interactively.
	var capture *exec.Cmd
	if windowID != "" {
		capture = exec.CommandContext(ctx, "xwd", "-id", windowID, "-display", ":0")
	} else if windowName != "" {
		capture = exec.CommandContext(ctx, "xwd", "-name", windowName, "-display", ":0")
	} else {
		return fmt.Errorf("No windowName or windowID provided to server.")
	}
	// convert uses '-' to read from STDIN and write to STDOUT.
	convert := exec.CommandContext(ctx, "convert", "xwd:-", "png:-")

	pRead, pWrite := io.Pipe()
	var captureErr bytes.Buffer
	var convertErr bytes.Buffer

	capture.Stdout = pWrite
	capture.Stderr = &captureErr
	convert.Stdin = pRead
	convert.Stdout = w
	convert.Stderr = &convertErr

	// Start the commands
	if err := capture.Start(); err != nil {
		return err
	}
	if err := convert.Start(); err != nil {
		return err
	}
	if err := capture.Wait(); err != nil {
		if captureErr.Len() > 0 {
			return fmt.Errorf(captureErr.String())
		}
		return err
	}
	pWrite.Close()
	if captureErr.Len() > 0 {
		return fmt.Errorf(captureErr.String())
	}
	if err := convert.Wait(); err != nil {
		if convertErr.Len() > 0 {
			return fmt.Errorf(captureErr.String())
		}
		return err
	}
	if captureErr.Len() > 0 {
		return fmt.Errorf(captureErr.String())
	}

	return nil
}
