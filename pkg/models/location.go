package models

type OfficeLocation struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Radius    int     `json:"radius"`
	Address   string  `json:"address"`
	IsActive  bool    `json:"is_active"`
}
