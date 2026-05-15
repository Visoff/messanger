package services

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/Visoff/messanger/internal/repository"
	"github.com/Visoff/messanger/pkgs/httperrors"
	"github.com/google/uuid"
)

type WebPushService struct {
	repository      *repository.Queries
	pubsubService   *PubSubService
	vapidPublicKey  string
	vapidPrivateKey string
}

func NewWebPushService(repository *repository.Queries, pubsubService *PubSubService) *WebPushService {
	pubkey := os.Getenv("VAPID_PUBLIC_KEY")
	privatekey := os.Getenv("VAPID_PRIVATE_KEY")
	if pubkey == "" || privatekey == "" {
		privatekey, pubkey, _ = webpush.GenerateVAPIDKeys()
		log.Println("Generated VAPID keys")
		panic("VAPID_PUBLIC_KEY or VAPID_PRIVATE_KEY is not set, here are some: " + pubkey + " : " + privatekey)
	}
	return &WebPushService{
		repository:      repository,
		pubsubService:   pubsubService,
		vapidPublicKey:  pubkey,
		vapidPrivateKey: privatekey,
	}
}

func (s *WebPushService) GetVapidPublicKey() string {
	return s.vapidPublicKey
}

type WebPushSubscriptionDTO struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		Auth string `json:"auth"`
		P256 string `json:"p256dh"`
	} `json:"keys"`
}

func (d *WebPushSubscriptionDTO) Validate() error {
	errors := make(map[string]string)
	if d.Endpoint == "" {
		errors["endpoint"] = "Endpoint is required"
	}
	if d.Keys.Auth == "" {
		errors["auth"] = "Auth is required"
	}
	if d.Keys.P256 == "" {
		errors["p256"] = "P256 is required"
	}
	return httperrors.NewHTTPValidationError(errors)
}

func (s *WebPushService) SaveSubscription(ctx context.Context, subscription *WebPushSubscriptionDTO, user_id uuid.UUID) error {
	return s.repository.CreateWebPushSubscription(ctx, &repository.CreateWebPushSubscriptionParams{
		Endpoint: subscription.Endpoint,
		Auth:     subscription.Keys.Auth,
		P256dh:   subscription.Keys.P256,
		UserID:   user_id,
	})
}

type WebPushNotificationDTO struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (dto *WebPushNotificationDTO) Validate() error {
	errors := make(map[string]string)
	if dto.Title == "" {
		errors["title"] = "Title is required"
	}
	if dto.Body == "" {
		errors["body"] = "Body is required"
	}
	return httperrors.NewHTTPValidationError(errors)
}

func (s *WebPushService) SendNotification(ctx context.Context, dto *WebPushNotificationDTO) error {
	subscriptions, err := s.repository.GetAllSubscriptions(ctx)
	if err != nil {
		return err
	}
	msg, _ := json.Marshal(dto)
	for _, subscription := range subscriptions {
		_, err := webpush.SendNotificationWithContext(ctx, msg, &webpush.Subscription{
			Keys: webpush.Keys{
				Auth:   subscription.Auth,
				P256dh: subscription.P256dh,
			},
			Endpoint: subscription.Endpoint,
		}, &webpush.Options{
			Subscriber:      "mailto:ikalinin01@mail.ru",
			VAPIDPublicKey:  s.vapidPublicKey,
			VAPIDPrivateKey: s.vapidPrivateKey,
			TTL:             60,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *WebPushService) SendUserNotification(ctx context.Context, user_id uuid.UUID, dto *WebPushNotificationDTO) error {
	subscriptions, err := s.repository.GetUserSubscriptions(ctx, user_id)
	if err != nil {
		return err
	}
	msg, _ := json.Marshal(dto)
	for _, subscription := range subscriptions {
		_, err := webpush.SendNotificationWithContext(ctx, msg, &webpush.Subscription{
			Keys: webpush.Keys{
				Auth:   subscription.Auth,
				P256dh: subscription.P256dh,
			},
			Endpoint: subscription.Endpoint,
		}, &webpush.Options{
			Subscriber:      "mailto:ikalinin01@mail.ru",
			VAPIDPublicKey:  s.vapidPublicKey,
			VAPIDPrivateKey: s.vapidPrivateKey,
			TTL:             60,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *WebPushService) SendNewMessageNotification(ctx context.Context, msg *repository.Message, usrs []*repository.User, me *repository.User) error {
	user_id := me.ID

	display_msg := *msg.Content
	if len(display_msg) > 50 {
		display_msg = display_msg[:50] + "..."
	}
	for _, usr := range usrs {
		if usr.ID == user_id {
			continue
		}
		num_subs, err := s.pubsubService.NumberOfSubscribers(ctx, usr.ID.String())
		if err != nil {
			continue
		}
		if num_subs > 0 {
			s.pubsubService.Publish(ctx, usr.ID.String(), msg)
			continue
		}
		err = s.SendUserNotification(ctx, usr.ID, &WebPushNotificationDTO{
			Title: "New message from @" + me.Username,
			Body:  display_msg,
		})
		if err != nil {
			return nil
		}
	}

	return nil
}
