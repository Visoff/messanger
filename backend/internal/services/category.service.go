package services

import (
	"context"
	"encoding/json"

	"github.com/Visoff/messanger/internal/repository"
	"github.com/Visoff/messanger/pkgs/httperrors"
	"github.com/google/uuid"
)

type CategoryService struct {
	repository *repository.Queries
}

func NewCategoryService(repository *repository.Queries) *CategoryService {
	return &CategoryService{repository: repository}
}

type UserCategoryResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	ChatIds   []string  `json:"chat_ids"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

func (s *CategoryService) formatCategory(cat *repository.UserCategory) (*UserCategoryResponse, error) {
	var chatIds []string
	if len(cat.ChatIds) > 0 {
		if err := json.Unmarshal(cat.ChatIds, &chatIds); err != nil {
			chatIds = []string{}
		}
	} else {
		chatIds = []string{}
	}
	return &UserCategoryResponse{
		ID:        cat.ID,
		UserID:    cat.UserID,
		Name:      cat.Name,
		ChatIds:   chatIds,
		CreatedAt: cat.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: cat.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *CategoryService) ListCategories(ctx context.Context) ([]*UserCategoryResponse, error) {
	user_id, err := ExtractUserId(ctx)
	if err != nil {
		return nil, err
	}
	cats, err := s.repository.ListUserCategories(ctx, user_id)
	if err != nil {
		return nil, err
	}
	result := make([]*UserCategoryResponse, 0, len(cats))
	for _, cat := range cats {
		formatted, err := s.formatCategory(cat)
		if err != nil {
			return nil, err
		}
		result = append(result, formatted)
	}
	return result, nil
}

type CreateCategoryDTO struct {
	Name    string   `json:"name"`
	ChatIds []string `json:"chat_ids"`
}

func (dto *CreateCategoryDTO) Validate() error {
	errors := make(map[string]string)
	if dto.Name == "" {
		errors["name"] = "Name is required"
	}
	if len(errors) > 0 {
		return httperrors.NewHTTPValidationError(errors)
	}
	return nil
}

func (s *CategoryService) CreateCategory(ctx context.Context, dto *CreateCategoryDTO) (*UserCategoryResponse, error) {
	user_id, err := ExtractUserId(ctx)
	if err != nil {
		return nil, err
	}
	if dto.ChatIds == nil {
		dto.ChatIds = []string{}
	}
	chatIdsJSON, err := json.Marshal(dto.ChatIds)
	if err != nil {
		return nil, err
	}
	cat, err := s.repository.CreateUserCategory(ctx, &repository.CreateUserCategoryParams{
		UserID:  user_id,
		Name:    dto.Name,
		ChatIds: chatIdsJSON,
	})
	if err != nil {
		return nil, err
	}
	return s.formatCategory(cat)
}

type UpdateCategoryDTO struct {
	Name    string   `json:"name"`
	ChatIds []string `json:"chat_ids"`
}

func (dto *UpdateCategoryDTO) Validate() error {
	errors := make(map[string]string)
	if dto.Name == "" {
		errors["name"] = "Name is required"
	}
	if len(errors) > 0 {
		return httperrors.NewHTTPValidationError(errors)
	}
	return nil
}

func (s *CategoryService) UpdateCategory(ctx context.Context, id uuid.UUID, dto *UpdateCategoryDTO) (*UserCategoryResponse, error) {
	user_id, err := ExtractUserId(ctx)
	if err != nil {
		return nil, err
	}
	if dto.ChatIds == nil {
		dto.ChatIds = []string{}
	}
	chatIdsJSON, err := json.Marshal(dto.ChatIds)
	if err != nil {
		return nil, err
	}
	cat, err := s.repository.UpdateUserCategory(ctx, &repository.UpdateUserCategoryParams{
		ID:      id,
		Name:    dto.Name,
		ChatIds: chatIdsJSON,
		UserID:  user_id,
	})
	if err != nil {
		return nil, httperrors.NewHTTPNotFoundError("Category not found")
	}
	return s.formatCategory(cat)
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	user_id, err := ExtractUserId(ctx)
	if err != nil {
		return err
	}
	err = s.repository.DeleteUserCategory(ctx, &repository.DeleteUserCategoryParams{
		ID:     id,
		UserID: user_id,
	})
	if err != nil {
		return httperrors.NewHTTPNotFoundError("Category not found")
	}
	return nil
}
