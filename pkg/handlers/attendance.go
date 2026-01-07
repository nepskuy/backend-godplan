package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nepskuy/be-godplan/pkg/config"
	"github.com/nepskuy/be-godplan/pkg/database"
	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

type ClockInRequest struct {
	Latitude    float64 `json:"latitude" example:"-6.2088"`
	Longitude   float64 ` json:"longitude" example:"106.8456"`
	PhotoSelfie string  `json:"photo_selfie" example:"base64_encoded_image"`
	Force       bool    `json:"force" example:"false"`
	Accuracy    float64 `json:"accuracy" example:"15.5"` // GPS accuracy in meters
}

type ClockOutRequest struct {
	Latitude    float64 `json:"latitude" example:"-6.2088"`
	Longitude   float64 `json:"longitude" example:"106.8456"`
	PhotoSelfie string  `json:"photo_selfie" example:"base64_encoded_image"`
	Force       bool    `json:"force" example:"false"`
	Accuracy    float64 `json:"accuracy" example:"15.5"` // GPS accuracy in meters
}

type LocationCheckRequest struct {
	Latitude  float64 `json:"latitude" example:"-6.2088"`
	Longitude float64 `json:"longitude" example:"106.8456"`
	Accuracy  float64 `json:"accuracy" example:"15.5"` // GPS accuracy in meters
}

type AttendanceResponse struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"user_id"`
	Type            string    `json:"type"`
	Status          string    `json:"status"`
	Date            string    `json:"date"`           // Added for mobile
	Time            string    `json:"time"`           // Added for mobile
	LocationName    string    `json:"location_name"` // Added for mobile
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	PhotoSelfie     string    `json:"photo_selfie,omitempty"`
	InRange         bool      `json:"in_range"`
	ForceAttendance bool      `json:"force_attendance"`
	CreatedAt       time.Time `json:"created_at"`
	Distance        float64   `json:"distance,omitempty"`
	MaxRadius       float64   `json:"max_radius,omitempty"`
}

// CheckLocation godoc
// @Summary Check location validity
// @Description Validate if user location is within office range
// @Tags attendance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body LocationCheckRequest true "Location coordinates"
// @Success 200 {object} utils.GinResponse
// @Failure 400 {object} utils.GinResponse
// @Router /attendance/check-location [post]
func CheckLocation(c *gin.Context) {
	if config.IsDevelopment() {
		fmt.Println("🔵 CheckLocation started")
	}

	defer func() {
		if err := recover(); err != nil {
			if config.IsDevelopment() {
				fmt.Printf("🚨 CheckLocation PANIC: %v\n", err)
			}
			utils.GinErrorResponse(c, http.StatusInternalServerError, "Location check failed")
		}
	}()

	var req LocationCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if config.IsDevelopment() {
			fmt.Printf("🔴 CheckLocation bind error: %v\n", err)
		}
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 🚀 NEW: Gunakan adaptive location validation dengan GPS accuracy
	validation := utils.ValidateLocationAdaptive(req.Latitude, req.Longitude, req.Accuracy)

	// Enhanced response dengan informasi GPS quality
	response := map[string]interface{}{
		"in_range":         validation.InRange,
		"message":          validation.Message,
		"detailed_message": validation.DetailedMessage,
		"need_force":       validation.NeedForce,
		"distance":         validation.Distance,
		"max_radius":       validation.MaxRadius,
		"adaptive_radius":  validation.AdaptiveRadius,
		"gps_accuracy":     validation.GPSAccuracy,
		"gps_quality":      validation.GPSQuality,
		"recommendation":   validation.Recommendation,
	}

	utils.GinSuccessResponse(c, http.StatusOK, "Location validation successful", response)
}

// ClockIn godoc
// @Summary Clock in attendance
// @Description Record user clock-in with location validation
// @Tags attendance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ClockInRequest true "Clock-in data"
// @Success 201 {object} utils.GinResponse
// @Failure 400 {object} utils.GinResponse
// @Failure 401 {object} utils.GinResponse
// @Failure 500 {object} utils.GinResponse
// @Router /attendance/clock-in [post]
func ClockIn(c *gin.Context) {
	if config.IsDevelopment() {
		fmt.Println("🔵 ClockIn started")
	}

	defer func() {
		if err := recover(); err != nil {
			if config.IsDevelopment() {
				fmt.Printf("🚨 ClockIn PANIC: %v\n", err)
			}
			utils.GinErrorResponse(c, http.StatusInternalServerError, "Clock in failed")
		}
	}()

	var req ClockInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if config.IsDevelopment() {
			fmt.Printf("🔴 ClockIn bind error: %v\n", err)
		}
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	userIDVal, exists := c.Get("userID")
	if !exists {
		utils.GinErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusUnauthorized, "Invalid tenant ID")
		return
	}

	var userID uuid.UUID
	switch v := userIDVal.(type) {
	case uuid.UUID:
		userID = v
	case string:
		userID, err = uuid.Parse(v)
		if err != nil {
			utils.GinErrorResponse(c, http.StatusInternalServerError, "Invalid user ID format")
			return
		}
	default:
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Invalid user ID type")
		return
	}

	// Validate location with adaptive threshold based on GPS accuracy
	inRange, distance, adaptiveRadius := utils.IsWithinOfficeRangeAdaptive(req.Latitude, req.Longitude, req.Accuracy)
	cfg := config.Load()

	if !inRange && !req.Force {
		utils.GinErrorResponse(c, http.StatusBadRequest,
			fmt.Sprintf("Lokasi di luar jangkauan kantor. Jarak: %.0fm | Jangkauan adaptive: %.0fm (GPS accuracy: %.0fm). Gunakan force=true jika Anda yakin sudah di kantor.",
				distance, adaptiveRadius, req.Accuracy))
		return
	}

	// Determine status
	status := "approved"
	if !inRange && req.Force {
		status = "pending_forced"
	}

	// Check if already clocked in today
	var existingID uuid.UUID
	checkErr := database.DB.QueryRow(
		`SELECT id FROM godplan.attendances 
		 WHERE user_id = $1 AND tenant_id = $2 AND attendance_date = CURRENT_DATE`,
		userID, tenantID,
	).Scan(&existingID)

	if checkErr == nil {
		utils.GinErrorResponse(c, http.StatusBadRequest, "Sudah melakukan Clock In hari ini")
		return
	}

	// Insert new attendance record with correct schema columns
	var attendanceID uuid.UUID
	now := time.Now()
	err = database.DB.QueryRow(
		`INSERT INTO godplan.attendances (
			tenant_id, user_id, type, status, 
			check_in_time, check_in_lat, check_in_lng, check_in_photo,
			attendance_date, in_range, force_attendance, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_DATE, $9, $10, $11) RETURNING id`,
		tenantID, userID, "CheckIn", status,
		now, req.Latitude, req.Longitude, req.PhotoSelfie,
		inRange, req.Force, now,
	).Scan(&attendanceID)

	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("🔴 ClockIn database error: %v\n", err)
		}
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to clock in: "+err.Error())
		return
	}

	response := AttendanceResponse{
		ID:              attendanceID,
		UserID:          userID,
		Type:            "CheckIn",
		Status:          status,
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		PhotoSelfie:     req.PhotoSelfie,
		InRange:         inRange,
		ForceAttendance: req.Force,
		CreatedAt:       now,
		Distance:        distance,
		MaxRadius:       cfg.AttendanceRadiusMeters,
	}

	utils.GinSuccessResponse(c, http.StatusCreated, "Clock in successful", response)
}

// ClockOut godoc
// @Summary Clock out attendance
// @Description Record user clock-out with location validation
// @Tags attendance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ClockOutRequest true "Clock-out data"
// @Success 200 {object} utils.GinResponse
// @Failure 400 {object} utils.GinResponse
// @Failure 401 {object} utils.GinResponse
// @Failure 500 {object} utils.GinResponse
// @Router /attendance/clock-out [post]
func ClockOut(c *gin.Context) {
	if config.IsDevelopment() {
		fmt.Println("🔵 ClockOut started")
	}

	defer func() {
		if err := recover(); err != nil {
			if config.IsDevelopment() {
				fmt.Printf("🚨 ClockOut PANIC: %v\n", err)
			}
			utils.GinErrorResponse(c, http.StatusInternalServerError, "Clock out failed")
		}
	}()

	var req ClockOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	userIDVal, exists := c.Get("userID")
	if !exists {
		utils.GinErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusUnauthorized, "Invalid tenant ID")
		return
	}

	var userID uuid.UUID
	switch v := userIDVal.(type) {
	case uuid.UUID:
		userID = v
	case string:
		userID, err = uuid.Parse(v)
		if err != nil {
			utils.GinErrorResponse(c, http.StatusInternalServerError, "Invalid user ID format")
			return
		}
	default:
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Invalid user ID type")
		return
	}

	// Validate location with adaptive threshold based on GPS accuracy
	inRange, distance, adaptiveRadius := utils.IsWithinOfficeRangeAdaptive(req.Latitude, req.Longitude, req.Accuracy)
	cfg := config.Load()

	if !inRange && !req.Force {
		utils.GinErrorResponse(c, http.StatusBadRequest,
			fmt.Sprintf("Lokasi di luar jangkauan kantor. Jarak: %.0fm | Jangkauan adaptive: %.0fm (GPS accuracy: %.0fm). Gunakan force=true jika Anda yakin sudah di kantor.",
				distance, adaptiveRadius, req.Accuracy))
		return
	}

	// Find existing clock-in record for today
	var attendanceID uuid.UUID
	var checkInTime time.Time
	findErr := database.DB.QueryRow(
		`SELECT id, check_in_time FROM godplan.attendances 
		 WHERE user_id = $1 AND tenant_id = $2 AND attendance_date = CURRENT_DATE AND check_out_time IS NULL`,
		userID, tenantID,
	).Scan(&attendanceID, &checkInTime)

	if findErr != nil {
		utils.GinErrorResponse(c, http.StatusBadRequest, "Belum melakukan Clock In hari ini atau sudah Clock Out")
		return
	}

	// Calculate total hours
	now := time.Now()
	totalHours := now.Sub(checkInTime).Hours()

	// Update status if force is used
	status := "approved"
	if !inRange && req.Force {
		status = "pending_forced"
	}

	// Update existing record with checkout data
	_, err = database.DB.Exec(
		`UPDATE godplan.attendances SET
			check_out_time = $1,
			check_out_lat = $2,
			check_out_lng = $3,
			check_out_photo = $4,
			total_hours = $5,
			type = 'CheckOut',
			status = $6,
			updated_at = $7
		WHERE id = $8 AND tenant_id = $9`,
		now, req.Latitude, req.Longitude, req.PhotoSelfie,
		totalHours, status, now, attendanceID, tenantID,
	)

	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("🔴 ClockOut database error: %v\n", err)
		}
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to clock out: "+err.Error())
		return
	}

	response := AttendanceResponse{
		ID:              attendanceID,
		UserID:          userID,
		Type:            "CheckOut",
		Status:          status,
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		PhotoSelfie:     req.PhotoSelfie,
		InRange:         inRange,
		ForceAttendance: req.Force,
		CreatedAt:       now,
		Distance:        distance,
		MaxRadius:       cfg.AttendanceRadiusMeters,
	}

	utils.GinSuccessResponse(c, http.StatusOK, "Clock out successful", response)
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
// @Success 200 {object} utils.GinResponse
// @Failure 400 {object} utils.GinResponse
// @Failure 401 {object} utils.GinResponse
// @Failure 500 {object} utils.GinResponse
// @Router /attendance [get]
func GetAttendance(c *gin.Context) {
	if config.IsDevelopment() {
		fmt.Println("🔵 GetAttendance started")
	}

	defer func() {
		if err := recover(); err != nil {
			if config.IsDevelopment() {
				fmt.Printf("🚨 GetAttendance PANIC: %v\n", err)
			}
			utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to get attendance")
		}
	}()

	userIDVal, exists := c.Get("userID")
	if !exists {
		if config.IsDevelopment() {
			fmt.Println("🔴 GetAttendance: userID not found in context")
		}
		utils.GinErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusUnauthorized, "Invalid tenant ID")
		return
	}

	var userID uuid.UUID
	switch v := userIDVal.(type) {
	case uuid.UUID:
		userID = v
	case string:
		userID, err = uuid.Parse(v)
		if err != nil {
			utils.GinErrorResponse(c, http.StatusInternalServerError, "Invalid user ID format")
			return
		}
	default:
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Invalid user ID type")
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("🔵 GetAttendance: userID=%s\n", userID)
	}

	dateFilter := c.Query("date")
	limitStr := c.Query("limit")

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

	if dateFilter != "" {
		rows, err = database.DB.Query(
			`SELECT id, user_id, type, status, attendance_date, 
				COALESCE(TO_CHAR(check_in_time, 'HH24:MI'), TO_CHAR(created_at, 'HH24:MI')) as time,
				check_in_lat as latitude, check_in_lng as longitude, check_in_photo as photo_selfie, in_range, force_attendance, created_at 
			FROM godplan.attendances 
			WHERE user_id = $1 AND tenant_id = $2 AND attendance_date = $3 
			ORDER BY created_at DESC LIMIT $4`,
			userID, tenantID, dateFilter, limit,
		)
	} else {
		rows, err = database.DB.Query(
			`SELECT id, user_id, type, status, attendance_date, 
				COALESCE(TO_CHAR(check_in_time, 'HH24:MI'), TO_CHAR(created_at, 'HH24:MI')) as time,
				check_in_lat as latitude, check_in_lng as longitude, check_in_photo as photo_selfie, in_range, force_attendance, created_at 
			FROM godplan.attendances 
			WHERE user_id = $1 AND tenant_id = $2 
			ORDER BY created_at DESC LIMIT $3`,
			userID, tenantID, limit,
		)
	}

	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("🔴 GetAttendance database error: %v\n", err)
		}
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to fetch attendance records")
		return
	}
	defer rows.Close()

	var attendances []AttendanceResponse
	for rows.Next() {
		var att models.Attendance
		var attendanceDate string
		var attendanceTime string
		
		err := rows.Scan(
			&att.ID, &att.UserID, &att.Type, &att.Status,
			&attendanceDate, &attendanceTime,
			&att.Latitude, &att.Longitude, &att.PhotoSelfie,
			&att.InRange, &att.ForceAttendance, &att.CreatedAt,
		)
		if err != nil {
			if config.IsDevelopment() {
				fmt.Printf("🔴 GetAttendance scan error: %v\n", err)
			}
			continue
		}

		attendance := AttendanceResponse{
			ID:              att.ID,
			UserID:          att.UserID,
			Type:            att.Type,
			Status:          att.Status,
			Date:            attendanceDate,
			Time:            attendanceTime,
			LocationName:    "Kantor Pusat Godplan", // Default location name
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
		fmt.Printf("🔵 GetAttendance successful: found %d records\n", len(attendances))
	}

	utils.GinSuccessResponse(c, http.StatusOK, "Attendance records retrieved", attendances)
}
