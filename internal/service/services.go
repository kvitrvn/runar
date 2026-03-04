package service

import (
	"github.com/kvitrvn/runar/internal/config"
	"github.com/kvitrvn/runar/internal/repository"
)

// Services regroupe tous les services de l'application.
type Services struct {
	Client     *ClientService
	Invoice    *InvoiceService
	Quote      *QuoteService
	CreditNote *CreditNoteService
	Audit      *AuditService
	PDF        *PDFService
	Export     *ExportService
}

// NewServices crée tous les services.
func NewServices(repos *repository.Repositories, cfg *config.Config) *Services {
	pdfService := NewPDFService(cfg)
	auditService := NewAuditService(repos.Audit)

	return &Services{
		Client:     NewClientService(repos.Client, auditService),
		Invoice:    NewInvoiceService(repos.Invoice, repos.Client, auditService, pdfService, cfg),
		Quote:      NewQuoteService(repos.Quote, repos.Client, repos.Invoice, auditService, pdfService, cfg),
		CreditNote: NewCreditNoteService(repos.CreditNote, repos.Invoice, repos.Client, auditService, pdfService),
		Audit:      auditService,
		PDF:        pdfService,
		Export:     NewExportService(repos.Invoice, repos.CreditNote),
	}
}
