package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Visoff/messanger/docs"
	"github.com/Visoff/messanger/internal/controllers"
	"github.com/Visoff/messanger/internal/migrations"
	"github.com/Visoff/messanger/internal/repository"
	"github.com/Visoff/messanger/internal/services"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	httpswagger "github.com/swaggo/http-swagger"
)

// @title           Messenger API
// @version         1.0
// @description     API for a messenger application.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Type "Bearer " followed by your JWT token.
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
	topic_service := services.NewTopicService(repo)

	// controllers
	user_controller := controllers.NewUserController(user_service, auth_service)
	chat_controller := controllers.NewChatController(chat_service, auth_service)
	topic_controller := controllers.NewTopicController(topic_service, auth_service)

	mux := http.NewServeMux()

	mux.Handle("/users/", http.StripPrefix("/users", user_controller))
	mux.Handle("/chats/", http.StripPrefix("/chats", chat_controller))
	mux.Handle("/topics/", http.StripPrefix("/topics", topic_controller))

	mux.Handle("/docs/swagger.json", http.StripPrefix("/docs", http.FileServerFS(docs.Docs)))
	mux.Handle("/docs/", httpswagger.Handler(
		httpswagger.URL("http://localhost:8080/docs/swagger.json"),
	))

	log.Println("Server is running on port 8080")
	err = http.ListenAndServe(":8080", mux)

	if err != nil {
		panic(err)
	}
}
