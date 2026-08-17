package warranty

import (
	"github.com/zhangkui/go-home-inventory/internal/domain"
	"github.com/zhangkui/go-home-inventory/internal/inventory"
	"time"
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
	return item.PurchaseDate.In(s.zone).AddDate(0, item.WarrantyMonths+item.ExtendedWarrantyMonths, 0)
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
