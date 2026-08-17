package warranty

import (
	"testing"
	"time"

	"github.com/zhangkui/go-home-inventory/internal/domain"
	"github.com/zhangkui/go-home-inventory/internal/inventory"
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

func TestExpirationAddsMonthsWithoutCrossMonthDrift(t *testing.T) {
	zone := time.FixedZone("UTC+08", 8*60*60)
	service := NewService(inventory.NewStore(time.Now), zone, time.Now)

	tests := []struct {
		name           string
		purchaseDate   time.Time
		warrantyMonths int
		extendedMonths int
		want           time.Time
	}{
		{
			name:           "leap year month end",
			purchaseDate:   time.Date(2024, time.January, 31, 23, 59, 58, 765432100, zone),
			warrantyMonths: 1,
			want:           time.Date(2024, time.February, 29, 23, 59, 58, 765432100, zone),
		},
		{
			name:           "non-leap year month end",
			purchaseDate:   time.Date(2023, time.January, 31, 8, 30, 15, 0, zone),
			warrantyMonths: 1,
			want:           time.Date(2023, time.February, 28, 8, 30, 15, 0, zone),
		},
		{
			name:           "extended warranty",
			purchaseDate:   time.Date(2024, time.August, 31, 12, 45, 30, 0, zone),
			warrantyMonths: 3,
			extendedMonths: 3,
			want:           time.Date(2025, time.February, 28, 12, 45, 30, 0, zone),
		},
		{
			name:           "cross year",
			purchaseDate:   time.Date(2024, time.November, 30, 6, 5, 4, 0, zone),
			warrantyMonths: 3,
			want:           time.Date(2025, time.February, 28, 6, 5, 4, 0, zone),
		},
		{
			name:           "ordinary date",
			purchaseDate:   time.Date(2024, time.May, 15, 9, 10, 11, 0, zone),
			warrantyMonths: 7,
			want:           time.Date(2024, time.December, 15, 9, 10, 11, 0, zone),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			item := domain.Item{
				PurchaseDate:           test.purchaseDate,
				WarrantyMonths:         test.warrantyMonths,
				ExtendedWarrantyMonths: test.extendedMonths,
			}

			got := service.Expiration(item)
			if !got.Equal(test.want) {
				t.Fatalf("expiration = %s, want %s", got, test.want)
			}
			if got.Location() != zone {
				t.Fatalf("location = %s, want %s", got.Location(), zone)
			}
		})
	}
}
