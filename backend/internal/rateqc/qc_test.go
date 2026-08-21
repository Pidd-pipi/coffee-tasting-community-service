package rateqc

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/config"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/middleware"
)

func newRateEngine(lim *middleware.RateLimiter) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(lim.Limit())
	r.GET("/probe", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func doRateReq(r *gin.Engine, ip string) int {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.RemoteAddr = ip + ":1234"
	r.ServeHTTP(w, req)
	return w.Code
}

func TestR2ConcurrentRequestsNoRace(t *testing.T) {
	lim := middleware.NewRateLimiter(100000, time.Hour)
	r := newRateEngine(lim)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			for j := 0; j < 30; j++ {
				if code := doRateReq(r, fmt.Sprintf("10.0.0.%d", i)); code != http.StatusOK {
					t.Errorf("concurrent request status = %d", code)
				}
			}
		}(i)
	}
	close(start)
	wg.Wait()
}

func TestR3SecondRequestIsBlocked(t *testing.T) {
	lim := middleware.NewRateLimiter(1, time.Hour)
	r := newRateEngine(lim)
	if code := doRateReq(r, "3.4.5.6"); code != http.StatusOK {
		t.Fatalf("first request status = %d, want 200", code)
	}
	if code := doRateReq(r, "3.4.5.6"); code != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d, want 429", code)
	}
}

func TestR5WindowUsesSeconds(t *testing.T) {
	t.Setenv("RATE_LIMIT_WINDOW_SECONDS", "")
	cfg := config.Load()
	if cfg.RateLimitWin < time.Second {
		t.Fatalf("RateLimitWin = %v, want at least 1s", cfg.RateLimitWin)
	}
}

func TestR6RequestsDefaultIsReasonable(t *testing.T) {
	t.Setenv("RATE_LIMIT_REQUESTS", "")
	cfg := config.Load()
	if cfg.RateLimitReq < 2 {
		t.Fatalf("RateLimitReq = %d, want at least 2", cfg.RateLimitReq)
	}
}

func TestR7FirstRequestNotLimited(t *testing.T) {
	lim := middleware.NewRateLimiter(1, time.Hour)
	r := newRateEngine(lim)
	if code := doRateReq(r, "9.9.9.9"); code != http.StatusOK {
		t.Fatalf("first request status = %d, want 200", code)
	}
}
