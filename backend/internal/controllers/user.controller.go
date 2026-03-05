package controllers

import (
	"encoding/json"
	"fmt"
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

func NewUserController(userService *services.UserService) *UserController {
	c := &UserController{
		userService: userService,
		mux:         nil,
	}

	mux := http.NewServeMux()
	c.mux = mux

	mux.Handle("POST /register", handlers.Handler(c.RegisterUser))
	mux.Handle("POST /login", handlers.Handler(c.LoginUser))

	mux.Handle("GET /me", c.userService.ProtectRoute(handlers.Handler(c.GetMe)))

	return c
}

func (c *UserController) RegisterUser(w http.ResponseWriter, r *http.Request) error {
	var dto services.RegisterUserDTO

	if err := dtos.ParseFromBody(r, &dto); err != nil {
		return err
	}

	user, err := c.userService.RegisterUser(r.Context(), &dto)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)

	return nil
}

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
	w.Write(fmt.Appendf(nil, `{"token":"%s"}`, token))
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
