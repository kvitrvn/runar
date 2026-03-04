// Package tui implémente l'interface utilisateur TUI de l'application.
package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kvitrvn/runar/internal/config"
	"github.com/kvitrvn/runar/internal/service"
	"github.com/kvitrvn/runar/internal/tui/styles"
	"github.com/kvitrvn/runar/internal/tui/views"
)

// tickMsg est envoyé périodiquement pour rafraîchir les toasts.
type tickMsg time.Time

// App est le modèle principal de l'application TUI.
type App struct {
	services *service.Services
	config   *config.Config
	width    int
	height   int

	currentView ViewType
	mode        AppMode

	// Vues (chargées à la demande)
	dashboardView   views.DashboardView
	clientsView     views.ClientsView
	invoicesView    views.InvoicesView
	creditNotesView views.CreditNotesView
	quotesView      views.QuotesView

	// Saisie commande / recherche
	input textinput.Model

	// Toast courant
	toast *Toast

	// Filtre de recherche actif
	searchQuery string
}

// NewApp crée l'application TUI.
func NewApp(services *service.Services, cfg *config.Config) *App {
	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = 64

	return &App{
		services:        services,
		config:          cfg,
		currentView:     ViewPulse,
		mode:            ModeNormal,
		input:           ti,
		dashboardView:   views.NewDashboardView(services, cfg, 80, 20),
		clientsView:     views.NewClientsView(services, 80, 20),
		invoicesView:    views.NewInvoicesView(services, cfg, 80, 20),
		creditNotesView: views.NewCreditNotesView(services, 80, 20),
		quotesView:      views.NewQuotesView(services, cfg, 80, 20),
	}
}

// Init implémente tea.Model.
func (m *App) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, tickCmd(), m.dashboardView.Load())
}

func tickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update implémente tea.Model.
func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		mainH := m.height - 3
		m.dashboardView.SetSize(m.width, mainH)
		m.clientsView.SetSize(m.width, mainH)
		m.invoicesView.SetSize(m.width, mainH)
		m.creditNotesView.SetSize(m.width, mainH)
		m.quotesView.SetSize(m.width, mainH)

	case tickMsg:
		if m.toast != nil && !m.toast.IsVisible() {
			m.toast = nil
		}
		cmds = append(cmds, tickCmd())

	case tea.KeyMsg:
		// Les touches globales (Ctrl+C, :, /, ?) ne sont interceptées
		// que si la vue active n'a pas de saisie en cours.
		activeInputBusy := m.isActiveViewInputBusy()

		if !activeInputBusy {
			switch m.mode {
			case ModeCommand:
				if cmd := m.handleCommandKey(msg); cmd != nil {
					cmds = append(cmds, cmd)
				}
				// Mettre à jour l'input de saisie
				updated, inputCmd := m.input.Update(msg)
				m.input = updated
				cmds = append(cmds, inputCmd)
				return m, tea.Batch(cmds...)
			case ModeSearch:
				if cmd := m.handleSearchKey(msg); cmd != nil {
					cmds = append(cmds, cmd)
				}
				updated, inputCmd := m.input.Update(msg)
				m.input = updated
				cmds = append(cmds, inputCmd)
				return m, tea.Batch(cmds...)
			case ModeHelp:
				m.handleHelpKey(msg)
				return m, nil
			default:
				// Touches globales mode Normal
				switch msg.String() {
				case "ctrl+c":
					return m, tea.Quit
				case ":":
					m.enterCommandMode()
					return m, textinput.Blink
				case "/":
					m.enterSearchMode()
					return m, textinput.Blink
				case "?":
					m.mode = ModeHelp
					return m, nil
				case "q":
					// Quitter seulement si on est sur une vue sans sous-mode actif
					if !m.isViewInSubMode() {
						return m, tea.Quit
					}
				case "tab":
					m.cycleView()
					cmd := m.loadCurrentView()
					return m, cmd
				}
			}
		}

	// Dashboard
	case views.DashboardLoadedMsg:
		var cmd tea.Cmd
		m.dashboardView, cmd = m.dashboardView.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	// Export CSV
	case views.ExportDoneMsg:
		if msg.Err != nil {
			m.showToast("Export échoué : "+msg.Err.Error(), ToastError)
		} else {
			m.showToast("Export créé : "+msg.Path, ToastSuccess)
		}
		return m, nil

	// Messages des vues
	case views.ClientsLoadedMsg, views.ClientSavedMsg, views.ClientDeletedMsg:
		var cmd tea.Cmd
		m.clientsView, cmd = m.clientsView.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case views.InvoicesLoadedMsg, views.InvoiceSavedMsg, views.InvoicePaidMsg, views.InvoiceDeletedMsg,
		views.InvoiceIssuedMsg, views.InvoiceSentMsg, views.InvoiceDetailLoadedMsg:
		var cmd tea.Cmd
		m.invoicesView, cmd = m.invoicesView.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case views.InvoicePDFMsg:
		if msg.Err != nil {
			m.showToast("PDF facture : "+msg.Err.Error(), ToastError)
		} else {
			m.showToast("PDF généré : "+msg.Path, ToastSuccess)
		}
		var cmd tea.Cmd
		m.invoicesView, cmd = m.invoicesView.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case views.OpenCreditNoteFormMsg:
		m.currentView = ViewCreditNotes
		m.creditNotesView.OpenFormForInvoice(msg.InvoiceID, msg.InvoiceNumber, msg.TotalHT, msg.VATAmount)
		return m, m.creditNotesView.Load()

	case views.CreditNotesLoadedMsg, views.CreditNoteSavedMsg, views.CreditNoteDetailLoadedMsg:
		var cmd tea.Cmd
		m.creditNotesView, cmd = m.creditNotesView.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case views.CreditNotePDFMsg:
		if msg.Err != nil {
			m.showToast("PDF avoir : "+msg.Err.Error(), ToastError)
		} else {
			m.showToast("PDF généré : "+msg.Path, ToastSuccess)
		}
		var cmd tea.Cmd
		m.creditNotesView, cmd = m.creditNotesView.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case views.QuotesLoadedMsg, views.QuoteSavedMsg, views.QuoteStateChangedMsg, views.QuoteDetailLoadedMsg:
		var cmd tea.Cmd
		m.quotesView, cmd = m.quotesView.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case views.QuotePDFMsg:
		if msg.Err != nil {
			m.showToast("PDF devis : "+msg.Err.Error(), ToastError)
		} else {
			m.showToast("PDF généré : "+msg.Path, ToastSuccess)
		}
		var cmd tea.Cmd
		m.quotesView, cmd = m.quotesView.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case views.QuoteConvertedMsg:
		var cmd tea.Cmd
		m.quotesView, cmd = m.quotesView.Update(msg)
		cmds = append(cmds, cmd)
		if msg.Err != nil {
			m.showToast("Conversion échouée : "+msg.Err.Error(), ToastError)
		} else {
			m.showToast("Devis converti → Facture "+msg.InvoiceNumber, ToastSuccess)
			m.currentView = ViewInvoices
			cmds = append(cmds, m.invoicesView.Load(m.searchQuery))
		}
		return m, tea.Batch(cmds...)
	}

	// Déléguer à la vue active
	var viewCmd tea.Cmd
	switch m.currentView {
	case ViewClients:
		m.clientsView, viewCmd = m.clientsView.Update(msg)
	case ViewInvoices:
		m.invoicesView, viewCmd = m.invoicesView.Update(msg)
	case ViewCreditNotes:
		m.creditNotesView, viewCmd = m.creditNotesView.Update(msg)
	case ViewQuotes:
		m.quotesView, viewCmd = m.quotesView.Update(msg)
	}
	if viewCmd != nil {
		cmds = append(cmds, viewCmd)
	}

	return m, tea.Batch(cmds...)
}

// isActiveViewInputBusy retourne true si la vue active a un champ texte focalisé.
func (m *App) isActiveViewInputBusy() bool {
	switch m.currentView {
	case ViewClients:
		return m.clientsView.IsInputActive()
	case ViewInvoices:
		return m.invoicesView.IsInputActive()
	case ViewCreditNotes:
		return m.creditNotesView.IsInputActive()
	case ViewQuotes:
		return m.quotesView.IsInputActive()
	}
	return false
}

// isViewInSubMode retourne true si la vue active est dans un sous-mode (detail, form…).
func (m *App) isViewInSubMode() bool {
	switch m.currentView {
	case ViewClients:
		return m.clientsView.IsInSubMode()
	case ViewInvoices:
		return m.invoicesView.IsInSubMode()
	case ViewCreditNotes:
		return m.creditNotesView.IsInSubMode()
	case ViewQuotes:
		return m.quotesView.IsInSubMode()
	}
	return false
}

// handleCommandKey gère les touches en mode commande.
func (m *App) handleCommandKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.exitInputMode()
	case "enter":
		cmd := m.executeCommand(m.input.Value())
		m.exitInputMode()
		return cmd
	case "tab":
		suggestions := Autocomplete(m.input.Value())
		if len(suggestions) > 0 {
			m.input.SetValue(suggestions[0].Name)
			m.input.CursorEnd()
		}
	}
	return nil
}

// handleSearchKey gère les touches en mode recherche.
func (m *App) handleSearchKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.searchQuery = ""
		m.exitInputMode()
	case "enter":
		m.searchQuery = m.input.Value()
		m.exitInputMode()
		return m.loadCurrentView()
	}
	return nil
}

// handleHelpKey gère les touches en mode aide.
func (m *App) handleHelpKey(msg tea.KeyMsg) {
	switch msg.String() {
	case "?", "esc", "q":
		m.mode = ModeNormal
	}
}

func (m *App) enterCommandMode() {
	m.mode = ModeCommand
	m.input.SetValue("")
	m.input.Focus()
}

func (m *App) enterSearchMode() {
	m.mode = ModeSearch
	m.input.SetValue("")
	m.input.Focus()
}

func (m *App) exitInputMode() {
	m.mode = ModeNormal
	m.input.SetValue("")
	m.input.Blur()
}

// executeCommand exécute la commande saisie et charge la vue si nécessaire.
func (m *App) executeCommand(raw string) tea.Cmd {
	cmd := ParseCommand(raw)
	if cmd == nil {
		m.showToast(fmt.Sprintf("Commande inconnue : %q", raw), ToastError)
		return nil
	}
	if cmd.IsQuit {
		return tea.Quit
	}
	if cmd.Name == "help" {
		m.mode = ModeHelp
		return nil
	}
	if cmd.Action != "" {
		return m.executeAction(cmd.Action)
	}
	prev := m.currentView
	m.currentView = cmd.View
	m.searchQuery = ""
	if m.currentView != prev {
		return m.loadCurrentView()
	}
	return nil
}

// executeAction exécute une action non-navigation (export, etc.).
func (m *App) executeAction(action string) tea.Cmd {
	svc := m.services.Export
	outputDir := "./exports"
	if m.config != nil && m.config.PDF.OutputDir != "" {
		outputDir = m.config.PDF.OutputDir
	}
	year := time.Now().Year()

	switch action {
	case "export-factures":
		m.showToast("Export factures en cours...", ToastInfo)
		return func() tea.Msg {
			path, err := svc.ExportInvoicesCSV(year, outputDir)
			return views.ExportDoneMsg{Path: path, Err: err}
		}
	case "export-avoirs":
		m.showToast("Export avoirs en cours...", ToastInfo)
		return func() tea.Msg {
			path, err := svc.ExportCreditNotesCSV(outputDir)
			return views.ExportDoneMsg{Path: path, Err: err}
		}
	}
	return nil
}

// loadCurrentView déclenche le chargement des données pour la vue courante.
func (m *App) loadCurrentView() tea.Cmd {
	switch m.currentView {
	case ViewPulse:
		return m.dashboardView.Load()
	case ViewClients:
		return m.clientsView.Load(m.searchQuery)
	case ViewInvoices:
		return m.invoicesView.Load(m.searchQuery)
	case ViewCreditNotes:
		return m.creditNotesView.Load()
	case ViewQuotes:
		return m.quotesView.Load(m.searchQuery)
	}
	return nil
}

// cycleView passe à la vue suivante.
func (m *App) cycleView() {
	allViews := []ViewType{ViewPulse, ViewClients, ViewInvoices, ViewQuotes, ViewCreditNotes}
	for i, v := range allViews {
		if v == m.currentView {
			m.currentView = allViews[(i+1)%len(allViews)]
			return
		}
	}
}

// showToast affiche une notification temporaire.
func (m *App) showToast(msg string, kind ToastType) {
	t := NewToast(msg, kind)
	m.toast = &t
}

// ─── Rendu ───────────────────────────────────────────────────────────────────

// View implémente tea.Model.
func (m *App) View() string {
	if m.width == 0 {
		return "Chargement..."
	}

	header := m.renderHeader()
	commandBar := m.renderCommandBar()
	infoBar := m.renderInfoBar()

	fixedH := lipgloss.Height(header) + lipgloss.Height(commandBar) + lipgloss.Height(infoBar)
	mainH := m.height - fixedH
	if mainH < 1 {
		mainH = 1
	}

	mainPane := m.renderMainPane(mainH)

	return lipgloss.JoinVertical(lipgloss.Left, header, commandBar, mainPane, infoBar)
}

func (m *App) renderHeader() string {
	year := time.Now().Year()
	left := fmt.Sprintf(" runar  %d  :%s", year, m.currentView.String())

	modeLabel := ""
	switch m.mode {
	case ModeCommand:
		modeLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Bold(true).Render(" [COMMANDE]")
	case ModeSearch:
		modeLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#0EA5E9")).Bold(true).Render(" [RECHERCHE]")
	case ModeHelp:
		modeLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#8B5CF6")).Bold(true).Render(" [AIDE]")
	}

	sellerName := ""
	if m.config != nil {
		sellerName = m.config.Seller.Name
	}
	right := ""
	if sellerName != "" {
		right = lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render(sellerName + " ")
	}

	content := left + modeLabel
	pad := m.width - lipgloss.Width(content) - lipgloss.Width(right)
	if pad < 0 {
		pad = 0
	}
	return styles.StyleHeader.Width(m.width).Render(content + strings.Repeat(" ", pad) + right)
}

func (m *App) renderCommandBar() string {
	baseStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#111827")).
		Foreground(lipgloss.Color("#F9FAFB")).
		Padding(0, 1)
	hintStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#111827")).
		Foreground(lipgloss.Color("#6B7280")).
		Padding(0, 1)
	prefixStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#0EA5E9")).Bold(true)

	hint := "[?] Aide  [Tab] Vue suivante"
	if m.mode == ModeCommand || m.mode == ModeSearch {
		hint = "[Enter] Valider  [Esc] Annuler  [Tab] Compléter"
	}
	hintR := hintStyle.Render(hint)
	leftW := m.width - lipgloss.Width(hintR)
	if leftW < 1 {
		leftW = 1
	}

	var leftContent string
	switch m.mode {
	case ModeCommand:
		suggestions := Autocomplete(m.input.Value())
		sugg := ""
		if len(suggestions) > 0 {
			names := make([]string, len(suggestions))
			for i, s := range suggestions {
				names[i] = s.Name
			}
			sugg = lipgloss.NewStyle().Foreground(lipgloss.Color("#374151")).Render("  " + strings.Join(names, "  "))
		}
		leftContent = prefixStyle.Render(":") + m.input.View() + sugg
	case ModeSearch:
		leftContent = prefixStyle.Render("/") + m.input.View()
	default:
		filter := ""
		if m.searchQuery != "" {
			filter = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).
				Render(" [/" + m.searchQuery + "]")
		}
		leftContent = prefixStyle.Render(":") + m.currentView.String() + filter
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, baseStyle.Width(leftW).Render(leftContent), hintR)
}

func (m *App) renderMainPane(mainH int) string {
	// Mode aide : overlay centré
	if m.mode == ModeHelp {
		helpPanel := views.RenderHelpPanel(m.width)
		helpH := lipgloss.Height(helpPanel)
		helpW := lipgloss.Width(helpPanel)
		padTop := (mainH - helpH) / 2
		padLeft := (m.width - helpW) / 2
		if padTop < 0 {
			padTop = 0
		}
		if padLeft < 0 {
			padLeft = 0
		}
		// Préfixer CHAQUE ligne du panneau pour le centrer horizontalement
		lines := strings.Split(helpPanel, "\n")
		pad := strings.Repeat(" ", padLeft)
		for i, l := range lines {
			lines[i] = pad + l
		}
		return strings.Repeat("\n", padTop) + strings.Join(lines, "\n")
	}

	title, content := m.renderActiveView()

	sep := styles.StyleTitle.Render("─ "+title+" ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#374151")).
			Render(strings.Repeat("─", maxInt(0, m.width-lipgloss.Width(title)-6)))

	// Toast superposé si présent
	toastStr := ""
	if m.toast != nil && m.toast.IsVisible() {
		toastStr = "\n" + m.renderToast()
	}

	return lipgloss.NewStyle().Height(mainH).MaxHeight(mainH).Render(sep + "\n" + content + toastStr)
}

func (m *App) renderActiveView() (title, content string) {
	switch m.currentView {
	case ViewClients:
		return "CLIENTS", m.clientsView.View()
	case ViewInvoices:
		return "FACTURES", m.invoicesView.View()
	case ViewPulse:
		return "DASHBOARD", m.dashboardView.View()
	case ViewQuotes:
		return "DEVIS", m.quotesView.View()
	case ViewCreditNotes:
		return "AVOIRS", m.creditNotesView.View()
	}
	return "INCONNU", ""
}

func renderPlaceholderView(subtitle string, actions []string) string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(styles.StyleMuted.Render("  "+subtitle) + "\n\n")
	for _, a := range actions {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render("  "+a) + "\n")
	}
	return sb.String()
}

func (m *App) renderInfoBar() string {
	baseStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#1F2937")).
		Foreground(lipgloss.Color("#6B7280")).
		Padding(0, 1)
	rightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#1F2937")).
		Foreground(lipgloss.Color("#9CA3AF")).
		Padding(0, 1)

	left := fmt.Sprintf("vue: %s", m.currentView.String())
	if m.searchQuery != "" {
		left += fmt.Sprintf("  │  filtre: %q", m.searchQuery)
	}
	right := rightStyle.Render("runar v0.1.0")
	rightW := lipgloss.Width(right)
	leftW := m.width - rightW
	if leftW < 1 {
		leftW = 1
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, baseStyle.Width(leftW).Render(left), right)
}

func (m *App) renderToast() string {
	if m.toast == nil {
		return ""
	}
	switch m.toast.Type {
	case ToastSuccess:
		return styles.StyleToastSuccess.Render("✓ " + m.toast.Message)
	case ToastError:
		return styles.StyleToastError.Render("✗ " + m.toast.Message)
	case ToastWarning:
		return styles.StyleToastWarning.Render("⚠ " + m.toast.Message)
	default:
		return styles.StyleToastInfo.Render("ℹ " + m.toast.Message)
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
