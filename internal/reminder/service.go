package reminder

import (
	"context"
	"errors"
	"github.com/zhangkui/go-home-inventory/internal/domain"
	"github.com/zhangkui/go-home-inventory/internal/inventory"
	"github.com/zhangkui/go-home-inventory/internal/warranty"
	"sort"
	"time"
)

var ErrInvalidRange = errors.New("invalid date range")

type Service struct {
	items      *inventory.Store
	warranties *warranty.Service
	zone       *time.Location
	now        func() time.Time
}

func NewService(items *inventory.Store, warranties *warranty.Service, zone *time.Location, now func() time.Time) *Service {
	if zone == nil {
		zone = time.UTC
	}
	if now == nil {
		now = time.Now
	}
	return &Service{items: items, warranties: warranties, zone: zone, now: now}
}
func (s *Service) Upcoming(ctx context.Context, from, to time.Time) ([]domain.Reminder, error) {
	if to.Before(from) {
		return nil, ErrInvalidRange
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	result := make([]domain.Reminder, 0)
	for _, item := range s.items.List() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		expires := s.warranties.Expiration(item)
		if expires.Before(from) || expires.After(to) {
			continue
		}
		result = append(result, domain.Reminder{ItemID: item.ID, ItemName: item.Name, SerialNumber: item.SerialNumber, ExpiresAt: expires, DaysLeft: int(expires.Sub(s.now().In(s.zone)).Hours() / 24)})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ExpiresAt.Before(result[j].ExpiresAt) })
	return result, nil
}
func (s *Service) BatchSummary(ctx context.Context, from, to time.Time) (domain.ReminderSummary, error) {
	values, err := s.Upcoming(ctx, from, to)
	if err != nil {
		return domain.ReminderSummary{}, err
	}
	return domain.ReminderSummary{From: from, To: to, Count: len(values), Reminders: values}, nil
}
