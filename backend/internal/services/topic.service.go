package services

import (
	"github.com/Visoff/messanger/internal/repository"
)

type TopicService struct {
	repository *repository.Queries
}

func NewTopicService(repository *repository.Queries) *TopicService {
	return &TopicService{repository: repository}
}
