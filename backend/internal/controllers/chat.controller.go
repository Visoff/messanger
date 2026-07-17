package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Visoff/messanger/internal/services"
	"github.com/Visoff/messanger/pkgs/dtos"
	"github.com/Visoff/messanger/pkgs/handlers"
	"github.com/Visoff/messanger/pkgs/httperrors"
	"github.com/google/uuid"
)

type ChatController struct {
	chatService          *services.ChatService
	userService          *services.UserService
	pubsubService        *services.PubSubService
	webpushService       *services.WebPushService
	mux                  *http.ServeMux
	fileStorageUrl       string
	publicFileStorageUrl string
}

func NewChatController(chatService *services.ChatService, userService *services.UserService, pubsubService *services.PubSubService, webpushService *services.WebPushService, authService *services.AuthService, fileStorageUrl string, publicFileStorageUrl string) *ChatController {
	c := &ChatController{
		chatService:          chatService,
		pubsubService:        pubsubService,
		userService:          userService,
		webpushService:       webpushService,
		mux:                  nil,
		fileStorageUrl:       fileStorageUrl,
		publicFileStorageUrl: publicFileStorageUrl,
	}

	mux := http.NewServeMux()
	c.mux = mux

	mux.Handle("GET /", authService.ProtectRoute(handlers.Handler(c.ListChats)))
	mux.Handle("POST /group", authService.ProtectRoute(handlers.Handler(c.CreateChat)))
	mux.Handle("POST /channel", authService.ProtectRoute(handlers.Handler(c.CreateChannel)))
	mux.Handle("POST /private", authService.ProtectRoute(handlers.Handler(c.CreatePrivateChat)))

	mux.Handle("GET /{id}", authService.ProtectRoute(handlers.Handler(c.GetChat)))

	/*
		mux.Handle("PUT /{id}", handlers.Handler(c.UpdateChat))
		mux.Handle("DELETE /{id}", handlers.Handler(c.DeleteChat))
	*/

	mux.Handle("GET /{id}/topics", authService.ProtectRoute(handlers.Handler(c.ListTopics)))
	mux.Handle("POST /{id}/topics", authService.ProtectRoute(handlers.Handler(c.CreateTopic)))

	mux.Handle("GET /{id}/messages", authService.ProtectRoute(handlers.Handler(c.ListMessages)))
	mux.Handle("POST /{id}/messages", authService.ProtectRoute(handlers.Handler(c.CreateMessage)))

	mux.Handle("POST /{id}/invite/{user_id}", authService.ProtectRoute(handlers.Handler(c.InviteUser)))
	mux.Handle("POST /{id}/invitation", authService.ProtectRoute(handlers.Handler(c.CreateInvitation)))
	mux.Handle("PUT /{id}", authService.ProtectRoute(handlers.Handler(c.UpdateChat)))
	mux.Handle("POST /{id}/avatar", authService.ProtectRoute(handlers.Handler(c.UploadChatAvatar)))
	mux.Handle("POST /{id}/leave", authService.ProtectRoute(handlers.Handler(c.LeaveChat)))
	mux.Handle("PUT /{id}/mute", authService.ProtectRoute(handlers.Handler(c.MuteChat)))
	mux.Handle("GET /{id}/members", authService.ProtectRoute(handlers.Handler(c.ListChatMembers)))
	mux.Handle("GET /{id}/my-role", authService.ProtectRoute(handlers.Handler(c.GetMyRole)))

	return c
}

func (c *ChatController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mux.ServeHTTP(w, r)
}

// ListChats returns a list of all chats with their last message.
// @Summary      List all chats
// @Description  Returns a list of all chats with their most recent message.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Success      200  {object}  []services.ChatWithLastMessage
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /chats/ [get]
// @Security     BearerAuth
func (c *ChatController) ListChats(w http.ResponseWriter, r *http.Request) error {
	chats, err := c.chatService.ListChats(r.Context())
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chats)
	return nil
}

// CreateChat creates a new group chat and adds the authenticated user as owner.
// @Summary      Create a group chat
// @Description  Create a new group chat. The authenticated user becomes the owner.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        request body services.CreateChatDTO true "Chat details"
// @Success      200  {object}  repository.Chat
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /chats/group [post]
// @Security     BearerAuth
func (c *ChatController) CreateChat(w http.ResponseWriter, r *http.Request) error {
	var dto services.CreateChatDTO

	if err := dtos.ParseFromBody(r, &dto); err != nil {
		return err
	}

	chat, err := c.chatService.CreateChat(r.Context(), &dto)
	if err != nil {
		return err
	}

	go c.pubsubService.Publish(context.Background(), chat.ID.String(), map[string]interface{}{
		"type": "chat_created",
		"chat": chat,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chat)
	return nil
}

// CreateChannel creates a new channel and adds the authenticated user as owner.
// @Summary      Create a channel
// @Description  Create a new channel. The authenticated user becomes the owner.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        request body services.CreateChatDTO true "Channel details"
// @Success      200  {object}  repository.Chat
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /chats/channel [post]
// @Security     BearerAuth
func (c *ChatController) CreateChannel(w http.ResponseWriter, r *http.Request) error {
	var dto services.CreateChatDTO

	if err := dtos.ParseFromBody(r, &dto); err != nil {
		return err
	}

	chat, err := c.chatService.CreateChannel(r.Context(), &dto)
	if err != nil {
		return err
	}

	go c.pubsubService.Publish(context.Background(), chat.ID.String(), map[string]interface{}{
		"type": "chat_created",
		"chat": chat,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chat)
	return nil
}

// ListTopics returns a list of all topics in a chat.
// @Summary      List all topics in a chat
// @Description  Returns a list of all topics in a chat.
// @Tags         topics
// @Accept       json
// @Produce      json
// @Param        id path string true "Chat ID"
// @Success      200  {object}  []repository.Topic
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /chats/{id}/topics [get]
// @Security     BearerAuth
func (c *ChatController) ListTopics(w http.ResponseWriter, r *http.Request) error {
	chat_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}
	topics, err := c.chatService.ListTopics(r.Context(), chat_id)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(topics)
	return nil
}

// ListMessages returns a list of all messages in a chat.
// @Summary      List all messages in a chat
// @Description  Returns a list of all messages in a chat.
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        id path string true "Chat ID"
// @Success      200  {object}  []repository.Message
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /chats/{id}/messages [get]
// @Security     BearerAuth
func (c *ChatController) ListMessages(w http.ResponseWriter, r *http.Request) error {
	chat_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}
	messages, err := c.chatService.ListMessages(r.Context(), chat_id)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(messages)
	return nil
}

// GetChat returns a chat by ID.
// @Summary      Get a chat by ID
// @Description  Returns a chat by ID.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        id path string true "Chat ID"
// @Success      200  {object}  repository.Chat
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /chats/{id} [get]
// @Security     BearerAuth
func (c *ChatController) GetChat(w http.ResponseWriter, r *http.Request) error {
	chat_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}
	chat, err := c.chatService.GetChat(r.Context(), chat_id)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chat)
	return nil
}

// CreateTopic creates a new topic in a chat.
// @Summary      Create a new topic in a chat
// @Description  Creates a new topic in a chat.
// @Tags         topics
// @Accept       json
// @Produce      json
// @Param        id path string true "Chat ID"
// @Param        request body services.CreateTopicDTO true "Topic details"
// @Success      200  {object}  repository.Topic
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /chats/{id}/topics [post]
// @Security     BearerAuth
func (c *ChatController) CreateTopic(w http.ResponseWriter, r *http.Request) error {
	chat_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}
	var dto services.CreateTopicDTO
	if err := dtos.ParseFromBody(r, &dto); err != nil {
		return err
	}
	chat, err := c.chatService.CreateTopic(r.Context(), chat_id, &dto)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chat)
	return nil
}

// CreateMessage creates a new message in a chat.
// @Summary      Create a new message in a chat
// @Description  Creates a new message in a chat.
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        id path string true "Chat ID"
// @Param        request body services.CreateMessageDTO true "Message details"
// @Success      200  {object}  repository.Message
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /chats/{id}/messages [post]
// @Security     BearerAuth
func (c *ChatController) CreateMessage(w http.ResponseWriter, r *http.Request) error {
	chat_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}
	var dto services.CreateMessageDTO
	if err := dtos.ParseFromBody(r, &dto); err != nil {
		return err
	}
	msg, err := c.chatService.CreateMessage(r.Context(), chat_id, &dto)
	if err != nil {
		return err
	}

	me, err := c.userService.GetMe(r)
	if err != nil {
		return err
	}

	usrs, err := c.chatService.ListChatMembers(r.Context(), chat_id)
	if err != nil {
		return err
	}

	go c.webpushService.SendNewMessageNotification(context.Background(), msg, usrs, me)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(msg)
	return nil
}

// CreateInvitation creates an invitation link for a chat.
// @Summary      Create invitation
// @Description  Create an invitation link for a chat. Returns the invitation ID.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        id path string true "Chat ID"
// @Success      200  {object}  services.InvitationResponse
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /chats/{id}/invitation [post]
// @Security     BearerAuth
func (c *ChatController) CreateInvitation(w http.ResponseWriter, r *http.Request) error {
	chat_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}
	id, err := c.chatService.CreateInvitation(r.Context(), chat_id)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(id)
	return nil
}

// InviteUser adds a user to a chat directly.
// @Summary      Invite user to chat
// @Description  Add a user directly to a chat by user ID.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        id path string true "Chat ID"
// @Param        user_id path string true "User ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /chats/{id}/invite/{user_id} [post]
// @Security     BearerAuth
func (c *ChatController) InviteUser(w http.ResponseWriter, r *http.Request) error {
	chat_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}
	user_id, err := handlers.GetParamID(r, "user_id")
	if err != nil {
		return err
	}

	err = c.chatService.InviteUser(r.Context(), chat_id, user_id)
	if err != nil {
		return err
	}

	go c.pubsubService.Publish(context.Background(), user_id.String(), map[string]interface{}{
		"type": "user_added_to_chat",
		"chat_id": chat_id.String(),
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	return nil
}

// CreatePrivateChat creates a new private chat with another user.
// @Summary      Create private chat
// @Description  Create a new private (1-on-1) chat with another user.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        user_id query string true "The other user's ID"
// @Success      200  {object}  repository.Chat
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      409  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /chats/private [post]
// @Security     BearerAuth
func (c *ChatController) CreatePrivateChat(w http.ResponseWriter, r *http.Request) error {
	user1_id, err := services.ExtractUserId(r.Context())
	if err != nil {
		return err
	}
	user2_id, err := uuid.Parse(r.URL.Query().Get("user_id"))
	if err != nil {
		return httperrors.NewHTTPBadRequestError("invalid user_id")
	}
	chat, err := c.chatService.CreatePrivateChat(r.Context(), user1_id, user2_id)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chat)
	return nil
}

// UpdateChat updates a chat's details.
// @Summary      Update chat
// @Description  Update a chat's title and metadata.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        id path string true "Chat ID"
// @Param        request body services.UpdateChatDTO true "Chat details"
// @Success      200  {object}  repository.Chat
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /chats/{id} [put]
// @Security     BearerAuth
func (c *ChatController) UpdateChat(w http.ResponseWriter, r *http.Request) error {
	chat_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}
	var dto services.UpdateChatDTO
	if err := dtos.ParseFromBody(r, &dto); err != nil {
		return err
	}
	chat, err := c.chatService.UpdateChat(r.Context(), chat_id, &dto)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chat)
	return nil
}

// LeaveChat removes the authenticated user from a chat.
// @Summary      Leave chat
// @Description  Remove the authenticated user from a chat.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        id path string true "Chat ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /chats/{id}/leave [post]
// @Security     BearerAuth
func (c *ChatController) LeaveChat(w http.ResponseWriter, r *http.Request) error {
	chat_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}
	err = c.chatService.LeaveChat(r.Context(), chat_id)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	return nil
}

type MuteChatDTO struct {
	Muted bool `json:"muted"`
}

func (d *MuteChatDTO) Validate() error {
	return nil
}

// MuteChat toggles mute status for a chat.
// @Summary      Mute/unmute chat
// @Description  Mute or unmute notifications for a chat.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        id path string true "Chat ID"
// @Param        request body MuteChatDTO true "Mute status"
// @Success      200  {object}  repository.Chat
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /chats/{id}/mute [put]
// @Security     BearerAuth
func (c *ChatController) MuteChat(w http.ResponseWriter, r *http.Request) error {
	chat_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}
	var dto MuteChatDTO
	if err := dtos.ParseFromBody(r, &dto); err != nil {
		return err
	}
	chat, err := c.chatService.MuteChat(r.Context(), chat_id, dto.Muted)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chat)
	return nil
}

// UploadChatAvatar uploads an avatar for a chat.
// @Summary      Upload chat avatar
// @Description  Upload an avatar image for a chat.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        id path string true "Chat ID"
// @Success      200  {object}  repository.Chat
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /chats/{id}/avatar [post]
// @Security     BearerAuth
func (c *ChatController) UploadChatAvatar(w http.ResponseWriter, r *http.Request) error {
	chat_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}

	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return httperrors.NewHTTPBadRequestError("File too large or invalid form")
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		return httperrors.NewHTTPBadRequestError("No file provided")
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return httperrors.NewHTTPBadRequestError("Only image files are allowed")
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	uuid, err := c.uploadToFileStorage(fileBytes)
	if err != nil {
		return err
	}

	avatarUrl := c.publicFileStorageUrl + "/" + uuid

	chat, err := c.chatService.UpdateChatAvatar(r.Context(), chat_id, avatarUrl)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chat)
	return nil
}

func (c *ChatController) uploadToFileStorage(data []byte) (string, error) {
	req, err := http.NewRequest("POST", c.fileStorageUrl+"/file", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	apiKey := os.Getenv("FILE_STORAGE_API_KEY")
	if apiKey != "" {
		req.Header.Set("X-Api-Key", apiKey)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("file_storage error: %s", string(body))
	}

	var result struct {
		UUID string `json:"uuid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.UUID, nil
}

// ListChatMembers returns all members of a chat.
// @Summary      List chat members
// @Description  Returns all members of a chat.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        id path string true "Chat ID"
// @Success      200  {object}  []repository.User
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /chats/{id}/members [get]
// @Security     BearerAuth
func (c *ChatController) ListChatMembers(w http.ResponseWriter, r *http.Request) error {
	chat_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}
	members, err := c.chatService.ListChatMembers(r.Context(), chat_id)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(members)
	return nil
}

// GetMyRole returns the authenticated user's role in a chat.
// @Summary      Get my role in chat
// @Description  Returns the role of the authenticated user in a chat.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        id path string true "Chat ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /chats/{id}/my-role [get]
// @Security     BearerAuth
func (c *ChatController) GetMyRole(w http.ResponseWriter, r *http.Request) error {
	chat_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}
	role, err := c.chatService.GetMyRole(r.Context(), chat_id)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"role": string(*role)})
	return nil
}
