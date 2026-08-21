package rbacqc

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/config"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/constants"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/middleware"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/router"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/util"
)

func TestK1ForbiddenStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.UserKey, &util.Claims{UserID: 2, Role: "user"})
		c.Next()
	})
	r.Use(middleware.RequireRole("admin"))
	r.GET("/probe", func(c *gin.Context) { c.Status(http.StatusOK) })
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("forbidden status = %d, want 403", w.Code)
	}
}

func TestK3AdminConstantCorrect(t *testing.T) {
	if constants.RoleAdmin != "admin" {
		t.Fatalf("RoleAdmin = %q, want %q", constants.RoleAdmin, "admin")
	}
}

func TestK5UserConstantCorrect(t *testing.T) {
	if constants.RoleUser != "user" {
		t.Fatalf("RoleUser = %q, want %q", constants.RoleUser, "user")
	}
}

func TestK4UserCannotCreateBean(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.TastingNote{}, &model.BrewRecipe{}, &model.CoffeeBean{}, &model.Comment{}, &model.Like{}, &model.UserFollow{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &config.Config{ServerPort: "8080", JWTSecret: "s", JWTExpire: time.Hour, RateLimitReq: 1000, RateLimitWin: time.Hour, UploadDir: "/tmp"}
	r := router.Setup(cfg, db, lg)
	token, err := util.GenerateToken(2, "bob", "user", "s", time.Hour)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/beans", bytes.NewBufferString(`{"name":"test","process_method":"washed"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("user create bean status = %d, want 403", w.Code)
	}
}

func TestK6ValidRolesContainsAdmin(t *testing.T) {
	roles := constants.ValidRoles()
	for _, r := range roles {
		if r == "admin" {
			return
		}
	}
	t.Fatalf("ValidRoles() = %v, want contain admin", roles)
}

func TestK7BeanListIsPublic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.TastingNote{}, &model.BrewRecipe{}, &model.CoffeeBean{}, &model.Comment{}, &model.Like{}, &model.UserFollow{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &config.Config{ServerPort: "8080", JWTSecret: "s", JWTExpire: time.Hour, RateLimitReq: 1000, RateLimitWin: time.Hour, UploadDir: "/tmp"}
	r := router.Setup(cfg, db, lg)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/beans", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("bean list status = %d, want 200", w.Code)
	}
}
