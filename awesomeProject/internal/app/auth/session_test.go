package auth

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type memorySessions struct {
	users map[string]SessionUser
	down  bool
	ttl   time.Duration
}

func (m *memorySessions) Put(_ context.Context, id string, u SessionUser, ttl time.Duration) error {
	if m.down {
		return errors.New("unavailable")
	}
	m.users[id] = u
	m.ttl = ttl
	return nil
}
func (m *memorySessions) Get(_ context.Context, id string) (SessionUser, error) {
	if m.down {
		return SessionUser{}, errors.New("unavailable")
	}
	u, ok := m.users[id]
	if !ok {
		return u, ErrInvalidToken
	}
	return u, nil
}
func (m *memorySessions) Delete(_ context.Context, id string) error { delete(m.users, id); return nil }
func TestSessionCookieAndLinkedJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &memorySessions{users: map[string]SessionUser{}}
	s, _ := New("01234567890123456789012345678901", &fakeBlacklist{blocked: map[string]bool{}}, time.Hour)
	s.SetSessions(store)
	sid, err := s.CreateSession(context.Background(), 6, "anna_lab4")
	if err != nil {
		t.Fatal(err)
	}
	other, _ := s.CreateSession(context.Background(), 6, "anna_lab4")
	if len(sid) != 43 || sid == other || store.ttl != time.Hour {
		t.Fatal("bad session randomness or TTL")
	}
	token, _ := s.IssueForSession(6, "anna_lab4", sid)
	r := gin.New()
	r.GET("/public", s.Optional(), func(c *gin.Context) { c.Status(200) })
	r.PUT("/private", s.Required(), func(c *gin.Context) {
		if c.GetUint("user_id") != 6 {
			t.Error("wrong session user")
		}
		c.Status(200)
	})
	request := func(cookie, header, path, method string, want int) {
		t.Helper()
		q := httptest.NewRequest(method, path, nil)
		if cookie != "" {
			q.AddCookie(&http.Cookie{Name: SessionCookie, Value: cookie})
		}
		if header != "" {
			q.Header.Set("Authorization", header)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, q)
		if w.Code != want {
			t.Fatalf("%s got %d want %d", path, w.Code, want)
		}
	}
	request(sid, "", "/private", "PUT", 200)
	request("", "Bearer "+token, "/private", "PUT", 200)
	request(sid, "Bearer invalid", "/private", "PUT", 401)
	request("unknown", "", "/private", "PUT", 401)
	request("", "", "/private", "PUT", 401)
	if err := s.DeleteSession(context.Background(), sid); err != nil {
		t.Fatal(err)
	}
	request(sid, "", "/private", "PUT", 401)
	request("", "Bearer "+token, "/private", "PUT", 401)
	request(other, "", "/private", "PUT", 200)
	store.down = true
	request(other, "", "/private", "PUT", 503)
	request(other, "", "/public", "GET", 200)
}
func TestSessionCookieAttributes(t *testing.T) {
	s, _ := New("01234567890123456789012345678901", &fakeBlacklist{blocked: map[string]bool{}}, time.Hour)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "https://example.test/", nil)
	s.SetSessionCookie(c, "opaque-session-id", 3600)
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatal("no cookie")
	}
	k := cookies[0]
	if k.Name != SessionCookie || !k.HttpOnly || !k.Secure || k.SameSite != http.SameSiteLaxMode || k.Path != "/" || k.MaxAge != 3600 {
		t.Fatal("incorrect cookie attributes")
	}
}
