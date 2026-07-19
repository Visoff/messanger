package services

import (
	"context"
	"encoding/json"
	"sort"

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

type ChatWithLastMessage struct {
	*repository.Chat
	LastMessage *repository.Message `json:"last_message"`
}

func (s *ChatService) ListChats(ctx context.Context) ([]*ChatWithLastMessage, error) {
	user_id, err := ExtractUserId(ctx)
	if err != nil {
		return []*ChatWithLastMessage{}, err
	}
	chats, err := s.repository.ListUserChats(ctx, user_id)
	if err != nil {
		return []*ChatWithLastMessage{}, nil
	}

	chatIDs := make([]uuid.UUID, 0, len(chats))
	for _, chat := range chats {
		chatIDs = append(chatIDs, chat.ID)
	}

	lastMessages := make(map[uuid.UUID]*repository.Message)
	if len(chatIDs) > 0 {
		msgs, err := s.repository.ListChatsLastMessages(ctx, chatIDs)
		if err == nil {
			for _, msg := range msgs {
				lastMessages[msg.ChatID] = msg
			}
		}
	}

	result := make([]*ChatWithLastMessage, 0, len(chats))
	for _, chat := range chats {
		if chat.Type == repository.ChatTypePrivate && chat.Title == "" {
			members, err := s.repository.ListChatMembers(ctx, chat.ID)
			if err == nil {
				for _, member := range members {
					if member.ID != user_id {
						chat.Title = member.Username
						chat.AvatarUrl = member.AvatarUrl
						break
					}
				}
			}
		}

		item := &ChatWithLastMessage{Chat: chat}
		msg, err := s.repository.GetChatLastMessage(ctx, chat.ID)
		if err == nil && msg != nil {
			item.LastMessage = msg
		}

		result = append(result, item)
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].LastMessage != nil && result[j].LastMessage != nil {
			return result[i].LastMessage.CreatedAt.After(result[j].LastMessage.CreatedAt)
		}
		if result[i].LastMessage != nil {
			return true
		}
		if result[j].LastMessage != nil {
			return false
		}
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result, nil
}

type CreateChatDTO struct {
	Title string `json:"title" example:"General Chat"`
}

func (dto *CreateChatDTO) Validate() error {
	errors := make(map[string]string)
	if dto.Title == "" {
		errors["title"] = "Title is required"
	}
	if len(errors) > 0 {
		return httperrors.NewHTTPValidationError(errors)
	}
	return nil
}

func (s *ChatService) CreateChat(ctx context.Context, dto *CreateChatDTO) (*repository.Chat, error) {
	return s.createChatWithType(ctx, dto.Title, repository.ChatTypeGroup)
}

func (s *ChatService) CreateChannel(ctx context.Context, dto *CreateChatDTO) (*repository.Chat, error) {
	return s.createChatWithType(ctx, dto.Title, repository.ChatTypeChannel)
}

func (s *ChatService) createChatWithType(ctx context.Context, title string, chatType repository.ChatType) (*repository.Chat, error) {
	user_id, err := ExtractUserId(ctx)
	if err != nil {
		return nil, err
	}
	qtx, tx, err := s.repository.NewTx(ctx)
	defer tx.Rollback(ctx)
	if err != nil {
		return nil, err
	}
	chat, err := qtx.CreateChat(ctx, &repository.CreateChatParams{
		Title: title,
		Type:  chatType,
	})
	if err != nil {
		return nil, err
	}
	err = qtx.AddUserToChat(ctx, &repository.AddUserToChatParams{
		ChatID: chat.ID,
		UserID: user_id,
		Role:   repository.ChatRoleOwner,
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

func (s *ChatService) GetChat(ctx context.Context, chat_id uuid.UUID) (*repository.Chat, error) {
	chat, err := s.repository.GetChat(ctx, chat_id)
	if err != nil {
		return nil, err
	}
	if chat.Type == repository.ChatTypePrivate && chat.Title == "" {
		user_id, err := ExtractUserId(ctx)
		if err != nil {
			return chat, nil
		}
		members, err := s.repository.ListChatMembers(ctx, chat.ID)
		if err == nil {
			for _, member := range members {
				if member.ID != user_id {
					chat.Title = member.Username
					chat.AvatarUrl = member.AvatarUrl
					break
				}
			}
		}
	}
	return chat, nil
}

func (s *ChatService) CreatePrivateChat(ctx context.Context, user1_id, user2_id uuid.UUID) (*repository.Chat, error) {
	existing, err := s.repository.CheckPrivateChatExists(ctx, &repository.CheckPrivateChatExistsParams{UserID: user1_id, UserID_2: user2_id})
	if err == nil && existing != uuid.Nil {
		chat, err := s.repository.GetChat(ctx, existing)
		if err == nil {
			return chat, httperrors.NewHTTPConflictError("Private chat already exists")
		}
	}

	chat, err := s.repository.CreateChat(ctx, &repository.CreateChatParams{
		Title: "",
		Type:  repository.ChatTypePrivate,
	})
	if err != nil {
		return nil, err
	}
	err = s.repository.AddUserToChat(ctx, &repository.AddUserToChatParams{
		ChatID: chat.ID,
		UserID: user1_id,
		Role:   repository.ChatRoleAdmin,
	})
	if err != nil {
		return nil, err
	}
	err = s.repository.AddUserToChat(ctx, &repository.AddUserToChatParams{
		ChatID: chat.ID,
		UserID: user2_id,
		Role:   repository.ChatRoleAdmin,
	})
	if err != nil {
		return nil, err
	}
	return chat, nil
}

type UpdateChatDTO struct {
	Title    string          `json:"title"`
	Metadata json.RawMessage `json:"metadata"`
}

func (dto *UpdateChatDTO) Validate() error {
	return nil
}

func (s *ChatService) UpdateChat(ctx context.Context, chat_id uuid.UUID, dto *UpdateChatDTO) (*repository.Chat, error) {
	metadata := []byte("{}")
	if dto.Metadata != nil {
		metadata = dto.Metadata
	}
	return s.repository.UpdateChat(ctx, &repository.UpdateChatParams{
		ID:       chat_id,
		Title:    dto.Title,
		Metadata: metadata,
	})
}

func (s *ChatService) UpdateChatAvatar(ctx context.Context, chat_id uuid.UUID, avatarUrl string) (*repository.Chat, error) {
	user_id, err := ExtractUserId(ctx)
	if err != nil {
		return nil, err
	}
	members, err := s.repository.ListChatMembers(ctx, chat_id)
	if err != nil {
		return nil, httperrors.NewHTTPNotFoundError("Chat not found")
	}
	isMember := false
	for _, m := range members {
		if m.ID == user_id {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, httperrors.NewHTTPForbiddenError("You are not a member of this chat")
	}
	return s.repository.UpdateChatAvatar(ctx, &repository.UpdateChatAvatarParams{
		ID:        chat_id,
		AvatarUrl: &avatarUrl,
	})
}
