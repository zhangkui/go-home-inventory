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

// TestCanceledContextStopsWithEmptyInventory verifies that a pre-canceled
// context is reported even when the inventory is empty. With no items the
// iteration body never runs, so only the pre-scan context check can surface
// the cancellation — the "空库存也如此" case.
func TestCanceledContextStopsWithEmptyInventory(t *testing.T) {
	store := inventory.NewStore(time.Now)
	service := NewService(store, warranty.NewService(store, time.UTC, time.Now), time.UTC, time.Now)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := service.BatchSummary(ctx, time.Now().Add(-time.Hour), time.Now().AddDate(2, 0, 0))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

// midScanCancelContext reports a live context the first time Err() is called
// (so the pre-scan check passes) and a canceled context thereafter. This
// deterministically models a caller canceling mid-scan without relying on
// goroutine timing, isolating the per-iteration context check.
type midScanCancelContext struct {
	context.Context
	calls int
}

func (c *midScanCancelContext) Err() error {
	c.calls++
	if c.calls == 1 {
		return nil
	}
	return context.Canceled
}

// TestScanCancellationReportedMidIteration verifies that cancellation is
// detected during iteration, not only before it. This fails if the
// per-iteration ctx.Err() check is removed (even with the pre-scan check
// intact), so it guards the "遍历中及时返回" behavior.
func TestScanCancellationReportedMidIteration(t *testing.T) {
	store := inventory.NewStore(time.Now)
	if _, err := store.Create(domain.Item{Name: "Router", SerialNumber: "R-1", PurchaseDate: time.Now(), WarrantyMonths: 12}); err != nil {
		t.Fatal(err)
	}
	service := NewService(store, warranty.NewService(store, time.UTC, time.Now), time.UTC, time.Now)
	ctx := &midScanCancelContext{Context: context.Background()}
	_, err := service.BatchSummary(ctx, time.Now().Add(-time.Hour), time.Now().AddDate(2, 0, 0))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if ctx.calls < 2 {
		t.Fatalf("ctx.Err() invoked %d times, want >=2 (pre-scan + per-iteration)", ctx.calls)
	}
}

// TestBatchSummaryNormalCompletion verifies the unchanged behavior of range
// validation, filtering, sorting, counting, and content once the fix is in
// place: items are filtered to the window, sorted by expiry, and counted.
func TestBatchSummaryNormalCompletion(t *testing.T) {
	store := inventory.NewStore(time.Now)
	// Created out of expiry order so the result sort is actually exercised.
	for _, item := range []domain.Item{
		{Name: "Beta", SerialNumber: "s-beta", PurchaseDate: time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC), WarrantyMonths: 12},  // expires 2025-03-10
		{Name: "Alpha", SerialNumber: "s-alpha", PurchaseDate: time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC), WarrantyMonths: 12}, // expires 2025-02-20
		{Name: "Gamma", SerialNumber: "s-gamma", PurchaseDate: time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC), WarrantyMonths: 12},   // expires 2025-08-01 (out of range)
	} {
		if _, err := store.Create(item); err != nil {
			t.Fatal(err)
		}
	}
	now := func() time.Time { return time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC) }
	service := NewService(store, warranty.NewService(store, time.UTC, now), time.UTC, now)
	summary, err := service.BatchSummary(context.Background(), time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC), time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if summary.Count != 2 || len(summary.Reminders) != 2 {
		t.Fatalf("summary = %+v", summary)
	}
	if summary.Reminders[0].ItemName != "Alpha" || summary.Reminders[1].ItemName != "Beta" {
		t.Fatalf("reminders = %+v", summary.Reminders)
	}
}
