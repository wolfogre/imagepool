package main

import (
	"net/http"
	"log"
	"flag"
	"fmt"
	"time"

	"gopkg.in/redis.v3"
	"github.com/qiniu/api.v7/auth/qbox"
)

// TODO 添加资源回收的逻辑

func main() {
	access := flag.String("access", "", "Access key")
	secret := flag.String("secret", "", "Secret key")
	port := flag.Int("port", 46243, "Server port")
	redis_addr := flag.String("redis", "", "Server port")
	redis_pass := flag.String("pass", "", "Redis password")
	redis_db := flag.Int64("db", 0, "Server port")
	bucket := flag.String("bucket", "", "Bucket")
	domain := flag.String("domain", "", "Domain")
	flag.Parse()

	if *access == "" || *secret == "" || *redis_addr == "" || *redis_pass == "" || *bucket == "" || *domain == ""{
		flag.PrintDefaults()
		return
	}

	log.SetFlags(0)
	log.SetOutput(NewLogWriter{})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", *port), &MainHandler{
		Mac:    qbox.NewMac(*access, *secret),
		Redis:  redis.NewClient(&redis.Options{
			Addr: *redis_addr,
			Password: *redis_pass,
			DB: *redis_db,
		}),
		Domain: *domain,
		Bucket: *bucket,
	}))
}

type NewLogWriter struct{}

func (h NewLogWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().Local().Format("2006/01/02 15:04:05.999999 ") + string(bytes))
}

