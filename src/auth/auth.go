package auth

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	log "github.com/schollz/logger"
)

func GenerateKeypair(fname string) (privatekey string, publickey string, err error) {
	privatekey = fname + ".priv"
	publickey = fname + ".pub"
	cmd := exec.Command("openssl", "genrsa", "-out", privatekey, "2048")
	err = cmd.Run()
	cmd = exec.Command("openssl", "rsa", "-in", privatekey, "-out", publickey, "-pubout")
	err = cmd.Run()
	return
}

func SignFile(privatekey, fname, signaturefile string) (err error) {
	cmd := exec.Command("openssl", "dgst", "-sign", privatekey, "-out", signaturefile, fname)
	err = cmd.Run()
	return
}

func VerifyFile(publickey, fname, signaturefile string) (err error) {
	log.Debug("openssl dgst -verify ", publickey, " -signature ", signaturefile, " ", fname)
	cmd := exec.Command("openssl", "dgst", "-verify", publickey, "-signature", signaturefile, fname)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("output: %s", output)
	if !bytes.Contains(output, []byte("OK")) {
		err = fmt.Errorf(string(output))
	}
	return
}

// VerifyString verifies plaintext public key, data and base64-encoded signature
func VerifyString(publickey, data, signature string) (err error) {
	publickeyfile, err := ioutil.TempFile("", "verify")
	if err != nil {
		return
	}
	defer os.Remove(publickeyfile.Name())
	publickeyfile.WriteString(publickey)
	err = publickeyfile.Close()
	if err != nil {
		return
	}

	datafile, err := ioutil.TempFile("", "verify")
	if err != nil {
		return
	}
	defer os.Remove(datafile.Name())
	datafile.WriteString(data)
	err = datafile.Close()
	if err != nil {
		return
	}

	signaturefile, err := ioutil.TempFile("", "verify")
	if err != nil {
		return
	}
	defer os.Remove(signaturefile.Name())
	b, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return
	}
	signaturefile.Write(b)
	err = signaturefile.Close()
	if err != nil {
		return
	}

	err = VerifyFile(publickeyfile.Name(), datafile.Name(), signaturefile.Name())
	return
}
