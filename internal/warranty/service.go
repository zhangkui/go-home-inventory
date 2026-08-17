package warranty

import (
	"time"

	"github.com/zhangkui/go-home-inventory/internal/domain"
	"github.com/zhangkui/go-home-inventory/internal/inventory"
)

type Service struct {
	items *inventory.Store
	zone  *time.Location
	now   func() time.Time
}

func NewService(items *inventory.Store, zone *time.Location, now func() time.Time) *Service {
	if zone == nil {
		zone = time.UTC
	}
	if now == nil {
		now = time.Now
	}
	return &Service{items: items, zone: zone, now: now}
}
func (s *Service) Expiration(item domain.Item) time.Time {
	purchasedAt := item.PurchaseDate.In(s.zone)
	return addMonthsClamped(purchasedAt, item.WarrantyMonths+item.ExtendedWarrantyMonths)
}

func addMonthsClamped(value time.Time, months int) time.Time {
	monthIndex := int(value.Month()) - 1 + months
	year := value.Year() + monthIndex/12
	monthIndex %= 12
	if monthIndex < 0 {
		monthIndex += 12
		year--
	}

	month := time.Month(monthIndex + 1)
	day := value.Day()
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if day > lastDay {
		day = lastDay
	}

	return time.Date(year, month, day, value.Hour(), value.Minute(), value.Second(), value.Nanosecond(), value.Location())
}

func (s *Service) Status(itemID string) (domain.WarrantyStatus, error) {
	item, err := s.items.Get(itemID)
	if err != nil {
		return domain.WarrantyStatus{}, err
	}
	expires, now, status := s.Expiration(item), s.now().In(s.zone), "active"
	if !now.Before(expires) {
		status = "expired"
	}
	return domain.WarrantyStatus{ItemID: itemID, ExpiresAt: expires, Status: status, Remaining: int(expires.Sub(now).Hours() / 24)}, nil
}
