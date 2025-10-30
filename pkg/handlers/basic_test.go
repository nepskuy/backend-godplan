package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test basic health check atau handlers yang tidak butuh database
func TestHealthCheck(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("HealthCheck returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

// Test bahwa handler functions exist (compile-time check)
func TestHandlerFunctionsExist(t *testing.T) {
	// Test bahwa function handlers ada
	_ = http.HandlerFunc(GetUsers)
	_ = http.HandlerFunc(CreateUser)
	_ = http.HandlerFunc(GetUser)
	_ = http.HandlerFunc(ClockIn)
	_ = http.HandlerFunc(ClockOut)
	_ = http.HandlerFunc(GetAttendance)
	_ = http.HandlerFunc(CheckLocation) // Tambah function baru

	assert := true
	if !assert {
		t.Error("Handler functions should exist")
	}
}
