package service

import (
	"github.com/google/uuid"
	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/repository"
)

// CRMService defines business logic for CRM projects
type CRMService interface {
	CreateProject(project *models.CRMProject) error
	GetProjectByID(tenantID uuid.UUID, id uuid.UUID) (*models.CRMProject, error)
	GetProjectsByManager(tenantID uuid.UUID, managerID uuid.UUID) ([]models.CRMProject, error)
	UpdateProject(project *models.CRMProject) error
	DeleteProject(tenantID uuid.UUID, id uuid.UUID) error
	ValidateProjectAccess(tenantID uuid.UUID, projectID, managerID uuid.UUID) (bool, error)
}

type crmServiceImpl struct {
	crmRepo repository.CRMRepository
}

func NewCRMService(crmRepo repository.CRMRepository) CRMService {
	return &crmServiceImpl{crmRepo: crmRepo}
}

func (s *crmServiceImpl) CreateProject(project *models.CRMProject) error {
	// Set sensible defaults
	if project.Stage == "" {
		project.Stage = "new"
	}
	if project.Urgency == "" {
		project.Urgency = "medium"
	}
	if project.Status == "" {
		project.Status = "active"
	}
	return s.crmRepo.CreateProject(project)
}

func (s *crmServiceImpl) GetProjectByID(tenantID uuid.UUID, id uuid.UUID) (*models.CRMProject, error) {
	return s.crmRepo.GetProjectByID(tenantID, id)
}

func (s *crmServiceImpl) GetProjectsByManager(tenantID uuid.UUID, managerID uuid.UUID) ([]models.CRMProject, error) {
	return s.crmRepo.GetProjectsByManager(tenantID, managerID)
}

func (s *crmServiceImpl) UpdateProject(project *models.CRMProject) error {
	return s.crmRepo.UpdateProject(project)
}

func (s *crmServiceImpl) DeleteProject(tenantID uuid.UUID, id uuid.UUID) error {
	return s.crmRepo.DeleteProject(tenantID, id)
}

func (s *crmServiceImpl) ValidateProjectAccess(tenantID uuid.UUID, projectID, managerID uuid.UUID) (bool, error) {
	return s.crmRepo.ValidateProjectAccess(tenantID, projectID, managerID)
}
