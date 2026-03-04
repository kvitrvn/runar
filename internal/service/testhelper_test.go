package service_test

import (
	"testing"
	"time"

	"github.com/kvitrvn/runar/internal/config"
	"github.com/kvitrvn/runar/internal/domain"
	"github.com/kvitrvn/runar/internal/repository"
	"github.com/kvitrvn/runar/internal/service"
	"github.com/shopspring/decimal"
)

// setupServices crée des services avec une DB SQLite en mémoire.
func setupServices(t *testing.T) *service.Services {
	t.Helper()

	db, err := repository.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Impossible d'initialiser la DB de test: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	repos := repository.NewRepositories(db)
	cfg := testConfig(t)
	return service.NewServices(repos, cfg)
}

// testConfig retourne une configuration de test.
func testConfig(t *testing.T) *config.Config {
	t.Helper()
	return &config.Config{
		Seller: config.Seller{
			Name:       "Test Auto-Entrepreneur",
			SIRET:      "73282932000074",
			Address:    "1 Rue de la Paix",
			PostalCode: "75001",
			City:       "Paris",
			Country:    "France",
			Email:      "test@example.com",
		},
		VAT: config.VATConf{
			Applicable:    false,
			ExemptionText: domain.VATMentionExemption,
		},
		Payment: config.Payment{
			DefaultDeadline: "30 jours",
			LatePenaltyRate: 13.25,
			RecoveryFee:     40,
		},
		Database: config.DBConf{Path: ":memory:"},
		PDF:      config.PDFConf{OutputDir: t.TempDir()},
	}
}

// createTestClient crée un client de test et le sauvegarde en DB.
func createTestClient(t *testing.T, svc *service.Services) *domain.Client {
	t.Helper()
	c := &domain.Client{
		Name:       "Client Test SA",
		SIRET:      "73282932000074",
		SIREN:      "732829320",
		Address:    "10 Rue du Test",
		PostalCode: "75001",
		City:       "Paris",
		Country:    "France",
		Email:      "client@test.fr",
	}
	if err := svc.Client.Create(c); err != nil {
		t.Fatalf("Création client test: %v", err)
	}
	return c
}

// createTestInvoice crée une facture de test valide.
func createTestInvoice(t *testing.T, svc *service.Services, clientID int) *domain.Invoice {
	t.Helper()
	inv := &domain.Invoice{
		ClientID:         clientID,
		IssueDate:        time.Now(),
		DueDate:          time.Now().Add(30 * 24 * time.Hour),
		DeliveryDate:     time.Now(),
		State:            domain.InvoiceStateDraft,
		VATApplicable:    false,
		VATExemptionText: domain.VATMentionExemption,
		PaymentDeadline:  "30 jours",
		LatePenaltyRate:  decimal.NewFromFloat(13.25),
		RecoveryFee:      decimal.NewFromInt(40),
		Lines: []domain.InvoiceLine{
			{
				Description: "Prestation de service",
				Quantity:    decimal.NewFromFloat(1),
				UnitPriceHT: decimal.NewFromFloat(1000),
				VATRate:     decimal.Zero,
			},
		},
	}
	if err := svc.Invoice.Create(inv); err != nil {
		t.Fatalf("Création facture test: %v", err)
	}
	return inv
}
