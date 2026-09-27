package handler

import (
	"awesomeProject/internal/app/currentuser"
	"awesomeProject/internal/app/repository"
	"awesomeProject/internal/app/serializer"
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type HeatingEquipmentHandler struct {
	repo         repository.HeatingEquipmentRepository
	mediaBaseURL string
}

func NewHeatingEquipmentHandler(repo repository.HeatingEquipmentRepository, mediaBaseURL string) *HeatingEquipmentHandler {
	return &HeatingEquipmentHandler{repo: repo, mediaBaseURL: mediaBaseURL}
}

func (h *HeatingEquipmentHandler) List(c *gin.Context) {
	filter := repository.EquipmentFilter{Name: strings.TrimSpace(c.Query("name"))}
	if v := c.Query("min_power"); v != "" {
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil || parsed < 0 {
			c.Status(http.StatusBadRequest)
			return
		}
		filter.MinPower = &parsed
	}
	if v := c.Query("max_cost"); v != "" {
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil || parsed < 0 {
			c.Status(http.StatusBadRequest)
			return
		}
		filter.MaxCost = &parsed
	}

	items, err := h.repo.ListPublished(filter)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	userID := currentuser.Get().ID
	c.JSON(http.StatusOK, serializer.HeatingEquipmentList(items, userID, h.mediaBaseURL, h.repo.MediaBucket()))
}

func (h *HeatingEquipmentHandler) CreateDraft(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(64 << 20); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	description := strings.TrimSpace(c.PostForm("description"))
	power, errPower := strconv.ParseFloat(strings.TrimSpace(c.PostForm("power")), 64)
	cost, errCost := strconv.ParseFloat(strings.TrimSpace(c.PostForm("cost")), 64)
	if name == "" || errPower != nil || errCost != nil || power <= 0 || cost <= 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	image, err := c.FormFile("image")
	if err != nil || validateMediaFile(image, "image/") != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	video, err := c.FormFile("video")
	if err != nil || validateMediaFile(video, "video/") != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	item, err := h.repo.CreateDraft(currentuser.Get().ID, repository.DraftData{
		Name: name, Description: description, Power: power, Cost: cost,
	}, image, video)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDraftExists):
			c.Status(http.StatusConflict)
		default:
			c.Status(http.StatusInternalServerError)
		}
		return
	}
	c.JSON(http.StatusCreated, serializer.HeatingEquipmentFromModel(*item, currentuser.Get().ID, h.mediaBaseURL, h.repo.MediaBucket()))
}

func (h *HeatingEquipmentHandler) Draft(c *gin.Context) {
	item, err := h.repo.GetDraft(currentuser.Get().ID)
	if err != nil {
		handleRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, serializer.HeatingEquipmentFromModel(*item, currentuser.Get().ID, h.mediaBaseURL, h.repo.MediaBucket()))
}

func (h *HeatingEquipmentHandler) Publish(c *gin.Context) {
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}
	item, err := h.repo.PublishDraft(currentuser.Get().ID, id)
	if err != nil {
		handleRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, serializer.HeatingEquipmentFromModel(*item, currentuser.Get().ID, h.mediaBaseURL, h.repo.MediaBucket()))
}

func (h *HeatingEquipmentHandler) Feed(c *gin.Context) {
	item, err := h.repo.GetFirstPublished()
	if err != nil {
		handleRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, serializer.HeatingEquipmentFromModel(*item, currentuser.Get().ID, h.mediaBaseURL, h.repo.MediaBucket()))
}

func (h *HeatingEquipmentHandler) FeedByID(c *gin.Context) {
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}
	modelItem, err := h.repo.GetPublishedByID(id)
	if c.Query("next") == "true" {
		modelItem, err = h.repo.GetNextPublished(id)
	}
	if err != nil {
		handleRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, serializer.HeatingEquipmentFromModel(*modelItem, currentuser.Get().ID, h.mediaBaseURL, h.repo.MediaBucket()))
}

type likeRequest struct {
	Like *int `json:"like" binding:"required"`
}

func (h *HeatingEquipmentHandler) Like(c *gin.Context) {
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}
	var req likeRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Like == nil || (*req.Like != 0 && *req.Like != 1) {
		c.Status(http.StatusBadRequest)
		return
	}
	item, err := h.repo.SetLike(currentuser.Get().ID, id, *req.Like)
	if err != nil {
		handleRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, serializer.HeatingEquipmentFromModel(*item, currentuser.Get().ID, h.mediaBaseURL, h.repo.MediaBucket()))
}

func (h *HeatingEquipmentHandler) Delete(c *gin.Context) {
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}
	if err := h.repo.SoftDelete(currentuser.Get().ID, id); err != nil {
		handleRepositoryError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

func validateMediaFile(header *multipart.FileHeader, expectedPrefix string) error {
	file, err := header.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		return err
	}
	contentType := http.DetectContentType(buffer[:n])
	if !strings.HasPrefix(contentType, expectedPrefix) {
		return errors.New("unexpected media type")
	}
	// PutObject uses this header value; keep it consistent with detected bytes.
	header.Header.Set("Content-Type", contentType)
	return nil
}

func parsePositiveID(value string) (uint, bool) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

func handleRepositoryError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		c.Status(http.StatusNotFound)
	case errors.Is(err, repository.ErrInvalidLikeValue):
		c.Status(http.StatusBadRequest)
	default:
		c.Status(http.StatusInternalServerError)
	}
}
