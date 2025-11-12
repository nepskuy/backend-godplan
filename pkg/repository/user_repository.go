package repository

import (
	"database/sql"

	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

// UserRepositoryInterface mendefinisikan contract untuk user repository
type UserRepositoryInterface interface {
	CreateUser(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id int) (*models.User, error)
	GetUserByID(userID int64) (*models.User, error)
}

type UserRepository struct {
	db *sql.DB
}

var _ UserRepositoryInterface = (*UserRepository)(nil)

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user *models.User) error {
	query := `INSERT INTO users (username, name, email, password, role) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	err := r.db.QueryRow(query, user.Username, user.Name, user.Email, user.Password, user.Role).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return utils.ErrInternalServer
	}
	return nil
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, name, email, password, role, created_at FROM users WHERE email = $1`
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, utils.ErrUserNotFound
	}
	if err != nil {
		return nil, utils.ErrInternalServer
	}
	return user, nil
}

func (r *UserRepository) FindByID(id int) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, name, email, role, created_at FROM users WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Email,
		&user.Role,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, utils.ErrUserNotFound
	}
	if err != nil {
		return nil, utils.ErrInternalServer
	}
	return user, nil
}

func (r *UserRepository) GetUserByID(userID int64) (*models.User, error) {
	user := &models.User{}

	query := `
		SELECT 
			id, username, name, email, 
			COALESCE(employee_id, '') as employee_id,
			COALESCE(nisn, '') as nisn, 
			COALESCE(department, '') as department,
			COALESCE(position, '') as position,
			COALESCE(status, '') as status,
			COALESCE(phone, '') as phone,
			role, 
			created_at, 
			COALESCE(updated_at, created_at) as updated_at
		FROM users 
		WHERE id = $1
	`

	err := r.db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Email,
		&user.EmployeeID,
		&user.NISN,
		&user.Department,
		&user.Position,
		&user.Status,
		&user.Phone,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, utils.ErrUserNotFound
	}
	if err != nil {
		return nil, utils.ErrInternalServer
	}
	return user, nil
}
