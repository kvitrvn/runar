package styles

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/kvitrvn/runar/internal/domain"
)

// RenderInvoiceState retourne le badge coloré de l'état d'une facture.
func RenderInvoiceState(state domain.InvoiceState, isOverdue bool) string {
	if isOverdue {
		return StyleOverdue.Render("⚠ ÉCHUE")
	}
	switch state {
	case domain.InvoiceStateDraft:
		return StyleDraft.Render("BROUILLON")
	case domain.InvoiceStateIssued:
		return StyleIssued.Render("ÉMISE")
	case domain.InvoiceStateSent:
		return StyleSent.Render("ENVOYÉE")
	case domain.InvoiceStatePaid:
		return StylePaid.Render("✓ PAYÉE")
	case domain.InvoiceStateCanceled:
		return StyleCanceled.Render("ANNULÉE")
	default:
		return StyleMuted.Render("INCONNU")
	}
}

// RenderQuoteState retourne le badge coloré de l'état d'un devis.
func RenderQuoteState(state domain.QuoteState) string {
	switch state {
	case domain.QuoteStateDraft:
		return StyleDraft.Render("BROUILLON")
	case domain.QuoteStateSent:
		return StyleIssued.Render("ENVOYÉ")
	case domain.QuoteStateAccepted:
		return StylePaid.Render("✓ ACCEPTÉ")
	case domain.QuoteStateRefused:
		return StyleCanceled.Render("REFUSÉ")
	case domain.QuoteStateExpired:
		return StyleOverdue.Render("EXPIRÉ")
	default:
		return StyleMuted.Render("INCONNU")
	}
}

// BadgePaid retourne le badge d'une facture payée.
func BadgePaid() string {
	return lipgloss.NewStyle().Foreground(ColorSuccess).Render("✓ PAYÉE")
}

// BadgeOverdue retourne le badge d'une facture échue.
func BadgeOverdue() string {
	return lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render("⚠ ÉCHUE")
}

// BadgeLocked retourne le badge d'immuabilité (facture payée/annulée).
func BadgeLocked() string {
	return lipgloss.NewStyle().
		Foreground(ColorForeground).
		Background(ColorDanger).
		Bold(true).
		Padding(0, 1).
		Render("🔒 IMMUABLE")
}

// BadgeDraft retourne le badge brouillon.
func BadgeDraft() string {
	return lipgloss.NewStyle().Foreground(ColorMuted).Render("BROUILLON")
}

// RenderImmutableError affiche un message d'erreur pour tentative de modification d'une facture immuable.
func RenderImmutableError(invoiceNumber string) string {
	return lipgloss.NewStyle().
		Foreground(ColorForeground).
		Background(ColorDanger).
		Padding(1, 2).
		Bold(true).
		Render(fmt.Sprintf(
			"❌ MODIFICATION INTERDITE\n\n"+
				"La facture %s est PAYÉE et ne peut être modifiée.\n"+
				"Action autorisée : Créer un avoir (touche 'c')",
			invoiceNumber,
		))
}

// RenderValidationErrors affiche les erreurs de validation avec amendes potentielles.
func RenderValidationErrors(errors []domain.ValidationError) string {
	if len(errors) == 0 {
		return ""
	}
	content := "⚠  ERREURS DE VALIDATION\n\n"
	for _, err := range errors {
		content += fmt.Sprintf("• %s: %s\n", err.Field, err.Message)
		if err.Fine.IsPositive() {
			content += fmt.Sprintf("  Amende potentielle: %s €\n", err.Fine.StringFixed(2))
		}
	}
	return lipgloss.NewStyle().
		Foreground(ColorWarning).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorWarning).
		Padding(1, 2).
		Render(content)
}
