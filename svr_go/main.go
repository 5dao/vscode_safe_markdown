package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	listen         = ":8080"
	privateKeyPath = "rsa_pri"
	pubKeyPath     = "rsa_pub"
)

var rsaPri *rsa.PrivateKey
var rsaPub *rsa.PublicKey

func RSAEncrypt(origData []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, rsaPub, origData)
}

func RSADecrypt(ciphertext []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, rsaPri, ciphertext)
}

func AESEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func AESDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

type CiphertextJSON struct {
	Key       string `json:"key"`        // RAS(uuid)
	Body      string `json:"body"`       // AES-128-cbc(txt,uuid)
	MakeTs    string `json:"make_ts"`    // 2006-01-02 15:04:05
	EditCount int64  `json:"edit_count"` //
	LastTs    string `json:"last_ts"`    // 2006-01-02 15:04:05
}

type JSONResult struct {
	Err  string      `json:"err"`
	Data interface{} `json:"data"`
}

func main() {
	if err := LoadKey(); err != nil {
		log.Printf("key: %v", err)
		return
	}

	go func() {
		http.HandleFunc("/decode", decode)
		http.HandleFunc("/encode", encode)
		err := http.ListenAndServe(listen, nil)
		if err != nil {
			log.Printf("ListenAndServe: %v", err)
			os.Exit(1)
		}
	}()

	select {}
}

func decode(w http.ResponseWriter, req *http.Request) {
	rs := &JSONResult{}

	err := req.ParseForm()
	if err != nil {
		rs.Err = fmt.Sprintf("ParseForm: %v", err)
		SendJSON(w, rs)
		return
	}
	bd := req.PostFormValue("bd")

	saveJSON := CiphertextJSON{}
	err = json.Unmarshal([]byte(bd), &saveJSON)
	if err != nil {
		rs.Err = fmt.Sprintf("Unmarshal: %v", err)
		SendJSON(w, rs)
		return
	}

	keyEnBuf, err := hex.DecodeString(saveJSON.Key)
	if err != nil {
		rs.Err = fmt.Sprintf("DecodeString Key: %v", err)
		SendJSON(w, rs)
		return
	}
	key, err := RSADecrypt(keyEnBuf)
	if err != nil {
		rs.Err = fmt.Sprintf("RSADecrypt: %v", err)
		SendJSON(w, rs)
		return
	}

	bodyBuf, err := hex.DecodeString(saveJSON.Body)
	if err != nil {
		rs.Err = fmt.Sprintf("RSADecrypt Body: %v", err)
		SendJSON(w, rs)
		return
	}

	body, err := AESDecrypt(bodyBuf, key)
	if err != nil {
		rs.Err = fmt.Sprintf("RSADecrypt AESDecrypt: %v", err)
		SendJSON(w, rs)
		return
	}
	rs.Data = string(body)
	SendJSON(w, rs)
}

func encode(w http.ResponseWriter, req *http.Request) {
	rs := &JSONResult{}
	err := req.ParseForm()
	if err != nil {
		rs.Err = fmt.Sprintf("ParseForm: %v", err)
		SendJSON(w, rs)
		return
	}
	bd := req.PostFormValue("bd")

	doubleJSON := CiphertextJSON{}
	err = json.Unmarshal([]byte(bd), &doubleJSON)
	if err == nil {
		rs.Err = fmt.Sprintf("Unmarshal: %v", err)
		SendJSON(w, rs)
		return
	}

	//uuid 16 byte
	newUUID := uuid.NewV4()
	uuidKey := []byte(newUUID[:])
	//rsa(key)
	uuidKeyEn, err := RSAEncrypt(uuidKey)
	if err != nil {
		rs.Err = fmt.Sprintf("RSAEncrypt: %v", err)
		SendJSON(w, rs)
		return
	}

	//aes(bd,key)
	saveBd, err := AESEncrypt([]byte(bd), uuidKey)
	if err != nil {
		rs.Err = fmt.Sprintf("AESEncrypt: %v", err)
		SendJSON(w, rs)
		return
	}
	saveJSON := CiphertextJSON{
		Key:    hex.EncodeToString(uuidKeyEn),
		Body:   hex.EncodeToString(saveBd),
		LastTs: time.Now().Format("2006-01-02 15:04:05"),
	}
	rs.Data = saveJSON
	SendJSON(w, rs)
}

func SendJSON(w http.ResponseWriter, rs *JSONResult) {
	w.Header().Set("Content-Type", "application/json")
	outBuf, _ := json.Marshal(rs)
	w.Write(outBuf)
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func LoadKey() error {
	fmt.Print("pass:")
	pwdBuf, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("ReadPassword: %v", err)
	}
	fmt.Println("")

	{
		priKeyExist, err := PathExists(privateKeyPath)
		if err != nil || !priKeyExist {
			return fmt.Errorf("rsa_pri not exist: %v", err)
		}
		pemBuf, err := ioutil.ReadFile(privateKeyPath)
		if err != nil {
			return fmt.Errorf("rsa_pri read: %v", err)
		}
		block, _ := pem.Decode(pemBuf)
		if block == nil {
			return fmt.Errorf("rsa_pri pem=nil")
		}
		privateKeyDer, err := x509.DecryptPEMBlock(block, pwdBuf)
		if err != nil {
			return fmt.Errorf("rsa_pri DecryptPEMBlock: %v", err)
		}
		rsaPri, err = x509.ParsePKCS1PrivateKey(privateKeyDer)
		if err != nil {
			return fmt.Errorf("rsa_pri ParsePKCS1PrivateKey: %v", err)
		}
	}

	{
		pubKeyExist, err := PathExists(pubKeyPath)
		if err != nil || !pubKeyExist {
			return fmt.Errorf("rsa_pub PathExists: %v", err)
		}

		pubKeyPem, err := ioutil.ReadFile(pubKeyPath)
		if err != nil {
			return fmt.Errorf("rsa_pub ReadFile: %v", err)
		}

		block, _ := pem.Decode(pubKeyPem)
		if block == nil {
			return fmt.Errorf("rsa_pub pem=nil")
		}
		pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("rsa_pub ParsePKIXPublicKey: %v", err)
		}
		rsaPub = pubInterface.(*rsa.PublicKey)
	}

	return nil
}
