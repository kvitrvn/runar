package domain_test

import (
	"testing"
	"time"

	"github.com/kvitrvn/runar/internal/domain"
	"github.com/shopspring/decimal"
)

// ─── CanEdit ────────────────────────────────────────────────────────────────

func TestInvoice_CanEdit(t *testing.T) {
	// LEGAL: Art. L441-9 - Facture payée ou annulée = immuable
	tests := []struct {
		state domain.InvoiceState
		want  bool
	}{
		{domain.InvoiceStateDraft, true},
		{domain.InvoiceStateIssued, true},
		{domain.InvoiceStateSent, true},
		{domain.InvoiceStatePaid, false},     // CRITIQUE: jamais éditable
		{domain.InvoiceStateCanceled, false}, // CRITIQUE: jamais éditable
	}
	for _, tt := range tests {
		inv := &domain.Invoice{State: tt.state}
		if got := inv.CanEdit(); got != tt.want {
			t.Errorf("CanEdit() avec state=%s = %v, attendu %v", tt.state, got, tt.want)
		}
	}
}

// ─── CanDelete ──────────────────────────────────────────────────────────────

func TestInvoice_CanDelete(t *testing.T) {
	tests := []struct {
		state domain.InvoiceState
		want  bool
	}{
		{domain.InvoiceStateDraft, true},
		{domain.InvoiceStateIssued, false},
		{domain.InvoiceStateSent, false},
		{domain.InvoiceStatePaid, false},
		{domain.InvoiceStateCanceled, false},
	}
	for _, tt := range tests {
		inv := &domain.Invoice{State: tt.state}
		if got := inv.CanDelete(); got != tt.want {
			t.Errorf("CanDelete() avec state=%s = %v, attendu %v", tt.state, got, tt.want)
		}
	}
}

// ─── CanMarkAsPaid ──────────────────────────────────────────────────────────

func TestInvoice_CanMarkAsPaid(t *testing.T) {
	tests := []struct {
		state domain.InvoiceState
		want  bool
	}{
		{domain.InvoiceStateDraft, false},
		{domain.InvoiceStateIssued, true},
		{domain.InvoiceStateSent, true},
		{domain.InvoiceStatePaid, false},
		{domain.InvoiceStateCanceled, false},
	}
	for _, tt := range tests {
		inv := &domain.Invoice{State: tt.state}
		if got := inv.CanMarkAsPaid(); got != tt.want {
			t.Errorf("CanMarkAsPaid() avec state=%s = %v, attendu %v", tt.state, got, tt.want)
		}
	}
}

// ─── IsOverdue ───────────────────────────────────────────────────────────────

func TestInvoice_IsOverdue(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	future := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name    string
		state   domain.InvoiceState
		dueDate time.Time
		want    bool
	}{
		{"issued échue", domain.InvoiceStateIssued, past, true},
		{"sent échue", domain.InvoiceStateSent, past, true},
		{"issued non échue", domain.InvoiceStateIssued, future, false},
		{"paid (jamais échue)", domain.InvoiceStatePaid, past, false},
		{"draft (jamais échue)", domain.InvoiceStateDraft, past, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &domain.Invoice{State: tt.state, DueDate: tt.dueDate}
			if got := inv.IsOverdue(); got != tt.want {
				t.Errorf("IsOverdue() = %v, attendu %v", got, tt.want)
			}
		})
	}
}

// ─── CalculateTotals ─────────────────────────────────────────────────────────

func TestInvoice_CalculateTotals_SansTVA(t *testing.T) {
	// LEGAL: Précision décimale obligatoire (jamais float)
	inv := &domain.Invoice{
		VATApplicable: false,
		Lines: []domain.InvoiceLine{
			{Quantity: decimal.NewFromFloat(2), UnitPriceHT: decimal.NewFromFloat(100)},
			{Quantity: decimal.NewFromFloat(1), UnitPriceHT: decimal.NewFromFloat(50.5)},
		},
	}
	inv.CalculateTotals()

	expectedHT := decimal.NewFromFloat(250.5)
	if !inv.TotalHT.Equal(expectedHT) {
		t.Errorf("TotalHT = %s, attendu %s", inv.TotalHT, expectedHT)
	}
	if !inv.TotalTTC.Equal(expectedHT) {
		t.Errorf("TotalTTC = %s, attendu %s (sans TVA = HT)", inv.TotalTTC, expectedHT)
	}
	if !inv.VATAmount.Equal(decimal.Zero) {
		t.Errorf("VATAmount = %s, attendu 0 (franchise)", inv.VATAmount)
	}
}

func TestInvoice_CalculateTotals_AvecTVA(t *testing.T) {
	inv := &domain.Invoice{
		VATApplicable: true,
		Lines: []domain.InvoiceLine{
			{
				Quantity:    decimal.NewFromFloat(1),
				UnitPriceHT: decimal.NewFromFloat(1000),
				VATRate:     decimal.NewFromFloat(20),
			},
		},
	}
	inv.CalculateTotals()

	expectedHT := decimal.NewFromFloat(1000)
	expectedVAT := decimal.NewFromFloat(200)
	expectedTTC := decimal.NewFromFloat(1200)

	if !inv.TotalHT.Equal(expectedHT) {
		t.Errorf("TotalHT = %s, attendu %s", inv.TotalHT, expectedHT)
	}
	if !inv.VATAmount.Equal(expectedVAT) {
		t.Errorf("VATAmount = %s, attendu %s", inv.VATAmount, expectedVAT)
	}
	if !inv.TotalTTC.Equal(expectedTTC) {
		t.Errorf("TotalTTC = %s, attendu %s", inv.TotalTTC, expectedTTC)
	}
}

func TestInvoice_CalculateTotals_PrecisionDecimale(t *testing.T) {
	// LEGAL: 0.1 + 0.2 doit être exact avec decimal (pas de float)
	inv := &domain.Invoice{
		VATApplicable: false,
		Lines: []domain.InvoiceLine{
			{Quantity: decimal.NewFromFloat(1), UnitPriceHT: decimal.NewFromFloat(0.1)},
			{Quantity: decimal.NewFromFloat(1), UnitPriceHT: decimal.NewFromFloat(0.2)},
		},
	}
	inv.CalculateTotals()

	expected := decimal.NewFromFloat(0.3)
	if !inv.TotalHT.Equal(expected) {
		t.Errorf("Précision décimale: TotalHT = %s, attendu %s", inv.TotalHT, expected)
	}
}

// ─── Validate ────────────────────────────────────────────────────────────────

func validInvoice() *domain.Invoice {
	return &domain.Invoice{
		Number:       "2026-0001",
		ClientID:     1,
		IssueDate:    time.Now(),
		DueDate:      time.Now().Add(30 * 24 * time.Hour),
		DeliveryDate: time.Now(),
		Lines: []domain.InvoiceLine{
			{
				Description: "Prestation de service",
				Quantity:    decimal.NewFromFloat(1),
				UnitPriceHT: decimal.NewFromFloat(1000),
			},
		},
		VATApplicable:    false,
		VATExemptionText: domain.VATMentionExemption,
		PaymentDeadline:  "30 jours",
		LatePenaltyRate:  decimal.NewFromFloat(13.25),
		RecoveryFee:      decimal.NewFromInt(40),
	}
}

func TestInvoice_Validate_Valide(t *testing.T) {
	inv := validInvoice()
	errs := inv.Validate()
	if len(errs) != 0 {
		t.Errorf("Facture valide: attendu 0 erreurs, got %d: %v", len(errs), errs)
	}
}

func TestInvoice_Validate_MentionTVAManquante(t *testing.T) {
	// LEGAL: Mention TVA exacte obligatoire (Art. 293B CGI) - amende 15€
	inv := validInvoice()
	inv.VATExemptionText = "TVA non applicable" // Incomplet = infraction
	errs := inv.Validate()
	if len(errs) == 0 {
		t.Fatal("Mention TVA incomplète doit être une erreur")
	}
	found := false
	for _, e := range errs {
		if e.Field == "vat_exemption_text" {
			found = true
			if !e.Fine.Equal(decimal.NewFromInt(15)) {
				t.Errorf("Amende TVA = %s, attendu 15€", e.Fine)
			}
		}
	}
	if !found {
		t.Error("Erreur vat_exemption_text non trouvée")
	}
}

func TestInvoice_Validate_SansNumero(t *testing.T) {
	// LEGAL: Numéro obligatoire (Art. 242 nonies A CGI) - amende 15€
	inv := validInvoice()
	inv.Number = ""
	errs := inv.Validate()
	if len(errs) == 0 {
		t.Fatal("Numéro manquant doit être une erreur")
	}
	found := false
	for _, e := range errs {
		if e.Field == "number" {
			found = true
		}
	}
	if !found {
		t.Error("Erreur number non trouvée")
	}
}

func TestInvoice_Validate_RecoveryFeeIncorrecte(t *testing.T) {
	// LEGAL: Indemnité forfaitaire doit être exactement 40€ (Art. L441-6)
	inv := validInvoice()
	inv.RecoveryFee = decimal.NewFromInt(35) // Incorrect
	errs := inv.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "recovery_fee" {
			found = true
		}
	}
	if !found {
		t.Error("Erreur recovery_fee non trouvée pour valeur != 40€")
	}
}

func TestInvoice_Validate_LigneDescriptionVide(t *testing.T) {
	// LEGAL: Description obligatoire sur chaque ligne (Art. L441-3) - amende 15€
	inv := validInvoice()
	inv.Lines[0].Description = ""
	errs := inv.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "lines[0].description" {
			found = true
		}
	}
	if !found {
		t.Error("Erreur description ligne vide non trouvée")
	}
}

func TestInvoice_Validate_ClientB2B_SansSIREN(t *testing.T) {
	// LEGAL: SIREN obligatoire pour clients professionnels depuis 2024 (Décret n° 2022-1299)
	inv := validInvoice()
	inv.Client = &domain.Client{
		Name:  "Acme SARL",
		SIRET: "73282932000074", // Entreprise identifiée par SIRET
		SIREN: "",               // Mais SIREN manquant
	}
	errs := inv.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "client.siren" {
			found = true
		}
	}
	if !found {
		t.Error("Erreur SIREN client B2B non trouvée")
	}
}

func TestInvoice_Validate_AmendeTotale(t *testing.T) {
	// LEGAL: Calcul amende totale plafonnée à 25% du montant TTC
	list := &domain.ValidationErrorList{
		Errors: []domain.ValidationError{
			{Field: "f1", Fine: decimal.NewFromInt(15)},
			{Field: "f2", Fine: decimal.NewFromInt(15)},
			{Field: "f3", Fine: decimal.NewFromInt(15)},
		},
	}
	ttc := decimal.NewFromFloat(100) // Plafond = 25€
	fine := list.TotalFine(ttc)
	// 3 * 15 = 45€ mais plafonné à 25€
	expected := decimal.NewFromFloat(25)
	if !fine.Equal(expected) {
		t.Errorf("Amende plafonnée = %s, attendu %s", fine, expected)
	}
}

// ─── ErrImmutableInvoice ────────────────────────────────────────────────────

func TestErrImmutableInvoice_Message(t *testing.T) {
	err := &domain.ErrImmutableInvoice{
		InvoiceNumber: "2026-0001",
		State:         "paid",
	}
	msg := err.Error()
	if msg == "" {
		t.Error("Message d'erreur immuable vide")
	}
	// Doit contenir le numéro pour traçabilité
	if len(msg) < 10 {
		t.Errorf("Message d'erreur trop court: %q", msg)
	}
}
