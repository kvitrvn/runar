package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kvitrvn/runar/internal/config"
	"github.com/kvitrvn/runar/internal/domain"
	"github.com/kvitrvn/runar/internal/repository"
	"github.com/shopspring/decimal"
)

// QuoteService gère les opérations sur les devis.
type QuoteService struct {
	quoteRepo   repository.QuoteRepository
	clientRepo  repository.ClientRepository
	invoiceRepo repository.InvoiceRepository
	audit       *AuditService
	pdf         *PDFService
	cfg         *config.Config
	numbering   domain.NumberingConfig
}

// NewQuoteService crée un service devis.
func NewQuoteService(
	quoteRepo repository.QuoteRepository,
	clientRepo repository.ClientRepository,
	invoiceRepo repository.InvoiceRepository,
	audit *AuditService,
	pdf *PDFService,
	cfg *config.Config,
) *QuoteService {
	return &QuoteService{
		quoteRepo:   quoteRepo,
		clientRepo:  clientRepo,
		invoiceRepo: invoiceRepo,
		audit:       audit,
		pdf:         pdf,
		cfg:         cfg,
		numbering:   domain.DefaultNumberingConfig(),
	}
}

// Create crée un nouveau devis.
func (s *QuoteService) Create(q *domain.Quote) error {
	q.CalculateTotals()

	// Appliquer le taux d'acompte par défaut si non spécifié et config > 0.
	if q.DepositRate.IsZero() && s.cfg.Payment.DefaultDepositRate > 0 {
		q.DepositRate = decimal.NewFromFloat(s.cfg.Payment.DefaultDepositRate)
	}
	assignQuoteDepositPaymentRef(q, s.cfg)

	number, err := s.generateNextNumber(q.IssueDate.Year())
	if err != nil {
		return err
	}
	q.Number = number
	q.State = domain.QuoteStateDraft

	if err := s.quoteRepo.Create(q); err != nil {
		return err
	}

	s.audit.Log("quote", q.ID, domain.AuditActionCreated, "", "")
	return nil
}

// Update met à jour un devis brouillon.
func (s *QuoteService) Update(id int, updates *domain.Quote) error {
	existing, err := s.quoteRepo.GetByID(id)
	if err != nil {
		return err
	}
	if !existing.CanEdit() {
		s.audit.Log("quote", id, domain.AuditActionDenied,
			string(existing.State),
			"tentative de modification d'un devis non brouillon",
		)
		return fmt.Errorf("seul un devis brouillon peut être modifié (état: %s)", existing.State)
	}

	updates.ID = existing.ID
	updates.Number = existing.Number
	updates.State = existing.State
	updates.CreatedAt = existing.CreatedAt
	updates.DepositPaid = false
	updates.DepositPaidAt = nil
	updates.PDFPath = ""
	updates.DepositPaymentRef = existing.DepositPaymentRef
	if !updates.RequiresDeposit() {
		updates.DepositPaymentRef = ""
	} else {
		assignQuoteDepositPaymentRef(updates, s.cfg)
	}
	updates.CalculateTotals()

	oldVal, _ := json.Marshal(existing)
	if err := s.quoteRepo.Replace(id, updates); err != nil {
		return err
	}
	newVal, _ := json.Marshal(updates)
	s.audit.Log("quote", id, domain.AuditActionUpdated, string(oldVal), string(newVal))
	return nil
}

// Delete supprime définitivement un devis brouillon.
func (s *QuoteService) Delete(id int) error {
	q, err := s.quoteRepo.GetByID(id)
	if err != nil {
		return err
	}
	if !q.CanDelete() {
		return fmt.Errorf("seul un brouillon peut être supprimé (état: %s)", q.State)
	}

	oldVal, _ := json.Marshal(q)
	if err := s.quoteRepo.Delete(id); err != nil {
		return err
	}
	s.audit.Log("quote", id, domain.AuditActionDeleted, string(oldVal), "")
	return nil
}

// MarkDepositAsPaid marque l'acompte d'un devis accepté comme payé.
func (s *QuoteService) MarkDepositAsPaid(id int) error {
	q, err := s.quoteRepo.GetByID(id)
	if err != nil {
		return err
	}
	if q.State != domain.QuoteStateAccepted {
		return fmt.Errorf("l'acompte ne peut être marqué payé que sur un devis accepté (état: %s)", q.State)
	}
	if !q.RequiresDeposit() {
		return fmt.Errorf("ce devis n'a pas d'acompte configuré")
	}
	if q.DepositPaid {
		return fmt.Errorf("l'acompte est déjà marqué comme payé")
	}
	now := time.Now()
	q.DepositPaid = true
	q.DepositPaidAt = &now
	if err := s.quoteRepo.Update(id, q); err != nil {
		return err
	}
	s.audit.Log("quote", id, domain.AuditActionUpdated, "deposit_unpaid",
		fmt.Sprintf(`{"action":"deposit_paid","amount":"%s"}`, q.DepositAmount().StringFixed(2)))
	return nil
}

// GetByID retourne un devis par son ID.
func (s *QuoteService) GetByID(id int) (*domain.Quote, error) {
	q, err := s.quoteRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if err := s.loadClient(q); err != nil {
		return nil, err
	}
	return q, nil
}

// List retourne la liste des devis.
func (s *QuoteService) List(search string) ([]domain.Quote, error) {
	return s.quoteRepo.List(search)
}

// MarkAsSent marque un devis comme envoyé (draft → sent).
func (s *QuoteService) MarkAsSent(id int) error {
	q, err := s.quoteRepo.GetByID(id)
	if err != nil {
		return err
	}
	if q.State != domain.QuoteStateDraft {
		return fmt.Errorf("seul un brouillon peut être marqué envoyé (état: %s)", q.State)
	}
	prev := string(q.State)
	q.State = domain.QuoteStateSent
	if err := s.quoteRepo.Update(id, q); err != nil {
		return err
	}
	s.audit.Log("quote", id, domain.AuditActionUpdated, prev, string(q.State))
	return nil
}

// MarkAsAccepted marque un devis comme accepté (draft/sent → accepted).
func (s *QuoteService) MarkAsAccepted(id int) error {
	q, err := s.quoteRepo.GetByID(id)
	if err != nil {
		return err
	}
	if q.State != domain.QuoteStateDraft && q.State != domain.QuoteStateSent {
		return fmt.Errorf("impossible d'accepter un devis %q", q.State)
	}
	prev := string(q.State)
	q.State = domain.QuoteStateAccepted
	if err := s.quoteRepo.Update(id, q); err != nil {
		return err
	}
	s.audit.Log("quote", id, domain.AuditActionUpdated, prev, string(q.State))
	return nil
}

// MarkAsRefused marque un devis comme refusé (draft/sent → refused).
func (s *QuoteService) MarkAsRefused(id int) error {
	q, err := s.quoteRepo.GetByID(id)
	if err != nil {
		return err
	}
	if q.State != domain.QuoteStateDraft && q.State != domain.QuoteStateSent {
		return fmt.Errorf("impossible de refuser un devis %q", q.State)
	}
	prev := string(q.State)
	q.State = domain.QuoteStateRefused
	if err := s.quoteRepo.Update(id, q); err != nil {
		return err
	}
	s.audit.Log("quote", id, domain.AuditActionUpdated, prev, string(q.State))
	return nil
}

// PrepareInvoiceFromQuote construit une domain.Invoice depuis un devis accepté.
// Le appelant doit persister l'invoice via InvoiceService.Create().
func (s *QuoteService) PrepareInvoiceFromQuote(quoteID int) (*domain.Invoice, error) {
	q, err := s.quoteRepo.GetByID(quoteID)
	if err != nil {
		return nil, err
	}
	if !q.CanConvertToInvoice() {
		return nil, fmt.Errorf("seul un devis accepté peut être converti (état: %s)", q.State)
	}

	inv := &domain.Invoice{
		ClientID:         q.ClientID,
		QuoteID:          &quoteID,
		IssueDate:        time.Now(),
		DueDate:          time.Now().AddDate(0, 0, 30),
		DeliveryDate:     time.Now(),
		State:            domain.InvoiceStateDraft,
		Notes:            q.Notes,
		VATApplicable:    s.cfg.VAT.Applicable,
		VATExemptionText: s.cfg.VAT.ExemptionText,
		PaymentDeadline:  s.cfg.Payment.DefaultDeadline,
		LatePenaltyRate:  decimal.NewFromFloat(s.cfg.Payment.LatePenaltyRate),
		RecoveryFee:      decimal.NewFromFloat(s.cfg.Payment.RecoveryFee),
	}

	for _, ql := range q.Lines {
		il := domain.InvoiceLine{
			Description: ql.Description,
			Quantity:    ql.Quantity,
			UnitPriceHT: ql.UnitPriceHT,
			VATRate:     ql.VATRate,
		}
		il.Calculate()
		inv.Lines = append(inv.Lines, il)
	}

	// LEGAL: Si acompte payé, déduire en ligne négative sur la facture.
	if q.DepositPaid && q.RequiresDeposit() {
		depositLine := domain.InvoiceLine{
			Description: fmt.Sprintf("Acompte versé (%.0f%%) - %s", q.DepositRate.InexactFloat64(), q.Number),
			Quantity:    decimal.NewFromInt(1),
			UnitPriceHT: q.DepositAmount().Neg(),
			VATRate:     decimal.Zero,
		}
		depositLine.Calculate()
		inv.Lines = append(inv.Lines, depositLine)
	}

	inv.CalculateTotals()

	s.audit.Log("quote", quoteID, domain.AuditActionUpdated, string(domain.QuoteStateAccepted),
		fmt.Sprintf(`{"action":"converted_to_invoice","client_id":%d}`, q.ClientID))
	return inv, nil
}

// GeneratePDF génère le PDF d'un devis.
func (s *QuoteService) GeneratePDF(id int) (string, error) {
	q, err := s.quoteRepo.GetByID(id)
	if err != nil {
		return "", err
	}
	if err := s.loadClient(q); err != nil {
		return "", err
	}
	pdfPath, err := s.pdf.GenerateQuote(q)
	if err != nil {
		return "", fmt.Errorf("génération PDF devis: %w", err)
	}
	q.PDFPath = pdfPath
	if err := s.quoteRepo.Update(id, q); err != nil {
		return "", err
	}
	s.audit.Log("quote", id, domain.AuditActionPDFGenerated, "", fmt.Sprintf(`{"pdf_path":%q}`, pdfPath))
	return pdfPath, nil
}

func (s *QuoteService) generateNextNumber(year int) (string, error) {
	lastSeq, err := s.quoteRepo.GetLastSequence(year)
	if err != nil {
		return "", err
	}
	return s.numbering.FormatQuoteNumber(year, lastSeq+1), nil
}

func (s *QuoteService) loadClient(q *domain.Quote) error {
	if q == nil || q.Client != nil || q.ClientID <= 0 {
		return nil
	}

	client, err := s.clientRepo.GetByID(q.ClientID)
	if err != nil {
		return fmt.Errorf("chargement client devis: %w", err)
	}
	q.Client = client
	return nil
}

// CreditNoteService gère les opérations sur les avoirs.
// LEGAL: Un avoir ne peut être créé que depuis une facture payée.
type CreditNoteService struct {
	cnRepo      repository.CreditNoteRepository
	invoiceRepo repository.InvoiceRepository
	clientRepo  repository.ClientRepository
	audit       *AuditService
	pdf         *PDFService
	numbering   domain.NumberingConfig
}

// NewCreditNoteService crée un service avoir.
func NewCreditNoteService(
	cnRepo repository.CreditNoteRepository,
	invoiceRepo repository.InvoiceRepository,
	clientRepo repository.ClientRepository,
	audit *AuditService,
	pdf *PDFService,
) *CreditNoteService {
	return &CreditNoteService{
		cnRepo:      cnRepo,
		invoiceRepo: invoiceRepo,
		clientRepo:  clientRepo,
		audit:       audit,
		pdf:         pdf,
		numbering:   domain.DefaultNumberingConfig(),
	}
}

// CreateFromInvoice crée un avoir depuis une facture payée.
// LEGAL: L'avoir doit référencer la facture d'origine (Art. 272 CGI).
func (s *CreditNoteService) CreateFromInvoice(invoiceID int, cn *domain.CreditNote) error {
	invoice, err := s.invoiceRepo.GetByID(invoiceID)
	if err != nil {
		return err
	}

	if !invoice.CanCancel() {
		return &domain.ErrImmutableInvoice{
			InvoiceNumber: invoice.Number,
			State:         string(invoice.State),
		}
	}

	// Numérotation continue
	lastSeq, err := s.cnRepo.GetLastSequence(cn.IssueDate.Year())
	if err != nil {
		return fmt.Errorf("numérotation avoir: %w", err)
	}
	cn.Number = s.numbering.FormatCreditNoteNumber(cn.IssueDate.Year(), lastSeq+1)

	// Référence obligatoire
	cn.InvoiceID = invoiceID
	cn.InvoiceReference = invoice.Number

	// Validation
	if errs := cn.Validate(); len(errs) > 0 {
		return &domain.ValidationErrorList{Errors: errs}
	}

	if err := s.cnRepo.Create(cn); err != nil {
		return err
	}

	s.audit.Log("credit_note", cn.ID, domain.AuditActionCreated, "", "")
	return nil
}

// List retourne tous les avoirs.
func (s *CreditNoteService) List() ([]domain.CreditNote, error) {
	return s.cnRepo.List()
}

// GetByID retourne un avoir par son ID.
func (s *CreditNoteService) GetByID(id int) (*domain.CreditNote, error) {
	return s.cnRepo.GetByID(id)
}

// GeneratePDF génère le PDF d'un avoir.
// LEGAL: Le PDF ne doit jamais être supprimé (conservation 10 ans).
func (s *CreditNoteService) GeneratePDF(id int) (string, error) {
	cn, err := s.cnRepo.GetByID(id)
	if err != nil {
		return "", err
	}
	// Charger le client via la facture d'origine
	if cn.Client == nil && cn.InvoiceID > 0 {
		invoice, err := s.invoiceRepo.GetByID(cn.InvoiceID)
		if err == nil && invoice.ClientID > 0 {
			client, err := s.clientRepo.GetByID(invoice.ClientID)
			if err == nil {
				cn.Client = client
			}
		}
	}
	pdfPath, err := s.pdf.GenerateCreditNote(cn)
	if err != nil {
		return "", fmt.Errorf("génération PDF avoir: %w", err)
	}
	if err := s.cnRepo.UpdatePDFPath(id, pdfPath); err != nil {
		return "", err
	}
	s.audit.Log("credit_note", id, domain.AuditActionPDFGenerated, "", fmt.Sprintf(`{"pdf_path":%q}`, pdfPath))
	return pdfPath, nil
}
