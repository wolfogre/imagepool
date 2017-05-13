package main

import (
	"net/http"
	"log"
	"flag"
	"fmt"

	"github.com/go-redis/redis"
	"qiniupkg.com/api.v7/kodo"
	"qiniupkg.com/api.v7/conf"
)

// TODO 添加资源回收的逻辑

func main() {
	access := flag.String("access", "", "Access key")
	secret := flag.String("secret", "", "Secret key")
	port := flag.Int("port", 46243, "Server port")
	redis_addr := flag.String("redis", "", "Server port")
	redis_pass := flag.String("pass", "", "Redis password")
	redis_db := flag.Int("db", 0, "Server port")
	bucket := flag.String("bucket", "", "Bucket")
	domain := flag.String("domain", "", "Domain")
	flag.Parse()

	if *access == "" || *secret == "" || *redis_addr == "" || *redis_pass == "" || *bucket == "" || *domain == ""{
		flag.PrintDefaults()
		return
	}
	conf.ACCESS_KEY = *access
	conf.SECRET_KEY = *secret
	rc := redis.NewClient(&redis.Options{
		Addr: *redis_addr,
		Password: *redis_pass,
		DB: *redis_db,
	})
	kc := kodo.New(0, nil)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", *port), &MainHandler{
		Kodo: kc,
		Redis: rc,
		Domain: *domain,
		Bucket: *bucket,
	}))
}


