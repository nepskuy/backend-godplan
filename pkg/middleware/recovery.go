package middleware

import (
	"log"
	"net/http"

	"github.com/nepskuy/be-godplan/pkg/utils"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("🚨 PANIC RECOVERED: %v", err)
				utils.ErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
