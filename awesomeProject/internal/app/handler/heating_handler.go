package handler

import (
	"awesomeProject/internal/app/model"
	"awesomeProject/internal/app/repository"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const currentUserID uint = 1 // авторизация будет в следующих лабораторных

type HeatingHandler struct {
	repo repository.HeatingRepository
}

func NewHeatingHandler(repo repository.HeatingRepository) *HeatingHandler {
	return &HeatingHandler{repo: repo}
}

func (h *HeatingHandler) FeedHandler(c *gin.Context) {
	idStr := c.Query("id")
	next := c.Query("next") == "true"

	var calc *model.HeatingCalculation
	var err error
	if idStr == "" {
		calc, err = h.repo.GetFirstPublished()
	} else {
		id64, parseErr := strconv.ParseUint(idStr, 10, 64)
		if parseErr != nil {
			c.String(http.StatusBadRequest, "Некорректный id")
			return
		}
		if next {
			calc, err = h.repo.GetNextPublished(uint(id64))
		} else {
			calc, err = h.repo.GetByID(uint(id64))
		}
	}
	if err != nil {
		c.String(http.StatusNotFound, "Опубликованный расчет не найден")
		return
	}

	c.HTML(http.StatusOK, "feed.html", gin.H{
		"calculation": calc,
		"likes_count": calc.LikesCount(),
		"active_page": "feed",
	})
}

func (h *HeatingHandler) DraftHandler(c *gin.Context) {
	draft, err := h.repo.GetDraft(currentUserID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.HTML(http.StatusOK, "draft.html", gin.H{
		"draft":       draft,
		"has_draft":   draft != nil,
		"active_page": "draft",
	})
}

func (h *HeatingHandler) GridHandler(c *gin.Context) {
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
	c.HTML(http.StatusOK, "grid.html", gin.H{
		"items":       items,
		"min_power":   minPowerStr,
		"active_page": "grid",
	})
}

func (h *HeatingHandler) CreateDraftHandler(c *gin.Context) {
	parsePositive := func(name string) (float64, bool) {
		value, err := strconv.ParseFloat(strings.TrimSpace(c.PostForm(name)), 64)
		return value, err == nil && value > 0
	}

	title := strings.TrimSpace(c.PostForm("title"))
	imageURL := strings.TrimSpace(c.PostForm("image_url"))
	videoURL := strings.TrimSpace(c.PostForm("video_url"))
	description := strings.TrimSpace(c.PostForm("description"))
	floorType := strings.TrimSpace(c.PostForm("floor_type"))
	wallMaterial := strings.TrimSpace(c.PostForm("wall_material"))

	power, okPower := parsePositive("power")
	cost, okCost := parsePositive("cost")
	area, okArea := parsePositive("area")
	ceilingHeight, okCeiling := parsePositive("ceiling_height")
	desiredTemperature, okTemperature := parsePositive("desired_temperature")
	electricityTariff, okTariff := parsePositive("electricity_tariff")

	if title == "" || imageURL == "" || videoURL == "" || description == "" ||
		!okPower || !okCost || !okArea || !okCeiling || !okTemperature || !okTariff ||
		floorType == "" || wallMaterial == "" {
		c.String(http.StatusBadRequest, "Заполните все поля расчета корректными значениями")
		return
	}

	data := repository.DraftData{
		Title:              title,
		ImageURL:           imageURL,
		VideoURL:           videoURL,
		Description:        description,
		Power:              power,
		Cost:               cost,
		Area:               area,
		CeilingHeight:      ceilingHeight,
		FloorType:          floorType,
		WallMaterial:       wallMaterial,
		DesiredTemperature: desiredTemperature,
		ElectricityTariff:  electricityTariff,
	}

	if _, err := h.repo.CreateDraft(currentUserID, data); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Redirect(http.StatusFound, "/draft")
}

func (h *HeatingHandler) PublishDraftHandler(c *gin.Context) {
	id64, err := strconv.ParseUint(c.PostForm("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Некорректный id черновика")
		return
	}

	parsePositive := func(name string) (float64, bool) {
		v, parseErr := strconv.ParseFloat(strings.TrimSpace(c.PostForm(name)), 64)
		return v, parseErr == nil && v > 0
	}

	description := strings.TrimSpace(c.PostForm("description"))
	power, okPower := parsePositive("power")
	cost, okCost := parsePositive("cost")
	area, okArea := parsePositive("area")
	ceilingHeight, okCeiling := parsePositive("ceiling_height")
	desiredTemperature, okTemperature := parsePositive("desired_temperature")
	electricityTariff, okTariff := parsePositive("electricity_tariff")
	floorType := strings.TrimSpace(c.PostForm("floor_type"))
	wallMaterial := strings.TrimSpace(c.PostForm("wall_material"))

	if description == "" || !okPower || !okCost || !okArea || !okCeiling || !okTemperature || !okTariff || floorType == "" || wallMaterial == "" {
		c.String(http.StatusBadRequest, "Заполните все параметры расчета корректными значениями")
		return
	}

	data := repository.PublishDraftData{
		Description:        description,
		Power:              power,
		Cost:               cost,
		Area:               area,
		CeilingHeight:      ceilingHeight,
		FloorType:          floorType,
		WallMaterial:       wallMaterial,
		DesiredTemperature: desiredTemperature,
		ElectricityTariff:  electricityTariff,
	}

	if err := h.repo.PublishDraft(currentUserID, uint(id64), data); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Redirect(http.StatusFound, "/grid")
}

func (h *HeatingHandler) DeleteCalculationHandler(c *gin.Context) {
	id64, err := strconv.ParseUint(c.PostForm("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Некорректный id")
		return
	}
	if err := h.repo.DeleteCalculationSQL(uint(id64)); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Redirect(http.StatusFound, "/grid")
}
