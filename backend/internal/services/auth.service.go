package services

import (
	"context"
	"net/http"
	"strings"

	"github.com/Visoff/messanger/pkgs/handlers"
	"github.com/Visoff/messanger/pkgs/httperrors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	jwt_secret string
}

func NewAuthService(jwt_secret string) *AuthService {
	return &AuthService{
		jwt_secret: jwt_secret,
	}
}

func (s *AuthService) HashPassword(password string) string {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed)
}

func (s *AuthService) CheckPassword(password string, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (s *AuthService) GenerateToken(user_id string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user_id,
	})
	tokenString, _ := token.SignedString([]byte(s.jwt_secret))
	return tokenString
}

func (s *AuthService) GetUserId(token string) (string, error) {
	t, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, nil
		}
		return []byte(s.jwt_secret), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := t.Claims.(jwt.MapClaims); ok {
		return claims["user_id"].(string), nil
	}
	return "", nil
}

func (s *AuthService) ProtectRoute(handler handlers.Handler) handlers.Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		token := r.Header.Get("Authorization")
		if token == "" {
			return httperrors.NewHTTPUnauthorizedError("Unauthorized")
		}
		if !strings.HasPrefix(token, "Bearer ") {
			return httperrors.NewHTTPUnauthorizedError("Unauthorized")
		}

		user_id, err := s.GetUserId(strings.TrimPrefix(token, "Bearer "))
		if err != nil {
			return httperrors.NewHTTPUnauthorizedError("Unauthorized")
		}
		ctx := context.WithValue(r.Context(), "user_id", user_id)
		return handler(w, r.WithContext(ctx))
	}
}


func (s *AuthService) PullUserIdFromAuth(r *http.Request) string {
	if user_id, ok := r.Context().Value("user_id").(string); ok {
		return user_id
	}
	return ""
}
