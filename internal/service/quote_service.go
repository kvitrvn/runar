package service

import (
	"github.com/kvitrvn/runar/internal/config"
	"github.com/kvitrvn/runar/internal/domain"
	"github.com/kvitrvn/runar/internal/repository"
)

// QuoteService gère les opérations sur les devis.
type QuoteService struct {
	quoteRepo   repository.QuoteRepository
	clientRepo  repository.ClientRepository
	invoiceRepo repository.InvoiceRepository
	audit       *AuditService
	cfg         *config.Config
	numbering   domain.NumberingConfig
}

// NewQuoteService crée un service devis.
func NewQuoteService(
	quoteRepo repository.QuoteRepository,
	clientRepo repository.ClientRepository,
	invoiceRepo repository.InvoiceRepository,
	audit *AuditService,
	cfg *config.Config,
) *QuoteService {
	return &QuoteService{
		quoteRepo:   quoteRepo,
		clientRepo:  clientRepo,
		invoiceRepo: invoiceRepo,
		audit:       audit,
		cfg:         cfg,
		numbering:   domain.DefaultNumberingConfig(),
	}
}

// Create crée un nouveau devis.
func (s *QuoteService) Create(q *domain.Quote) error {
	q.CalculateTotals()

	number, err := s.generateNextNumber(q.IssueDate.Year())
	if err != nil {
		return err
	}
	q.Number = number

	if err := s.quoteRepo.Create(q); err != nil {
		return err
	}

	s.audit.Log("quote", q.ID, domain.AuditActionCreated, "", "")
	return nil
}

// GetByID retourne un devis par son ID.
func (s *QuoteService) GetByID(id int) (*domain.Quote, error) {
	return s.quoteRepo.GetByID(id)
}

// List retourne la liste des devis.
func (s *QuoteService) List(search string) ([]domain.Quote, error) {
	return s.quoteRepo.List(search)
}

func (s *QuoteService) generateNextNumber(year int) (string, error) {
	lastSeq, err := s.quoteRepo.GetLastSequence(year)
	if err != nil {
		return "", err
	}
	return s.numbering.FormatQuoteNumber(year, lastSeq+1), nil
}

// CreditNoteService gère les opérations sur les avoirs.
// LEGAL: Un avoir ne peut être créé que depuis une facture payée.
type CreditNoteService struct {
	cnRepo      repository.CreditNoteRepository
	invoiceRepo repository.InvoiceRepository
	audit       *AuditService
	numbering   domain.NumberingConfig
}

// NewCreditNoteService crée un service avoir.
func NewCreditNoteService(
	cnRepo repository.CreditNoteRepository,
	invoiceRepo repository.InvoiceRepository,
	audit *AuditService,
) *CreditNoteService {
	return &CreditNoteService{
		cnRepo:      cnRepo,
		invoiceRepo: invoiceRepo,
		audit:       audit,
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
