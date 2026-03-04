package domain_test

import (
	"testing"

	"github.com/kvitrvn/runar/internal/domain"
)

func TestNumberingConfig_FormatInvoiceNumber(t *testing.T) {
	cfg := domain.DefaultNumberingConfig()
	tests := []struct {
		year int
		seq  int
		want string
	}{
		// LEGAL: Format attendu "ANNÉE-SEQUENCE" (Art. 242 nonies A CGI)
		{2026, 1, "2026-0001"},
		{2026, 99, "2026-0099"},
		{2026, 1000, "2026-1000"},
		{2027, 1, "2027-0001"},
	}
	for _, tt := range tests {
		got := cfg.FormatInvoiceNumber(tt.year, tt.seq)
		if got != tt.want {
			t.Errorf("FormatInvoiceNumber(%d, %d) = %q, attendu %q", tt.year, tt.seq, got, tt.want)
		}
	}
}

func TestNumberingConfig_FormatQuoteNumber(t *testing.T) {
	cfg := domain.DefaultNumberingConfig()
	got := cfg.FormatQuoteNumber(2026, 1)
	expected := "DEV-2026-0001"
	if got != expected {
		t.Errorf("FormatQuoteNumber(2026, 1) = %q, attendu %q", got, expected)
	}
}

func TestNumberingConfig_FormatCreditNoteNumber(t *testing.T) {
	// LEGAL: Séquence avoirs séparée des factures (pour clarté)
	cfg := domain.DefaultNumberingConfig()
	got := cfg.FormatCreditNoteNumber(2026, 1)
	expected := "A-2026-0001"
	if got != expected {
		t.Errorf("FormatCreditNoteNumber(2026, 1) = %q, attendu %q", got, expected)
	}
}

func TestNumberingConfig_InvoiceEtAvoirDistincts(t *testing.T) {
	// LEGAL: Les numéros de factures et d'avoirs ne peuvent pas se confondre
	cfg := domain.DefaultNumberingConfig()
	inv := cfg.FormatInvoiceNumber(2026, 1)
	cn := cfg.FormatCreditNoteNumber(2026, 1)
	if inv == cn {
		t.Errorf("Numéro facture et avoir ne devraient pas être identiques: %q", inv)
	}
}

func TestNumberingConfig_Continuity(t *testing.T) {
	// LEGAL: Séquence continue sans trou (Art. 242 nonies A CGI)
	cfg := domain.DefaultNumberingConfig()
	prev := cfg.FormatInvoiceNumber(2026, 1)
	for seq := 2; seq <= 10; seq++ {
		curr := cfg.FormatInvoiceNumber(2026, seq)
		if curr <= prev {
			t.Errorf("Numérotation non croissante: %q puis %q", prev, curr)
		}
		prev = curr
	}
}
