package domain

import "fmt"

// NumberingConfig définit le format de numérotation.
type NumberingConfig struct {
	InvoicePrefix   string // "" → "2026-0001"
	QuotePrefix     string // "DEV" → "DEV-2026-0001"
	CreditNotePrefix string // "A" → "A-2026-0001"
	SequenceWidth   int    // 4 → padding 4 chiffres
}

// DefaultNumberingConfig retourne la configuration par défaut.
func DefaultNumberingConfig() NumberingConfig {
	return NumberingConfig{
		InvoicePrefix:    "",
		QuotePrefix:      "DEV",
		CreditNotePrefix: "A",
		SequenceWidth:    4,
	}
}

// FormatInvoiceNumber génère un numéro de facture.
// LEGAL: Format continu annuel sans trou (Art. 242 nonies A CGI).
func (cfg NumberingConfig) FormatInvoiceNumber(year, seq int) string {
	if cfg.InvoicePrefix == "" {
		return fmt.Sprintf("%d-%0*d", year, cfg.SequenceWidth, seq)
	}
	return fmt.Sprintf("%s-%d-%0*d", cfg.InvoicePrefix, year, cfg.SequenceWidth, seq)
}

// FormatQuoteNumber génère un numéro de devis.
func (cfg NumberingConfig) FormatQuoteNumber(year, seq int) string {
	if cfg.QuotePrefix == "" {
		return fmt.Sprintf("%d-%0*d", year, cfg.SequenceWidth, seq)
	}
	return fmt.Sprintf("%s-%d-%0*d", cfg.QuotePrefix, year, cfg.SequenceWidth, seq)
}

// FormatCreditNoteNumber génère un numéro d'avoir.
// LEGAL: Séquence séparée des factures pour clarté.
func (cfg NumberingConfig) FormatCreditNoteNumber(year, seq int) string {
	if cfg.CreditNotePrefix == "" {
		return fmt.Sprintf("%d-%0*d", year, cfg.SequenceWidth, seq)
	}
	return fmt.Sprintf("%s-%d-%0*d", cfg.CreditNotePrefix, year, cfg.SequenceWidth, seq)
}
