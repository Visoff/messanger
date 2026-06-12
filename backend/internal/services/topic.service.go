package services

import (
	"context"

	"github.com/Visoff/messanger/internal/repository"
	"github.com/Visoff/messanger/pkgs/httperrors"
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

	topic, err := s.repository.GetTopic(ctx, topic_id)
	if err != nil {
		return nil, httperrors.NewHTTPNotFoundError("Topic not found")
	}

	chat, err := s.repository.GetChat(ctx, topic.ChatID)
	if err != nil {
		return nil, httperrors.NewHTTPNotFoundError("Chat not found")
	}
	if chat.Type == repository.ChatTypeChannel {
		role, err := s.repository.GetUserChatRole(ctx, &repository.GetUserChatRoleParams{
			UserID: user_id,
			ChatID: topic.ChatID,
		})
		if err != nil {
			return nil, httperrors.NewHTTPForbiddenError("You are not a member of this channel")
		}
		if role != repository.ChatRoleOwner && role != repository.ChatRoleAdmin {
			return nil, httperrors.NewHTTPForbiddenError("Only admins can send messages in channels")
		}
	}

	return s.repository.CreateTopicMessage(ctx, &repository.CreateTopicMessageParams{
		TopicID:        &topic_id,
		SenderID:       user_id,
		Content:        &dto.Content,
		ReplyMessageID: dto.ReplyMessageID,
	})
}

func (s *TopicService) GetTopic(ctx context.Context, topic_id uuid.UUID) (*repository.Topic, error) {
	return s.repository.GetTopic(ctx, topic_id)
}

func (s *TopicService) ListTopicMembers(ctx context.Context, topic_id uuid.UUID) ([]*repository.User, error) {
	return s.repository.ListTopicMembers(ctx, topic_id)
}
