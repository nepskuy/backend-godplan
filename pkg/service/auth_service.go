package service

import (
	"github.com/google/uuid"
	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/repository"
	"github.com/nepskuy/be-godplan/pkg/utils"
	"golang.org/x/crypto/bcrypt"
	"os"
)

type AuthService struct {
	userRepo *repository.UserRepository
	jwtUtil  *utils.JWTUtil
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-key-change-in-production"
	}
	return &AuthService{
		userRepo: userRepo,
		jwtUtil:  utils.NewJWTUtil(jwtSecret),
	}
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

	// Generate real JWT token
	token, err := s.jwtUtil.GenerateToken(user.ID, user.Email, user.Role, user.TenantID)
	if err != nil {
		return "", nil, utils.ErrInternalServer
	}

	user.Password = ""

	return token, user, nil
}
