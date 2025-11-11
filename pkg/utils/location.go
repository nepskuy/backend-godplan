package utils

import (
	"fmt"
	"math"

	"github.com/nepskuy/be-godplan/pkg/config"
)

// CalculateDistance menghitung jarak dalam meter antara dua koordinat menggunakan rumus Haversine
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Radius bumi dalam meter

	// Convert degrees to radians
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lon2 - lon1) * math.Pi / 180

	// Haversine formula
	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*
			math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := R * c // Jarak dalam meter

	return distance
}

// IsWithinOfficeRange mengecek apakah koordinat berada dalam jangkauan kantor
func IsWithinOfficeRange(userLat, userLon float64) (bool, float64) {
	cfg := config.Load()

	// Jika location check dimatikan, selalu return true
	if !cfg.EnableLocationCheck {
		return true, 0
	}

	distance := CalculateDistance(userLat, userLon, cfg.OfficeLatitude, cfg.OfficeLongitude)
	inRange := distance <= cfg.AttendanceRadiusMeters

	return inRange, distance
}

// LocationValidationResponse adalah response untuk validasi lokasi
type LocationValidationResponse struct {
	InRange   bool    `json:"in_range"`
	Message   string  `json:"message"`
	NeedForce bool    `json:"need_force"`
	Distance  float64 `json:"distance"`
	MaxRadius float64 `json:"max_radius"`
}

// ValidateLocation memvalidasi lokasi user terhadap kantor
func ValidateLocation(userLat, userLon float64) LocationValidationResponse {
	cfg := config.Load()

	// Jika location check dimatikan
	if !cfg.EnableLocationCheck {
		return LocationValidationResponse{
			InRange:   true,
			Message:   "Location check is disabled",
			NeedForce: false,
			Distance:  0,
			MaxRadius: cfg.AttendanceRadiusMeters,
		}
	}

	distance := CalculateDistance(userLat, userLon, cfg.OfficeLatitude, cfg.OfficeLongitude)
	inRange := distance <= cfg.AttendanceRadiusMeters

	response := LocationValidationResponse{
		InRange:   inRange,
		Distance:  distance,
		MaxRadius: cfg.AttendanceRadiusMeters,
	}

	if inRange {
		response.Message = "Lokasi valid, dalam jangkauan kantor"
		response.NeedForce = false
	} else {
		response.Message = fmt.Sprintf("Lokasi di luar jangkauan kantor (%.0f meter dari radius %0.f meter)", distance, cfg.AttendanceRadiusMeters)
		response.NeedForce = true
	}

	return response
}

// FormatDistance memformat jarak menjadi string yang mudah dibaca
func FormatDistance(meters float64) string {
	if meters < 1 {
		return "kurang dari 1 meter"
	} else if meters < 1000 {
		return fmt.Sprintf("%.0f meter", meters)
	} else {
		km := meters / 1000
		return fmt.Sprintf("%.1f km", km)
	}
}

// GetOfficeLocation mengembalikan informasi lokasi kantor
func GetOfficeLocation() (float64, float64, float64) {
	cfg := config.Load()
	return cfg.OfficeLatitude, cfg.OfficeLongitude, cfg.AttendanceRadiusMeters
}
