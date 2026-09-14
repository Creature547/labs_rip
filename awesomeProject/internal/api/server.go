package api

import (
	"awesomeProject/internal/app/dsn"
	"awesomeProject/internal/app/handler"
	"awesomeProject/internal/app/repository"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func StartServer() {
	_ = godotenv.Load()

	repo, err := repository.NewPostgresHeatingRepository(dsn.FromEnv())
	if err != nil {
		log.Fatalf("ошибка подключения к PostgreSQL: %v", err)
	}

	r := gin.Default()
	r.Static("/static", "./resources")
	r.LoadHTMLGlob("templates/*")

	h := handler.NewHeatingHandler(repo)

	
	r.GET("/feed", h.FeedHandler)
	r.GET("/draft", h.DraftHandler)
	r.GET("/grid", h.GridHandler)


	r.POST("/draft/create", h.CreateDraftHandler)
	r.POST("/draft/publish", h.PublishDraftHandler)
	r.POST("/heating/delete", h.DeleteCalculationHandler)

	
	r.NoRoute(func(c *gin.Context) {
		if c.Request.URL.Path == "/" {
			c.Redirect(302, "/feed")
			return
		}
		c.Status(404)
	})

	if err := r.Run("0.0.0.0:8080"); err != nil {
		log.Fatal(err)
	}
}
