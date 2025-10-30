package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nepskuy/be-godplan/pkg/database"
	"github.com/nepskuy/be-godplan/pkg/models"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func GetUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT id, username, email, created_at FROM users")
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch users"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
		if err != nil {
			http.Error(w, `{"error": "Failed to scan user"}`, http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error": "Failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	var id int
	err = database.DB.QueryRow(
		"INSERT INTO users (username, password, email) VALUES ($1, $2, $3) RETURNING id",
		user.Username, string(hashedPassword), user.Email,
	).Scan(&id)

	if err != nil {
		http.Error(w, `{"error": "Failed to create user"}`, http.StatusInternalServerError)
		return
	}

	user.ID = id
	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, `{"error": "Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	var user models.User
	err = database.DB.QueryRow(
		"SELECT id, username, email, created_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)

	if err != nil {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
