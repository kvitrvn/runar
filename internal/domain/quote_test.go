package domain_test

import (
	"testing"
	"time"

	"github.com/kvitrvn/runar/internal/domain"
	"github.com/shopspring/decimal"
)

func TestQuote_CanConvertToInvoice(t *testing.T) {
	tests := []struct {
		state domain.QuoteState
		want  bool
	}{
		{domain.QuoteStateDraft, false},
		{domain.QuoteStateSent, false},
		{domain.QuoteStateAccepted, true}, // Seul état permettant la conversion
		{domain.QuoteStateRefused, false},
		{domain.QuoteStateExpired, false},
	}
	for _, tt := range tests {
		q := &domain.Quote{State: tt.state}
		if got := q.CanConvertToInvoice(); got != tt.want {
			t.Errorf("CanConvertToInvoice() state=%s = %v, attendu %v", tt.state, got, tt.want)
		}
	}
}

func TestQuote_IsExpired(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	future := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name       string
		expiryDate time.Time
		want       bool
	}{
		{"date passée = expiré", past, true},
		{"date future = non expiré", future, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &domain.Quote{ExpiryDate: tt.expiryDate}
			if got := q.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, attendu %v", got, tt.want)
			}
		})
	}
}

func TestQuote_CalculateTotals(t *testing.T) {
	q := &domain.Quote{
		Lines: []domain.QuoteLine{
			{Quantity: decimal.NewFromFloat(2), UnitPriceHT: decimal.NewFromFloat(500), VATRate: decimal.NewFromFloat(20)},
			{Quantity: decimal.NewFromFloat(1), UnitPriceHT: decimal.NewFromFloat(200), VATRate: decimal.NewFromFloat(20)},
		},
	}
	q.CalculateTotals()

	expectedHT := decimal.NewFromFloat(1200)
	expectedVAT := decimal.NewFromFloat(240)
	expectedTTC := decimal.NewFromFloat(1440)

	if !q.TotalHT.Equal(expectedHT) {
		t.Errorf("TotalHT = %s, attendu %s", q.TotalHT, expectedHT)
	}
	if !q.VATAmount.Equal(expectedVAT) {
		t.Errorf("VATAmount = %s, attendu %s", q.VATAmount, expectedVAT)
	}
	if !q.TotalTTC.Equal(expectedTTC) {
		t.Errorf("TotalTTC = %s, attendu %s", q.TotalTTC, expectedTTC)
	}
}
