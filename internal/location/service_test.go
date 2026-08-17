package location

import (
	"github.com/zhangkui/go-home-inventory/internal/domain"
	"github.com/zhangkui/go-home-inventory/internal/inventory"
	"testing"
	"time"
)

func TestFilterByLocation(t *testing.T) {
	store := inventory.NewStore(time.Now)
	item, err := store.Create(domain.Item{Name: "TV", SerialNumber: "tv-1", PurchaseDate: time.Now(), WarrantyMonths: 12, Room: "Living Room", Position: "North Wall"})
	if err != nil {
		t.Fatal(err)
	}
	items := NewService(store).Filter("Living Room", "north wall")
	if len(items) != 1 || items[0].ID != item.ID {
		t.Fatalf("items = %+v", items)
	}
}
