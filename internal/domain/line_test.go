package domain_test

import (
	"testing"

	"github.com/kvitrvn/runar/internal/domain"
	"github.com/shopspring/decimal"
)

func TestInvoiceLine_Calculate(t *testing.T) {
	line := &domain.InvoiceLine{
		Quantity:    decimal.NewFromFloat(3),
		UnitPriceHT: decimal.NewFromFloat(100),
		VATRate:     decimal.NewFromFloat(20),
	}
	line.Calculate()

	expectedHT := decimal.NewFromFloat(300)
	expectedTTC := decimal.NewFromFloat(360)

	if !line.TotalHT.Equal(expectedHT) {
		t.Errorf("TotalHT = %s, attendu %s", line.TotalHT, expectedHT)
	}
	if !line.TotalTTC.Equal(expectedTTC) {
		t.Errorf("TotalTTC = %s, attendu %s", line.TotalTTC, expectedTTC)
	}
}

func TestInvoiceLine_Calculate_SansTVA(t *testing.T) {
	line := &domain.InvoiceLine{
		Quantity:    decimal.NewFromFloat(2),
		UnitPriceHT: decimal.NewFromFloat(150),
		VATRate:     decimal.Zero,
	}
	line.Calculate()

	expected := decimal.NewFromFloat(300)
	if !line.TotalHT.Equal(expected) {
		t.Errorf("TotalHT = %s, attendu %s", line.TotalHT, expected)
	}
	if !line.TotalTTC.Equal(expected) {
		t.Errorf("TotalTTC = %s, attendu %s (sans TVA = HT)", line.TotalTTC, expected)
	}
}

func TestInvoiceLine_Calculate_Arrondi(t *testing.T) {
	// LEGAL: Arrondi au centime (2 décimales)
	line := &domain.InvoiceLine{
		Quantity:    decimal.NewFromFloat(1),
		UnitPriceHT: decimal.NewFromFloat(99.999),
		VATRate:     decimal.Zero,
	}
	line.Calculate()

	// 99.999 arrondi à 2 décimales = 100.00
	expected := decimal.NewFromFloat(100.00)
	if !line.TotalHT.Equal(expected) {
		t.Errorf("TotalHT arrondi = %s, attendu %s", line.TotalHT, expected)
	}
}

func TestQuoteLine_Calculate(t *testing.T) {
	line := &domain.QuoteLine{
		Quantity:    decimal.NewFromFloat(5),
		UnitPriceHT: decimal.NewFromFloat(200),
		VATRate:     decimal.NewFromFloat(20),
	}
	line.Calculate()

	expectedHT := decimal.NewFromFloat(1000)
	expectedTTC := decimal.NewFromFloat(1200)

	if !line.TotalHT.Equal(expectedHT) {
		t.Errorf("QuoteLine TotalHT = %s, attendu %s", line.TotalHT, expectedHT)
	}
	if !line.TotalTTC.Equal(expectedTTC) {
		t.Errorf("QuoteLine TotalTTC = %s, attendu %s", line.TotalTTC, expectedTTC)
	}
}
