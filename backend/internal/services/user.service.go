package services

import (
	"context"
	"net/http"
	"strings"

	"github.com/Visoff/messanger/internal/repository"
	"github.com/Visoff/messanger/pkgs/handlers"
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

func (s *UserService) ProtectRoute(handler handlers.Handler) handlers.Handler {
	return s.authService.ProtectRoute(handler)
}

type AccessToken struct {
	Token string `json:"token"`
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
	user_id := s.authService.PullUserIdFromAuth(r)
	id, err := uuid.Parse(user_id)
	if err != nil {
		return nil, httperrors.NewHTTPUnauthorizedError("Unauthorized")
	}
	user, err := s.repository.GetUserById(r.Context(), id)
	if err != nil {
		return nil, httperrors.NewHTTPNotFoundError("User not found")
	}
	return user, nil
}
