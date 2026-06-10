package controllers

import (
	"net/http"
	"os"

	"github.com/Visoff/messanger/internal/services"
	"github.com/Visoff/messanger/pkgs/handlers"
	"github.com/google/uuid"
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

	mux.Handle("GET /{id}", handlers.Handler(c.AcceptInvitation))

	return c
}

func (c *InvitationController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mux.ServeHTTP(w, r)
}

func (c *InvitationController) AcceptInvitation(w http.ResponseWriter, r *http.Request) error {
	frontendUrl := os.Getenv("FRONTEND_URL")
	if frontendUrl == "" {
		frontendUrl = "http://localhost:5173"
	}

	invitation_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		http.Redirect(w, r, frontendUrl+"/login", http.StatusFound)
		return nil
	}

	token := r.URL.Query().Get("token")
	var user_id *uuid.UUID
	if token != "" {
		id, err := c.authService.ValidateToken(token)
		if err == nil {
			uid := uuid.MustParse(id)
			user_id = &uid
		}
	}

	if user_id == nil {
		http.Redirect(w, r, frontendUrl+"/login", http.StatusFound)
		return nil
	}

	invitation, err := c.chatService.GetInvitation(r.Context(), invitation_id)
	if err != nil {
		http.Redirect(w, r, frontendUrl, http.StatusFound)
		return nil
	}

	_, err = c.chatService.AcceptInvitation(r.Context(), invitation_id)
	if err != nil {
		http.Redirect(w, r, frontendUrl+"/login", http.StatusFound)
		return nil
	}

	http.Redirect(w, r, frontendUrl+"/?chat_id="+invitation.ChatID.String(), http.StatusFound)
	return nil
}
