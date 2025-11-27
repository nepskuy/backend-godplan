package repository

import (
	"database/sql"

	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

// CRMRepository defines access methods for CRM projects (deals)
type CRMRepository interface {
	CreateProject(project *models.CRMProject) error
	GetProjectByID(id string) (*models.CRMProject, error)
	GetProjectsByManager(managerID string) ([]models.CRMProject, error)
	UpdateProject(project *models.CRMProject) error
	DeleteProject(id string) error
	ValidateProjectAccess(projectID, managerID string) (bool, error)
}

type crmRepositoryImpl struct {
	db *sql.DB
}

func NewCRMRepository(db *sql.DB) CRMRepository {
	return &crmRepositoryImpl{db: db}
}

func (r *crmRepositoryImpl) scanProject(row *sql.Row) (*models.CRMProject, error) {
	project := &models.CRMProject{}
	err := row.Scan(
		&project.ID,
		&project.Title,
		&project.Client,
		&project.Value,
		&project.Stage,
		&project.Urgency,
		&project.Deadline,
		&project.ContactPerson,
		&project.Description,
		&project.Category,
		&project.Status,
		&project.ManagerID,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrTaskNotFound // reuse generic not-found error
	}
	if err != nil {
		return nil, utils.ErrInternalServer
	}
	return project, nil
}

func (r *crmRepositoryImpl) CreateProject(project *models.CRMProject) error {
	query := `INSERT INTO godplan.projects 
		(title, client, value, stage, urgency, deadline, contact_person, description, category, status, manager_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query,
		project.Title,
		project.Client,
		project.Value,
		project.Stage,
		project.Urgency,
		project.Deadline,
		project.ContactPerson,
		project.Description,
		project.Category,
		project.Status,
		project.ManagerID,
	).Scan(&project.ID, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		return utils.ErrInternalServer
	}
	return nil
}

func (r *crmRepositoryImpl) GetProjectByID(id string) (*models.CRMProject, error) {
	query := `SELECT id, title, client, value, stage, urgency, deadline, contact_person, description, category, status, manager_id, created_at, updated_at
		FROM godplan.projects WHERE id = $1`
	row := r.db.QueryRow(query, id)
	return r.scanProject(row)
}

func (r *crmRepositoryImpl) GetProjectsByManager(managerID string) ([]models.CRMProject, error) {
	query := `SELECT id, title, client, value, stage, urgency, deadline, contact_person, description, category, status, manager_id, created_at, updated_at
		FROM godplan.projects
		WHERE manager_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, managerID)
	if err != nil {
		return nil, utils.ErrInternalServer
	}
	defer rows.Close()

	var projects []models.CRMProject
	for rows.Next() {
		var p models.CRMProject
		if err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Client,
			&p.Value,
			&p.Stage,
			&p.Urgency,
			&p.Deadline,
			&p.ContactPerson,
			&p.Description,
			&p.Category,
			&p.Status,
			&p.ManagerID,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, utils.ErrInternalServer
		}
		projects = append(projects, p)
	}

	return projects, nil
}

func (r *crmRepositoryImpl) UpdateProject(project *models.CRMProject) error {
	query := `UPDATE godplan.projects
		SET title = $1, client = $2, value = $3, stage = $4, urgency = $5,
		    deadline = $6, contact_person = $7, description = $8, category = $9,
		    status = $10, manager_id = $11, updated_at = CURRENT_TIMESTAMP
		WHERE id = $12`

	_, err := r.db.Exec(query,
		project.Title,
		project.Client,
		project.Value,
		project.Stage,
		project.Urgency,
		project.Deadline,
		project.ContactPerson,
		project.Description,
		project.Category,
		project.Status,
		project.ManagerID,
		project.ID,
	)
	if err != nil {
		return utils.ErrInternalServer
	}
	return nil
}

func (r *crmRepositoryImpl) DeleteProject(id string) error {
	query := `DELETE FROM godplan.projects WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return utils.ErrInternalServer
	}
	return nil
}

// ValidateProjectAccess ensures the given manager owns the project
func (r *crmRepositoryImpl) ValidateProjectAccess(projectID, managerID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM godplan.projects WHERE id = $1 AND manager_id = $2`
	if err := r.db.QueryRow(query, projectID, managerID).Scan(&count); err != nil {
		return false, utils.ErrInternalServer
	}
	return count > 0, nil
}
