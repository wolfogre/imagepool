package main

import (
	"net/http"
	"log"
	"time"

	"qiniupkg.com/api.v7/kodo"
	"github.com/go-redis/redis"
)

type MainHandler struct {
	Kodo   *kodo.Client
	Redis  *redis.Client
	Domain string
	Bucket string
}

func (h *MainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%v] [%v] [%v]\n", r.RemoteAddr, r.Method, r.RequestURI)
	switch r.Method {
	case "HEAD":
		h.ServeHead(w, r)
	case "GET":
		h.ServeGet(w, r)
	default:
		h.ServeDefault(w, r)
	}
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
	if r.RequestURI == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Not Found\n"))
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

	url := h.Kodo.MakePrivateUrl(kodo.MakeBaseUrl(h.Domain, key), &kodo.GetPolicy{})
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	return
}

func (h *MainHandler) ServeDefault(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte("405 - Method Not Allowed\n"))
}
