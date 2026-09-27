package handler

import (
	"awesomeProject/internal/app/repository"
	"awesomeProject/internal/app/serializer"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	repo repository.UserRepository
}

func NewUserHandler(repo repository.UserRepository) *UserHandler { return &UserHandler{repo: repo} }

type registerRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	req.Login = strings.TrimSpace(req.Login)
	if req.Login == "" || len(req.Password) < 4 {
		c.Status(http.StatusBadRequest)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	user, err := h.repo.RegisterUser(req.Login, string(hash))
	if err != nil {
		if errors.Is(err, repository.ErrLoginExists) {
			c.Status(http.StatusConflict)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusCreated, serializer.UserFromModel(*user))
}

// Authenticate is a stub for lab 4.
func (h *UserHandler) Authenticate(c *gin.Context) { c.Status(http.StatusOK) }

// Logout is a stub for lab 4.
func (h *UserHandler) Logout(c *gin.Context) { c.Status(http.StatusOK) }
