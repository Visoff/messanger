package services

import (
	"context"

	"github.com/Visoff/messanger/internal/repository"
	"github.com/Visoff/messanger/pkgs/httperrors"
	"github.com/google/uuid"
)

type ChatService struct {
	repository *repository.Queries
}

func NewChatService(repository *repository.Queries) *ChatService {
	return &ChatService{repository: repository}
}

func (s *ChatService) ListChats(ctx context.Context) ([]*repository.Chat, error) {
	user_id, err := ExtractUserId(ctx)
	if err != nil {return []*repository.Chat{}, err}
	list, err := s.repository.ListChats(ctx, user_id)
	if err != nil {
		return []*repository.Chat{}, nil
	}
	return list, nil
}

type CreateChatDTO struct {
	Title string `json:"title" example:"General Chat"`
	Type  string `json:"type"  example:"group" enums:"private,group,channel"`
}

func (dto *CreateChatDTO) Validate() error {
	errors := make(map[string]string)
	if dto.Title == "" {
		errors["title"] = "Title is required"
	}
	if dto.Type == "" {
		errors["type"] = "Type is required"
	}
	if dto.Type != "private" && dto.Type != "group" && dto.Type != "channel" {
		errors["type"] = "Invalid type"
	}
	if len(errors) > 0 {
		return httperrors.NewHTTPValidationError(errors)
	}
	return nil
}

func (s *ChatService) CreateChat(ctx context.Context, dto *CreateChatDTO) (*repository.Chat, error) {
	user_id, err := ExtractUserId(ctx)
	if err != nil {return nil, err}
	qtx, tx, err := s.repository.NewTx(ctx)
	defer tx.Rollback(ctx)
	if err != nil {
		return nil, err
	}
	chat, err := qtx.CreateChat(ctx, &repository.CreateChatParams{
		Title: dto.Title,
		Type: repository.ChatType(dto.Type),
	})
	if err != nil {
		return nil, err
	}
	err = qtx.AddUserToChat(ctx, &repository.AddUserToChatParams{
		ChatID: chat.ID,
		UserID: user_id,
		Role: repository.ChatRoleOwner,
	})
	if err != nil {
		return nil, err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}
	return chat, nil
}

func (s *ChatService) ListTopics(ctx context.Context, chat_id uuid.UUID) ([]*repository.Topic, error) {
	return s.repository.ListChatTopics(ctx, chat_id)
}

func (s *ChatService) ListMessages(ctx context.Context, chat_id uuid.UUID) ([]*repository.Message, error) {
	return s.repository.ListChatMessages(ctx, chat_id)
}
