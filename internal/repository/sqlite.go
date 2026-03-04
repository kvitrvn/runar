package repository

import (
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // Driver SQLite
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// InitDB ouvre la base de données SQLite et applique les migrations.
// LEGAL: La base doit exister et être accessible pour garantir la conservation 10 ans.
func InitDB(dbPath string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", fmt.Sprintf("%s?_foreign_keys=on&_journal_mode=WAL", dbPath))
	if err != nil {
		return nil, fmt.Errorf("ouverture base de données: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("connexion base de données: %w", err)
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}

	return db, nil
}

// runMigrations applique les migrations SQL dans l'ordre.
func runMigrations(db *sqlx.DB) error {
	// Créer table de suivi des migrations
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("création table migrations: %w", err)
	}

	// Lister les migrations disponibles
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("lecture migrations: %w", err)
	}

	// Trier par nom (ordre alphabétique = ordre numérique)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		version := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))

		// Vérifier si déjà appliquée
		var count int
		err := db.Get(&count, "SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version)
		if err != nil {
			return fmt.Errorf("vérification migration %s: %w", version, err)
		}
		if count > 0 {
			continue // Déjà appliquée
		}

		// Lire et exécuter
		sql, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("lecture migration %s: %w", entry.Name(), err)
		}

		if _, err := db.Exec(string(sql)); err != nil {
			return fmt.Errorf("exécution migration %s: %w", entry.Name(), err)
		}

		// Enregistrer comme appliquée
		if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			return fmt.Errorf("enregistrement migration %s: %w", version, err)
		}
	}

	return nil
}

// Repositories regroupe tous les repositories de l'application.
type Repositories struct {
	Client     ClientRepository
	Invoice    InvoiceRepository
	Quote      QuoteRepository
	CreditNote CreditNoteRepository
	Audit      AuditRepository
}

// NewRepositories crée tous les repositories.
func NewRepositories(db *sqlx.DB) *Repositories {
	return &Repositories{
		Client:     NewClientRepository(db),
		Invoice:    NewInvoiceRepository(db),
		Quote:      NewQuoteRepository(db),
		CreditNote: NewCreditNoteRepository(db),
		Audit:      NewAuditRepository(db),
	}
}
