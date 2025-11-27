package service

import (
	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/repository"
)

// CRMService defines business logic for CRM projects
type CRMService interface {
	CreateProject(project *models.CRMProject) error
	GetProjectByID(id string) (*models.CRMProject, error)
	GetProjectsByManager(managerID string) ([]models.CRMProject, error)
	UpdateProject(project *models.CRMProject) error
	DeleteProject(id string) error
	ValidateProjectAccess(projectID, managerID string) (bool, error)
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

func (s *crmServiceImpl) GetProjectByID(id string) (*models.CRMProject, error) {
	return s.crmRepo.GetProjectByID(id)
}

func (s *crmServiceImpl) GetProjectsByManager(managerID string) ([]models.CRMProject, error) {
	return s.crmRepo.GetProjectsByManager(managerID)
}

func (s *crmServiceImpl) UpdateProject(project *models.CRMProject) error {
	return s.crmRepo.UpdateProject(project)
}

func (s *crmServiceImpl) DeleteProject(id string) error {
	return s.crmRepo.DeleteProject(id)
}

func (s *crmServiceImpl) ValidateProjectAccess(projectID, managerID string) (bool, error) {
	return s.crmRepo.ValidateProjectAccess(projectID, managerID)
}
