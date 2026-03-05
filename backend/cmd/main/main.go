package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Visoff/messanger/internal/controllers"
	"github.com/Visoff/messanger/internal/migrations"
	"github.com/Visoff/messanger/internal/repository"
	"github.com/Visoff/messanger/internal/services"
	"github.com/Visoff/messanger/pkgs/database"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	ctx := context.Background()
	pool, err := database.New(ctx, "postgres://postgres:postgres@localhost/postgres?sslmode=disable")

	if err = pool.Ping(ctx); err != nil {
		panic(err)
	}

	if err = migrations.Migrate(pool); err != nil {
		panic(err)
	}

	// repository
	repo := repository.New(pool)

	// services
	auth_service := services.NewAuthService("secret")
	user_service := services.NewUserService(repo, auth_service)

	// controllers
	user_controller := controllers.NewUserController(user_service)

	mux := http.NewServeMux()

	mux.Handle("/users/", http.StripPrefix("/users", user_controller))

	log.Println("Server is running on port 8080")
	err = http.ListenAndServe(":8080", mux)

	if err != nil {
		panic(err)
	}
}
