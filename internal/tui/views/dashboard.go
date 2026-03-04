package views

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kvitrvn/runar/internal/config"
	"github.com/kvitrvn/runar/internal/domain"
	"github.com/kvitrvn/runar/internal/repository"
	"github.com/kvitrvn/runar/internal/service"
	"github.com/kvitrvn/runar/internal/tui/styles"
	"github.com/shopspring/decimal"
)

// LEGAL: Seuil de franchise en base TVA pour prestations de services (Art. 293B CGI).
// Valeur 2024-2026 : 36 800 € (seuil majoré : 39 100 €).
const vatFranchiseThreshold = 36_800.0

// eInvoicingDeadline = date de l'obligation de facturation électronique pour les TPE/auto-entrepreneurs.
var eInvoicingDeadline = time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)

// DashboardLoadedMsg est envoyé lorsque les données du dashboard sont chargées.
type DashboardLoadedMsg struct {
	Invoices []domain.Invoice
	Quotes   []domain.Quote
	Err      error
}

// ExportDoneMsg est envoyé après un export CSV.
type ExportDoneMsg struct {
	Path string
	Err  error
}

// DashboardView est la vue tableau de bord (:pulse).
type DashboardView struct {
	services *service.Services
	config   *config.Config
	invoices []domain.Invoice
	quotes   []domain.Quote
	year     int
	loaded   bool
	err      string
	width    int
	height   int
}

// NewDashboardView crée la vue dashboard.
func NewDashboardView(svc *service.Services, cfg *config.Config, width, height int) DashboardView {
	return DashboardView{
		services: svc,
		config:   cfg,
		year:     time.Now().Year(),
		width:    width,
		height:   height,
	}
}

// Load charge les données depuis les services (asynchrone).
func (v DashboardView) Load() tea.Cmd {
	svc := v.services
	return func() tea.Msg {
		invoices, err := svc.Invoice.List(repository.InvoiceFilters{})
		if err != nil {
			return DashboardLoadedMsg{Err: err}
		}
		quotes, err := svc.Quote.List("")
		if err != nil {
			return DashboardLoadedMsg{Err: err}
		}
		return DashboardLoadedMsg{Invoices: invoices, Quotes: quotes}
	}
}

// SetSize met à jour les dimensions.
func (v *DashboardView) SetSize(w, h int) {
	v.width = w
	v.height = h
}

// Update gère les messages.
func (v DashboardView) Update(msg tea.Msg) (DashboardView, tea.Cmd) {
	if m, ok := msg.(DashboardLoadedMsg); ok {
		if m.Err != nil {
			v.err = m.Err.Error()
		} else {
			v.invoices = m.Invoices
			v.quotes = m.Quotes
			v.loaded = true
			v.err = ""
		}
	}
	return v, nil
}

// View rend le dashboard.
func (v DashboardView) View() string {
	if v.err != "" {
		return "\n" + styles.StyleDanger.Render("  ⚠ "+v.err)
	}
	if !v.loaded {
		return "\n" + styles.StyleMuted.Render("  Chargement du tableau de bord...")
	}
	return v.render()
}

// ─── Styles locaux ────────────────────────────────────────────────────────────

var (
	dashLabel   = lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Width(26)
	dashValue   = lipgloss.NewStyle().Foreground(lipgloss.Color("#F9FAFB")).Bold(true)
	dashSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Bold(true)
	dashWarn    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Bold(true)
	dashDanger  = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Bold(true)
	dashMuted   = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	dashBlue    = lipgloss.NewStyle().Foreground(lipgloss.Color("#0EA5E9"))
	dashSep     = lipgloss.NewStyle().Foreground(lipgloss.Color("#374151"))
	dashBox     = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#374151")).
			Padding(0, 1)
)

// ─── Rendu principal ──────────────────────────────────────────────────────────

func (v DashboardView) render() string {
	w := v.width - 4
	if w < 40 {
		w = 40
	}
	sep := dashSep.Render(strings.Repeat("─", w))

	// Ligne intro
	sellerLine := "  Exercice " + fmt.Sprint(v.year)
	if v.config != nil && v.config.Seller.Name != "" {
		pad := w - len("Exercice "+fmt.Sprint(v.year)) - len(v.config.Seller.Name) - 4
		if pad < 2 {
			pad = 2
		}
		sellerLine = "  " + dashMuted.Render("Exercice "+fmt.Sprint(v.year)) +
			strings.Repeat(" ", pad) + dashBlue.Render(v.config.Seller.Name)
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(sellerLine + "\n\n")

	// Bloc CA
	sb.WriteString(v.renderCABlock(w))
	sb.WriteString("\n\n")

	// Factures + Devis côte à côte
	colW := (w - 2) / 2
	left := v.renderInvoiceStats(colW)
	right := v.renderQuoteStats(colW)
	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right))
	sb.WriteString("\n\n")

	// Graphique CA mensuel
	sb.WriteString(v.renderMonthlyChart(w))
	sb.WriteString("\n")

	// Alertes
	alerts := v.renderAlerts(w)
	if alerts != "" {
		sb.WriteString("\n")
		sb.WriteString(alerts)
	}

	_ = sep
	return sb.String()
}

// ─── Bloc chiffre d'affaires ─────────────────────────────────────────────────

func (v DashboardView) renderCABlock(w int) string {
	now := time.Now()
	year := v.year

	caAnnuel := decimal.Zero
	caMois := decimal.Zero
	for _, inv := range v.invoices {
		if inv.State != domain.InvoiceStatePaid || inv.PaidDate == nil {
			continue
		}
		if inv.PaidDate.Year() == year {
			caAnnuel = caAnnuel.Add(inv.TotalTTC)
			if int(inv.PaidDate.Month()) == int(now.Month()) {
				caMois = caMois.Add(inv.TotalTTC)
			}
		}
	}

	threshold := decimal.NewFromFloat(vatFranchiseThreshold)
	pct := decimal.Zero
	if threshold.IsPositive() && caAnnuel.IsPositive() {
		pct = caAnnuel.Div(threshold).Mul(decimal.NewFromInt(100)).Round(1)
	}

	// Couleur selon le taux d'occupation du seuil TVA
	barColor := lipgloss.Color("#10B981")
	pctStyle := dashSuccess
	if pct.GreaterThan(decimal.NewFromInt(80)) {
		barColor = lipgloss.Color("#EF4444")
		pctStyle = dashDanger
	} else if pct.GreaterThan(decimal.NewFromInt(60)) {
		barColor = lipgloss.Color("#F59E0B")
		pctStyle = dashWarn
	}

	// Barre de progression TVA (20 chars)
	const barTotal = 20
	filled := int(pct.Mul(decimal.NewFromInt(barTotal)).Div(decimal.NewFromInt(100)).IntPart())
	if filled > barTotal {
		filled = barTotal
	}
	bar := lipgloss.NewStyle().Foreground(barColor).Render(strings.Repeat("█", filled)) +
		dashMuted.Render(strings.Repeat("░", barTotal-filled))

	var sb strings.Builder
	sb.WriteString(styles.StyleTitle.Render("  CHIFFRE D'AFFAIRES") + "\n")
	sb.WriteString(dashLabel.Render("  CA annuel "+fmt.Sprint(year)) +
		dashValue.Render(fmtAmount(caAnnuel)) + "\n")
	sb.WriteString(dashLabel.Render("  Ce mois ("+monthFR(int(now.Month()))+")") +
		dashValue.Render(fmtAmount(caMois)) + "\n")
	sb.WriteString(dashLabel.Render("  Seuil franchise TVA") +
		dashMuted.Render(fmtAmount(threshold)) + "\n")
	sb.WriteString(dashLabel.Render("  Taux occupé") +
		pctStyle.Render(pct.StringFixed(1)+"% ") + bar + "\n")

	return dashBox.Width(w).Render(sb.String())
}

// ─── Stats factures ───────────────────────────────────────────────────────────

func (v DashboardView) renderInvoiceStats(w int) string {
	counts := map[domain.InvoiceState]int{}
	overdue := 0
	for _, inv := range v.invoices {
		counts[inv.State]++
		if inv.IsOverdue() {
			overdue++
		}
	}

	lbl := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Width(18)
	val := func(n int, s lipgloss.Style) string { return lbl.Render("") + s.Render(fmt.Sprint(n)) + "\n" }
	row := func(label string, n int, s lipgloss.Style) string {
		return lbl.Render("  "+label) + s.Render(fmt.Sprint(n)) + "\n"
	}

	var sb strings.Builder
	sb.WriteString(styles.StyleTitle.Render("  FACTURES") + "\n\n")
	sb.WriteString(row("Brouillons", counts[domain.InvoiceStateDraft], dashMuted))
	sb.WriteString(row("Émises", counts[domain.InvoiceStateIssued], dashBlue))
	sb.WriteString(row("Envoyées", counts[domain.InvoiceStateSent], dashBlue))
	sb.WriteString(row("Payées", counts[domain.InvoiceStatePaid], dashSuccess))
	if counts[domain.InvoiceStateCanceled] > 0 {
		sb.WriteString(row("Annulées", counts[domain.InvoiceStateCanceled], dashMuted))
	}
	if overdue > 0 {
		sb.WriteString(row("En retard ⚠", overdue, dashDanger))
	}

	_ = val
	return dashBox.Width(w).Render(sb.String())
}

// ─── Stats devis ──────────────────────────────────────────────────────────────

func (v DashboardView) renderQuoteStats(w int) string {
	counts := map[domain.QuoteState]int{}
	expiringSoon := 0
	limit := time.Now().AddDate(0, 0, 14)
	for _, q := range v.quotes {
		counts[q.State]++
		active := q.State == domain.QuoteStateDraft || q.State == domain.QuoteStateSent
		if active && !q.ExpiryDate.IsZero() && q.ExpiryDate.Before(limit) && q.ExpiryDate.After(time.Now()) {
			expiringSoon++
		}
	}

	lbl := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Width(18)
	row := func(label string, n int, s lipgloss.Style) string {
		return lbl.Render("  "+label) + s.Render(fmt.Sprint(n)) + "\n"
	}

	var sb strings.Builder
	sb.WriteString(styles.StyleTitle.Render("  DEVIS") + "\n\n")
	sb.WriteString(row("Brouillons", counts[domain.QuoteStateDraft], dashMuted))
	sb.WriteString(row("Envoyés", counts[domain.QuoteStateSent], dashBlue))
	sb.WriteString(row("Acceptés", counts[domain.QuoteStateAccepted], dashSuccess))
	sb.WriteString(row("Refusés", counts[domain.QuoteStateRefused], dashMuted))
	if expiringSoon > 0 {
		sb.WriteString(row("Expirent <14j ⚠", expiringSoon, dashWarn))
	}

	return dashBox.Width(w).Render(sb.String())
}

// ─── Graphique CA mensuel ─────────────────────────────────────────────────────

func (v DashboardView) renderMonthlyChart(w int) string {
	year := v.year
	var monthly [12]decimal.Decimal
	for _, inv := range v.invoices {
		if inv.State != domain.InvoiceStatePaid || inv.PaidDate == nil {
			continue
		}
		if inv.PaidDate.Year() != year {
			continue
		}
		m := int(inv.PaidDate.Month()) - 1
		monthly[m] = monthly[m].Add(inv.TotalTTC)
	}

	maxCA := decimal.Zero
	for _, ca := range monthly {
		if ca.GreaterThan(maxCA) {
			maxCA = ca
		}
	}

	// Largeur de barre disponible : w - "  Mar  " (7) - "  1 250,00 €" (14) = w - 21
	barMaxW := w - 21
	if barMaxW < 5 {
		barMaxW = 5
	}
	if barMaxW > 40 {
		barMaxW = 40
	}

	currentMonth := int(time.Now().Month())
	monthNames := []string{"Jan", "Fév", "Mar", "Avr", "Mai", "Jun", "Jul", "Aoû", "Sep", "Oct", "Nov", "Déc"}

	var sb strings.Builder
	sb.WriteString(styles.StyleTitle.Render("  CA MENSUEL "+fmt.Sprint(year)) + "\n\n")

	for i, ca := range monthly {
		month := i + 1
		if year == time.Now().Year() && month > currentMonth {
			break // Ne pas afficher les mois futurs pour l'année courante
		}

		barLen := 0
		if maxCA.IsPositive() {
			barLen = int(ca.Div(maxCA).Mul(decimal.NewFromInt(int64(barMaxW))).IntPart())
		}

		barColor := lipgloss.Color("#0EA5E9")
		amtStyle := dashValue
		if month == currentMonth {
			barColor = lipgloss.Color("#10B981")
			amtStyle = dashSuccess
		}

		bar := lipgloss.NewStyle().Foreground(barColor).Render(strings.Repeat("█", barLen)) +
			dashMuted.Render(strings.Repeat("░", barMaxW-barLen))

		sb.WriteString(fmt.Sprintf("  %s  %s  %s\n",
			dashMuted.Render(monthNames[i]),
			bar,
			amtStyle.Render(fmtAmount(ca)),
		))
	}

	return dashBox.Width(w).Render(sb.String())
}

// ─── Alertes ──────────────────────────────────────────────────────────────────

func (v DashboardView) renderAlerts(w int) string {
	var alerts []string

	// 1. Factures en retard
	overdue := 0
	for _, inv := range v.invoices {
		if inv.IsOverdue() {
			overdue++
		}
	}
	if overdue > 0 {
		alerts = append(alerts, dashDanger.Render("  ⚠  ")+
			fmt.Sprintf("%d facture(s) en retard — :factures pour les voir", overdue))
	}

	// 2. Devis expirant dans < 14 jours
	expiringSoon := 0
	limit := time.Now().AddDate(0, 0, 14)
	for _, q := range v.quotes {
		active := q.State == domain.QuoteStateDraft || q.State == domain.QuoteStateSent
		if active && !q.ExpiryDate.IsZero() && q.ExpiryDate.Before(limit) && q.ExpiryDate.After(time.Now()) {
			expiringSoon++
		}
	}
	if expiringSoon > 0 {
		alerts = append(alerts, dashWarn.Render("  ⚠  ")+
			fmt.Sprintf("%d devis expire(nt) dans moins de 14 jours — :devis pour les voir", expiringSoon))
	}

	// 3. Seuil TVA
	caAnnuel := decimal.Zero
	for _, inv := range v.invoices {
		if inv.State == domain.InvoiceStatePaid && inv.PaidDate != nil && inv.PaidDate.Year() == v.year {
			caAnnuel = caAnnuel.Add(inv.TotalTTC)
		}
	}
	threshold := decimal.NewFromFloat(vatFranchiseThreshold)
	if threshold.IsPositive() && caAnnuel.IsPositive() {
		pct := caAnnuel.Div(threshold).Mul(decimal.NewFromInt(100))
		if pct.GreaterThan(decimal.NewFromInt(80)) {
			alerts = append(alerts, dashDanger.Render("  ⚠  ")+
				// LEGAL: Alerte seuil franchise TVA (Art. 293B CGI)
				fmt.Sprintf("CA à %.1f%% du seuil franchise TVA (36 800 €) — vérifiez votre régime", pct.InexactFloat64()))
		} else if pct.GreaterThan(decimal.NewFromInt(60)) {
			alerts = append(alerts, dashWarn.Render("  ⚠  ")+
				fmt.Sprintf("CA à %.1f%% du seuil franchise TVA (36 800 €)", pct.InexactFloat64()))
		}
	}

	// 4. Facturation électronique 2027
	// LEGAL: Obligation de facturation électronique pour tous les assujettis à la TVA (LF 2024)
	daysLeft := int(time.Until(eInvoicingDeadline).Hours() / 24)
	if daysLeft > 0 {
		alerts = append(alerts, dashMuted.Render("  ℹ  ")+
			fmt.Sprintf("Facturation électronique obligatoire dans %d jours (01/01/2027) — :export pour préparer vos données", daysLeft))
	}

	// 5. Aucun vendeur configuré
	if v.config != nil && v.config.Seller.Name == "" {
		alerts = append(alerts, dashWarn.Render("  ⚠  ")+
			"Infos vendeur non configurées — Créez config.yaml (runar --init-config)")
	}

	if len(alerts) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(styles.StyleTitle.Render("  ALERTES") + "\n\n")
	for _, a := range alerts {
		sb.WriteString(a + "\n")
	}
	return dashBox.Width(w).Render(sb.String())
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// fmtAmount formate un montant avec 2 décimales et le symbole €.
func fmtAmount(d decimal.Decimal) string {
	return d.StringFixed(2) + " €"
}

// monthFR retourne le nom court du mois en français.
func monthFR(m int) string {
	names := []string{"", "Jan", "Fév", "Mar", "Avr", "Mai", "Jun", "Jul", "Aoû", "Sep", "Oct", "Nov", "Déc"}
	if m < 1 || m > 12 {
		return ""
	}
	return names[m]
}
