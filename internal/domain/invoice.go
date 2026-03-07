package domain

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// ErrImmutableInvoice est retourné lors d'une tentative de modification d'une facture immuable.
// LEGAL: Modification d'une facture payée = amende 75 000€ (Art. L441-9 Code de Commerce).
type ErrImmutableInvoice struct {
	InvoiceNumber string
	State         string
}

func (e *ErrImmutableInvoice) Error() string {
	return fmt.Sprintf("FACTURE IMMUABLE %s (état: %s): modification interdite par la loi (Art. L441-9)", e.InvoiceNumber, e.State)
}

// Invoice représente une facture.
type Invoice struct {
	ID       int
	Number   string  // Ex: "2026-0001"
	ClientID int
	Client   *Client // Relation chargée à la demande
	QuoteID  *int    // Lien vers le devis source, si conversion

	// Dates
	IssueDate    time.Time
	DueDate      time.Time
	DeliveryDate time.Time
	PaidDate     *time.Time // nil si non payée
	PaidLockedAt *time.Time // Timestamp du verrouillage immuable

	// État
	State InvoiceState

	// Lignes
	Lines []InvoiceLine

	// Montants (calculés depuis Lines)
	TotalHT   decimal.Decimal
	TotalTTC  decimal.Decimal
	VATAmount decimal.Decimal

	// TVA
	VATApplicable    bool
	VATExemptionText string // LEGAL: "TVA non applicable, article 293B du CGI"

	// Paiement
	PaymentDeadline      string          // "30 jours", "45 jours fin de mois"
	LatePenaltyRate      decimal.Decimal // Taux annuel (ex: 13.25 pour 13.25%)
	RecoveryFee          decimal.Decimal // LEGAL: 40€ forfaitaire obligatoire
	EarlyPaymentDiscount string          // Escompte optionnel

	// Paiement par virement
	PaymentRef string // Code virement 8 chars A-Z0-9, unique par facture, utilisé comme libellé de virement

	// Mentions 2027
	OperationCategory OperationCategory
	DeliveryAddress   string // Si différente de l'adresse client

	// Métadonnées
	Notes       string
	PDFPath     string // LEGAL: PDF jamais supprimé
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CanEdit retourne true si la facture peut être modifiée.
// LEGAL: Facture payée ou annulée = IMMUABLE (Art. L441-9 Code de Commerce).
func (inv *Invoice) CanEdit() bool {
	return inv.State != InvoiceStatePaid && inv.State != InvoiceStateCanceled
}

// CanDelete retourne true si la facture peut être supprimée.
// LEGAL: Seul un brouillon peut être supprimé.
func (inv *Invoice) CanDelete() bool {
	return inv.State == InvoiceStateDraft
}

// CanMarkAsPaid retourne true si la facture peut être marquée comme payée.
func (inv *Invoice) CanMarkAsPaid() bool {
	return inv.State == InvoiceStateIssued || inv.State == InvoiceStateSent
}

// CanCancel retourne true si la facture peut être annulée via un avoir.
func (inv *Invoice) CanCancel() bool {
	return inv.State == InvoiceStatePaid || inv.State == InvoiceStateIssued || inv.State == InvoiceStateSent
}

// IsOverdue retourne true si la facture est échue et non payée.
func (inv *Invoice) IsOverdue() bool {
	return inv.State != InvoiceStatePaid &&
		inv.State != InvoiceStateDraft &&
		time.Now().After(inv.DueDate)
}

// CalculateTotals recalcule les totaux depuis les lignes.
// LEGAL: Les montants doivent être exacts (pas de float).
func (inv *Invoice) CalculateTotals() {
	inv.TotalHT = decimal.Zero
	inv.VATAmount = decimal.Zero

	for i := range inv.Lines {
		inv.Lines[i].Calculate()
		inv.TotalHT = inv.TotalHT.Add(inv.Lines[i].TotalHT)
		if inv.VATApplicable {
			lineVAT := inv.Lines[i].TotalHT.Mul(inv.Lines[i].VATRate).Div(decimal.NewFromInt(100)).Round(2)
			inv.VATAmount = inv.VATAmount.Add(lineVAT)
		}
	}

	inv.TotalTTC = inv.TotalHT.Add(inv.VATAmount)
}

// Validate vérifie la conformité légale de la facture.
// LEGAL: Toutes les mentions obligatoires doivent être présentes.
func (inv *Invoice) Validate() []ValidationError {
	var errors []ValidationError

	// Numéro obligatoire
	if inv.Number == "" {
		errors = append(errors, ValidationError{
			Field:   "number",
			Message: "Numéro de facture obligatoire",
			Fine:    fineAmount(),
		})
	}

	// Date émission obligatoire
	if inv.IssueDate.IsZero() {
		errors = append(errors, ValidationError{
			Field:   "issue_date",
			Message: "Date d'émission obligatoire",
			Fine:    fineAmount(),
		})
	}

	// Date livraison obligatoire
	if inv.DeliveryDate.IsZero() {
		errors = append(errors, ValidationError{
			Field:   "delivery_date",
			Message: "Date de livraison/fin de prestation obligatoire",
			Fine:    fineAmount(),
		})
	}

	// Client obligatoire
	if inv.ClientID == 0 {
		errors = append(errors, ValidationError{
			Field:   "client_id",
			Message: "Client obligatoire",
			Fine:    fineAmount(),
		})
	}

	// LEGAL: SIREN client obligatoire si B2B (Décret n° 2022-1299, depuis 2024)
	if inv.Client != nil && inv.Client.IsCompany() && inv.Client.SIREN == "" {
		errors = append(errors, ValidationError{
			Field:   "client.siren",
			Message: "SIREN obligatoire pour clients professionnels (depuis 2024)",
			Fine:    fineAmount(),
		})
	}

	// Au moins une ligne obligatoire
	if len(inv.Lines) == 0 {
		errors = append(errors, ValidationError{
			Field:   "lines",
			Message: "Au moins une ligne de facturation obligatoire",
			Fine:    fineAmount(),
		})
	}

	// Validation de chaque ligne
	for i, line := range inv.Lines {
		if line.Description == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("lines[%d].description", i),
				Message: "Description obligatoire",
				Fine:    fineAmount(),
			})
		}
		if line.Quantity.LessThanOrEqual(decimal.Zero) {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("lines[%d].quantity", i),
				Message: "Quantité doit être > 0",
				Fine:    fineAmount(),
			})
		}
		// Les lignes de déduction (acompte, avoir partiel) peuvent avoir un prix négatif : autorisé.
	}

	// LEGAL: Mention TVA exacte obligatoire si franchise en base (Art. 293B CGI)
	if !inv.VATApplicable && inv.VATExemptionText != VATMentionExemption {
		errors = append(errors, ValidationError{
			Field:   "vat_exemption_text",
			Message: fmt.Sprintf("Mention TVA obligatoire exacte: %q", VATMentionExemption),
			Fine:    fineAmount(),
		})
	}

	// Délai de paiement obligatoire
	if inv.PaymentDeadline == "" {
		errors = append(errors, ValidationError{
			Field:   "payment_deadline",
			Message: "Délai de paiement obligatoire",
			Fine:    fineAmount(),
		})
	}

	// Taux pénalités de retard obligatoire
	if inv.LatePenaltyRate.LessThanOrEqual(decimal.Zero) {
		errors = append(errors, ValidationError{
			Field:   "late_penalty_rate",
			Message: "Taux de pénalités de retard obligatoire (Taux BCE + 10 points)",
			Fine:    fineAmount(),
		})
	}

	// LEGAL: Indemnité forfaitaire 40€ obligatoire (Art. L441-6 Code de Commerce)
	if !inv.RecoveryFee.Equal(decimal.NewFromInt(40)) {
		errors = append(errors, ValidationError{
			Field:   "recovery_fee",
			Message: "Indemnité forfaitaire de recouvrement doit être exactement 40€",
			Fine:    fineAmount(),
		})
	}

	return errors
}
