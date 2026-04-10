package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Visoff/messanger/internal/services"
	"github.com/Visoff/messanger/pkgs/dtos"
	"github.com/Visoff/messanger/pkgs/handlers"
)

type ChatController struct {
	chatService   *services.ChatService
	pubsubService *services.PubSubService
	mux           *http.ServeMux
}

func NewChatController(chatService *services.ChatService, pubsubService *services.PubSubService, authService *services.AuthService) *ChatController {
	c := &ChatController{
		chatService: chatService,
		pubsubService: pubsubService,
		mux:         nil,
	}

	mux := http.NewServeMux()
	c.mux = mux

	mux.Handle("GET /", authService.ProtectRoute(handlers.Handler(c.ListChats)))
	mux.Handle("POST /", authService.ProtectRoute(handlers.Handler(c.CreateChat)))

	mux.Handle("GET /{id}", authService.ProtectRoute(handlers.Handler(c.GetChat)))

	/*
		mux.Handle("PUT /{id}", handlers.Handler(c.UpdateChat))
		mux.Handle("DELETE /{id}", handlers.Handler(c.DeleteChat))
	*/

	mux.Handle("GET /{id}/topics", authService.ProtectRoute(handlers.Handler(c.ListTopics)))
	mux.Handle("POST /{id}/topics", authService.ProtectRoute(handlers.Handler(c.CreateTopic)))

	mux.Handle("GET /{id}/messages", authService.ProtectRoute(handlers.Handler(c.ListMessages)))
	mux.Handle("POST /{id}/messages", authService.ProtectRoute(handlers.Handler(c.CreateMessage)))

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

// ListMessages returns a list of all messages in a chat.
// @Summary      List all messages in a chat
// @Description  Returns a list of all messages in a chat.
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        id path int true "Chat ID"
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
// @Param        id path int true "Chat ID"
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
// @Param        id path int true "Chat ID"
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
// @Param        id path int true "Chat ID"
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

	c.pubsubService.Publish(r.Context(), "messanger", msg)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(msg)
	return nil
}
