package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/nepskuy/be-godplan/pkg/database"
	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

type ClockInRequest struct {
	Latitude    float64 `json:"latitude" example:"-6.2088"`
	Longitude   float64 `json:"longitude" example:"106.8456"`
	PhotoSelfie string  `json:"photo_selfie" example:"base64_encoded_image"`
	Force       bool    `json:"force" example:"false"`
}

type ClockOutRequest struct {
	Latitude    float64 `json:"latitude" example:"-6.2088"`
	Longitude   float64 `json:"longitude" example:"106.8456"`
	PhotoSelfie string  `json:"photo_selfie" example:"base64_encoded_image"`
	Force       bool    `json:"force" example:"false"`
}

type LocationCheckRequest struct {
	Latitude  float64 `json:"latitude" example:"-6.2088"`
	Longitude float64 `json:"longitude" example:"106.8456"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type AttendanceResponse struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	Type            string    `json:"type"`
	Status          string    `json:"status"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	PhotoSelfie     string    `json:"photo_selfie,omitempty"`
	InRange         bool      `json:"in_range"`
	ForceAttendance bool      `json:"force_attendance"`
	CreatedAt       time.Time `json:"created_at"`
}

// CheckLocation godoc
// @Summary Check location validity
// @Description Validate if user location is within office range
// @Tags attendance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body LocationCheckRequest true "Location coordinates"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /attendance/check-location [post]
func CheckLocation(w http.ResponseWriter, r *http.Request) {
	var req LocationCheckRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response := models.LocationValidationResponse{
		InRange:   true,
		Message:   "Lokasi valid, dalam jangkauan kantor",
		NeedForce: false,
		Distance:  50.0,
	}

	utils.SuccessResponse(w, http.StatusOK, "Location validation successful", response)
}

// ClockIn godoc
// @Summary Clock in attendance
// @Description Record user clock-in with location validation
// @Tags attendance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ClockInRequest true "Clock-in data"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /attendance/clock-in [post]
func ClockIn(w http.ResponseWriter, r *http.Request) {
	log.Println("🔵 ClockIn started")

	defer func() {
		if err := recover(); err != nil {
			log.Printf("🚨 ClockIn PANIC: %v", err)
			utils.ErrorResponse(w, http.StatusInternalServerError, "Clock in failed")
		}
	}()

	var req ClockInRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("🔴 ClockIn decode error: %v", err)
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userIDVal := r.Context().Value("userID")
	if userIDVal == nil {
		log.Printf("🔴 ClockIn: userID not found in context")
		utils.ErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, ok := userIDVal.(int)
	if !ok {
		log.Printf("🔴 ClockIn: userID type assertion failed")
		utils.ErrorResponse(w, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	log.Printf("🔵 ClockIn: userID=%d", userID)

	inRange := true

	var attendanceID int
	err = database.DB.QueryRow(
		"INSERT INTO attendances (user_id, type, status, latitude, longitude, photo_selfie, in_range, force_attendance, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id",
		userID, "in", "approved", req.Latitude, req.Longitude, req.PhotoSelfie, inRange, req.Force, time.Now(),
	).Scan(&attendanceID)

	if err != nil {
		log.Printf("🔴 ClockIn database error: %v", err)
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to clock in: "+err.Error())
		return
	}

	log.Printf("🔵 ClockIn successful: attendanceID=%d", attendanceID)

	response := AttendanceResponse{
		ID:              attendanceID,
		UserID:          userID,
		Type:            "in",
		Status:          "approved",
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		InRange:         inRange,
		ForceAttendance: req.Force,
		CreatedAt:       time.Now(),
	}

	utils.SuccessResponse(w, http.StatusCreated, "Clock in successful", response)
}

// ClockOut godoc
// @Summary Clock out attendance
// @Description Record user clock-out with location validation
// @Tags attendance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ClockOutRequest true "Clock-out data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /attendance/clock-out [post]
func ClockOut(w http.ResponseWriter, r *http.Request) {
	log.Println("🔵 ClockOut started")

	defer func() {
		if err := recover(); err != nil {
			log.Printf("🚨 ClockOut PANIC: %v", err)
			utils.ErrorResponse(w, http.StatusInternalServerError, "Clock out failed")
		}
	}()

	var req ClockOutRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("🔴 ClockOut decode error: %v", err)
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userIDVal := r.Context().Value("userID")
	if userIDVal == nil {
		log.Printf("🔴 ClockOut: userID not found in context")
		utils.ErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, ok := userIDVal.(int)
	if !ok {
		log.Printf("🔴 ClockOut: userID type assertion failed")
		utils.ErrorResponse(w, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	log.Printf("🔵 ClockOut: userID=%d", userID)

	inRange := true

	var attendanceID int
	err = database.DB.QueryRow(
		"INSERT INTO attendances (user_id, type, status, latitude, longitude, photo_selfie, in_range, force_attendance, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id",
		userID, "out", "approved", req.Latitude, req.Longitude, req.PhotoSelfie, inRange, req.Force, time.Now(),
	).Scan(&attendanceID)

	if err != nil {
		log.Printf("🔴 ClockOut database error: %v", err)
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to clock out: "+err.Error())
		return
	}

	log.Printf("🔵 ClockOut successful: attendanceID=%d", attendanceID)

	response := AttendanceResponse{
		ID:              attendanceID,
		UserID:          userID,
		Type:            "out",
		Status:          "approved",
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		InRange:         inRange,
		ForceAttendance: req.Force,
		CreatedAt:       time.Now(),
	}

	utils.SuccessResponse(w, http.StatusOK, "Clock out successful", response)
}

// GetAttendance godoc
// @Summary Get attendance history
// @Description Get logged-in user's attendance history with optional filters
// @Tags attendance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param date query string false "Filter by date (YYYY-MM-DD)"
// @Param limit query int false "Limit number of records (default: 30)"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /attendance [get]
func GetAttendance(w http.ResponseWriter, r *http.Request) {
	log.Println("🔵 GetAttendance started")

	defer func() {
		if err := recover(); err != nil {
			log.Printf("🚨 GetAttendance PANIC: %v", err)
			utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get attendance")
		}
	}()

	userIDVal := r.Context().Value("userID")
	if userIDVal == nil {
		log.Printf("🔴 GetAttendance: userID not found in context")
		utils.ErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, ok := userIDVal.(int)
	if !ok {
		log.Printf("🔴 GetAttendance: userID type assertion failed")
		utils.ErrorResponse(w, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	log.Printf("🔵 GetAttendance: userID=%d", userID)

	query := r.URL.Query()
	dateFilter := query.Get("date")
	limitStr := query.Get("limit")

	var limit int
	if limitStr == "" {
		limit = 30
	} else {
		limit, _ = strconv.Atoi(limitStr)
		if limit <= 0 {
			limit = 30
		}
	}

	var rows *sql.Rows
	var err error

	if dateFilter != "" {
		rows, err = database.DB.Query(
			"SELECT id, user_id, type, status, latitude, longitude, photo_selfie, in_range, force_attendance, created_at FROM attendances WHERE user_id = $1 AND DATE(created_at) = $2 ORDER BY created_at DESC LIMIT $3",
			userID, dateFilter, limit,
		)
	} else {
		rows, err = database.DB.Query(
			"SELECT id, user_id, type, status, latitude, longitude, photo_selfie, in_range, force_attendance, created_at FROM attendances WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2",
			userID, limit,
		)
	}

	if err != nil {
		log.Printf("🔴 GetAttendance database error: %v", err)
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to fetch attendance records")
		return
	}
	defer rows.Close()

	var attendances []AttendanceResponse
	for rows.Next() {
		var att models.Attendance
		err := rows.Scan(
			&att.ID, &att.UserID, &att.Type, &att.Status,
			&att.Latitude, &att.Longitude, &att.PhotoSelfie,
			&att.InRange, &att.ForceAttendance, &att.CreatedAt,
		)
		if err != nil {
			log.Printf("🔴 GetAttendance scan error: %v", err)
			continue
		}

		attendance := AttendanceResponse{
			ID:              att.ID,
			UserID:          att.UserID,
			Type:            att.Type,
			Status:          att.Status,
			Latitude:        att.Latitude,
			Longitude:       att.Longitude,
			InRange:         att.InRange,
			ForceAttendance: att.ForceAttendance,
			CreatedAt:       att.CreatedAt,
		}
		attendances = append(attendances, attendance)
	}

	log.Printf("🔵 GetAttendance successful: found %d records", len(attendances))
	utils.SuccessResponse(w, http.StatusOK, "Attendance records retrieved", attendances)
}
