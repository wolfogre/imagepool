package main

import (
	"path/filepath"
	"os"
	"path"
	"strings"
	"io/ioutil"
	"regexp"
	"net/http"
	"errors"
	"fmt"
	"log"
)

var (
	reg = regexp.MustCompile(`!\[.*]\((?P<url>.*)\)`)
	count, sum int64
)

func main() {
	filepath.Walk(`C:\Users\wolfo\AppData\Local\YNote\data\`, walkfunc)
	log.Println(count, " ", sum)
}

func walkfunc(p string, i os.FileInfo, e error) error {
	if e != nil {
		return e
	}

	if i.IsDir() {
		return nil
	}

	if strings.ToLower(path.Ext(p)) != ".md" {
		return nil
	}

	log.Println(p)

	buffer, err := ioutil.ReadFile(p)
	if err != nil {
		return err
	}

	content := string(buffer)
	matchs := reg.FindAllStringSubmatch(content, -1)
	for _, v := range matchs {
		if strings.HasPrefix(v[1], "http://image.wolfogre.com") {
			log.Println(v[1])
			resp, err := http.Head(v[1])
			if err != nil {
				return err
			}
			if resp.StatusCode != http.StatusOK {
				return errors.New(fmt.Sprintf("%v return %v", v[1], resp.StatusCode))
			}
			count++
			if resp.ContentLength > 0 {
				sum += resp.ContentLength
			}
		}
	}
	return nil
}