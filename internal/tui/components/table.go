package components

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TableModel est un wrapper autour de bubbles/table avec le thème runar.
type TableModel struct {
	table table.Model
}

// NewTable crée un TableModel stylisé.
func NewTable(columns []table.Column, rows []table.Row, height int) TableModel {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#374151")).
		BorderBottom(true).
		Foreground(lipgloss.Color("#9CA3AF")).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#F9FAFB")).
		Background(lipgloss.Color("#1D4ED8")).
		Bold(true)
	s.Cell = s.Cell.
		Foreground(lipgloss.Color("#D1D5DB"))

	t.SetStyles(s)
	return TableModel{table: t}
}

// Init implémente tea.Model.
func (m TableModel) Init() tea.Cmd {
	return nil
}

// Update implémente tea.Model, gère la navigation j/k/g/G.
func (m TableModel) Update(msg tea.Msg) (TableModel, tea.Cmd) {
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View implémente tea.Model.
func (m TableModel) View() string {
	return m.table.View()
}

// SelectedRow retourne la ligne sélectionnée.
func (m TableModel) SelectedRow() table.Row {
	return m.table.SelectedRow()
}

// SetRows met à jour les lignes.
func (m *TableModel) SetRows(rows []table.Row) {
	m.table.SetRows(rows)
}

// SetColumns met à jour les colonnes.
func (m *TableModel) SetColumns(cols []table.Column) {
	m.table.SetColumns(cols)
}

// SetHeight ajuste la hauteur de la table.
func (m *TableModel) SetHeight(h int) {
	m.table.SetHeight(h)
}

// Focus active la navigation clavier sur la table.
func (m *TableModel) Focus() {
	m.table.Focus()
}

// Blur désactive la navigation clavier.
func (m *TableModel) Blur() {
	m.table.Blur()
}
