package main

import (
	"qiniupkg.com/api.v7/kodo"
	"qiniupkg.com/api.v7/conf"
	"net/http"
	"log"
	"flag"
	"fmt"


	"gopkg.in/redis.v3"
	"time"
)

var (
	client *redis.Client
	bucket kodo.Bucket

	domain = flag.String("domain", "", "Domain")
)

func main() {
	access := flag.String("access", "", "Access key")
	secret := flag.String("secret", "", "Secret key")
	port := flag.Int("port", 46243, "Server port")
	redis_addr := flag.String("redis", "", "Server port")
	redis_pass := flag.String("pass", "", "Redis password")
	redis_db := flag.Int64("db", 0, "Server port")
	bucket_name := flag.String("bucket", "", "Bucket")
	flag.Parse()
	if *access == "" || *secret == "" || *redis_addr == "" || *redis_pass == "" || *bucket_name == "" || *domain == ""{
		flag.PrintDefaults()
		return
	}
	conf.ACCESS_KEY = *access
	conf.SECRET_KEY = *secret
	client = redis.NewClient(&redis.Options{
		Addr: *redis_addr,
		Password: *redis_pass,
		DB: *redis_db,
	})
	_, err := client.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}
	bucket = kodo.New(0, nil).Bucket(*bucket_name)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", *port), &mainHandler{}))
}

type mainHandler struct {

}

func (h *mainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

//func (h *mainHandler) ServePut(w http.ResponseWriter, r *http.Request) {
//	buffer, err := ioutil.ReadAll(r.Body)
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		w.Write([]byte("500 - " + err.Error() + "\n"))
//		return
//	}
//	r.Body.Close()
//
//	reader := bytes.NewReader(buffer)
//	hasher := sha256.New()
//	if _, err := io.Copy(hasher, reader); err != nil {
//		log.Fatal(err)
//		w.WriteHeader(http.StatusInternalServerError)
//		w.Write([]byte("500 - " + err.Error() + "\n"))
//		return
//	}
//	hash := hex.EncodeToString(hasher.Sum(nil))
//	reader.Reset(buffer)
//	err = upload(hash, reader, r.ContentLength)
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		w.Write([]byte("500 - " + err.Error() + "\n"))
//		return
//	}
//	w.WriteHeader(http.StatusCreated)
//	w.Write([]byte("\n\nhttp://" + HOST + "/" + hash + "\n\n"))
//}


func (h *mainHandler) ServeHead(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	key := r.RequestURI[1:]

	if _, err := bucket.Stat(nil, key); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := client.Set(key, time.Now().String(), 0).Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}

func (h *mainHandler) ServeGet(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Not Found\n"))
		return
	}

	key := r.RequestURI[1:]

	_, err := client.Get(key).Result()
	if err != nil && err != redis.Nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - " + err.Error() + "\n"))
		return
	}

	if err == redis.Nil {
		if _, err := bucket.Stat(nil, key); err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 - Not Found\n"))
			return
		}
	}

	if err := client.Set(key, time.Now().String(), 0).Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - " + err.Error() + "\n"))
		return
	}

	http.Redirect(w, r, downloadUrl(*domain, key), http.StatusTemporaryRedirect)
	return
}

func (h *mainHandler) ServeDefault(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte("405 - Method Not Allowed\n"))
}

//func upload(key string, data io.Reader, size int64) error {
//	c := kodo.New(0, nil)
//	policy := &kodo.PutPolicy{
//		Scope:   BUCKET,
//		Expires: 3600,
//	}
//	token := c.MakeUptoken(policy)
//	uploader := kodocli.NewUploader(0, nil)
//	return uploader.Put(nil, nil, token, key, data, size, nil)
//}

func downloadUrl(domain, key string) string {
	baseUrl := kodo.MakeBaseUrl(domain, key)
	policy := kodo.GetPolicy{}
	c := kodo.New(0, nil)
	return c.MakePrivateUrl(baseUrl, &policy)
}
