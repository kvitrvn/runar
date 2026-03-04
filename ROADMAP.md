# 🗺️ ROADMAP.md - Plan de Développement

## 📊 Vue d'Ensemble

Développement en **7 sprints de 1-2 semaines** pour un MVP fonctionnel et conforme.

```
Phase 1: Fondations (S1-S2)     ████████████░░░░░░░░░░░░░░░░  40%
Phase 2: TUI Core (S3-S4)       ░░░░░░░░░░░░████████████░░░░  30%
Phase 3: Métier (S5-S6)         ░░░░░░░░░░░░░░░░░░░░████████  20%
Phase 4: Polish (S7)            ░░░░░░░░░░░░░░░░░░░░░░░░████  10%
```

---

## 🏗️ PHASE 1 : Fondations (Sprint 1-2)

### Sprint 1 : Setup Projet & Base de Données

**Durée** : 1 semaine
**Objectif** : Projet Go fonctionnel avec persistance SQLite

#### Tâches

- [ ] **Setup Projet Go**
  ```bash
  go mod init github.com/yourname/runar
  go get github.com/charmbracelet/bubbletea@latest
  go get github.com/mattn/go-sqlite3@latest
  go get github.com/shopspring/decimal@latest
  ```

- [ ] **Structure Dossiers**
  - Créer architecture complète selon ARCHITECTURE.md
  - Setup Makefile avec commandes de base
  - Configuration viper pour config.yaml

- [ ] **Base de Données SQLite**
  - [ ] Migration 001: Tables principales
    ```sql
    -- clients, invoices, invoice_lines, quotes, quote_lines
    ```
  - [ ] Migration 002: Audit trail
    ```sql
    -- audit_log table
    ```
  - [ ] Migration 003: Index et contraintes
  - [ ] Script d'init avec données de test

- [ ] **Repository Layer**
  - [ ] ClientRepository (CRUD complet)
  - [ ] InvoiceRepository (CRUD + helpers)
  - [ ] AuditRepository
  - [ ] Tests unitaires repositories

- [ ] **Validation Critère de Succès**
  - ✅ `make test` passe à 100%
  - ✅ DB créée avec données test
  - ✅ Repositories fonctionnels

### Sprint 2 : Domain & Service Layer

**Durée** : 1 semaine
**Objectif** : Logique métier avec règles légales

#### Tâches

- [ ] **Domain Models**
  - [ ] `client.go` avec validation
  - [ ] `invoice.go` avec états et règles
  - [ ] `quote.go`
  - [ ] `creditnote.go`
  - [ ] `validation.go` avec validateurs SIRET/SIREN
  - [ ] `numbering.go` générateur de numéros

- [ ] **Service Layer**
  - [ ] ClientService (CRUD)
  - [ ] InvoiceService avec règles légales :
    - [ ] Création avec numéro auto
    - [ ] Update avec guard immuabilité
    - [ ] MarkAsPaid avec verrouillage
    - [ ] Validation complète mentions
  - [ ] QuoteService
  - [ ] AuditService

- [ ] **Tests Règles Légales** 🔴 CRITIQUE
  - [ ] Test: Facture payée ne peut être modifiée
  - [ ] Test: Numérotation continue sans trou
  - [ ] Test: Validation toutes mentions obligatoires
  - [ ] Test: Calcul décimaux précis
  - [ ] Test: Audit log complet

- [ ] **Validation Critère de Succès**
  - ✅ Tous tests règles légales passent
  - ✅ Coverage > 80% sur domain/service
  - ✅ Impossible de modifier facture payée

---

## 🎨 PHASE 2 : TUI Core (Sprint 3-4)

### Sprint 3 : Application Bubbletea de Base

**Durée** : 1-2 semaines
**Objectif** : TUI fonctionnel avec navigation k9s-like

#### Tâches

- [ ] **App Structure**
  - [ ] `tui/app.go` : Modèle principal
  - [ ] `tui/context.go` : State management
  - [ ] `tui/navigation.go` : Système de navigation
  - [ ] Integration services dans TUI

- [ ] **Composants de Base**
  - [ ] `components/table.go` : Table interactive
  - [ ] `components/statusbar.go` : Barre de statut
  - [ ] `components/commandbar.go` : Barre de commande
  - [ ] `components/help.go` : Panneau d'aide

- [ ] **Styles Lipgloss**
  - [ ] `styles/theme.go` : Palette de couleurs
  - [ ] `styles/colors.go` : Couleurs par état
  - [ ] `styles/states.go` : Styles factures
  - [ ] Badges et indicateurs

- [ ] **Command System**
  - [ ] Parser commandes (`:clients`, `:factures`, etc.)
  - [ ] Autocomplétion
  - [ ] Historique commandes

- [ ] **Validation Critère de Succès**
  - ✅ App démarre sans erreur
  - ✅ Navigation clavier fonctionnelle
  - ✅ Command mode opérationnel

### Sprint 4 : Vues Clients & Factures

**Durée** : 1-2 semaines
**Objectif** : CRUD Clients et Factures complet dans TUI

#### Tâches

- [ ] **Vue Clients** (`:clients`)
  - [ ] Liste clients avec table
  - [ ] Formulaire nouveau client
  - [ ] Formulaire édition client
  - [ ] Vue détail client
  - [ ] Validation SIRET/SIREN temps réel
  - [ ] Recherche/filtre clients

- [ ] **Vue Factures** (`:factures`)
  - [ ] Liste factures avec états visuels
  - [ ] Formulaire nouvelle facture :
    - [ ] Sélection client
    - [ ] Multi-lignes dynamiques
    - [ ] Calcul automatique totaux
    - [ ] Validation temps réel
  - [ ] Vue détail facture
  - [ ] Badge immuable pour factures payées
  - [ ] Action "Marquer comme payée"
  - [ ] Recherche/filtre factures

- [ ] **Gestion Erreurs UI**
  - [ ] Messages erreur pour immuabilité
  - [ ] Affichage erreurs validation
  - [ ] Calcul et affichage amendes potentielles
  - [ ] Confirmations actions destructrices

- [ ] **Validation Critère de Succès**
  - ✅ Workflow complet création facture
  - ✅ Impossible modifier facture payée (UI + backend)
  - ✅ Tous états visuels corrects
  - ✅ Recherche/filtre fonctionnels

---

## 📄 PHASE 3 : Fonctionnalités Métier (Sprint 5-6)

### Sprint 5 : Génération PDF & Avoirs

**Durée** : 1-2 semaines
**Objectif** : Export PDF conforme + système d'avoirs

#### Tâches

- [ ] **Service PDF**
  - [ ] `pdf/generator.go` : Générateur Maroto
  - [ ] `pdf/invoice_template.go` : Template facture
    - [ ] Header avec infos vendeur
    - [ ] Bloc client
    - [ ] Tableau lignes
    - [ ] Totaux HT/TTC
    - [ ] Mentions légales obligatoires :
      - [ ] TVA non applicable
      - [ ] Pénalités retard
      - [ ] Indemnité forfaitaire
      - [ ] SIREN client si B2B
  - [ ] Tests génération PDF

- [ ] **Système Avoirs**
  - [ ] `domain/creditnote.go`
  - [ ] CreditNoteService :
    - [ ] Création avoir depuis facture payée
    - [ ] Validation référence facture
    - [ ] Numérotation séparée
  - [ ] Vue TUI avoirs
  - [ ] Génération PDF avoir

- [ ] **Integration TUI**
  - [ ] Action "p" : Générer PDF
  - [ ] Action "v" : Ouvrir PDF (xdg-open)
  - [ ] Action "c" : Créer avoir (factures payées)
  - [ ] Vue liste avoirs

- [ ] **Validation Critère de Succès**
  - ✅ PDF facture conforme (toutes mentions)
  - ✅ Avoir créable uniquement depuis facture payée
  - ✅ PDF avoir correct avec référence

### Sprint 6 : Devis & Conversion

**Durée** : 1-2 semaines
**Objectif** : Gestion complète devis

#### Tâches

- [ ] **Service Devis**
  - [ ] QuoteService complet
  - [ ] Gestion états (draft, sent, accepted, refused, expired)
  - [ ] Conversion devis → facture :
    - [ ] Copie lignes
    - [ ] Génération nouveau numéro facture
    - [ ] Lien devis → facture
    - [ ] Marquage devis comme "accepté"

- [ ] **Vue TUI Devis** (`:devis`)
  - [ ] Liste devis avec filtres
  - [ ] Formulaire nouveau devis
  - [ ] Vue détail devis
  - [ ] Action conversion en facture
  - [ ] Indicateur expiration

- [ ] **PDF Devis**
  - [ ] Template PDF devis
  - [ ] Mentions spécifiques devis :
    - [ ] Date expiration
    - [ ] Conditions acceptation
    - [ ] "Devis valable jusqu'au..."

- [ ] **Validation Critère de Succès**
  - ✅ Workflow complet devis
  - ✅ Conversion devis → facture seamless
  - ✅ Numérotation facture correcte après conversion
  - ✅ Lien traçable devis ↔ facture

---

## 🎯 PHASE 4 : Polish & Préparation Production (Sprint 7)

### Sprint 7 : Dashboard, Export & Doc

**Durée** : 1 semaine
**Objectif** : Finitions, documentation, préparation release

#### Tâches

- [ ] **Dashboard** (`:pulse`)
  - [ ] Widget CA annuel
  - [ ] Widget factures par état
  - [ ] Widget alertes :
    - [ ] Factures échues
    - [ ] Devis expirant
    - [ ] Seuils TVA
    - [ ] Facturation électronique 2027
  - [ ] Graphique CA mensuel (ASCII)
  - [ ] Indicateur seuil TVA

- [ ] **Export Comptable**
  - [ ] Export CSV factures
  - [ ] Export CSV avoirs
  - [ ] Format compatible Excel
  - [ ] Filtres période

- [ ] **Configuration**
  - [ ] Fichier config.yaml :
    ```yaml
    seller:
      name: "Votre Nom"
      siret: "12345678901234"
      address: "123 Rue..."
      city: "Paris"
      postal_code: "75001"
      email: "contact@..."
      phone: "+33..."
    
    vat:
      applicable: false
      exemption_text: "TVA non applicable, article 293B du CGI"
    
    payment:
      default_deadline: "30 jours"
      late_penalty_rate: 13.25
      recovery_fee: 40
    
    database:
      path: "./runar.db"
    
    pdf:
      output_dir: "./invoices"
    ```
  - [ ] Wizard premier lancement

- [ ] **Documentation**
  - [ ] README.md complet :
    - [ ] Installation
    - [ ] Utilisation
    - [ ] Screenshots
    - [ ] Troubleshooting
  - [ ] Guide utilisateur (docs/user-guide.md)
  - [ ] Changelog
  - [ ] LICENSE

- [ ] **Tests E2E**
  - [ ] Scénario complet :
    1. Créer client
    2. Créer devis
    3. Convertir en facture
    4. Marquer comme payée
    5. Créer avoir
    6. Générer tous les PDFs
    7. Export CSV

- [ ] **Build & Packaging**
  - [ ] Makefile targets :
    ```makefile
    build:        # Build binaire
    install:      # Install dans $GOPATH/bin
    test:         # Tous les tests
    test-legal:   # Tests règles légales
    lint:         # Linter
    clean:        # Nettoyage
    release:      # Build multi-platform
    ```
  - [ ] GitHub Actions CI/CD (optionnel)
  - [ ] Releases GitHub avec binaires

- [ ] **Validation Finale**
  - ✅ Tous tests passent
  - ✅ Coverage > 80%
  - ✅ Linter sans warning
  - ✅ Documentation complète
  - ✅ Binaire fonctionnel standalone

---

## 🚀 Post-MVP (Optionnel)

### Améliorations Futures

- **Phase 5** : Email & Cloud
  - [ ] Envoi factures par email
  - [ ] Backup cloud (S3, Dropbox)
  - [ ] Sync multi-devices

- **Phase 6** : Analytics Avancés
  - [ ] Graphiques interactifs (termgraph)
  - [ ] Prévisions CA
  - [ ] Analyse rentabilité par client

- **Phase 7** : Facturation Électronique 2027
  - [ ] Génération Factur-X (PDF + XML)
  - [ ] Intégration PDP/PPF
  - [ ] E-reporting automatique
  - [ ] Signature électronique

- **Phase 8** : Multi-Utilisateurs
  - [ ] Authentification
  - [ ] Gestion permissions
  - [ ] Audit trail par user

---

## 📝 Checklist Pré-Release

### Qualité Code

- [ ] Tous tests passent
- [ ] Coverage > 80% (domain, service, repository)
- [ ] Aucun warning linter
- [ ] Documentation Godoc complète
- [ ] README à jour

### Conformité Légale

- [ ] Immuabilité factures payées ✅
- [ ] Numérotation continue ✅
- [ ] Toutes mentions obligatoires ✅
- [ ] Conservation 10 ans ✅
- [ ] Audit trail ✅
- [ ] Calculs décimaux précis ✅

### UX/UI

- [ ] Tous raccourcis fonctionnels
- [ ] Aide contextuelle complète
- [ ] Messages erreur clairs
- [ ] Responsive (80+ colonnes)
- [ ] Performance <100ms

### Packaging

- [ ] Binaire Linux 64-bit
- [ ] Binaire macOS (ARM + Intel)
- [ ] Binaire Windows (optionnel)
- [ ] Installation simple
- [ ] Config wizard

---

## 📊 Métriques de Succès

### Technique

| Métrique | Cible | Statut |
|----------|-------|--------|
| Coverage Tests | >80% | ⏳ |
| Temps Build | <30s | ⏳ |
| Taille Binaire | <20MB | ⏳ |
| Startup Time | <500ms | ⏳ |

### Fonctionnel

| Feature | Priorité | Statut |
|---------|----------|--------|
| CRUD Clients | P0 | ⏳ |
| CRUD Factures | P0 | ⏳ |
| Génération PDF | P0 | ⏳ |
| Immuabilité | P0 | ⏳ |
| Avoirs | P1 | ⏳ |
| Devis | P1 | ⏳ |
| Dashboard | P2 | ⏳ |
| Export CSV | P2 | ⏳ |

### Légal

| Règle | Implémenté | Testé |
|-------|------------|-------|
| Mentions obligatoires | ⏳ | ⏳ |
| Numérotation continue | ⏳ | ⏳ |
| Immuabilité payées | ⏳ | ⏳ |
| SIREN B2B (2024) | ⏳ | ⏳ |
| Conservation 10 ans | ⏳ | ⏳ |
| Audit trail | ⏳ | ⏳ |

---

## 🎯 Priorités par Sprint

### Sprint 1-2 : 🔴 Critique
- Base de données
- Repositories
- Services avec règles légales
- Tests immuabilité

### Sprint 3-4 : 🟠 Important
- TUI fonctionnel
- Navigation
- CRUD UI Clients/Factures
- États visuels

### Sprint 5-6 : 🟡 Nécessaire
- PDF conforme
- Avoirs
- Devis
- Conversion

### Sprint 7 : 🟢 Nice to Have
- Dashboard
- Export
- Documentation
- Polish

---

## 📅 Timeline Estimée

```
S1  ████ Fondations - DB & Repos
S2  ████ Fondations - Domain & Services
S3  ████ TUI - App Structure
S4  ████ TUI - Vues Clients/Factures
S5  ████ Métier - PDF & Avoirs
S6  ████ Métier - Devis
S7  ████ Polish - Dashboard & Doc
    ├────────────────────────────┤
    0                      14 semaines
```

**Durée Totale Estimée** : 10-14 semaines (2.5-3.5 mois)
**Effort** : 1 développeur à temps partiel (20h/semaine)

---

## 🎓 Pour Claude Code

### Commencer par...

1. **Lire** : CLAUDE.md → LEGAL.md → ARCHITECTURE.md → ROADMAP.md (ce fichier)
2. **Sprint 1** : Créer structure projet + DB
3. **Tests d'abord** : TDD pour règles légales critiques
4. **Itérer** : Un sprint à la fois

### À chaque commit

```
git commit -m "feat(invoice): implement immutability guard

[LEGAL] Implements article L441-9 compliance
Paid invoices can no longer be modified

- Add InvoiceState enum
- Add CanEdit() method with guards
- Add tests for immutability
- Update service layer to block updates
```

### Si bloqué

1. Revenir à LEGAL.md pour clarifier règle
2. Vérifier ARCHITECTURE.md pour pattern
3. Demander clarification sur règle légale
4. Ne JAMAIS skipper une règle légale

---

**Dernière Mise à Jour** : 2026-03-04
**Prochaine Révision** : Fin Sprint 2
