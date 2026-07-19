package services

import (
	"context"

	"github.com/Visoff/messanger/internal/repository"
	"github.com/Visoff/messanger/pkgs/httperrors"
	"github.com/google/uuid"
)

func (s *ChatService) ListTopics(ctx context.Context, chat_id uuid.UUID) ([]*repository.Topic, error) {
	return s.repository.ListChatTopics(ctx, chat_id)
}

func (s *ChatService) ListMessages(ctx context.Context, chat_id uuid.UUID) ([]*repository.Message, error) {
	return s.repository.ListChatMessages(ctx, chat_id)
}

type CreateTopicDTO struct {
	Title string `json:"title" example:"General Chat"`
	Type  string `json:"type"  example:"group" enums:"text_topic,voice_topic"`
}

func (dto *CreateTopicDTO) Validate() error {
	errors := make(map[string]string)
	if dto.Title == "" {
		errors["title"] = "Title is required"
	}
	if dto.Type == "" {
		errors["type"] = "Type is required"
	}
	if dto.Type != "text_topic" && dto.Type != "voice_topic" {
		errors["type"] = "Invalid type"
	}
	if len(errors) > 0 {
		return httperrors.NewHTTPValidationError(errors)
	}
	return nil
}

func (s *ChatService) CreateTopic(ctx context.Context, chat_id uuid.UUID, dto *CreateTopicDTO) (*repository.Topic, error) {
	return s.repository.CreateChatTopic(ctx, &repository.CreateChatTopicParams{
		ChatID: chat_id,
		Title:  dto.Title,
		Type:   repository.TopicType(dto.Type),
	})
}

type CreateMessageDTO struct {
	Content        string     `json:"content" example:"Hello, world!"`
	ReplyMessageID *uuid.UUID `json:"reply_message_id"`
}

func (dto *CreateMessageDTO) Validate() error {
	errors := make(map[string]string)
	if dto.Content == "" {
		errors["content"] = "Content is required"
	}
	if len(errors) > 0 {
		return httperrors.NewHTTPValidationError(errors)
	}
	return nil
}

func (s *ChatService) CreateMessage(ctx context.Context, chat_id uuid.UUID, dto *CreateMessageDTO) (*repository.Message, error) {
	user_id, err := ExtractUserId(ctx)
	if err != nil {
		return nil, err
	}

	chat, err := s.repository.GetChat(ctx, chat_id)
	if err != nil {
		return nil, httperrors.NewHTTPNotFoundError("Chat not found")
	}
	if chat.Type == repository.ChatTypeChannel {
		role, err := s.repository.GetUserChatRole(ctx, &repository.GetUserChatRoleParams{
			UserID: user_id,
			ChatID: chat_id,
		})
		if err != nil {
			return nil, httperrors.NewHTTPForbiddenError("You are not a member of this channel")
		}
		if role != repository.ChatRoleOwner && role != repository.ChatRoleAdmin {
			return nil, httperrors.NewHTTPForbiddenError("Only admins can send messages in channels")
		}
	}

	msg, err := s.repository.CreateChatMessage(ctx, &repository.CreateChatMessageParams{
		ChatID:         chat_id,
		SenderID:       user_id,
		Content:        &dto.Content,
		ReplyMessageID: dto.ReplyMessageID,
	})
	if err != nil {
		return nil, err
	}
	s.repository.UpdateChatUpdatedAt(ctx, chat_id)
	return msg, nil
}
