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

// GetInvitationInfo returns information about an invitation.
// @Summary      Get invitation info
// @Description  Returns information about a chat invitation by its ID.
// @Tags         invitations
// @Accept       json
// @Produce      json
// @Param        id path string true "Invitation ID"
// @Success      200  {object}  services.InvitationInfo
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      404  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /invitation/{id}/info [get]
// @Security     BearerAuth
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

// AcceptInvitationJson accepts a chat invitation.
// @Summary      Accept invitation
// @Description  Accept a chat invitation and join the chat.
// @Tags         invitations
// @Accept       json
// @Produce      json
// @Param        id path string true "Invitation ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /invitation/{id}/accept [post]
// @Security     BearerAuth
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

// RejectInvitation rejects (deletes) a chat invitation.
// @Summary      Reject invitation
// @Description  Reject or delete a chat invitation.
// @Tags         invitations
// @Accept       json
// @Produce      json
// @Param        id path string true "Invitation ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /invitation/{id} [delete]
// @Security     BearerAuth
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
