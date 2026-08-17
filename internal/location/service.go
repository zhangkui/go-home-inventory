package location

import (
	"github.com/zhangkui/go-home-inventory/internal/domain"
	"github.com/zhangkui/go-home-inventory/internal/inventory"
	"strings"
)

type Service struct{ items *inventory.Store }

func NewService(items *inventory.Store) *Service { return &Service{items: items} }
func (s *Service) Set(itemID, room, position string) (domain.Item, error) {
	return s.items.SetLocation(itemID, room, position)
}
func (s *Service) Filter(room, position string) []domain.Item {
	room, position = strings.TrimSpace(room), strings.TrimSpace(position)
	result := make([]domain.Item, 0)
	for _, item := range s.items.List() {
		if room != "" && item.Room != room {
			continue
		}
		if position != "" && !strings.EqualFold(item.Position, position) {
			continue
		}
		result = append(result, item)
	}
	return result
}
