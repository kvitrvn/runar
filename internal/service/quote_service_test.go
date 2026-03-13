package service_test

import (
	"testing"
	"time"

	"github.com/kvitrvn/runar/internal/config"
	"github.com/kvitrvn/runar/internal/domain"
	"github.com/kvitrvn/runar/internal/service"
	"github.com/shopspring/decimal"
)

func TestQuoteService_Create_GeneratesDepositPaymentRefWhenBankInfoValid(t *testing.T) {
	svc := setupServicesWithConfig(t, func(cfg *config.Config) {
		cfg.Payment.IBAN = "FR76 3000 6000 0112 3456 7890 189"
		cfg.Payment.BIC = "AGRIFRPP"
	})
	client := createTestClient(t, svc)

	q := &domain.Quote{
		ClientID:    client.ID,
		IssueDate:   time.Now(),
		ExpiryDate:  time.Now().AddDate(0, 0, 30),
		DepositRate: decimal.NewFromInt(30),
		Lines: []domain.QuoteLine{
			{
				Description: "Prestation",
				Quantity:    decimal.NewFromInt(1),
				UnitPriceHT: decimal.NewFromInt(1000),
			},
		},
	}

	if err := svc.Quote.Create(q); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if q.DepositPaymentRef == "" {
		t.Fatal("DepositPaymentRef doit être généré")
	}

	loaded, err := svc.Quote.GetByID(q.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if loaded.DepositPaymentRef == "" {
		t.Fatal("DepositPaymentRef doit être persisté")
	}
	if loaded.DepositPaymentRef != q.DepositPaymentRef {
		t.Fatalf("DepositPaymentRef persisté = %q, attendu %q", loaded.DepositPaymentRef, q.DepositPaymentRef)
	}
}

func TestQuoteService_Create_DoesNotGenerateDepositPaymentRefWithoutDeposit(t *testing.T) {
	svc := setupServicesWithConfig(t, func(cfg *config.Config) {
		cfg.Payment.IBAN = "FR76 3000 6000 0112 3456 7890 189"
		cfg.Payment.BIC = "AGRIFRPP"
	})
	client := createTestClient(t, svc)

	q := &domain.Quote{
		ClientID:   client.ID,
		IssueDate:  time.Now(),
		ExpiryDate: time.Now().AddDate(0, 0, 30),
		Lines: []domain.QuoteLine{
			{
				Description: "Prestation",
				Quantity:    decimal.NewFromInt(1),
				UnitPriceHT: decimal.NewFromInt(1000),
			},
		},
	}

	if err := svc.Quote.Create(q); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if q.DepositPaymentRef != "" {
		t.Fatalf("DepositPaymentRef = %q, attendu vide", q.DepositPaymentRef)
	}
}

func TestQuoteService_Create_DoesNotGenerateDepositPaymentRefWithoutValidBankInfo(t *testing.T) {
	tests := []struct {
		name string
		iban string
		bic  string
	}{
		{name: "sans iban", iban: "", bic: "AGRIFRPP"},
		{name: "iban invalide", iban: "FR76 XXXX", bic: "AGRIFRPP"},
		{name: "sans bic", iban: "FR76 3000 6000 0112 3456 7890 189", bic: ""},
		{name: "bic invalide", iban: "FR76 3000 6000 0112 3456 7890 189", bic: "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := setupServicesWithConfig(t, func(cfg *config.Config) {
				cfg.Payment.IBAN = tt.iban
				cfg.Payment.BIC = tt.bic
			})
			client := createTestClient(t, svc)

			q := &domain.Quote{
				ClientID:    client.ID,
				IssueDate:   time.Now(),
				ExpiryDate:  time.Now().AddDate(0, 0, 30),
				DepositRate: decimal.NewFromInt(30),
				Lines: []domain.QuoteLine{
					{
						Description: "Prestation",
						Quantity:    decimal.NewFromInt(1),
						UnitPriceHT: decimal.NewFromInt(1000),
					},
				},
			}

			if err := svc.Quote.Create(q); err != nil {
				t.Fatalf("Create: %v", err)
			}
			if q.DepositPaymentRef != "" {
				t.Fatalf("DepositPaymentRef = %q, attendu vide", q.DepositPaymentRef)
			}
		})
	}
}

func TestQuoteService_Update_UpdatesDraft(t *testing.T) {
	svc := setupServicesWithConfig(t, func(cfg *config.Config) {
		cfg.Payment.IBAN = "FR76 3000 6000 0112 3456 7890 189"
		cfg.Payment.BIC = "AGRIFRPP"
	})
	client := createTestClient(t, svc)
	q := createTestQuote(t, svc, client.ID)

	issueDate := time.Now().AddDate(0, 0, 2)
	expiryDate := time.Now().AddDate(0, 0, 45)
	updates := &domain.Quote{
		ClientID:    client.ID,
		IssueDate:   issueDate,
		ExpiryDate:  expiryDate,
		Notes:       "Devis mis à jour",
		DepositRate: decimal.NewFromInt(25),
		Lines: []domain.QuoteLine{
			{
				Description: "Audit initial",
				Quantity:    decimal.NewFromInt(2),
				UnitPriceHT: decimal.NewFromInt(500),
			},
			{
				Description: "Implémentation",
				Quantity:    decimal.NewFromInt(3),
				UnitPriceHT: decimal.NewFromInt(250),
			},
		},
	}

	if err := svc.Quote.Update(q.ID, updates); err != nil {
		t.Fatalf("Update: %v", err)
	}

	loaded, err := svc.Quote.GetByID(q.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if loaded.Number != q.Number {
		t.Fatalf("Number = %q, attendu %q", loaded.Number, q.Number)
	}
	if loaded.State != domain.QuoteStateDraft {
		t.Fatalf("State = %q, attendu %q", loaded.State, domain.QuoteStateDraft)
	}
	if !loaded.IssueDate.Equal(issueDate) {
		t.Fatalf("IssueDate = %v, attendu %v", loaded.IssueDate, issueDate)
	}
	if !loaded.ExpiryDate.Equal(expiryDate) {
		t.Fatalf("ExpiryDate = %v, attendu %v", loaded.ExpiryDate, expiryDate)
	}
	if loaded.Notes != "Devis mis à jour" {
		t.Fatalf("Notes = %q, attendu devis mis à jour", loaded.Notes)
	}
	if len(loaded.Lines) != 2 {
		t.Fatalf("Lines = %d, attendu 2", len(loaded.Lines))
	}
	if loaded.Lines[0].Description != "Audit initial" {
		t.Fatalf("Line[0].Description = %q, attendu Audit initial", loaded.Lines[0].Description)
	}
	if !loaded.TotalHT.Equal(decimal.NewFromInt(1750)) {
		t.Fatalf("TotalHT = %s, attendu 1750", loaded.TotalHT)
	}
	if !loaded.TotalTTC.Equal(decimal.NewFromInt(1750)) {
		t.Fatalf("TotalTTC = %s, attendu 1750", loaded.TotalTTC)
	}
	if loaded.DepositPaymentRef == "" {
		t.Fatal("DepositPaymentRef doit être généré lors de l'ajout d'un acompte")
	}
	if !loaded.DepositRate.Equal(decimal.NewFromInt(25)) {
		t.Fatalf("DepositRate = %s, attendu 25", loaded.DepositRate)
	}
}

func TestQuoteService_Update_RejectsNonDraft(t *testing.T) {
	svc := setupServices(t)
	client := createTestClient(t, svc)
	q := createTestQuote(t, svc, client.ID)

	if err := svc.Quote.MarkAsSent(q.ID); err != nil {
		t.Fatalf("MarkAsSent: %v", err)
	}

	err := svc.Quote.Update(q.ID, &domain.Quote{
		ClientID:   client.ID,
		IssueDate:  time.Now(),
		ExpiryDate: time.Now().AddDate(0, 0, 30),
		Lines: []domain.QuoteLine{
			{
				Description: "Tentative",
				Quantity:    decimal.NewFromInt(1),
				UnitPriceHT: decimal.NewFromInt(100),
			},
		},
	})
	if err == nil {
		t.Fatal("Update sur devis non brouillon doit échouer")
	}
}

func TestQuoteService_Delete_DeletesDraft(t *testing.T) {
	svc := setupServices(t)
	client := createTestClient(t, svc)
	q := createTestQuote(t, svc, client.ID)

	if err := svc.Quote.Delete(q.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := svc.Quote.GetByID(q.ID); err == nil {
		t.Fatal("GetByID après Delete doit échouer")
	}
}

func TestQuoteService_Delete_RejectsNonDraft(t *testing.T) {
	svc := setupServices(t)
	client := createTestClient(t, svc)
	q := createTestQuote(t, svc, client.ID)

	if err := svc.Quote.MarkAsSent(q.ID); err != nil {
		t.Fatalf("MarkAsSent: %v", err)
	}

	if err := svc.Quote.Delete(q.ID); err == nil {
		t.Fatal("Delete sur devis non brouillon doit échouer")
	}
}

func TestQuoteService_GetByID_LoadsClient(t *testing.T) {
	svc := setupServices(t)
	client := createTestClient(t, svc)
	q := createTestQuote(t, svc, client.ID)

	loaded, err := svc.Quote.GetByID(q.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if loaded.Client == nil {
		t.Fatal("GetByID doit charger le client")
	}
	if loaded.Client.Name != client.Name {
		t.Fatalf("Client.Name = %q, attendu %q", loaded.Client.Name, client.Name)
	}
}

func TestGetQuoteDepositPaymentInfo(t *testing.T) {
	cfg := &config.Config{}
	cfg.Payment.IBAN = "fr76 3000 6000 0112 3456 7890 189"
	cfg.Payment.BIC = "agri frpp xxx"

	q := &domain.Quote{
		DepositRate:       decimal.NewFromInt(30),
		DepositPaymentRef: "ABC12345",
		TotalHT:           decimal.NewFromInt(1000),
	}

	info := service.GetQuoteDepositPaymentInfo(q, cfg)
	if info == nil {
		t.Fatal("GetQuoteDepositPaymentInfo doit retourner un bloc")
	}
	if info.Amount != "300.00" {
		t.Fatalf("Amount = %q, attendu 300.00", info.Amount)
	}
	if info.IBAN != "FR7630006000011234567890189" {
		t.Fatalf("IBAN = %q, attendu normalisé", info.IBAN)
	}
	if info.BIC != "AGRIFRPPXXX" {
		t.Fatalf("BIC = %q, attendu normalisé", info.BIC)
	}
	if info.PaymentRef != "ABC12345" {
		t.Fatalf("PaymentRef = %q, attendu ABC12345", info.PaymentRef)
	}
}

func TestGetQuoteDepositPaymentInfo_HidesWithoutDisplayConditions(t *testing.T) {
	cfg := &config.Config{}
	cfg.Payment.IBAN = "FR76 3000 6000 0112 3456 7890 189"
	cfg.Payment.BIC = "AGRIFRPP"

	tests := []struct {
		name string
		q    *domain.Quote
		cfg  *config.Config
	}{
		{
			name: "sans acompte",
			q: &domain.Quote{
				DepositPaymentRef: "ABC12345",
				TotalHT:           decimal.NewFromInt(1000),
			},
			cfg: cfg,
		},
		{
			name: "sans code",
			q: &domain.Quote{
				DepositRate: decimal.NewFromInt(30),
				TotalHT:     decimal.NewFromInt(1000),
			},
			cfg: cfg,
		},
		{
			name: "iban invalide",
			q: &domain.Quote{
				DepositRate:       decimal.NewFromInt(30),
				DepositPaymentRef: "ABC12345",
				TotalHT:           decimal.NewFromInt(1000),
			},
			cfg: &config.Config{Payment: config.Payment{IBAN: "FR76 XXXX", BIC: "AGRIFRPP"}},
		},
		{
			name: "bic invalide",
			q: &domain.Quote{
				DepositRate:       decimal.NewFromInt(30),
				DepositPaymentRef: "ABC12345",
				TotalHT:           decimal.NewFromInt(1000),
			},
			cfg: &config.Config{Payment: config.Payment{IBAN: "FR76 3000 6000 0112 3456 7890 189", BIC: "123"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if info := service.GetQuoteDepositPaymentInfo(tt.q, tt.cfg); info != nil {
				t.Fatalf("GetQuoteDepositPaymentInfo() = %+v, attendu nil", info)
			}
		})
	}
}
