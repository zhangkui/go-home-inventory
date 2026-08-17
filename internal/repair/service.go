package repair

import (
	"errors"
	"fmt"
	"github.com/zhangkui/go-home-inventory/internal/domain"
	"github.com/zhangkui/go-home-inventory/internal/inventory"
	"sort"
	"strings"
	"sync"
	"time"
)

var ErrInvalidRepair = errors.New("invalid repair")

type Service struct {
	mu      sync.RWMutex
	items   *inventory.Store
	repairs map[string][]domain.Repair
	nextID  int64
	now     func() time.Time
}

func NewService(items *inventory.Store, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{items: items, repairs: make(map[string][]domain.Repair), now: now}
}
func (s *Service) Add(value domain.Repair) (domain.Repair, error) {
	if _, err := s.items.Get(value.ItemID); err != nil {
		return domain.Repair{}, err
	}
	if value.OccurredAt.IsZero() || value.CostCents <= 0 {
		return domain.Repair{}, ErrInvalidRepair
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	value.ID, value.Provider, value.Notes, value.CreatedAt = fmt.Sprintf("repair-%d", s.nextID), strings.TrimSpace(value.Provider), strings.TrimSpace(value.Notes), s.now()
	s.repairs[value.ItemID] = append(s.repairs[value.ItemID], value)
	return value, nil
}
func (s *Service) History(itemID string) ([]domain.Repair, error) {
	if _, err := s.items.Get(itemID); err != nil {
		return nil, err
	}
	s.mu.RLock()
	result := append([]domain.Repair(nil), s.repairs[itemID]...)
	s.mu.RUnlock()
	sort.Slice(result, func(i, j int) bool { return result[i].OccurredAt.Before(result[j].OccurredAt) })
	return result, nil
}
