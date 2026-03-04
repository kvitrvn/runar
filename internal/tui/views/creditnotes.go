package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kvitrvn/runar/internal/domain"
	"github.com/kvitrvn/runar/internal/service"
	"github.com/kvitrvn/runar/internal/tui/components"
	"github.com/kvitrvn/runar/internal/tui/styles"
	"github.com/shopspring/decimal"
)

// CreditNoteMode représente le sous-mode de la vue avoirs.
type CreditNoteMode int

const (
	CreditNoteModeList   CreditNoteMode = iota
	CreditNoteModeForm                  // formulaire création
	CreditNoteModeDetail                // vue détail
)

// Messages internes.
type CreditNotesLoadedMsg struct{ CreditNotes []domain.CreditNote; Err error }
type CreditNoteSavedMsg struct{ Err error }
type CreditNotePDFMsg struct{ Path string; Err error }

// CreditNotesView est la vue de gestion des avoirs.
type CreditNotesView struct {
	services  *service.Services
	mode      CreditNoteMode
	cns       []domain.CreditNote
	table     components.TableModel
	form      *components.Form
	selected  *domain.CreditNote
	invoiceID int    // facture source pour la création
	invNumber string // numéro de la facture source
	err       string
	width     int
	height    int
}

// NewCreditNotesView crée la vue avoirs.
func NewCreditNotesView(services *service.Services, width, height int) CreditNotesView {
	cols := cnColumns(width)
	t := components.NewTable(cols, nil, height-6)
	return CreditNotesView{
		services: services,
		table:    t,
		width:    width,
		height:   height,
	}
}

// Load déclenche le chargement des avoirs.
func (v CreditNotesView) Load() tea.Cmd {
	svc := v.services.CreditNote
	return func() tea.Msg {
		cns, err := svc.List()
		return CreditNotesLoadedMsg{CreditNotes: cns, Err: err}
	}
}

// OpenFormForInvoice ouvre le formulaire de création d'un avoir pour une facture donnée.
func (v *CreditNotesView) OpenFormForInvoice(invoiceID int, invoiceNumber, totalHT, vatAmount string) {
	v.invoiceID = invoiceID
	v.invNumber = invoiceNumber
	v.mode = CreditNoteModeForm
	v.form = newCreditNoteForm(invoiceNumber, totalHT, vatAmount, v.width)
	v.err = ""
}

// SetSize ajuste les dimensions et recalcule les colonnes.
func (v *CreditNotesView) SetSize(w, h int) {
	v.width = w
	v.height = h
	v.table.SetColumns(cnColumns(w))
	v.table.SetHeight(h - 6)
}

// IsInputActive retourne true si le formulaire est actif.
func (v CreditNotesView) IsInputActive() bool {
	return v.mode == CreditNoteModeForm
}

// IsInSubMode retourne true si la vue n'est pas en mode liste.
func (v CreditNotesView) IsInSubMode() bool {
	return v.mode != CreditNoteModeList
}

// Update gère les messages.
func (v CreditNotesView) Update(msg tea.Msg) (CreditNotesView, tea.Cmd) {
	switch msg := msg.(type) {
	case CreditNotesLoadedMsg:
		if msg.Err != nil {
			v.err = msg.Err.Error()
		} else {
			v.cns = msg.CreditNotes
			v.table.SetRows(cnRows(msg.CreditNotes))
			v.err = ""
		}

	case CreditNoteSavedMsg:
		if msg.Err != nil {
			v.err = msg.Err.Error()
		} else {
			v.mode = CreditNoteModeList
			v.form = nil
			v.invoiceID = 0
			v.invNumber = ""
			v.err = ""
			return v, v.Load()
		}

	case CreditNotePDFMsg:
		if msg.Err != nil {
			v.err = msg.Err.Error()
		} else {
			v.err = ""
			// Le PDF est généré — on pourrait l'ouvrir avec xdg-open (Sprint 6)
		}

	case tea.KeyMsg:
		switch v.mode {
		case CreditNoteModeList:
			return v.handleListKey(msg)
		case CreditNoteModeForm:
			return v.handleFormKey(msg)
		case CreditNoteModeDetail:
			return v.handleDetailKey(msg)
		}
	}

	if v.mode == CreditNoteModeList {
		updated, cmd := v.table.Update(msg)
		v.table = updated
		return v, cmd
	}
	return v, nil
}

func (v CreditNotesView) handleListKey(msg tea.KeyMsg) (CreditNotesView, tea.Cmd) {
	switch msg.String() {
	case "enter":
		row := v.table.SelectedRow()
		if row != nil {
			var id int
			fmt.Sscanf(row[0], "%d", &id)
			for i := range v.cns {
				if v.cns[i].ID == id {
					v.selected = &v.cns[i]
					v.mode = CreditNoteModeDetail
					break
				}
			}
		}
	case "p":
		row := v.table.SelectedRow()
		if row != nil {
			var id int
			fmt.Sscanf(row[0], "%d", &id)
			return v, v.generatePDF(id)
		}
	default:
		updated, cmd := v.table.Update(msg)
		v.table = updated
		return v, cmd
	}
	return v, nil
}

func (v CreditNotesView) handleFormKey(msg tea.KeyMsg) (CreditNotesView, tea.Cmd) {
	event, cmd := v.form.Update(msg)
	switch event {
	case components.FormEventCancel:
		v.mode = CreditNoteModeList
		v.form = nil
		v.invoiceID = 0
		v.invNumber = ""
	case components.FormEventSubmit:
		return v, v.saveCreditNote()
	}
	return v, cmd
}

func (v CreditNotesView) handleDetailKey(msg tea.KeyMsg) (CreditNotesView, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		v.mode = CreditNoteModeList
		v.selected = nil
	case "p":
		if v.selected != nil {
			return v, v.generatePDF(v.selected.ID)
		}
	}
	return v, nil
}

func (v CreditNotesView) saveCreditNote() tea.Cmd {
	svc := v.services.CreditNote
	invoiceID := v.invoiceID
	reason := v.form.Value(0)
	htStr := v.form.Value(1)
	vatStr := v.form.Value(2)
	return func() tea.Msg {
		totalHT, _ := decimal.NewFromString(htStr)
		vatAmount, _ := decimal.NewFromString(vatStr)
		cn := &domain.CreditNote{
			IssueDate: time.Now(),
			Reason:    reason,
			TotalHT:   totalHT,
			VATAmount: vatAmount,
			TotalTTC:  totalHT.Add(vatAmount),
		}
		err := svc.CreateFromInvoice(invoiceID, cn)
		return CreditNoteSavedMsg{Err: err}
	}
}

func (v CreditNotesView) generatePDF(id int) tea.Cmd {
	svc := v.services.CreditNote
	return func() tea.Msg {
		path, err := svc.GeneratePDF(id)
		return CreditNotePDFMsg{Path: path, Err: err}
	}
}

// View rend la vue avoirs.
func (v CreditNotesView) View() string {
	switch v.mode {
	case CreditNoteModeForm:
		return v.renderForm()
	case CreditNoteModeDetail:
		return v.renderDetail()
	default:
		return v.renderList()
	}
}

func (v CreditNotesView) renderList() string {
	var sb strings.Builder
	if v.err != "" {
		sb.WriteString(styles.StyleDanger.Render("⚠ "+v.err) + "\n\n")
	}
	sb.WriteString(styles.StyleMuted.Render(fmt.Sprintf("  %d avoir(s)", len(v.cns))) + "\n\n")
	sb.WriteString(v.table.View() + "\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).
		Render("  p: Générer PDF  Enter: Détail  j/k: Navigation"))
	return sb.String()
}

func (v CreditNotesView) renderForm() string {
	if v.form == nil {
		return ""
	}
	header := styles.StyleTitle.Render("NOUVEL AVOIR") + "\n"
	if v.invNumber != "" {
		header += styles.StyleMuted.Render("  Facture d'origine : "+v.invNumber) + "\n"
	}
	header += "\n"
	return header + v.form.View()
}

func (v CreditNotesView) renderDetail() string {
	if v.selected == nil {
		return ""
	}
	cn := v.selected

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#374151")).
		Padding(1, 2).
		Width(v.width - 4)

	label := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Width(20)
	value := lipgloss.NewStyle().Foreground(lipgloss.Color("#F9FAFB"))
	row := func(l, val string) string { return label.Render(l) + value.Render(val) + "\n" }

	var content strings.Builder
	content.WriteString(styles.StyleTitle.Render("AVOIR "+cn.Number) + "\n\n")
	content.WriteString(row("Numéro", cn.Number))
	content.WriteString(row("Facture d'origine", cn.InvoiceReference))
	content.WriteString(row("Date", cn.IssueDate.Format("02/01/2006")))
	content.WriteString(row("Motif", cn.Reason))
	content.WriteString("\n")
	totalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Bold(true)
	content.WriteString(totalStyle.Render(fmt.Sprintf("  Total TTC : %s €", cn.TotalTTC.StringFixed(2))) + "\n")
	if cn.PDFPath != "" {
		content.WriteString("\n" + styles.StyleSuccess.Render("  PDF : "+cn.PDFPath) + "\n")
	}
	content.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).
		Render("  p: Générer PDF  Esc: Retour"))

	return "\n" + box.Render(content.String())
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func cnColumns(width int) []table.Column {
	// Colonnes fixes : ID(5)+NUMÉRO(14)+FACTURE D'ORIGINE(18)+TOTAL TTC(12)+DATE(12) = 61 + 6*2 = 73
	motifW := max(18, width-73)
	return []table.Column{
		{Title: "ID", Width: 5},
		{Title: "NUMÉRO", Width: 14},
		{Title: "FACTURE D'ORIGINE", Width: 18},
		{Title: "MOTIF", Width: motifW},
		{Title: "TOTAL TTC", Width: 12},
		{Title: "DATE", Width: 12},
	}
}

func cnRows(cns []domain.CreditNote) []table.Row {
	rows := make([]table.Row, len(cns))
	for i, cn := range cns {
		rows[i] = table.Row{
			fmt.Sprint(cn.ID),
			cn.Number,
			cn.InvoiceReference,
			truncate(cn.Reason, 30),
			cn.TotalTTC.StringFixed(2) + "€",
			cn.IssueDate.Format("02/01/2006"),
		}
	}
	return rows
}

func newCreditNoteForm(invoiceNumber, totalHT, vatAmount string, width int) *components.Form {
	fields := []components.FormField{
		components.NewField("Motif", "Erreur de facturation sur "+invoiceNumber, true),
		components.NewField("Montant HT €", totalHT, true),
		components.NewField("TVA €", vatAmount, false),
	}
	return components.NewForm("NOUVEL AVOIR — "+invoiceNumber, fields, width)
}
