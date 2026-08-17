package httpapi

import (
	"encoding/json"
	"errors"
	"github.com/zhangkui/go-home-inventory/internal/domain"
	"github.com/zhangkui/go-home-inventory/internal/inventory"
	"github.com/zhangkui/go-home-inventory/internal/location"
	"github.com/zhangkui/go-home-inventory/internal/reminder"
	"github.com/zhangkui/go-home-inventory/internal/repair"
	"github.com/zhangkui/go-home-inventory/internal/warranty"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	items      *inventory.Store
	locations  *location.Service
	warranties *warranty.Service
	repairs    *repair.Service
	reminders  *reminder.Service
	zone       *time.Location
}

func NewServer(items *inventory.Store, locations *location.Service, warranties *warranty.Service, repairs *repair.Service, reminders *reminder.Service, zone *time.Location) http.Handler {
	server := &Server{items: items, locations: locations, warranties: warranties, repairs: repairs, reminders: reminders, zone: zone}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", server.health)
	mux.HandleFunc("/items", server.itemsCollection)
	mux.HandleFunc("/items/", server.itemResource)
	mux.HandleFunc("/reminders/upcoming", server.upcoming)
	mux.HandleFunc("/reminders/summary", server.summary)
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) itemsCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var item domain.Item
		if !decodeJSON(w, r, &item) {
			return
		}
		created, err := s.items.Create(item)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.locations.Filter(r.URL.Query().Get("room"), r.URL.Query().Get("position")))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) itemResource(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != "items" {
		http.NotFound(w, r)
		return
	}
	id := parts[1]
	if len(parts) == 3 {
		switch parts[2] {
		case "location":
			s.location(w, r, id)
		case "warranty":
			s.warranty(w, r, id)
		case "repairs":
			s.repairsResource(w, r, id)
		default:
			http.NotFound(w, r)
		}
		return
	}
	switch r.Method {
	case http.MethodGet:
		item, err := s.items.Get(id)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodPut:
		var item domain.Item
		if !decodeJSON(w, r, &item) {
			return
		}
		updated, err := s.items.Update(id, item)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, updated)
	case http.MethodDelete:
		if err := s.items.Delete(id); err != nil {
			writeError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) location(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		Room     string `json:"room"`
		Position string `json:"position"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := s.locations.Set(id, request.Room, request.Position)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) warranty(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	status, err := s.warranties.Status(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) repairsResource(w http.ResponseWriter, r *http.Request, id string) {
	switch r.Method {
	case http.MethodPost:
		var value domain.Repair
		if !decodeJSON(w, r, &value) {
			return
		}
		value.ItemID = id
		created, err := s.repairs.Add(value)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	case http.MethodGet:
		history, err := s.repairs.History(id)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, history)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) upcoming(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	from, to, ok := s.rangeParams(w, r)
	if !ok {
		return
	}
	values, err := s.reminders.Upcoming(r.Context(), from, to)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, values)
}

func (s *Server) summary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	from, err := time.ParseInLocation(time.DateOnly, request.From, s.zone)
	if err != nil {
		http.Error(w, "invalid from date", http.StatusBadRequest)
		return
	}
	to, err := time.ParseInLocation(time.DateOnly, request.To, s.zone)
	if err != nil {
		http.Error(w, "invalid to date", http.StatusBadRequest)
		return
	}
	value, err := s.reminders.BatchSummary(r.Context(), from, to)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}

func (s *Server) rangeParams(w http.ResponseWriter, r *http.Request) (time.Time, time.Time, bool) {
	from, err := time.ParseInLocation(time.DateOnly, r.URL.Query().Get("from"), s.zone)
	if err != nil {
		http.Error(w, "invalid from date", http.StatusBadRequest)
		return time.Time{}, time.Time{}, false
	}
	to, err := time.ParseInLocation(time.DateOnly, r.URL.Query().Get("to"), s.zone)
	if err != nil {
		http.Error(w, "invalid to date", http.StatusBadRequest)
		return time.Time{}, time.Time{}, false
	}
	return from, to, true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return false
	}
	return true
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, inventory.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, inventory.ErrSerialDuplicate):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, inventory.ErrInvalidItem), errors.Is(err, repair.ErrInvalidRepair), errors.Is(err, reminder.ErrInvalidRange):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
