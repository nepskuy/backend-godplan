package utils

import (
	"github.com/gin-gonic/gin"
)

type GinResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func GinSuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, GinResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func GinErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, GinResponse{
		Success: false,
		Error:   message,
	})
}
