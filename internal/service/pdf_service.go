package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/props"
	appcfg "github.com/kvitrvn/runar/internal/config"
	"github.com/kvitrvn/runar/internal/domain"
)

// PDFService gère la génération de PDFs.
// LEGAL: Les PDFs ne doivent jamais être supprimés (conservation 10 ans).
type PDFService struct {
	cfg *appcfg.Config
}

// NewPDFService crée un service PDF.
func NewPDFService(cfg *appcfg.Config) *PDFService {
	return &PDFService{cfg: cfg}
}

// GenerateInvoice génère le PDF d'une facture et retourne son chemin.
// LEGAL: Toutes les mentions obligatoires doivent figurer dans le PDF.
func (s *PDFService) GenerateInvoice(invoice *domain.Invoice) (string, error) {
	if err := os.MkdirAll(s.cfg.PDF.OutputDir, 0750); err != nil {
		return "", fmt.Errorf("création répertoire PDF: %w", err)
	}

	m := buildInvoicePDF(invoice, &s.cfg.Seller)

	doc, err := m.Generate()
	if err != nil {
		return "", fmt.Errorf("génération PDF facture: %w", err)
	}

	filename := fmt.Sprintf("facture-%s.pdf", invoice.Number)
	pdfPath := filepath.Join(s.cfg.PDF.OutputDir, filename)

	if err := doc.Save(pdfPath); err != nil {
		return "", fmt.Errorf("écriture PDF: %w", err)
	}

	return pdfPath, nil
}

// GenerateCreditNote génère le PDF d'un avoir.
func (s *PDFService) GenerateCreditNote(cn *domain.CreditNote) (string, error) {
	if err := os.MkdirAll(s.cfg.PDF.OutputDir, 0750); err != nil {
		return "", fmt.Errorf("création répertoire PDF: %w", err)
	}

	m := buildCreditNotePDF(cn, &s.cfg.Seller)

	doc, err := m.Generate()
	if err != nil {
		return "", fmt.Errorf("génération PDF avoir: %w", err)
	}

	filename := fmt.Sprintf("avoir-%s.pdf", cn.Number)
	pdfPath := filepath.Join(s.cfg.PDF.OutputDir, filename)

	if err := doc.Save(pdfPath); err != nil {
		return "", fmt.Errorf("écriture PDF avoir: %w", err)
	}

	return pdfPath, nil
}

// GenerateQuote génère le PDF d'un devis.
func (s *PDFService) GenerateQuote(quote *domain.Quote) (string, error) {
	if err := os.MkdirAll(s.cfg.PDF.OutputDir, 0750); err != nil {
		return "", fmt.Errorf("création répertoire PDF: %w", err)
	}

	m := buildQuotePDF(quote, &s.cfg.Seller)

	doc, err := m.Generate()
	if err != nil {
		return "", fmt.Errorf("génération PDF devis: %w", err)
	}

	filename := fmt.Sprintf("devis-%s.pdf", quote.Number)
	pdfPath := filepath.Join(s.cfg.PDF.OutputDir, filename)

	if err := doc.Save(pdfPath); err != nil {
		return "", fmt.Errorf("écriture PDF: %w", err)
	}

	return pdfPath, nil
}

// ─── Styles partagés ────────────────────────────────────────────────────────

var (
	colorBlack  = &props.Color{Red: 0, Green: 0, Blue: 0}
	colorGray   = &props.Color{Red: 100, Green: 100, Blue: 100}
	colorLight  = &props.Color{Red: 240, Green: 240, Blue: 240}
	colorBlue   = &props.Color{Red: 14, Green: 165, Blue: 233}
	colorDanger = &props.Color{Red: 220, Green: 38, Blue: 38}

	cellHeader = &props.Cell{
		BackgroundColor: colorBlue,
		BorderColor:     colorBlue,
		BorderType:      border.Full,
		BorderThickness: 0.3,
	}
	cellAlt = &props.Cell{
		BackgroundColor: colorLight,
	}
	cellBorder = &props.Cell{
		BorderColor:     colorGray,
		BorderType:      border.Bottom,
		BorderThickness: 0.2,
	}

	txtNormal = props.Text{Size: 10, Color: colorBlack}
	txtSmall  = props.Text{Size: 9, Color: colorGray}
	txtBold   = props.Text{Size: 10, Style: fontstyle.Bold, Color: colorBlack}
	txtTitle  = props.Text{Size: 16, Style: fontstyle.Bold, Color: colorBlue, Align: align.Right}
	txtRight  = props.Text{Size: 10, Align: align.Right, Color: colorBlack}
	txtRightB = props.Text{Size: 11, Style: fontstyle.Bold, Align: align.Right, Color: colorBlack}
	txtCenter = props.Text{Size: 9, Align: align.Center, Color: colorBlack}
	txtWhiteB = props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Center,
		Color: &props.Color{Red: 255, Green: 255, Blue: 255}}
)

// ─── Template Facture ───────────────────────────────────────────────────────

func buildInvoicePDF(inv *domain.Invoice, seller *appcfg.Seller) core.Maroto {
	cfg := config.NewBuilder().
		WithPageSize(pagesize.A4).
		WithLeftMargin(15).
		WithRightMargin(15).
		WithTopMargin(15).
		WithBottomMargin(20).
		Build()

	m := maroto.New(cfg)

	// ── En-tête : vendeur à gauche, titre à droite
	m.AddRows(
		row.New(8).Add(
			col.New(7).Add(text.New(seller.Name, props.Text{Size: 14, Style: fontstyle.Bold})),
			col.New(5).Add(text.New("FACTURE", txtTitle)),
		),
		row.New(5).Add(
			col.New(7).Add(text.New("SIRET : "+seller.SIRET, txtSmall)),
			col.New(5).Add(text.New("N° "+inv.Number, props.Text{Size: 12, Style: fontstyle.Bold, Align: align.Right, Color: colorBlue})),
		),
		row.New(5).Add(
			col.New(7).Add(text.New(seller.Address, txtSmall)),
			col.New(5).Add(text.New("Émise le : "+inv.IssueDate.Format("02/01/2006"), props.Text{Size: 9, Align: align.Right, Color: colorGray})),
		),
		row.New(5).Add(
			col.New(7).Add(text.New(seller.PostalCode+" "+seller.City, txtSmall)),
			col.New(5).Add(text.New("Échéance : "+inv.DueDate.Format("02/01/2006"), props.Text{Size: 9, Align: align.Right, Color: colorGray})),
		),
	)

	if seller.Email != "" {
		m.AddRow(5,
			col.New(7).Add(text.New(seller.Email, txtSmall)),
			col.New(5).Add(text.New("Livraison : "+inv.DeliveryDate.Format("02/01/2006"), props.Text{Size: 9, Align: align.Right, Color: colorGray})),
		)
	}
	if seller.Phone != "" {
		m.AddRow(5, col.New(7).Add(text.New(seller.Phone, txtSmall)), col.New(5))
	}

	m.AddRow(6) // espace

	// ── Bloc client
	m.AddRow(6,
		col.New(12).Add(text.New("FACTURER À :", props.Text{Size: 9, Style: fontstyle.Bold, Color: colorGray})),
	)

	clientName := "Client inconnu"
	clientSIREN := ""
	clientAddr := ""
	clientCity := ""
	if inv.Client != nil {
		clientName = inv.Client.Name
		if inv.Client.SIREN != "" {
			clientSIREN = "SIREN : " + inv.Client.SIREN
		} else if inv.Client.SIRET != "" {
			clientSIREN = "SIRET : " + inv.Client.SIRET
		}
		clientAddr = inv.Client.Address
		clientCity = inv.Client.PostalCode + " " + inv.Client.City
	}

	m.AddRow(7, col.New(12).Add(text.New(clientName, props.Text{Size: 11, Style: fontstyle.Bold})))
	if clientSIREN != "" {
		m.AddRow(5, col.New(12).Add(text.New(clientSIREN, txtSmall)))
	}
	if clientAddr != "" {
		m.AddRow(5, col.New(12).Add(text.New(clientAddr, txtSmall)))
	}
	if clientCity != "" {
		m.AddRow(5, col.New(12).Add(text.New(clientCity, txtSmall)))
	}

	m.AddRow(8) // espace

	// ── Tableau lignes : en-tête
	m.AddRows(
		row.New(8).Add(
			col.New(6).WithStyle(cellHeader).Add(text.New("DESCRIPTION", txtWhiteB)),
			col.New(2).WithStyle(cellHeader).Add(text.New("QTÉ", txtWhiteB)),
			col.New(2).WithStyle(cellHeader).Add(text.New("PRIX HT", txtWhiteB)),
			col.New(2).WithStyle(cellHeader).Add(text.New("TOTAL HT", txtWhiteB)),
		),
	)

	for i, line := range inv.Lines {
		cell := cellBorder
		if i%2 == 1 {
			cell = cellAlt
		}
		m.AddRows(
			row.New(7).Add(
				col.New(6).WithStyle(cell).Add(text.New(line.Description, txtNormal)),
				col.New(2).WithStyle(cell).Add(text.New(line.Quantity.StringFixed(2), txtCenter)),
				col.New(2).WithStyle(cell).Add(text.New(line.UnitPriceHT.StringFixed(2)+" €", txtRight)),
				col.New(2).WithStyle(cell).Add(text.New(line.TotalHT.StringFixed(2)+" €", txtRight)),
			),
		)
	}

	m.AddRow(4) // espace

	// ── Totaux
	m.AddRow(7,
		col.New(8),
		col.New(2).Add(text.New("Total HT", txtBold)),
		col.New(2).Add(text.New(inv.TotalHT.StringFixed(2)+" €", txtRightB)),
	)

	if inv.VATApplicable && inv.VATAmount.IsPositive() {
		m.AddRow(6,
			col.New(8),
			col.New(2).Add(text.New("TVA", txtNormal)),
			col.New(2).Add(text.New(inv.VATAmount.StringFixed(2)+" €", txtRight)),
		)
	}

	m.AddRows(
		row.New(8).Add(
			col.New(8),
			col.New(4).WithStyle(&props.Cell{
				BackgroundColor: colorBlue,
				BorderType:      border.Full,
				BorderColor:     colorBlue,
				BorderThickness: 0.3,
			}).Add(text.New("TOTAL TTC  "+inv.TotalTTC.StringFixed(2)+" €",
				props.Text{Size: 11, Style: fontstyle.Bold, Align: align.Center,
					Color: &props.Color{Red: 255, Green: 255, Blue: 255}})),
		),
	)

	m.AddRow(6) // espace

	// ── Mentions légales TVA
	if !inv.VATApplicable {
		m.AddRow(6, col.New(12).Add(text.New(
			"💡 "+domain.VATMentionExemption,
			props.Text{Size: 9, Style: fontstyle.Italic, Color: colorGray},
		)))
	}

	m.AddRow(4) // espace

	// ── Conditions de paiement
	// LEGAL: Délai, pénalités, indemnité 40€ — obligatoires (Art. L441-6 Code de Commerce)
	m.AddRow(6, col.New(12).Add(text.New("CONDITIONS DE PAIEMENT",
		props.Text{Size: 9, Style: fontstyle.Bold, Color: colorGray})))

	paymentLines := []string{
		"Délai de paiement : " + inv.PaymentDeadline,
		"Pénalités de retard : " + inv.LatePenaltyRate.StringFixed(2) + "% par an à compter du lendemain de la date d'échéance",
		"Indemnité forfaitaire pour frais de recouvrement : " + inv.RecoveryFee.StringFixed(0) + " € (Art. L441-6 Code de Commerce)",
	}
	for _, line := range paymentLines {
		m.AddRow(5, col.New(12).Add(text.New("• "+line, txtSmall)))
	}

	// ── Pied de page (footer)
	footer := seller.Name + "  |  SIRET : " + seller.SIRET
	if seller.VATNumber != "" {
		footer += "  |  TVA : " + seller.VATNumber
	}
	if err := m.RegisterFooter(
		row.New(6).Add(col.New(12).Add(text.New(footer,
			props.Text{Size: 8, Align: align.Center, Color: colorGray}))),
	); err != nil {
		_ = err // non-fatal
	}

	return m
}

// ─── Template Avoir ─────────────────────────────────────────────────────────

func buildCreditNotePDF(cn *domain.CreditNote, seller *appcfg.Seller) core.Maroto {
	cfg := config.NewBuilder().
		WithPageSize(pagesize.A4).
		WithLeftMargin(15).
		WithRightMargin(15).
		WithTopMargin(15).
		WithBottomMargin(20).
		Build()

	m := maroto.New(cfg)

	// ── En-tête
	m.AddRows(
		row.New(8).Add(
			col.New(7).Add(text.New(seller.Name, props.Text{Size: 14, Style: fontstyle.Bold})),
			col.New(5).Add(text.New("AVOIR", props.Text{Size: 16, Style: fontstyle.Bold, Align: align.Right, Color: colorDanger})),
		),
		row.New(5).Add(
			col.New(7).Add(text.New("SIRET : "+seller.SIRET, txtSmall)),
			col.New(5).Add(text.New("N° "+cn.Number, props.Text{Size: 12, Style: fontstyle.Bold, Align: align.Right, Color: colorDanger})),
		),
		row.New(5).Add(
			col.New(7).Add(text.New(seller.Address, txtSmall)),
			col.New(5).Add(text.New("Émis le : "+cn.IssueDate.Format("02/01/2006"), props.Text{Size: 9, Align: align.Right, Color: colorGray})),
		),
		row.New(5).Add(
			col.New(7).Add(text.New(seller.PostalCode+" "+seller.City, txtSmall)),
			col.New(5).Add(text.New("Réf. facture : "+cn.InvoiceReference, props.Text{Size: 9, Align: align.Right, Color: colorGray})),
		),
	)

	m.AddRow(6) // espace

	// LEGAL: Référence obligatoire à la facture d'origine (Art. 272 CGI)
	m.AddRow(7, col.New(12).Add(text.New(
		"Avoir en réponse à la facture "+cn.InvoiceReference+" — Motif : "+cn.Reason,
		props.Text{Size: 10, Style: fontstyle.Italic, Color: colorGray},
	)))

	m.AddRow(8) // espace

	// ── Lignes
	m.AddRows(
		row.New(8).Add(
			col.New(8).WithStyle(cellHeader).Add(text.New("DESCRIPTION", txtWhiteB)),
			col.New(2).WithStyle(cellHeader).Add(text.New("QTÉ", txtWhiteB)),
			col.New(2).WithStyle(cellHeader).Add(text.New("MONTANT HT", txtWhiteB)),
		),
	)

	if len(cn.Lines) == 0 {
		m.AddRows(
			row.New(7).Add(
				col.New(8).WithStyle(cellBorder).Add(text.New("Avoir — "+cn.Reason, txtNormal)),
				col.New(2).WithStyle(cellBorder).Add(text.New("1", txtCenter)),
				col.New(2).WithStyle(cellBorder).Add(text.New(cn.TotalHT.StringFixed(2)+" €", txtRight)),
			),
		)
	}
	for i, line := range cn.Lines {
		cell := cellBorder
		if i%2 == 1 {
			cell = cellAlt
		}
		m.AddRows(
			row.New(7).Add(
				col.New(8).WithStyle(cell).Add(text.New(line.Description, txtNormal)),
				col.New(2).WithStyle(cell).Add(text.New(line.Quantity.StringFixed(2), txtCenter)),
				col.New(2).WithStyle(cell).Add(text.New(line.TotalHT.StringFixed(2)+" €", txtRight)),
			),
		)
	}

	m.AddRow(4)

	// ── Totaux
	m.AddRow(7,
		col.New(8),
		col.New(2).Add(text.New("Total HT", txtBold)),
		col.New(2).Add(text.New(cn.TotalHT.StringFixed(2)+" €", txtRightB)),
	)
	m.AddRows(
		row.New(8).Add(
			col.New(8),
			col.New(4).WithStyle(&props.Cell{
				BackgroundColor: colorDanger,
				BorderType:      border.Full,
				BorderColor:     colorDanger,
				BorderThickness: 0.3,
			}).Add(text.New("TOTAL TTC  "+cn.TotalTTC.StringFixed(2)+" €",
				props.Text{Size: 11, Style: fontstyle.Bold, Align: align.Center,
					Color: &props.Color{Red: 255, Green: 255, Blue: 255}})),
		),
	)

	// Pied de page
	footer := seller.Name + "  |  SIRET : " + seller.SIRET
	if err := m.RegisterFooter(
		row.New(6).Add(col.New(12).Add(text.New(footer,
			props.Text{Size: 8, Align: align.Center, Color: colorGray}))),
	); err != nil {
		_ = err
	}

	return m
}

// ─── Template Devis ──────────────────────────────────────────────────────────

func buildQuotePDF(q *domain.Quote, seller *appcfg.Seller) core.Maroto {
	cfg := config.NewBuilder().
		WithPageSize(pagesize.A4).
		WithLeftMargin(15).
		WithRightMargin(15).
		WithTopMargin(15).
		WithBottomMargin(20).
		Build()

	m := maroto.New(cfg)

	m.AddRows(
		row.New(8).Add(
			col.New(7).Add(text.New(seller.Name, props.Text{Size: 14, Style: fontstyle.Bold})),
			col.New(5).Add(text.New("DEVIS", props.Text{Size: 16, Style: fontstyle.Bold, Align: align.Right,
				Color: &props.Color{Red: 139, Green: 92, Blue: 246}})),
		),
		row.New(5).Add(
			col.New(7).Add(text.New("SIRET : "+seller.SIRET, txtSmall)),
			col.New(5).Add(text.New("N° "+q.Number, props.Text{Size: 12, Style: fontstyle.Bold, Align: align.Right,
				Color: &props.Color{Red: 139, Green: 92, Blue: 246}})),
		),
		row.New(5).Add(
			col.New(7).Add(text.New(seller.Address, txtSmall)),
			col.New(5).Add(text.New("Émis le : "+q.IssueDate.Format("02/01/2006"), props.Text{Size: 9, Align: align.Right, Color: colorGray})),
		),
		row.New(5).Add(
			col.New(7).Add(text.New(seller.PostalCode+" "+seller.City, txtSmall)),
			col.New(5).Add(text.New("Valable jusqu'au : "+q.ExpiryDate.Format("02/01/2006"),
				props.Text{Size: 9, Align: align.Right, Color: colorGray})),
		),
	)

	m.AddRow(8)

	if q.Notes != "" {
		m.AddRow(6, col.New(12).Add(text.New(q.Notes, props.Text{Size: 9, Style: fontstyle.Italic, Color: colorGray})))
		m.AddRow(4)
	}

	// Lignes
	m.AddRows(
		row.New(8).Add(
			col.New(6).WithStyle(cellHeader).Add(text.New("DESCRIPTION", txtWhiteB)),
			col.New(2).WithStyle(cellHeader).Add(text.New("QTÉ", txtWhiteB)),
			col.New(2).WithStyle(cellHeader).Add(text.New("PRIX HT", txtWhiteB)),
			col.New(2).WithStyle(cellHeader).Add(text.New("TOTAL HT", txtWhiteB)),
		),
	)
	for i, line := range q.Lines {
		cell := cellBorder
		if i%2 == 1 {
			cell = cellAlt
		}
		m.AddRows(
			row.New(7).Add(
				col.New(6).WithStyle(cell).Add(text.New(line.Description, txtNormal)),
				col.New(2).WithStyle(cell).Add(text.New(line.Quantity.StringFixed(2), txtCenter)),
				col.New(2).WithStyle(cell).Add(text.New(line.UnitPriceHT.StringFixed(2)+" €", txtRight)),
				col.New(2).WithStyle(cell).Add(text.New(line.TotalHT.StringFixed(2)+" €", txtRight)),
			),
		)
	}

	m.AddRow(4)
	m.AddRow(7,
		col.New(8),
		col.New(2).Add(text.New("Total HT", txtBold)),
		col.New(2).Add(text.New(q.TotalHT.StringFixed(2)+" €", txtRightB)),
	)
	m.AddRow(7,
		col.New(8),
		col.New(2).Add(text.New("Total TTC", props.Text{Size: 11, Style: fontstyle.Bold})),
		col.New(2).Add(text.New(q.TotalTTC.StringFixed(2)+" €", props.Text{Size: 11, Style: fontstyle.Bold, Align: align.Right})),
	)

	m.AddRow(8)
	m.AddRow(6, col.New(12).Add(text.New(
		"Devis valable jusqu'au "+q.ExpiryDate.Format("02/01/2006")+". Pour accepter, merci de nous retourner ce document signé.",
		props.Text{Size: 9, Style: fontstyle.Italic, Color: colorGray},
	)))

	footer := seller.Name + "  |  SIRET : " + seller.SIRET
	if err := m.RegisterFooter(
		row.New(6).Add(col.New(12).Add(text.New(footer,
			props.Text{Size: 8, Align: align.Center, Color: colorGray}))),
	); err != nil {
		_ = err
	}

	return m
}
