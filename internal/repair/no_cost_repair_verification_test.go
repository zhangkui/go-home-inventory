package repair

import (
	"github.com/zhangkui/go-home-inventory/internal/domain"
	"github.com/zhangkui/go-home-inventory/internal/inventory"
	"testing"
	"time"
)

func TestVerificationAllowsNoCostWarrantyRepair(t *testing.T) {
	store := inventory.NewStore(time.Now)
	item, err := store.Create(domain.Item{Name: "Fridge", SerialNumber: "FR-1", PurchaseDate: time.Now(), WarrantyMonths: 36})
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(store, time.Now)
	if _, err := service.Add(domain.Repair{ItemID: item.ID, OccurredAt: time.Now(), CostCents: 0, Provider: "Manufacturer", Notes: "covered by warranty"}); err != nil {
		t.Fatalf("no-cost repair should be recorded: %v", err)
	}
	history, err := service.History(item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].CostCents != 0 {
		t.Fatalf("history = %+v", history)
	}
}
