package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nepskuy/be-godplan/pkg/repository"
)

// GinGetProfile handler untuk Gin
func GinGetProfile(userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Convert userID to int64
		var id int64
		switch v := userID.(type) {
		case int:
			id = int64(v)
		case int64:
			id = v
		case float64:
			id = int64(v)
		case string:
			parsedID, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
				return
			}
			id = parsedID
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID type"})
			return
		}

		// Get user with employee data from repository
		user, err := userRepo.GetUserWithEmployeeData(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile"})
			return
		}

		if user == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}
