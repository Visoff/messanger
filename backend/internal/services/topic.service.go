package services

import (
	"context"

	"github.com/Visoff/messanger/internal/repository"
	"github.com/google/uuid"
)

type TopicService struct {
	repository *repository.Queries
}

func NewTopicService(repository *repository.Queries) *TopicService {
	return &TopicService{repository: repository}
}

func (s *TopicService) ListMessages(ctx context.Context, topic_id uuid.UUID) ([]*repository.Message, error) {
	return s.repository.ListTopicMessages(ctx, &topic_id)
}
