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
