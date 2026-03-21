package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Visoff/messanger/internal/services"
	"github.com/Visoff/messanger/pkgs/dtos"
	"github.com/Visoff/messanger/pkgs/handlers"
)

type UserController struct {
	userService *services.UserService
	mux         *http.ServeMux
}

func (c *UserController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mux.ServeHTTP(w, r)
}

func NewUserController(userService *services.UserService, authService *services.AuthService) *UserController {
	c := &UserController{
		userService: userService,
		mux:         nil,
	}

	mux := http.NewServeMux()
	c.mux = mux

	mux.Handle("POST /register", handlers.Handler(c.RegisterUser))
	mux.Handle("POST /login", handlers.Handler(c.LoginUser))

	mux.Handle("GET /me", authService.ProtectRoute(handlers.Handler(c.GetMe)))

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

func (c *UserController) GetMe(w http.ResponseWriter, r *http.Request) error {
	user, err := c.userService.GetMe(r)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
	return nil
}
