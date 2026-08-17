package inventory

import (
	"errors"
	"github.com/zhangkui/go-home-inventory/internal/domain"
	"testing"
	"time"
)

func sampleItem(serial string) domain.Item {
	return domain.Item{Name: "Camera", SerialNumber: serial, PurchaseDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), WarrantyMonths: 24}
}

func TestStoreCreateGetDelete(t *testing.T) {
	store := NewStore(time.Now)
	created, err := store.Create(sampleItem(" cam-1 "))
	if err != nil {
		t.Fatal(err)
	}
	if created.SerialNumber != "CAM-1" {
		t.Fatalf("serial = %q", created.SerialNumber)
	}
	if _, err := store.Get(created.ID); err != nil {
		t.Fatal(err)
	}
	if err := store.Delete(created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v", err)
	}
}

func TestStoreRejectsDuplicateSerial(t *testing.T) {
	store := NewStore(time.Now)
	if _, err := store.Create(sampleItem("abc")); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Create(sampleItem(" ABC ")); !errors.Is(err, ErrSerialDuplicate) {
		t.Fatalf("error = %v", err)
	}
}
