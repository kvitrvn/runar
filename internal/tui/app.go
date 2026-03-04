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
	"github.com/kvitrvn/runar/internal/tui/components"
	"github.com/kvitrvn/runar/internal/tui/styles"
	"github.com/kvitrvn/runar/internal/tui/views"
)

// tickMsg est envoyé périodiquement pour rafraîchir les toasts.
type tickMsg time.Time

// App est le modèle principal de l'application TUI.
type App struct {
	services    *service.Services
	config      *config.Config
	width       int
	height      int
	currentView ViewType
	mode        AppMode

	// Composants
	commandBar components.CommandBar
	statusBar  components.StatusBar

	// Saisie active (commande ou recherche)
	input textinput.Model

	// Toast courant (notification temporaire)
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
		services:    services,
		config:      cfg,
		currentView: ViewPulse,
		mode:        ModeNormal,
		input:       ti,
	}
}

// Init implémente tea.Model.
func (m *App) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		tickCmd(),
	)
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

	case tickMsg:
		// Nettoyer toast expiré
		if m.toast != nil && !m.toast.IsVisible() {
			m.toast = nil
		}
		cmds = append(cmds, tickCmd())

	case tea.KeyMsg:
		switch m.mode {
		case ModeCommand:
			cmds = append(cmds, m.handleCommandKey(msg)...)
		case ModeSearch:
			cmds = append(cmds, m.handleSearchKey(msg)...)
		case ModeHelp:
			cmds = append(cmds, m.handleHelpKey(msg)...)
		default:
			cmds = append(cmds, m.handleNormalKey(msg)...)
		}
	}

	// Mettre à jour l'input si actif
	if m.mode == ModeCommand || m.mode == ModeSearch {
		var inputCmd tea.Cmd
		m.input, inputCmd = m.input.Update(msg)
		cmds = append(cmds, inputCmd)
	}

	return m, tea.Batch(cmds...)
}

// handleNormalKey gère les touches en mode normal.
func (m *App) handleNormalKey(msg tea.KeyMsg) []tea.Cmd {
	switch msg.String() {
	case "ctrl+c":
		return []tea.Cmd{tea.Quit}
	case ":":
		m.enterCommandMode()
	case "/":
		m.enterSearchMode()
	case "?":
		m.mode = ModeHelp
	case "q":
		return []tea.Cmd{tea.Quit}
	case "tab":
		m.cycleView()
	}
	return nil
}

// handleCommandKey gère les touches en mode commande.
func (m *App) handleCommandKey(msg tea.KeyMsg) []tea.Cmd {
	switch msg.String() {
	case "esc":
		m.exitInputMode()
	case "enter":
		m.executeCommand(m.input.Value())
		m.exitInputMode()
	case "tab":
		// Autocomplétion : prendre la première suggestion
		suggestions := Autocomplete(m.input.Value())
		if len(suggestions) > 0 {
			m.input.SetValue(suggestions[0].Name)
			// Repositionner le curseur à la fin
			m.input.CursorEnd()
		}
	}
	return nil
}

// handleSearchKey gère les touches en mode recherche.
func (m *App) handleSearchKey(msg tea.KeyMsg) []tea.Cmd {
	switch msg.String() {
	case "esc":
		m.searchQuery = ""
		m.exitInputMode()
	case "enter":
		m.searchQuery = m.input.Value()
		m.exitInputMode()
		if m.searchQuery != "" {
			m.showToast(fmt.Sprintf("Filtre actif : %q", m.searchQuery), ToastInfo)
		}
	}
	return nil
}

// handleHelpKey gère les touches en mode aide.
func (m *App) handleHelpKey(msg tea.KeyMsg) []tea.Cmd {
	switch msg.String() {
	case "?", "esc", "q":
		m.mode = ModeNormal
	}
	return nil
}

// enterCommandMode active le mode commande.
func (m *App) enterCommandMode() {
	m.mode = ModeCommand
	m.input.SetValue("")
	m.input.Focus()
}

// enterSearchMode active le mode recherche.
func (m *App) enterSearchMode() {
	m.mode = ModeSearch
	m.input.SetValue("")
	m.input.Focus()
}

// exitInputMode revient en mode normal.
func (m *App) exitInputMode() {
	m.mode = ModeNormal
	m.input.SetValue("")
	m.input.Blur()
}

// executeCommand exécute la commande saisie.
func (m *App) executeCommand(raw string) {
	cmd := ParseCommand(raw)
	if cmd == nil {
		m.showToast(fmt.Sprintf("Commande inconnue : %q  (tapez :help)", raw), ToastError)
		return
	}
	if cmd.IsQuit {
		// Sera traité via tea.Quit dans le prochain Update
		// On fait ça en passant par un message clavier synthétique
		// mais le plus simple est de quitter directement
		return
	}
	if cmd.Name == "help" {
		m.mode = ModeHelp
		return
	}
	m.currentView = cmd.View
	m.searchQuery = ""
}

// cycleView passe à la vue suivante.
func (m *App) cycleView() {
	views := []ViewType{ViewPulse, ViewClients, ViewInvoices, ViewQuotes, ViewCreditNotes}
	for i, v := range views {
		if v == m.currentView {
			m.currentView = views[(i+1)%len(views)]
			return
		}
	}
}

// showToast affiche une notification temporaire.
func (m *App) showToast(msg string, kind ToastType) {
	t := NewToast(msg, kind)
	m.toast = &t
}

// View implémente tea.Model — rendu 4 zones.
func (m *App) View() string {
	if m.width == 0 {
		return "Chargement..."
	}

	header := m.renderHeader()
	commandBar := m.renderCommandBar()
	mainPane := m.renderMainPane()
	infoBar := m.renderInfoBar()

	// Hauteur réservée aux zones fixes
	fixedHeight := lipgloss.Height(header) +
		lipgloss.Height(commandBar) +
		lipgloss.Height(infoBar)
	mainHeight := m.height - fixedHeight
	if mainHeight < 1 {
		mainHeight = 1
	}

	// Si aide visible, overlay centré sur le main pane
	if m.mode == ModeHelp {
		helpPanel := views.RenderHelpPanel(m.width)
		helpHeight := lipgloss.Height(helpPanel)
		helpWidth := lipgloss.Width(helpPanel)
		padTop := (mainHeight - helpHeight) / 2
		padLeft := (m.width - helpWidth) / 2
		if padTop < 0 {
			padTop = 0
		}
		if padLeft < 0 {
			padLeft = 0
		}
		mainPane = strings.Repeat("\n", padTop) +
			strings.Repeat(" ", padLeft) + helpPanel
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		commandBar,
		mainPane,
		infoBar,
	)
}

// renderHeader rend la ligne d'en-tête.
func (m *App) renderHeader() string {
	year := time.Now().Year()

	leftPart := fmt.Sprintf(" runar  %d  :%s", year, m.currentView.String())

	modeLabel := ""
	switch m.mode {
	case ModeCommand:
		modeLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Bold(true).Render(" [COMMANDE]")
	case ModeSearch:
		modeLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#0EA5E9")).Bold(true).Render(" [RECHERCHE]")
	case ModeHelp:
		modeLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#8B5CF6")).Bold(true).Render(" [AIDE]")
	}

	// Vendeur depuis config
	sellerName := ""
	if m.config != nil {
		sellerName = m.config.Seller.Name
	}

	rightPart := ""
	if sellerName != "" {
		rightPart = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			Render(sellerName + " ")
	}

	totalText := leftPart + modeLabel
	textWidth := lipgloss.Width(totalText) + lipgloss.Width(rightPart)
	padWidth := m.width - textWidth
	if padWidth < 0 {
		padWidth = 0
	}

	return styles.StyleHeader.Width(m.width).Render(
		totalText + strings.Repeat(" ", padWidth) + rightPart,
	)
}

// renderCommandBar rend la barre de commande.
func (m *App) renderCommandBar() string {
	baseStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#111827")).
		Foreground(lipgloss.Color("#F9FAFB")).
		Padding(0, 1)

	hintStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#111827")).
		Foreground(lipgloss.Color("#6B7280")).
		Padding(0, 1)

	prefixStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#0EA5E9")).
		Bold(true)

	hint := "[?] Aide  [Tab] Vue suivante"
	if m.mode == ModeCommand || m.mode == ModeSearch {
		hint = "[Enter] Valider  [Esc] Annuler  [Tab] Compléter"
	}
	hintRendered := hintStyle.Render(hint)
	hintWidth := lipgloss.Width(hintRendered)

	leftWidth := m.width - hintWidth
	if leftWidth < 1 {
		leftWidth = 1
	}

	var leftContent string
	switch m.mode {
	case ModeCommand:
		suggestions := Autocomplete(m.input.Value())
		suggStr := ""
		if len(suggestions) > 0 {
			names := make([]string, len(suggestions))
			for i, s := range suggestions {
				names[i] = s.Name
			}
			suggStr = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#374151")).
				Render("  " + strings.Join(names, "  "))
		}
		leftContent = prefixStyle.Render(":") + m.input.View() + suggStr
	case ModeSearch:
		leftContent = prefixStyle.Render("/") + m.input.View()
	default:
		filter := ""
		if m.searchQuery != "" {
			filter = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F59E0B")).
				Render(" [/" + m.searchQuery + "]")
		}
		leftContent = prefixStyle.Render(":") + m.currentView.String() + filter
	}

	left := baseStyle.Width(leftWidth).Render(leftContent)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, hintRendered)
}

// renderMainPane rend le contenu principal.
func (m *App) renderMainPane() string {
	fixedLines := 3 // header + commandbar + infobar
	mainHeight := m.height - fixedLines
	if mainHeight < 4 {
		mainHeight = 4
	}

	title, content := m.renderView()

	inner := styles.StyleTitle.Render("─ "+title+" ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#374151")).
			Render(strings.Repeat("─", max(0, m.width-lipgloss.Width(title)-6))) + "\n" +
		content

	// Toast en bas du main pane si actif
	if m.toast != nil && m.toast.IsVisible() {
		toastStr := m.renderToast()
		inner += "\n" + toastStr
	}

	return lipgloss.NewStyle().Height(mainHeight).MaxHeight(mainHeight).Render(inner)
}

// renderView retourne le titre et contenu pour la vue courante.
func (m *App) renderView() (string, string) {
	switch m.currentView {
	case ViewPulse:
		return "DASHBOARD", m.renderPlaceholder(
			"Tableau de bord — Sprint 7",
			[]string{"CA annuel", "Factures par état", "Alertes TVA", "Graphique mensuel"},
		)
	case ViewClients:
		return "CLIENTS", m.renderPlaceholder(
			"Liste des clients — Sprint 4",
			[]string{"n: Nouveau  e: Éditer  d: Supprimer  f: Factures  Enter: Voir"},
		)
	case ViewInvoices:
		return "FACTURES", m.renderPlaceholder(
			"Liste des factures — Sprint 4",
			[]string{"n: Nouvelle  e: Éditer  p: PDF  m: Marquer payée  c: Avoir"},
		)
	case ViewQuotes:
		return "DEVIS", m.renderPlaceholder(
			"Liste des devis — Sprint 5",
			[]string{"n: Nouveau  e: Éditer  f: Convertir en facture  p: PDF"},
		)
	case ViewCreditNotes:
		return "AVOIRS", m.renderPlaceholder(
			"Liste des avoirs — Sprint 5",
			[]string{"c: Créer avoir  p: PDF  Enter: Voir"},
		)
	}
	return "INCONNU", ""
}

// renderPlaceholder rend un panneau placeholder pour les vues Sprint 4+.
func (m *App) renderPlaceholder(subtitle string, actions []string) string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(styles.StyleMuted.Render("  "+subtitle) + "\n\n")
	for _, a := range actions {
		sb.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			Render("  "+a) + "\n")
	}
	return sb.String()
}

// renderInfoBar rend la barre d'information du bas.
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

	right := "runar v0.1.0"

	rightRendered := rightStyle.Render(right)
	rightWidth := lipgloss.Width(rightRendered)
	leftWidth := m.width - rightWidth
	if leftWidth < 1 {
		leftWidth = 1
	}

	leftRendered := baseStyle.Width(leftWidth).Render(left)
	return lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, rightRendered)
}

// renderToast rend le toast courant.
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

// max retourne le maximum de deux entiers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
