package components

import (
	"github.com/charmbracelet/lipgloss"
)

// StatusBar est la barre d'info du bas (ligne 4 du layout).
type StatusBar struct {
	left  string // Stats / message principal
	right string // Info secondaire (filtre actif, position)
	width int
}

// NewStatusBar crée une StatusBar.
func NewStatusBar(width int) StatusBar {
	return StatusBar{width: width}
}

// SetLeft définit le contenu gauche.
func (s *StatusBar) SetLeft(text string) {
	s.left = text
}

// SetRight définit le contenu droit.
func (s *StatusBar) SetRight(text string) {
	s.right = text
}

// SetWidth ajuste la largeur.
func (s *StatusBar) SetWidth(w int) {
	s.width = w
}

// View rend la barre de statut.
func (s StatusBar) View() string {
	baseStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#1F2937")).
		Foreground(lipgloss.Color("#6B7280")).
		Padding(0, 1)

	rightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#1F2937")).
		Foreground(lipgloss.Color("#9CA3AF")).
		Padding(0, 1)

	right := rightStyle.Render(s.right)
	rightWidth := lipgloss.Width(right)
	leftWidth := s.width - rightWidth
	if leftWidth < 1 {
		leftWidth = 1
	}

	left := baseStyle.Width(leftWidth).Render(s.left)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}
