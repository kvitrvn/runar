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
	"github.com/kvitrvn/runar/internal/repository"
	"github.com/kvitrvn/runar/internal/service"
	"github.com/kvitrvn/runar/internal/tui/components"
	"github.com/kvitrvn/runar/internal/tui/styles"
	"github.com/shopspring/decimal"
)

// InvoiceMode représente le sous-mode de la vue factures.
type InvoiceMode int

const (
	InvoiceModeList          InvoiceMode = iota
	InvoiceModeForm                      // formulaire création/édition
	InvoiceModeDetail                    // vue détail
	InvoiceModeConfirmPaid               // confirmation paiement
	InvoiceModeConfirmDelete             // confirmation suppression brouillon
)

// Messages internes.
type InvoicesLoadedMsg struct{ Invoices []domain.Invoice; Err error }
type InvoiceSavedMsg struct{ Err error }
type InvoicePaidMsg struct{ Err error }
type InvoiceDeletedMsg struct{ Err error }
type InvoiceIssuedMsg struct{ Err error }
type InvoiceSentMsg struct{ Err error }
type InvoicePDFMsg struct{ Path string; Err error }
type InvoiceDetailLoadedMsg struct{ Invoice *domain.Invoice; Err error }
type OpenCreditNoteFormMsg struct {
	InvoiceID     int
	InvoiceNumber string
	TotalHT       string // pré-remplissage montant HT
	VATAmount     string // pré-remplissage TVA
}

// ─── InvoiceForm (formulaire multi-étapes) ───────────────────────────────────

// invoiceFormStep représente l'étape active du formulaire.
type invoiceFormStep int

const (
	invoiceStepBasic invoiceFormStep = iota // infos de base
	invoiceStepLines                        // lignes
)

// lineEntry est une ligne de facture en cours de saisie.
type lineEntry struct {
	description string
	quantity    string
	unitPrice   string
}

// invoiceForm gère la saisie d'une nouvelle facture.
type invoiceForm struct {
	step        invoiceFormStep
	basicForm   *components.Form
	lineForm    *components.Form
	lines       []lineEntry
	focusedLine int
	cfg         *config.Config
}

func newInvoiceForm(cfg *config.Config, width int) *invoiceForm {
	basic := components.NewForm("NOUVELLE FACTURE — Informations", []components.FormField{
		components.NewField("Client ID", "1", true),
		components.NewField("Date émission", time.Now().Format("2006-01-02"), true),
		components.NewField("Date échéance", time.Now().AddDate(0, 0, 30).Format("2006-01-02"), true),
		components.NewField("Date livraison", time.Now().Format("2006-01-02"), true),
		components.NewField("Notes", "", false),
	}, width)

	return &invoiceForm{
		step:      invoiceStepBasic,
		basicForm: basic,
		cfg:       cfg,
	}
}

func newLineForm(width int) *components.Form {
	return components.NewForm("LIGNE DE FACTURATION", []components.FormField{
		components.NewField("Description", "Prestation de service", true),
		components.NewField("Quantité", "1", true),
		components.NewField("Prix unitaire HT €", "0.00", true),
	}, width)
}

// buildInvoice construit une domain.Invoice depuis le formulaire.
func (f *invoiceForm) buildInvoice() (*domain.Invoice, error) {
	bf := f.basicForm
	var clientID int
	fmt.Sscanf(bf.Value(0), "%d", &clientID)

	issueDate, err := time.Parse("2006-01-02", bf.Value(1))
	if err != nil {
		return nil, fmt.Errorf("date émission invalide (format attendu: AAAA-MM-JJ)")
	}
	dueDate, err := time.Parse("2006-01-02", bf.Value(2))
	if err != nil {
		return nil, fmt.Errorf("date échéance invalide (format attendu: AAAA-MM-JJ)")
	}
	deliveryDate, err := time.Parse("2006-01-02", bf.Value(3))
	if err != nil {
		return nil, fmt.Errorf("date livraison invalide (format attendu: AAAA-MM-JJ)")
	}

	inv := &domain.Invoice{
		ClientID:         clientID,
		IssueDate:        issueDate,
		DueDate:          dueDate,
		DeliveryDate:     deliveryDate,
		State:            domain.InvoiceStateDraft,
		Notes:            bf.Value(4),
		VATApplicable:    f.cfg.VAT.Applicable,
		VATExemptionText: f.cfg.VAT.ExemptionText,
		PaymentDeadline:  f.cfg.Payment.DefaultDeadline,
		LatePenaltyRate:  decimal.NewFromFloat(f.cfg.Payment.LatePenaltyRate),
		RecoveryFee:      decimal.NewFromFloat(f.cfg.Payment.RecoveryFee),
	}

	for i, le := range f.lines {
		qty, _ := decimal.NewFromString(le.quantity)
		price, _ := decimal.NewFromString(le.unitPrice)
		line := domain.InvoiceLine{
			LineOrder:   i + 1,
			Description: le.description,
			Quantity:    qty,
			UnitPriceHT: price,
		}
		if f.cfg.VAT.Applicable {
			line.VATRate = decimal.NewFromFloat(f.cfg.VAT.DefaultRate)
		}
		line.Calculate()
		inv.Lines = append(inv.Lines, line)
	}

	return inv, nil
}

// ─── InvoicesView ────────────────────────────────────────────────────────────

// InvoicesView est la vue complète de gestion des factures.
type InvoicesView struct {
	services *service.Services
	config   *config.Config
	mode     InvoiceMode
	invoices []domain.Invoice
	filtered []domain.Invoice
	table    components.TableModel
	form     *invoiceForm
	selected *domain.Invoice
	search   string
	err      string
	formErr  string // erreur dans le formulaire
	width    int
	height   int
}

// NewInvoicesView crée la vue factures.
func NewInvoicesView(services *service.Services, cfg *config.Config, width, height int) InvoicesView {
	cols := invoiceColumns(width)
	t := components.NewTable(cols, nil, height-6)
	return InvoicesView{
		services: services,
		config:   cfg,
		table:    t,
		width:    width,
		height:   height,
	}
}

// Load déclenche le chargement des factures.
func (v InvoicesView) Load(search string) tea.Cmd {
	svc := v.services.Invoice
	year := time.Now().Year()
	return func() tea.Msg {
		invoices, err := svc.List(repository.InvoiceFilters{
			Year:   year,
			Search: search,
		})
		return InvoicesLoadedMsg{Invoices: invoices, Err: err}
	}
}

// SetSize ajuste les dimensions et recalcule les colonnes.
func (v *InvoicesView) SetSize(w, h int) {
	v.width = w
	v.height = h
	v.table.SetColumns(invoiceColumns(w))
	v.table.SetHeight(h - 6)
}

// IsInSubMode retourne true si la vue n'est pas en mode liste.
func (v InvoicesView) IsInSubMode() bool {
	return v.mode != InvoiceModeList
}

// IsInputActive retourne true si un formulaire est actif.
func (v InvoicesView) IsInputActive() bool {
	return v.mode == InvoiceModeForm
}

// Update gère les messages de la vue.
func (v InvoicesView) Update(msg tea.Msg) (InvoicesView, tea.Cmd) {
	switch msg := msg.(type) {
	case InvoicesLoadedMsg:
		if msg.Err != nil {
			v.err = msg.Err.Error()
		} else {
			v.invoices = msg.Invoices
			v.filtered = filterInvoices(msg.Invoices, v.search)
			v.table.SetRows(invoiceRows(v.filtered))
			v.err = ""
		}

	case InvoiceSavedMsg:
		if msg.Err != nil {
			v.formErr = msg.Err.Error()
		} else {
			v.mode = InvoiceModeList
			v.form = nil
			v.formErr = ""
			return v, v.Load(v.search)
		}

	case InvoicePaidMsg:
		if msg.Err != nil {
			v.err = msg.Err.Error()
			v.mode = InvoiceModeDetail
		} else {
			v.mode = InvoiceModeList
			v.selected = nil
			return v, v.Load(v.search)
		}

	case InvoiceDeletedMsg:
		if msg.Err != nil {
			v.err = msg.Err.Error()
			v.mode = InvoiceModeList
		} else {
			v.mode = InvoiceModeList
			v.selected = nil
			return v, v.Load(v.search)
		}

	case InvoiceIssuedMsg:
		if msg.Err != nil {
			v.err = msg.Err.Error()
		} else {
			v.err = ""
			return v, v.Load(v.search)
		}

	case InvoiceSentMsg:
		if msg.Err != nil {
			v.err = msg.Err.Error()
		} else {
			v.err = ""
			return v, v.Load(v.search)
		}

	case InvoicePDFMsg:
		if msg.Err != nil {
			v.err = msg.Err.Error()
		} else {
			v.err = ""
		}

	case InvoiceDetailLoadedMsg:
		if msg.Err != nil {
			v.err = msg.Err.Error()
		} else {
			v.selected = msg.Invoice
			v.mode = InvoiceModeDetail
			v.err = ""
		}

	case tea.KeyMsg:
		switch v.mode {
		case InvoiceModeList:
			return v.handleListKey(msg)
		case InvoiceModeForm:
			return v.handleFormKey(msg)
		case InvoiceModeDetail:
			return v.handleDetailKey(msg)
		case InvoiceModeConfirmPaid:
			return v.handleConfirmPaidKey(msg)
		case InvoiceModeConfirmDelete:
			return v.handleConfirmDeleteKey(msg)
		}
	}

	if v.mode == InvoiceModeList {
		updated, cmd := v.table.Update(msg)
		v.table = updated
		return v, cmd
	}

	return v, nil
}

func (v InvoicesView) handleListKey(msg tea.KeyMsg) (InvoicesView, tea.Cmd) {
	switch msg.String() {
	case "n":
		v.mode = InvoiceModeForm
		v.form = newInvoiceForm(v.config, v.width)
		v.formErr = ""
	case "i":
		sel := v.selectedInvoice()
		if sel != nil {
			return v, v.markAsIssued(sel.ID)
		}
	case "s":
		sel := v.selectedInvoice()
		if sel != nil {
			return v, v.markAsSent(sel.ID)
		}
	case "e":
		sel := v.selectedInvoice()
		if sel != nil && !sel.CanEdit() {
			v.err = styles.RenderImmutableError(sel.Number)
		}
	case "m":
		sel := v.selectedInvoice()
		if sel != nil {
			if sel.CanMarkAsPaid() {
				v.selected = sel
				v.mode = InvoiceModeConfirmPaid
			} else if sel.State == domain.InvoiceStatePaid {
				v.err = "Facture déjà payée"
			} else {
				v.err = fmt.Sprintf("Impossible : état %q (requis: émise ou envoyée)", sel.State)
			}
		}
	case "d":
		sel := v.selectedInvoice()
		if sel != nil {
			if sel.CanDelete() {
				v.selected = sel
				v.mode = InvoiceModeConfirmDelete
			} else {
				v.err = "Seuls les brouillons peuvent être supprimés"
			}
		}
	case "c":
		sel := v.selectedInvoice()
		if sel != nil {
			if sel.CanCancel() {
				return v, func() tea.Msg {
					return OpenCreditNoteFormMsg{
						InvoiceID:     sel.ID,
						InvoiceNumber: sel.Number,
						TotalHT:       sel.TotalHT.StringFixed(2),
						VATAmount:     sel.VATAmount.StringFixed(2),
					}
				}
			}
			v.err = fmt.Sprintf("Impossible de créer un avoir : facture %q (état: %s)", sel.Number, sel.State)
		}
	case "p":
		sel := v.selectedInvoice()
		if sel != nil {
			return v, v.generatePDF(sel.ID)
		}
	case "enter":
		sel := v.selectedInvoice()
		if sel != nil {
			svc := v.services.Invoice
			id := sel.ID
			return v, func() tea.Msg {
				inv, err := svc.GetByID(id)
				return InvoiceDetailLoadedMsg{Invoice: inv, Err: err}
			}
		}
	default:
		updated, cmd := v.table.Update(msg)
		v.table = updated
		return v, cmd
	}
	return v, nil
}

func (v InvoicesView) handleFormKey(msg tea.KeyMsg) (InvoicesView, tea.Cmd) {
	if v.form == nil {
		return v, nil
	}

	if v.form.step == invoiceStepBasic {
		event, cmd := v.form.basicForm.Update(msg)
		switch event {
		case components.FormEventCancel:
			v.mode = InvoiceModeList
			v.form = nil
			v.formErr = ""
		case components.FormEventSubmit:
			// Passer à l'étape lignes
			v.form.step = invoiceStepLines
			v.form.lineForm = newLineForm(v.width)
			v.formErr = ""
		}
		return v, cmd
	}

	// Étape lignes
	if v.form.lineForm != nil {
		// Saisie d'une ligne
		event, cmd := v.form.lineForm.Update(msg)
		switch event {
		case components.FormEventCancel:
			if len(v.form.lines) == 0 {
				// Retour à l'étape de base si pas de lignes
				v.form.step = invoiceStepBasic
				v.form.lineForm = nil
			} else {
				v.form.lineForm = nil
			}
		case components.FormEventSubmit:
			// Valider avant d'ajouter la ligne
			lf := v.form.lineForm
			desc := lf.Value(0)
			qtyStr := lf.Value(1)
			priceStr := lf.Value(2)
			qty, _ := decimal.NewFromString(qtyStr)
			if desc == "" {
				v.formErr = "La description est obligatoire"
				return v, nil
			}
			if qty.IsZero() || qty.IsNegative() {
				v.formErr = "La quantité doit être supérieure à 0"
				return v, nil
			}
			v.form.lines = append(v.form.lines, lineEntry{
				description: desc,
				quantity:    qtyStr,
				unitPrice:   priceStr,
			})
			v.form.lineForm = nil
			v.formErr = ""
		}
		return v, cmd
	}

	// Navigation dans la liste des lignes
	switch msg.String() {
	case "a":
		v.form.lineForm = newLineForm(v.width)
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
			// Soumettre la facture
			return v, v.saveInvoice()
		}
		v.formErr = "Ajoutez au moins une ligne (touche 'a')"
	case "esc":
		v.form.step = invoiceStepBasic
		v.form.lineForm = nil
	}
	return v, nil
}

func (v InvoicesView) handleDetailKey(msg tea.KeyMsg) (InvoicesView, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		v.mode = InvoiceModeList
		v.selected = nil
		v.err = ""
	case "i":
		if v.selected != nil {
			return v, v.markAsIssued(v.selected.ID)
		}
	case "s":
		if v.selected != nil {
			return v, v.markAsSent(v.selected.ID)
		}
	case "m":
		if v.selected != nil && v.selected.CanMarkAsPaid() {
			v.mode = InvoiceModeConfirmPaid
		}
	case "e":
		if v.selected != nil && !v.selected.CanEdit() {
			v.err = fmt.Sprintf("Facture %s immuable (état: %s)", v.selected.Number, v.selected.State)
		}
	case "c":
		if v.selected != nil {
			if v.selected.CanCancel() {
				sel := v.selected
				return v, func() tea.Msg {
					return OpenCreditNoteFormMsg{
						InvoiceID:     sel.ID,
						InvoiceNumber: sel.Number,
						TotalHT:       sel.TotalHT.StringFixed(2),
						VATAmount:     sel.VATAmount.StringFixed(2),
					}
				}
			}
			v.err = fmt.Sprintf("Impossible de créer un avoir : facture %q (état: %s)", v.selected.Number, v.selected.State)
		}
	case "p":
		if v.selected != nil {
			return v, v.generatePDF(v.selected.ID)
		}
	}
	return v, nil
}

func (v InvoicesView) handleConfirmPaidKey(msg tea.KeyMsg) (InvoicesView, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if v.selected != nil {
			return v, v.markAsPaid(v.selected.ID)
		}
	case "n", "N", "esc":
		if v.mode == InvoiceModeConfirmPaid {
			if v.selected != nil {
				v.mode = InvoiceModeDetail
			} else {
				v.mode = InvoiceModeList
			}
		}
	}
	return v, nil
}

func (v InvoicesView) handleConfirmDeleteKey(msg tea.KeyMsg) (InvoicesView, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if v.selected != nil {
			return v, v.deleteInvoice(v.selected.ID)
		}
	case "n", "N", "esc":
		v.mode = InvoiceModeList
		v.selected = nil
	}
	return v, nil
}

// saveInvoice déclenche la sauvegarde via le service.
func (v InvoicesView) saveInvoice() tea.Cmd {
	f := v.form
	svc := v.services.Invoice
	return func() tea.Msg {
		inv, err := f.buildInvoice()
		if err != nil {
			return InvoiceSavedMsg{Err: err}
		}
		err = svc.Create(inv)
		return InvoiceSavedMsg{Err: err}
	}
}

// markAsPaid déclenche le marquage comme payée.
func (v InvoicesView) markAsPaid(id int) tea.Cmd {
	svc := v.services.Invoice
	return func() tea.Msg {
		err := svc.MarkAsPaid(id, time.Now())
		return InvoicePaidMsg{Err: err}
	}
}

// markAsIssued émet une facture brouillon.
func (v InvoicesView) markAsIssued(id int) tea.Cmd {
	svc := v.services.Invoice
	return func() tea.Msg {
		return InvoiceIssuedMsg{Err: svc.MarkAsIssued(id)}
	}
}

// markAsSent marque une facture comme envoyée.
func (v InvoicesView) markAsSent(id int) tea.Cmd {
	svc := v.services.Invoice
	return func() tea.Msg {
		return InvoiceSentMsg{Err: svc.MarkAsSent(id)}
	}
}

// deleteInvoice supprime un brouillon.
// LEGAL: seuls les brouillons peuvent être supprimés.
func (v InvoicesView) deleteInvoice(id int) tea.Cmd {
	svc := v.services.Invoice
	return func() tea.Msg {
		return InvoiceDeletedMsg{Err: svc.Delete(id)}
	}
}

// generatePDF déclenche la génération du PDF de la facture.
func (v InvoicesView) generatePDF(id int) tea.Cmd {
	svc := v.services.Invoice
	return func() tea.Msg {
		path, err := svc.GeneratePDF(id)
		return InvoicePDFMsg{Path: path, Err: err}
	}
}

// selectedInvoice retourne la facture sélectionnée dans la table.
func (v InvoicesView) selectedInvoice() *domain.Invoice {
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

// View rend la vue factures.
func (v InvoicesView) View() string {
	switch v.mode {
	case InvoiceModeForm:
		return v.renderForm()
	case InvoiceModeDetail:
		return v.renderDetail()
	case InvoiceModeConfirmPaid:
		return v.renderConfirmPaid()
	case InvoiceModeConfirmDelete:
		return v.renderConfirmDelete()
	default:
		return v.renderList()
	}
}

func (v InvoicesView) renderList() string {
	var sb strings.Builder

	if v.err != "" {
		sb.WriteString(styles.StyleDanger.Render("⚠ "+v.err) + "\n\n")
	}

	count := len(v.filtered)
	sb.WriteString(styles.StyleMuted.Render(fmt.Sprintf("  %d facture(s)", count)) + "\n\n")
	sb.WriteString(v.table.View() + "\n")

	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(
		"  n: Nouvelle  i: Émettre  s: Envoyer  m: Payée  c: Avoir  p: PDF  d: Supprimer  Enter: Détail  j/k: Nav",
	)
	sb.WriteString(hint)
	return sb.String()
}

func (v InvoicesView) renderForm() string {
	if v.form == nil {
		return ""
	}

	var sb strings.Builder

	if v.formErr != "" {
		sb.WriteString(styles.StyleDanger.Render("⚠ "+v.formErr) + "\n\n")
	}

	if v.form.step == invoiceStepBasic {
		sb.WriteString(v.form.basicForm.View())
		return sb.String()
	}

	// Étape lignes
	sb.WriteString(styles.StyleTitle.Render("NOUVELLE FACTURE — Lignes") + "\n\n")

	if len(v.form.lines) == 0 {
		sb.WriteString(styles.StyleMuted.Render("  Aucune ligne. Appuyez sur 'a' pour ajouter.\n\n"))
	} else {
		// Afficher les lignes existantes
		// Width() pour chaque cellule header → gère correctement les chars multi-octets (QTÉ)
		// et aligne avec les données (%12s€ = 13 chars visuels → Width(13))
		hs := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Bold(true)
		sb.WriteString(
			"  "+hs.Copy().Width(40).Render("DESCRIPTION")+
				"  "+hs.Copy().Width(8).Align(lipgloss.Right).Render("QTÉ")+
				"  "+hs.Copy().Width(13).Align(lipgloss.Right).Render("PRIX HT")+
				"  "+hs.Copy().Width(13).Align(lipgloss.Right).Render("TOTAL HT")+"\n")
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#374151")).
			Render(strings.Repeat("─", 82)) + "\n")

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
			sb.WriteString(lineStyle.Render(fmt.Sprintf("%s%-40s  %8s  %12s€  %12s€",
				prefix,
				truncate(line.description, 40),
				line.quantity,
				price.StringFixed(2),
				total.StringFixed(2),
			)) + "\n")
		}

		// Afficher le total
		inv, err := v.form.buildInvoice()
		if err == nil {
			inv.CalculateTotals()
			sb.WriteString("\n")
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#374151")).
				Render(strings.Repeat("─", 82)) + "\n")
			totalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Bold(true)
			// %-67s aligne le montant en bout de la colonne TOTAL HT (pos 82)
			sb.WriteString(totalStyle.Render(fmt.Sprintf("  %-67s%12s€", "TOTAL HT", inv.TotalHT.StringFixed(2))) + "\n")
			if v.form.cfg.VAT.Applicable {
				sb.WriteString(totalStyle.Render(fmt.Sprintf("  %-67s%12s€", "TVA", inv.VATAmount.StringFixed(2))) + "\n")
			}
			sb.WriteString(totalStyle.Render(fmt.Sprintf("  %-67s%12s€", "TOTAL TTC", inv.TotalTTC.StringFixed(2))) + "\n")
		}
	}

	sb.WriteString("\n")

	if v.form.lineForm != nil {
		sb.WriteString(v.form.lineForm.View())
	} else {
		hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(
			"  a: Ajouter ligne  d: Supprimer ligne  j/k: Sélectionner  Enter: Valider la facture  Esc: Retour",
		)
		sb.WriteString(hint)
	}

	return sb.String()
}

func (v InvoicesView) renderDetail() string {
	if v.selected == nil {
		return ""
	}
	inv := v.selected

	isLocked := !inv.CanEdit()
	isOverdue := inv.IsOverdue()

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(func() lipgloss.Color {
			if isLocked {
				return lipgloss.Color("#7F1D1D")
			}
			return lipgloss.Color("#374151")
		}()).
		Padding(1, 2).
		Width(v.width - 4)

	label := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Width(18)
	value := lipgloss.NewStyle().Foreground(lipgloss.Color("#F9FAFB"))

	row := func(l, val string) string {
		return label.Render(l) + value.Render(val) + "\n"
	}

	var content strings.Builder

	// En-tête avec badge immutabilité
	title := "FACTURE " + inv.Number
	if isLocked {
		title += "  " + styles.BadgeLocked()
	}
	content.WriteString(styles.StyleTitle.Render(title) + "\n\n")

	if isLocked {
		content.WriteString(styles.StyleImmutableBanner.Render(
			"⚠  Cette facture est PAYÉE — toute modification est interdite (Art. L441-9)") + "\n\n")
	}

	// Infos facture
	content.WriteString(row("Numéro", inv.Number))
	clientLabel := fmt.Sprint(inv.ClientID)
	if inv.Client != nil {
		clientLabel = inv.Client.Name
	}
	content.WriteString(row("Client", clientLabel))
	content.WriteString(row("État", styles.RenderInvoiceState(inv.State, isOverdue)))
	content.WriteString(row("Date émission", inv.IssueDate.Format("02/01/2006")))
	content.WriteString(row("Date échéance", inv.DueDate.Format("02/01/2006")))
	if !inv.DeliveryDate.IsZero() {
		content.WriteString(row("Date livraison", inv.DeliveryDate.Format("02/01/2006")))
	}
	if inv.PaidDate != nil {
		content.WriteString(row("Date paiement", inv.PaidDate.Format("02/01/2006")))
	}
	content.WriteString("\n")

	// Lignes
	if len(inv.Lines) > 0 {
		// 4 colonnes : desc(30) + qty(6) + pu(12) + total(12) = 2+30+2+6+2+12+2+12 = 68 chars
		hd := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Bold(true)
		content.WriteString(
			"  "+hd.Copy().Width(30).Render("DESCRIPTION")+
				"  "+hd.Copy().Width(6).Align(lipgloss.Right).Render("QTÉ")+
				"  "+hd.Copy().Width(12).Align(lipgloss.Right).Render("PU HT")+
				"  "+hd.Copy().Width(12).Align(lipgloss.Right).Render("TOTAL HT")+"\n")
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#374151")).
			Render(strings.Repeat("─", 68)) + "\n")
		rowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#D1D5DB"))
		for _, line := range inv.Lines {
			content.WriteString(rowStyle.Render(fmt.Sprintf("  %-30s  %6s  %11s€  %11s€",
				truncate(line.Description, 30),
				line.Quantity.StringFixed(2),
				line.UnitPriceHT.StringFixed(2),
				line.TotalHT.StringFixed(2),
			)) + "\n")
		}
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#374151")).
			Render(strings.Repeat("─", 68)) + "\n")
	}

	// Totaux — label sur 50 chars, montant aligné à droite sur 13 chars
	totalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Bold(true)
	content.WriteString(totalStyle.Render(fmt.Sprintf("  %-50s%12s€", "TOTAL HT", inv.TotalHT.StringFixed(2))) + "\n")
	if inv.VATApplicable {
		content.WriteString(totalStyle.Render(fmt.Sprintf("  %-50s%12s€", "TVA", inv.VATAmount.StringFixed(2))) + "\n")
	} else {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).
			Render("  "+domain.VATMentionExemption) + "\n")
	}
	content.WriteString(totalStyle.Render(fmt.Sprintf("  %-50s%12s€", "TOTAL TTC", inv.TotalTTC.StringFixed(2))) + "\n")
	content.WriteString("\n")

	// Conditions paiement
	pmt := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
	content.WriteString(pmt.Render(fmt.Sprintf("  Délai: %s  |  Pénalités: %s%%/an  |  Indemnité: %s€",
		inv.PaymentDeadline,
		inv.LatePenaltyRate.StringFixed(2),
		inv.RecoveryFee.StringFixed(0),
	)) + "\n")
	if inv.PaymentRef != "" {
		content.WriteString(pmt.Render("  Libellé virement : "+inv.PaymentRef) + "\n")
	}

	// PDFPath
	if inv.PDFPath != "" {
		content.WriteString(styles.StyleSuccess.Render("  PDF : "+inv.PDFPath) + "\n")
	}

	// Actions disponibles
	content.WriteString("\n")
	var actions []string
	if inv.State == domain.InvoiceStateDraft {
		actions = append(actions, "i: Émettre", "s: Envoyer")
	} else if inv.State == domain.InvoiceStateIssued {
		actions = append(actions, "s: Envoyer")
	}
	if inv.CanMarkAsPaid() {
		actions = append(actions, "m: Payée")
	}
	if inv.CanCancel() {
		actions = append(actions, "c: Avoir")
	}
	actions = append(actions, "p: PDF")
	actions = append(actions, "Esc: Retour")
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).
		Render("  " + strings.Join(actions, "  ")))

	return "\n" + box.Render(content.String())
}

func (v InvoicesView) renderConfirmPaid() string {
	if v.selected == nil {
		return ""
	}
	inv := v.selected
	msg := fmt.Sprintf("Marquer la facture %s comme PAYÉE ?\n\n"+
		"⚠  ATTENTION: Cette action est IRRÉVERSIBLE.\n"+
		"   La facture deviendra IMMUABLE (Art. L441-9 Code de Commerce).\n\n"+
		"   Montant: %s€ TTC",
		inv.Number, inv.TotalTTC.StringFixed(2))
	return "\n" + lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B")).
		Bold(true).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#F59E0B")).
		Render(msg) +
		"\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).
			Render("  [Y] Confirmer  [N/Esc] Annuler")
}

func (v InvoicesView) renderConfirmDelete() string {
	if v.selected == nil {
		return ""
	}
	msg := fmt.Sprintf("Supprimer le brouillon %s ?", v.selected.Number)
	return "\n" + lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EF4444")).
		Bold(true).
		Render("⚠  "+msg) +
		"\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).
			Render("  [Y] Confirmer  [N/Esc] Annuler")
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func invoiceColumns(width int) []table.Column {
	// ID(5)+NUMÉRO(12)+TTC(13)+ÉTAT(12)+ÉCHÉANCE(12)+VIR(12) = 66 + 7cols*2pad = 80 → CLIENT adaptatif plafonné
	clientW := min(24, max(16, width-90))
	return []table.Column{
		{Title: "ID", Width: 5},
		{Title: "NUMÉRO", Width: 12},
		{Title: "CLIENT", Width: clientW},
		{Title: "MONTANT TTC", Width: 13},
		{Title: "ÉTAT", Width: 12},
		{Title: "ÉCHÉANCE", Width: 12},
		{Title: "LIBELLÉ VIR", Width: 12},
	}
}

func invoiceRows(invoices []domain.Invoice) []table.Row {
	rows := make([]table.Row, len(invoices))
	for i, inv := range invoices {
		clientName := fmt.Sprint(inv.ClientID)
		if inv.Client != nil {
			clientName = inv.Client.Name
		}
		stateStr := string(inv.State)
		if inv.IsOverdue() {
			stateStr = "ÉCHUE"
		}
		rows[i] = table.Row{
			fmt.Sprint(inv.ID),
			inv.Number,
			clientName,
			inv.TotalTTC.StringFixed(2) + "€",
			stateStr,
			inv.DueDate.Format("02/01/06"),
			inv.PaymentRef,
		}
	}
	return rows
}

func filterInvoices(invoices []domain.Invoice, search string) []domain.Invoice {
	if search == "" {
		return invoices
	}
	q := strings.ToLower(search)
	var out []domain.Invoice
	for _, inv := range invoices {
		if strings.Contains(strings.ToLower(inv.Number), q) ||
			strings.Contains(strings.ToLower(string(inv.State)), q) {
			out = append(out, inv)
		}
	}
	return out
}
