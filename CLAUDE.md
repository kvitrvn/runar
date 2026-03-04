# AutoGest - Application TUI de Gestion Auto-Entreprise

## 🎯 Vision du Projet

AutoGest est une application TUI (Text User Interface) en Go inspirée de k9s, conçue pour gérer une auto-entreprise française : clients, devis, factures et avoirs, avec génération PDF et respect strict des obligations légales françaises.

## 🏛️ Conformité Légale ABSOLUE

**PRIORITÉ MAXIMALE** : Ce projet doit respecter 100% des obligations légales françaises pour les auto-entrepreneurs. Toute fonctionnalité doit être validée contre les règles du Code de Commerce et du Code Général des Impôts.

### Règles Critiques NON-NÉGOCIABLES

1. **IMMUABILITÉ DES FACTURES PAYÉES**
   - Une facture marquée comme "payée" NE PEUT JAMAIS être modifiée
   - Seule action possible : créer un avoir (credit note)
   - Toute tentative de modification doit être bloquée avec message d'erreur explicite
   - Implémenter des guards dans le code pour empêcher toute modification

2. **NUMÉROTATION CONTINUE**
   - Les numéros de facture doivent suivre une séquence continue sans trou
   - Format : ANNÉE-SEQUENCE (ex: 2026-0001, 2026-0002)
   - Pas de suppression de numéro
   - Générateur automatique avec vérification de continuité

3. **MENTIONS OBLIGATOIRES**
   - Toutes les mentions listées dans LEGAL.md doivent être présentes
   - Validation automatique avant sauvegarde
   - Alertes visuelles pour mentions manquantes
   - Calcul automatique de l'amende potentielle (15€ par mention manquante)

4. **CONSERVATION**
   - Toutes les factures doivent être conservées 10 ans
   - Les PDFs générés ne doivent JAMAIS être supprimés
   - Implémenter un système d'archivage avec timestamps

5. **AUDIT TRAIL**
   - Toute action sur une facture doit être loggée
   - Qui a fait quoi, quand, ancienne et nouvelle valeur
   - Table audit_log obligatoire

## 📁 Structure du Projet

```
runar/
├── CLAUDE.md                    # Ce fichier - Vision et directives
├── LEGAL.md                     # Règles légales détaillées
├── ARCHITECTURE.md              # Architecture technique
├── TUI-DESIGN.md               # Spécifications UI/UX
├── API.md                      # Documentation API interne
├── ROADMAP.md                  # Plan de développement
├── cmd/
│   └── runar/
│       └── main.go
├── internal/
│   ├── tui/                    # Interface Bubbletea
│   │   ├── app.go              # Point d'entrée TUI
│   │   ├── context.go          # State management global
│   │   ├── navigation.go       # Système de navigation k9s-like
│   │   ├── views/              # Vues principales
│   │   │   ├── clients.go      # Vue clients
│   │   │   ├── invoices.go     # Vue factures
│   │   │   ├── quotes.go       # Vue devis
│   │   │   ├── creditnotes.go  # Vue avoirs
│   │   │   ├── dashboard.go    # Dashboard
│   │   │   └── help.go         # Aide contextuelle
│   │   ├── components/         # Composants réutilisables
│   │   │   ├── table.go        # Table avec tri/filtre
│   │   │   ├── form.go         # Formulaires
│   │   │   ├── detail.go       # Vue détail
│   │   │   ├── statusbar.go    # Barre de statut
│   │   │   └── commandbar.go   # Barre de commande
│   │   └── styles/             # Thèmes Lipgloss
│   │       ├── theme.go        # Thème principal
│   │       ├── colors.go       # Palette de couleurs
│   │       └── states.go       # Styles par état
│   ├── domain/                 # Logique métier
│   │   ├── client.go           # Entité Client
│   │   ├── invoice.go          # Entité Facture
│   │   ├── quote.go            # Entité Devis
│   │   ├── creditnote.go       # Entité Avoir
│   │   ├── line.go             # Lignes de facture/devis
│   │   ├── validation.go       # Validations légales
│   │   ├── numbering.go        # Générateur de numéros
│   │   └── states.go           # Machine à états
│   ├── service/                # Services métier
│   │   ├── client_service.go
│   │   ├── invoice_service.go  # Avec règles légales
│   │   ├── quote_service.go
│   │   ├── audit_service.go    # Service d'audit
│   │   └── pdf_service.go      # Génération PDF
│   ├── repository/             # Couche données
│   │   ├── sqlite.go           # Connexion SQLite
│   │   ├── client_repo.go
│   │   ├── invoice_repo.go
│   │   ├── quote_repo.go
│   │   ├── audit_repo.go
│   │   └── migrations/         # Migrations SQL
│   │       ├── 001_initial.sql
│   │       ├── 002_audit.sql
│   │       └── 003_legal_2027.sql
│   ├── pdf/                    # Génération PDF
│   │   ├── generator.go        # Générateur principal
│   │   ├── invoice_template.go # Template facture
│   │   ├── quote_template.go   # Template devis
│   │   └── styles.go           # Styles PDF
│   └── config/
│       ├── config.go           # Configuration app
│       └── seller.go           # Infos auto-entrepreneur
├── testdata/                   # Données de test
│   ├── clients.json
│   └── invoices.json
├── docs/                       # Documentation
│   ├── screenshots/
│   └── user-guide.md
├── go.mod
├── go.sum
├── Makefile                    # Commandes build/test
└── README.md                   # Readme public
```

## 🎨 Principes de Design TUI

### Navigation k9s-like

L'application doit reproduire l'expérience de navigation de k9s :

1. **Command Mode** : `:` pour entrer des commandes (`:clients`, `:factures`, etc.)
2. **Search Mode** : `/` pour filtrer les listes
3. **Vim Bindings** : `j/k` pour naviguer, `g/G` pour début/fin
4. **Context Awareness** : Aide contextuelle avec `?`
5. **Visual Feedback** : Codes couleur pour états, badges, alertes

### Composants Principaux

- **Header** : Contexte, mode, année fiscale
- **Command Bar** : Commande active + suggestions
- **Main Pane** : Liste/Tableau/Détails
- **Info Bar** : Stats, filtres actifs, messages
- **Help Panel** : Raccourcis disponibles (toggle avec `?`)

### États Visuels

- 🟢 **Vert** : Payé, Accepté, Validé
- 🟡 **Jaune** : Brouillon, En attente
- 🔴 **Rouge** : Échu, Refusé, Erreur
- 🔒 **Badge Lock** : Facture payée (immuable)
- ⚠️ **Badge Warning** : Validation manquante

## 🛠️ Stack Technique

### Dépendances Principales

```go
// TUI Framework
github.com/charmbracelet/bubbletea
github.com/charmbracelet/bubbles
github.com/charmbracelet/lipgloss

// Database
github.com/mattn/go-sqlite3
github.com/jmoiron/sqlx

// PDF Generation
github.com/johnfercher/maroto/v2

// Decimal Precision (CRUCIAL pour argent)
github.com/shopspring/decimal

// Validation
github.com/go-playground/validator/v10

// Configuration
github.com/spf13/viper
```

### Patterns Obligatoires

1. **Repository Pattern** : Abstraire la persistence
2. **Service Layer** : Logique métier séparée
3. **Validation Layer** : Règles légales isolées
4. **Audit Layer** : Traçabilité systématique
5. **Immutability Guards** : Protection anti-modification

## 🚦 Workflow de Développement

### Phase 1 : Fondations (Sprint 1-2)

1. Setup projet Go avec structure complète
2. Migrations SQLite avec toutes les tables
3. Modèles de domaine avec validations
4. Repository layer pour Client et Invoice
5. Service layer avec règles légales
6. Tests unitaires pour règles critiques

### Phase 2 : TUI Core (Sprint 3-4)

1. App Bubbletea de base avec navigation
2. Vue Clients (liste + CRUD)
3. Vue Factures (liste + détails)
4. Command bar avec autocomplétion
5. Styles et thème
6. Aide contextuelle

### Phase 3 : Fonctionnalités Métier (Sprint 5-6)

1. Génération PDF factures
2. Système d'avoirs
3. Gestion devis
4. Conversion devis → facture
5. Validation légale complète
6. Audit trail

### Phase 4 : Polish (Sprint 7)

1. Dashboard analytics
2. Export comptable
3. Documentation complète
4. Tests E2E
5. Préparation facturation électronique 2027

## 🧪 Tests Obligatoires

### Tests Critiques

1. **Immuabilité** : Tenter de modifier facture payée doit échouer
2. **Numérotation** : Pas de trous dans la séquence
3. **Validation** : Toutes mentions obligatoires présentes
4. **Calculs** : Précision décimale pour montants
5. **Audit** : Toutes actions loggées

### Commandes Make

```makefile
make test           # Tous les tests
make test-legal     # Tests règles légales uniquement
make test-coverage  # Coverage report
make lint           # Linter
make build          # Build binaire
make run            # Run en mode dev
```

## 📝 Conventions de Code

### Nommage

- **Entités** : Singulier, PascalCase (`Invoice`, `Client`)
- **Services** : Suffix `Service` (`InvoiceService`)
- **Repos** : Suffix `Repository` (`InvoiceRepository`)
- **Constantes** : UPPER_SNAKE_CASE pour états (`INVOICE_PAID`)
- **Erreurs** : Prefix `Err` (`ErrImmutableInvoice`)

### Commentaires

```go
// LEGAL: Explication de la règle légale
// WHY: Justification d'une décision technique
// TODO: Tâche à faire
// FIXME: Bug à corriger
// HACK: Solution temporaire
```

### Gestion Erreurs

```go
// Erreurs métier typées
type ImmutableInvoiceError struct {
    InvoiceNumber string
    Message       string
}

func (e *ImmutableInvoiceError) Error() string {
    return fmt.Sprintf("❌ FACTURE IMMUABLE %s: %s", e.InvoiceNumber, e.Message)
}
```

## 🎯 Objectifs de Qualité

- ✅ **Coverage** : >80% pour domain et service layers
- ✅ **Linting** : 0 warning sur golangci-lint
- ✅ **Documentation** : Godoc pour toutes les fonctions publiques
- ✅ **Performance** : <100ms pour toute opération UI
- ✅ **Sécurité** : Pas de secrets en dur, validation inputs

## 📚 Références

- [Code de Commerce - Facturation](https://www.legifrance.gouv.fr/)
- [CGI Article 293B - Franchise TVA](https://www.legifrance.gouv.fr/)
- [Bubbletea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [k9s Source Code](https://github.com/derailed/k9s)

## 🚨 Red Flags - À NE JAMAIS FAIRE

❌ Modifier une facture payée
❌ Supprimer un numéro de facture
❌ Stocker les montants en float (utiliser decimal.Decimal)
❌ Oublier une mention légale obligatoire
❌ Skip les validations "pour aller vite"
❌ Hardcoder des valeurs qui devraient être en config
❌ Ignorer les erreurs de génération PDF
❌ Ne pas logger les actions critiques

## 💡 Best Practices

✅ Toujours valider avant persist
✅ Utiliser des transactions pour opérations multi-tables
✅ Logger toutes les actions métier importantes
✅ Préfixer les messages d'erreur avec emojis pour clarté
✅ Utiliser des types forts (pas de string pour états)
✅ Documenter les règles légales dans le code
✅ Tests pour chaque règle légale
✅ Backup automatique de la DB

## 🎓 Pour Claude Code

Quand tu travailles sur ce projet :

1. **Commence TOUJOURS par lire** : CLAUDE.md → LEGAL.md → ARCHITECTURE.md
2. **Vérifie la conformité légale** avant toute implémentation
3. **Pose des questions** si une règle légale n'est pas claire
4. **Priorise la sécurité juridique** sur la rapidité de dev
5. **Tests d'abord** pour les règles critiques
6. **Documentation inline** pour les règles légales
7. **Commit messages** descriptifs avec contexte légal si applicable

### Format Commit Messages

```
type(scope): description courte

[LEGAL] si impact légal
[BREAKING] si breaking change

Explication détaillée
Référence légale si applicable
```

Exemples :
```
feat(invoice): add immutability guard for paid invoices

[LEGAL] Implements article L441-9 compliance
Paid invoices can no longer be modified, only credited

fix(validation): add missing SIREN validation for B2B clients

[LEGAL] Required since 2024 decree
```

## 🔄 État du Projet

**Version Actuelle** : 0.0.0 (non démarré)
**Phase** : Phase 1 - Fondations
**Prochaine Milestone** : Setup projet + DB schema

---

**Dernière mise à jour** : 2026-03-04
**Mainteneur** : Vous
**Licence** : À définir
