package api

import (
	"awesomeProject/internal/app/dsn"
	"awesomeProject/internal/app/handler"
	"awesomeProject/internal/app/repository"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func StartServer() {
	_ = godotenv.Load()

	repo, err := repository.New(repository.Settings{
		PostgresDSN:    dsn.FromEnv(),
		MinioEndpoint:  os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey: os.Getenv("MINIO_SECRET_KEY"),
		MinioBucket:    os.Getenv("MINIO_BUCKET_NAME"),
		MinioUseSSL:    strings.EqualFold(os.Getenv("MINIO_USE_SSL"), "true"),
	})
	if err != nil {
		log.Fatalf("repository initialization failed: %v", err)
	}

	r := NewRouter(repo, repo, os.Getenv("MINIO_PUBLIC_URL"))
	if err := r.Run("0.0.0.0:8080"); err != nil {
		log.Fatal(err)
	}
}

func NewRouter(equipmentRepo repository.HeatingEquipmentRepository, userRepo repository.UserRepository, mediaBaseURL string) *gin.Engine {
	r := gin.Default()
	r.MaxMultipartMemory = 64 << 20

	equipment := handler.NewHeatingEquipmentHandler(equipmentRepo, mediaBaseURL)
	users := handler.NewUserHandler(userRepo)

	api := r.Group("/api")
	{
		api.GET("/heating-equipment", equipment.List)
		api.POST("/heating-equipment", equipment.CreateDraft)
		api.GET("/heating-equipment/draft", equipment.Draft)
		api.PUT("/heating-equipment/:id/publish", equipment.Publish)
		api.DELETE("/heating-equipment/:id", equipment.Delete)
		api.POST("/heating-equipment/:id/like", equipment.Like)

		api.GET("/heating-equipment/feed", equipment.Feed)
		api.GET("/heating-equipment/feed/:id", equipment.FeedByID)

		api.POST("/users", users.Register)
		api.POST("/users/auth", users.Authenticate)
		api.POST("/users/logout", users.Logout)
	}

	return r
}
