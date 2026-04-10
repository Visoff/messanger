package services

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

type PubSubService struct {
	redis redis.Client
}

func NewPubSubService() (*PubSubService, error) {
	s := &PubSubService{
		redis: *redis.NewClient(&redis.Options{
			Addr: os.Getenv("REDIS_URL"),
		}),
	}

	if err := s.redis.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	log.Println("Connected to Redis")

	return s, nil
}

func (s *PubSubService) Subscribe(ctx context.Context, topic string) (<-chan *redis.Message, error) {
	pubsub := s.redis.Subscribe(ctx, topic)
	return pubsub.Channel(), nil
}

func (s *PubSubService) Publish(ctx context.Context, topic string, input any) error {
	message, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return s.redis.Publish(ctx, topic, message).Err()
}
