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
	return addMonthsClamped(item.PurchaseDate.In(s.zone), item.WarrantyMonths+item.ExtendedWarrantyMonths)
}

// addMonthsClamped adds whole months to base, clamping the day to the last
// calendar day of the target month instead of letting an overflow day (e.g.
// Jan 31 + 1 month -> Feb 31) roll into the next month (Mar 2/3). The time of
// day and the configured zone are preserved. Negative deltas are supported.
func addMonthsClamped(base time.Time, months int) time.Time {
	year, month, day := base.Date()
	hour, min, sec := base.Clock()
	ns := base.Nanosecond()

	// Fold the month delta into a 0-based month index, then split back into a
	// year/month pair. Go's % truncates toward zero, so normalize negatives.
	totalMonths := int(month-1) + months
	targetYear := year + totalMonths/12
	targetMonth := totalMonths%12 + 1
	if targetMonth <= 0 {
		targetMonth += 12
		targetYear--
	}

	// Last day of the target month = day before the 1st of the next month.
	// time.Date normalizes month 13 into January of the following year.
	lastDay := time.Date(targetYear, time.Month(targetMonth+1), 1, 0, 0, 0, 0, base.Location()).AddDate(0, 0, -1).Day()
	if day > lastDay {
		day = lastDay
	}

	return time.Date(targetYear, time.Month(targetMonth), day, hour, min, sec, ns, base.Location())
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
