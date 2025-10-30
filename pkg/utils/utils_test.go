package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseFunctions(t *testing.T) {
	t.Run("JSONResponse success", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]string{"test": "data"}

		JSONResponse(w, http.StatusOK, data)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		w := httptest.NewRecorder()

		ErrorResponse(w, http.StatusBadRequest, "Test error")

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("SuccessResponse", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]string{"user": "john"}

		SuccessResponse(w, http.StatusCreated, "User created", data)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}
	})
}

func TestAppError(t *testing.T) {
	err := ErrInvalidCredentials

	if err.Error() != "Invalid credentials" {
		t.Errorf("Expected error message 'Invalid credentials', got '%s'", err.Error())
	}

	if err.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, err.Code)
	}
}
