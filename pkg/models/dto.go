package models

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" example:"admin@godplan.com"`
	Password string `json:"password" example:"password"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// UserRegistrationRequest represents user registration request
// @Description User registration request body
type UserRegistrationRequest struct {
	Username string `json:"username" example:"johndoe"`
	FullName string `json:"full_name" example:"John Doe"`
	Email    string `json:"email" example:"john@example.com"`
	Password string `json:"password" example:"password123"`
	Phone    string `json:"phone,omitempty" example:"+628123456789"`
	Role     string `json:"role,omitempty" example:"employee"`
}

// ClockInRequest represents clock-in request
type ClockInRequest struct {
	Latitude    float64 `json:"latitude" example:"-6.2088"`
	Longitude   float64 `json:"longitude" example:"106.8456"`
	PhotoSelfie string  `json:"photo_selfie" example:"base64_encoded_image"`
	Force       bool    `json:"force" example:"false"`
}

// ClockOutRequest represents clock-out request
type ClockOutRequest struct {
	Latitude    float64 `json:"latitude" example:"-6.2088"`
	Longitude   float64 `json:"longitude" example:"106.8456"`
	PhotoSelfie string  `json:"photo_selfie" example:"base64_encoded_image"`
	Force       bool    `json:"force" example:"false"`
}

// LocationCheckRequest represents location check request
type LocationCheckRequest struct {
	Latitude  float64 `json:"latitude" example:"-6.2088"`
	Longitude float64 `json:"longitude" example:"106.8456"`
}

// LocationValidationResponse represents location validation response
type LocationValidationResponse struct {
	InRange   bool    `json:"in_range"`
	Message   string  `json:"message"`
	NeedForce bool    `json:"need_force"`
	Distance  float64 `json:"distance"`
}
