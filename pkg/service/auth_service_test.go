package service

import (
    "testing"

    "github.com/nepskuy/be-godplan/pkg/models"
    "github.com/nepskuy/be-godplan/pkg/utils"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "golang.org/x/crypto/bcrypt"
)

// MockUserRepository dengan method yang sesuai interface
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) CreateUser(user *models.User) error {
    args := m.Called(user)
    return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
    args := m.Called(email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(id int) (*models.User, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.User), args.Error(1)
}

func TestAuthService_Register_Success(t *testing.T) {
    mockRepo := new(MockUserRepository)
    authService := NewAuthService(mockRepo)

    // Setup expectations - user belum ada
    mockRepo.On("FindByEmail", "newuser@example.com").
        Return((*models.User)(nil), utils.ErrUserNotFound)

    mockRepo.On("CreateUser", mock.AnythingOfType("*models.User")).
        Return(nil).Run(func(args mock.Arguments) {
        user := args.Get(0).(*models.User)
        user.ID = 1 // Simulate ID assignment (int, not uint)
    })

    user := &models.User{
        Username: "newuser",
        Name:     "New User",
        Email:    "newuser@example.com",
        Password: "password123",
        Role:     "employee",
    }

    err := authService.Register(user)

    assert.NoError(t, err)
    assert.Equal(t, 1, user.ID) // FIX: Expect int, not uint
    mockRepo.AssertExpectations(t)
}

func TestAuthService_Register_EmailExists(t *testing.T) {
    mockRepo := new(MockUserRepository)
    authService := NewAuthService(mockRepo)

    existingUser := &models.User{
        ID:    1, // int
        Email: "existing@example.com",
    }

    mockRepo.On("FindByEmail", "existing@example.com").
        Return(existingUser, nil)

    user := &models.User{
        Username: "existinguser",
        Name:     "Existing User",
        Email:    "existing@example.com",
        Password: "password123",
    }

    err := authService.Register(user)

    assert.Error(t, err)
    assert.Equal(t, utils.ErrEmailExists, err)
    mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
    mockRepo := new(MockUserRepository)
    authService := NewAuthService(mockRepo)

    // Generate valid bcrypt hash
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

    existingUser := &models.User{
        ID:       1, // int
        Username: "testuser",
        Name:     "Test User",
        Email:    "test@example.com",
        Password: string(hashedPassword),
        Role:     "employee",
    }

    mockRepo.On("FindByEmail", "test@example.com").
        Return(existingUser, nil)

    token, user, err := authService.Login("test@example.com", "password123")

    assert.NoError(t, err)
    assert.NotEmpty(t, token)
    assert.Equal(t, "test@example.com", user.Email)
    assert.Empty(t, user.Password)
    mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
    mockRepo := new(MockUserRepository)
    authService := NewAuthService(mockRepo)

    mockRepo.On("FindByEmail", "nonexistent@example.com").
        Return((*models.User)(nil), utils.ErrUserNotFound)

    token, user, err := authService.Login("nonexistent@example.com", "password123")

    assert.Error(t, err)
    assert.Equal(t, utils.ErrInvalidCredentials, err)
    assert.Empty(t, token)
    assert.Nil(t, user)
    mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
    mockRepo := new(MockUserRepository)
    authService := NewAuthService(mockRepo)

    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

    existingUser := &models.User{
        ID:       1, // int
        Username: "testuser",
        Name:     "Test User",
        Email:    "test@example.com",
        Password: string(hashedPassword),
        Role:     "employee",
    }

    mockRepo.On("FindByEmail", "test@example.com").
        Return(existingUser, nil)

    token, user, err := authService.Login("test@example.com", "wrongpassword")

    assert.Error(t, err)
    assert.Equal(t, utils.ErrInvalidCredentials, err)
    assert.Empty(t, token)
    assert.Nil(t, user)
    mockRepo.AssertExpectations(t)
}
