package beanqc

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

func newBeanDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.CoffeeBean{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func newBeanService(db *gorm.DB) *service.BeanService {
	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	return service.NewBeanService(repository.NewCoffeeBeanRepository(db), lg)
}

func TestB1UpdateKeepsNameOnEmpty(t *testing.T) {
	db := newBeanDB(t)
	svc := newBeanService(db)
	created, err := svc.Create(&model.CoffeeBean{Name: "yirgacheffe", ProcessMethod: "washed", FlavorTags: "[]"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	updated, err := svc.Update(created.ID, &model.CoffeeBean{Name: "", ProcessMethod: "washed", Description: "new desc"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "yirgacheffe" {
		t.Fatalf("Name = %q, want preserved %q", updated.Name, "yirgacheffe")
	}
}

func TestB2CreateRejectsEmptyProcess(t *testing.T) {
	db := newBeanDB(t)
	svc := newBeanService(db)
	_, err := svc.Create(&model.CoffeeBean{Name: "x", ProcessMethod: "", FlavorTags: "[]"})
	var appErr *util.AppError
	if !errors.As(err, &appErr) || appErr.HTTPStatus != 422 {
		t.Fatalf("Create empty process = %v, want 422 AppError", err)
	}
}

func TestB3UpdateMissingBeanErrors(t *testing.T) {
	db := newBeanDB(t)
	svc := newBeanService(db)
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				t.Fatalf("Update missing bean panicked: %v", rec)
			}
		}()
		if _, err := svc.Update(99999, &model.CoffeeBean{Name: "x", ProcessMethod: "washed"}); err == nil {
			t.Fatalf("Update missing bean = nil error, want error")
		}
	}()
}

func TestB4UpdateKeepsProcessOnEmpty(t *testing.T) {
	db := newBeanDB(t)
	svc := newBeanService(db)
	created, err := svc.Create(&model.CoffeeBean{Name: "yirgacheffe", ProcessMethod: "washed", FlavorTags: "[]"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	updated, err := svc.Update(created.ID, &model.CoffeeBean{Name: "yirgacheffe", ProcessMethod: "", Description: "d"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.ProcessMethod != "washed" {
		t.Fatalf("ProcessMethod = %q, want preserved %q", updated.ProcessMethod, "washed")
	}
}
