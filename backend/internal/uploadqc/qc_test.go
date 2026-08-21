package uploadqc

import (
	"bytes"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/config"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/handler"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/middleware"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/util"
)

func newFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fw.Write(content); err != nil {
		t.Fatalf("write: %v", err)
	}
	w.Close()
	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	_, fh, err := req.FormFile("file")
	if err != nil {
		t.Fatalf("form file: %v", err)
	}
	return fh
}

func TestU1ValidUploadSucceeds(t *testing.T) {
	dir := t.TempDir()
	fh := newFileHeader(t, "cup.png", []byte("fakepng"))
	path, err := util.SaveUpload(dir, fh)
	if err != nil {
		t.Fatalf("SaveUpload = %v, want nil", err)
	}
	if path == "" {
		t.Fatalf("SaveUpload path is empty")
	}
}

func TestU2WebpUploadSucceeds(t *testing.T) {
	dir := t.TempDir()
	fh := newFileHeader(t, "cup.webp", []byte("fakewebp"))
	if _, err := util.SaveUpload(dir, fh); err != nil {
		t.Fatalf("SaveUpload webp = %v, want nil", err)
	}
}

func TestU3TenKBUploadSucceeds(t *testing.T) {
	dir := t.TempDir()
	content := bytes.Repeat([]byte("x"), 10*1024)
	fh := newFileHeader(t, "big.png", content)
	if _, err := util.SaveUpload(dir, fh); err != nil {
		t.Fatalf("SaveUpload 10KB = %v, want nil", err)
	}
}

func TestU4HandlerUploadOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &config.Config{UploadDir: t.TempDir()}
	h := handler.NewUploadHandler(cfg, lg)
	r := gin.New()
	r.Use(middleware.ErrorHandler(lg))
	r.POST("/uploads", h.Upload)
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, _ := w.CreateFormFile("file", "cup.png")
	fw.Write([]byte("fakepng"))
	w.Close()
	req := httptest.NewRequest(http.MethodPost, "/uploads", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("upload status = %d, want 200", rec.Code)
	}
}

func TestU5TwoUploadsDistinctFiles(t *testing.T) {
	dir := t.TempDir()
	path1, err := util.SaveUpload(dir, newFileHeader(t, "a.png", []byte("a")))
	if err != nil {
		t.Fatalf("first upload: %v", err)
	}
	path2, err := util.SaveUpload(dir, newFileHeader(t, "b.png", []byte("b")))
	if err != nil {
		t.Fatalf("second upload: %v", err)
	}
	if path1 == path2 {
		t.Fatalf("two uploads got same path %q, want distinct", path1)
	}
}
