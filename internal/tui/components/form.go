package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormEvent indique ce que le formulaire demande à son parent.
type FormEvent int

const (
	FormEventNone   FormEvent = iota
	FormEventSubmit           // l'utilisateur a soumis
	FormEventCancel           // l'utilisateur a annulé (Esc)
)

// FormField est un champ de saisie avec libellé et message d'erreur.
type FormField struct {
	Label       string
	Placeholder string
	Input       textinput.Model
	Error       string
	Required    bool
}

// NewField crée un champ de formulaire.
func NewField(label, placeholder string, required bool) FormField {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Width = 40
	return FormField{
		Label:       label,
		Placeholder: placeholder,
		Required:    required,
		Input:       ti,
	}
}

// Form est un formulaire de saisie multi-champs.
type Form struct {
	Title  string
	Fields []FormField
	focused int
	width   int
}

// NewForm crée un formulaire.
func NewForm(title string, fields []FormField, width int) *Form {
	f := &Form{Title: title, Fields: fields, width: width}
	if len(fields) > 0 {
		f.Fields[0].Input.Focus()
	}
	return f
}

// Init implémente la partie Init de tea.Model.
func (f *Form) Init() tea.Cmd {
	return textinput.Blink
}

// Update gère les événements clavier du formulaire.
// Retourne l'événement détecté (submit/cancel/none).
func (f *Form) Update(msg tea.Msg) (FormEvent, tea.Cmd) {
	var cmds []tea.Cmd

	keyMsg, isKey := msg.(tea.KeyMsg)
	if !isKey {
		// Mettre à jour l'input focalisé même pour les non-key messages (ex: blink)
		updated, cmd := f.Fields[f.focused].Input.Update(msg)
		f.Fields[f.focused].Input = updated
		return FormEventNone, cmd
	}

	switch keyMsg.String() {
	case "esc":
		return FormEventCancel, nil

	case "tab", "down":
		f.Fields[f.focused].Input.Blur()
		f.Fields[f.focused].Error = "" // effacer erreur au changement
		f.focused = (f.focused + 1) % len(f.Fields)
		f.Fields[f.focused].Input.Focus()
		cmds = append(cmds, textinput.Blink)

	case "shift+tab", "up":
		f.Fields[f.focused].Input.Blur()
		f.Fields[f.focused].Error = ""
		f.focused = (f.focused - 1 + len(f.Fields)) % len(f.Fields)
		f.Fields[f.focused].Input.Focus()
		cmds = append(cmds, textinput.Blink)

	case "enter":
		if f.focused == len(f.Fields)-1 {
			return FormEventSubmit, nil
		}
		// Passer au champ suivant sur Enter (sauf dernier)
		f.Fields[f.focused].Input.Blur()
		f.Fields[f.focused].Error = ""
		f.focused++
		f.Fields[f.focused].Input.Focus()
		cmds = append(cmds, textinput.Blink)

	default:
		updated, cmd := f.Fields[f.focused].Input.Update(msg)
		f.Fields[f.focused].Input = updated
		cmds = append(cmds, cmd)
	}

	return FormEventNone, tea.Batch(cmds...)
}

// View rend le formulaire.
func (f *Form) View() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Width(20)

	focusedLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#0EA5E9")).
		Bold(true).
		Width(20)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EF4444"))

	inputBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#374151")).
		Padding(0, 1)

	focusedInputBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#0EA5E9")).
		Padding(0, 1)

	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#0EA5E9")).
		Bold(true).
		MarginBottom(1)

	sb.WriteString(titleStyle.Render(f.Title) + "\n\n")

	for i, field := range f.Fields {
		focused := i == f.focused
		label := field.Label
		if field.Required {
			label += " *"
		}

		var lbl, inp string
		if focused {
			lbl = focusedLabelStyle.Render(label)
			inp = focusedInputBorderStyle.Render(field.Input.View())
		} else {
			lbl = labelStyle.Render(label)
			inp = inputBorderStyle.Render(field.Input.View())
		}

		sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, lbl, inp) + "\n")

		if field.Error != "" {
			sb.WriteString("                     " + errorStyle.Render("⚠ "+field.Error) + "\n")
		}
	}

	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	required := lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render("*") + " obligatoire"
	sb.WriteString("\n" + hintStyle.Render(fmt.Sprintf(
		"Tab/↓: suivant  Shift+Tab/↑: précédent  Enter: %s  Esc: annuler",
		func() string {
			if f.focused == len(f.Fields)-1 {
				return "valider"
			}
			return "suivant"
		}(),
	)) + "  " + required + "\n")

	return sb.String()
}

// Value retourne la valeur d'un champ par index.
func (f *Form) Value(i int) string {
	if i < 0 || i >= len(f.Fields) {
		return ""
	}
	return strings.TrimSpace(f.Fields[i].Input.Value())
}

// SetValue définit la valeur d'un champ.
func (f *Form) SetValue(i int, v string) {
	if i >= 0 && i < len(f.Fields) {
		f.Fields[i].Input.SetValue(v)
	}
}

// SetError définit un message d'erreur sur un champ.
func (f *Form) SetError(i int, err string) {
	if i >= 0 && i < len(f.Fields) {
		f.Fields[i].Error = err
	}
}

// ClearErrors efface tous les messages d'erreur.
func (f *Form) ClearErrors() {
	for i := range f.Fields {
		f.Fields[i].Error = ""
	}
}

// Reset vide tous les champs et remet le focus au premier.
func (f *Form) Reset() {
	for i := range f.Fields {
		f.Fields[i].Input.SetValue("")
		f.Fields[i].Input.Blur()
		f.Fields[i].Error = ""
	}
	f.focused = 0
	if len(f.Fields) > 0 {
		f.Fields[0].Input.Focus()
	}
}

// HasActiveInput retourne true si un champ est en cours de saisie.
func (f *Form) HasActiveInput() bool {
	return f.Fields[f.focused].Input.Focused()
}
