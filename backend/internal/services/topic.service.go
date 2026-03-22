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

func (s *TopicService) CreateMessage(ctx context.Context, topic_id uuid.UUID, dto *CreateMessageDTO) (*repository.Message, error) {
	user_id, err := ExtractUserId(ctx)
	if err != nil {return nil, err}
	return s.repository.CreateTopicMessage(ctx, &repository.CreateTopicMessageParams{
		TopicID: &topic_id,
		SenderID: user_id,
		Content: &dto.Content,
	})
}
