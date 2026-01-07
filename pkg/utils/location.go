package utils

import (
	"fmt"
	"log"
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

// GPSAccuracy represents the quality of GPS signal
type GPSAccuracy int

const (
	AccuracyExcellent GPSAccuracy = iota // < 10m
	AccuracyGood                          // 10-25m
	AccuracyFair                          // 25-50m
	AccuracyPoor                          // > 50m
)

// LocationCheckRequest with accuracy information from device
type LocationCheckWithAccuracy struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Accuracy  float64 `json:"accuracy"` // GPS accuracy in meters dari device
	Altitude  float64 `json:"altitude,omitempty"`
	Heading   float64 `json:"heading,omitempty"`
	Speed     float64 `json:"speed,omitempty"`
}

// AdaptiveThreshold calculates dynamic threshold based on GPS accuracy
// Semakin buruk GPS accuracy, semakin besar threshold yang diberikan
func AdaptiveThreshold(baseRadius float64, gpsAccuracy float64) float64 {
	// Base threshold
	threshold := baseRadius

	// Jika GPS accuracy buruk, tambahkan tolerance
	// Formula: threshold = baseRadius + (gpsAccuracy * multiplier)
	// multiplier menurun seiring GPS accuracy membaik
	
	switch {
	case gpsAccuracy == 0:
		// No accuracy data provided, use base radius only
		return threshold
		
	case gpsAccuracy < 10:
		// Excellent GPS - minimal adjustment
		threshold += gpsAccuracy * 0.5
		log.Printf("📍 [GPS] Excellent accuracy (%.1fm) - threshold: %.1fm", gpsAccuracy, threshold)
		
	case gpsAccuracy < 25:
		// Good GPS - moderate adjustment  
		threshold += gpsAccuracy * 1.0
		log.Printf("📍 [GPS] Good accuracy (%.1fm) - threshold: %.1fm", gpsAccuracy, threshold)
		
	case gpsAccuracy < 50:
		// Fair GPS - significant adjustment
		threshold += gpsAccuracy * 1.5
		log.Printf("⚠️ [GPS] Fair accuracy (%.1fm) - threshold: %.1fm", gpsAccuracy, threshold)
		
	default:
		// Poor GPS - maximum adjustment
		threshold += gpsAccuracy * 2.0
		log.Printf("⚠️ [GPS] Poor accuracy (%.1fm) - threshold: %.1fm", gpsAccuracy, threshold)
	}

	// Cap maximum threshold untuk keamanan
	maxThreshold := baseRadius * 3 // Maksimal 3x radius dasar
	if threshold > maxThreshold {
		log.Printf("⚠️ [GPS] Threshold capped from %.1fm to %.1fm", threshold, maxThreshold)
		threshold = maxThreshold
	}

	return threshold
}

// IsWithinOfficeRange mengecek apakah koordinat berada dalam jangkauan kantor
// DEPRECATED: Use IsWithinOfficeRangeAdaptive for better accuracy
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

// IsWithinOfficeRangeAdaptive - Enhanced version dengan adaptive threshold
func IsWithinOfficeRangeAdaptive(userLat, userLon, gpsAccuracy float64) (bool, float64, float64) {
	cfg := config.Load()

	// Jika location check dimatikan, selalu return true
	if !cfg.EnableLocationCheck {
		return true, 0, cfg.AttendanceRadiusMeters
	}

	distance := CalculateDistance(userLat, userLon, cfg.OfficeLatitude, cfg.OfficeLongitude)
	
	// Hitung adaptive threshold berdasarkan GPS accuracy
	adaptiveRadius := AdaptiveThreshold(cfg.AttendanceRadiusMeters, gpsAccuracy)
	
	inRange := distance <= adaptiveRadius

	log.Printf("📍 [Location] Distance: %.1fm | Base Radius: %.1fm | Adaptive Radius: %.1fm | In Range: %v",
		distance, cfg.AttendanceRadiusMeters, adaptiveRadius, inRange)

	return inRange, distance, adaptiveRadius
}

// LocationValidationResponse adalah response untuk validasi lokasi
type LocationValidationResponse struct {
	InRange         bool    `json:"in_range"`
	Message         string  `json:"message"`
	DetailedMessage string  `json:"detailed_message,omitempty"`
	NeedForce       bool    `json:"need_force"`
	Distance        float64 `json:"distance"`
	MaxRadius       float64 `json:"max_radius"`
	AdaptiveRadius  float64 `json:"adaptive_radius,omitempty"`
	GPSAccuracy     float64 `json:"gps_accuracy,omitempty"`
	GPSQuality      string  `json:"gps_quality,omitempty"`
	Recommendation  string  `json:"recommendation,omitempty"`
}

// GetGPSQuality returns a human-readable GPS quality assessment
func GetGPSQuality(accuracy float64) string {
	switch {
	case accuracy == 0:
		return "unknown"
	case accuracy < 10:
		return "excellent"
	case accuracy < 25:
		return "good"
	case accuracy < 50:
		return "fair"
	default:
		return "poor"
	}
}

// ValidateLocation memvalidasi lokasi user terhadap kantor
// DEPRECATED: Use ValidateLocationAdaptive for better accuracy
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
		// Format jarak jadi lebih user-friendly
		distanceStr := FormatDistance(distance)
		radiusStr := FormatDistance(cfg.AttendanceRadiusMeters)
		response.Message = fmt.Sprintf("Anda berada %s dari lokasi kantor (jangkauan: %s)", distanceStr, radiusStr)
		response.NeedForce = true
	}

	return response
}

// ValidateLocationAdaptive - Enhanced validation dengan GPS accuracy consideration
func ValidateLocationAdaptive(userLat, userLon, gpsAccuracy float64) LocationValidationResponse {
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

	// Gunakan adaptive range checking
	inRange, distance, adaptiveRadius := IsWithinOfficeRangeAdaptive(userLat, userLon, gpsAccuracy)
	gpsQuality := GetGPSQuality(gpsAccuracy)

	response := LocationValidationResponse{
		InRange:        inRange,
		Distance:       distance,
		MaxRadius:      cfg.AttendanceRadiusMeters,
		AdaptiveRadius: adaptiveRadius,
		GPSAccuracy:    gpsAccuracy,
		GPSQuality:     gpsQuality,
	}

	// Generate detailed messages
	distanceStr := FormatDistance(distance)
	baseRadiusStr := FormatDistance(cfg.AttendanceRadiusMeters)
	adaptiveRadiusStr := FormatDistance(adaptiveRadius)

	if inRange {
		response.Message = "Lokasi valid, dalam jangkauan kantor"
		response.DetailedMessage = fmt.Sprintf("Jarak: %s | Jangkauan: %s | Akurasi GPS: %s (%.0fm)",
			distanceStr, adaptiveRadiusStr, gpsQuality, gpsAccuracy)
		response.NeedForce = false

		// Add recommendations for poor GPS
		if gpsQuality == "poor" || gpsQuality == "fair" {
			response.Recommendation = "Untuk akurasi lebih baik, coba pindah ke area dengan sinyal GPS lebih kuat (dekat jendela atau outdoor)"
		}
	} else {
		response.Message = fmt.Sprintf("Anda berada %s dari lokasi kantor", distanceStr)
		
		// Detailed explanation
		marginStr := FormatDistance(distance - adaptiveRadius)
		response.DetailedMessage = fmt.Sprintf(
			"Jarak: %s | Jangkauan base: %s | Jangkauan adaptive: %s | Margin: %s | Akurasi GPS: %s (%.0fm)",
			distanceStr, baseRadiusStr, adaptiveRadiusStr, marginStr, gpsQuality, gpsAccuracy,
		)

		// Check if just barely out of range
		overagePercentage := ((distance - adaptiveRadius) / adaptiveRadius) * 100
		
		if overagePercentage < 10 {
			// Very close - likely GPS error
			response.NeedForce = false // Auto-approve
			response.Message += " - Kemungkinan error GPS signal, attendance di-approve otomatis"
			response.Recommendation = "Anda sangat dekat dengan jangkauan kantor. Coba refresh lokasi atau pindah sedikit."
		} else if overagePercentage < 30 {
			// Moderately out of range
			response.NeedForce = true
			response.Recommendation = "Jika Anda yakin sudah di kantor, coba: 1) Pindah dekat jendela 2) Pastikan GPS device aktif 3) Refresh lokasi"
		} else {
			// Significantly out of range
			response.NeedForce = true
			response.Recommendation = "Lokasi Anda cukup jauh dari kantor. Pastikan Anda berada di lokasi yang benar."
		}
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

// EstimateGPSAccuracyFromContext provides estimation when device doesn't report accuracy
// This is a fallback for older devices or browsers
func EstimateGPSAccuracyFromContext(isIndoor bool, deviceType string) float64 {
	// Indoor typically has worse GPS
	if isIndoor {
		return 50.0 // Assume fair accuracy
	}

	// Mobile devices typically have better GPS than desktops
	if deviceType == "mobile" {
		return 15.0 // Assume good accuracy
	}

	// Desktop/unknown - assume moderate accuracy
	return 30.0
}
