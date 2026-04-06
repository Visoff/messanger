package main

import (
	"log"
	"net/http"
	"runtime"

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

	service := NewWebRTCService()

	ws_updater := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	controller := NewWebRTCController(&ws_updater, service)
	http.Handle("/", controller)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println(err)
	}
}
