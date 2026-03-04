package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kvitrvn/runar/internal/domain"
)

type pickerEvent int

const (
	pickerNone      pickerEvent = iota
	pickerSelected              // client sélectionné (Enter)
	pickerCancelled             // annulé (Esc)
)

// clientPicker est un sélecteur de client avec filtre live et navigation clavier.
type clientPicker struct {
	input    textinput.Model
	all      []domain.Client
	filtered []domain.Client
	cursor   int
}

func newClientPicker() clientPicker {
	ti := textinput.New()
	ti.Placeholder = "Tapez le nom du client..."
	ti.Prompt = ""
	ti.CharLimit = 64
	ti.Focus()
	return clientPicker{input: ti}
}

func (p *clientPicker) SetClients(clients []domain.Client) {
	p.all = clients
	p.refilter()
}

func (p *clientPicker) refilter() {
	q := strings.ToLower(p.input.Value())
	if q == "" {
		p.filtered = p.all
	} else {
		var out []domain.Client
		for _, c := range p.all {
			if strings.Contains(strings.ToLower(c.Name), q) ||
				strings.Contains(strings.ToLower(c.City), q) {
				out = append(out, c)
			}
		}
		p.filtered = out
	}
	if len(p.filtered) == 0 {
		p.cursor = 0
	} else if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
	}
}

// Selected retourne le client en surbrillance.
func (p *clientPicker) Selected() *domain.Client {
	if len(p.filtered) == 0 {
		return nil
	}
	c := p.filtered[p.cursor]
	return &c
}

// Update gère les événements clavier et les messages textinput.
func (p *clientPicker) Update(msg tea.Msg) (pickerEvent, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "j", "down":
			if p.cursor < len(p.filtered)-1 {
				p.cursor++
			}
			return pickerNone, nil
		case "k", "up":
			if p.cursor > 0 {
				p.cursor--
			}
			return pickerNone, nil
		case "enter":
			if len(p.filtered) > 0 {
				return pickerSelected, nil
			}
			return pickerNone, nil
		case "esc":
			return pickerCancelled, nil
		}
	}
	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)
	p.refilter()
	return pickerNone, cmd
}

// View rend le picker : input de recherche + liste filtrée.
func (p clientPicker) View(width int) string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F9FAFB"))
	sb.WriteString(titleStyle.Render("SÉLECTION DU CLIENT") + "\n\n")

	inputW := min(60, width-8)
	if inputW < 10 {
		inputW = 10
	}
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#1D4ED8")).
		Padding(0, 1).
		MarginLeft(2).
		Width(inputW)
	sb.WriteString(inputBox.Render(p.input.View()) + "\n\n")

	if len(p.all) == 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).
			Render("  Chargement des clients...") + "\n")
		return sb.String()
	}

	if len(p.filtered) == 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).
			Render("  Aucun client trouvé.") + "\n")
		return sb.String()
	}

	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).
		Render(fmt.Sprintf("  %d client(s)  ·  j/k: naviguer  ·  Enter: sélectionner  ·  Esc: annuler\n\n",
			len(p.filtered))))

	// Fenêtre glissante de 8 entrées
	const maxVisible = 8
	start := 0
	if p.cursor >= maxVisible {
		start = p.cursor - maxVisible + 1
	}
	end := min(len(p.filtered), start+maxVisible)

	for i := start; i < end; i++ {
		c := p.filtered[i]
		city := ""
		if c.City != "" {
			city = "  " + truncate(c.City, 18)
		}
		if i == p.cursor {
			line := fmt.Sprintf("  ▶ %-38s%s", truncate(c.Name, 38), city)
			sb.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F9FAFB")).
				Background(lipgloss.Color("#1D4ED8")).
				Bold(true).
				Render(line) + "\n")
		} else {
			line := fmt.Sprintf("    %-38s%s", truncate(c.Name, 38), city)
			sb.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#D1D5DB")).
				Render(line) + "\n")
		}
	}

	return sb.String()
}
