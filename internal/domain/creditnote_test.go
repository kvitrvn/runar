package domain_test

import (
	"testing"
	"time"

	"github.com/kvitrvn/runar/internal/domain"
)

func TestCreditNote_Validate_ReferenceObligatoire(t *testing.T) {
	// LEGAL: Référence à la facture d'origine obligatoire (Art. 272 CGI)
	cn := &domain.CreditNote{
		Number:           "A-2026-0001",
		InvoiceReference: "", // Manquant
		IssueDate:        time.Now(),
		Reason:           "Annulation prestation",
	}
	errs := cn.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "invoice_reference" {
			found = true
		}
	}
	if !found {
		t.Error("Erreur invoice_reference manquant non trouvée")
	}
}

func TestCreditNote_Validate_MotifObligatoire(t *testing.T) {
	cn := &domain.CreditNote{
		Number:           "A-2026-0001",
		InvoiceReference: "2026-0001",
		IssueDate:        time.Now(),
		Reason:           "", // Manquant
	}
	errs := cn.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "reason" {
			found = true
		}
	}
	if !found {
		t.Error("Erreur reason manquant non trouvée")
	}
}

func TestCreditNote_Validate_Valide(t *testing.T) {
	cn := &domain.CreditNote{
		Number:           "A-2026-0001",
		InvoiceID:        1,
		InvoiceReference: "2026-0001",
		IssueDate:        time.Now(),
		Reason:           "Erreur sur facture originale",
	}
	errs := cn.Validate()
	if len(errs) != 0 {
		t.Errorf("Avoir valide: attendu 0 erreurs, got %d: %v", len(errs), errs)
	}
}
