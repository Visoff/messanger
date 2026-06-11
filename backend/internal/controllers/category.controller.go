package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Visoff/messanger/internal/services"
	"github.com/Visoff/messanger/pkgs/dtos"
	"github.com/Visoff/messanger/pkgs/handlers"
)

type CategoryController struct {
	categoryService *services.CategoryService
	mux             *http.ServeMux
}

func NewCategoryController(categoryService *services.CategoryService, authService *services.AuthService) *CategoryController {
	c := &CategoryController{
		categoryService: categoryService,
		mux:             nil,
	}

	mux := http.NewServeMux()
	c.mux = mux

	mux.Handle("GET /", authService.ProtectRoute(handlers.Handler(c.ListCategories)))
	mux.Handle("POST /", authService.ProtectRoute(handlers.Handler(c.CreateCategory)))
	mux.Handle("PUT /{id}", authService.ProtectRoute(handlers.Handler(c.UpdateCategory)))
	mux.Handle("DELETE /{id}", authService.ProtectRoute(handlers.Handler(c.DeleteCategory)))

	return c
}

func (c *CategoryController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mux.ServeHTTP(w, r)
}

func (c *CategoryController) ListCategories(w http.ResponseWriter, r *http.Request) error {
	categories, err := c.categoryService.ListCategories(r.Context())
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(categories)
	return nil
}

func (c *CategoryController) CreateCategory(w http.ResponseWriter, r *http.Request) error {
	var dto services.CreateCategoryDTO
	if err := dtos.ParseFromBody(r, &dto); err != nil {
		return err
	}
	cat, err := c.categoryService.CreateCategory(r.Context(), &dto)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cat)
	return nil
}

func (c *CategoryController) UpdateCategory(w http.ResponseWriter, r *http.Request) error {
	id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}
	var dto services.UpdateCategoryDTO
	if err := dtos.ParseFromBody(r, &dto); err != nil {
		return err
	}
	cat, err := c.categoryService.UpdateCategory(r.Context(), id, &dto)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cat)
	return nil
}

func (c *CategoryController) DeleteCategory(w http.ResponseWriter, r *http.Request) error {
	id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}
	if err := c.categoryService.DeleteCategory(r.Context(), id); err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	return nil
}
