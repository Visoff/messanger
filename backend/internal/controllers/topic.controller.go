package controllers

import (
	"net/http"

	"github.com/Visoff/messanger/internal/services"
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

	return c
}

func (c *TopicController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mux.ServeHTTP(w, r)
}
