package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kvitrvn/runar/internal/config"
	"github.com/kvitrvn/runar/internal/domain"
)

// PDFService gère la génération de PDFs.
// LEGAL: Les PDFs ne doivent jamais être supprimés (conservation 10 ans).
// TODO Sprint 5: Implémenter avec maroto/v2
type PDFService struct {
	cfg *config.Config
}

// NewPDFService crée un service PDF.
func NewPDFService(cfg *config.Config) *PDFService {
	return &PDFService{cfg: cfg}
}

// GenerateInvoice génère le PDF d'une facture et retourne son chemin.
// LEGAL: Toutes les mentions obligatoires doivent figurer dans le PDF.
func (s *PDFService) GenerateInvoice(invoice *domain.Invoice) (string, error) {
	if err := os.MkdirAll(s.cfg.PDF.OutputDir, 0750); err != nil {
		return "", fmt.Errorf("création répertoire PDF: %w", err)
	}

	filename := fmt.Sprintf("facture-%s.pdf", invoice.Number)
	pdfPath := filepath.Join(s.cfg.PDF.OutputDir, filename)

	// TODO Sprint 5: Générer avec maroto/v2
	// Pour l'instant, créer un fichier vide en placeholder
	if err := os.WriteFile(pdfPath, []byte("PDF placeholder - Sprint 5"), 0600); err != nil {
		return "", fmt.Errorf("écriture PDF: %w", err)
	}

	return pdfPath, nil
}

// GenerateQuote génère le PDF d'un devis.
func (s *PDFService) GenerateQuote(quote *domain.Quote) (string, error) {
	if err := os.MkdirAll(s.cfg.PDF.OutputDir, 0750); err != nil {
		return "", fmt.Errorf("création répertoire PDF: %w", err)
	}

	filename := fmt.Sprintf("devis-%s.pdf", quote.Number)
	pdfPath := filepath.Join(s.cfg.PDF.OutputDir, filename)

	// TODO Sprint 5
	if err := os.WriteFile(pdfPath, []byte("PDF placeholder - Sprint 5"), 0600); err != nil {
		return "", fmt.Errorf("écriture PDF: %w", err)
	}

	return pdfPath, nil
}
