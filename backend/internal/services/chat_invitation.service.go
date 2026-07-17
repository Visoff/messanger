package services

import (
	"context"

	"github.com/Visoff/messanger/internal/repository"
	"github.com/Visoff/messanger/pkgs/httperrors"
	"github.com/google/uuid"
)

type InvitationResponse struct {
	ID uuid.UUID `json:"id"`
}

func (s *ChatService) CreateInvitation(ctx context.Context, chat_id uuid.UUID) (*InvitationResponse, error) {
	user_id, err := ExtractUserId(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := s.repository.GetInvitationByUserAndChat(ctx, &repository.GetInvitationByUserAndChatParams{
		UserID: user_id,
		ChatID: chat_id,
	})
	if err == nil && existing != nil {
		return &InvitationResponse{ID: existing.ID}, nil
	}

	id, err := s.repository.CreateChatInvitation(ctx, &repository.CreateChatInvitationParams{
		ChatID: chat_id,
		UserID: user_id,
	})
	if err != nil {
		return nil, err
	}
	return &InvitationResponse{ID: id}, nil
}

func (s *ChatService) GetInvitation(ctx context.Context, id uuid.UUID) (*repository.Invitation, error) {
	return s.repository.GetInvitationById(ctx, id)
}

func (s *ChatService) UseInvitation(ctx context.Context, id uuid.UUID) error {
	return s.repository.UseInvitation(ctx, id)
}

type InvitationInfo struct {
	Invitation repository.Invitation `json:"invitation"`
	Chat       *repository.Chat      `json:"chat"`
	Creator    *repository.User      `json:"creator"`
}

func (s *ChatService) GetInvitationInfo(ctx context.Context, id uuid.UUID) (*InvitationInfo, error) {
	invitation, err := s.repository.GetInvitationById(ctx, id)
	if err != nil {
		return nil, httperrors.NewHTTPNotFoundError("Invitation not found")
	}

	chat, err := s.repository.GetChat(ctx, invitation.ChatID)
	if err != nil {
		return nil, httperrors.NewHTTPNotFoundError("Chat not found")
	}

	creator, err := s.repository.GetUserById(ctx, invitation.UserID)
	if err != nil {
		return nil, httperrors.NewHTTPNotFoundError("Creator not found")
	}

	if chat.Type == repository.ChatTypePrivate && chat.Title == "" {
		members, err := s.repository.ListChatMembers(ctx, chat.ID)
		if err == nil {
			for _, member := range members {
				if member.ID != invitation.UserID {
					chat.Title = member.Username
					chat.AvatarUrl = member.AvatarUrl
					break
				}
			}
		}
		if chat.Title == "" {
			chat.Title = creator.Username
		}
	}

	return &InvitationInfo{
		Invitation: *invitation,
		Chat:       chat,
		Creator:    creator,
	}, nil
}

func (s *ChatService) AcceptInvitation(ctx context.Context, invitation_id uuid.UUID) (*repository.Chat, error) {
	invitation, err := s.repository.GetInvitationById(ctx, invitation_id)
	if err != nil {
		return nil, httperrors.NewHTTPNotFoundError("Invitation not found")
	}
	user_id, err := ExtractUserId(ctx)
	if err != nil {
		return nil, err
	}
	err = s.repository.AddUserToChat(ctx, &repository.AddUserToChatParams{
		ChatID: invitation.ChatID,
		UserID: user_id,
		Role:   repository.ChatRoleMember,
	})
	if err != nil {
		return nil, err
	}
	s.repository.UseInvitation(ctx, invitation_id)
	return s.repository.GetChat(ctx, invitation.ChatID)
}
