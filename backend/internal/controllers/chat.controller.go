package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Visoff/messanger/internal/services"
	"github.com/Visoff/messanger/pkgs/dtos"
	"github.com/Visoff/messanger/pkgs/handlers"
)

type ChatController struct {
	chatService *services.ChatService
	mux         *http.ServeMux
}

func NewChatController(chatService *services.ChatService) *ChatController {
	c := &ChatController{
		chatService: chatService,
		mux:         nil,
	}

	mux := http.NewServeMux()
	c.mux = mux

	mux.Handle("GET /", handlers.Handler(c.ListChats))
	mux.Handle("POST /", handlers.Handler(c.CreateChat))

	/*
	mux.Handle("GET /{id}", handlers.Handler(c.GetChat))
	mux.Handle("PUT /{id}", handlers.Handler(c.UpdateChat))
	mux.Handle("DELETE /{id}", handlers.Handler(c.DeleteChat))
	*/

	mux.Handle("GET /{id}/topics", handlers.Handler(c.ListTopics))

	return c
}

func (c *ChatController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mux.ServeHTTP(w, r)
}

// ListChats returns a list of all chats.
// @Summary      List all chats
// @Description  Returns a list of all chats.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Success      200  {object}  []repository.Chat
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

// CreateChat creates a new chat and adds the authenticated user as owner.
// @Summary      Create a chat
// @Description  Create a new chat (private, group, or channel). The authenticated user becomes the owner.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        request body services.CreateChatDTO true "Chat details"
// @Success      200  {object}  repository.Chat
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /chats/ [post]
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
// @Param        id path int true "Chat ID"
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
