package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Visoff/messanger/internal/services"
	"github.com/Visoff/messanger/pkgs/dtos"
	"github.com/Visoff/messanger/pkgs/handlers"
)

type TopicController struct {
	topicService *services.TopicService
	mux          *http.ServeMux
}

func NewTopicController(topicService *services.TopicService, auth_service *services.AuthService) *TopicController {
	c := &TopicController{
		topicService: topicService,
		mux:         nil,
	}

	mux := http.NewServeMux()
	c.mux = mux

	/*
	mux.Handle("PUT /{id}", handlers.Handler(c.UpdateTopic))
	mux.Handle("DELETE /{id}", handlers.Handler(c.DeleteTopic))
	*/

	mux.Handle("GET /{id}", auth_service.ProtectRoute(handlers.Handler(c.GetTopic)))
	mux.Handle("GET /{id}/messages", auth_service.ProtectRoute(handlers.Handler(c.ListMessages)))
	mux.Handle("POST /{id}/messages", auth_service.ProtectRoute(handlers.Handler(c.CreateMessage)))

	return c
}

func (c *TopicController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mux.ServeHTTP(w, r)
}

// ListMessages returns a list of all messages in a topic.
// @Summary      List all messages in a topic
// @Description  Returns a list of all messages in a topic.
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        id path int true "Topic ID"
// @Success      200  {object}  []repository.Message
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /topics/{id}/messages [get]
// @Security     BearerAuth
func (c *TopicController) ListMessages(w http.ResponseWriter, r *http.Request) error {
	topic_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}
	messages, err := c.topicService.ListMessages(r.Context(), topic_id)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(messages)
	return nil
}

// CreateMessage creates a new message in a topic.
// @Summary      Create a new message in a topic
// @Description  Creates a new message in a topic.
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        id path int true "Topic ID"
// @Param        request body services.CreateMessageDTO true "Message details"
// @Success      200  {object}  repository.Message
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /topics/{id}/messages [post]
// @Security     BearerAuth
func (c *TopicController) CreateMessage(w http.ResponseWriter, r *http.Request) error {
	topic_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}
	var dto services.CreateMessageDTO
	if err := dtos.ParseFromBody(r, &dto); err != nil {
		return err
	}
	msg, err := c.topicService.CreateMessage(r.Context(), topic_id, &dto)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(msg)
	return nil
}

// GetTopic returns a topic by ID.
// @Summary      Get a topic by ID
// @Description  Returns a topic by ID.
// @Tags         topics
// @Accept       json
// @Produce      json
// @Param        id path int true "Topic ID"
// @Success      200  {object}  repository.Topic
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /topics/{id} [get]
// @Security     BearerAuth
func (c *TopicController) GetTopic(w http.ResponseWriter, r *http.Request) error {
	topic_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}
	topic, err := c.topicService.GetTopic(r.Context(), topic_id)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(topic)
	return nil
}
