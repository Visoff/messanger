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
	if err != nil {return []*ChatWithLastMessage{}, err}
	chats, err := s.repository.ListChats(ctx, user_id)
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
						break
					}
				}
			}
		}

		item := &ChatWithLastMessage{Chat: chat}
		if msg, ok := lastMessages[chat.ID]; ok {
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
	user_id, err := ExtractUserId(ctx)
	if err != nil {return nil, err}
	qtx, tx, err := s.repository.NewTx(ctx)
	defer tx.Rollback(ctx)
	if err != nil {
		return nil, err
	}
	chat, err := qtx.CreateChat(ctx, &repository.CreateChatParams{
		Title: dto.Title,
		Type: repository.ChatTypeGroup,
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
					break
				}
			}
		}
	}
	return chat, nil
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
		Title: dto.Title,
		Type: repository.TopicType(dto.Type),
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
	if err != nil {return nil, err}
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

func (s *ChatService) InviteUser(ctx context.Context, chat_id uuid.UUID, user_id uuid.UUID) error {
	return s.repository.AddUserToChat(ctx, &repository.AddUserToChatParams{
		ChatID: chat_id,
		UserID: user_id,
		Role: repository.ChatRoleMember,
	})
}

type InvitationResponse struct {
	ID uuid.UUID `json:"id"`
}

func (s *ChatService) CreateInvitation(ctx context.Context, chat_id uuid.UUID) (*InvitationResponse, error) {
	user_id, err := ExtractUserId(ctx)
	if err != nil {return nil, err}

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
	if err != nil {return nil, err}
	return &InvitationResponse{ID: id}, nil
}

func (s *ChatService) GetInvitation(ctx context.Context, id uuid.UUID) (*repository.Invitation, error) {
	return s.repository.GetInvitationById(ctx, id)
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

func (s *ChatService) ListChatMembers(ctx context.Context, chat_id uuid.UUID) ([]*repository.User, error) {
	return s.repository.ListChatMembers(ctx, chat_id)
}

func (s *ChatService) CreatePrivateChat(ctx context.Context, user1_id, user2_id uuid.UUID) (*repository.Chat, error) {
	existing, err := s.repository.CheckPrivateChatExists(ctx, &repository.CheckPrivateChatExistsParams{UserID: user1_id, UserID_2: user2_id })
	if err == nil && existing != uuid.Nil {
		chat, err := s.repository.GetChat(ctx, existing)
		if err == nil {
			return chat, httperrors.NewHTTPConflictError("Private chat already exists")
		}
	}

	chat, err := s.repository.CreateChat(ctx, &repository.CreateChatParams{
		Title: "",
		Type: repository.ChatTypePrivate,
	})
	if err != nil {
		return nil, err
	}
	err = s.repository.AddUserToChat(ctx, &repository.AddUserToChatParams{
		ChatID: chat.ID,
		UserID: user1_id,
		Role: repository.ChatRoleAdmin,
	})
	if err != nil {
		return nil, err
	}
	err = s.repository.AddUserToChat(ctx, &repository.AddUserToChatParams{
		ChatID: chat.ID,
		UserID: user2_id,
		Role: repository.ChatRoleAdmin,
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

func (s *ChatService) MuteChat(ctx context.Context, chat_id uuid.UUID, muted bool) (*repository.Chat, error) {
	metadata := map[string]interface{}{"muted": muted}
	metaBytes, _ := json.Marshal(metadata)
	return s.repository.UpdateChatMuted(ctx, &repository.UpdateChatMutedParams{
		ID:       chat_id,
		Metadata: metaBytes,
	})
}
