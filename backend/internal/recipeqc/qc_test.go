package recipeqc

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

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/handler"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/repository"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/service"
)

func newRecipeDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.BrewRecipe{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func newRecipeService(db *gorm.DB) *service.RecipeService {
	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	return service.NewRecipeService(repository.NewBrewRecipeRepository(db), lg)
}

func names(items []model.BrewRecipe) map[string]bool {
	m := map[string]bool{}
	for _, it := range items {
		m[it.Name] = true
	}
	return m
}

func TestP1EmptyStepRecipeStillShown(t *testing.T) {
	db := newRecipeDB(t)
	svc := newRecipeService(db)
	if _, err := svc.Create(1, &model.BrewRecipe{Name: "empty-steps", Steps: "", WaterTemp: 92}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := svc.Create(1, &model.BrewRecipe{Name: "full-steps", Steps: `[{"step_number":1}]`, WaterTemp: 92}); err != nil {
		t.Fatalf("create: %v", err)
	}
	items, _, err := svc.List("", "", 1, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	got := names(items)
	if !got["empty-steps"] || !got["full-steps"] {
		t.Fatalf("List names = %v, want both empty-steps and full-steps", got)
	}
}

func TestP2ZeroTempRecipeStillShown(t *testing.T) {
	db := newRecipeDB(t)
	svc := newRecipeService(db)
	if _, err := svc.Create(1, &model.BrewRecipe{Name: "cold-brew", Steps: `[{"step_number":1}]`, WaterTemp: 0}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := svc.Create(1, &model.BrewRecipe{Name: "hot", Steps: `[{"step_number":1}]`, WaterTemp: 93}); err != nil {
		t.Fatalf("create: %v", err)
	}
	items, _, err := svc.List("", "", 1, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	got := names(items)
	if !got["cold-brew"] || !got["hot"] {
		t.Fatalf("List names = %v, want both cold-brew and hot", got)
	}
}

func newRecipeRouter(t *testing.T, db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := newRecipeService(db)
	h := handler.NewRecipeHandler(svc, lg)
	r := gin.New()
	r.GET("/recipes", h.List)
	return r
}

func TestP3HandlerTotalCountsAll(t *testing.T) {
	db := newRecipeDB(t)
	for i := 0; i < 15; i++ {
		if err := db.Create(&model.BrewRecipe{UserID: 1, Name: "r", Steps: `[{"step_number":1}]`, WaterTemp: 92}).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
	}
	r := newRecipeRouter(t, db)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/recipes?page=1&page_size=10", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !contains(body, `"total":15`) {
		t.Fatalf("response total missing or wrong: %s", body)
	}
}

func TestP4HandlerPageSizeKept(t *testing.T) {
	db := newRecipeDB(t)
	if err := db.Create(&model.BrewRecipe{UserID: 1, Name: "r", Steps: `[{"step_number":1}]`, WaterTemp: 92}).Error; err != nil {
		t.Fatalf("create: %v", err)
	}
	r := newRecipeRouter(t, db)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/recipes?page=1&page_size=10", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !contains(body, `"page_size":10`) {
		t.Fatalf("response page_size missing or wrong: %s", body)
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


func TestP5GetKeepsAllSteps(t *testing.T) {
	db := newRecipeDB(t)
	svc := newRecipeService(db)
	created, err := svc.Create(1, &model.BrewRecipe{Name: "multi-step", Steps: `[{"step_number":1},{"step_number":2}]`, WaterTemp: 92})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := svc.Get(created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Steps != `[{"step_number":1},{"step_number":2}]` {
		t.Fatalf("Steps = %q, want full JSON", got.Steps)
	}
}
