package reminder

import (
	"context"
	"errors"
	"github.com/zhangkui/go-home-inventory/internal/domain"
	"github.com/zhangkui/go-home-inventory/internal/inventory"
	"github.com/zhangkui/go-home-inventory/internal/warranty"
	"testing"
	"time"
)

func TestVerificationCanceledSummaryStopsBeforeWork(t *testing.T) {
	store := inventory.NewStore(time.Now)
	if _, err := store.Create(domain.Item{Name: "Router", SerialNumber: "R-1", PurchaseDate: time.Now(), WarrantyMonths: 12}); err != nil {
		t.Fatal(err)
	}
	service := NewService(store, warranty.NewService(store, time.UTC, time.Now), time.UTC, time.Now)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := service.BatchSummary(ctx, time.Now().Add(-time.Hour), time.Now().AddDate(2, 0, 0))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}
