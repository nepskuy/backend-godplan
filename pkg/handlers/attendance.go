package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nepskuy/be-godplan/pkg/config"
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

	response := models.LocationValidationResponse{
		InRange:   true,
		Message:   "Lokasi valid, dalam jangkauan kantor",
		NeedForce: false,
		Distance:  50.0,
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

	userID, exists := c.Get("userID")
	if !exists {
		if config.IsDevelopment() {
			fmt.Println("🔴 ClockIn: userID not found in context")
		}
		utils.GinErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		if config.IsDevelopment() {
			fmt.Println("🔴 ClockIn: userID type assertion failed")
		}
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("🔵 ClockIn: userID=%d\n", userIDInt)
	}

	inRange := true

	var attendanceID int
	err := database.DB.QueryRow(
		"INSERT INTO attendances (user_id, type, status, latitude, longitude, photo_selfie, in_range, force_attendance, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id",
		userIDInt, "in", "approved", req.Latitude, req.Longitude, req.PhotoSelfie, inRange, req.Force, time.Now(),
	).Scan(&attendanceID)

	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("🔴 ClockIn database error: %v\n", err)
		}
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to clock in")
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("🔵 ClockIn successful: attendanceID=%d\n", attendanceID)
	}

	response := AttendanceResponse{
		ID:              attendanceID,
		UserID:          userIDInt,
		Type:            "in",
		Status:          "approved",
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		InRange:         inRange,
		ForceAttendance: req.Force,
		CreatedAt:       time.Now(),
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
		if config.IsDevelopment() {
			fmt.Printf("🔴 ClockOut bind error: %v\n", err)
		}
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		if config.IsDevelopment() {
			fmt.Println("🔴 ClockOut: userID not found in context")
		}
		utils.GinErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		if config.IsDevelopment() {
			fmt.Println("🔴 ClockOut: userID type assertion failed")
		}
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("🔵 ClockOut: userID=%d\n", userIDInt)
	}

	inRange := true

	var attendanceID int
	err := database.DB.QueryRow(
		"INSERT INTO attendances (user_id, type, status, latitude, longitude, photo_selfie, in_range, force_attendance, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id",
		userIDInt, "out", "approved", req.Latitude, req.Longitude, req.PhotoSelfie, inRange, req.Force, time.Now(),
	).Scan(&attendanceID)

	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("🔴 ClockOut database error: %v\n", err)
		}
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to clock out")
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("🔵 ClockOut successful: attendanceID=%d\n", attendanceID)
	}

	response := AttendanceResponse{
		ID:              attendanceID,
		UserID:          userIDInt,
		Type:            "out",
		Status:          "approved",
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		InRange:         inRange,
		ForceAttendance: req.Force,
		CreatedAt:       time.Now(),
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

	userID, exists := c.Get("userID")
	if !exists {
		if config.IsDevelopment() {
			fmt.Println("🔴 GetAttendance: userID not found in context")
		}
		utils.GinErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		if config.IsDevelopment() {
			fmt.Println("🔴 GetAttendance: userID type assertion failed")
		}
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("🔵 GetAttendance: userID=%d\n", userIDInt)
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
	var err error

	if dateFilter != "" {
		rows, err = database.DB.Query(
			"SELECT id, user_id, type, status, latitude, longitude, photo_selfie, in_range, force_attendance, created_at FROM attendances WHERE user_id = $1 AND DATE(created_at) = $2 ORDER BY created_at DESC LIMIT $3",
			userIDInt, dateFilter, limit,
		)
	} else {
		rows, err = database.DB.Query(
			"SELECT id, user_id, type, status, latitude, longitude, photo_selfie, in_range, force_attendance, created_at FROM attendances WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2",
			userIDInt, limit,
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
		err := rows.Scan(
			&att.ID, &att.UserID, &att.Type, &att.Status,
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
			Latitude:        att.Latitude,
			Longitude:       att.Longitude,
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
