package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kvitrvn/runar/internal/config"
	"github.com/kvitrvn/runar/internal/domain"
	"github.com/kvitrvn/runar/internal/service"
	"github.com/kvitrvn/runar/internal/tui/components"
	"github.com/kvitrvn/runar/internal/tui/styles"
	"github.com/shopspring/decimal"
)

// QuoteMode représente le sous-mode de la vue devis.
type QuoteMode int

const (
	QuoteModeList           QuoteMode = iota
	QuoteModeForm                     // formulaire création
	QuoteModeDetail                   // vue détail
	QuoteModeConfirmConvert           // confirmation conversion en facture
)

// Messages internes.
type QuotesLoadedMsg struct{ Quotes []domain.Quote; Err error }
type QuoteSavedMsg struct{ Err error }
type QuotePDFMsg struct{ Path string; Err error }
type QuoteStateChangedMsg struct{ Err error }
type QuoteConvertedMsg struct{ InvoiceNumber string; Err error }

// ─── QuoteForm (formulaire multi-étapes) ─────────────────────────────────────

type quoteFormStep int

const (
	quoteStepBasic quoteFormStep = iota
	quoteStepLines
)

type quoteLineEntry struct {
	description string
	quantity    string
	unitPrice   string
}

type quoteForm struct {
	step        quoteFormStep
	basicForm   *components.Form
	lineForm    *components.Form
	lines       []quoteLineEntry
	focusedLine int
	cfg         *config.Config
}

func newQuoteForm(cfg *config.Config, width int) *quoteForm {
	basic := components.NewForm("NOUVEAU DEVIS — Informations", []components.FormField{
		components.NewField("Client ID", "1", true),
		components.NewField("Date émission", time.Now().Format("2006-01-02"), true),
		components.NewField("Date expiration", time.Now().AddDate(0, 0, 30).Format("2006-01-02"), true),
		components.NewField("Notes", "", false),
	}, width)
	return &quoteForm{step: quoteStepBasic, basicForm: basic, cfg: cfg}
}

func newQuoteLineForm(width int) *components.Form {
	return components.NewForm("LIGNE DE DEVIS", []components.FormField{
		components.NewField("Description", "Prestation de service", true),
		components.NewField("Quantité", "1", true),
		components.NewField("Prix unitaire HT €", "0.00", true),
	}, width)
}

func (f *quoteForm) buildQuote() (*domain.Quote, error) {
	bf := f.basicForm
	var clientID int
	fmt.Sscanf(bf.Value(0), "%d", &clientID)

	issueDate, err := time.Parse("2006-01-02", bf.Value(1))
	if err != nil {
		return nil, fmt.Errorf("date émission invalide (format attendu: AAAA-MM-JJ)")
	}
	expiryDate, err := time.Parse("2006-01-02", bf.Value(2))
	if err != nil {
		return nil, fmt.Errorf("date expiration invalide (format attendu: AAAA-MM-JJ)")
	}

	q := &domain.Quote{
		ClientID:   clientID,
		IssueDate:  issueDate,
		ExpiryDate: expiryDate,
		Notes:      bf.Value(3),
	}

	for i, le := range f.lines {
		qty, _ := decimal.NewFromString(le.quantity)
		price, _ := decimal.NewFromString(le.unitPrice)
		line := domain.QuoteLine{
			LineOrder:   i + 1,
			Description: le.description,
			Quantity:    qty,
			UnitPriceHT: price,
		}
		if f.cfg.VAT.Applicable {
			line.VATRate = decimal.NewFromFloat(f.cfg.VAT.DefaultRate)
		}
		line.Calculate()
		q.Lines = append(q.Lines, line)
	}

	return q, nil
}

// ─── QuotesView ───────────────────────────────────────────────────────────────

// QuotesView est la vue complète de gestion des devis.
type QuotesView struct {
	services *service.Services
	config   *config.Config
	mode     QuoteMode
	quotes   []domain.Quote
	filtered []domain.Quote
	table    components.TableModel
	form     *quoteForm
	selected *domain.Quote
	search   string
	err      string
	formErr  string
	width    int
	height   int
}

// NewQuotesView crée la vue devis.
func NewQuotesView(services *service.Services, cfg *config.Config, width, height int) QuotesView {
	cols := quoteColumns(width)
	t := components.NewTable(cols, nil, height-6)
	return QuotesView{
		services: services,
		config:   cfg,
		table:    t,
		width:    width,
		height:   height,
	}
}

// Load déclenche le chargement des devis.
func (v QuotesView) Load(search string) tea.Cmd {
	svc := v.services.Quote
	return func() tea.Msg {
		quotes, err := svc.List(search)
		return QuotesLoadedMsg{Quotes: quotes, Err: err}
	}
}

// SetSize ajuste les dimensions.
func (v *QuotesView) SetSize(w, h int) {
	v.width = w
	v.height = h
}

// IsInputActive retourne true si un formulaire est actif.
func (v QuotesView) IsInputActive() bool {
	return v.mode == QuoteModeForm
}

// Update gère les messages.
func (v QuotesView) Update(msg tea.Msg) (QuotesView, tea.Cmd) {
	switch msg := msg.(type) {
	case QuotesLoadedMsg:
		if msg.Err != nil {
			v.err = msg.Err.Error()
		} else {
			v.quotes = msg.Quotes
			v.filtered = filterQuotes(msg.Quotes, v.search)
			v.table.SetRows(quoteRows(v.filtered))
			v.err = ""
		}

	case QuoteSavedMsg:
		if msg.Err != nil {
			v.formErr = msg.Err.Error()
		} else {
			v.mode = QuoteModeList
			v.form = nil
			v.formErr = ""
			return v, v.Load(v.search)
		}

	case QuotePDFMsg:
		if msg.Err != nil {
			v.err = msg.Err.Error()
		} else {
			v.err = ""
		}

	case QuoteStateChangedMsg:
		if msg.Err != nil {
			v.err = msg.Err.Error()
		} else {
			v.err = ""
			return v, v.Load(v.search)
		}

	case QuoteConvertedMsg:
		// Géré par l'App (toast + switch view)

	case tea.KeyMsg:
		switch v.mode {
		case QuoteModeList:
			return v.handleListKey(msg)
		case QuoteModeForm:
			return v.handleFormKey(msg)
		case QuoteModeDetail:
			return v.handleDetailKey(msg)
		case QuoteModeConfirmConvert:
			return v.handleConfirmConvertKey(msg)
		}
	}

	if v.mode == QuoteModeList {
		updated, cmd := v.table.Update(msg)
		v.table = updated
		return v, cmd
	}
	return v, nil
}

func (v QuotesView) handleListKey(msg tea.KeyMsg) (QuotesView, tea.Cmd) {
	switch msg.String() {
	case "n":
		v.mode = QuoteModeForm
		v.form = newQuoteForm(v.config, v.width)
		v.formErr = ""
	case "s":
		if sel := v.selectedQuote(); sel != nil {
			return v, v.changeState(sel.ID, "sent")
		}
	case "a":
		if sel := v.selectedQuote(); sel != nil {
			return v, v.changeState(sel.ID, "accepted")
		}
	case "r":
		if sel := v.selectedQuote(); sel != nil {
			return v, v.changeState(sel.ID, "refused")
		}
	case "f":
		sel := v.selectedQuote()
		if sel != nil {
			if sel.CanConvertToInvoice() {
				v.selected = sel
				v.mode = QuoteModeConfirmConvert
			} else {
				v.err = fmt.Sprintf("Seul un devis accepté peut être converti (état: %s)", sel.State)
			}
		}
	case "p":
		if sel := v.selectedQuote(); sel != nil {
			return v, v.generatePDF(sel.ID)
		}
	case "enter":
		if sel := v.selectedQuote(); sel != nil {
			v.selected = sel
			v.mode = QuoteModeDetail
		}
	default:
		updated, cmd := v.table.Update(msg)
		v.table = updated
		return v, cmd
	}
	return v, nil
}

func (v QuotesView) handleFormKey(msg tea.KeyMsg) (QuotesView, tea.Cmd) {
	if v.form == nil {
		return v, nil
	}

	if v.form.step == quoteStepBasic {
		event, cmd := v.form.basicForm.Update(msg)
		switch event {
		case components.FormEventCancel:
			v.mode = QuoteModeList
			v.form = nil
			v.formErr = ""
		case components.FormEventSubmit:
			v.form.step = quoteStepLines
			v.form.lineForm = newQuoteLineForm(v.width)
			v.formErr = ""
		}
		return v, cmd
	}

	// Étape lignes
	if v.form.lineForm != nil {
		event, cmd := v.form.lineForm.Update(msg)
		switch event {
		case components.FormEventCancel:
			if len(v.form.lines) == 0 {
				v.form.step = quoteStepBasic
				v.form.lineForm = nil
			} else {
				v.form.lineForm = nil
			}
		case components.FormEventSubmit:
			lf := v.form.lineForm
			v.form.lines = append(v.form.lines, quoteLineEntry{
				description: lf.Value(0),
				quantity:    lf.Value(1),
				unitPrice:   lf.Value(2),
			})
			v.form.lineForm = nil
			v.formErr = ""
		}
		return v, cmd
	}

	// Navigation liste des lignes
	switch msg.String() {
	case "a":
		v.form.lineForm = newQuoteLineForm(v.width)
	case "d":
		if len(v.form.lines) > 0 && v.form.focusedLine < len(v.form.lines) {
			v.form.lines = append(v.form.lines[:v.form.focusedLine], v.form.lines[v.form.focusedLine+1:]...)
			if v.form.focusedLine >= len(v.form.lines) && v.form.focusedLine > 0 {
				v.form.focusedLine--
			}
		}
	case "j", "down":
		if v.form.focusedLine < len(v.form.lines)-1 {
			v.form.focusedLine++
		}
	case "k", "up":
		if v.form.focusedLine > 0 {
			v.form.focusedLine--
		}
	case "enter":
		if len(v.form.lines) > 0 {
			return v, v.saveQuote()
		}
		v.formErr = "Ajoutez au moins une ligne (touche 'a')"
	case "esc":
		v.form.step = quoteStepBasic
		v.form.lineForm = nil
	}
	return v, nil
}

func (v QuotesView) handleDetailKey(msg tea.KeyMsg) (QuotesView, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		v.mode = QuoteModeList
		v.selected = nil
		v.err = ""
	case "s":
		if v.selected != nil {
			return v, v.changeState(v.selected.ID, "sent")
		}
	case "a":
		if v.selected != nil {
			return v, v.changeState(v.selected.ID, "accepted")
		}
	case "r":
		if v.selected != nil {
			return v, v.changeState(v.selected.ID, "refused")
		}
	case "f":
		if v.selected != nil {
			if v.selected.CanConvertToInvoice() {
				v.mode = QuoteModeConfirmConvert
			} else {
				v.err = fmt.Sprintf("Seul un devis accepté peut être converti (état: %s)", v.selected.State)
			}
		}
	case "p":
		if v.selected != nil {
			return v, v.generatePDF(v.selected.ID)
		}
	}
	return v, nil
}

func (v QuotesView) handleConfirmConvertKey(msg tea.KeyMsg) (QuotesView, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if v.selected != nil {
			return v, v.convertToInvoice(v.selected.ID)
		}
	case "n", "N", "esc":
		if v.selected != nil {
			v.mode = QuoteModeDetail
		} else {
			v.mode = QuoteModeList
		}
	}
	return v, nil
}

// ─── Commands ─────────────────────────────────────────────────────────────────

func (v QuotesView) saveQuote() tea.Cmd {
	f := v.form
	svc := v.services.Quote
	return func() tea.Msg {
		q, err := f.buildQuote()
		if err != nil {
			return QuoteSavedMsg{Err: err}
		}
		return QuoteSavedMsg{Err: svc.Create(q)}
	}
}

func (v QuotesView) changeState(id int, target string) tea.Cmd {
	svc := v.services.Quote
	return func() tea.Msg {
		var err error
		switch target {
		case "sent":
			err = svc.MarkAsSent(id)
		case "accepted":
			err = svc.MarkAsAccepted(id)
		case "refused":
			err = svc.MarkAsRefused(id)
		}
		return QuoteStateChangedMsg{Err: err}
	}
}

func (v QuotesView) generatePDF(id int) tea.Cmd {
	svc := v.services.Quote
	return func() tea.Msg {
		path, err := svc.GeneratePDF(id)
		return QuotePDFMsg{Path: path, Err: err}
	}
}

func (v QuotesView) convertToInvoice(quoteID int) tea.Cmd {
	quoteSvc := v.services.Quote
	invoiceSvc := v.services.Invoice
	return func() tea.Msg {
		inv, err := quoteSvc.PrepareInvoiceFromQuote(quoteID)
		if err != nil {
			return QuoteConvertedMsg{Err: err}
		}
		if err := invoiceSvc.Create(inv); err != nil {
			return QuoteConvertedMsg{Err: err}
		}
		return QuoteConvertedMsg{InvoiceNumber: inv.Number}
	}
}

func (v QuotesView) selectedQuote() *domain.Quote {
	row := v.table.SelectedRow()
	if row == nil {
		return nil
	}
	var id int
	fmt.Sscanf(row[0], "%d", &id)
	for i := range v.filtered {
		if v.filtered[i].ID == id {
			return &v.filtered[i]
		}
	}
	return nil
}

// ─── Rendu ────────────────────────────────────────────────────────────────────

// View rend la vue devis.
func (v QuotesView) View() string {
	switch v.mode {
	case QuoteModeForm:
		return v.renderForm()
	case QuoteModeDetail:
		return v.renderDetail()
	case QuoteModeConfirmConvert:
		return v.renderConfirmConvert()
	default:
		return v.renderList()
	}
}

func (v QuotesView) renderList() string {
	var sb strings.Builder
	if v.err != "" {
		sb.WriteString(styles.StyleDanger.Render("⚠ "+v.err) + "\n\n")
	}
	sb.WriteString(styles.StyleMuted.Render(fmt.Sprintf("  %d devis", len(v.filtered))) + "\n\n")
	sb.WriteString(v.table.View() + "\n")
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(
		"  n: Nouveau  s: Envoyer  a: Accepter  r: Refuser  f: → Facture  p: PDF  Enter: Détail",
	)
	sb.WriteString(hint)
	return sb.String()
}

func (v QuotesView) renderForm() string {
	if v.form == nil {
		return ""
	}
	var sb strings.Builder
	if v.formErr != "" {
		sb.WriteString(styles.StyleDanger.Render("⚠ "+v.formErr) + "\n\n")
	}

	if v.form.step == quoteStepBasic {
		sb.WriteString(v.form.basicForm.View())
		return sb.String()
	}

	// Étape lignes
	sb.WriteString(styles.StyleTitle.Render("NOUVEAU DEVIS — Lignes") + "\n\n")

	if len(v.form.lines) == 0 {
		sb.WriteString(styles.StyleMuted.Render("  Aucune ligne. Appuyez sur 'a' pour ajouter.\n\n"))
	} else {
		headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Bold(true)
		sb.WriteString(headerStyle.Render(fmt.Sprintf("  %-40s  %8s  %12s  %12s\n",
			"DESCRIPTION", "QTÉ", "PRIX HT", "TOTAL HT")))
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#374151")).
			Render("  " + strings.Repeat("─", 76) + "\n"))

		for i, line := range v.form.lines {
			qty, _ := decimal.NewFromString(line.quantity)
			price, _ := decimal.NewFromString(line.unitPrice)
			total := qty.Mul(price).Round(2)
			lineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#D1D5DB"))
			prefix := "  "
			if i == v.form.focusedLine {
				lineStyle = lineStyle.Background(lipgloss.Color("#1E3A5F"))
				prefix = "> "
			}
			sb.WriteString(lineStyle.Render(fmt.Sprintf("%s%-40s  %8s  %12s€  %12s€\n",
				prefix,
				truncate(line.description, 40),
				line.quantity,
				price.StringFixed(2),
				total.StringFixed(2),
			)))
		}

		// Total
		q, err := v.form.buildQuote()
		if err == nil {
			q.CalculateTotals()
			sb.WriteString("\n")
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#374151")).
				Render("  " + strings.Repeat("─", 76) + "\n"))
			totalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8B5CF6")).Bold(true)
			sb.WriteString(totalStyle.Render(fmt.Sprintf("  %-52s  %12s€\n", "TOTAL HT", q.TotalHT.StringFixed(2))))
			if v.form.cfg.VAT.Applicable {
				sb.WriteString(totalStyle.Render(fmt.Sprintf("  %-52s  %12s€\n", "TVA", q.VATAmount.StringFixed(2))))
			}
			sb.WriteString(totalStyle.Render(fmt.Sprintf("  %-52s  %12s€\n", "TOTAL TTC", q.TotalTTC.StringFixed(2))))
		}
	}

	sb.WriteString("\n")
	if v.form.lineForm != nil {
		sb.WriteString(v.form.lineForm.View())
	} else {
		hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(
			"  a: Ajouter ligne  d: Supprimer ligne  j/k: Sélectionner  Enter: Valider  Esc: Retour",
		)
		sb.WriteString(hint)
	}
	return sb.String()
}

func (v QuotesView) renderDetail() string {
	if v.selected == nil {
		return ""
	}
	q := v.selected
	isExpired := q.IsExpired() && q.State != domain.QuoteStateAccepted && q.State != domain.QuoteStateRefused

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#374151")).
		Padding(1, 2).
		Width(v.width - 4)

	label := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Width(20)
	value := lipgloss.NewStyle().Foreground(lipgloss.Color("#F9FAFB"))
	row := func(l, val string) string { return label.Render(l) + value.Render(val) + "\n" }

	var content strings.Builder
	content.WriteString(styles.StyleTitle.Render("DEVIS "+q.Number) + "\n\n")

	if v.err != "" {
		content.WriteString(styles.StyleDanger.Render("⚠ "+v.err) + "\n\n")
	}
	if isExpired {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Bold(true).
			Render("⚠  Devis EXPIRÉ") + "\n\n")
	}

	content.WriteString(row("Numéro", q.Number))
	content.WriteString(row("État", renderQuoteState(q.State)))
	content.WriteString(row("Date émission", q.IssueDate.Format("02/01/2006")))
	content.WriteString(row("Date expiration", q.ExpiryDate.Format("02/01/2006")))
	if q.Notes != "" {
		content.WriteString(row("Notes", q.Notes))
	}
	content.WriteString("\n")

	if len(q.Lines) > 0 {
		hdr := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Bold(true)
		content.WriteString(hdr.Render(fmt.Sprintf("  %-38s  %8s  %12s\n", "DESCRIPTION", "QTÉ", "TOTAL HT")))
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#374151")).
			Render("  " + strings.Repeat("─", 62) + "\n"))
		for _, line := range q.Lines {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#D1D5DB")).
				Render(fmt.Sprintf("  %-38s  %8s  %12s€\n",
					truncate(line.Description, 38),
					line.Quantity.StringFixed(2),
					line.TotalHT.StringFixed(2),
				)))
		}
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#374151")).
			Render("  " + strings.Repeat("─", 62) + "\n"))
	}

	totalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8B5CF6")).Bold(true)
	content.WriteString(totalStyle.Render(fmt.Sprintf("  %-50s %12s€\n", "TOTAL HT", q.TotalHT.StringFixed(2))))
	content.WriteString(totalStyle.Render(fmt.Sprintf("  %-50s %12s€\n", "TOTAL TTC", q.TotalTTC.StringFixed(2))))
	if q.PDFPath != "" {
		content.WriteString("\n" + styles.StyleSuccess.Render("  PDF : "+q.PDFPath) + "\n")
	}

	content.WriteString("\n")
	var actions []string
	if q.State == domain.QuoteStateDraft {
		actions = append(actions, "s: Envoyer")
	}
	if q.State == domain.QuoteStateDraft || q.State == domain.QuoteStateSent {
		actions = append(actions, "a: Accepter", "r: Refuser")
	}
	if q.CanConvertToInvoice() {
		actions = append(actions, "f: → Facture")
	}
	actions = append(actions, "p: PDF", "Esc: Retour")
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).
		Render("  " + strings.Join(actions, "  ")))

	return "\n" + box.Render(content.String())
}

func (v QuotesView) renderConfirmConvert() string {
	if v.selected == nil {
		return ""
	}
	q := v.selected
	msg := fmt.Sprintf("Convertir le devis %s en FACTURE ?\n\n"+
		"  Client ID  : %d\n"+
		"  Total HT   : %s€\n"+
		"  Total TTC  : %s€\n\n"+
		"  Une nouvelle facture brouillon sera créée.",
		q.Number, q.ClientID, q.TotalHT.StringFixed(2), q.TotalTTC.StringFixed(2))
	return "\n" + lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8B5CF6")).
		Bold(true).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8B5CF6")).
		Render(msg) +
		"\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).
			Render("  [Y] Confirmer  [N/Esc] Annuler")
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func quoteColumns(width int) []table.Column {
	return []table.Column{
		{Title: "ID", Width: 5},
		{Title: "NUMÉRO", Width: 14},
		{Title: "TOTAL TTC", Width: 12},
		{Title: "ÉTAT", Width: 12},
		{Title: "ÉMISSION", Width: 12},
		{Title: "EXPIRATION", Width: 12},
	}
}

func quoteRows(quotes []domain.Quote) []table.Row {
	rows := make([]table.Row, len(quotes))
	for i, q := range quotes {
		expiry := q.ExpiryDate.Format("02/01/2006")
		if q.IsExpired() && q.State != domain.QuoteStateAccepted && q.State != domain.QuoteStateRefused {
			expiry = "! " + expiry
		}
		rows[i] = table.Row{
			fmt.Sprint(q.ID),
			q.Number,
			q.TotalTTC.StringFixed(2) + "€",
			string(q.State),
			q.IssueDate.Format("02/01/2006"),
			expiry,
		}
	}
	return rows
}

func filterQuotes(quotes []domain.Quote, search string) []domain.Quote {
	if search == "" {
		return quotes
	}
	s := strings.ToLower(search)
	var out []domain.Quote
	for _, q := range quotes {
		if strings.Contains(strings.ToLower(q.Number), s) ||
			strings.Contains(strings.ToLower(string(q.State)), s) {
			out = append(out, q)
		}
	}
	return out
}

func renderQuoteState(state domain.QuoteState) string {
	switch state {
	case domain.QuoteStateDraft:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("brouillon")
	case domain.QuoteStateSent:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#0EA5E9")).Render("envoyé")
	case domain.QuoteStateAccepted:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Bold(true).Render("accepté")
	case domain.QuoteStateRefused:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render("refusé")
	case domain.QuoteStateExpired:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Render("expiré")
	default:
		return string(state)
	}
}
