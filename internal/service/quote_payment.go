package service

import (
	appcfg "github.com/kvitrvn/runar/internal/config"
	"github.com/kvitrvn/runar/internal/domain"
)

// QuoteDepositPaymentInfo contient les informations affichables pour payer l'acompte d'un devis.
type QuoteDepositPaymentInfo struct {
	Amount     string
	IBAN       string
	BIC        string
	PaymentRef string
}

func canGenerateQuoteDepositPaymentRef(q *domain.Quote, cfg *appcfg.Config) bool {
	if q == nil || cfg == nil {
		return false
	}
	return q.RequiresDeposit() &&
		domain.ValidateIBAN(cfg.Payment.IBAN) &&
		domain.ValidateBIC(cfg.Payment.BIC)
}

func assignQuoteDepositPaymentRef(q *domain.Quote, cfg *appcfg.Config) {
	if !canGenerateQuoteDepositPaymentRef(q, cfg) || q.DepositPaymentRef != "" {
		return
	}
	q.DepositPaymentRef = generatePaymentRef()
}

// GetQuoteDepositPaymentInfo retourne les informations de paiement de l'acompte
// si le devis doit les afficher.
func GetQuoteDepositPaymentInfo(q *domain.Quote, cfg *appcfg.Config) *QuoteDepositPaymentInfo {
	if !canGenerateQuoteDepositPaymentRef(q, cfg) || q.DepositPaymentRef == "" {
		return nil
	}

	return &QuoteDepositPaymentInfo{
		Amount:     q.DepositAmount().StringFixed(2),
		IBAN:       domain.NormalizeIBAN(cfg.Payment.IBAN),
		BIC:        domain.NormalizeBIC(cfg.Payment.BIC),
		PaymentRef: q.DepositPaymentRef,
	}
}
