package controllers

import (
	"net/http"
	"time"

	"github.com/Visoff/messanger/internal/services"
	"github.com/Visoff/messanger/pkgs/dtos"
	"github.com/Visoff/messanger/pkgs/handlers"
)

type PubSubController struct {
	pubsubService  *services.PubSubService
	webpushService *services.WebPushService
	mux            *http.ServeMux
}

func NewPubSubController(pubsubService *services.PubSubService, webpushService *services.WebPushService, authService *services.AuthService) *PubSubController {
	c := &PubSubController{
		pubsubService:  pubsubService,
		webpushService: webpushService,
		mux:            nil,
	}

	mux := http.NewServeMux()
	c.mux = mux

	mux.Handle("GET /sse", handlers.Handler(c.SSE))
	mux.Handle("GET /push/pubkey", handlers.Handler(c.GetPushPubKey))
	mux.Handle("POST /push/subscribe", authService.ProtectRoute(handlers.Handler(c.SubscribePush)))
	mux.Handle("POST /push/notify", handlers.Handler(c.NotifyPush))

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

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return nil
		case <-ticker.C:
			w.Write([]byte(":ping\n\n"))
			flusher.Flush()
		case msg := <-ch:
			w.Write([]byte("data: "))
			w.Write([]byte(msg.Payload))
			w.Write([]byte("\n\n"))
			flusher.Flush()
		}
	}
}

func (c *PubSubController) GetPushPubKey(w http.ResponseWriter, r *http.Request) error {
	key := c.webpushService.GetVapidPublicKey()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(key))
	return nil
}

func (c *PubSubController) SubscribePush(w http.ResponseWriter, r *http.Request) error {
	user_id, err := services.ExtractUserId(r.Context())
	if err != nil {
		return err
	}
	var dto services.WebPushSubscriptionDTO
	if err := dtos.ParseFromBody(r, &dto); err != nil {
		return err
	}
	err = c.webpushService.SaveSubscription(r.Context(), &dto, user_id)
	if err != nil {
		return err
	}
	return nil
}

func (c *PubSubController) NotifyPush(w http.ResponseWriter, r *http.Request) error {
	var dto services.WebPushNotificationDTO
	if err := dtos.ParseFromBody(r, &dto); err != nil {
		return err
	}
	err := c.webpushService.SendNotification(r.Context(), &dto)
	if err != nil {
		return err
	}
	return nil
}
