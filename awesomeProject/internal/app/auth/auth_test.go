package auth

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type fakeBlacklist struct {
	blocked     map[string]bool
	unavailable bool
	lastTTL     time.Duration
}

func (b *fakeBlacklist) Contains(_ context.Context, id string) (bool, error) {
	if b.unavailable {
		return false, errors.New("Redis unavailable")
	}
	return b.blocked[id], nil
}
func (b *fakeBlacklist) Add(_ context.Context, id, _ string, ttl time.Duration) error {
	if b.unavailable {
		return errors.New("Redis unavailable")
	}
	b.blocked[id] = true
	b.lastTTL = ttl
	return nil
}
func TestTokenRevocationAndIsolation(t *testing.T) {
	b := &fakeBlacklist{blocked: map[string]bool{}}
	s, err := New("01234567890123456789012345678901", b, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	a, err := s.Issue(7, "anna")
	if err != nil {
		t.Fatal(err)
	}
	other, err := s.Issue(7, "anna")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := s.Verify(context.Background(), a)
	if err != nil || claims.UserID != 7 {
		t.Fatalf("identity: %v", err)
	}
	if err = s.Revoke(context.Background(), claims); err != nil {
		t.Fatal(err)
	}
	if b.lastTTL <= 0 || b.lastTTL > time.Hour {
		t.Fatal("incorrect blacklist TTL")
	}
	if _, err = s.Verify(context.Background(), a); !errors.Is(err, ErrInvalidToken) {
		t.Fatal("logged-out token accepted")
	}
	if _, err = s.Verify(context.Background(), other); err != nil {
		t.Fatal("other login session revoked")
	}
}
func TestPublicAndPrivateAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	b := &fakeBlacklist{blocked: map[string]bool{}}
	s, _ := New("01234567890123456789012345678901", b, time.Hour)
	token, _ := s.Issue(7, "anna")
	r := gin.New()
	r.GET("/public", s.Optional(), func(c *gin.Context) { c.Status(200) })
	r.PUT("/private", s.Required(), func(c *gin.Context) { c.Status(200) })
	tests := []struct {
		path, method, header string
		status               int
		down                 bool
	}{
		{"/public", "GET", "", 200, false},
		{"/public", "GET", "Bearer invalid", 200, false},
		{"/private", "PUT", "", 401, false},
		{"/private", "PUT", "Bearer invalid", 401, false},
		{"/private", "PUT", "Bearer " + token, 200, false},
		{"/private", "PUT", "Bearer " + token, 503, true},
		{"/public", "GET", "Bearer " + token, 200, true},
	}
	for _, test := range tests {
		b.unavailable = test.down
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(test.method, test.path, nil)
		req.Header.Set("Authorization", test.header)
		r.ServeHTTP(recorder, req)
		if recorder.Code != test.status {
			t.Fatalf("%s: got %d want %d", test.path, recorder.Code, test.status)
		}
	}
}
func TestRejectWrongSignatureAlgorithmAndExpiry(t *testing.T) {
	secret := "01234567890123456789012345678901"
	s, _ := New(secret, &fakeBlacklist{blocked: map[string]bool{}}, time.Hour)
	raw, _ := s.Issue(7, "anna")
	other, _ := New("abcdefghijklmnopqrstuvwxyz012345", &fakeBlacklist{blocked: map[string]bool{}}, time.Hour)
	if _, err := other.Verify(context.Background(), raw); err == nil {
		t.Fatal("wrong secret accepted")
	}
	claims := Claims{UserID: 7, Login: "anna", Role: "creator", RegisteredClaims: jwt.RegisteredClaims{
		Issuer: Issuer, Subject: "7", ID: "abc", ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
	}}
	expired, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if _, err := s.Verify(context.Background(), expired); err == nil {
		t.Fatal("expired token accepted")
	}
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Hour))
	wrongAlgorithm, _ := jwt.NewWithClaims(jwt.SigningMethodHS384, claims).SignedString([]byte(secret))
	if _, err := s.Verify(context.Background(), wrongAlgorithm); err == nil {
		t.Fatal("unexpected algorithm accepted")
	}
	claims.ExpiresAt = nil
	noExpiry, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if _, err := s.Verify(context.Background(), noExpiry); err == nil {
		t.Fatal("token without expiry accepted")
	}
}
