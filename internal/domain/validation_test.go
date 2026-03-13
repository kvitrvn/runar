package domain_test

import (
	"testing"

	"github.com/kvitrvn/runar/internal/domain"
)

// ─── SIRET ───────────────────────────────────────────────────────────────────

func TestValidateSIRET(t *testing.T) {
	tests := []struct {
		name  string
		siret string
		valid bool
	}{
		// SIRET réels valides (Luhn)
		{"valide sans espaces", "73282932000074", true},
		{"valide avec espaces", "732 829 320 00074", true},
		// Cas invalides
		{"trop court", "123456789012", false},
		{"trop long", "1234567890123456", false},
		{"caractères non numériques", "7328A932000074", false},
		{"Luhn invalide", "73282932000075", false}, // Dernier chiffre modifié
		{"vide", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.ValidateSIRET(tt.siret)
			if got != tt.valid {
				t.Errorf("ValidateSIRET(%q) = %v, attendu %v", tt.siret, got, tt.valid)
			}
		})
	}
}

// ─── SIREN ───────────────────────────────────────────────────────────────────

func TestValidateSIREN(t *testing.T) {
	tests := []struct {
		name  string
		siren string
		valid bool
	}{
		// Le SIREN est les 9 premiers chiffres du SIRET
		{"valide", "732829320", true},
		{"valide avec espaces", "732 829 320", true},
		{"trop court", "12345678", false},
		{"trop long", "1234567890", false},
		{"non numérique", "73282932A", false},
		{"Luhn invalide", "732829321", false},
		{"vide", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.ValidateSIREN(tt.siren)
			if got != tt.valid {
				t.Errorf("ValidateSIREN(%q) = %v, attendu %v", tt.siren, got, tt.valid)
			}
		})
	}
}

// ─── Email ───────────────────────────────────────────────────────────────────

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"user.name+tag@sub.domain.fr", true},
		{"invalid", false},
		{"@domain.com", false},
		{"user@", false},
		{"user@domain", false},
		{"", false},
	}
	for _, tt := range tests {
		got := domain.ValidateEmail(tt.email)
		if got != tt.valid {
			t.Errorf("ValidateEmail(%q) = %v, attendu %v", tt.email, got, tt.valid)
		}
	}
}

// ─── IBAN ───────────────────────────────────────────────────────────────────

func TestValidateIBAN(t *testing.T) {
	tests := []struct {
		name  string
		iban  string
		valid bool
	}{
		{"valide avec espaces", "FR76 3000 6000 0112 3456 7890 189", true},
		{"valide sans espaces", "FR7630006000011234567890189", true},
		{"minuscule", "fr7630006000011234567890189", true},
		{"clé invalide", "FR7630006000011234567890188", false},
		{"format invalide", "FR76-3000-6000-0112", false},
		{"vide", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.ValidateIBAN(tt.iban)
			if got != tt.valid {
				t.Errorf("ValidateIBAN(%q) = %v, attendu %v", tt.iban, got, tt.valid)
			}
		})
	}
}

// ─── BIC ────────────────────────────────────────────────────────────────────

func TestValidateBIC(t *testing.T) {
	tests := []struct {
		name  string
		bic   string
		valid bool
	}{
		{"8 caractères", "AGRIFRPP", true},
		{"11 caractères", "AGRIFRPPXXX", true},
		{"avec espaces", "agri frpp xxx", true},
		{"trop court", "AGRIFRP", false},
		{"pays invalide", "AGRI1RPP", false},
		{"caractère invalide", "AGRIFRP!", false},
		{"vide", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.ValidateBIC(tt.bic)
			if got != tt.valid {
				t.Errorf("ValidateBIC(%q) = %v, attendu %v", tt.bic, got, tt.valid)
			}
		})
	}
}

// ─── SIRET issu d'un SIREN valide ────────────────────────────────────────────

func TestSIRET_ContientSIREN(t *testing.T) {
	// Le SIREN d'un SIRET valide doit aussi être valide
	siret := "73282932000074"
	siren := siret[:9]
	if !domain.ValidateSIRET(siret) {
		t.Fatalf("SIRET %s devrait être valide", siret)
	}
	if !domain.ValidateSIREN(siren) {
		t.Errorf("SIREN %s extrait de SIRET valide devrait aussi être valide", siren)
	}
}
