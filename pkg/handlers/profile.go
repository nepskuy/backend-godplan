package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nepskuy/be-godplan/pkg/repository"
)

// GinGetProfile handler untuk Gin
func GinGetProfile(userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		userIDVal, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		tenantIDStr := c.GetString("tenant_id")
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid tenant ID"})
			return
		}

		// Convert userID to uuid.UUID
		var id uuid.UUID
		switch v := userIDVal.(type) {
		case uuid.UUID:
			id = v
		case string:
			parsedID, err := uuid.Parse(v)
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
		user, err := userRepo.GetUserWithEmployeeData(tenantID, id)
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
