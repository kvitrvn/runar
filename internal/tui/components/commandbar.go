package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CommandBar est la barre de commande k9s-like (ligne 2 du layout).
type CommandBar struct {
	input       textinput.Model
	active      bool   // true = mode commande ou recherche
	prefix      string // ":" ou "/"
	currentView string // Vue courante affichée en mode normal
	hint        string // Hint affiché à droite
	width       int
	autocomplete []string // Suggestions d'autocomplétion
}

// NewCommandBar crée une CommandBar.
func NewCommandBar(currentView string, width int) CommandBar {
	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = 64
	ti.Width = width - 20

	return CommandBar{
		input:       ti,
		currentView: currentView,
		hint:        "[?] Aide",
		width:       width,
	}
}

// Activate passe en mode saisie avec le préfixe donné (":" ou "/").
func (c *CommandBar) Activate(prefix string) {
	c.active = true
	c.prefix = prefix
	c.input.SetValue("")
	c.input.Focus()
	c.autocomplete = nil
}

// Deactivate repasse en mode normal.
func (c *CommandBar) Deactivate() {
	c.active = false
	c.prefix = ""
	c.input.SetValue("")
	c.input.Blur()
	c.autocomplete = nil
}

// Value retourne la valeur saisie (sans le préfixe).
func (c *CommandBar) Value() string {
	return c.input.Value()
}

// SetCurrentView met à jour la vue affichée en mode normal.
func (c *CommandBar) SetCurrentView(v string) {
	c.currentView = v
}

// SetAutocomplete met à jour les suggestions.
func (c *CommandBar) SetAutocomplete(suggestions []string) {
	c.autocomplete = suggestions
}

// SetWidth ajuste la largeur.
func (c *CommandBar) SetWidth(w int) {
	c.width = w
	c.input.Width = w - 20
}

// Update gère les messages clavier en mode actif.
func (c CommandBar) Update(msg tea.Msg) (CommandBar, tea.Cmd) {
	if !c.active {
		return c, nil
	}
	var cmd tea.Cmd
	c.input, cmd = c.input.Update(msg)
	return c, cmd
}

// View rend la barre de commande.
func (c CommandBar) View() string {
	leftStyle := lipgloss.NewStyle().
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

	var left string
	if c.active {
		left = fmt.Sprintf("%s%s", prefixStyle.Render(c.prefix), c.input.View())
		// Afficher autocomplétion si disponible
		if len(c.autocomplete) > 0 {
			suggestions := strings.Join(c.autocomplete, "  ")
			left += lipgloss.NewStyle().Foreground(lipgloss.Color("#374151")).Render("  " + suggestions)
		}
	} else {
		left = prefixStyle.Render(":") + c.currentView
	}

	hint := hintStyle.Render(c.hint)

	// Calculer l'espace disponible pour le contenu gauche
	hintWidth := lipgloss.Width(hint)
	leftWidth := c.width - hintWidth
	if leftWidth < 1 {
		leftWidth = 1
	}

	leftRendered := leftStyle.Width(leftWidth).Render(left)
	return lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, hint)
}
