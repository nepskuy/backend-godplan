package models

import (
	"time"
)

type Attendance struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
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
