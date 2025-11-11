package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/nepskuy/be-godplan/pkg/config"
	"github.com/nepskuy/be-godplan/pkg/database"
	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

// ClockInHTTP untuk Vercel/gorilla-mux compatibility
func ClockInHTTP(w http.ResponseWriter, r *http.Request) {
	if config.IsDevelopment() {
		log.Println("🔵 ClockInHTTP started")
	}

	defer func() {
		if err := recover(); err != nil {
			if config.IsDevelopment() {
				log.Printf("🚨 ClockInHTTP PANIC: %v", err)
			}
			utils.ErrorResponse(w, http.StatusInternalServerError, "Clock in failed")
		}
	}()

	var req ClockInRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		if config.IsDevelopment() {
			log.Printf("🔴 ClockInHTTP decode error: %v", err)
		}
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get userID from context (set by AuthMiddleware)
	userIDVal := r.Context().Value("userID")
	if userIDVal == nil {
		if config.IsDevelopment() {
			log.Printf("🔴 ClockInHTTP: userID not found in context")
		}
		utils.ErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, ok := userIDVal.(int)
	if !ok {
		if config.IsDevelopment() {
			log.Printf("🔴 ClockInHTTP: userID type assertion failed")
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	if config.IsDevelopment() {
		log.Printf("🔵 ClockInHTTP: userID=%d", userID)
	}

	// 🔥 NEW: Validasi lokasi
	inRange, distance := utils.IsWithinOfficeRange(req.Latitude, req.Longitude)
	cfg := config.Load()

	// Jika tidak dalam range dan tidak menggunakan force, tolak attendance
	if !inRange && !req.Force {
		if config.IsDevelopment() {
			log.Printf("🔴 ClockInHTTP: Location out of range (%.0fm > %.0fm)", distance, cfg.AttendanceRadiusMeters)
		}
		utils.ErrorResponse(w, http.StatusBadRequest,
			"Lokasi di luar jangkauan kantor. Gunakan force=true untuk tetap melanjutkan.")
		return
	}

	// Tentukan status berdasarkan lokasi dan force
	status := "approved"
	if !inRange && req.Force {
		status = "forced"
		if config.IsDevelopment() {
			log.Printf("🟡 ClockInHTTP: Using forced attendance (%.0fm > %.0fm)", distance, cfg.AttendanceRadiusMeters)
		}
	}

	var attendanceID int
	err = database.DB.QueryRow(
		"INSERT INTO attendances (user_id, type, status, latitude, longitude, photo_selfie, in_range, force_attendance, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id",
		userID, "in", status, req.Latitude, req.Longitude, req.PhotoSelfie, inRange, req.Force, time.Now(),
	).Scan(&attendanceID)

	if err != nil {
		if config.IsDevelopment() {
			log.Printf("🔴 ClockInHTTP database error: %v", err)
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to clock in")
		return
	}

	if config.IsDevelopment() {
		log.Printf("🔵 ClockInHTTP successful: attendanceID=%d, status=%s, inRange=%t", attendanceID, status, inRange)
	}

	response := AttendanceResponse{
		ID:              attendanceID,
		UserID:          userID,
		Type:            "in",
		Status:          status,
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		PhotoSelfie:     req.PhotoSelfie,
		InRange:         inRange,
		ForceAttendance: req.Force,
		CreatedAt:       time.Now(),
		Distance:        distance,
		MaxRadius:       cfg.AttendanceRadiusMeters,
	}

	utils.SuccessResponse(w, http.StatusCreated, "Clock in successful", response)
}

// ClockOutHTTP untuk Vercel/gorilla-mux compatibility
func ClockOutHTTP(w http.ResponseWriter, r *http.Request) {
	if config.IsDevelopment() {
		log.Println("🔵 ClockOutHTTP started")
	}

	defer func() {
		if err := recover(); err != nil {
			if config.IsDevelopment() {
				log.Printf("🚨 ClockOutHTTP PANIC: %v", err)
			}
			utils.ErrorResponse(w, http.StatusInternalServerError, "Clock out failed")
		}
	}()

	var req ClockOutRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		if config.IsDevelopment() {
			log.Printf("🔴 ClockOutHTTP decode error: %v", err)
		}
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get userID from context (set by AuthMiddleware)
	userIDVal := r.Context().Value("userID")
	if userIDVal == nil {
		if config.IsDevelopment() {
			log.Printf("🔴 ClockOutHTTP: userID not found in context")
		}
		utils.ErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, ok := userIDVal.(int)
	if !ok {
		if config.IsDevelopment() {
			log.Printf("🔴 ClockOutHTTP: userID type assertion failed")
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	if config.IsDevelopment() {
		log.Printf("🔵 ClockOutHTTP: userID=%d", userID)
	}

	// 🔥 NEW: Validasi lokasi
	inRange, distance := utils.IsWithinOfficeRange(req.Latitude, req.Longitude)
	cfg := config.Load()

	// Jika tidak dalam range dan tidak menggunakan force, tolak attendance
	if !inRange && !req.Force {
		if config.IsDevelopment() {
			log.Printf("🔴 ClockOutHTTP: Location out of range (%.0fm > %.0fm)", distance, cfg.AttendanceRadiusMeters)
		}
		utils.ErrorResponse(w, http.StatusBadRequest,
			"Lokasi di luar jangkauan kantor. Gunakan force=true untuk tetap melanjutkan.")
		return
	}

	// Tentukan status berdasarkan lokasi dan force
	status := "approved"
	if !inRange && req.Force {
		status = "forced"
		if config.IsDevelopment() {
			log.Printf("🟡 ClockOutHTTP: Using forced attendance (%.0fm > %.0fm)", distance, cfg.AttendanceRadiusMeters)
		}
	}

	var attendanceID int
	err = database.DB.QueryRow(
		"INSERT INTO attendances (user_id, type, status, latitude, longitude, photo_selfie, in_range, force_attendance, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id",
		userID, "out", status, req.Latitude, req.Longitude, req.PhotoSelfie, inRange, req.Force, time.Now(),
	).Scan(&attendanceID)

	if err != nil {
		if config.IsDevelopment() {
			log.Printf("🔴 ClockOutHTTP database error: %v", err)
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to clock out")
		return
	}

	if config.IsDevelopment() {
		log.Printf("🔵 ClockOutHTTP successful: attendanceID=%d, status=%s, inRange=%t", attendanceID, status, inRange)
	}

	response := AttendanceResponse{
		ID:              attendanceID,
		UserID:          userID,
		Type:            "out",
		Status:          status,
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		PhotoSelfie:     req.PhotoSelfie,
		InRange:         inRange,
		ForceAttendance: req.Force,
		CreatedAt:       time.Now(),
		Distance:        distance,
		MaxRadius:       cfg.AttendanceRadiusMeters,
	}

	utils.SuccessResponse(w, http.StatusOK, "Clock out successful", response)
}

// CheckLocationHTTP untuk Vercel/gorilla-mux compatibility
func CheckLocationHTTP(w http.ResponseWriter, r *http.Request) {
	if config.IsDevelopment() {
		log.Println("🔵 CheckLocationHTTP started")
	}

	defer func() {
		if err := recover(); err != nil {
			if config.IsDevelopment() {
				log.Printf("🚨 CheckLocationHTTP PANIC: %v", err)
			}
			utils.ErrorResponse(w, http.StatusInternalServerError, "Location check failed")
		}
	}()

	var req LocationCheckRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		if config.IsDevelopment() {
			log.Printf("🔴 CheckLocationHTTP decode error: %v", err)
		}
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 🔥 NEW: Gunakan location validation dari utils
	validation := utils.ValidateLocation(req.Latitude, req.Longitude)

	response := models.LocationValidationResponse{
		InRange:   validation.InRange,
		Message:   validation.Message,
		NeedForce: validation.NeedForce,
		Distance:  validation.Distance,
	}

	utils.SuccessResponse(w, http.StatusOK, "Location validation successful", response)
}

// GetAttendanceHTTP untuk Vercel/gorilla-mux compatibility
func GetAttendanceHTTP(w http.ResponseWriter, r *http.Request) {
	if config.IsDevelopment() {
		log.Println("🔵 GetAttendanceHTTP started")
	}

	defer func() {
		if err := recover(); err != nil {
			if config.IsDevelopment() {
				log.Printf("🚨 GetAttendanceHTTP PANIC: %v", err)
			}
			utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get attendance")
		}
	}()

	// Get userID from context (set by AuthMiddleware)
	userIDVal := r.Context().Value("userID")
	if userIDVal == nil {
		if config.IsDevelopment() {
			log.Printf("🔴 GetAttendanceHTTP: userID not found in context")
		}
		utils.ErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, ok := userIDVal.(int)
	if !ok {
		if config.IsDevelopment() {
			log.Printf("🔴 GetAttendanceHTTP: userID type assertion failed")
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	if config.IsDevelopment() {
		log.Printf("🔵 GetAttendanceHTTP: userID=%d", userID)
	}

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
		if config.IsDevelopment() {
			log.Printf("🔴 GetAttendanceHTTP database error: %v", err)
		}
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
			if config.IsDevelopment() {
				log.Printf("🔴 GetAttendanceHTTP scan error: %v", err)
			}
			continue
		}

		attendance := AttendanceResponse{
			ID:              att.ID,
			UserID:          att.UserID,
			Type:            att.Type,
			Status:          att.Status,
			Latitude:        att.Latitude,
			Longitude:       att.Longitude,
			PhotoSelfie:     att.PhotoSelfie,
			InRange:         att.InRange,
			ForceAttendance: att.ForceAttendance,
			CreatedAt:       att.CreatedAt,
		}
		attendances = append(attendances, attendance)
	}

	if config.IsDevelopment() {
		log.Printf("🔵 GetAttendanceHTTP successful: found %d records", len(attendances))
	}

	utils.SuccessResponse(w, http.StatusOK, "Attendance records retrieved", attendances)
}
