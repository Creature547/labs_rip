package api

import (
	"awesomeProject/internal/app/handler"
	"awesomeProject/internal/app/repository"

	"github.com/gin-gonic/gin"
)


func StartServer() {
	r := gin.Default()

	
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
