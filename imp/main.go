package main

import (
	"os"
	"bufio"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"qiniupkg.com/api.v7/conf"
	"qiniupkg.com/api.v7/kodo"
	"qiniupkg.com/api.v7/kodocli"
	"bytes"
	"crypto/sha256"
	"io"
	"encoding/hex"
	"strings"
	"net/http"
	"qiniupkg.com/x/errors.v7"
)

func main() {
	if err := loadConfig(); err != nil {
		fmt.Println("Load config failed: " + err.Error())
		return
	}
	if (len(os.Args) < 2) {
		fmt.Println("Please input file path")
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
		fmt.Println("Head failed: ",err.Error())
		return
	}

	fmt.Println(url)
}

var config struct{
	Access string `json:"access"`
	Secret string `json:"secret"`
	Bucket string `json:"bucket"`
	Domain string `json:"domain"`
}

func upload(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buffer, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	reader := bytes.NewReader(buffer)
	hasher := sha256.New()
	if _, err := io.Copy(hasher, reader); err != nil {
		return "", err
	}
	hash := hex.EncodeToString(hasher.Sum(nil))
	reader.Seek(0, os.SEEK_SET)

	key := hash
	if strings.LastIndex(path, ".") != -1 {
		key = key + path[strings.LastIndex(path, "."):]
	}

	conf.ACCESS_KEY = config.Access
	conf.SECRET_KEY = config.Secret
	token := kodo.New(0, nil).MakeUptoken(&kodo.PutPolicy{
		Scope: config.Bucket,
		Expires: 3600,
	})
	uploader := kodocli.NewUploader(0, nil)
	return key, uploader.Put(nil, nil, token, key, reader, int64(reader.Len()), nil)
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