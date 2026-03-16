package services

import (
	"context"

	"github.com/Visoff/messanger/internal/repository"
	"github.com/Visoff/messanger/pkgs/httperrors"
)

type ChatService struct {
	repository *repository.Queries
}

func NewChatService(repository *repository.Queries) *ChatService {
	return &ChatService{repository: repository}
}

func (s *ChatService) ListChats(ctx context.Context) []*repository.Chat {
	list, err := s.repository.ListChats(ctx)
	if err != nil {
		return []*repository.Chat{}
	}
	return list
}

type CreateChatDTO struct {
	Title string `json:"title"`
	Type  string `json:"type"`
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
	return s.repository.CreateChat(ctx, &repository.CreateChatParams{
		Title: dto.Title,
		Type: repository.ChatType(dto.Type),
	})
}
