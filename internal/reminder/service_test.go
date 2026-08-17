package reminder

import (
	"context"
	"github.com/zhangkui/go-home-inventory/internal/domain"
	"github.com/zhangkui/go-home-inventory/internal/inventory"
	"github.com/zhangkui/go-home-inventory/internal/warranty"
	"testing"
	"time"
)

func TestUpcomingReturnsOnlyRangeMatches(t *testing.T) {
	store := inventory.NewStore(time.Now)
	for _, item := range []domain.Item{{Name: "Phone", SerialNumber: "p-1", PurchaseDate: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), WarrantyMonths: 12}, {Name: "Watch", SerialNumber: "w-1", PurchaseDate: time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC), WarrantyMonths: 12}} {
		if _, err := store.Create(item); err != nil {
			t.Fatal(err)
		}
	}
	now := func() time.Time { return time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC) }
	service := NewService(store, warranty.NewService(store, time.UTC, now), time.UTC, now)
	got, err := service.Upcoming(context.Background(), time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC), time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ItemName != "Phone" {
		t.Fatalf("reminders = %+v", got)
	}
}
