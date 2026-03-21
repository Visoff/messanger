package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Visoff/messanger/internal/services"
	"github.com/Visoff/messanger/pkgs/handlers"
)

type TopicController struct {
	topicService *services.TopicService
	mux          *http.ServeMux
}

func NewTopicController(topicService *services.TopicService) *TopicController {
	c := &TopicController{
		topicService: topicService,
		mux:         nil,
	}

	mux := http.NewServeMux()
	c.mux = mux

	/*
	mux.Handle("GET /{id}", handlers.Handler(c.GetTopic))
	mux.Handle("PUT /{id}", handlers.Handler(c.UpdateTopic))
	mux.Handle("DELETE /{id}", handlers.Handler(c.DeleteTopic))
	*/

	mux.Handle("GET /{id}/messages", handlers.Handler(c.ListMessages))

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
