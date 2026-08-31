package api

import (
	"awesomeProject/internal/app/handler"
	"awesomeProject/internal/app/repository"

	"github.com/gin-gonic/gin"
)

// StartServer запускает веб-сервер
func StartServer() {
	r := gin.Default()

	// Указываем путь к твоим CSS из папки resources
	r.Static("/static", "./resources")
	r.LoadHTMLGlob("templates/*")

	repo := repository.NewInMemoryHeatingRepository()
	h := handler.NewHeatingHandler(repo)

	r.GET("/", h.FeedHandler)
	r.GET("/feed", h.FeedHandler)
	r.GET("/draft", h.DraftHandler)
	r.GET("/grid", h.GridHandler)

	r.Run("0.0.0.0:8080")
}
