package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func IsEmpty(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true
	}
	return false
}

func UTCTime() float64 {
	// javascript: (new Date()).getTime()
	return float64(time.Now().UTC().UnixNano() / int64(time.Millisecond))
}

// CopyMax copies only the maxBytes and then returns an error if it
// copies equal to or greater than maxBytes (meaning that it did not
// complete the copy).
func CopyMax(dst io.Writer, src io.Reader, maxBytes int64) (n int64, err error) {
	n, err = io.CopyN(dst, src, maxBytes)
	if err != nil && err != io.EOF {
		return
	}

	if n >= maxBytes {
		err = fmt.Errorf("upload exceeds maximum size")
	} else {
		err = nil
	}
	return
}

func SHA256(fname string) (hash string, err error) {
	f, err := os.Open(fname)
	if err != nil {
		return
	}
	defer f.Close()

	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return
	}

	hash = fmt.Sprintf("%x", h.Sum(nil))
	return
}

// ConvertAudio converts to something else and removes the original
func ConvertAudio(fname string) (newfname string, err error) {
	// newfname = fname
	// return
	newfname = fname + ".ogg"
	cmd := exec.Command("ffmpeg", "-i", fname, "-codec:a", "libvorbis", "-qscale:a", "7", newfname)
	// newfname = fname + ".ogg"
	// cmd := exec.Command("ffmpeg", "-i", fname,"-codec:a","libvorbis","-qscale:a","7",newfname)
	//newfname = fname + ".mp3"
	//cmd := exec.Command("ffmpeg", "-i", fname,"-codec:a","libmp3lame","-b:a","320k",newfname)
	//cmd := exec.Command("ffmpeg", "-i", fname,"-codec:a","libmp3lame","-qscale:a","4","-af","silenceremove=start_periods=1:start_duration=0.01:start_threshold=-30dB:detection=peak,aformat=dblp,areverse,silenceremove=start_periods=1:start_duration=0.01:start_threshold=-30dB:detection=peak,aformat=dblp,areverse",newfname)
	//cmd := exec.Command("ffmpeg", "-i", fname,"-codec:a","libmp3lame","-qscale:a","4","-af","afade=t=in:st=0:d=0.01,afade=t=out:st=3.99:d=0.01",newfname)
	err = cmd.Run()
	os.Remove(fname) // remove original
	return
}

// ConvertToWav converts to something else and removes the original
func ConvertToWav(fname string) (newfname string, err error) {
	// newfname = fname
	// return
	newfname = strings.TrimSuffix(fname, filepath.Ext(fname))+".wav"
	fmt.Println("ffmpeg", "-y","-i", fname, newfname)
	cmd := exec.Command("ffmpeg", "-y","-i", fname, newfname)
	err = cmd.Run()
	return
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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// RandString returns a random string
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
