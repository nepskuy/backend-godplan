package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/nepskuy/be-godplan/pkg/utils"
)

// Simpan JWT util instance
var jwtUtil = utils.NewJWTUtil("your-secret-key-change-in-production")

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for public routes
		if r.URL.Path == "/api/v1/auth/login" || r.URL.Path == "/api/v1/auth/register" || r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ErrorResponse(w, http.StatusUnauthorized, "Invalid authorization format")
			return
		}

		tokenString := parts[1]

		// Validate token and get claims
		claims, err := jwtUtil.ValidateToken(tokenString)
		if err != nil {
			utils.ErrorResponse(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Extract userID from claims struct
		userID := int(claims.UserID)

		// Create new context with userID
		ctx := context.WithValue(r.Context(), "userID", userID)

		// Token valid, continue to next handler with new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
