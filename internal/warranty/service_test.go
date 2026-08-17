package warranty

import (
	"github.com/zhangkui/go-home-inventory/internal/domain"
	"github.com/zhangkui/go-home-inventory/internal/inventory"
	"testing"
	"time"
)

func TestStatusUsesBaseAndExtendedMonths(t *testing.T) {
	store := inventory.NewStore(time.Now)
	item, err := store.Create(domain.Item{Name: "Washer", SerialNumber: "w-1", PurchaseDate: time.Date(2024, 5, 10, 0, 0, 0, 0, time.UTC), WarrantyMonths: 12, ExtendedWarrantyMonths: 6})
	if err != nil {
		t.Fatal(err)
	}
	status, err := NewService(store, time.UTC, func() time.Time { return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) }).Status(item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !status.ExpiresAt.Equal(time.Date(2025, 11, 10, 0, 0, 0, 0, time.UTC)) || status.Status != "active" {
		t.Fatalf("status = %+v", status)
	}
}
