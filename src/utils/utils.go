package utils

import (
	"crypto/md5"
	"io"
	"math/rand"
	"os"
)

// ConvertAudio converts to something else and removes the original
func ConvertAudio(fname string) (newfname string, err error) {
	newfname = fname
	return
	// newfname = fname + ".ogg"
	// cmd := exec.Command("ffmpeg", "-i", fname,"-codec:a","libvorbis","-qscale:a","7",newfname)
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
