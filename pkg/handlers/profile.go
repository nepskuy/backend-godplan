package handlers

import (
	"log"
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
			log.Printf("[Profile] Error: userID not found in context")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized - no user ID"})
			return
		}
		log.Printf("[Profile] userID from context: %v (type: %T)", userIDVal, userIDVal)

		tenantIDStr := c.GetString("tenant_id")
		log.Printf("[Profile] tenant_id from context: %s", tenantIDStr)
		
		if tenantIDStr == "" {
			log.Printf("[Profile] Error: tenant_id is empty")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid tenant ID - empty"})
			return
		}
		
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			log.Printf("[Profile] Error parsing tenant_id '%s': %v", tenantIDStr, err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid tenant ID format"})
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
				log.Printf("[Profile] Error parsing userID string '%s': %v", v, err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
				return
			}
			id = parsedID
		default:
			log.Printf("[Profile] Error: userID has unexpected type %T", userIDVal)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID type"})
			return
		}

		log.Printf("[Profile] Fetching user data for userID=%s, tenantID=%s", id.String(), tenantID.String())

		// Get user with employee data from repository
		user, err := userRepo.GetUserWithEmployeeData(tenantID, id)
		if err != nil {
			log.Printf("[Profile] Error from GetUserWithEmployeeData: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile", "details": err.Error()})
			return
		}

		if user == nil {
			log.Printf("[Profile] User not found for userID=%s, tenantID=%s", id.String(), tenantID.String())
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		log.Printf("[Profile] Successfully fetched profile for user: %s", user.Email)
		c.JSON(http.StatusOK, user)
	}
}
