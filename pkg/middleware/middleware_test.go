package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nepskuy/be-godplan/pkg/utils"
)

func TestCORSMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req, err := http.NewRequest("OPTIONS", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	corsHandler := CORS(handler)

	corsHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check CORS headers
	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, Authorization",
	}

	for header, expectedValue := range expectedHeaders {
		if actualValue := rr.Header().Get(header); actualValue != expectedValue {
			t.Errorf("Header %s: got %v want %v", header, actualValue, expectedValue)
		}
	}
}

func TestAuthMiddlewarePublicRoutes(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.SuccessResponse(w, http.StatusOK, "OK", nil)
	})

	// Test public route - should pass without auth
	req, err := http.NewRequest("POST", "/api/v1/auth/login", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	authHandler := AuthMiddleware(handler)

	authHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Public route should pass without auth: got %v want %v", status, http.StatusOK)
	}
}
