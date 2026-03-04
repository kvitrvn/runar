package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kvitrvn/runar/internal/domain"
	"github.com/kvitrvn/runar/internal/service"
	"github.com/kvitrvn/runar/internal/tui/components"
	"github.com/kvitrvn/runar/internal/tui/styles"
)

// ClientMode représente le sous-mode de la vue clients.
type ClientMode int

const (
	ClientModeList          ClientMode = iota
	ClientModeNew                      // formulaire création
	ClientModeEdit                     // formulaire édition
	ClientModeDetail                   // vue détail
	ClientModeConfirmDelete            // confirmation suppression
)

// Messages internes à la vue.
type ClientsLoadedMsg struct{ Clients []domain.Client; Err error }
type ClientSavedMsg struct{ Err error }
type ClientDeletedMsg struct{ Err error }


// ClientsView est la vue complète de gestion des clients.
type ClientsView struct {
	services  *service.Services
	mode      ClientMode
	clients   []domain.Client
	filtered  []domain.Client
	table     components.TableModel
	form      *components.Form
	editingID int
	selected  *domain.Client
	search    string
	err       string
	width     int
	height    int
}

// NewClientsView crée la vue clients.
func NewClientsView(services *service.Services, width, height int) ClientsView {
	cols := clientColumns(width)
	t := components.NewTable(cols, nil, height-6)
	return ClientsView{
		services: services,
		table:    t,
		width:    width,
		height:   height,
	}
}

// Load déclenche le chargement des clients (cmd à exécuter par l'App).
func (v ClientsView) Load(search string) tea.Cmd {
	svc := v.services.Client
	return func() tea.Msg {
		clients, err := svc.List(search)
		return ClientsLoadedMsg{Clients: clients, Err: err}
	}
}

// SetSize ajuste les dimensions.
func (v *ClientsView) SetSize(w, h int) {
	v.width = w
	v.height = h
	v.table.SetColumns(clientColumns(w))
	v.table.SetHeight(h - 6)
}

// IsInputActive retourne true si un formulaire est actif.
func (v ClientsView) IsInputActive() bool {
	return v.mode == ClientModeNew || v.mode == ClientModeEdit
}

// IsInSubMode retourne true si la vue n'est pas en mode liste.
func (v ClientsView) IsInSubMode() bool {
	return v.mode != ClientModeList
}

// Update gère les messages de la vue.
func (v ClientsView) Update(msg tea.Msg) (ClientsView, tea.Cmd) {
	switch msg := msg.(type) {
	case ClientsLoadedMsg:
		if msg.Err != nil {
			v.err = msg.Err.Error()
		} else {
			v.clients = msg.Clients
			v.filtered = filterClients(msg.Clients, v.search)
			v.table.SetRows(clientRows(v.filtered))
			v.err = ""
		}

	case ClientSavedMsg:
		if msg.Err != nil {
			v.err = msg.Err.Error()
		} else {
			v.mode = ClientModeList
			v.form = nil
			v.err = ""
			return v, v.Load(v.search)
		}

	case ClientDeletedMsg:
		if msg.Err != nil {
			v.err = msg.Err.Error()
			v.mode = ClientModeList
		} else {
			v.mode = ClientModeList
			v.selected = nil
			v.err = ""
			return v, v.Load(v.search)
		}

	case tea.KeyMsg:
		switch v.mode {
		case ClientModeList:
			return v.handleListKey(msg)
		case ClientModeNew, ClientModeEdit:
			return v.handleFormKey(msg)
		case ClientModeDetail:
			return v.handleDetailKey(msg)
		case ClientModeConfirmDelete:
			return v.handleConfirmKey(msg)
		}
	}

	// Déléguer à la table en mode liste
	if v.mode == ClientModeList {
		updated, cmd := v.table.Update(msg)
		v.table = updated
		return v, cmd
	}

	return v, nil
}

func (v ClientsView) handleListKey(msg tea.KeyMsg) (ClientsView, tea.Cmd) {
	switch msg.String() {
	case "n":
		v.mode = ClientModeNew
		v.editingID = 0
		v.form = newClientForm(v.width)
	case "e":
		sel := v.selectedClient()
		if sel != nil {
			v.mode = ClientModeEdit
			v.editingID = sel.ID
			v.form = newClientForm(v.width)
			populateClientForm(v.form, sel)
		}
	case "d":
		sel := v.selectedClient()
		if sel != nil {
			v.selected = sel
			v.mode = ClientModeConfirmDelete
		}
	case "enter":
		sel := v.selectedClient()
		if sel != nil {
			v.selected = sel
			v.mode = ClientModeDetail
		}
	case "f":
		sel := v.selectedClient()
		if sel != nil {
			return v, func() tea.Msg {
				return OpenInvoiceFormForClientMsg{ClientID: sel.ID, ClientName: sel.Name}
			}
		}
	case "v":
		sel := v.selectedClient()
		if sel != nil {
			return v, func() tea.Msg {
				return OpenQuoteFormForClientMsg{ClientID: sel.ID, ClientName: sel.Name}
			}
		}
	default:
		updated, cmd := v.table.Update(msg)
		v.table = updated
		return v, cmd
	}
	return v, nil
}

func (v ClientsView) handleFormKey(msg tea.KeyMsg) (ClientsView, tea.Cmd) {
	event, cmd := v.form.Update(msg)
	switch event {
	case components.FormEventCancel:
		v.mode = ClientModeList
		v.form = nil
	case components.FormEventSubmit:
		return v, v.saveClient()
	}
	return v, cmd
}

func (v ClientsView) handleDetailKey(msg tea.KeyMsg) (ClientsView, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		v.mode = ClientModeList
		v.selected = nil
	case "e":
		if v.selected != nil {
			v.mode = ClientModeEdit
			v.editingID = v.selected.ID
			v.form = newClientForm(v.width)
			populateClientForm(v.form, v.selected)
		}
	case "d":
		if v.selected != nil {
			v.mode = ClientModeConfirmDelete
		}
	case "f":
		if v.selected != nil {
			sel := v.selected
			return v, func() tea.Msg {
				return OpenInvoiceFormForClientMsg{ClientID: sel.ID, ClientName: sel.Name}
			}
		}
	case "v":
		if v.selected != nil {
			sel := v.selected
			return v, func() tea.Msg {
				return OpenQuoteFormForClientMsg{ClientID: sel.ID, ClientName: sel.Name}
			}
		}
	}
	return v, nil
}

func (v ClientsView) handleConfirmKey(msg tea.KeyMsg) (ClientsView, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if v.selected != nil {
			return v, v.deleteClient(v.selected.ID)
		}
	case "n", "N", "esc":
		v.mode = ClientModeList
		v.selected = nil
	}
	return v, nil
}

// saveClient déclenche la sauvegarde du client via le service.
func (v ClientsView) saveClient() tea.Cmd {
	form := v.form
	svc := v.services.Client
	editingID := v.editingID

	return func() tea.Msg {
		client := &domain.Client{
			Name:       form.Value(0),
			SIRET:      form.Value(1),
			SIREN:      form.Value(2),
			Address:    form.Value(3),
			PostalCode: form.Value(4),
			City:       form.Value(5),
			Country:    form.Value(6),
			Email:      form.Value(7),
			Phone:      form.Value(8),
			Notes:      form.Value(9),
		}
		var err error
		if editingID > 0 {
			err = svc.Update(editingID, client)
		} else {
			err = svc.Create(client)
		}
		return ClientSavedMsg{Err: err}
	}
}

// deleteClient déclenche la suppression via le service.
func (v ClientsView) deleteClient(id int) tea.Cmd {
	svc := v.services.Client
	return func() tea.Msg {
		err := svc.Delete(id)
		return ClientDeletedMsg{Err: err}
	}
}

// selectedClient retourne le client correspondant à la ligne sélectionnée.
func (v ClientsView) selectedClient() *domain.Client {
	row := v.table.SelectedRow()
	if row == nil {
		return nil
	}
	// La première colonne est l'ID
	var id int
	fmt.Sscanf(row[0], "%d", &id)
	for i := range v.filtered {
		if v.filtered[i].ID == id {
			return &v.filtered[i]
		}
	}
	return nil
}

// View rend la vue clients.
func (v ClientsView) View() string {
	switch v.mode {
	case ClientModeNew, ClientModeEdit:
		return v.renderForm()
	case ClientModeDetail:
		return v.renderDetail()
	case ClientModeConfirmDelete:
		return v.renderConfirmDelete()
	default:
		return v.renderList()
	}
}

func (v ClientsView) renderList() string {
	var sb strings.Builder

	if v.err != "" {
		sb.WriteString(styles.StyleDanger.Render("⚠ "+v.err) + "\n\n")
	}

	count := len(v.filtered)
	sb.WriteString(styles.StyleMuted.Render(fmt.Sprintf("  %d client(s)", count)) + "\n\n")
	sb.WriteString(v.table.View() + "\n")

	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(
		"  n: Nouveau  e: Éditer  d: Supprimer  f: Facture  v: Devis  Enter: Détail  j/k: Navigation",
	)
	sb.WriteString(hint)
	return sb.String()
}

func (v ClientsView) renderForm() string {
	if v.form == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("\n")
	if v.err != "" {
		sb.WriteString(styles.StyleDanger.Render("⚠ "+v.err) + "\n\n")
	}
	sb.WriteString(v.form.View())
	return sb.String()
}

func (v ClientsView) renderDetail() string {
	if v.selected == nil {
		return ""
	}
	c := v.selected

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#374151")).
		Padding(1, 2).
		Width(v.width - 4)

	label := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Width(18)
	value := lipgloss.NewStyle().Foreground(lipgloss.Color("#F9FAFB"))

	row := func(l, val string) string {
		return label.Render(l) + value.Render(val) + "\n"
	}

	var content strings.Builder
	content.WriteString(styles.StyleTitle.Render("CLIENT #"+fmt.Sprint(c.ID)+" — "+c.Name) + "\n\n")
	content.WriteString(row("Nom", c.Name))
	if c.SIRET != "" {
		content.WriteString(row("SIRET", c.SIRET))
	}
	if c.SIREN != "" {
		content.WriteString(row("SIREN", c.SIREN))
	}
	if c.Address != "" {
		content.WriteString(row("Adresse", c.Address))
	}
	if c.PostalCode != "" || c.City != "" {
		content.WriteString(row("Ville", c.PostalCode+" "+c.City))
	}
	if c.Country != "" {
		content.WriteString(row("Pays", c.Country))
	}
	if c.Email != "" {
		content.WriteString(row("Email", c.Email))
	}
	if c.Phone != "" {
		content.WriteString(row("Téléphone", c.Phone))
	}
	if c.Notes != "" {
		content.WriteString("\n" + label.Render("Notes") + value.Render(c.Notes) + "\n")
	}

	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).
		Render("\n  e: Éditer  d: Supprimer  f: Nouvelle facture  v: Nouveau devis  Esc: Retour")
	content.WriteString(hint)

	return "\n" + box.Render(content.String())
}

func (v ClientsView) renderConfirmDelete() string {
	if v.selected == nil {
		return ""
	}
	msg := fmt.Sprintf("Supprimer le client %q (ID %d) ?", v.selected.Name, v.selected.ID)
	return "\n" + lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EF4444")).
		Bold(true).
		Render("⚠  "+msg) +
		"\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).
			Render("  [Y] Confirmer  [N/Esc] Annuler")
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func clientColumns(width int) []table.Column {
	// Colonnes fixes : ID(5)+SIRET(16)+EMAIL(24)+VILLE(14) = 59 + 5*2pad = 69
	// NOM remplit l'espace restant — le nom est la donnée la plus utile à afficher en entier
	nomW := max(16, width-69)
	return []table.Column{
		{Title: "ID", Width: 5},
		{Title: "NOM", Width: nomW},
		{Title: "SIRET", Width: 16},
		{Title: "EMAIL", Width: 24},
		{Title: "VILLE", Width: 14},
	}
}

func clientRows(clients []domain.Client) []table.Row {
	rows := make([]table.Row, len(clients))
	for i, c := range clients {
		rows[i] = table.Row{
			fmt.Sprint(c.ID),
			c.Name,
			c.SIRET,
			truncate(c.Email, 24),
			c.City,
		}
	}
	return rows
}

func filterClients(clients []domain.Client, search string) []domain.Client {
	if search == "" {
		return clients
	}
	q := strings.ToLower(search)
	var out []domain.Client
	for _, c := range clients {
		if strings.Contains(strings.ToLower(c.Name), q) ||
			strings.Contains(c.SIRET, q) ||
			strings.Contains(strings.ToLower(c.Email), q) ||
			strings.Contains(strings.ToLower(c.City), q) {
			out = append(out, c)
		}
	}
	return out
}

func newClientForm(width int) *components.Form {
	fields := []components.FormField{
		components.NewField("Nom", "Acme Corporation", true),
		components.NewField("SIRET", "", false),
		components.NewField("SIREN", "", false),
		components.NewField("Adresse", "123 Rue de la Paix", false),
		components.NewField("Code postal", "75001", false),
		components.NewField("Ville", "Paris", false),
		components.NewField("Pays", "France", false),
		components.NewField("Email", "contact@example.com", false),
		components.NewField("Téléphone", "+33 6 00 00 00 00", false),
		components.NewField("Notes", "", false),
	}
	return components.NewForm("CLIENT", fields, width)
}

func populateClientForm(f *components.Form, c *domain.Client) {
	f.SetValue(0, c.Name)
	f.SetValue(1, c.SIRET)
	f.SetValue(2, c.SIREN)
	f.SetValue(3, c.Address)
	f.SetValue(4, c.PostalCode)
	f.SetValue(5, c.City)
	f.SetValue(6, c.Country)
	f.SetValue(7, c.Email)
	f.SetValue(8, c.Phone)
	f.SetValue(9, c.Notes)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
