package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// Quote représente un devis.
type Quote struct {
	ID         int
	Number     string // Ex: "DEV-2026-0001"
	ClientID   int
	Client     *Client
	IssueDate  time.Time
	ExpiryDate time.Time
	State      QuoteState
	Lines      []QuoteLine
	TotalHT    decimal.Decimal
	TotalTTC   decimal.Decimal
	VATAmount  decimal.Decimal
	Notes      string
	PDFPath    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// CanConvertToInvoice retourne true si le devis peut être converti en facture.
func (q *Quote) CanConvertToInvoice() bool {
	return q.State == QuoteStateAccepted
}

// IsExpired retourne true si le devis est expiré.
func (q *Quote) IsExpired() bool {
	return time.Now().After(q.ExpiryDate)
}

// CalculateTotals recalcule les totaux depuis les lignes.
func (q *Quote) CalculateTotals() {
	q.TotalHT = decimal.Zero
	q.VATAmount = decimal.Zero

	for i := range q.Lines {
		q.Lines[i].Calculate()
		q.TotalHT = q.TotalHT.Add(q.Lines[i].TotalHT)
		lineVAT := q.Lines[i].TotalHT.Mul(q.Lines[i].VATRate).Div(decimal.NewFromInt(100)).Round(2)
		q.VATAmount = q.VATAmount.Add(lineVAT)
	}

	q.TotalTTC = q.TotalHT.Add(q.VATAmount)
}
