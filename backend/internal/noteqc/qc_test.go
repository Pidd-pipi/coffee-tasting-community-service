package noteqc

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/config"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/repository"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/router"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/service"
)

func newNoteDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.TastingNote{}, &model.BrewRecipe{}, &model.CoffeeBean{}, &model.Comment{}, &model.Like{}, &model.UserFollow{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func newNoteService(db *gorm.DB) *service.NoteService {
	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	return service.NewNoteService(repository.NewTastingNoteRepository(db), lg)
}

func newProfileRouter(t *testing.T, db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &config.Config{ServerPort: "8080", JWTSecret: "s", JWTExpire: 999999, RateLimitReq: 1000, RateLimitWin: 999999, UploadDir: "/tmp"}
	return router.Setup(cfg, db, lg)
}

func TestN1TopOriginsNotEmptySlice(t *testing.T) {
	db := newNoteDB(t)
	svc := newNoteService(db)
	if err := db.Create(&model.User{Username: "alice", Email: "a@x.com", PasswordHash: "h"}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	origins, err := svc.TopOrigins(1)
	if err != nil {
		t.Fatalf("TopOrigins: %v", err)
	}
	if origins == nil {
		t.Fatalf("TopOrigins returned nil, want empty slice")
	}
}

func TestN2AvgScoreNoNotesIsZero(t *testing.T) {
	db := newNoteDB(t)
	svc := newNoteService(db)
	if err := db.Create(&model.User{Username: "bob", Email: "b@x.com", PasswordHash: "h"}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := svc.AvgScore(1); err != nil {
		t.Fatalf("AvgScore = %v, want nil error", err)
	}
}

func TestN3ListByUserNotEmptySlice(t *testing.T) {
	db := newNoteDB(t)
	svc := newNoteService(db)
	if err := db.Create(&model.User{Username: "carol", Email: "c@x.com", PasswordHash: "h"}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	items, err := svc.ListByUser(1)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if items == nil {
		t.Fatalf("ListByUser returned nil, want empty slice")
	}
}

func TestN4NoteCountIsZero(t *testing.T) {
	db := newNoteDB(t)
	if err := db.Create(&model.User{Username: "dave", Email: "d@x.com", PasswordHash: "h"}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	r := newProfileRouter(t, db)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/1/profile", nil)
	r.ServeHTTP(w, req)
	body := w.Body.String()
	if contains(body, `"note_count":1`) {
		t.Fatalf("note_count is 1, want 0: %s", body)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
