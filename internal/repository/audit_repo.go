package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kvitrvn/runar/internal/domain"
)

// AuditLog représente une entrée du journal d'audit en SQL.
type AuditLog struct {
	ID         int       `db:"id"`
	EntityType string    `db:"entity_type"`
	EntityID   int       `db:"entity_id"`
	Action     string    `db:"action"`
	UserID     string    `db:"user_id"`
	OldValue   string    `db:"old_value"`
	NewValue   string    `db:"new_value"`
	IPAddress  string    `db:"ip_address"`
	Timestamp  time.Time `db:"timestamp"`
}

// AuditRepository définit les opérations sur l'audit.
// LEGAL: Traçabilité obligatoire (Art. L47 A Livre des procédures fiscales).
type AuditRepository interface {
	Log(entry domain.AuditLog) error
	GetByEntity(entityType string, entityID int) ([]domain.AuditLog, error)
}

type auditRepository struct {
	db *sqlx.DB
}

// NewAuditRepository crée un repository audit.
func NewAuditRepository(db *sqlx.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Log(entry domain.AuditLog) error {
	query := `
		INSERT INTO audit_log (entity_type, entity_id, action, user_id, old_value, new_value, ip_address, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	userID := entry.UserID
	if userID == "" {
		userID = "owner"
	}
	ts := entry.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}
	_, err := r.db.Exec(query,
		entry.EntityType, entry.EntityID, entry.Action,
		userID, entry.OldValue, entry.NewValue, entry.IPAddress, ts,
	)
	return err
}

func (r *auditRepository) GetByEntity(entityType string, entityID int) ([]domain.AuditLog, error) {
	var rows []AuditLog
	err := r.db.Select(&rows, `
		SELECT * FROM audit_log
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY timestamp DESC
	`, entityType, entityID)
	if err != nil {
		return nil, err
	}

	logs := make([]domain.AuditLog, len(rows))
	for i, row := range rows {
		logs[i] = domain.AuditLog{
			ID:         row.ID,
			EntityType: row.EntityType,
			EntityID:   row.EntityID,
			Action:     row.Action,
			UserID:     row.UserID,
			OldValue:   row.OldValue,
			NewValue:   row.NewValue,
			IPAddress:  row.IPAddress,
			Timestamp:  row.Timestamp,
		}
	}
	return logs, nil
}
