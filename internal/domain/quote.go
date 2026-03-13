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
	// Acompte
	DepositRate       decimal.Decimal // pourcentage, 0 = pas d'acompte
	DepositPaid       bool
	DepositPaidAt     *time.Time
	DepositPaymentRef string // Libellé virement unique pour l'acompte
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// RequiresDeposit retourne true si le devis a un acompte configuré (taux > 0).
func (q *Quote) RequiresDeposit() bool {
	return q.DepositRate.IsPositive()
}

// DepositAmount retourne le montant de l'acompte (TotalHT * DepositRate / 100), arrondi à 2 décimales.
func (q *Quote) DepositAmount() decimal.Decimal {
	if !q.RequiresDeposit() {
		return decimal.Zero
	}
	return q.TotalHT.Mul(q.DepositRate).Div(decimal.NewFromInt(100)).Round(2)
}

// CanConvertToInvoice retourne true si le devis peut être converti en facture.
// LEGAL: Un devis avec acompte ne peut être converti que si l'acompte est payé.
func (q *Quote) CanConvertToInvoice() bool {
	if q.State != QuoteStateAccepted {
		return false
	}
	if q.RequiresDeposit() && !q.DepositPaid {
		return false
	}
	return true
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
