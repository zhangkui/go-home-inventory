package warranty

import (
	"github.com/zhangkui/go-home-inventory/internal/domain"
	"github.com/zhangkui/go-home-inventory/internal/inventory"
	"testing"
	"time"
)

func TestVerificationMonthEndExpirationClampsToCalendarMonth(t *testing.T) {
	store := inventory.NewStore(time.Now)
	item, err := store.Create(domain.Item{Name: "Tablet", SerialNumber: "TAB-1", PurchaseDate: time.Date(2024, 1, 31, 8, 30, 0, 0, time.UTC), WarrantyMonths: 1})
	if err != nil {
		t.Fatal(err)
	}
	status, err := NewService(store, time.UTC, time.Now).Status(item.ID)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2024, 2, 29, 8, 30, 0, 0, time.UTC)
	if !status.ExpiresAt.Equal(want) {
		t.Fatalf("expiration = %s, want %s", status.ExpiresAt, want)
	}
}
