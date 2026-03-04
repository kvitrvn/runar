package service

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kvitrvn/runar/internal/repository"
)

// ExportService gère l'export des données comptables en CSV.
type ExportService struct {
	invoiceRepo repository.InvoiceRepository
	cnRepo      repository.CreditNoteRepository
}

// NewExportService crée un service d'export.
func NewExportService(invoiceRepo repository.InvoiceRepository, cnRepo repository.CreditNoteRepository) *ExportService {
	return &ExportService{invoiceRepo: invoiceRepo, cnRepo: cnRepo}
}

// ExportInvoicesCSV exporte les factures d'une année en CSV (séparateur ';', compatible Excel FR).
// Retourne le chemin du fichier créé.
func (s *ExportService) ExportInvoicesCSV(year int, outputDir string) (string, error) {
	filters := repository.InvoiceFilters{}
	if year > 0 {
		filters.Year = year
	}
	invoices, err := s.invoiceRepo.List(filters)
	if err != nil {
		return "", fmt.Errorf("chargement factures: %w", err)
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Comma = ';'

	// En-tête CSV
	_ = w.Write([]string{
		"Numéro", "Date émission", "Échéance", "Date paiement",
		"Client ID", "Total HT", "TVA", "Total TTC", "État",
		"Délai paiement", "Réf virement", "PDF",
	})

	for _, inv := range invoices {
		paidDate := ""
		if inv.PaidDate != nil {
			paidDate = inv.PaidDate.Format("02/01/2006")
		}
		_ = w.Write([]string{
			inv.Number,
			inv.IssueDate.Format("02/01/2006"),
			inv.DueDate.Format("02/01/2006"),
			paidDate,
			fmt.Sprint(inv.ClientID),
			inv.TotalHT.StringFixed(2),
			inv.VATAmount.StringFixed(2),
			inv.TotalTTC.StringFixed(2),
			string(inv.State),
			inv.PaymentDeadline,
			inv.PaymentRef,
			inv.PDFPath,
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", fmt.Errorf("écriture CSV: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0750); err != nil {
		return "", fmt.Errorf("création répertoire: %w", err)
	}

	suffix := "all"
	if year > 0 {
		suffix = fmt.Sprint(year)
	}
	path := filepath.Join(outputDir, fmt.Sprintf("export-factures-%s-%s.csv",
		suffix, time.Now().Format("20060102")))

	if err := os.WriteFile(path, buf.Bytes(), 0640); err != nil {
		return "", fmt.Errorf("écriture fichier: %w", err)
	}
	return path, nil
}

// ExportCreditNotesCSV exporte les avoirs en CSV (séparateur ';', compatible Excel FR).
// Retourne le chemin du fichier créé.
func (s *ExportService) ExportCreditNotesCSV(outputDir string) (string, error) {
	cns, err := s.cnRepo.List()
	if err != nil {
		return "", fmt.Errorf("chargement avoirs: %w", err)
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Comma = ';'

	_ = w.Write([]string{
		"Numéro", "Date", "Facture d'origine", "Motif",
		"Total HT", "TVA", "Total TTC", "PDF",
	})

	for _, cn := range cns {
		_ = w.Write([]string{
			cn.Number,
			cn.IssueDate.Format("02/01/2006"),
			cn.InvoiceReference,
			cn.Reason,
			cn.TotalHT.StringFixed(2),
			cn.VATAmount.StringFixed(2),
			cn.TotalTTC.StringFixed(2),
			cn.PDFPath,
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", fmt.Errorf("écriture CSV: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0750); err != nil {
		return "", fmt.Errorf("création répertoire: %w", err)
	}

	path := filepath.Join(outputDir, fmt.Sprintf("export-avoirs-%s.csv",
		time.Now().Format("20060102")))

	if err := os.WriteFile(path, buf.Bytes(), 0640); err != nil {
		return "", fmt.Errorf("écriture fichier: %w", err)
	}
	return path, nil
}
