package domain

import "github.com/shopspring/decimal"

// InvoiceLine représente une ligne de facture.
type InvoiceLine struct {
	ID          int
	InvoiceID   int
	LineOrder   int
	Description string
	Quantity    decimal.Decimal
	UnitPriceHT decimal.Decimal
	VATRate     decimal.Decimal // En pourcentage (ex: 20.0 pour 20%)
	TotalHT     decimal.Decimal
	TotalTTC    decimal.Decimal
}

// Calculate recalcule les totaux de la ligne.
func (l *InvoiceLine) Calculate() {
	l.TotalHT = l.UnitPriceHT.Mul(l.Quantity).Round(2)
	vatAmount := l.TotalHT.Mul(l.VATRate).Div(decimal.NewFromInt(100)).Round(2)
	l.TotalTTC = l.TotalHT.Add(vatAmount)
}

// QuoteLine représente une ligne de devis.
type QuoteLine struct {
	ID          int
	QuoteID     int
	LineOrder   int
	Description string
	Quantity    decimal.Decimal
	UnitPriceHT decimal.Decimal
	VATRate     decimal.Decimal
	TotalHT     decimal.Decimal
	TotalTTC    decimal.Decimal
}

// Calculate recalcule les totaux de la ligne.
func (l *QuoteLine) Calculate() {
	l.TotalHT = l.UnitPriceHT.Mul(l.Quantity).Round(2)
	vatAmount := l.TotalHT.Mul(l.VATRate).Div(decimal.NewFromInt(100)).Round(2)
	l.TotalTTC = l.TotalHT.Add(vatAmount)
}

// CreditNoteLine représente une ligne d'avoir.
type CreditNoteLine struct {
	ID           int
	CreditNoteID int
	LineOrder    int
	Description  string
	Quantity     decimal.Decimal
	UnitPriceHT  decimal.Decimal
	VATRate      decimal.Decimal
	TotalHT      decimal.Decimal // Négatif
	TotalTTC     decimal.Decimal // Négatif
}
