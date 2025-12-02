package models

import (
	"time"

	"github.com/google/uuid"
)

type Attendance struct {
	ID              uuid.UUID `json:"id"`
	TenantID        uuid.UUID `json:"tenant_id"`
	UserID          uuid.UUID `json:"user_id"`
	Type            string    `json:"type"`
	Status          string    `json:"status"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	PhotoSelfie     string    `json:"photo_selfie"`
	InRange         bool      `json:"in_range"`
	ForceAttendance bool      `json:"force_attendance"`
	CreatedAt       time.Time `json:"created_at"`
}

type ClockRequest struct {
	Latitude    float64 `json:"latitude" binding:"required"`
	Longitude   float64 `json:"longitude" binding:"required"`
	PhotoSelfie string  `json:"photo_selfie" binding:"required"`
	Force       bool    `json:"force"`
}
