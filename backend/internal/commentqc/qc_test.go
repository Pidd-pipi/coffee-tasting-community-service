package commentqc

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/constants"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/middleware"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/repository"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/service"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/util"
)

func newCommentDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.TastingNote{}, &model.Comment{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func newCommentService(db *gorm.DB) *service.CommentService {
	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	return service.NewCommentService(repository.NewCommentRepository(db), repository.NewTastingNoteRepository(db), lg)
}

func createCommentUser(t *testing.T, db *gorm.DB, name string) *model.User {
	t.Helper()
	u := &model.User{Username: name, Email: name + "@test.local", PasswordHash: "h", Role: constants.RoleUser}
	if err := db.Create(u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u
}

func createCommentNote(t *testing.T, db *gorm.DB, uid uint) *model.TastingNote {
	t.Helper()
	n := &model.TastingNote{UserID: uid, CoffeeName: "yirgacheffe", RoastLevel: constants.RoastLight, FlavorTags: "[]"}
	if err := db.Create(n).Error; err != nil {
		t.Fatalf("create note: %v", err)
	}
	return n
}

func TestC1DeleteMissingCommentIs404(t *testing.T) {
	db := newCommentDB(t)
	svc := newCommentService(db)
	u := createCommentUser(t, db, "alice")
	err := svc.Delete(u.ID, 424242)
	var appErr *util.AppError
	if !errors.As(err, &appErr) || appErr.HTTPStatus != http.StatusNotFound {
		t.Fatalf("Delete missing comment = %v, want 404 AppError", err)
	}
}

func TestC2CreateCommentOnMissingNoteIs404(t *testing.T) {
	db := newCommentDB(t)
	svc := newCommentService(db)
	u := createCommentUser(t, db, "bob")
	_, err := svc.Create(u.ID, 424242, "hello")
	var appErr *util.AppError
	if !errors.As(err, &appErr) || appErr.HTTPStatus != http.StatusNotFound {
		t.Fatalf("Create comment on missing note = %v, want 404 AppError", err)
	}
}

func TestC3NonOwnerDeleteIs403(t *testing.T) {
	db := newCommentDB(t)
	svc := newCommentService(db)
	owner := createCommentUser(t, db, "owner")
	other := createCommentUser(t, db, "other")
	note := createCommentNote(t, db, owner.ID)
	cm, err := svc.Create(owner.ID, note.ID, "hello")
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}
	err = svc.Delete(other.ID, cm.ID)
	var appErr *util.AppError
	if !errors.As(err, &appErr) || appErr.HTTPStatus != http.StatusForbidden {
		t.Fatalf("non-owner delete = %v, want 403 AppError", err)
	}
}

func TestC4WrappedAppErrorKeepsStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	r := gin.New()
	r.Use(middleware.ErrorHandler(lg))
	r.GET("/probe", func(c *gin.Context) {
		c.Error(fmt.Errorf("wrapped: %w", util.NewAppError(http.StatusNotFound, constants.CodeNotFound, "missing")))
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("wrapped AppError status = %d, want 404", w.Code)
	}
}
