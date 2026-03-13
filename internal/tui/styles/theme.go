package styles

import "github.com/charmbracelet/lipgloss"

// Styles des zones de layout.
var (
	StyleHeader = lipgloss.NewStyle().
			Background(ColorBackground).
			Foreground(ColorForeground).
			Bold(true).
			Padding(0, 1)

	StyleCommandBar = lipgloss.NewStyle().
			Background(lipgloss.Color("#111827")).
			Foreground(ColorForeground).
			Padding(0, 1)

	StyleInfoBar = lipgloss.NewStyle().
			Background(ColorBackground).
			Foreground(ColorMuted).
			Padding(0, 1)

	StyleMainBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted)

	StyleTitle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	StyleMuted = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleSuccess = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	StyleDanger = lipgloss.NewStyle().
			Foreground(ColorDanger)

	StyleWarning = lipgloss.NewStyle().
			Foreground(ColorWarning)

	// Styles états de facture.
	StyleDraft = lipgloss.NewStyle().
			Foreground(ColorStateDraft)

	StyleIssued = lipgloss.NewStyle().
			Foreground(ColorStateIssued).
			Bold(true)

	StyleSent = lipgloss.NewStyle().
			Foreground(ColorStateSent).
			Bold(true)

	StylePaid = lipgloss.NewStyle().
			Foreground(ColorStatePaid).
			Bold(true)

	StyleOverdue = lipgloss.NewStyle().
			Foreground(ColorStateOverdue).
			Bold(true)

	StyleCanceled = lipgloss.NewStyle().
			Foreground(ColorStateCanceled)

	// Style erreur immuabilité (fond rouge foncé).
	StyleImmutableBanner = lipgloss.NewStyle().
				Foreground(ColorForeground).
				Background(lipgloss.Color("#7F1D1D")).
				Bold(true).
				Padding(0, 1)

	// Help panel.
	StyleHelpPanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSecondary).
			Padding(1, 2)

	StyleHelpCategory = lipgloss.NewStyle().
				Foreground(ColorSecondary).
				Bold(true)

	StyleHelpKey = lipgloss.NewStyle().
			Foreground(ColorPrimary)

	// Toast styles.
	StyleToastSuccess = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Background(lipgloss.Color("#065F46")).
				Padding(0, 2)

	StyleToastError = lipgloss.NewStyle().
			Foreground(ColorForeground).
			Background(ColorDanger).
			Padding(0, 2)

	StyleToastWarning = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1C1917")).
				Background(ColorWarning).
				Padding(0, 2)

	StyleToastInfo = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0C4A6E")).
			Background(ColorPrimary).
			Padding(0, 2)
)
