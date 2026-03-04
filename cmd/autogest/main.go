package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kvitrvn/runar/internal/config"
	"github.com/kvitrvn/runar/internal/repository"
	"github.com/kvitrvn/runar/internal/service"
	"github.com/kvitrvn/runar/internal/tui"
)

func main() {
	// 1. Charger configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erreur configuration: %v", err)
	}

	// 2. Initialiser base de données avec migrations
	db, err := repository.InitDB(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Erreur base de données: %v", err)
	}
	defer db.Close()

	// 3. Créer repositories
	repos := repository.NewRepositories(db)

	// 4. Créer services
	services := service.NewServices(repos, cfg)

	// 5. Créer et lancer TUI
	app := tui.NewApp(services, cfg)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Erreur TUI: %v", err)
	}
}
