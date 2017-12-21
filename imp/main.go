package main

import (
	"os"
	"bufio"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"bytes"
	"crypto/sha256"
	"io"
	"encoding/hex"
	"strings"
	"net/http"
	"errors"
	"os/exec"
	"context"
	"runtime"

	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	"path"
)

func main() {
	if err := loadConfig(); err != nil {
		fmt.Println("Load config failed: " + err.Error())
		return
	}
	if len(os.Args) < 2 {
		fmt.Println("Please input file path or url")
		return
	}
	path := os.Args[1]
	key, err := upload(path)
	if err != nil {
		fmt.Println("Upload failed: ",err.Error())
		return
	}

	url := "http://" + config.Domain + "/" + key
	if err := head(url); err != nil {
		fmt.Println("Head failed: ", err.Error())
		return
	}

	fmt.Println()
	fmt.Println(url)

	fmt.Println()
	if err := clip(url); err != nil {
		fmt.Println("copy failed: ", err.Error())
	} else {
		fmt.Println("copied!")
	}
}

var config struct{
	Access string `json:"access"`
	Secret string `json:"secret"`
	Bucket string `json:"bucket"`
	Domain string `json:"domain"`
}

func upload(filename string) (string, error) {
	var buffer []byte

	if strings.HasPrefix(filename, "https://") || strings.HasPrefix(filename, "http://") {
		client := http.Client{}
		req, err := http.NewRequest("GET", filename, nil)
		if err != nil {
			return "", err
		}
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		if resp.StatusCode != http.StatusOK {
			return "", errors.New(fmt.Sprintf("%v return %v", filename, resp.StatusCode))
		}
		buffer, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", err
		}
	} else {
		file, err := os.Open(filename)
		if err != nil {
			return "", err
		}
		buffer, err = ioutil.ReadAll(file)
		file.Close()
		if err != nil {
			return "", err
		}
	}

	reader := bytes.NewReader(buffer)
	hasher := sha256.New()
	if _, err := io.Copy(hasher, reader); err != nil {
		return "", err
	}
	hash := hex.EncodeToString(hasher.Sum(nil))
	reader.Seek(0, io.SeekStart)

	key := hash + path.Ext(filename)

	putPolicy := &storage.PutPolicy{
		Scope: config.Bucket,
	}
	upToken := putPolicy.UploadToken(qbox.NewMac(config.Access, config.Secret))
	formUploader := storage.NewFormUploader(&storage.Config{
		Zone: &storage.ZoneHuadong,
		UseHTTPS: false,
		UseCdnDomains: false,
	})
	return key, formUploader.Put(context.Background(), nil, upToken, key, reader, int64(reader.Len()), nil)
}

func loadConfig() error {
	home, err := HomeDir()
	if err != nil {
		return err
	}
	path := home + "/.impconfig"
	if file, err := os.Open(path); os.IsNotExist(err) {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("access:")
		if !scanner.Scan() {
			return errors.New("refused input")
		}
		config.Access = scanner.Text()
		fmt.Print("secret:")
		if !scanner.Scan() {
			return errors.New("refused input")
		}
		config.Secret = scanner.Text()
		fmt.Print("bucket:")
		if !scanner.Scan() {
			return errors.New("refused input")
		}
		config.Bucket = scanner.Text()
		fmt.Print("domain:")
		if !scanner.Scan() {
			return errors.New("refused input")
		}
		config.Domain = scanner.Text()


		file, err = os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
		buffer, err := json.Marshal(config)
		if err != nil {
			return err
		}
		_, err = file.Write(buffer)
		return err
	} else {
		if err != nil {
			return err
		}
		defer file.Close()
		buffer, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}
		return json.Unmarshal(buffer, &config)
	}
}

func head(url string) error {
	client := http.Client{}
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("%v return %v", url, resp.StatusCode))
	}
	return nil
}

func clip(content string) error {
	if runtime.GOOS != "windows" {
		return errors.New("Only supports Windows for now")
	}
	cmd := exec.Command("clip")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, content)
	}()

	err = cmd.Run()
	return err
}