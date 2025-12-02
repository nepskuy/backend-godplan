package service

import (
	"github.com/google/uuid"
	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/repository"
	"github.com/nepskuy/be-godplan/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Register(user *models.User) error {
	// Cek apakah user sudah ada
	existingUser, err := s.userRepo.GetUserByEmail(user.TenantID, user.Email)
	if err == nil && existingUser != nil {
		return utils.ErrEmailExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.ErrInternalServer
	}
	user.Password = string(hashedPassword)

	if user.Role == "" {
		user.Role = "employee"
	}

	return s.userRepo.CreateUser(user)
}

func (s *AuthService) Login(tenantID uuid.UUID, email, password string) (string, *models.User, error) {
	user, err := s.userRepo.GetUserByEmail(tenantID, email)
	if err != nil {
		return "", nil, utils.ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, utils.ErrInvalidCredentials
	}

	// Generate Token (assuming utils.GenerateToken exists and accepts UUIDs)
	// Note: The original code had "mock-jwt-token", I'll keep it as placeholder or update if I know the JWT util
	// But wait, the handler uses utils.NewJWTUtil. The service here seems to be a leftover or alternative implementation.
	// I will just return the mock token as before but with correct signature.
	token := "mock-jwt-token"

	user.Password = ""

	return token, user, nil
}
