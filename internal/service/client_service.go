package service

import (
	"encoding/json"

	"github.com/kvitrvn/runar/internal/domain"
	"github.com/kvitrvn/runar/internal/repository"
)

// ClientService gère les opérations sur les clients.
type ClientService struct {
	repo  repository.ClientRepository
	audit *AuditService
}

// NewClientService crée un service client.
func NewClientService(repo repository.ClientRepository, audit *AuditService) *ClientService {
	return &ClientService{repo: repo, audit: audit}
}

// Create crée un nouveau client après validation.
func (s *ClientService) Create(c *domain.Client) error {
	if errs := c.Validate(); len(errs) > 0 {
		return &domain.ValidationErrorList{Errors: errs}
	}

	if err := s.repo.Create(c); err != nil {
		return err
	}

	newVal, _ := json.Marshal(c)
	s.audit.Log("client", c.ID, domain.AuditActionCreated, "", string(newVal))
	return nil
}

// Update met à jour un client existant.
func (s *ClientService) Update(id int, c *domain.Client) error {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if errs := c.Validate(); len(errs) > 0 {
		return &domain.ValidationErrorList{Errors: errs}
	}

	oldVal, _ := json.Marshal(existing)
	if err := s.repo.Update(id, c); err != nil {
		return err
	}

	newVal, _ := json.Marshal(c)
	s.audit.Log("client", id, domain.AuditActionUpdated, string(oldVal), string(newVal))
	return nil
}

// GetByID retourne un client par son ID.
func (s *ClientService) GetByID(id int) (*domain.Client, error) {
	return s.repo.GetByID(id)
}

// List retourne la liste des clients avec filtre optionnel.
func (s *ClientService) List(search string) ([]domain.Client, error) {
	return s.repo.List(search)
}

// Delete supprime un client.
func (s *ClientService) Delete(id int) error {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	oldVal, _ := json.Marshal(existing)
	s.audit.Log("client", id, domain.AuditActionDeleted, string(oldVal), "")
	return nil
}
