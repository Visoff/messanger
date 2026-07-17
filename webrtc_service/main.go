package main

import (
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	runtime.SetMutexProfileFraction(1)
	err := godotenv.Load()
	if err != nil {
		log.Println(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	var originCheck func(r *http.Request) bool
	if allowedOrigins == "" || allowedOrigins == "*" {
		originCheck = func(r *http.Request) bool { return true }
	} else {
		origins := make(map[string]bool)
		for _, o := range strings.Split(allowedOrigins, ",") {
			origins[strings.TrimSpace(o)] = true
		}
		originCheck = func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return origin == "" || origins[origin]
		}
	}

	service := NewWebRTCService()

	ws_updater := websocket.Upgrader{
		CheckOrigin: originCheck,
	}

	controller := NewWebRTCController(&ws_updater, service)
	http.Handle("/", controller)
	log.Println("Listening on port " + port)
	err = http.ListenAndServe(":" + port, nil)
	if err != nil {
		log.Println(err)
	}
}
