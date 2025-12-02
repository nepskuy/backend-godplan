package models

import "github.com/google/uuid"

type OfficeLocation struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Name      string    `json:"name"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Radius    int       `json:"radius"`
	Address   string    `json:"address"`
	IsActive  bool      `json:"is_active"`
}
