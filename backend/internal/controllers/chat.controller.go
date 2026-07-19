package controllers

import (
	"context"
	"encoding/json"
	"net/http"

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

func (c *ChatController) ListChats(w http.ResponseWriter, r *http.Request) error {
	chats, err := c.chatService.ListChats(r.Context())
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chats)
	return nil
}

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
