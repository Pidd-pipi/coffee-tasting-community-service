package tokenqc

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/config"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/middleware"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/util"
)

func TestT1GeneratedTokenIsValid(t *testing.T) {
	token, err := util.GenerateToken(1, "alice", "user", "test-secret", time.Hour)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, err := util.ParseToken(token, "test-secret"); err != nil {
		t.Fatalf("ParseToken = %v, want nil", err)
	}
}

func TestT2ExpiredTokenIsExpired(t *testing.T) {
	claims := util.Claims{
		UserID: 1, Username: "alice", Role: "user",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour))},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	_, err = util.ParseToken(token, "test-secret")
	if !errors.Is(err, util.ErrTokenExpired) {
		t.Fatalf("ParseToken expired = %v, want ErrTokenExpired", err)
	}
}

func TestT3HMACTokenAccepted(t *testing.T) {
	claims := util.Claims{
		UserID: 1, Username: "alice", Role: "user",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if _, err := util.ParseToken(token, "test-secret"); err != nil {
		t.Fatalf("ParseToken HMAC = %v, want nil", err)
	}
}

func TestT4AuthAcceptsValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{JWTSecret: "test-secret"}
	claims := util.Claims{
		UserID: 1, Username: "alice", Role: "user",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	r := gin.New()
	r.Use(middleware.AuthRequired(cfg))
	r.GET("/probe", func(c *gin.Context) { c.Status(http.StatusOK) })
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("auth status = %d, want 200", w.Code)
	}
}
