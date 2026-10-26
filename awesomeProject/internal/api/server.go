package api

import (
	_ "awesomeProject/docs"
	"awesomeProject/internal/app/auth"
	"awesomeProject/internal/app/dsn"
	"awesomeProject/internal/app/handler"
	"awesomeProject/internal/app/repository"
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"os"
	"strings"
	"time"

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

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr, Password: os.Getenv("REDIS_PASSWORD"), DB: 0,
		DialTimeout: 2 * time.Second, ReadTimeout: 2 * time.Second, WriteTimeout: 2 * time.Second})
	defer redisClient.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	err = redisClient.Ping(ctx).Err()
	cancel()
	if err != nil {
		log.Fatalf("Redis initialization failed: %v", err)
	}
	service, err := auth.New(os.Getenv("JWT_SECRET"), &auth.RedisBlacklist{Client: redisClient}, time.Hour)
	if err != nil {
		log.Fatal(err)
	}
	service.SetSessions(&auth.RedisSessions{Client: redisClient, Lookup: func(login string) (uint, error) {
		user, err := repo.GetUserByLogin(login)
		if errors.Is(err, repository.ErrNotFound) {
			return 0, auth.ErrInvalidToken
		}
		if err != nil {
			return 0, err
		}
		return user.ID, nil
	}})
	r := NewRouter(repo, repo, os.Getenv("MINIO_PUBLIC_URL"), service)
	if err := r.Run("0.0.0.0:8080"); err != nil {
		log.Fatal(err)
	}
}

func NewRouter(equipmentRepo repository.HeatingEquipmentRepository, userRepo repository.UserRepository, mediaBaseURL string, service *auth.Service) *gin.Engine {
	r := gin.Default()
	r.MaxMultipartMemory = 64 << 20

	equipment := handler.NewHeatingEquipmentHandler(equipmentRepo, mediaBaseURL)
	users := handler.NewUserHandler(userRepo, service)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	{
		api.GET("/heating-equipment", service.Optional(), equipment.List)
		api.POST("/heating-equipment", service.Required(), equipment.CreateDraft)
		api.GET("/heating-equipment/draft", service.Required(), equipment.Draft)
		api.PUT("/heating-equipment/:id/publish", service.Required(), equipment.Publish)
		api.DELETE("/heating-equipment/:id", service.Required(), equipment.Delete)
		api.POST("/heating-equipment/:id/like", service.Required(), equipment.Like)

		api.GET("/heating-equipment/feed", service.Optional(), equipment.Feed)
		api.GET("/heating-equipment/feed/:id", service.Optional(), equipment.FeedByID)

		api.POST("/users", users.Register)
		api.POST("/users/auth", users.Authenticate)
		api.POST("/users/logout", service.Required(), users.Logout)
	}

	return r
}
