package inventory

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/zhangkui/go-home-inventory/internal/domain"
)

var (
	ErrNotFound        = errors.New("item not found")
	ErrSerialDuplicate = errors.New("serial number already exists")
	ErrInvalidItem     = errors.New("invalid item")
)

type Store struct {
	mu         sync.RWMutex
	items      map[string]domain.Item
	serialToID map[string]string
	nextID     int64
	now        func() time.Time
}

func NewStore(now func() time.Time) *Store {
	if now == nil {
		now = time.Now
	}
	return &Store{items: make(map[string]domain.Item), serialToID: make(map[string]string), now: now}
}

func normalizeSerial(value string) string { return strings.ToUpper(strings.TrimSpace(value)) }

func validate(item domain.Item) error {
	if strings.TrimSpace(item.Name) == "" || normalizeSerial(item.SerialNumber) == "" || item.PurchaseDate.IsZero() {
		return ErrInvalidItem
	}
	if item.WarrantyMonths < 0 || item.ExtendedWarrantyMonths < 0 {
		return ErrInvalidItem
	}
	return nil
}

func (s *Store) Create(item domain.Item) (domain.Item, error) {
	if err := validate(item); err != nil {
		return domain.Item{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	serial := normalizeSerial(item.SerialNumber)
	if _, ok := s.serialToID[serial]; ok {
		return domain.Item{}, ErrSerialDuplicate
	}
	s.nextID++
	now := s.now()
	item.ID = fmt.Sprintf("item-%d", s.nextID)
	item.Name, item.Category, item.SerialNumber = strings.TrimSpace(item.Name), strings.TrimSpace(item.Category), serial
	item.Room, item.Position = strings.TrimSpace(item.Room), strings.TrimSpace(item.Position)
	item.CreatedAt, item.UpdatedAt = now, now
	s.items[item.ID], s.serialToID[serial] = item, item.ID
	return item, nil
}

func (s *Store) Update(id string, replacement domain.Item) (domain.Item, error) {
	if err := validate(replacement); err != nil {
		return domain.Item{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.items[id]
	if !ok {
		return domain.Item{}, ErrNotFound
	}
	serial := normalizeSerial(replacement.SerialNumber)
	if owner, exists := s.serialToID[serial]; exists && owner != id {
		return domain.Item{}, ErrSerialDuplicate
	}
	replacement.ID, replacement.CreatedAt, replacement.UpdatedAt = current.ID, current.CreatedAt, s.now()
	replacement.Name, replacement.Category, replacement.SerialNumber = strings.TrimSpace(replacement.Name), strings.TrimSpace(replacement.Category), serial
	replacement.Room, replacement.Position = strings.TrimSpace(replacement.Room), strings.TrimSpace(replacement.Position)
	s.items[id], s.serialToID[serial] = replacement, id
	return replacement, nil
}

func (s *Store) Get(id string) (domain.Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[id]
	if !ok {
		return domain.Item{}, ErrNotFound
	}
	return item, nil
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return ErrNotFound
	}
	delete(s.items, id)
	delete(s.serialToID, normalizeSerial(item.SerialNumber))
	return nil
}

func (s *Store) List() []domain.Item {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.Item, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items
}

func (s *Store) SetLocation(id, room, position string) (domain.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return domain.Item{}, ErrNotFound
	}
	item.Room, item.Position, item.UpdatedAt = strings.TrimSpace(room), strings.TrimSpace(position), s.now()
	s.items[id] = item
	return item, nil
}
