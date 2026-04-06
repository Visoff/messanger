package handlers

import (
	"bufio"
	"log"
	"net"
	"net/http"
)

func MiddlewareChain(middlewares ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

func AllowCors(handler Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return nil
		}
		return handler(w, r)
	}
}

type req_with_status struct {
	http.ResponseWriter
	Status int
}

func (r *req_with_status) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *req_with_status) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := r.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, nil
}

func Logging(logger *log.Logger) Middleware {
	return func(next Handler) Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			rs := &req_with_status{w, http.StatusOK}
			err := next(rs, r)
			logger.Println(r.Method, r.RequestURI, rs.Status)
			return err
		}
	}
}
