package domain

import "time"

// AuditLog représente une entrée du journal d'audit.
// LEGAL: Traçabilité obligatoire de toutes les actions (Art. L47 A LPF).
// LEGAL: Conservation 6 ans minimum (durée du droit de contrôle fiscal).
type AuditLog struct {
	ID         int
	EntityType string    // "invoice", "quote", "client", "credit_note"
	EntityID   int
	Action     string    // "created", "updated", "paid_and_locked", "pdf_generated", etc.
	UserID     string    // Toujours "owner" pour un auto-entrepreneur
	OldValue   string    // JSON de l'état avant modification
	NewValue   string    // JSON de l'état après modification
	IPAddress  string    // Optionnel
	Timestamp  time.Time
}

// Actions d'audit standard
const (
	AuditActionCreated      = "created"
	AuditActionUpdated      = "updated"
	AuditActionDeleted      = "deleted"
	AuditActionPaidLocked   = "paid_and_locked"
	AuditActionPDFGenerated = "pdf_generated"
	AuditActionCanceled     = "canceled"
	AuditActionDenied       = "modification_denied" // Tentative de modification interdite
	AuditActionExported     = "exported"
)
