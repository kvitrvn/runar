package styles

import "github.com/charmbracelet/lipgloss"

// Palette principale.
var (
	ColorPrimary    = lipgloss.Color("#0EA5E9") // Bleu ciel
	ColorSecondary  = lipgloss.Color("#8B5CF6") // Violet
	ColorSuccess    = lipgloss.Color("#10B981") // Vert
	ColorWarning    = lipgloss.Color("#F59E0B") // Orange
	ColorDanger     = lipgloss.Color("#EF4444") // Rouge
	ColorMuted      = lipgloss.Color("#6B7280") // Gris
	ColorBackground = lipgloss.Color("#1F2937") // Gris foncé
	ColorForeground = lipgloss.Color("#F9FAFB") // Blanc cassé

	// Couleurs par état de facture.
	ColorStateDraft    = ColorMuted
	ColorStateIssued   = ColorPrimary
	ColorStateSent     = ColorSecondary
	ColorStatePaid     = ColorSuccess
	ColorStateOverdue  = ColorDanger
	ColorStateCanceled = ColorMuted
)
