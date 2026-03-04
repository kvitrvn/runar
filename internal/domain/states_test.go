package domain_test

import (
	"testing"

	"github.com/kvitrvn/runar/internal/domain"
)

func TestVATMentionExemption_TexteExact(t *testing.T) {
	// LEGAL: Le texte exact est obligatoire (Art. 293B CGI)
	// Toute variation = infraction
	expected := "TVA non applicable, article 293B du CGI"
	if domain.VATMentionExemption != expected {
		t.Errorf("VATMentionExemption = %q, attendu %q", domain.VATMentionExemption, expected)
	}
}

func TestDefaultRecoveryFee_40Euros(t *testing.T) {
	// LEGAL: Indemnité forfaitaire = 40€ fixe (Art. L441-6 Code de Commerce)
	if domain.DefaultRecoveryFee != "40" {
		t.Errorf("DefaultRecoveryFee = %q, attendu %q", domain.DefaultRecoveryFee, "40")
	}
}

func TestInvoiceState_Valeurs(t *testing.T) {
	// Vérifier que les constantes correspondent aux valeurs SQL attendues
	states := map[domain.InvoiceState]string{
		domain.InvoiceStateDraft:    "draft",
		domain.InvoiceStateIssued:   "issued",
		domain.InvoiceStateSent:     "sent",
		domain.InvoiceStatePaid:     "paid",
		domain.InvoiceStateCanceled: "canceled",
	}
	for state, expected := range states {
		if string(state) != expected {
			t.Errorf("InvoiceState %s != %s", state, expected)
		}
	}
}

func TestInvoice_CanCancel(t *testing.T) {
	tests := []struct {
		state domain.InvoiceState
		want  bool
	}{
		{domain.InvoiceStateDraft, false},
		{domain.InvoiceStateIssued, true},
		{domain.InvoiceStateSent, true},
		{domain.InvoiceStatePaid, true}, // Avoir possible sur facture payée
		{domain.InvoiceStateCanceled, false},
	}
	for _, tt := range tests {
		inv := &domain.Invoice{State: tt.state}
		if got := inv.CanCancel(); got != tt.want {
			t.Errorf("CanCancel() state=%s = %v, attendu %v", tt.state, got, tt.want)
		}
	}
}
