package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Visoff/messanger/internal/controllers"
	"github.com/Visoff/messanger/internal/migrations"
	"github.com/Visoff/messanger/internal/repository"
	"github.com/Visoff/messanger/internal/services"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://postgres:postgres@localhost/postgres?sslmode=disable")

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
	chat_service := services.NewChatService(repo)

	// controllers
	user_controller := controllers.NewUserController(user_service)
	chat_controller := controllers.NewChatController(chat_service)

	mux := http.NewServeMux()

	mux.Handle("/users/", http.StripPrefix("/users", user_controller))
	mux.Handle("/chats/", http.StripPrefix("/chats", chat_controller))

	log.Println("Server is running on port 8080")
	err = http.ListenAndServe(":8080", mux)

	if err != nil {
		panic(err)
	}
}
