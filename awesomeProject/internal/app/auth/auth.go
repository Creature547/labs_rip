package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

const Issuer = "heating-api"

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	UserID    uint   `json:"user_id"`
	Login     string `json:"login"`
	Role      string `json:"role"`
	SessionID string `json:"sid,omitempty"`
	jwt.RegisteredClaims
}

type Blacklist interface {
	Contains(context.Context, string) (bool, error)
	Add(context.Context, string, string, time.Duration) error
}

type RedisBlacklist struct{ Client *redis.Client }

func (b *RedisBlacklist) Contains(ctx context.Context, id string) (bool, error) {
	count, err := b.Client.Exists(ctx, "blacklist:"+id).Result()
	return count > 0, err
}
func (b *RedisBlacklist) Add(ctx context.Context, id, login string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	return b.Client.Set(ctx, "blacklist:"+id, login, ttl).Err()
}

type Service struct {
	secret    []byte
	blacklist Blacklist
	ttl       time.Duration
	sessions  SessionStore
}

func New(secret string, blacklist Blacklist, ttl time.Duration) (*Service, error) {
	if len(secret) < 32 || blacklist == nil || ttl < time.Second {
		return nil, errors.New("JWT_SECRET must have at least 32 bytes; blacklist and positive TTL required")
	}
	return &Service{secret: []byte(secret), blacklist: blacklist, ttl: ttl}, nil
}
func (s *Service) Issue(userID uint, login string) (string, error) {
	return s.IssueForSession(userID, login, "")
}
func (s *Service) IssueForSession(userID uint, login, sessionID string) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	now := time.Now()
	claims := Claims{UserID: userID, Login: login, Role: "creator", SessionID: sessionID, RegisteredClaims: jwt.RegisteredClaims{
		Issuer: Issuer, Subject: strconv.FormatUint(uint64(userID), 10),
		ID: hex.EncodeToString(nonce), IssuedAt: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
	}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}
func (s *Service) Verify(ctx context.Context, raw string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (interface{}, error) {
		return s.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(Issuer), jwt.WithExpirationRequired())
	if err != nil || token == nil || !token.Valid || claims.UserID == 0 || claims.ID == "" ||
		claims.Subject != strconv.FormatUint(uint64(claims.UserID), 10) || claims.Role != "creator" {
		return nil, ErrInvalidToken
	}
	blocked, err := s.blacklist.Contains(ctx, claims.ID)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, ErrInvalidToken
	}
	if claims.SessionID != "" {
		if s.sessions == nil {
			return nil, ErrInvalidToken
		}
		user, err := s.sessions.Get(ctx, claims.SessionID)
		if err != nil {
			return nil, err
		}
		if user.ID != claims.UserID || user.Login != claims.Login {
			return nil, ErrInvalidToken
		}
	}
	return claims, nil
}
func (s *Service) Revoke(ctx context.Context, claims *Claims) error {
	return s.blacklist.Add(ctx, claims.ID, claims.Login, time.Until(claims.ExpiresAt.Time))
}
func (s *Service) ExpiresIn() int64 { return int64(s.ttl / time.Second) }

// Optional preserves public reading even with an expired or malformed token.
func (s *Service) Optional() gin.HandlerFunc {
	return func(c *gin.Context) {
		if claims, err := s.requestClaims(c); err == nil {
			c.Set("auth_claims", claims)
			c.Set("user_id", claims.UserID)
		}
		c.Next()
	}
}
func (s *Service) Required() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := s.requestClaims(c)
		if err != nil {
			status := http.StatusServiceUnavailable
			if errors.Is(err, ErrInvalidToken) {
				status = http.StatusUnauthorized
			}
			c.AbortWithStatusJSON(status, gin.H{"error": http.StatusText(status)})
			return
		}
		c.Set("auth_claims", claims)
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
func bearer(header string) string {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
