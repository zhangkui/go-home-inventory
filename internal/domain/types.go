package domain

import "time"

type Item struct {
	ID                     string    `json:"id"`
	Name                   string    `json:"name"`
	Category               string    `json:"category"`
	SerialNumber           string    `json:"serial_number"`
	Room                   string    `json:"room"`
	Position               string    `json:"position"`
	PurchaseDate           time.Time `json:"purchase_date"`
	WarrantyMonths         int       `json:"warranty_months"`
	ExtendedWarrantyMonths int       `json:"extended_warranty_months"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type Repair struct {
	ID         string    `json:"id"`
	ItemID     string    `json:"item_id"`
	OccurredAt time.Time `json:"occurred_at"`
	CostCents  int64     `json:"cost_cents"`
	Provider   string    `json:"provider"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
}

type WarrantyStatus struct {
	ItemID    string    `json:"item_id"`
	ExpiresAt time.Time `json:"expires_at"`
	Status    string    `json:"status"`
	Remaining int       `json:"remaining_days"`
}

type Reminder struct {
	ItemID       string    `json:"item_id"`
	ItemName     string    `json:"item_name"`
	SerialNumber string    `json:"serial_number"`
	ExpiresAt    time.Time `json:"expires_at"`
	DaysLeft     int       `json:"days_left"`
}

type ReminderSummary struct {
	From      time.Time  `json:"from"`
	To        time.Time  `json:"to"`
	Count     int        `json:"count"`
	Reminders []Reminder `json:"reminders"`
}
