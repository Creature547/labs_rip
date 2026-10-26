package handler

import (
	"awesomeProject/internal/app/auth"
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
	auth *auth.Service
}

func NewUserHandler(repo repository.UserRepository, service *auth.Service) *UserHandler {
	return &UserHandler{repo: repo, auth: service}
}

type registerRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register godoc
// @Summary Регистрация
// @Tags Пользователи
// @Produce json
// @Accept json
// @Param credentials body registerRequest true "credentials"
// @Success 201 {object} serializer.User
// @Failure 400 "Неверные параметры"
// @Failure 404 "Не найдено"
// @Failure 500 "Ошибка сервера"
// @Failure 409 "Конфликт"
// @Router /users [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	req.Login = strings.TrimSpace(req.Login)
	if req.Login == "" || len(req.Login) > 50 || len(req.Password) < 4 || len(req.Password) > 72 {
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

// Authenticate checks the password, creates a Redis session and issues a linked JWT.
// Authenticate godoc
// @Summary Вход — получить JWT и cookie сессии
// @Tags Пользователи
// @Produce json
// @Accept json
// @Param credentials body registerRequest true "credentials"
// @Success 200 {object} authResponse
// @Failure 400 "Неверные параметры"
// @Failure 404 "Не найдено"
// @Failure 500 "Ошибка сервера"
// @Failure 401 "Неверный логин или пароль"
// @Router /users/auth [post]
func (h *UserHandler) Authenticate(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	req.Login = strings.TrimSpace(req.Login)
	user, err := h.repo.GetUserByLogin(req.Login)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid login or password"})
		} else {
			c.Status(http.StatusInternalServerError)
		}
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid login or password"})
		return
	}
	sessionID, err := h.auth.CreateSession(c.Request.Context(), user.ID, user.Login)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Redis unavailable; session was not created"})
		return
	}
	token, err := h.auth.IssueForSession(user.ID, user.Login, sessionID)
	if err != nil {
		_ = h.auth.DeleteSession(c.Request.Context(), sessionID)
		c.Status(http.StatusInternalServerError)
		return
	}
	if sessionID != "" {
		h.auth.SetSessionCookie(c, sessionID, int(h.auth.ExpiresIn()))
	}
	c.JSON(http.StatusOK, gin.H{"access_token": token, "token_type": "Bearer", "expires_in": h.auth.ExpiresIn(), "user": serializer.UserFromModel(*user)})
}

// Logout deletes the active session and revokes the supplied JWT, if any.
// Logout godoc
// @Summary Выход — удалить сессию и отозвать JWT
// @Tags Пользователи
// @Produce json
// @Success 200 "Успех"
// @Failure 400 "Неверные параметры"
// @Failure 404 "Не найдено"
// @Failure 500 "Ошибка сервера"
// @Security BearerAuth
// @Failure 401 "Не авторизован"
// @Failure 503 "Redis недоступен"
// @Router /users/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	value, ok := c.Get("auth_claims")
	claims, valid := value.(*auth.Claims)
	if !ok || !valid {
		c.Status(http.StatusUnauthorized)
		return
	}
	if claims.ID != "" {
		if err := h.auth.Revoke(c.Request.Context(), claims); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Redis unavailable; token was not revoked"})
			return
		}
	}
	if err := h.auth.DeleteSession(c.Request.Context(), claims.SessionID); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Redis unavailable; session was not deleted"})
		return
	}
	h.auth.SetSessionCookie(c, "", -1)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// authResponse describes the authentication response for Swagger.
type authResponse struct {
	AccessToken string          `json:"access_token"`
	TokenType   string          `json:"token_type"`
	ExpiresIn   int64           `json:"expires_in"`
	User        serializer.User `json:"user"`
}
