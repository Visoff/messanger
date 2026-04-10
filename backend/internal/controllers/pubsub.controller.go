package controllers

import (
	"log"
	"net/http"
	"os"

	"github.com/Visoff/messanger/internal/services"
	"github.com/Visoff/messanger/pkgs/handlers"
)

type PubSubController struct {
	pubsubService *services.PubSubService
	mux           *http.ServeMux
}

func NewPubSubController(pubsubService *services.PubSubService, authService *services.AuthService) *PubSubController {
	c := &PubSubController{
		pubsubService: pubsubService,
		mux: nil,
	}

	mux := http.NewServeMux()
	c.mux = mux

	//mux.Handle("GET /sse", authService.ProtectRoute(handlers.Handler(c.SSE)))
	mux.Handle("GET /sse", handlers.Handler(c.SSE))
	mux.Handle("GET /push/pubkey", handlers.Handler(c.GetPushPubKey))

	return c
}

func (c *PubSubController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mux.ServeHTTP(w, r)
}

func (c *PubSubController) SSE(w http.ResponseWriter, r *http.Request) error {
	ch, err := c.pubsubService.Subscribe(r.Context(), "messanger")
	if err != nil {
		return err
	}

	flusher := w.(http.Flusher)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Accel-Buffering", "no")

	w.Write([]byte(":ok\n\n"))
	flusher.Flush()

	for msg := range ch {
		w.Write([]byte("data: "))
		w.Write([]byte(msg.Payload))
		w.Write([]byte("\n\n"))
		flusher.Flush()
	}
	return nil
}

func (c *PubSubController) GetPushPubKey(w http.ResponseWriter, r *http.Request) error {
	key, ok := os.LookupEnv("PUSH_PUBKEY")
	if !ok {
		log.Println("PUSH_PUBKEY is not set")
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(key))
	return nil
}
