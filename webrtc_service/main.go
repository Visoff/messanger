package main

import (
	"log"
	"net/http"
	"os"
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	service := NewWebRTCService()

	ws_updater := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	controller := NewWebRTCController(&ws_updater, service)
	http.Handle("/", controller)
	err = http.ListenAndServe(":" + port, nil)
	if err != nil {
		log.Println(err)
	}
}
