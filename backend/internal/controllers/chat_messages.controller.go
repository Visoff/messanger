package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Visoff/messanger/internal/services"
	"github.com/Visoff/messanger/pkgs/dtos"
	"github.com/Visoff/messanger/pkgs/handlers"
)

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
