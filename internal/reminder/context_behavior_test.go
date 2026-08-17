package reminder

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zhangkui/go-home-inventory/internal/domain"
	"github.com/zhangkui/go-home-inventory/internal/inventory"
	"github.com/zhangkui/go-home-inventory/internal/warranty"
)

func TestUpcomingReturnsContextErrorWithEmptyInventory(t *testing.T) {
	store := inventory.NewStore(time.Now)
	service := NewService(store, warranty.NewService(store, time.UTC, time.Now), time.UTC, time.Now)
	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(1, 0, 0)

	tests := []struct {
		name string
		ctx  context.Context
		want error
	}{
		{name: "canceled", ctx: canceledContext(), want: context.Canceled},
		{name: "deadline exceeded", ctx: expiredContext(), want: context.DeadlineExceeded},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.Upcoming(test.ctx, from, to)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestUpcomingStopsWhenContextCanceledDuringTraversal(t *testing.T) {
	store := inventory.NewStore(time.Now)
	for _, serial := range []string{"p-1", "p-2"} {
		if _, err := store.Create(domain.Item{Name: "Phone", SerialNumber: serial, PurchaseDate: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), WarrantyMonths: 12}); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	now := func() time.Time {
		cancel()
		return time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	}
	service := NewService(store, warranty.NewService(store, time.UTC, now), time.UTC, now)

	_, err := service.Upcoming(ctx, time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC), time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestUpcomingValidatesRangeBeforeContext(t *testing.T) {
	store := inventory.NewStore(time.Now)
	service := NewService(store, warranty.NewService(store, time.UTC, time.Now), time.UTC, time.Now)
	from := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	_, err := service.Upcoming(canceledContext(), from, from.Add(-time.Hour))
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("error = %v, want ErrInvalidRange", err)
	}
}

func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func expiredContext() context.Context {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	cancel()
	return ctx
}
