package httpapi

import (
	"github.com/zhangkui/go-home-inventory/internal/inventory"
	"github.com/zhangkui/go-home-inventory/internal/location"
	"github.com/zhangkui/go-home-inventory/internal/reminder"
	"github.com/zhangkui/go-home-inventory/internal/repair"
	"github.com/zhangkui/go-home-inventory/internal/warranty"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealth(t *testing.T) {
	store := inventory.NewStore(time.Now)
	warranties := warranty.NewService(store, time.UTC, time.Now)
	handler := NewServer(store, location.NewService(store), warranties, repair.NewService(store, time.Now), reminder.NewService(store, warranties, time.UTC, time.Now), time.UTC)
	request, response := httptest.NewRequest(http.MethodGet, "/healthz", nil), httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
}
