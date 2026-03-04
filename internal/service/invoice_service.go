package service

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kvitrvn/runar/internal/config"
	"github.com/kvitrvn/runar/internal/domain"
	"github.com/kvitrvn/runar/internal/repository"
	"github.com/shopspring/decimal"
)

// generatePaymentRef génère un code de virement unique de 8 caractères alphanumériques (A-Z0-9).
// Ce code sert de libellé de virement bancaire pour identifier le paiement.
func generatePaymentRef() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	for i, c := range b {
		b[i] = chars[int(c)%len(chars)]
	}
	return string(b)
}

// InvoiceService gère les opérations sur les factures avec les règles légales.
type InvoiceService struct {
	invoiceRepo repository.InvoiceRepository
	clientRepo  repository.ClientRepository
	audit       *AuditService
	pdf         *PDFService
	cfg         *config.Config
	numbering   domain.NumberingConfig
}

// NewInvoiceService crée un service facture.
func NewInvoiceService(
	invoiceRepo repository.InvoiceRepository,
	clientRepo repository.ClientRepository,
	audit *AuditService,
	pdf *PDFService,
	cfg *config.Config,
) *InvoiceService {
	return &InvoiceService{
		invoiceRepo: invoiceRepo,
		clientRepo:  clientRepo,
		audit:       audit,
		pdf:         pdf,
		cfg:         cfg,
		numbering:   domain.DefaultNumberingConfig(),
	}
}

// NewFromConfig crée une facture pré-remplie avec les valeurs de config.
func (s *InvoiceService) NewFromConfig() *domain.Invoice {
	return &domain.Invoice{
		State:            domain.InvoiceStateDraft,
		IssueDate:        time.Now(),
		DueDate:          time.Now().AddDate(0, 0, 30),
		VATApplicable:    s.cfg.VAT.Applicable,
		VATExemptionText: s.cfg.VAT.ExemptionText,
		PaymentDeadline:  s.cfg.Payment.DefaultDeadline,
		LatePenaltyRate:  decimal.NewFromFloat(s.cfg.Payment.LatePenaltyRate),
		RecoveryFee:      decimal.NewFromFloat(s.cfg.Payment.RecoveryFee),
	}
}

// Create crée une nouvelle facture.
// LEGAL: Validation complète + numérotation continue garantie.
func (s *InvoiceService) Create(inv *domain.Invoice) error {
	// Calculer totaux depuis les lignes
	inv.CalculateTotals()

	// Générer numéro (avant validation car le numéro est requis)
	number, err := s.GenerateNextNumber(inv.IssueDate.Year())
	if err != nil {
		return fmt.Errorf("génération numéro: %w", err)
	}
	inv.Number = number

	// Générer le libellé de virement unique
	if inv.PaymentRef == "" {
		inv.PaymentRef = generatePaymentRef()
	}

	// Charger client pour validation B2B
	if inv.ClientID > 0 && inv.Client == nil {
		client, err := s.clientRepo.GetByID(inv.ClientID)
		if err != nil {
			return fmt.Errorf("chargement client: %w", err)
		}
		inv.Client = client
	}

	// Validation légale complète
	if errs := inv.Validate(); len(errs) > 0 {
		return &domain.ValidationErrorList{Errors: errs}
	}

	if err := s.invoiceRepo.Create(inv); err != nil {
		return fmt.Errorf("sauvegarde facture: %w", err)
	}

	newVal, _ := json.Marshal(inv)
	s.audit.Log("invoice", inv.ID, domain.AuditActionCreated, "", string(newVal))
	return nil
}

// Update met à jour une facture.
// LEGAL: Impossible si la facture est payée ou annulée (Art. L441-9 Code de Commerce).
func (s *InvoiceService) Update(id int, updates *domain.Invoice) error {
	existing, err := s.invoiceRepo.GetByID(id)
	if err != nil {
		return err
	}

	// RÈGLE CRITIQUE: Facture payée = immuable
	if !existing.CanEdit() {
		// LEGAL: Logger la tentative de modification interdite
		s.audit.Log("invoice", id, domain.AuditActionDenied,
			string(existing.State),
			"tentative de modification d'une facture immuable",
		)
		return &domain.ErrImmutableInvoice{
			InvoiceNumber: existing.Number,
			State:         string(existing.State),
		}
	}

	updates.CalculateTotals()

	if errs := updates.Validate(); len(errs) > 0 {
		return &domain.ValidationErrorList{Errors: errs}
	}

	oldVal, _ := json.Marshal(existing)
	if err := s.invoiceRepo.Update(id, updates); err != nil {
		return err
	}

	newVal, _ := json.Marshal(updates)
	s.audit.Log("invoice", id, domain.AuditActionUpdated, string(oldVal), string(newVal))
	return nil
}

// MarkAsIssued émet une facture brouillon (draft → issued).
func (s *InvoiceService) MarkAsIssued(id int) error {
	invoice, err := s.invoiceRepo.GetByID(id)
	if err != nil {
		return err
	}
	if invoice.State != domain.InvoiceStateDraft {
		return fmt.Errorf("seule une facture brouillon peut être émise (état: %s)", invoice.State)
	}
	prev := string(invoice.State)
	invoice.State = domain.InvoiceStateIssued
	if err := s.invoiceRepo.Update(id, invoice); err != nil {
		return err
	}
	s.audit.Log("invoice", id, domain.AuditActionUpdated, prev, string(domain.InvoiceStateIssued))
	return nil
}

// MarkAsSent marque une facture comme envoyée (draft/issued → sent).
func (s *InvoiceService) MarkAsSent(id int) error {
	invoice, err := s.invoiceRepo.GetByID(id)
	if err != nil {
		return err
	}
	if invoice.State != domain.InvoiceStateDraft && invoice.State != domain.InvoiceStateIssued {
		return fmt.Errorf("seule une facture brouillon ou émise peut être envoyée (état: %s)", invoice.State)
	}
	prev := string(invoice.State)
	invoice.State = domain.InvoiceStateSent
	if err := s.invoiceRepo.Update(id, invoice); err != nil {
		return err
	}
	s.audit.Log("invoice", id, domain.AuditActionUpdated, prev, string(domain.InvoiceStateSent))
	return nil
}

// Delete supprime définitivement un brouillon.
// LEGAL: Seul un brouillon (jamais émis) peut être supprimé.
func (s *InvoiceService) Delete(id int) error {
	invoice, err := s.invoiceRepo.GetByID(id)
	if err != nil {
		return err
	}
	if !invoice.CanDelete() {
		return fmt.Errorf("seul un brouillon peut être supprimé (état: %s)", invoice.State)
	}
	if err := s.invoiceRepo.SoftDelete(id); err != nil {
		return err
	}
	s.audit.Log("invoice", id, domain.AuditActionUpdated, string(invoice.State), "deleted")
	return nil
}

// MarkAsPaid marque une facture comme payée et la VERROUILLE définitivement.
// LEGAL: Après cet appel, la facture devient IMMUABLE (Art. L441-9 Code de Commerce).
func (s *InvoiceService) MarkAsPaid(id int, paidDate time.Time) error {
	invoice, err := s.invoiceRepo.GetByID(id)
	if err != nil {
		return err
	}

	if !invoice.CanMarkAsPaid() {
		return fmt.Errorf("facture %s dans l'état '%s' ne peut pas être marquée comme payée",
			invoice.Number, invoice.State)
	}

	oldVal, _ := json.Marshal(map[string]string{"state": string(invoice.State)})

	now := time.Now()
	invoice.State = domain.InvoiceStatePaid
	invoice.PaidDate = &paidDate
	invoice.PaidLockedAt = &now

	if err := s.invoiceRepo.Update(id, invoice); err != nil {
		return err
	}

	// LEGAL: Logger le verrouillage comme action critique
	newVal, _ := json.Marshal(map[string]interface{}{
		"state":          string(domain.InvoiceStatePaid),
		"paid_date":      paidDate,
		"paid_locked_at": now,
	})
	s.audit.Log("invoice", id, domain.AuditActionPaidLocked, string(oldVal), string(newVal))
	return nil
}

// GetByID retourne une facture par son ID.
func (s *InvoiceService) GetByID(id int) (*domain.Invoice, error) {
	return s.invoiceRepo.GetByID(id)
}

// List retourne la liste des factures avec filtres.
func (s *InvoiceService) List(filters repository.InvoiceFilters) ([]domain.Invoice, error) {
	return s.invoiceRepo.List(filters)
}

// GenerateNextNumber génère le prochain numéro de facture.
// LEGAL: Numérotation continue sans trou (Art. 242 nonies A annexe II CGI).
func (s *InvoiceService) GenerateNextNumber(year int) (string, error) {
	lastSeq, err := s.invoiceRepo.GetLastSequence(year)
	if err != nil {
		return "", fmt.Errorf("récupération dernière séquence: %w", err)
	}

	nextSeq := lastSeq + 1
	number := s.numbering.FormatInvoiceNumber(year, nextSeq)

	// Vérification de doublon (protection contre race condition)
	exists, err := s.invoiceRepo.NumberExists(number)
	if err != nil {
		return "", err
	}
	if exists {
		return "", fmt.Errorf("numéro %s existe déjà (race condition détectée)", number)
	}

	return number, nil
}

// GeneratePDF génère le PDF d'une facture.
// LEGAL: Le PDF ne doit jamais être supprimé (conservation 10 ans).
func (s *InvoiceService) GeneratePDF(id int) (string, error) {
	invoice, err := s.invoiceRepo.GetByID(id)
	if err != nil {
		return "", err
	}

	if invoice.Client == nil {
		client, err := s.clientRepo.GetByID(invoice.ClientID)
		if err != nil {
			return "", err
		}
		invoice.Client = client
	}

	pdfPath, err := s.pdf.GenerateInvoice(invoice)
	if err != nil {
		return "", fmt.Errorf("génération PDF: %w", err)
	}

	invoice.PDFPath = pdfPath
	if err := s.invoiceRepo.Update(id, invoice); err != nil {
		return "", err
	}

	s.audit.Log("invoice", id, domain.AuditActionPDFGenerated, "", fmt.Sprintf(`{"pdf_path":%q}`, pdfPath))
	return pdfPath, nil
}
