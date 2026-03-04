# 🎨 TUI-DESIGN.md - Design de l'Interface Utilisateur

## 🎯 Philosophie de Design

AutoGest s'inspire de **k9s** pour créer une expérience utilisateur :
- ⚡ **Rapide** : Navigation 100% clavier
- 🎯 **Efficace** : Raccourcis intuitifs (Vim-like)
- 📊 **Informative** : État visible en un coup d'œil
- 🎨 **Claire** : Codes couleur cohérents
- 💪 **Puissante** : Filtres, tris, recherche

---

## 📐 Layout Global

### Structure de l'Écran

```
┌─────────────────────────────────────────────────────────────────┐
│ 🏢 AutoGest | 2026 | Context: Tous clients | Mode: Normal     │ Header
├─────────────────────────────────────────────────────────────────┤
│ :factures                                          [F1] Aide   │ Command Bar
├─────────────────────────────────────────────────────────────────┤
│ ┌─ FACTURES ─────────────────────────────────────────────────┐ │
│ │ NUM         CLIENT         MONTANT  ÉTAT     ÉCHÉANCE      │ │
│ │ ────────────────────────────────────────────────────────── │ │
│ │ 2026-0001   Acme Corp      1,250€   PAYÉE    15/02/2026 ✓ │ │ Main Pane
│ │>2026-0002   Digital SA       890€   ÉCHUE    01/03/2026 ⚠ │ │ (Table)
│ │ 2026-0003   Solutions      2,340€   ÉMISE    20/03/2026   │ │
│ │ 2026-0004   StartUp Inc      450€   BROUILLON 25/03/2026  │ │
│ │                                                            │ │
│ └────────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│ 📊 4 factures | 3 émises | 1 échue | Total: 4,930€ TTC        │ Info Bar
└─────────────────────────────────────────────────────────────────┘
 ↑ Ligne 2/80   ↑ Statut     ↑ Filtre actif      ↑ Stats
```

### Zones Fonctionnelles

1. **Header (1 ligne)** : Contexte global, année, mode
2. **Command Bar (1 ligne)** : Commande active + hints
3. **Main Pane (variable)** : Contenu principal
4. **Info Bar (1 ligne)** : Stats, filtres, messages
5. **Help Panel (overlay)** : Aide contextuelle (toggle)

---

## 🎨 Palette de Couleurs (Lipgloss)

### Couleurs Principales

```go
package styles

import "github.com/charmbracelet/lipgloss"

var (
    // Palette de base
    ColorPrimary     = lipgloss.Color("#0EA5E9")  // Bleu ciel
    ColorSecondary   = lipgloss.Color("#8B5CF6")  // Violet
    ColorSuccess     = lipgloss.Color("#10B981")  // Vert
    ColorWarning     = lipgloss.Color("#F59E0B")  // Orange
    ColorDanger      = lipgloss.Color("#EF4444")  // Rouge
    ColorMuted       = lipgloss.Color("#6B7280")  // Gris
    ColorBackground  = lipgloss.Color("#1F2937")  // Gris foncé
    ColorForeground  = lipgloss.Color("#F9FAFB")  // Blanc cassé
    
    // Couleurs par état de facture
    ColorStateDraft    = ColorMuted
    ColorStateIssued   = ColorPrimary
    ColorStateSent     = ColorSecondary
    ColorStatePaid     = ColorSuccess
    ColorStateOverdue  = ColorDanger
    ColorStateCanceled = ColorMuted
)
```

### Styles par État

```go
// États de facture
var (
    StyleDraft = lipgloss.NewStyle().
        Foreground(ColorStateDraft).
        Bold(true)
    
    StyleIssued = lipgloss.NewStyle().
        Foreground(ColorStateIssued).
        Bold(true)
    
    StylePaid = lipgloss.NewStyle().
        Foreground(ColorStatePaid).
        Bold(true)
    
    StyleOverdue = lipgloss.NewStyle().
        Foreground(ColorStateOverdue).
        Bold(true).
        Blink(true)  // Clignotant pour attirer l'attention
    
    // Style spécial pour factures IMMUABLES
    StyleImmutable = lipgloss.NewStyle().
        Foreground(ColorForeground).
        Background(lipgloss.Color("#7F1D1D")).  // Rouge foncé
        Bold(true).
        Padding(0, 1).
        Render("🔒 PAYÉE")
)
```

### Badges et Indicateurs

```go
// Badges visuels
func BadgePaid() string {
    return lipgloss.NewStyle().
        Foreground(ColorSuccess).
        Render("✓ PAYÉE")
}

func BadgeOverdue() string {
    return lipgloss.NewStyle().
        Foreground(ColorDanger).
        Blink(true).
        Render("⚠ ÉCHUE")
}

func BadgeLocked() string {
    return lipgloss.NewStyle().
        Foreground(ColorForeground).
        Background(ColorDanger).
        Bold(true).
        Render("🔒 IMMUABLE")
}

func BadgeDraft() string {
    return lipgloss.NewStyle().
        Foreground(ColorMuted).
        Render("📝 BROUILLON")
}
```

---

## 📋 Vues Principales

### 1. Vue Dashboard (`:pulse`)

```
┌─ DASHBOARD ────────────────────────────────────────────────────┐
│                                                                 │
│  💰 CHIFFRE D'AFFAIRES 2026                                    │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│  Total HT:  45,890 €    │  Total TTC:  45,890 €               │
│  Seuils TVA: ✓ Franchise en base (39,100 € max)               │
│                                                                 │
│  📊 FACTURES EN COURS                                          │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│  Brouillons:    3  │  Émises:       8  │  Envoyées:      4   │
│  Payées:       45  │  Échues:       2  │  Total:        62   │
│                                                                 │
│  ⚠️  ALERTES                                                    │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│  • 2 factures échues (total: 3,450 €)                          │
│  • Facturation électronique obligatoire dans 529 jours         │
│  • 3 devis expirant dans moins de 7 jours                      │
│                                                                 │
│  📈 ÉVOLUTION MENSUELLE                                        │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│  Janvier:  3,890 €  ████████░░                                 │
│  Février:  8,450 €  █████████████████░░                        │
│  Mars:    12,340 €  █████████████████████████                  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 2. Vue Clients (`:clients`)

```
┌─ CLIENTS ──────────────────────────────────────────────────────┐
│ NOM              SIRET           EMAIL            CA 2026      │
│ ────────────────────────────────────────────────────────────── │
│>Acme Corporation 12345678901234  contact@acme.fr  15,890 €    │
│ Digital SARL     98765432109876  info@digital.fr   8,450 €    │
│ StartUp SAS      11223344556677  hello@startup.fr  2,340 €    │
│ Solutions Inc    44556677889900  team@solutions.fr 9,120 €    │
│                                                                 │
│ [Actions]                                                       │
│ n: Nouveau  e: Éditer  d: Supprimer  f: Factures  Enter: Voir │
└─────────────────────────────────────────────────────────────────┘
```

### 3. Vue Factures (`:factures`)

```
┌─ FACTURES ─────────────────────────────────────────────────────┐
│ NUM       CLIENT          MONTANT   ÉTAT        ÉCHÉANCE       │
│ ────────────────────────────────────────────────────────────── │
│ 2026-0001 Acme Corp       1,250€    ✓ PAYÉE    15/02/2026     │
│>2026-0002 Digital SA        890€    ⚠ ÉCHUE    01/03/2026     │
│ 2026-0003 Solutions       2,340€    ÉMISE      20/03/2026     │
│ 2026-0004 StartUp Inc       450€    📝 BROUILLON 25/03/2026   │
│ 2026-0005 Acme Corp       5,600€    🔒 PAYÉE   10/03/2026     │
│                                                                 │
│ [Actions]                                                       │
│ n: Nouvelle  e: Éditer  p: PDF  m: Marquer payée  c: Avoir    │
└─────────────────────────────────────────────────────────────────┘
```

### 4. Vue Détail Facture

```
┌─ FACTURE 2026-0002 ────────────────────────────────────────────┐
│                                                                 │
│  CLIENT                           │  FACTURE                   │
│  Digital SARL                     │  Numéro: 2026-0002         │
│  98765432109876                   │  Émise: 01/02/2026         │
│  123 Rue de la Tech               │  Échéance: 01/03/2026      │
│  75001 Paris                      │  État: ⚠ ÉCHUE (3j)        │
│  contact@digital.fr               │                            │
│                                                                 │
│  LIGNES ──────────────────────────────────────────────────────│
│  Description              Qté    Prix HT    Total HT          │
│  ────────────────────────────────────────────────────────────│
│  Développement site web    1     750.00€     750.00€          │
│  Hébergement 1 an          1     140.00€     140.00€          │
│  ────────────────────────────────────────────────────────────│
│                                   TOTAL HT:   890.00€          │
│                                   TVA (0%):     0.00€          │
│                                   TOTAL TTC:  890.00€          │
│                                                                 │
│  💡 TVA non applicable, article 293B du CGI                    │
│                                                                 │
│  PAIEMENT ────────────────────────────────────────────────────│
│  Délai: 30 jours                                               │
│  Pénalités retard: 13.25% / an                                │
│  Indemnité forfaitaire: 40€                                    │
│                                                                 │
│ [Actions]                                                       │
│ p: Générer PDF  s: Envoyer  m: Marquer payée  e: Éditer       │
└─────────────────────────────────────────────────────────────────┘
```

### 5. Vue Détail Facture PAYÉE (Immuable)

```
┌─ FACTURE 2026-0001 🔒 PAYÉE (IMMUABLE) ───────────────────────┐
│                                                                 │
│  ⚠️  ATTENTION: Cette facture est PAYÉE et ne peut être       │
│     modifiée. Seule action possible: créer un avoir (c)        │
│                                                                 │
│  CLIENT                           │  FACTURE                   │
│  Acme Corporation                 │  Numéro: 2026-0001         │
│  12345678901234                   │  Émise: 15/01/2026         │
│  456 Avenue Business              │  Payée: 15/02/2026 ✓       │
│  69001 Lyon                       │  Verrouillée: 15/02/2026   │
│  contact@acme.fr                  │                            │
│                                                                 │
│  [Lignes, totaux identiques...]                                │
│                                                                 │
│ [Actions LIMITÉES]                                              │
│ p: Voir PDF  v: Ouvrir PDF  c: Créer AVOIR  ESC: Retour       │
└─────────────────────────────────────────────────────────────────┘
```

---

## ⌨️ Navigation et Raccourcis

### Navigation Globale

| Touche | Action |
|--------|--------|
| `:` | Mode commande (taper `:clients`, `:factures`, etc.) |
| `/` | Mode recherche/filtre |
| `?` | Aide contextuelle (toggle) |
| `ESC` | Annuler / Retour |
| `Tab` | Cycle entre vues |
| `Ctrl+c` ou `q` | Quitter |

### Navigation Liste (Vim-like)

| Touche | Action |
|--------|--------|
| `j` ou `↓` | Ligne suivante |
| `k` ou `↑` | Ligne précédente |
| `g` | Première ligne |
| `G` | Dernière ligne |
| `Ctrl+d` | Page down |
| `Ctrl+u` | Page up |
| `h` ou `←` | Colonne précédente (si applicable) |
| `l` ou `→` | Colonne suivante (si applicable) |

### Commandes (Mode `:`)

```
:clients       → Vue clients
:factures      → Vue factures
:devis         → Vue devis
:avoirs        → Vue avoirs
:pulse         → Dashboard
:help          → Aide complète
:quit ou :q    → Quitter
```

### Actions sur Entités Sélectionnées

#### Clients
| Touche | Action |
|--------|--------|
| `n` | Nouveau client |
| `e` | Éditer client |
| `d` | Supprimer client |
| `f` | Voir factures du client |
| `Enter` | Voir détails |

#### Factures
| Touche | Action |
|--------|--------|
| `n` | Nouvelle facture |
| `e` | Éditer (si non payée) |
| `d` | Supprimer (si brouillon) |
| `p` | Générer PDF |
| `v` | Ouvrir PDF |
| `m` | Marquer comme payée |
| `c` | Créer avoir (si payée) |
| `s` | Envoyer par email |
| `Ctrl+d` | Dupliquer |
| `Enter` | Voir détails |

#### Devis
| Touche | Action |
|--------|--------|
| `n` | Nouveau devis |
| `e` | Éditer |
| `d` | Supprimer |
| `f` | Convertir en facture |
| `p` | Générer PDF |
| `s` | Envoyer |
| `Enter` | Voir détails |

---

## 🔍 Filtres et Recherche

### Mode Recherche (`/`)

```
┌─ FACTURES ─────────────────────────────────────────────────────┐
│ / Recherche: acme                                    [ESC] Annuler
│ ────────────────────────────────────────────────────────────── │
│ 2026-0001 Acme Corp       1,250€    ✓ PAYÉE    15/02/2026     │
│ 2026-0005 Acme Corp       5,600€    🔒 PAYÉE   10/03/2026     │
│                                                                 │
│ 2 résultats pour "acme"                                        │
└─────────────────────────────────────────────────────────────────┘
```

### Filtres Avancés

```go
// Syntaxe filtres
/état:payée          // Factures payées uniquement
/client:acme         // Factures pour "acme"
/montant:>1000       // Factures > 1000€
/échue               // Factures échues
/brouillon           // Brouillons
/2026-01             // Factures de janvier 2026

// Combinaisons
/état:payée client:acme
```

---

## 🎭 États Visuels

### Indicateurs d'État Facture

```go
func RenderInvoiceState(state domain.InvoiceState, isOverdue bool) string {
    if isOverdue {
        return StyleOverdue.Render("⚠ ÉCHUE")
    }
    
    switch state {
    case domain.InvoiceStateDraft:
        return StyleDraft.Render("📝 BROUILLON")
    case domain.InvoiceStateIssued:
        return StyleIssued.Render("ÉMISE")
    case domain.InvoiceStateSent:
        return StyleIssued.Render("ENVOYÉE")
    case domain.InvoiceStatePaid:
        return StylePaid.Render("✓ PAYÉE")
    case domain.InvoiceStateCanceled:
        return StyleMuted.Render("ANNULÉE")
    default:
        return StyleMuted.Render("INCONNU")
    }
}
```

### Codes Couleur Cohérents

| Couleur | Usage |
|---------|-------|
| 🟢 Vert | Succès, Payé, Validé |
| 🔵 Bleu | Info, Émise, Actif |
| 🟡 Jaune | Attention, Brouillon, En attente |
| 🔴 Rouge | Erreur, Échu, Critique |
| ⚪ Gris | Neutre, Annulé, Désactivé |

---

## 📱 Composants Réutilisables (Bubbles)

### Table Interactive

```go
package components

import (
    "github.com/charmbracelet/bubbles/table"
    "github.com/charmbracelet/lipgloss"
)

type TableModel struct {
    table table.Model
}

func NewTable(columns []table.Column, rows []table.Row) TableModel {
    t := table.New(
        table.WithColumns(columns),
        table.WithRows(rows),
        table.WithFocused(true),
        table.WithHeight(20),
    )
    
    s := table.DefaultStyles()
    s.Header = s.Header.
        BorderStyle(lipgloss.NormalBorder()).
        BorderForeground(lipgloss.Color("240")).
        BorderBottom(true).
        Bold(true)
    
    s.Selected = s.Selected.
        Foreground(lipgloss.Color("229")).
        Background(lipgloss.Color("57")).
        Bold(true)
    
    t.SetStyles(s)
    
    return TableModel{table: t}
}
```

### Formulaire Interactif

```go
package components

import (
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/lipgloss"
)

type FormField struct {
    Label    string
    Input    textinput.Model
    Required bool
    Error    string
}

type FormModel struct {
    Fields  []FormField
    Focused int
}

func NewForm(fields []FormField) FormModel {
    for i := range fields {
        fields[i].Input.Focus()
        if i == 0 {
            fields[i].Input.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
        }
    }
    
    return FormModel{
        Fields:  fields,
        Focused: 0,
    }
}
```

### Panneau d'Aide (Help Panel)

```go
type HelpPanel struct {
    visible bool
    keys    []KeyBinding
}

type KeyBinding struct {
    Key         string
    Description string
    Category    string
}

func (h HelpPanel) View() string {
    if !h.visible {
        return ""
    }
    
    help := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("62")).
        Padding(1, 2).
        Width(60)
    
    content := "AIDE - Raccourcis Clavier\n\n"
    
    categories := map[string][]KeyBinding{}
    for _, kb := range h.keys {
        categories[kb.Category] = append(categories[kb.Category], kb)
    }
    
    for cat, bindings := range categories {
        content += lipgloss.NewStyle().Bold(true).Render(cat) + "\n"
        for _, kb := range bindings {
            content += fmt.Sprintf("  %-10s  %s\n", kb.Key, kb.Description)
        }
        content += "\n"
    }
    
    return help.Render(content)
}
```

---

## 🚨 Messages d'Erreur et Avertissements

### Erreurs Légales

```go
// Affichage erreur immuabilité
func RenderImmutableError(invoiceNumber string) string {
    return lipgloss.NewStyle().
        Foreground(ColorForeground).
        Background(ColorDanger).
        Padding(1, 2).
        Bold(true).
        Render(fmt.Sprintf(
            "❌ MODIFICATION INTERDITE\n\n"+
            "La facture %s est PAYÉE et ne peut être modifiée.\n"+
            "Action autorisée : Créer un avoir (touche 'c')",
            invoiceNumber,
        ))
}
```

### Validation Mentions Manquantes

```go
func RenderValidationErrors(errors []domain.ValidationError) string {
    content := "⚠️  ERREURS DE VALIDATION\n\n"
    totalFine := decimal.Zero
    
    for _, err := range errors {
        content += fmt.Sprintf("• %s: %s\n", err.Field, err.Message)
        if err.Fine.GreaterThan(decimal.Zero) {
            content += fmt.Sprintf("  Amende: %.2f €\n", err.Fine)
            totalFine = totalFine.Add(err.Fine)
        }
    }
    
    content += fmt.Sprintf("\nAmende totale potentielle: %.2f €", totalFine)
    
    return lipgloss.NewStyle().
        Foreground(ColorWarning).
        Border(lipgloss.RoundedBorder()).
        Padding(1, 2).
        Render(content)
}
```

---

## 📊 Widgets Spéciaux

### Graphique CA Mensuel (ASCII)

```go
func RenderMonthlyRevenue(data map[string]decimal.Decimal) string {
    maxAmount := decimal.Zero
    for _, amount := range data {
        if amount.GreaterThan(maxAmount) {
            maxAmount = amount
        }
    }
    
    barWidth := 30
    output := ""
    
    months := []string{"Jan", "Fév", "Mar", "Avr", "Mai", "Jun",
                       "Jul", "Aoû", "Sep", "Oct", "Nov", "Déc"}
    
    for _, month := range months {
        amount := data[month]
        if amount.IsZero() {
            continue
        }
        
        ratio := amount.Div(maxAmount)
        filledBars := int(ratio.Mul(decimal.NewFromInt(int64(barWidth))).IntPart())
        
        bar := strings.Repeat("█", filledBars) + 
              strings.Repeat("░", barWidth-filledBars)
        
        output += fmt.Sprintf("%s: %7s €  %s\n", 
            month, 
            amount.StringFixed(0),
            bar)
    }
    
    return output
}
```

### Indicateur Seuils TVA

```go
func RenderVATThresholdIndicator(currentCA decimal.Decimal, isService bool) string {
    threshold := decimal.NewFromInt(39100) // Services
    if !isService {
        threshold = decimal.NewFromInt(91900) // Vente
    }
    
    percentage := currentCA.Div(threshold).Mul(decimal.NewFromInt(100))
    
    color := ColorSuccess
    if percentage.GreaterThan(decimal.NewFromInt(80)) {
        color = ColorWarning
    }
    if percentage.GreaterThan(decimal.NewFromInt(95)) {
        color = ColorDanger
    }
    
    barWidth := 40
    filled := int(percentage.Div(decimal.NewFromInt(100)).
                   Mul(decimal.NewFromInt(int64(barWidth))).IntPart())
    
    bar := lipgloss.NewStyle().Foreground(color).
        Render(strings.Repeat("█", filled)) +
        strings.Repeat("░", barWidth-filled)
    
    return fmt.Sprintf(
        "Seuil TVA: %s  %.1f%% (%s€ / %s€)",
        bar,
        percentage,
        currentCA.StringFixed(0),
        threshold.StringFixed(0),
    )
}
```

---

## 🎬 Animations et Transitions

### Spinner Loading

```go
import "github.com/charmbracelet/bubbles/spinner"

type LoadingModel struct {
    spinner  spinner.Model
    message  string
}

func NewLoading(message string) LoadingModel {
    s := spinner.New()
    s.Spinner = spinner.Dot
    s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
    
    return LoadingModel{
        spinner: s,
        message: message,
    }
}

func (m LoadingModel) View() string {
    return fmt.Sprintf("%s %s", m.spinner.View(), m.message)
}
```

### Toast Notifications

```go
type Toast struct {
    Message  string
    Type     ToastType // success, warning, error, info
    Duration time.Duration
    ShowTime time.Time
}

func (t Toast) IsVisible() bool {
    return time.Since(t.ShowTime) < t.Duration
}

func (t Toast) View() string {
    if !t.IsVisible() {
        return ""
    }
    
    var style lipgloss.Style
    switch t.Type {
    case ToastSuccess:
        style = lipgloss.NewStyle().
            Foreground(ColorSuccess).
            Background(lipgloss.Color("#065F46"))
    case ToastError:
        style = lipgloss.NewStyle().
            Foreground(ColorForeground).
            Background(ColorDanger)
    // ... autres types
    }
    
    return style.
        Padding(0, 2).
        Render(t.Message)
}
```

---

## 📐 Responsive Layout

```go
func (m *App) handleWindowSize(width, height int) {
    // Ajuster hauteur des composants
    headerHeight := 1
    commandBarHeight := 1
    infoBarHeight := 1
    
    mainPaneHeight := height - headerHeight - commandBarHeight - infoBarHeight
    
    // Adapter table
    if m.currentView == ViewInvoices {
        m.invoicesView.table.SetHeight(mainPaneHeight - 4)
    }
    
    // Adapter largeur colonnes
    if width < 80 {
        // Mode compact : réduire colonnes
        m.invoicesView.table.SetColumns(compactColumns)
    } else {
        // Mode normal
        m.invoicesView.table.SetColumns(normalColumns)
    }
}
```

---

**Dernière Mise à Jour** : 2026-03-04
