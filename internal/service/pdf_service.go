package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"
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

	m := buildInvoicePDF(invoice, s.cfg)

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

	m := buildCreditNotePDF(cn, s.cfg)

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

	m := buildQuotePDF(quote, s.cfg)

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
	colorPurple = &props.Color{Red: 139, Green: 92, Blue: 246}

	cellHeader = &props.Cell{
		BackgroundColor: colorBlue,
		BorderColor:     colorBlue,
		BorderType:      border.Full,
		BorderThickness: 0.3,
	}
	cellHeaderDanger = &props.Cell{
		BackgroundColor: colorDanger,
		BorderColor:     colorDanger,
		BorderType:      border.Full,
		BorderThickness: 0.3,
	}
	cellHeaderPurple = &props.Cell{
		BackgroundColor: colorPurple,
		BorderColor:     colorPurple,
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
	cellTotalLine = &props.Cell{
		BorderColor:     colorBlack,
		BorderType:      border.Top,
		BorderThickness: 0.5,
	}

	txtNormal = props.Text{Size: 10, Color: colorBlack}
	txtSmall  = props.Text{Size: 9, Color: colorGray}
	txtBold   = props.Text{Size: 10, Style: fontstyle.Bold, Color: colorBlack}
	txtRight  = props.Text{Size: 10, Align: align.Right, Color: colorBlack}
	txtRightB = props.Text{Size: 12, Style: fontstyle.Bold, Align: align.Right, Color: colorBlack}
	txtCenter = props.Text{Size: 9, Align: align.Center, Color: colorBlack}
	txtWhiteB = props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Center,
		Color: &props.Color{Red: 255, Green: 255, Blue: 255}}
)

// ─── Footer commun ──────────────────────────────────────────────────────────

// registerFooter enregistre le footer multi-ligne commun à tous les documents.
// LEGAL: Mentions légales obligatoires en pied de page (Art. L441-6 Code de Commerce).
func registerFooter(m core.Maroto, cfg *appcfg.Config, isInvoice bool) {
	seller := &cfg.Seller

	// Ligne 1 : identité vendeur
	line1 := seller.Name + "  |  SIRET : " + seller.SIRET
	if seller.VATNumber != "" {
		line1 += "  |  N° TVA : " + seller.VATNumber
	}

	rows := []core.Row{
		row.New(5).Add(col.New(12).Add(text.New(line1,
			props.Text{Size: 8, Align: align.Center, Color: colorGray}))),
	}

	// Ligne 2 : mention TVA non applicable (si franchise en base)
	if !cfg.VAT.Applicable {
		rows = append(rows, row.New(5).Add(col.New(12).Add(text.New(
			domain.VATMentionExemption,
			props.Text{Size: 8, Style: fontstyle.Italic, Align: align.Center, Color: colorGray},
		))))
	}

	// Ligne 3 : conditions de paiement (factures uniquement)
	// LEGAL: Délai, pénalités, indemnité 40€ — obligatoires (Art. L441-6 Code de Commerce)
	if isInvoice {
		penaltyLine := fmt.Sprintf(
			"Délai : %s  —  Pénalités : %.2f%% l'an  —  Indemnité forfaitaire : 40 € (Art. L441-6 C. comm.)",
			cfg.Payment.DefaultDeadline,
			cfg.Payment.LatePenaltyRate,
		)
		rows = append(rows, row.New(5).Add(col.New(12).Add(text.New(penaltyLine,
			props.Text{Size: 7, Align: align.Center, Color: colorGray}))))
	}

	if err := m.RegisterFooter(rows...); err != nil {
		_ = err // non-fatal
	}
}

// ─── Bloc client partagé ────────────────────────────────────────────────────

// addClientBlock ajoute le bloc destinataire (FACTURER À / ADRESSÉ À) commun aux 3 templates.
func addClientBlock(m core.Maroto, client *domain.Client, label string) {
	m.AddRow(6, col.New(12).Add(text.New(label,
		props.Text{Size: 9, Style: fontstyle.Bold, Color: colorGray})))

	clientName := "Client non renseigné"
	if client != nil {
		clientName = client.Name
	}
	m.AddRow(7, col.New(12).Add(text.New(clientName,
		props.Text{Size: 11, Style: fontstyle.Bold})))

	if client != nil {
		if client.SIREN != "" {
			m.AddRow(5, col.New(12).Add(text.New("SIREN : "+client.SIREN, txtSmall)))
		} else if client.SIRET != "" {
			m.AddRow(5, col.New(12).Add(text.New("SIRET : "+client.SIRET, txtSmall)))
		}
		if client.Address != "" {
			m.AddRow(5, col.New(12).Add(text.New(client.Address, txtSmall)))
		}
		if client.PostalCode != "" || client.City != "" {
			m.AddRow(5, col.New(12).Add(text.New(client.PostalCode+" "+client.City, txtSmall)))
		}
	}
}

// ─── En-tête vendeur partagé ────────────────────────────────────────────────

func addSellerHeader(m core.Maroto, seller *appcfg.Seller, titleText string, titleColor *props.Color,
	numberText string, date1Label, date1Value, date2Label, date2Value string) {
	m.AddRows(
		row.New(8).Add(
			col.New(7).Add(text.New(seller.Name, props.Text{Size: 14, Style: fontstyle.Bold})),
			col.New(5).Add(text.New(titleText, props.Text{Size: 16, Style: fontstyle.Bold, Align: align.Right, Color: titleColor})),
		),
		row.New(5).Add(
			col.New(7).Add(text.New("SIRET : "+seller.SIRET, txtSmall)),
			col.New(5).Add(text.New(numberText, props.Text{Size: 12, Style: fontstyle.Bold, Align: align.Right, Color: titleColor})),
		),
		row.New(5).Add(
			col.New(7).Add(text.New(seller.Address, txtSmall)),
			col.New(5).Add(text.New(date1Label+date1Value, props.Text{Size: 9, Align: align.Right, Color: colorGray})),
		),
		row.New(5).Add(
			col.New(7).Add(text.New(seller.PostalCode+" "+seller.City, txtSmall)),
			col.New(5).Add(text.New(date2Label+date2Value, props.Text{Size: 9, Align: align.Right, Color: colorGray})),
		),
	)
	if seller.Email != "" {
		m.AddRow(5, col.New(7).Add(text.New(seller.Email, txtSmall)), col.New(5))
	}
	if seller.Phone != "" {
		m.AddRow(5, col.New(7).Add(text.New(seller.Phone, txtSmall)), col.New(5))
	}
}

// ─── Tableau lignes partagé ─────────────────────────────────────────────────

// addLinesTable ajoute l'en-tête du tableau et les lignes de facturation.
// Layout 12 cols : DESCRIPTION(7) | QTÉ(1) | P.U. HT(2) | TOTAL HT(2)
func addLinesTable(m core.Maroto, hdrCell *props.Cell,
	lines []struct{ desc, qty, pu, total string }) {
	m.AddRow(8,
		col.New(7).WithStyle(hdrCell).Add(text.New("DESCRIPTION", txtWhiteB)),
		col.New(1).WithStyle(hdrCell).Add(text.New("QTÉ", txtWhiteB)),
		col.New(2).WithStyle(hdrCell).Add(text.New("P.U. HT", txtWhiteB)),
		col.New(2).WithStyle(hdrCell).Add(text.New("TOTAL HT", txtWhiteB)),
	)
	for i, l := range lines {
		cell := cellBorder
		if i%2 == 1 {
			cell = cellAlt
		}
		m.AddRow(7,
			col.New(7).WithStyle(cell).Add(text.New(l.desc, txtNormal)),
			col.New(1).WithStyle(cell).Add(text.New(l.qty, txtCenter)),
			col.New(2).WithStyle(cell).Add(text.New(l.pu, txtRight)),
			col.New(2).WithStyle(cell).Add(text.New(l.total, txtRight)),
		)
	}
}

// ─── Template Facture ───────────────────────────────────────────────────────

func buildInvoicePDF(inv *domain.Invoice, cfg *appcfg.Config) core.Maroto {
	seller := &cfg.Seller

	marotoCfg := config.NewBuilder().
		WithPageSize(pagesize.A4).
		WithLeftMargin(15).
		WithRightMargin(15).
		WithTopMargin(15).
		WithBottomMargin(25).
		Build()

	m := maroto.New(marotoCfg)

	// ── En-tête vendeur
	addSellerHeader(m, seller, "FACTURE", colorBlue,
		"N° "+inv.Number,
		"Émise le : ", inv.IssueDate.Format("02/01/2006"),
		"Échéance : ", inv.DueDate.Format("02/01/2006"),
	)

	if inv.DeliveryDate != (inv.IssueDate) {
		m.AddRow(5,
			col.New(7),
			col.New(5).Add(text.New("Livraison : "+inv.DeliveryDate.Format("02/01/2006"),
				props.Text{Size: 9, Align: align.Right, Color: colorGray})),
		)
	}

	m.AddRow(6) // espace
	addClientBlock(m, inv.Client, "FACTURER À :")
	m.AddRow(8) // espace

	// ── Tableau lignes
	lines := make([]struct{ desc, qty, pu, total string }, len(inv.Lines))
	for i, l := range inv.Lines {
		lines[i] = struct{ desc, qty, pu, total string }{
			desc:  l.Description,
			qty:   l.Quantity.StringFixed(2),
			pu:    l.UnitPriceHT.StringFixed(2) + " €",
			total: l.TotalHT.StringFixed(2) + " €",
		}
	}
	addLinesTable(m, cellHeader, lines)

	m.AddRow(4) // espace

	// ── Totaux
	m.AddRow(7,
		col.New(8),
		col.New(2).Add(text.New("Total HT", txtBold)),
		col.New(2).Add(text.New(inv.TotalHT.StringFixed(2)+" €", txtRight)),
	)

	if inv.VATApplicable && inv.VATAmount.IsPositive() {
		m.AddRow(6,
			col.New(8),
			col.New(2).Add(text.New("TVA", txtNormal)),
			col.New(2).Add(text.New(inv.VATAmount.StringFixed(2)+" €", txtRight)),
		)
	}

	// TOTAL TTC — sans fond coloré, bordure top
	m.AddRow(8,
		col.New(8),
		col.New(4).WithStyle(cellTotalLine).Add(
			text.New("TOTAL TTC   "+inv.TotalTTC.StringFixed(2)+" €", txtRightB),
		),
	)

	m.AddRow(6) // espace

	// ── Coordonnées bancaires (si IBAN configuré)
	if cfg.Payment.IBAN != "" {
		m.AddRow(6, col.New(12).Add(text.New("COORDONNÉES BANCAIRES",
			props.Text{Size: 9, Style: fontstyle.Bold, Color: colorGray})))
		m.AddRow(5, col.New(12).Add(text.New("IBAN : "+cfg.Payment.IBAN, txtSmall)))
		if cfg.Payment.BIC != "" {
			m.AddRow(5, col.New(12).Add(text.New("BIC  : "+cfg.Payment.BIC, txtSmall)))
		}
		if inv.PaymentRef != "" {
			m.AddRow(5, col.New(12).Add(text.New("Libellé virement : "+inv.PaymentRef, txtSmall)))
			m.AddRow(5, col.New(12).Add(text.New(
				"Merci d'utiliser ce libellé lors de votre virement bancaire.",
				props.Text{Size: 9, Style: fontstyle.Italic, Color: colorGray},
			)))
		}
		m.AddRow(4) // espace
	}

	registerFooter(m, cfg, true)
	return m
}

// ─── Template Avoir ─────────────────────────────────────────────────────────

func buildCreditNotePDF(cn *domain.CreditNote, cfg *appcfg.Config) core.Maroto {
	seller := &cfg.Seller

	marotoCfg := config.NewBuilder().
		WithPageSize(pagesize.A4).
		WithLeftMargin(15).
		WithRightMargin(15).
		WithTopMargin(15).
		WithBottomMargin(25).
		Build()

	m := maroto.New(marotoCfg)

	// ── En-tête vendeur
	addSellerHeader(m, seller, "AVOIR", colorDanger,
		"N° "+cn.Number,
		"Émis le : ", cn.IssueDate.Format("02/01/2006"),
		"Réf. facture : ", cn.InvoiceReference,
	)

	m.AddRow(6) // espace

	// LEGAL: Référence obligatoire à la facture d'origine (Art. 272 CGI)
	m.AddRow(7, col.New(12).Add(text.New(
		"Avoir en réponse à la facture "+cn.InvoiceReference+" — Motif : "+cn.Reason,
		props.Text{Size: 10, Style: fontstyle.Italic, Color: colorGray},
	)))

	m.AddRow(6) // espace
	addClientBlock(m, cn.Client, "ADRESSÉ À :")
	m.AddRow(8) // espace

	// ── Tableau lignes
	var lines []struct{ desc, qty, pu, total string }
	if len(cn.Lines) == 0 {
		lines = []struct{ desc, qty, pu, total string }{
			{desc: "Avoir — " + cn.Reason, qty: "1", pu: "", total: cn.TotalHT.StringFixed(2) + " €"},
		}
	} else {
		lines = make([]struct{ desc, qty, pu, total string }, len(cn.Lines))
		for i, l := range cn.Lines {
			lines[i] = struct{ desc, qty, pu, total string }{
				desc:  l.Description,
				qty:   l.Quantity.StringFixed(2),
				pu:    l.UnitPriceHT.StringFixed(2) + " €",
				total: l.TotalHT.StringFixed(2) + " €",
			}
		}
	}
	addLinesTable(m, cellHeaderDanger, lines)

	m.AddRow(4) // espace

	// ── Totaux
	m.AddRow(7,
		col.New(8),
		col.New(2).Add(text.New("Total HT", txtBold)),
		col.New(2).Add(text.New(cn.TotalHT.StringFixed(2)+" €", txtRight)),
	)

	// TOTAL TTC — sans fond coloré, bordure top
	m.AddRow(8,
		col.New(8),
		col.New(4).WithStyle(cellTotalLine).Add(
			text.New("TOTAL TTC   "+cn.TotalTTC.StringFixed(2)+" €", txtRightB),
		),
	)

	registerFooter(m, cfg, false)
	return m
}

// ─── Template Devis ──────────────────────────────────────────────────────────

func buildQuotePDF(q *domain.Quote, cfg *appcfg.Config) core.Maroto {
	seller := &cfg.Seller

	marotoCfg := config.NewBuilder().
		WithPageSize(pagesize.A4).
		WithLeftMargin(15).
		WithRightMargin(15).
		WithTopMargin(15).
		WithBottomMargin(25).
		Build()

	m := maroto.New(marotoCfg)

	// ── En-tête vendeur
	addSellerHeader(m, seller, "DEVIS", colorPurple,
		"N° "+q.Number,
		"Émis le : ", q.IssueDate.Format("02/01/2006"),
		"Valable jusqu'au : ", q.ExpiryDate.Format("02/01/2006"),
	)

	m.AddRow(6) // espace
	addClientBlock(m, q.Client, "ADRESSÉ À :")
	m.AddRow(8) // espace

	if q.Notes != "" {
		m.AddRow(6, col.New(12).Add(text.New(q.Notes,
			props.Text{Size: 9, Style: fontstyle.Italic, Color: colorGray})))
		m.AddRow(4)
	}

	// ── Tableau lignes
	lines := make([]struct{ desc, qty, pu, total string }, len(q.Lines))
	for i, l := range q.Lines {
		lines[i] = struct{ desc, qty, pu, total string }{
			desc:  l.Description,
			qty:   l.Quantity.StringFixed(2),
			pu:    l.UnitPriceHT.StringFixed(2) + " €",
			total: l.TotalHT.StringFixed(2) + " €",
		}
	}
	addLinesTable(m, cellHeaderPurple, lines)

	m.AddRow(4) // espace

	// ── Totaux
	m.AddRow(7,
		col.New(8),
		col.New(2).Add(text.New("Total HT", txtBold)),
		col.New(2).Add(text.New(q.TotalHT.StringFixed(2)+" €", txtRight)),
	)

	// TOTAL TTC — sans fond coloré, bordure top
	m.AddRow(8,
		col.New(8),
		col.New(4).WithStyle(cellTotalLine).Add(
			text.New("TOTAL TTC   "+q.TotalTTC.StringFixed(2)+" €", txtRightB),
		),
	)

	m.AddRow(8) // espace
	m.AddRow(6, col.New(12).Add(text.New(
		"Devis valable jusqu'au "+q.ExpiryDate.Format("02/01/2006")+". Pour accepter, merci de nous retourner ce document signé.",
		props.Text{Size: 9, Style: fontstyle.Italic, Color: colorGray},
	)))

	registerFooter(m, cfg, false)
	return m
}
