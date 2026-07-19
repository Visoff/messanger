package services

import (
	"context"
	"encoding/json"

	"github.com/Visoff/messanger/internal/repository"
	"github.com/Visoff/messanger/pkgs/httperrors"
	"github.com/google/uuid"
)

func (s *ChatService) InviteUser(ctx context.Context, chat_id uuid.UUID, user_id uuid.UUID) error {
	return s.repository.AddUserToChat(ctx, &repository.AddUserToChatParams{
		ChatID: chat_id,
		UserID: user_id,
		Role:   repository.ChatRoleMember,
	})
}

func (s *ChatService) ListChatMembers(ctx context.Context, chat_id uuid.UUID) ([]*repository.User, error) {
	return s.repository.ListChatMembers(ctx, chat_id)
}

func (s *ChatService) ListChatMembersWithRoles(ctx context.Context, chat_id uuid.UUID) ([]*repository.ListChatMembersWithRolesRow, error) {
	return s.repository.ListChatMembersWithRoles(ctx, chat_id)
}

func (s *ChatService) LeaveChat(ctx context.Context, chat_id uuid.UUID) error {
	user_id, err := ExtractUserId(ctx)
	if err != nil {
		return err
	}
	return s.repository.RemoveUserFromChat(ctx, &repository.RemoveUserFromChatParams{
		UserID: user_id,
		ChatID: chat_id,
	})
}

func (s *ChatService) GetMyRole(ctx context.Context, chat_id uuid.UUID) (*repository.ChatRole, error) {
	user_id, err := ExtractUserId(ctx)
	if err != nil {
		return nil, err
	}
	role, err := s.repository.GetUserChatRole(ctx, &repository.GetUserChatRoleParams{
		UserID: user_id,
		ChatID: chat_id,
	})
	if err != nil {
		return nil, httperrors.NewHTTPNotFoundError("You are not a member of this chat")
	}
	return &role, nil
}

func (s *ChatService) MuteChat(ctx context.Context, chat_id uuid.UUID, muted bool) (*repository.Chat, error) {
	metadata := map[string]interface{}{"muted": muted}
	metaBytes, _ := json.Marshal(metadata)
	return s.repository.UpdateChatMuted(ctx, &repository.UpdateChatMutedParams{
		ID:       chat_id,
		Metadata: metaBytes,
	})
}
