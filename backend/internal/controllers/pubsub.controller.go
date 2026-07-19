package controllers

import (
	"context"
	"net/http"
	"strconv"
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

	mux.Handle("GET /sse", authService.ProtectRoute(handlers.Handler(c.SSE)))
	mux.Handle("GET /poll", authService.ProtectRoute(handlers.Handler(c.Poll)))
	mux.Handle("GET /push/pubkey", handlers.Handler(c.GetPushPubKey))
	mux.Handle("POST /push/subscribe", authService.ProtectRoute(handlers.Handler(c.SubscribePush)))
	mux.Handle("POST /push/notify", handlers.Handler(c.NotifyPush))

	return c
}

func (c *PubSubController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mux.ServeHTTP(w, r)
}

// SSE subscribes to Server-Sent Events for real-time updates.
// @Summary      SSE stream
// @Description  Subscribe to Server-Sent Events for real-time chat updates.
// @Tags         pubsub
// @Produce      text/event-stream
// @Success      200
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /pubsub/sse [get]
// @Security     BearerAuth
func (c *PubSubController) SSE(w http.ResponseWriter, r *http.Request) error {
	user_id, err := services.ExtractUserId(r.Context())
	if err != nil {
		return err
	}
	ch, err := c.pubsubService.Subscribe(r.Context(), user_id.String())
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

// Poll is a long-polling fallback for SSE. Blocks up to `timeout` seconds.
func (c *PubSubController) Poll(w http.ResponseWriter, r *http.Request) error {
	user_id, err := services.ExtractUserId(r.Context())
	if err != nil {
		return err
	}

	timeout := 30
	if t := r.URL.Query().Get("timeout"); t != "" {
		if v, err := strconv.Atoi(t); err == nil && v > 0 && v <= 60 {
			timeout = v
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeout)*time.Second)
	defer cancel()

	ch, err := c.pubsubService.Subscribe(ctx, user_id.String())
	if err != nil {
		return err
	}

	flusher := w.(http.Flusher)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")

	select {
	case <-ctx.Done():
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
		return nil
	case msg := <-ch:
		w.WriteHeader(http.StatusOK)
		payload := msg.Payload
		w.Write([]byte("[" + payload + "]"))
		flusher.Flush()
		return nil
	}
}

// GetPushPubKey returns the VAPID public key for Web Push.
// @Summary      Get VAPID public key
// @Description  Returns the VAPID public key needed for Web Push subscription.
// @Tags         pubsub
// @Produce      plain
// @Success      200  {string}  string
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /pubsub/push/pubkey [get]
func (c *PubSubController) GetPushPubKey(w http.ResponseWriter, r *http.Request) error {
	key := c.webpushService.GetVapidPublicKey()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(key))
	return nil
}

// SubscribePush subscribes the authenticated user to Web Push notifications.
// @Summary      Subscribe to push notifications
// @Description  Save a Web Push subscription for the authenticated user.
// @Tags         pubsub
// @Accept       json
// @Produce      json
// @Param        request body services.WebPushSubscriptionDTO true "Push subscription details"
// @Success      200
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /pubsub/push/subscribe [post]
// @Security     BearerAuth
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

// NotifyPush sends a push notification to all subscribers.
// @Summary      Send push notification
// @Description  Send a push notification to all subscribed users.
// @Tags         pubsub
// @Accept       json
// @Produce      json
// @Param        request body services.WebPushNotificationDTO true "Notification details"
// @Success      200
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /pubsub/push/notify [post]
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
