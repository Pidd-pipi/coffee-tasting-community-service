package followqc

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/handler"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/middleware"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/repository"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/service"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/util"
)

func newFollowDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.UserFollow{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func newFollowService(db *gorm.DB) *service.FollowService {
	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	return service.NewFollowService(repository.NewUserFollowRepository(db), lg)
}

func TestF1SelfFollowDisallowed(t *testing.T) {
	db := newFollowDB(t)
	svc := newFollowService(db)
	_, err := svc.Follow(1, 1)
	var appErr *util.AppError
	if !errors.As(err, &appErr) || appErr.HTTPStatus != 422 {
		t.Fatalf("Follow self = %v, want 422 AppError", err)
	}
}

func TestF2UnfollowRemovesEdge(t *testing.T) {
	db := newFollowDB(t)
	svc := newFollowService(db)
	if _, err := svc.Follow(1, 2); err != nil {
		t.Fatalf("follow: %v", err)
	}
	if err := svc.Unfollow(1, 2); err != nil {
		t.Fatalf("unfollow: %v", err)
	}
	followers, following, err := svc.Counts(2)
	if err != nil {
		t.Fatalf("counts: %v", err)
	}
	if followers != 0 || following != 0 {
		t.Fatalf("counts(2) = %d/%d, want 0/0", followers, following)
	}
}

func TestF3CountFollowers(t *testing.T) {
	db := newFollowDB(t)
	svc := newFollowService(db)
	if _, err := svc.Follow(2, 1); err != nil {
		t.Fatalf("follow: %v", err)
	}
	followers, _, err := svc.Counts(1)
	if err != nil {
		t.Fatalf("counts: %v", err)
	}
	if followers != 1 {
		t.Fatalf("followers(1) = %d, want 1", followers)
	}
}

func TestF4CountFollowing(t *testing.T) {
	db := newFollowDB(t)
	svc := newFollowService(db)
	if _, err := svc.Follow(1, 2); err != nil {
		t.Fatalf("follow: %v", err)
	}
	_, following, err := svc.Counts(1)
	if err != nil {
		t.Fatalf("counts: %v", err)
	}
	if following != 1 {
		t.Fatalf("following(1) = %d, want 1", following)
	}
}

func TestF5HandlerFollowDirection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newFollowDB(t)
	svc := newFollowService(db)
	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	h := handler.NewFollowHandler(svc, lg)
	r := gin.New()
	r.Use(middleware.ErrorHandler(lg))
	r.POST("/users/:id/follow", func(c *gin.Context) {
		c.Set(middleware.UserKey, &util.Claims{UserID: 1})
		h.Follow(c)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users/2/follow", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("follow status = %d, want 201", w.Code)
	}
	var count int64
	if err := db.Model(&model.UserFollow{}).Where("follower_id = ? AND following_id = ?", 1, 2).Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("edge (1->2) count = %d, want 1", count)
	}
}
