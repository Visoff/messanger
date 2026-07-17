package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Visoff/messanger/pkgs/dtos"
	"github.com/Visoff/messanger/pkgs/handlers"
)

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
		"type":    "user_added_to_chat",
		"chat_id": chat_id.String(),
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	return nil
}

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
