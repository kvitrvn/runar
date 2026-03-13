package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kvitrvn/runar/internal/tui/styles"
)

// KeyBinding représente un raccourci clavier avec sa description.
type KeyBinding struct {
	Key         string
	Description string
	Category    string
}

// globalBindings liste les raccourcis globaux.
var globalBindings = []KeyBinding{
	{Key: ":", Description: "Mode commande", Category: "Navigation"},
	{Key: "/", Description: "Recherche / filtre", Category: "Navigation"},
	{Key: "?", Description: "Afficher / masquer l'aide", Category: "Navigation"},
	{Key: "ESC", Description: "Annuler / retour", Category: "Navigation"},
	{Key: "Tab", Description: "Vue suivante", Category: "Navigation"},
	{Key: "q / Ctrl+c", Description: "Quitter", Category: "Navigation"},
	{Key: "j / ↓", Description: "Ligne suivante", Category: "Liste"},
	{Key: "k / ↑", Description: "Ligne précédente", Category: "Liste"},
	{Key: "g", Description: "Première ligne", Category: "Liste"},
	{Key: "G", Description: "Dernière ligne", Category: "Liste"},
	{Key: "Ctrl+d", Description: "Page suivante", Category: "Liste"},
	{Key: "Ctrl+u", Description: "Page précédente", Category: "Liste"},
	{Key: "Enter", Description: "Voir détail", Category: "Liste"},
}

// invoiceBindings liste les raccourcis spécifiques aux factures.
var invoiceBindings = []KeyBinding{
	{Key: "n", Description: "Nouvelle facture", Category: "Factures"},
	{Key: "i", Description: "Émettre (brouillon → émise)", Category: "Factures"},
	{Key: "s", Description: "Marquer envoyée (brouillon/émise → envoyée)", Category: "Factures"},
	{Key: "m", Description: "Marquer comme payée (émise/envoyée)", Category: "Factures"},
	{Key: "e", Description: "Éditer (si non payée / annulée)", Category: "Factures"},
	{Key: "d", Description: "Supprimer (si brouillon uniquement)", Category: "Factures"},
	{Key: "p", Description: "Générer PDF", Category: "Factures"},
	{Key: "c", Description: "Créer avoir (si émise / envoyée / payée)", Category: "Factures"},
}

// clientBindings liste les raccourcis spécifiques aux clients.
var clientBindings = []KeyBinding{
	{Key: "n", Description: "Nouveau client", Category: "Clients"},
	{Key: "e", Description: "Éditer le client sélectionné", Category: "Clients"},
	{Key: "d", Description: "Supprimer le client sélectionné", Category: "Clients"},
}

// quoteBindings liste les raccourcis spécifiques aux devis.
var quoteBindings = []KeyBinding{
	{Key: "n", Description: "Nouveau devis", Category: "Devis"},
	{Key: "e", Description: "Éditer le devis brouillon", Category: "Devis"},
	{Key: "d", Description: "Supprimer le devis brouillon", Category: "Devis"},
	{Key: "s", Description: "Marquer envoyé", Category: "Devis"},
	{Key: "a", Description: "Accepter le devis", Category: "Devis"},
	{Key: "r", Description: "Refuser le devis", Category: "Devis"},
	{Key: "f", Description: "Convertir en facture (si accepté)", Category: "Devis"},
	{Key: "p", Description: "Générer PDF", Category: "Devis"},
}

// creditNoteBindings liste les raccourcis spécifiques aux avoirs.
var creditNoteBindings = []KeyBinding{
	{Key: "p", Description: "Générer PDF", Category: "Avoirs"},
}

// RenderHelpPanel rend le panneau d'aide (overlay).
func RenderHelpPanel(width int) string {
	panelWidth := 62
	if width < panelWidth+4 {
		panelWidth = width - 4
	}

	var sb strings.Builder
	sb.WriteString(styles.StyleTitle.Render("AIDE — Raccourcis Clavier") + "\n\n")

	// Regrouper par catégorie
	categoryOrder := []string{"Navigation", "Liste", "Factures", "Clients", "Devis", "Avoirs"}
	allBindings := append(globalBindings, invoiceBindings...)
	allBindings = append(allBindings, clientBindings...)
	allBindings = append(allBindings, quoteBindings...)
	allBindings = append(allBindings, creditNoteBindings...)
	categories := make(map[string][]KeyBinding)
	for _, kb := range allBindings {
		categories[kb.Category] = append(categories[kb.Category], kb)
	}

	for _, cat := range categoryOrder {
		bindings, ok := categories[cat]
		if !ok {
			continue
		}
		sb.WriteString(styles.StyleHelpCategory.Render(cat) + "\n")
		for _, kb := range bindings {
			sb.WriteString(fmt.Sprintf("  %s  %s\n",
				styles.StyleHelpKey.Width(14).Render(kb.Key),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#D1D5DB")).Render(kb.Description),
			))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(styles.StyleMuted.Render("Commandes : :pulse  :clients  :factures  :devis  :avoirs  :quit"))

	return styles.StyleHelpPanel.Width(panelWidth).Render(sb.String())
}
