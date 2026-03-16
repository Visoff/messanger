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

	return c
}

func (c *ChatController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mux.ServeHTTP(w, r)
}

func (c *ChatController) ListChats(w http.ResponseWriter, r *http.Request) error {
	chats := c.chatService.ListChats(r.Context())
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chat)
	return nil
}
