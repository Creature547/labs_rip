package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const SessionCookie = "session_id"

type SessionUser struct {
	ID    uint
	Login string
}
type SessionStore interface {
	Put(context.Context, string, SessionUser, time.Duration) error
	Get(context.Context, string) (SessionUser, error)
	Delete(context.Context, string) error
}

// Redis contains only session:<random id> = login, with a TTL.
type RedisSessions struct {
	Client *redis.Client
	Lookup func(string) (uint, error)
}

func (r *RedisSessions) Put(ctx context.Context, id string, user SessionUser, ttl time.Duration) error {
	return r.Client.Set(ctx, "session:"+id, user.Login, ttl).Err()
}
func (r *RedisSessions) Get(ctx context.Context, id string) (SessionUser, error) {
	login, err := r.Client.Get(ctx, "session:"+id).Result()
	if errors.Is(err, redis.Nil) {
		return SessionUser{}, ErrInvalidToken
	}
	if err != nil {
		return SessionUser{}, err
	}
	uid, err := r.Lookup(login)
	if err != nil {
		return SessionUser{}, err
	}
	if uid == 0 {
		return SessionUser{}, ErrInvalidToken
	}
	return SessionUser{ID: uid, Login: login}, nil
}
func (r *RedisSessions) Delete(ctx context.Context, id string) error {
	return r.Client.Del(ctx, "session:"+id).Err()
}
func (s *Service) SetSessions(store SessionStore) { s.sessions = store }
func (s *Service) CreateSession(ctx context.Context, uid uint, login string) (string, error) {
	if s.sessions == nil {
		return "", nil
	}
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	id := base64.RawURLEncoding.EncodeToString(nonce)
	if err := s.sessions.Put(ctx, id, SessionUser{ID: uid, Login: login}, s.ttl); err != nil {
		return "", err
	}
	return id, nil
}
func (s *Service) DeleteSession(ctx context.Context, id string) error {
	if id == "" || s.sessions == nil {
		return nil
	}
	return s.sessions.Delete(ctx, id)
}
func (s *Service) SetSessionCookie(c *gin.Context, id string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(SessionCookie, id, maxAge, "/", "", c.Request.TLS != nil, true)
}
func (s *Service) requestClaims(c *gin.Context) (*Claims, error) {
	// An explicit Authorization header takes precedence; no silent cookie fallback.
	if header := c.GetHeader("Authorization"); header != "" {
		raw := bearer(header)
		if raw == "" {
			return nil, ErrInvalidToken
		}
		return s.Verify(c.Request.Context(), raw)
	}
	id, err := c.Cookie(SessionCookie)
	if err != nil || id == "" || s.sessions == nil {
		return nil, ErrInvalidToken
	}
	user, err := s.sessions.Get(c.Request.Context(), id)
	if err != nil {
		return nil, err
	}
	return &Claims{UserID: user.ID, Login: user.Login, Role: "creator", SessionID: id}, nil
}
