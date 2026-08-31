package handler

import (
	"awesomeProject/internal/app/model"
	"awesomeProject/internal/app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type HeatingHandler struct {
	repo repository.HeatingRepository
}

func NewHeatingHandler(repo repository.HeatingRepository) *HeatingHandler {
	return &HeatingHandler{repo: repo}
}

func (h *HeatingHandler) FeedHandler(c *gin.Context) {
	idStr := c.Query("id")
	nextStr := c.Query("next")

	var calcPtr *model.HeatingCalculation
	var err error

	if idStr == "" {
		calcPtr, err = h.repo.GetFirstPublished()
	} else {
		id, parseErr := strconv.Atoi(idStr)
		if parseErr != nil {
			c.String(http.StatusBadRequest, "Неверный ID")
			return
		}

		if nextStr == "true" {
			calcPtr, err = h.repo.GetNextPublished(id)
		} else {
			calcPtr, err = h.repo.GetByID(id)
		}
	}

	if err != nil {
		c.String(http.StatusNotFound, "Запись не найдена")
		return
	}

	c.HTML(http.StatusOK, "feed.html", gin.H{
		"calculation": calcPtr,
		"likes_count": calcPtr.LikesCount(),
		"active_page": "feed",
	})
}

func (h *HeatingHandler) DraftHandler(c *gin.Context) {
	draft, err := h.repo.GetDraft()
	if err != nil {
		c.String(http.StatusNotFound, "Черновик не найден")
		return
	}

	c.HTML(http.StatusOK, "draft.html", gin.H{
		"draft":       draft,
		"active_page": "draft",
	})
}

func (h *HeatingHandler) GridHandler(c *gin.Context) {
	minAreaStr := c.Query("min_area")
	minArea := 0
	if minAreaStr != "" {
		minArea, _ = strconv.Atoi(minAreaStr)
	}

	calculations := h.repo.GetAllPublished(minArea)

	c.HTML(http.StatusOK, "grid.html", gin.H{
		"calculations": calculations,
		"min_area":     minAreaStr,
		"active_page":  "grid",
	})
}
