package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Visoff/messanger/internal/services"
	"github.com/Visoff/messanger/pkgs/dtos"
	"github.com/Visoff/messanger/pkgs/handlers"
	"github.com/Visoff/messanger/pkgs/httperrors"
)

type UserController struct {
	userService         *services.UserService
	mux                 *http.ServeMux
	fileStorageUrl      string
	publicFileStorageUrl string
}

func (c *UserController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mux.ServeHTTP(w, r)
}

func NewUserController(userService *services.UserService, authService *services.AuthService, fileStorageUrl string, publicFileStorageUrl string) *UserController {
	c := &UserController{
		userService:          userService,
		mux:                  nil,
		fileStorageUrl:       fileStorageUrl,
		publicFileStorageUrl: publicFileStorageUrl,
	}

	mux := http.NewServeMux()
	c.mux = mux

	mux.Handle("POST /register", handlers.Handler(c.RegisterUser))
	mux.Handle("POST /login", handlers.Handler(c.LoginUser))

	mux.Handle("GET /me", authService.ProtectRoute(handlers.Handler(c.GetMe)))
	mux.Handle("PUT /me", authService.ProtectRoute(handlers.Handler(c.UpdateMe)))
	mux.Handle("POST /me/avatar", authService.ProtectRoute(handlers.Handler(c.UploadMyAvatar)))

	mux.Handle("GET /id/{id}", handlers.Handler(c.GetUserByID))
	mux.Handle("GET /username/{username}", handlers.Handler(c.GetUserByUsername))

	//mux.Handle("GET /", handlers.Handler(c.Search))

	return c
}

// RegisterUser registers a new user.
// @Summary      Register a user
// @Description  Register a new user.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body services.RegisterUserDTO true "User details"
// @Success      200  {object}  services.AccessToken
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /users/register [post]
func (c *UserController) RegisterUser(w http.ResponseWriter, r *http.Request) error {
	var dto services.RegisterUserDTO

	if err := dtos.ParseFromBody(r, &dto); err != nil {
		return err
	}

	token, err := c.userService.RegisterUser(r.Context(), &dto)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(token)

	return nil
}

// LoginUser logs in a user.
// @Summary      Login a user
// @Description  Log in a user.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body services.LoginUserDTO true "User details"
// @Success      200  {object}  services.AccessToken
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /users/login [post]
func (c *UserController) LoginUser(w http.ResponseWriter, r *http.Request) error {
	var dto services.LoginUserDTO

	if err := dtos.ParseFromBody(r, &dto); err != nil {
		return err
	}

	token, err := c.userService.LoginUser(r.Context(), &dto)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(token)

	return nil
}

// GetMe gets the current user.
// @Summary      Get current user
// @Description  Get the current user.
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  services.DisplayUser
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /users/me [get]
func (c *UserController) GetMe(w http.ResponseWriter, r *http.Request) error {
	user, err := c.userService.GetMe(r)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(c.userService.NewDisplayUser(user))
	return nil
}

// UpdateMe updates the current user.
// @Summary      Update current user
// @Description  Update the current user's profile.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body services.UpdateUserDTO true "User details"
// @Success      200  {object}  services.DisplayUser
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Router       /users/me [put]
// @Security     BearerAuth
func (c *UserController) UpdateMe(w http.ResponseWriter, r *http.Request) error {
	var dto services.UpdateUserDTO
	if err := dtos.ParseFromBody(r, &dto); err != nil {
		return err
	}
	user, err := c.userService.UpdateUser(r.Context(), &dto)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
	return nil
}

// GetUserByUsername gets a user by username.
// @Summary      Get user by username
// @Description  Get a user by username.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        username path string true "Username"
// @Success      200  {object}  services.DisplayUser
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /users/username/{username} [get]
func (c *UserController) GetUserByUsername(w http.ResponseWriter, r *http.Request) error {
	username, err := handlers.GetParamString(r, "username")
	if err != nil {
		return err
	}
	user, err := c.userService.GetUserByUsername(r.Context(), username)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(c.userService.NewDisplayUser(user))
	return nil
}

func (c *UserController) UploadMyAvatar(w http.ResponseWriter, r *http.Request) error {
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return httperrors.NewHTTPBadRequestError("File too large or invalid form")
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		return httperrors.NewHTTPBadRequestError("No file provided")
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return httperrors.NewHTTPBadRequestError("Only image files are allowed")
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	uuid, err := c.uploadToFileStorage(fileBytes)
	if err != nil {
		return err
	}

	avatarUrl := c.publicFileStorageUrl + "/" + uuid

	user, err := c.userService.UpdateMyAvatar(r.Context(), avatarUrl)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
	return nil
}

func (c *UserController) uploadToFileStorage(data []byte) (string, error) {
	resp, err := http.Post(c.fileStorageUrl+"/file", "application/octet-stream", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("file_storage error: %s", string(body))
	}

	var result struct {
		UUID string `json:"uuid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.UUID, nil
}

// GetUserByID gets a user by ID.
// @Summary      Get user by ID
// @Description  Get a user by ID.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id path int true "User ID"
// @Success      200  {object}  services.DisplayUser
// @Failure      400  {object}  httperrors.ErrorResponse
// @Failure      401  {object}  httperrors.ErrorResponse
// @Failure      500  {object}  httperrors.ErrorResponse
// @Router       /users/id/{id} [get]
func (c *UserController) GetUserByID(w http.ResponseWriter, r *http.Request) error {
	uid, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}
	user, err := c.userService.GetUserByID(r.Context(), uid)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(c.userService.NewDisplayUser(user))
	return nil
}
