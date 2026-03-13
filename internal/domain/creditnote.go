package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// CreditNote représente un avoir (note de crédit).
// LEGAL: Référence à la facture d'origine obligatoire (Art. 272 CGI).
// LEGAL: Conservation 10 ans, comme les factures (Art. L123-22 Code de Commerce).
type CreditNote struct {
	ID               int
	Number           string // Ex: "A-2026-0001"
	InvoiceID        int
	InvoiceReference string // Numéro et date de la facture d'origine (obligatoire)
	IssueDate        time.Time
	Reason           string  // Motif de l'avoir (obligatoire)
	Client           *Client // Relation chargée à la demande (depuis la facture d'origine)
	Lines            []CreditNoteLine
	TotalHT          decimal.Decimal // Négatif
	TotalTTC         decimal.Decimal // Négatif
	VATAmount        decimal.Decimal
	PDFPath          string
	CreatedAt        time.Time
}

// Validate vérifie la conformité légale de l'avoir.
func (cn *CreditNote) Validate() []ValidationError {
	var errors []ValidationError

	if cn.Number == "" {
		errors = append(errors, ValidationError{
			Field:   "number",
			Message: "Numéro d'avoir obligatoire",
			Fine:    fineAmount(),
		})
	}

	// LEGAL: Référence facture d'origine obligatoire
	if cn.InvoiceReference == "" {
		errors = append(errors, ValidationError{
			Field:   "invoice_reference",
			Message: "Référence à la facture d'origine obligatoire",
			Fine:    fineAmount(),
		})
	}

	if cn.IssueDate.IsZero() {
		errors = append(errors, ValidationError{
			Field:   "issue_date",
			Message: "Date d'émission obligatoire",
			Fine:    fineAmount(),
		})
	}

	if cn.Reason == "" {
		errors = append(errors, ValidationError{
			Field:   "reason",
			Message: "Motif de l'avoir obligatoire",
			Fine:    fineAmount(),
		})
	}

	return errors
}
