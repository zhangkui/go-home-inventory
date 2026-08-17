package repair

import (
	"errors"
	"github.com/zhangkui/go-home-inventory/internal/domain"
	"github.com/zhangkui/go-home-inventory/internal/inventory"
	"testing"
	"time"
)

func TestRepairHistoryIsChronological(t *testing.T) {
	store := inventory.NewStore(time.Now)
	item, err := store.Create(domain.Item{Name: "Laptop", SerialNumber: "l-1", PurchaseDate: time.Now(), WarrantyMonths: 12})
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(store, time.Now)
	for _, day := range []int{20, 10} {
		if _, err := service.Add(domain.Repair{ItemID: item.ID, OccurredAt: time.Date(2025, 1, day, 0, 0, 0, 0, time.UTC), CostCents: 5000}); err != nil {
			t.Fatal(err)
		}
	}
	history, err := service.History(item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 || history[0].OccurredAt.Day() != 10 {
		t.Fatalf("history = %+v", history)
	}
}
func TestAddRejectsNegativeCost(t *testing.T) {
	store := inventory.NewStore(time.Now)
	item, err := store.Create(domain.Item{Name: "Laptop", SerialNumber: "l-1", PurchaseDate: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(store, time.Now)

	_, err = service.Add(domain.Repair{ItemID: item.ID, OccurredAt: time.Now(), CostCents: -1})
	if !errors.Is(err, ErrInvalidRepair) {
		t.Fatalf("err = %v, want ErrInvalidRepair", err)
	}
}

func TestAddRejectsZeroTime(t *testing.T) {
	store := inventory.NewStore(time.Now)
	item, err := store.Create(domain.Item{Name: "Laptop", SerialNumber: "l-1", PurchaseDate: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(store, time.Now)

	_, err = service.Add(domain.Repair{ItemID: item.ID, CostCents: 0})
	if !errors.Is(err, ErrInvalidRepair) {
		t.Fatalf("err = %v, want ErrInvalidRepair", err)
	}
}

func TestAddReturnsNotFoundForMissingItem(t *testing.T) {
	service := NewService(inventory.NewStore(time.Now), time.Now)

	_, err := service.Add(domain.Repair{ItemID: "missing", OccurredAt: time.Now(), CostCents: -1})
	if !errors.Is(err, inventory.ErrNotFound) {
		t.Fatalf("err = %v, want inventory.ErrNotFound", err)
	}
}
