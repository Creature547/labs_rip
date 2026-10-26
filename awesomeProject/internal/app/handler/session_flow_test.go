package handler

import (
	"awesomeProject/internal/app/auth"
	"awesomeProject/internal/app/model"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type sessionRepo struct{ user model.User }

func (r *sessionRepo) RegisterUser(string, string) (*model.User, error) { return &r.user, nil }
func (r *sessionRepo) GetUserByLogin(string) (*model.User, error)       { return &r.user, nil }

type sessionMemory struct{ items map[string]auth.SessionUser }

func (m *sessionMemory) Put(_ context.Context, id string, u auth.SessionUser, _ time.Duration) error {
	m.items[id] = u
	return nil
}
func (m *sessionMemory) Get(_ context.Context, id string) (auth.SessionUser, error) {
	u, ok := m.items[id]
	if !ok {
		return u, auth.ErrInvalidToken
	}
	return u, nil
}
func (m *sessionMemory) Delete(_ context.Context, id string) error { delete(m.items, id); return nil }

type sessionBlacklist struct{ items map[string]bool }

func (m *sessionBlacklist) Contains(_ context.Context, id string) (bool, error) {
	return m.items[id], nil
}
func (m *sessionBlacklist) Add(_ context.Context, id, _ string, _ time.Duration) error {
	m.items[id] = true
	return nil
}
func TestLoginCookieAndLogoutFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, useJWT := range []bool{false, true} {
		t.Run(map[bool]string{false: "cookie logout", true: "JWT logout"}[useJWT], func(t *testing.T) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("lab4pass"), bcrypt.MinCost)
			repo := &sessionRepo{user: model.User{ID: 6, Login: "anna_lab4", PasswordHash: string(hash)}}
			sessions := &sessionMemory{items: map[string]auth.SessionUser{}}
			blacklist := &sessionBlacklist{items: map[string]bool{}}
			service, _ := auth.New("01234567890123456789012345678901", blacklist, time.Hour)
			service.SetSessions(sessions)
			h := NewUserHandler(repo, service)
			r := gin.New()
			r.POST("/login", h.Authenticate)
			r.POST("/logout", service.Required(), h.Logout)
			r.GET("/private", service.Required(), func(c *gin.Context) { c.Status(200) })
			w := httptest.NewRecorder()
			q := httptest.NewRequest("POST", "/login", strings.NewReader(`{"login":"anna_lab4","password":"lab4pass"}`))
			q.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, q)
			if w.Code != 200 {
				t.Fatalf("login: %d %s", w.Code, w.Body.String())
			}
			var body struct {
				Token string `json:"access_token"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			cookies := w.Result().Cookies()
			if len(cookies) != 1 {
				t.Fatal("missing session cookie")
			}
			cookie := cookies[0]
			if sessions.items[cookie.Value].Login != "anna_lab4" {
				t.Fatal("session is not linked to login")
			}
			q = httptest.NewRequest("POST", "/logout", nil)
			if useJWT {
				q.Header.Set("Authorization", "Bearer "+body.Token)
			} else {
				q.AddCookie(cookie)
			}
			w = httptest.NewRecorder()
			r.ServeHTTP(w, q)
			if w.Code != 200 {
				t.Fatalf("logout: %d %s", w.Code, w.Body.String())
			}
			if len(sessions.items) != 0 {
				t.Fatal("session was not deleted")
			}
			if c := w.Result().Cookies(); len(c) != 1 || c[0].MaxAge != -1 {
				t.Fatal("cookie was not expired")
			}
			if useJWT && len(blacklist.items) != 1 {
				t.Fatal("JWT was not blacklisted")
			}
			for _, byJWT := range []bool{false, true} {
				q = httptest.NewRequest(http.MethodGet, "/private", nil)
				if byJWT {
					q.Header.Set("Authorization", "Bearer "+body.Token)
				} else {
					q.AddCookie(cookie)
				}
				w = httptest.NewRecorder()
				r.ServeHTTP(w, q)
				if w.Code != 401 {
					t.Fatalf("old credential accepted: %d", w.Code)
				}
			}
		})
	}
}
