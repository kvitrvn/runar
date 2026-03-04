package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kvitrvn/runar/internal/config"
	"github.com/kvitrvn/runar/internal/repository"
	"github.com/kvitrvn/runar/internal/service"
	"github.com/kvitrvn/runar/internal/tui"
)

// Version est injectée à la compilation via -ldflags.
var Version = "dev"

func main() {
	initConfig := flag.Bool("init-config", false, "Crée un fichier config.yaml d'exemple dans le répertoire courant")
	showVersion := flag.Bool("version", false, "Affiche la version")
	flag.Parse()

	if *showVersion {
		fmt.Println("runar", Version)
		return
	}

	// Wizard premier lancement : créer config.yaml si absent
	if *initConfig {
		if _, err := os.Stat("config.yaml"); err == nil {
			fmt.Println("⚠  config.yaml existe déjà. Supprimez-le d'abord si vous souhaitez le réinitialiser.")
			os.Exit(1)
		}
		if err := config.WriteDefault("config.yaml"); err != nil {
			log.Fatalf("Impossible de créer config.yaml: %v", err)
		}
		fmt.Println("✓ config.yaml créé avec succès.")
		fmt.Println("  Éditez-le avec vos informations (SIRET, nom, adresse, IBAN...)")
		fmt.Println("  puis relancez : runar")
		return
	}

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
