package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/schollz/logger"
	"github.com/schollz/norns.online/src/auth"
	"github.com/schollz/norns.online/src/models"
	"github.com/schollz/norns.online/src/utils"
)

var sockets map[string]Client
var mutex sync.Mutex
var wsmutex sync.Mutex

const MaxBytesPerFile = 100000000

type Metadata struct {
	// user submits
	DataName string `json:"dataname"` // name of the data
	Username string `json:"username"`
	Type     string `json:"type"` // name of script, or "tape"
	Files    []File `json:"files"`

	// server fills in
	Date time.Time `json:"date"`
}

type File struct {
	Name      string `json:"name"`
	Target    string `json:"target"`
	Hash      string `json:"hash"`      // hash of the data
	Signature string `json:"signature"` // signature of the hash
}

type Client struct {
	Group string
	Room  string
	conn  *websocket.Conn
}

func Run() (err error) {
	sockets = make(map[string]Client)
	os.MkdirAll("share/keys", os.ModePerm)
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
		w.Write([]byte(err.Error()))
	}
	log.Infof("%v %v %v %s\n", r.RemoteAddr, r.Method, r.URL.Path, time.Since(t))
}

func handle(w http.ResponseWriter, r *http.Request) (err error) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	// very special paths
	if strings.HasPrefix(r.URL.Path, "/upload") {
		// PUT file
		// this is called from curl/wget upload
		return handleUpload(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/register") {
		// this is called from curl/wget upload
		return handleRegister(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/unregister") {
		// this is called from curl/wget upload
		return handleUnregister(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/directory.json") {
		// this is called from curl/wget upload
		return handleDirectory(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/delete") {
		// this is called from curl/wget upload
		return handleDelete(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/share") {
		if strings.HasSuffix(r.URL.Path, "/") {
			s := strings.TrimPrefix(r.URL.Path, "/share/")
			if s == "" {
				s = "norns.online/share"
			} else {
				parts := strings.Split(s, "/")
				s = ""
				for i, part := range parts {
					if part == "" {
						continue
					}
					s += fmt.Sprintf("<a href='/share/%s/%s'>%s</a>/", strings.Join(parts[:i], "/"), part, part)
				}
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			io.WriteString(w, `<style>
@font-face {
    font-family: 'uni 05_53';
    font-style: normal;
    font-weight: 400;
    src: local('uni 05_53'), local('uni05_53-Regular'), url(/static/font/uni-05_53_8dd1788135e93e81fb9993c630d92da3.woff) format('woff'), url(/static/font/uni-05_53_8dd1788135e93e81fb9993c630d92da3.ttf) format('truetype');
}
a {
    color: inherit;
}
a:hover {
    color: 9a9f9b;
}
pre {
font-family: 'uni 05_53', arial;
}
body{padding:1em;margin:auto;max-width:800px;color:#fff; font-family: 'uni 05_53', arial; background-color: #222222;font-size:2em;font-weight:bold;</style><pre style="margin-bottom:-1em;">
`+s+`
`)
			if !strings.HasSuffix(r.URL.Path, "/share/") {
				io.WriteString(w, `<a href="../">..</a>
`)
			}
			io.WriteString(w, `</pre>`)
		}
		http.FileServer(http.Dir(".")).ServeHTTP(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/share") {
		http.Redirect(w, r, "/share/", 302)
	} else if r.URL.Path == "/ws" {
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

func handleDirectory(w http.ResponseWriter, r *http.Request) (err error) {
	files := make(map[string]bool)
	err = filepath.Walk("share",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			path = filepath.ToSlash(path)
			path = strings.TrimPrefix(path, "share/")
			if strings.Count(path, "/") == 2 {
				files[path] = true
			}
			return nil
		})
	if err != nil {
		log.Error(err)
		return
	}

	b, _ := json.MarshalIndent(files, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
	return
}

func handleRegister(w http.ResponseWriter, r *http.Request) (err error) {
	var username, signature string
	if _, ok := r.URL.Query()["username"]; ok {
		username = r.URL.Query()["username"][0]
	} else {
		err = fmt.Errorf("no username")
		return
	}

	// make sure username has valid characters
	isAlpha := regexp.MustCompile(`^[A-Za-z0-9]+$`).MatchString
	if !isAlpha(username) {
		err = fmt.Errorf("bad username")
		return
	}

	// make sure username is not taken
	if _, err = os.Stat("share/keys/" + username); !os.IsNotExist(err) {
		err = fmt.Errorf("username taken")
		return
	}

	// signature should be base64 encoded
	if _, ok := r.URL.Query()["signature"]; ok {
		signature = strings.Replace(r.URL.Query()["signature"][0], " ", "+", -1)
	} else {
		err = fmt.Errorf("no signature")
		return
	}
	log.Debugf("signature: %+v", signature)

	// read in the public key to a file
	fname, err := writeAllBytes(r.Body, 20000)
	defer os.Remove(fname) // always remove
	if err != nil {
		return
	}
	log.Debugf("wrote '%s' input", fname)

	// verify signature using username
	pubkey, _ := ioutil.ReadFile(fname)
	err = auth.VerifyString(string(pubkey), username, signature)
	if err != nil {
		log.Error(err)
		err = fmt.Errorf("could not verify signature")
		return
	}

	// add to the database
	err = os.Rename(fname, "share/keys/"+username)
	if err == nil {
		w.Write([]byte("registration OK"))
	}
	return
}

func handleUnregister(w http.ResponseWriter, r *http.Request) (err error) {
	var username, signature string
	if _, ok := r.URL.Query()["username"]; ok {
		username = r.URL.Query()["username"][0]
	} else {
		err = fmt.Errorf("no username")
		return
	}

	// make sure username has valid characters
	isAlpha := regexp.MustCompile(`^[A-Za-z0-9]+$`).MatchString
	if !isAlpha(username) {
		err = fmt.Errorf("bad username")
		return
	}

	// make sure username exists
	if _, err = os.Stat("share/keys/" + username); os.IsNotExist(err) {
		err = nil
		return
	}

	// signature should be base64 encoded
	if _, ok := r.URL.Query()["signature"]; ok {
		signature = strings.Replace(r.URL.Query()["signature"][0], " ", "+", -1)
	} else {
		err = fmt.Errorf("no signature")
		return
	}
	log.Debugf("signature: %+v", signature)

	// read in the public key
	pubkey, err := ioutil.ReadFile("share/keys/" + username)
	if err != nil {
		return
	}

	// verify signature using username
	err = auth.VerifyString(string(pubkey), username, signature)
	if err != nil {
		log.Error(err)
		err = fmt.Errorf("could not verify signature")
		return
	}

	// remove to the database
	err = os.Remove("share/keys/" + username)
	if err == nil {
		w.Write([]byte("...unregistration OK"))
	}
	return
}

func handleDelete(w http.ResponseWriter, r *http.Request) (err error) {
	isValidFilename := regexp.MustCompile(`^[A-Za-z0-9.\-_]+$`).MatchString

	var username, signature, datatype, dataname string
	if _, ok := r.URL.Query()["username"]; ok {
		username = r.URL.Query()["username"][0]
	} else {
		err = fmt.Errorf("no username")
		return
	}

	// make sure username has valid characters
	isAlpha := regexp.MustCompile(`^[A-Za-z0-9]+$`).MatchString
	if !isAlpha(username) {
		err = fmt.Errorf("bad username")
		return
	}

	if _, ok := r.URL.Query()["type"]; ok {
		datatype = filepath.Base(r.URL.Query()["type"][0])
	} else {
		err = fmt.Errorf("no type")
		return
	}
	// make sure type has valid characters
	if !isValidFilename(datatype) {
		err = fmt.Errorf("bad type")
		return
	}

	if _, ok := r.URL.Query()["dataname"]; ok {
		dataname = filepath.Base(r.URL.Query()["dataname"][0])
	} else {
		err = fmt.Errorf("no dataname")
		return
	}
	// make sure dataname has valid characters
	if !isValidFilename(dataname) {
		err = fmt.Errorf("bad data name: '%s'", dataname)
		return
	}

	// signature should be base64 encoded
	if _, ok := r.URL.Query()["signature"]; ok {
		signature = strings.Replace(r.URL.Query()["signature"][0], " ", "+", -1)
	} else {
		err = fmt.Errorf("no signature")
		return
	}
	log.Debugf("signature: %+v", signature)

	// verify signature using username
	// verify the signature of the hash of data
	keyb, err := ioutil.ReadFile("share/keys/" + username)
	if err != nil {
		log.Error(err)
		return
	}
	err = auth.VerifyString(string(keyb), username, signature)
	if err != nil {
		log.Error(err)
		err = fmt.Errorf("could not verify signature")
		return
	}

	pathToFile := path.Join("share", datatype, username, dataname)
	if _, err = os.Stat(pathToFile); os.IsNotExist(err) {
		err = fmt.Errorf("does not exist")
		return
	}
	// add to the database
	err = os.RemoveAll(pathToFile)
	if err == nil {
		if utils.IsEmpty(path.Join("share", datatype, username)) {
			os.RemoveAll(path.Join("share", datatype, username))
			if utils.IsEmpty(path.Join("share", datatype)) {
				os.RemoveAll(path.Join("share", datatype))
			}
		}
		w.Write([]byte("data deleted"))
	}
	return
}

func handleUpload(w http.ResponseWriter, r *http.Request) (err error) {
	isValidFilename := regexp.MustCompile(`^[A-Za-z0-9.\-_]+$`).MatchString

	var m Metadata
	if _, ok := r.URL.Query()["username"]; ok {
		m.Username = r.URL.Query()["username"][0]
	} else {
		err = fmt.Errorf("no username")
		return
	}

	// make sure username exists
	if _, err = os.Stat("share/keys/" + m.Username); os.IsNotExist(err) {
		err = fmt.Errorf("need to register first")
		return
	}

	if _, ok := r.URL.Query()["type"]; ok {
		m.Type = r.URL.Query()["type"][0]
	} else {
		err = fmt.Errorf("no type")
		return
	}
	// make sure type has valid characters
	if !isValidFilename(m.Type) {
		err = fmt.Errorf("bad type")
		return
	}

	m.Files = make([]File, 1)
	if _, ok := r.URL.Query()["filename"]; ok {
		m.Files[0].Name = filepath.Base(r.URL.Query()["filename"][0])
	} else {
		err = fmt.Errorf("no type")
		return
	}

	if _, ok := r.URL.Query()["dataname"]; ok {
		m.DataName = r.URL.Query()["dataname"][0]
	} else {
		err = fmt.Errorf("no dataname")
		return
	}
	// make sure dataname has valid characters
	if !isValidFilename(m.DataName) {
		err = fmt.Errorf("bad data name: '%s'", m.DataName)
		return
	}

	if _, ok := r.URL.Query()["hash"]; ok {
		m.Files[0].Hash = r.URL.Query()["hash"][0]
	} else {
		err = fmt.Errorf("no hash")
		return
	}

	if _, ok := r.URL.Query()["target"]; ok {
		m.Files[0].Target = r.URL.Query()["target"][0]
	} else {
		err = fmt.Errorf("no target")
		return
	}

	if _, ok := r.URL.Query()["signature"]; ok {
		m.Files[0].Signature = strings.Replace(r.URL.Query()["signature"][0], " ", "+", -1)
	} else {
		err = fmt.Errorf("no signature")
		return
	}

	m.Date = time.Now()
	log.Debugf("m: %+v", m)

	fname, err := writeAllBytes(r.Body, 200000000)
	defer os.Remove(fname) // always remove
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("wrote '%s' input", fname)

	// verify the hash of data
	hash, err := utils.SHA256(fname)
	if err != nil {
		log.Error(err)
		return
	}
	if m.Files[0].Hash != hash {
		err = fmt.Errorf("obs hash '%s' does not equal spoken hash '%s'", hash, m.Files[0].Hash)
		return
	}

	// verify the signature of the hash of data
	keyb, err := ioutil.ReadFile("share/keys/" + m.Username)
	if err != nil {
		log.Error(err)
		return
	}
	err = auth.VerifyString(string(keyb), m.Files[0].Name+m.Files[0].Target+m.Files[0].Hash, m.Files[0].Signature)
	if err != nil {
		log.Error(err)
		err = fmt.Errorf("could not verify signature")
		return
	}

	// everything is good write it out
	os.MkdirAll(path.Join("share", m.Type, m.Username, m.DataName), os.ModePerm)
	err = os.Rename(fname, path.Join("share", m.Type, m.Username, m.DataName, m.Files[0].Name))
	if err != nil {
		log.Error(err)
		return
	}
	// check if metadata exists, and if it does combine meta datas
	metadatajson := path.Join("share", m.Type, m.Username, m.DataName, "metadata.json")
	bcurrent, err := ioutil.ReadFile(metadatajson)
	if err == nil {
		var mcurrent Metadata
		err = json.Unmarshal(bcurrent, &mcurrent)
		if err != nil {
			log.Error(err)
			return
		}
		// add in things not in current metadata
		for _, file := range mcurrent.Files {
			if file.Name != m.Files[0].Name {
				m.Files = append(m.Files, file)
			}
		}
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		log.Error(err)
		return
	}
	err = ioutil.WriteFile(metadatajson, b, 0644)
	if err == nil {
		w.Write([]byte("...upload OK"))
	}
	return
}

// writeAllBytes takes a reader and writes it to the content directory.
// It throws an error if the number of bytes written exceeds what is set.
func writeAllBytes(src io.Reader, maxbytes int64) (fname string, err error) {
	f, err := ioutil.TempFile(".", "sharetemp")
	if err != nil {
		log.Error(err)
		return
	}
	fname = f.Name()
	// w := gzip.NewWriter(f)
	n, err := utils.CopyMax(f, src, maxbytes)
	// w.Flush()
	// w.Close()
	f.Close()

	// if an error occured, then erase the temp file
	if err != nil {
		os.Remove(f.Name())
		return
	} else {
		log.Debugf("wrote %d bytes to %s", n, f.Name())
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
	group := m.Group
	room := m.Room
	if name == "" || (group == "" && room == "") {
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
		if m.TimeSend {
			// send back time
			m.TimeServer = utils.UTCTime()
			err = c.WriteJSON(m)
			if err != nil {
				break
			}
			continue
		}
		m.Sender = name // update the sender information
		if m.Audio != "" {
			log.Debugf("got audio from %s in group %s and room %s", name, group, room)
		}

		// send out audio data / img data to browser
		for name2, client := range sockets {
			if name == name2 {
				// never send back to self
				continue
			}
			sendData := false
			if room == client.Room && m.Audio != "" && client.Room != "" {
				sendData = true
			}
			if group == client.Group && client.Group != "" {
				sendData = true
			}
			if m.Recipient == name2 {
				sendData = true
			}
			if sendData {
				go func(name2 string, c2 *websocket.Conn, m models.Message) {
					c2.SetWriteDeadline(time.Now().Add(1 * time.Second))
					wsmutex.Lock()
					err := c2.WriteJSON(m)
					wsmutex.Unlock()
					if err != nil {
						log.Error(err)
						mutex.Lock()
						delete(sockets, name2)
						mutex.Unlock()
					}
					log.Debugf("sent data to %s", name2)
				}(name2, client.conn, m)
			}
		}
	}
	return
}
