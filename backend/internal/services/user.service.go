package services

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Visoff/messanger/internal/repository"
	"github.com/Visoff/messanger/pkgs/httperrors"
	"github.com/google/uuid"
)

type UserService struct {
	repository *repository.Queries
	authService    *AuthService
}

func NewUserService(repository *repository.Queries, authService *AuthService) *UserService {
	return &UserService{repository: repository, authService: authService}
}

type AccessToken struct {
	Token string `json:"token"`
}

type DisplayUser struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	AvatarUrl    *string    `json:"avatar_url"`
	Metadata     []byte     `json:"metadata"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
	LastSeenAt   time.Time  `json:"last_seen_at"`
}

func (s *UserService)NewDisplayUser(user *repository.User) *DisplayUser {
	return &DisplayUser{
		ID:           user.ID,
		Username:     user.Username,
		AvatarUrl:    user.AvatarUrl,
		Metadata:     user.Metadata,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		DeletedAt:    user.DeletedAt,
		LastSeenAt:   user.LastSeenAt,
	}
}

type RegisterUserDTO struct {
	Username     string `json:"username"`
	Password string `json:"password"`
}

func (dto *RegisterUserDTO) Validate() error {
	errors := make(map[string]string)
	if dto.Username == "" {
		errors["username"] = "Username is required"
	}
	if dto.Password == "" {
		errors["password"] = "Password is required"
	}
	return httperrors.NewHTTPValidationError(errors)
}

func (s *UserService) RegisterUser(ctx context.Context, dto *RegisterUserDTO) (*AccessToken, error) {
	usr, err := s.repository.CreateUser(ctx, &repository.CreateUserParams{
		Username: dto.Username,
		PasswordHash: s.authService.HashPassword(dto.Password),
	})
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return nil, httperrors.NewHTTPConflictError("User already exists")
		}
		return nil, err
	}
	return &AccessToken{Token: s.authService.GenerateToken(usr.ID.String())}, nil
}

type LoginUserDTO struct {
	Username string `json:"username"`
	Password  string `json:"password"`
}

func (dto *LoginUserDTO) Validate() error {
	errors := make(map[string]string)
	if dto.Username == "" {
		errors["username"] = "Username is required"
	}
	if dto.Password == "" {
		errors["password"] = "Password is required"
	}
	return httperrors.NewHTTPValidationError(errors)
}

func (s *UserService) LoginUser(ctx context.Context, dto *LoginUserDTO) (*AccessToken, error) {
	user, err := s.repository.GetUserByUsername(ctx, dto.Username)
	if err != nil {
		return nil, httperrors.NewHTTPNotFoundError("User not found")
	}
	if !s.authService.CheckPassword(dto.Password, user.PasswordHash) {
		return nil, httperrors.NewHTTPUnauthorizedError("Invalid password")
	}

	token := s.authService.GenerateToken(user.ID.String())
	return &AccessToken{Token: token}, nil
}

func (s *UserService) GetMe(r *http.Request) (*repository.User, error) {
	id, err := ExtractUserId(r.Context())
	if err != nil {
		return nil, httperrors.NewHTTPUnauthorizedError("Unauthorized")
	}
	user, err := s.repository.GetUserById(r.Context(), id)
	if err != nil {
		return nil, httperrors.NewHTTPNotFoundError("User not found")
	}
	return user, nil
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*repository.User, error) {
	user, err := s.repository.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, httperrors.NewHTTPNotFoundError("User not found")
	}
	return user, nil
}

type UpdateUserDTO struct {
	Username  string          `json:"username"`
	AvatarUrl *string         `json:"avatar_url"`
	Metadata  json.RawMessage `json:"metadata"`
}

func (dto *UpdateUserDTO) Validate() error {
	errors := make(map[string]string)
	if dto.Username == "" {
		errors["username"] = "Username is required"
	}
	if len(errors) > 0 {
		return httperrors.NewHTTPValidationError(errors)
	}
	return nil
}

func (s *UserService) UpdateUser(ctx context.Context, dto *UpdateUserDTO) (*DisplayUser, error) {
	user_id, err := ExtractUserId(ctx)
	if err != nil {
		return nil, err
	}
	metadata := []byte("{}")
	if dto.Metadata != nil {
		metadata = dto.Metadata
	}
	user, err := s.repository.UpdateUser(ctx, &repository.UpdateUserParams{
		ID:        user_id,
		Username:  dto.Username,
		AvatarUrl: dto.AvatarUrl,
		Metadata:  metadata,
	})
	if err != nil {
		return nil, err
	}
	return s.NewDisplayUser(user), nil
}
