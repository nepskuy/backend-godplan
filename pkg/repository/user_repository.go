package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nepskuy/be-godplan/pkg/config"
	"github.com/nepskuy/be-godplan/pkg/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetUserByID mendapatkan user berdasarkan ID
func (r *UserRepository) GetUserByID(tenantID uuid.UUID, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(`
		SELECT id, tenant_id, username, email, password, role, name, phone, avatar_url, is_active, created_at, updated_at
		FROM godplan.users 
		WHERE id = $1 AND tenant_id = $2 AND is_active = true
	`, id, tenantID).Scan(
		&user.ID,
		&user.TenantID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Name,
		&user.Phone,
		&user.AvatarURL,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// GetUserWithEmployeeData returns user data with employee information
func (r *UserRepository) GetUserWithEmployeeData(tenantID uuid.UUID, userID uuid.UUID) (*models.User, error) {
	query := `
		SELECT 
			u.id, u.tenant_id, u.username, u.email, u.name, u.phone, u.role,
			u.avatar_url, u.is_active, u.created_at, u.updated_at,
			e.employee_id, e.join_date, e.employment_type, e.work_schedule,
			d.name as department_name, 
			p.name as position_name,
			CASE 
				WHEN e.employment_type = 'full_time' THEN 'Aktif'
				WHEN e.employment_type = 'contract' THEN 'Kontrak'
				WHEN e.employment_type = 'probation' THEN 'Percobaan' 
				ELSE 'Tidak Terdefinisi'
			END as status
		FROM godplan.users u
		LEFT JOIN godplan.employees e ON e.user_id = u.id AND e.tenant_id = u.tenant_id
		LEFT JOIN godplan.departments d ON e.department_id = d.id AND e.tenant_id = d.tenant_id
		LEFT JOIN godplan.positions p ON e.position_id = p.id AND e.tenant_id = p.tenant_id
		WHERE u.id = $1 AND u.tenant_id = $2 AND u.is_active = true`

	var user models.User
	var employeeID, employmentType, workSchedule, departmentName, positionName, status sql.NullString
	var joinDate sql.NullTime

	err := r.db.QueryRow(query, userID, tenantID).Scan(
		&user.ID, &user.TenantID, &user.Username, &user.Email, &user.Name, &user.Phone, &user.Role,
		&user.AvatarURL, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		&employeeID, &joinDate, &employmentType, &workSchedule,
		&departmentName, &positionName, &status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	// Set employee data jika ada
	if employeeID.Valid {
		user.EmployeeID = employeeID.String
	}
	if employmentType.Valid {
		user.EmploymentType = employmentType.String
	}
	if workSchedule.Valid {
		user.WorkSchedule = workSchedule.String
	}
	if departmentName.Valid {
		user.Department = departmentName.String
	}
	if positionName.Valid {
		user.Position = positionName.String
	}
	if status.Valid {
		user.Status = status.String
	}
	if joinDate.Valid {
		user.JoinDate = joinDate.Time.Format("2006-01-02")
	}

	return &user, nil
}

// GetUserByEmail mendapatkan user berdasarkan email
func (r *UserRepository) GetUserByEmail(tenantID uuid.UUID, email string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(`
		SELECT id, tenant_id, username, email, password, role, name, phone, avatar_url, is_active, created_at, updated_at
		FROM godplan.users 
		WHERE email = $1 AND tenant_id = $2 AND is_active = true
	`, email, tenantID).Scan(
		&user.ID,
		&user.TenantID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Name,
		&user.Phone,
		&user.AvatarURL,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByUsername mendapatkan user berdasarkan username
func (r *UserRepository) GetUserByUsername(tenantID uuid.UUID, username string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(`
		SELECT id, tenant_id, username, email, password, role, name, phone, avatar_url, is_active, created_at, updated_at
		FROM godplan.users 
		WHERE username = $1 AND tenant_id = $2 AND is_active = true
	`, username, tenantID).Scan(
		&user.ID,
		&user.TenantID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Name,
		&user.Phone,
		&user.AvatarURL,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// CreateUser membuat user baru
func (r *UserRepository) CreateUser(user *models.User) error {
	var id uuid.UUID
	err := r.db.QueryRow(`
		INSERT INTO godplan.users (
			tenant_id, username, email, password, role, name, phone, avatar_url, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`,
		user.TenantID,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.Name,
		user.Phone,
		user.AvatarURL,
		user.IsActive,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		if config.IsDevelopment() {
			log.Printf("❌ Failed to create user in repository: %v", err)
		}
		return err
	}

	user.ID = id
	return nil
}

// UpdateUser mengupdate data user
func (r *UserRepository) UpdateUser(user *models.User) error {
	_, err := r.db.Exec(`
		UPDATE godplan.users 
		SET username = $1, email = $2, role = $3, name = $4, phone = $5, 
		    avatar_url = $6, is_active = $7, updated_at = $8
		WHERE id = $9 AND tenant_id = $10
	`,
		user.Username,
		user.Email,
		user.Role,
		user.Name,
		user.Phone,
		user.AvatarURL,
		user.IsActive,
		time.Now(),
		user.ID,
		user.TenantID,
	)

	if err != nil {
		if config.IsDevelopment() {
			log.Printf("❌ Failed to update user in repository: %v", err)
		}
		return err
	}

	return nil
}

// DeleteUser menghapus user (soft delete)
func (r *UserRepository) DeleteUser(tenantID uuid.UUID, id uuid.UUID) error {
	_, err := r.db.Exec(`
		UPDATE godplan.users 
		SET is_active = false, updated_at = $1
		WHERE id = $2 AND tenant_id = $3
	`, time.Now(), id, tenantID)

	if err != nil {
		if config.IsDevelopment() {
			log.Printf("❌ Failed to delete user in repository: %v", err)
		}
		return err
	}

	return nil
}

// GetAllUsers mendapatkan semua user aktif
func (r *UserRepository) GetAllUsers(tenantID uuid.UUID) ([]models.User, error) {
	rows, err := r.db.Query(`
		SELECT id, tenant_id, username, email, role, name, phone, avatar_url, is_active, created_at, updated_at
		FROM godplan.users 
		WHERE is_active = true AND tenant_id = $1
		ORDER BY created_at DESC
	`, tenantID)
	if err != nil {
		if config.IsDevelopment() {
			log.Printf("❌ Failed to get all users in repository: %v", err)
		}
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.TenantID,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.Name,
			&user.Phone,
			&user.AvatarURL,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			if config.IsDevelopment() {
				log.Printf("❌ Failed to scan user in repository: %v", err)
			}
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// UpdatePassword mengupdate password user
func (r *UserRepository) UpdatePassword(tenantID uuid.UUID, userID uuid.UUID, hashedPassword string) error {
	_, err := r.db.Exec(`
		UPDATE godplan.users 
		SET password = $1, updated_at = $2
		WHERE id = $3 AND tenant_id = $4
	`, hashedPassword, time.Now(), userID, tenantID)

	if err != nil {
		if config.IsDevelopment() {
			log.Printf("❌ Failed to update password in repository: %v", err)
		}
		return err
	}

	return nil
}
