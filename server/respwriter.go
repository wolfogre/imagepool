package main

import (
	"net/http"
	"time"
)

type ResponseWriter struct{
	w http.ResponseWriter
	start time.Time
	StatusCode int
}

func NewResponseWriter(writer http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		w: writer,
		start: time.Now(),
	}
}

func (rw ResponseWriter) Header() http.Header {
	return rw.w.Header()
}

func (rw ResponseWriter) Write(buf []byte) (int, error) {
	return rw.w.Write(buf)
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.w.WriteHeader(code)
}

func (rw *ResponseWriter) CostTime() time.Duration{
	return time.Now().Sub(rw.start)
}

