package handler

import (
	"awesomeProject/internal/app/model"
	"awesomeProject/internal/app/repository"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const currentUserID uint = 1
const defaultEquipmentImage = "/static/img/heating-placeholder.png"

type HeatingEquipmentHandler struct {
	repo repository.HeatingEquipmentRepository
}

func NewHeatingEquipmentHandler(repo repository.HeatingEquipmentRepository) *HeatingEquipmentHandler {
	return &HeatingEquipmentHandler{repo: repo}
}

func (h *HeatingEquipmentHandler) FeedHandler(c *gin.Context) {
	idStr := c.Query("id")
	next := c.Query("next") == "true"

	var equipment *model.HeatingEquipment
	var err error
	if idStr == "" {
		equipment, err = h.repo.GetFirstPublished()
	} else {
		id64, parseErr := strconv.ParseUint(idStr, 10, 64)
		if parseErr != nil {
			c.String(http.StatusBadRequest, "Некорректный id")
			return
		}
		if next {
			equipment, err = h.repo.GetNextPublished(uint(id64))
		} else {
			equipment, err = h.repo.GetByID(uint(id64))
		}
	}
	if err != nil {
		c.String(http.StatusNotFound, "Опубликованное отопительное оборудование не найдено")
		return
	}

	c.HTML(http.StatusOK, "heating_equipment_feed.html", gin.H{
		"equipment":     equipment,
		"likes_count":   equipment.LikesCount(),
		"default_image": defaultEquipmentImage,
		"active_page":   "feed",
	})
}

func (h *HeatingEquipmentHandler) DraftHandler(c *gin.Context) {
	draft, err := h.repo.GetDraft(currentUserID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.HTML(http.StatusOK, "heating_equipment_add.html", gin.H{
		"draft":         draft,
		"has_draft":     draft != nil,
		"default_image": defaultEquipmentImage,
		"active_page":   "draft",
	})
}

func (h *HeatingEquipmentHandler) GridHandler(c *gin.Context) {
	minPowerStr := c.Query("min_power")
	var minPower float64
	if minPowerStr != "" {
		parsed, err := strconv.ParseFloat(minPowerStr, 64)
		if err != nil || parsed < 0 {
			c.String(http.StatusBadRequest, "Некорректная минимальная мощность")
			return
		}
		minPower = parsed
	}

	items, err := h.repo.GetAllPublished(minPower)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.HTML(http.StatusOK, "heating_equipment_grid.html", gin.H{
		"items":         items,
		"min_power":     minPowerStr,
		"default_image": defaultEquipmentImage,
		"active_page":   "grid",
	})
}

func (h *HeatingEquipmentHandler) CreateDraftHandler(c *gin.Context) {
	parsePositive := func(name string) (float64, bool) {
		value, err := strconv.ParseFloat(strings.TrimSpace(c.PostForm(name)), 64)
		return value, err == nil && value > 0
	}

	name := strings.TrimSpace(c.PostForm("name"))
	imageURL := strings.TrimSpace(c.PostForm("image_url"))
	videoURL := strings.TrimSpace(c.PostForm("video_url"))
	power, okPower := parsePositive("power")
	cost, okCost := parsePositive("cost")

	if imageURL == "" {
		imageURL = defaultEquipmentImage
	}

	if name == "" || videoURL == "" || !okPower || !okCost {
		c.String(http.StatusBadRequest, "Заполните название, видео, мощность и стоимость корректными значениями")
		return
	}

	data := repository.DraftData{
		Name:     name,
		ImageURL: imageURL,
		VideoURL: videoURL,
		Power:    power,
		Cost:     cost,
	}

	if _, err := h.repo.CreateDraft(currentUserID, data); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Redirect(http.StatusFound, "/draft")
}

func (h *HeatingEquipmentHandler) PublishDraftHandler(c *gin.Context) {
	id64, err := strconv.ParseUint(c.PostForm("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Некорректный id черновика")
		return
	}

	description := strings.TrimSpace(c.PostForm("description"))
	if description == "" {
		c.String(http.StatusBadRequest, "Заполните краткое описание оборудования")
		return
	}

	data := repository.PublishDraftData{Description: description}
	if err := h.repo.PublishDraft(currentUserID, uint(id64), data); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Redirect(http.StatusFound, "/grid")
}

func (h *HeatingEquipmentHandler) DeleteEquipmentHandler(c *gin.Context) {
	id64, err := strconv.ParseUint(c.PostForm("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Некорректный id")
		return
	}
	if err := h.repo.DeleteEquipmentSQL(uint(id64)); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Redirect(http.StatusFound, "/grid")
}
