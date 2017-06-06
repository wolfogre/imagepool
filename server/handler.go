package main

import (
	"net/http"
	"log"
	"time"

	"qiniupkg.com/api.v7/kodo"
	"github.com/go-redis/redis"
	"strings"
	"strconv"
	"qiniupkg.com/api.v7/auth/qbox"
	"sync"
)

type MainHandler struct {
	Kodo   *kodo.Client
	Redis  *redis.Client
	Domain string
	Bucket string
	StatusCount int
	StatusCountLock sync.Mutex
}

func (h *MainHandler) ServeHTTP(ow http.ResponseWriter, r *http.Request) {
	w := NewResponseWriter(ow)
	switch r.Method {
	case "HEAD":
		h.ServeHead(w, r)
	case "GET":
		h.ServeGet(w, r)
	default:
		h.ServeDefault(w, r)
	}

	// Log _status per 1000
	if r.RequestURI == "/_status" {
		h.StatusCountLock.Lock()
		if h.StatusCount > 1000 {
			h.StatusCount = 0
			h.StatusCountLock.Unlock()
		} else {
			h.StatusCount++
			h.StatusCountLock.Unlock()
			return
		}
	}

	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.RemoteAddr
	}
	log.Printf("[%v]\t[%v]\t[%v]\t[%fms]\t[%v]\n", ip, r.Method, w.StatusCode, w.CostTime().Seconds() * 1000, r.RequestURI)
}

func (h *MainHandler) ServeHead(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	key := r.RequestURI[1:]
	bucket := h.Kodo.Bucket(h.Bucket)

	if _, err := bucket.Stat(nil, key); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := h.Redis.Set(key, time.Now().String(), 0).Err(); err != nil {
		log.Printf("%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}

func (h *MainHandler) ServeGet(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Not Found\n"))
		return
	}

	if r.RequestURI == "/_status" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("200 - OK\n"))
		return
	}

	key := r.RequestURI[1:]

	_, err := h.Redis.Get(key).Result()
	if err != nil && err != redis.Nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - " + err.Error() + "\n"))
		return
	}

	if err == redis.Nil {
		bucket := h.Kodo.Bucket(h.Bucket)
		if _, err := bucket.Stat(nil, key); err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 - Not Found\n"))
			return
		}
	}

	if err := h.Redis.Set(key, time.Now().String(), 0).Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - " + err.Error() + "\n"))
		return
	}

	url := makePrivateUrl(h.Kodo, kodo.MakeBaseUrl(h.Domain, key))
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	return
}

func (h *MainHandler) ServeDefault(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte("405 - Method Not Allowed\n"))
}

// Like kodo.Client{}.MakePrivateUrl()
func makePrivateUrl(client *kodo.Client, baseUrl string) string{
	now := time.Now()
	zero := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var expires int64 = 24 * 60 * 60 // one day
	deadline := zero.Unix() + expires

	if strings.Contains(baseUrl, "?") {
		baseUrl += "&e="
	} else {
		baseUrl += "?e="
	}
	baseUrl += strconv.FormatInt(deadline, 10)

	token := qbox.Sign(qbox.NewMac(client.AccessKey, client.SecretKey), []byte(baseUrl))
	return baseUrl + "&token=" + token
}
