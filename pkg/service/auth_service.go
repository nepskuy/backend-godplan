package service

import (
	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/repository"
	"github.com/nepskuy/be-godplan/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo repository.UserRepositoryInterface
}

func NewAuthService(userRepo repository.UserRepositoryInterface) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Register(user *models.User) error {

	_, err := s.userRepo.FindByEmail(user.Email)
	if err == nil {
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

func (s *AuthService) Login(email, password string) (string, *models.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", nil, utils.ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, utils.ErrInvalidCredentials
	}

	token := "mock-jwt-token"

	user.Password = ""

	return token, user, nil
}
