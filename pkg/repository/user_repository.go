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
	UpdateUser(user *models.User) error
}

type UserRepository struct {
	db *sql.DB
}

var _ UserRepositoryInterface = (*UserRepository)(nil)

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user *models.User) error {
	query := `
		INSERT INTO godplan.users 
		(username, email, password, role, full_name, phone, avatar_url, is_active, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.FullName,
		user.Phone,
		user.AvatarURL,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return utils.ErrInternalServer
	}
	return nil
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT 
			id, username, email, password, role, full_name, phone, 
			avatar_url, is_active, created_at, updated_at 
		FROM godplan.users 
		WHERE email = $1 AND is_active = true
	`

	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.FullName,
		&user.Phone,
		&user.AvatarURL,
		&user.IsActive,
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

func (r *UserRepository) FindByID(id int) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT 
			id, username, email, role, full_name, phone, 
			avatar_url, is_active, created_at, updated_at 
		FROM godplan.users 
		WHERE id = $1 AND is_active = true
	`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.FullName,
		&user.Phone,
		&user.AvatarURL,
		&user.IsActive,
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

func (r *UserRepository) GetUserByID(userID int64) (*models.User, error) {
	user := &models.User{}

	query := `
		SELECT 
			id, username, email, role, full_name, phone, 
			avatar_url, is_active, created_at, updated_at 
		FROM godplan.users 
		WHERE id = $1 AND is_active = true
	`

	err := r.db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.FullName,
		&user.Phone,
		&user.AvatarURL,
		&user.IsActive,
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

func (r *UserRepository) UpdateUser(user *models.User) error {
	query := `
		UPDATE godplan.users 
		SET username = $1, email = $2, role = $3, full_name = $4, 
			phone = $5, avatar_url = $6, is_active = $7, updated_at = $8
		WHERE id = $9
	`

	result, err := r.db.Exec(
		query,
		user.Username,
		user.Email,
		user.Role,
		user.FullName,
		user.Phone,
		user.AvatarURL,
		user.IsActive,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		return utils.ErrInternalServer
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.ErrInternalServer
	}

	if rowsAffected == 0 {
		return utils.ErrUserNotFound
	}

	return nil
}

// GetUserWithEmployeeData - Get user data with employee information (jika diperlukan)
func (r *UserRepository) GetUserWithEmployeeData(userID int64) (*models.User, *models.Employee, error) {
	user := &models.User{}
	employee := &models.Employee{}

	// Get user data
	userQuery := `
		SELECT id, username, email, role, full_name, phone, avatar_url, is_active, created_at, updated_at 
		FROM godplan.users 
		WHERE id = $1 AND is_active = true
	`

	err := r.db.QueryRow(userQuery, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.FullName,
		&user.Phone,
		&user.AvatarURL,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, nil, err
	}

	// Get employee data if exists
	employeeQuery := `
		SELECT id, user_id, employee_id, department_id, position_id, base_salary, 
			   join_date, employment_type, work_schedule, created_at, updated_at
		FROM godplan.employees 
		WHERE user_id = $1
	`

	err = r.db.QueryRow(employeeQuery, userID).Scan(
		&employee.ID,
		&employee.UserID,
		&employee.EmployeeID,
		&employee.DepartmentID,
		&employee.PositionID,
		&employee.BaseSalary,
		&employee.JoinDate,
		&employee.EmploymentType,
		&employee.WorkSchedule,
		&employee.CreatedAt,
		&employee.UpdatedAt,
	)

	// Employee data might not exist, so we don't return error if not found
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, err
	}

	return user, employee, nil
}
