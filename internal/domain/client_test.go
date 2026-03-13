package domain_test

import (
	"testing"

	"github.com/kvitrvn/runar/internal/domain"
)

func TestClient_IsCompany(t *testing.T) {
	tests := []struct {
		name string
		c    domain.Client
		want bool
	}{
		{"avec SIRET", domain.Client{SIRET: "73282932000074"}, true},
		{"avec SIREN", domain.Client{SIREN: "732829320"}, true},
		{"particulier", domain.Client{Name: "Jean Dupont"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.IsCompany(); got != tt.want {
				t.Errorf("IsCompany() = %v, attendu %v", got, tt.want)
			}
		})
	}
}

func TestClient_Validate_NomObligatoire(t *testing.T) {
	c := &domain.Client{
		Name:    "",
		Address: "1 rue Test",
	}
	errs := c.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "name" {
			found = true
		}
	}
	if !found {
		t.Error("Erreur name manquant non trouvée")
	}
}

func TestClient_Validate_SIRETInvalide(t *testing.T) {
	c := &domain.Client{
		Name:  "Acme",
		SIRET: "12345678901234", // Luhn invalide
	}
	errs := c.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "siret" {
			found = true
		}
	}
	if !found {
		t.Error("Erreur SIRET invalide non trouvée")
	}
}

func TestClient_Validate_EmailInvalide(t *testing.T) {
	c := &domain.Client{
		Name:  "Test",
		Email: "pas-un-email",
	}
	errs := c.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "email" {
			found = true
		}
	}
	if !found {
		t.Error("Erreur email invalide non trouvée")
	}
}

func TestClient_Validate_Valide(t *testing.T) {
	c := &domain.Client{
		Name:       "Dupont SA",
		SIRET:      "73282932000074",
		SIREN:      "732829320",
		Email:      "contact@dupont.fr",
		Address:    "1 Rue Test",
		PostalCode: "75001",
		City:       "Paris",
	}
	errs := c.Validate()
	if len(errs) != 0 {
		t.Errorf("Client valide: attendu 0 erreurs, got %d: %v", len(errs), errs)
	}
}
