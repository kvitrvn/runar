// Package tui implémente l'interface utilisateur TUI de l'application.
// Sprint 3: Implémentation complète avec Bubbletea.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kvitrvn/runar/internal/config"
	"github.com/kvitrvn/runar/internal/service"
)

// App est le modèle principal de l'application TUI.
// TODO Sprint 3: Implémenter la navigation k9s-like complète.
type App struct {
	services *service.Services
	config   *config.Config
	width    int
	height   int
}

// NewApp crée l'application TUI.
func NewApp(services *service.Services, cfg *config.Config) *App {
	return &App{
		services: services,
		config:   cfg,
	}
}

// Init implémente tea.Model.
func (m *App) Init() tea.Cmd {
	return nil
}

// Update implémente tea.Model.
func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View implémente tea.Model.
func (m *App) View() string {
	return "AutoGest - TUI en cours de développement (Sprint 3)\n\nAppuyez sur 'q' pour quitter."
}
