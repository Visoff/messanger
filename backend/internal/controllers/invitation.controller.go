package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Visoff/messanger/internal/services"
	"github.com/Visoff/messanger/pkgs/handlers"
)

type InvitationController struct {
	chatService *services.ChatService
	authService *services.AuthService
	mux         *http.ServeMux
}

func NewInvitationController(chatService *services.ChatService, authService *services.AuthService) *InvitationController {
	c := &InvitationController{
		chatService: chatService,
		authService: authService,
		mux:         nil,
	}

	mux := http.NewServeMux()
	c.mux = mux

	mux.Handle("GET /{id}/info", c.authService.ProtectRoute(handlers.Handler(c.GetInvitationInfo)))
	mux.Handle("POST /{id}/accept", c.authService.ProtectRoute(handlers.Handler(c.AcceptInvitationJson)))
	mux.Handle("DELETE /{id}", c.authService.ProtectRoute(handlers.Handler(c.RejectInvitation)))

	return c
}

func (c *InvitationController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mux.ServeHTTP(w, r)
}

func (c *InvitationController) GetInvitationInfo(w http.ResponseWriter, r *http.Request) error {
	invitation_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}

	info, err := c.chatService.GetInvitationInfo(r.Context(), invitation_id)
	if err != nil {
		http.Error(w, `{"error":"invitation not found"}`, http.StatusNotFound)
		return nil
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
	return nil
}

func (c *InvitationController) AcceptInvitationJson(w http.ResponseWriter, r *http.Request) error {
	invitation_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}

	chat, err := c.chatService.AcceptInvitation(r.Context(), invitation_id)
	if err != nil {
		http.Error(w, `{"error":"failed to accept invitation"}`, http.StatusBadRequest)
		return nil
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"chat_id": chat.ID.String()})
	return nil
}

func (c *InvitationController) RejectInvitation(w http.ResponseWriter, r *http.Request) error {
	invitation_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}

	if err := c.chatService.UseInvitation(r.Context(), invitation_id); err != nil {
		http.Error(w, `{"error":"failed to reject invitation"}`, http.StatusInternalServerError)
		return nil
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	return nil
}
