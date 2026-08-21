package likeqc

import (
	"errors"
	"log/slog"
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/repository"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/service"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/util"
)

func newLikeDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.TastingNote{}, &model.Like{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func newLikeService(db *gorm.DB) *service.LikeService {
	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	return service.NewLikeService(repository.NewLikeRepository(db), repository.NewTastingNoteRepository(db), lg)
}

func newLikeNote(t *testing.T, db *gorm.DB, uid uint) *model.TastingNote {
	t.Helper()
	n := &model.TastingNote{UserID: uid, CoffeeName: "yirgacheffe", RoastLevel: "light", FlavorTags: "[]"}
	if err := db.Create(n).Error; err != nil {
		t.Fatalf("create note: %v", err)
	}
	return n
}

func TestL1LikeAbsentNoteDenied(t *testing.T) {
	db := newLikeDB(t)
	svc := newLikeService(db)
	_, err := svc.Like(1, 99999)
	var appErr *util.AppError
	if !errors.As(err, &appErr) || appErr.HTTPStatus != 404 {
		t.Fatalf("Like absent note = %v, want 404 AppError", err)
	}
}

func TestL2UnlikeMissingNotFound(t *testing.T) {
	db := newLikeDB(t)
	svc := newLikeService(db)
	err := svc.Unlike(1, 99999)
	var appErr *util.AppError
	if !errors.As(err, &appErr) || appErr.HTTPStatus != 404 {
		t.Fatalf("Unlike missing = %v, want 404 AppError", err)
	}
}

func TestL3CountByNote(t *testing.T) {
	db := newLikeDB(t)
	svc := newLikeService(db)
	note := newLikeNote(t, db, 1)
	if _, err := svc.Like(2, note.ID); err != nil {
		t.Fatalf("like: %v", err)
	}
	count, err := svc.CountByNote(note.ID)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("CountByNote = %d, want 1", count)
	}
}

func TestL4CountByUserNotes(t *testing.T) {
	db := newLikeDB(t)
	svc := newLikeService(db)
	note := newLikeNote(t, db, 1)
	if _, err := svc.Like(2, note.ID); err != nil {
		t.Fatalf("like: %v", err)
	}
	count, err := svc.CountByUserNotes(1)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("CountByUserNotes = %d, want 1", count)
	}
}
