package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/Visoff/messanger/docs"
	"github.com/Visoff/messanger/internal/controllers"
	"github.com/Visoff/messanger/internal/migrations"
	"github.com/Visoff/messanger/internal/repository"
	"github.com/Visoff/messanger/internal/services"
	"github.com/Visoff/messanger/pkgs/handlers"
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
	if err != nil {
		log.Println(err)
	}

	connectionString := os.Getenv("DATABASE_URL")
	if connectionString == "" {
		panic("DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		panic(err)
	}

	if err = pool.Ping(ctx); err != nil {
		panic(err)
	}

	if err = migrations.Migrate(pool); err != nil {
		panic(err)
	}

	// repository
	repo := repository.New(pool)

	fileStorageUrl := os.Getenv("FILE_STORAGE_URL")
	if fileStorageUrl == "" {
		fileStorageUrl = "http://localhost:3001"
	}

	publicFileStorageUrl := os.Getenv("PUBLIC_FILE_STORAGE_URL")
	if publicFileStorageUrl == "" {
		publicFileStorageUrl = fileStorageUrl
	}

	// services
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		panic("JWT_SECRET is not set")
	}
	auth_service := services.NewAuthService(jwtSecret)
	user_service := services.NewUserService(repo, auth_service)
	chat_service := services.NewChatService(repo)
	topic_service := services.NewTopicService(repo)
	category_service := services.NewCategoryService(repo)
	pubsub_service, err := services.NewPubSubService()
	webpush_service := services.NewWebPushService(repo, pubsub_service)
	if err != nil {
		panic(err)
	}

	// controllers
	user_controller := controllers.NewUserController(user_service, auth_service, fileStorageUrl, publicFileStorageUrl)
	chat_controller := controllers.NewChatController(chat_service, user_service, pubsub_service, webpush_service, auth_service, fileStorageUrl, publicFileStorageUrl)
	topic_controller := controllers.NewTopicController(topic_service, user_service, webpush_service, auth_service)
	pubsub_controller := controllers.NewPubSubController(pubsub_service, webpush_service, auth_service)
	invitation_controller := controllers.NewInvitationController(chat_service, auth_service)
	category_controller := controllers.NewCategoryController(category_service, auth_service)

	mux := http.NewServeMux()

	mux.Handle("/users/", http.StripPrefix("/users", user_controller))
	mux.Handle("/chats/", http.StripPrefix("/chats", chat_controller))
	mux.Handle("/topics/", http.StripPrefix("/topics", topic_controller))
	mux.Handle("/pubsub/", http.StripPrefix("/pubsub", pubsub_controller))
	mux.Handle("/invitation/", http.StripPrefix("/invitation", invitation_controller))
	mux.Handle("/categories/", http.StripPrefix("/categories", category_controller))

	mux.Handle("/docs/swagger.json", http.StripPrefix("/docs", http.FileServerFS(docs.Docs)))
	mux.Handle("/docs/", httpswagger.Handler(
		httpswagger.URL("http://localhost:8080/docs/swagger.json"),
	))

	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "../frontend/build"
	}
	mux.Handle("/", spaFileServer(staticDir))

	log.Println("Server is running on port 8080")
	err = http.ListenAndServe(":8080", handlers.MiddlewareChain(
		handlers.Logging(log.Default()),
		handlers.AllowCors,
	)(handlers.ToErrorHandler(mux)))

	if err != nil {
		panic(err)
	}
}

func spaFileServer(dir string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := path.Clean(r.URL.Path)
		if clean == "/" || clean == "" {
			http.ServeFile(w, r, dir+"/index.html")
			return
		}
		if strings.HasPrefix(clean, "/_app/") || strings.HasPrefix(clean, "/favicon") ||
			strings.HasPrefix(clean, "/icons") || strings.HasPrefix(clean, "/manifest") ||
			strings.HasPrefix(clean, "/robots") || strings.HasPrefix(clean, "/sw") {
			fs.ServeHTTP(w, r)
			return
		}
		ext := path.Ext(clean)
		if ext != "" {
			fs.ServeHTTP(w, r)
			return
		}
		htmlPath := dir + clean + ".html"
		if _, err := os.Stat(htmlPath); err == nil {
			http.ServeFile(w, r, htmlPath)
			return
		}
		http.ServeFile(w, r, dir+"/index.html")
	})
}
