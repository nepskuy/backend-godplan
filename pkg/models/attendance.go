package models

import (
	"time"

	"github.com/google/uuid"
)

// Attendance represents attendance data aligned with godplan.attendances schema
type Attendance struct {
	ID              uuid.UUID  `json:"id"`
	TenantID        uuid.UUID  `json:"tenant_id"`
	UserID          uuid.UUID  `json:"user_id"`
	EmployeeID      *uuid.UUID `json:"employee_id,omitempty"`
	ScheduleID      *uuid.UUID `json:"schedule_id,omitempty"`
	AttendanceDate  string     `json:"attendance_date"`
	Type            string     `json:"type"`   // CheckIn, CheckOut
	Status          string     `json:"status"` // approved, pending, pending_forced, rejected

	// Check In Data
	CheckInTime  *time.Time `json:"check_in_time,omitempty"`
	CheckInLat   float64    `json:"check_in_lat"`
	CheckInLng   float64    `json:"check_in_lng"`
	CheckInPhoto string     `json:"check_in_photo,omitempty"`

	// Check Out Data
	CheckOutTime  *time.Time `json:"check_out_time,omitempty"`
	CheckOutLat   float64    `json:"check_out_lat,omitempty"`
	CheckOutLng   float64    `json:"check_out_lng,omitempty"`
	CheckOutPhoto string     `json:"check_out_photo,omitempty"`

	TotalHours      float64 `json:"total_hours,omitempty"`
	LateMinutes     float64 `json:"late_minutes,omitempty"`
	OvertimeMinutes float64 `json:"overtime_minutes,omitempty"`

	InRange         bool   `json:"in_range"`
	ForceAttendance bool   `json:"force_attendance"`
	Notes           string `json:"notes,omitempty"`

	ApprovedBy      *uuid.UUID `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	RejectionReason string     `json:"rejection_reason,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// Legacy fields for backward compatibility with API responses
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	PhotoSelfie string  `json:"photo_selfie,omitempty"`
}

type ClockRequest struct {
	Latitude    float64 `json:"latitude" binding:"required"`
	Longitude   float64 `json:"longitude" binding:"required"`
	PhotoSelfie string  `json:"photo_selfie"`
	Force       bool    `json:"force"`
}
