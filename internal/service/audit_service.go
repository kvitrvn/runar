package service

import (
	"time"

	"github.com/kvitrvn/runar/internal/domain"
	"github.com/kvitrvn/runar/internal/repository"
)

// AuditService gère la traçabilité des actions.
// LEGAL: Toute action sur une entité doit être loggée (Art. L47 A LPF).
type AuditService struct {
	repo repository.AuditRepository
}

// NewAuditService crée un service d'audit.
func NewAuditService(repo repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

// Log enregistre une action dans le journal d'audit.
func (s *AuditService) Log(entityType string, entityID int, action, oldValue, newValue string) {
	entry := domain.AuditLog{
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		UserID:     "owner",
		OldValue:   oldValue,
		NewValue:   newValue,
		Timestamp:  time.Now(),
	}
	// On ignore l'erreur de log pour ne pas bloquer l'opération principale
	_ = s.repo.Log(entry)
}

// GetHistory retourne l'historique d'une entité.
func (s *AuditService) GetHistory(entityType string, entityID int) ([]domain.AuditLog, error) {
	return s.repo.GetByEntity(entityType, entityID)
}
