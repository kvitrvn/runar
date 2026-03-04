package tui

import "strings"

// Command représente une commande exécutable via ':'.
type Command struct {
	Name        string
	Aliases     []string
	Description string
	View        ViewType
	IsQuit      bool
}

// commands liste toutes les commandes disponibles.
var commands = []Command{
	{Name: "pulse", Description: "Tableau de bord", View: ViewPulse},
	{Name: "clients", Description: "Liste des clients", View: ViewClients},
	{Name: "factures", Description: "Liste des factures", View: ViewInvoices},
	{Name: "devis", Description: "Liste des devis", View: ViewQuotes},
	{Name: "avoirs", Description: "Liste des avoirs", View: ViewCreditNotes},
	{Name: "quit", Aliases: []string{"q"}, Description: "Quitter l'application", IsQuit: true},
	{Name: "help", Aliases: []string{"h"}, Description: "Afficher l'aide"},
}

// ParseCommand recherche une commande par son nom ou alias.
// Retourne nil si non trouvée.
func ParseCommand(input string) *Command {
	input = strings.TrimSpace(strings.ToLower(input))
	for i, cmd := range commands {
		if cmd.Name == input {
			return &commands[i]
		}
		for _, alias := range cmd.Aliases {
			if alias == input {
				return &commands[i]
			}
		}
	}
	return nil
}

// Autocomplete retourne les commandes dont le nom commence par prefix.
func Autocomplete(prefix string) []Command {
	prefix = strings.ToLower(prefix)
	var matches []Command
	for _, cmd := range commands {
		if strings.HasPrefix(cmd.Name, prefix) {
			matches = append(matches, cmd)
		}
	}
	return matches
}

// AllCommands retourne toutes les commandes (pour l'aide).
func AllCommands() []Command {
	return commands
}
