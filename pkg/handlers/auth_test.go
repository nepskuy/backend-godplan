package handlers

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestRegisterHandlerInvalidJSON(t *testing.T) {
    req, err := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer([]byte("invalid json")))
    if err != nil {
        t.Fatal(err)
    }
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(Register)

    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusBadRequest {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
    }
}

func TestRegisterHandlerMissingFields(t *testing.T) {
    registerData := map[string]string{
        "email": "test@example.com",
    }

    jsonData, _ := json.Marshal(registerData)

    req, err := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
    if err != nil {
        t.Fatal(err)
    }
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(Register)

    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusBadRequest {
        t.Errorf("Register with missing fields returned wrong status code: got %v want %v", status, http.StatusBadRequest)
    }
}

func TestRegisterHandlerValidRequest(t *testing.T) {
    registerData := map[string]string{
        "username": "testuser",
        "name":     "Test User", 
        "email":    "test@example.com",
        "password": "password123",
    }

    jsonData, _ := json.Marshal(registerData)

    req, err := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
    if err != nil {
        t.Fatal(err)
    }
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(Register)

    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusCreated {
        t.Errorf("Valid register request returned wrong status code: got %v want %v", status, http.StatusCreated)
    }
}

func TestTaskHandlers(t *testing.T) {
    t.Run("GetTasks", func(t *testing.T) {
        req, err := http.NewRequest("GET", "/api/v1/tasks", nil)
        if err != nil {
            t.Fatal(err)
        }

        rr := httptest.NewRecorder()
        handler := http.HandlerFunc(GetTasks)

        handler.ServeHTTP(rr, req)

        if status := rr.Code; status != http.StatusOK {
            t.Errorf("GetTasks returned wrong status code: got %v want %v", status, http.StatusOK)
        }
    })

    t.Run("CreateTask", func(t *testing.T) {
        taskData := map[string]interface{}{
            "title":       "Test Task",
            "description": "Test Description",
        }

        jsonData, _ := json.Marshal(taskData)

        req, err := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(jsonData))
        if err != nil {
            t.Fatal(err)
        }
        req.Header.Set("Content-Type", "application/json")

        rr := httptest.NewRecorder()
        handler := http.HandlerFunc(CreateTask)

        handler.ServeHTTP(rr, req)

        if status := rr.Code; status != http.StatusCreated {
            t.Errorf("CreateTask returned wrong status code: got %v want %v", status, http.StatusCreated)
        }
    })

    t.Run("CreateTaskMissingTitle", func(t *testing.T) {
        taskData := map[string]interface{}{
            "description": "Test Description", 
        }

        jsonData, _ := json.Marshal(taskData)

        req, err := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(jsonData))
        if err != nil {
            t.Fatal(err)
        }
        req.Header.Set("Content-Type", "application/json")

        rr := httptest.NewRecorder()
        handler := http.HandlerFunc(CreateTask)

        handler.ServeHTTP(rr, req)

        if status := rr.Code; status != http.StatusBadRequest {
            t.Errorf("CreateTask with missing title returned wrong status code: got %v want %v", status, http.StatusBadRequest)
        }
    })
}

func TestLoginHandler(t *testing.T) {
    t.Run("LoginValidCredentials", func(t *testing.T) {
        loginData := map[string]string{
            "email":    "test@example.com",
            "password": "password123",
        }

        jsonData, _ := json.Marshal(loginData)

        req, err := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
        if err != nil {
            t.Fatal(err)
        }
        req.Header.Set("Content-Type", "application/json")

        rr := httptest.NewRecorder()
        handler := http.HandlerFunc(Login)

        handler.ServeHTTP(rr, req)

        if status := rr.Code; status != http.StatusOK {
            t.Errorf("Login with valid credentials returned wrong status code: got %v want %v", status, http.StatusOK)
        }
    })

    t.Run("LoginMissingFields", func(t *testing.T) {
        loginData := map[string]string{
            "email": "test@example.com",
        }

        jsonData, _ := json.Marshal(loginData)

        req, err := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
        if err != nil {
            t.Fatal(err)
        }
        req.Header.Set("Content-Type", "application/json")

        rr := httptest.NewRecorder()
        handler := http.HandlerFunc(Login)

        handler.ServeHTTP(rr, req)

        if status := rr.Code; status != http.StatusBadRequest {
            t.Errorf("Login with missing fields returned wrong status code: got %v want %v", status, http.StatusBadRequest)
        }
    })
}
