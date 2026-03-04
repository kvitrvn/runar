# 🏗️ ARCHITECTURE.md - Architecture Technique

## 📐 Vue d'Ensemble

AutoGest suit une architecture en couches (Layered Architecture) avec séparation stricte des responsabilités.

```
┌─────────────────────────────────────────┐
│         TUI Layer (Bubbletea)          │  Présentation
├─────────────────────────────────────────┤
│         Service Layer                   │  Logique Métier
├─────────────────────────────────────────┤
│         Domain Layer                    │  Entités & Règles
├─────────────────────────────────────────┤
│         Repository Layer                │  Persistance
├─────────────────────────────────────────┤
│         SQLite Database                 │  Stockage
└─────────────────────────────────────────┘
```

---

## 🗂️ Structure Détaillée des Packages

### `/cmd/autogest`

**Responsabilité** : Point d'entrée de l'application

```go
// main.go
package main

import (
    "log"
    "os"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/yourname/autogest/internal/config"
    "github.com/yourname/autogest/internal/repository"
    "github.com/yourname/autogest/internal/service"
    "github.com/yourname/autogest/internal/tui"
)

func main() {
    // 1. Charger configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. Initialiser DB
    db, err := repository.InitDB(cfg.DatabasePath)
    if err != nil {
        log.Fatal(err)
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
        log.Fatal(err)
    }
}
```

### `/internal/domain`

**Responsabilité** : Entités métier pures, règles de validation, types

#### `client.go`

```go
package domain

import (
    "time"
    "github.com/shopspring/decimal"
)

// Client représente un client de l'auto-entreprise
type Client struct {
    ID          int
    Name        string
    SIRET       string // 14 chiffres pour entreprises
    SIREN       string // 9 chiffres (partie du SIRET)
    Address     string
    PostalCode  string
    City        string
    Country     string
    Email       string
    Phone       string
    VATNumber   string // N° TVA intracommunautaire si EU
    Notes       string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// IsCompany retourne true si le client est une entreprise (B2B)
func (c *Client) IsCompany() bool {
    return c.SIRET != "" || c.SIREN != ""
}

// Validate vérifie la validité d'un client
func (c *Client) Validate() []ValidationError {
    var errors []ValidationError
    
    if c.Name == "" {
        errors = append(errors, ValidationError{
            Field:   "name",
            Message: "Le nom du client est obligatoire",
        })
    }
    
    if c.SIRET != "" && !ValidateSIRET(c.SIRET) {
        errors = append(errors, ValidationError{
            Field:   "siret",
            Message: "SIRET invalide (14 chiffres attendus)",
        })
    }
    
    if c.Email != "" && !ValidateEmail(c.Email) {
        errors = append(errors, ValidationError{
            Field:   "email",
            Message: "Email invalide",
        })
    }
    
    return errors
}
```

#### `invoice.go`

```go
package domain

import (
    "time"
    "github.com/shopspring/decimal"
)

// InvoiceState représente l'état d'une facture
type InvoiceState string

const (
    InvoiceStateDraft    InvoiceState = "draft"    // Brouillon
    InvoiceStateIssued   InvoiceState = "issued"   // Émise
    InvoiceStateSent     InvoiceState = "sent"     // Envoyée
    InvoiceStatePaid     InvoiceState = "paid"     // Payée (IMMUABLE)
    InvoiceStateCanceled InvoiceState = "canceled" // Annulée (via avoir)
)

// OperationCategory pour obligations 2027
type OperationCategory string

const (
    OperationService OperationCategory = "service" // Prestation
    OperationGoods   OperationCategory = "goods"   // Vente
    OperationMixed   OperationCategory = "mixed"   // Mixte
)

// Invoice représente une facture
type Invoice struct {
    ID                 int
    Number             string    // Ex: "2026-0001"
    ClientID           int
    Client             *Client   // Relation
    QuoteID            *int      // Si conversion depuis devis
    
    // Dates
    IssueDate          time.Time
    DueDate            time.Time
    DeliveryDate       time.Time
    PaidDate           *time.Time // nil si non payée
    
    // État
    State              InvoiceState
    
    // Lignes
    Lines              []InvoiceLine
    
    // Montants (calculés depuis Lines)
    TotalHT            decimal.Decimal
    TotalTTC           decimal.Decimal
    VATAmount          decimal.Decimal
    
    // TVA
    VATApplicable      bool
    VATExemptionText   string // "TVA non applicable, article 293B du CGI"
    
    // Paiement
    PaymentDeadline    string          // "30 jours", "45 jours fin de mois"
    LatePenaltyRate    decimal.Decimal // Taux annuel
    RecoveryFee        decimal.Decimal // 40 €
    EarlyPaymentDiscount string        // Escompte optionnel
    
    // Nouvelles mentions 2027
    OperationCategory  OperationCategory
    DeliveryAddress    string // Si différente de l'adresse client
    
    // Métadonnées
    Notes              string
    PDFPath            string
    CreatedAt          time.Time
    UpdatedAt          time.Time
    PaidLockedAt       *time.Time // Timestamp du verrouillage
}

// InvoiceLine représente une ligne de facture
type InvoiceLine struct {
    ID           int
    InvoiceID    int
    LineOrder    int    // Ordre d'affichage
    Description  string
    Quantity     decimal.Decimal
    UnitPriceHT  decimal.Decimal
    VATRate      decimal.Decimal // En pourcentage (20.0 pour 20%)
    TotalHT      decimal.Decimal
    TotalTTC     decimal.Decimal
}

// CanEdit retourne true si la facture peut être éditée
// LEGAL: Une facture payée ne peut JAMAIS être modifiée
func (inv *Invoice) CanEdit() bool {
    return inv.State != InvoiceStatePaid && inv.State != InvoiceStateCanceled
}

// CanDelete retourne true si la facture peut être supprimée
func (inv *Invoice) CanDelete() bool {
    return inv.State == InvoiceStateDraft
}

// CanMarkAsPaid retourne true si la facture peut être marquée comme payée
func (inv *Invoice) CanMarkAsPaid() bool {
    return inv.State == InvoiceStateIssued || inv.State == InvoiceStateSent
}

// IsOverdue retourne true si la facture est échue
func (inv *Invoice) IsOverdue() bool {
    return inv.State != InvoiceStatePaid && 
           inv.State != InvoiceStateDraft && 
           time.Now().After(inv.DueDate)
}

// CalculateTotals recalcule les totaux depuis les lignes
func (inv *Invoice) CalculateTotals() {
    inv.TotalHT = decimal.Zero
    inv.VATAmount = decimal.Zero
    
    for _, line := range inv.Lines {
        inv.TotalHT = inv.TotalHT.Add(line.TotalHT)
        
        if inv.VATApplicable {
            lineVAT := line.TotalHT.Mul(line.VATRate).Div(decimal.NewFromInt(100))
            inv.VATAmount = inv.VATAmount.Add(lineVAT)
        }
    }
    
    inv.TotalTTC = inv.TotalHT.Add(inv.VATAmount)
}

// Validate vérifie la conformité légale de la facture
// LEGAL: Toutes les mentions obligatoires doivent être présentes
func (inv *Invoice) Validate() []ValidationError {
    var errors []ValidationError
    
    // Numéro obligatoire
    if inv.Number == "" {
        errors = append(errors, ValidationError{
            Field:   "number",
            Message: "Numéro de facture obligatoire",
            Fine:    decimal.NewFromInt(15),
        })
    }
    
    // Date émission
    if inv.IssueDate.IsZero() {
        errors = append(errors, ValidationError{
            Field:   "issue_date",
            Message: "Date d'émission obligatoire",
            Fine:    decimal.NewFromInt(15),
        })
    }
    
    // Date livraison
    if inv.DeliveryDate.IsZero() {
        errors = append(errors, ValidationError{
            Field:   "delivery_date",
            Message: "Date de livraison/fin de prestation obligatoire",
            Fine:    decimal.NewFromInt(15),
        })
    }
    
    // Client
    if inv.ClientID == 0 {
        errors = append(errors, ValidationError{
            Field:   "client_id",
            Message: "Client obligatoire",
            Fine:    decimal.NewFromInt(15),
        })
    }
    
    // SIREN client si B2B (depuis 2024)
    if inv.Client != nil && inv.Client.IsCompany() && inv.Client.SIREN == "" {
        errors = append(errors, ValidationError{
            Field:   "client.siren",
            Message: "SIREN obligatoire pour clients professionnels (depuis 2024)",
            Fine:    decimal.NewFromInt(15),
        })
    }
    
    // Au moins une ligne
    if len(inv.Lines) == 0 {
        errors = append(errors, ValidationError{
            Field:   "lines",
            Message: "Au moins une ligne obligatoire",
            Fine:    decimal.NewFromInt(15),
        })
    }
    
    // Validation lignes
    for i, line := range inv.Lines {
        if line.Description == "" {
            errors = append(errors, ValidationError{
                Field:   fmt.Sprintf("lines[%d].description", i),
                Message: "Description obligatoire",
                Fine:    decimal.NewFromInt(15),
            })
        }
        
        if line.Quantity.LessThanOrEqual(decimal.Zero) {
            errors = append(errors, ValidationError{
                Field:   fmt.Sprintf("lines[%d].quantity", i),
                Message: "Quantité doit être > 0",
                Fine:    decimal.NewFromInt(15),
            })
        }
        
        if line.UnitPriceHT.LessThan(decimal.Zero) {
            errors = append(errors, ValidationError{
                Field:   fmt.Sprintf("lines[%d].unit_price_ht", i),
                Message: "Prix unitaire ne peut être négatif",
                Fine:    decimal.NewFromInt(15),
        })
        }
    }
    
    // Mention TVA si franchise en base
    if !inv.VATApplicable && inv.VATExemptionText != "TVA non applicable, article 293B du CGI" {
        errors = append(errors, ValidationError{
            Field:   "vat_exemption_text",
            Message: "Mention TVA obligatoire : 'TVA non applicable, article 293B du CGI'",
            Fine:    decimal.NewFromInt(15),
        })
    }
    
    // Délai de paiement
    if inv.PaymentDeadline == "" {
        errors = append(errors, ValidationError{
            Field:   "payment_deadline",
            Message: "Délai de paiement obligatoire",
            Fine:    decimal.NewFromInt(15),
        })
    }
    
    // Taux pénalités retard
    if inv.LatePenaltyRate.LessThanOrEqual(decimal.Zero) {
        errors = append(errors, ValidationError{
            Field:   "late_penalty_rate",
            Message: "Taux de pénalités de retard obligatoire",
            Fine:    decimal.NewFromInt(15),
        })
    }
    
    // Indemnité forfaitaire (40€)
    if !inv.RecoveryFee.Equal(decimal.NewFromInt(40)) {
        errors = append(errors, ValidationError{
            Field:   "recovery_fee",
            Message: "Indemnité forfaitaire de recouvrement doit être 40€",
            Fine:    decimal.NewFromInt(15),
        })
    }
    
    return errors
}
```

#### `quote.go`

```go
package domain

// QuoteState représente l'état d'un devis
type QuoteState string

const (
    QuoteStateDraft    QuoteState = "draft"    // Brouillon
    QuoteStateSent     QuoteState = "sent"     // Envoyé
    QuoteStateAccepted QuoteState = "accepted" // Accepté
    QuoteStateRefused  QuoteState = "refused"  // Refusé
    QuoteStateExpired  QuoteState = "expired"  // Expiré
)

// Quote représente un devis
type Quote struct {
    ID             int
    Number         string // Ex: "DEV-2026-0001"
    ClientID       int
    Client         *Client
    IssueDate      time.Time
    ExpiryDate     time.Time
    State          QuoteState
    Lines          []QuoteLine
    TotalHT        decimal.Decimal
    TotalTTC       decimal.Decimal
    VATAmount      decimal.Decimal
    Notes          string
    PDFPath        string
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

// QuoteLine représente une ligne de devis
type QuoteLine struct {
    ID           int
    QuoteID      int
    LineOrder    int
    Description  string
    Quantity     decimal.Decimal
    UnitPriceHT  decimal.Decimal
    VATRate      decimal.Decimal
    TotalHT      decimal.Decimal
    TotalTTC     decimal.Decimal
}

// CanConvertToInvoice retourne true si le devis peut être converti
func (q *Quote) CanConvertToInvoice() bool {
    return q.State == QuoteStateAccepted
}

// IsExpired retourne true si le devis est expiré
func (q *Quote) IsExpired() bool {
    return time.Now().After(q.ExpiryDate)
}
```

#### `validation.go`

```go
package domain

import (
    "fmt"
    "regexp"
    "strconv"
    "strings"
    
    "github.com/shopspring/decimal"
)

// ValidationError représente une erreur de validation
type ValidationError struct {
    Field   string
    Message string
    Fine    decimal.Decimal // Amende potentielle
}

func (e ValidationError) Error() string {
    if e.Fine.GreaterThan(decimal.Zero) {
        return fmt.Sprintf("%s: %s (Amende: %.2f €)", e.Field, e.Message, e.Fine)
    }
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateSIRET vérifie la validité d'un SIRET (14 chiffres)
func ValidateSIRET(siret string) bool {
    siret = strings.ReplaceAll(siret, " ", "")
    
    if len(siret) != 14 {
        return false
    }
    
    if _, err := strconv.Atoi(siret); err != nil {
        return false
    }
    
    return luhnCheck(siret)
}

// ValidateSIREN vérifie la validité d'un SIREN (9 chiffres)
func ValidateSIREN(siren string) bool {
    siren = strings.ReplaceAll(siren, " ", "")
    
    if len(siren) != 9 {
        return false
    }
    
    if _, err := strconv.Atoi(siren); err != nil {
        return false
    }
    
    return luhnCheck(siren)
}

// ValidateEmail vérifie la validité d'un email
func ValidateEmail(email string) bool {
    re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return re.MatchString(email)
}

// luhnCheck implémente l'algorithme de Luhn pour SIRET/SIREN
func luhnCheck(number string) bool {
    sum := 0
    alternate := false
    
    for i := len(number) - 1; i >= 0; i-- {
        n, _ := strconv.Atoi(string(number[i]))
        
        if alternate {
            n *= 2
            if n > 9 {
                n -= 9
            }
        }
        
        sum += n
        alternate = !alternate
    }
    
    return sum%10 == 0
}
```

### `/internal/service`

**Responsabilité** : Orchestration, règles métier, transactions

#### `invoice_service.go`

```go
package service

import (
    "fmt"
    "time"
    
    "github.com/shopspring/decimal"
    "github.com/yourname/autogest/internal/domain"
    "github.com/yourname/autogest/internal/repository"
)

type InvoiceService struct {
    invoiceRepo repository.InvoiceRepository
    clientRepo  repository.ClientRepository
    auditRepo   repository.AuditRepository
    pdfService  *PDFService
}

func NewInvoiceService(
    invoiceRepo repository.InvoiceRepository,
    clientRepo repository.ClientRepository,
    auditRepo repository.AuditRepository,
    pdfService *PDFService,
) *InvoiceService {
    return &InvoiceService{
        invoiceRepo: invoiceRepo,
        clientRepo:  clientRepo,
        auditRepo:   auditRepo,
        pdfService:  pdfService,
    }
}

// Create crée une nouvelle facture
func (s *InvoiceService) Create(invoice *domain.Invoice) error {
    // Validation
    if errs := invoice.Validate(); len(errs) > 0 {
        return &ValidationErrorList{Errors: errs}
    }
    
    // Générer numéro unique
    number, err := s.GenerateNextNumber(invoice.IssueDate.Year())
    if err != nil {
        return err
    }
    invoice.Number = number
    
    // Calculer totaux
    invoice.CalculateTotals()
    
    // Sauvegarder
    if err := s.invoiceRepo.Create(invoice); err != nil {
        return err
    }
    
    // Audit log
    s.auditRepo.Log(domain.AuditLog{
        EntityType: "invoice",
        EntityID:   invoice.ID,
        Action:     "created",
        NewValue:   toJSON(invoice),
        Timestamp:  time.Now(),
    })
    
    return nil
}

// Update met à jour une facture
// LEGAL: Impossible si facture payée
func (s *InvoiceService) Update(id int, updates *domain.Invoice) error {
    existing, err := s.invoiceRepo.GetByID(id)
    if err != nil {
        return err
    }
    
    // RÈGLE CRITIQUE: Facture payée = immuable
    if !existing.CanEdit() {
        return &ErrImmutableInvoice{
            InvoiceNumber: existing.Number,
            State:         string(existing.State),
        }
    }
    
    // Validation
    if errs := updates.Validate(); len(errs) > 0 {
        return &ValidationErrorList{Errors: errs}
    }
    
    // Recalculer totaux
    updates.CalculateTotals()
    
    // Audit log AVANT modification
    s.auditRepo.Log(domain.AuditLog{
        EntityType: "invoice",
        EntityID:   id,
        Action:     "updated",
        OldValue:   toJSON(existing),
        NewValue:   toJSON(updates),
        Timestamp:  time.Now(),
    })
    
    return s.invoiceRepo.Update(id, updates)
}

// MarkAsPaid marque une facture comme payée et la verrouille
// LEGAL: Après cet appel, la facture devient IMMUABLE
func (s *InvoiceService) MarkAsPaid(id int, paidDate time.Time) error {
    invoice, err := s.invoiceRepo.GetByID(id)
    if err != nil {
        return err
    }
    
    if !invoice.CanMarkAsPaid() {
        return fmt.Errorf("facture dans état %s ne peut être marquée comme payée", invoice.State)
    }
    
    // Mise à jour état
    now := time.Now()
    invoice.State = domain.InvoiceStatePaid
    invoice.PaidDate = &paidDate
    invoice.PaidLockedAt = &now
    
    if err := s.invoiceRepo.Update(id, invoice); err != nil {
        return err
    }
    
    // Audit log critique
    s.auditRepo.Log(domain.AuditLog{
        EntityType: "invoice",
        EntityID:   id,
        Action:     "paid_and_locked",
        OldValue:   fmt.Sprintf(`{"state": "%s"}`, domain.InvoiceStateIssued),
        NewValue:   fmt.Sprintf(`{"state": "%s", "paid_date": "%s"}`, domain.InvoiceStatePaid, paidDate),
        Timestamp:  now,
    })
    
    return nil
}

// GenerateNextNumber génère le prochain numéro de facture
// LEGAL: Numérotation continue sans trou
func (s *InvoiceService) GenerateNextNumber(year int) (string, error) {
    lastSeq, err := s.invoiceRepo.GetLastSequence(year)
    if err != nil {
        return "", err
    }
    
    nextSeq := lastSeq + 1
    number := fmt.Sprintf("%d-%04d", year, nextSeq)
    
    // Vérification doublon (sécurité)
    exists, err := s.invoiceRepo.NumberExists(number)
    if err != nil {
        return "", err
    }
    if exists {
        return "", fmt.Errorf("numéro %s existe déjà (race condition)", number)
    }
    
    return number, nil
}

// GeneratePDF génère le PDF d'une facture
func (s *InvoiceService) GeneratePDF(id int) (string, error) {
    invoice, err := s.invoiceRepo.GetByID(id)
    if err != nil {
        return "", err
    }
    
    // Charger client
    client, err := s.clientRepo.GetByID(invoice.ClientID)
    if err != nil {
        return "", err
    }
    invoice.Client = client
    
    pdfPath, err := s.pdfService.GenerateInvoice(invoice)
    if err != nil {
        return "", err
    }
    
    // Sauvegarder path
    invoice.PDFPath = pdfPath
    s.invoiceRepo.Update(id, invoice)
    
    // Audit
    s.auditRepo.Log(domain.AuditLog{
        EntityType: "invoice",
        EntityID:   id,
        Action:     "pdf_generated",
        NewValue:   fmt.Sprintf(`{"pdf_path": "%s"}`, pdfPath),
        Timestamp:  time.Now(),
    })
    
    return pdfPath, nil
}
```

### `/internal/repository`

**Responsabilité** : Accès données, SQL, migrations

#### `invoice_repository.go`

```go
package repository

import (
    "database/sql"
    "github.com/jmoiron/sqlx"
    "github.com/yourname/autogest/internal/domain"
)

type InvoiceRepository interface {
    Create(invoice *domain.Invoice) error
    Update(id int, invoice *domain.Invoice) error
    GetByID(id int) (*domain.Invoice, error)
    List(filters InvoiceFilters) ([]domain.Invoice, error)
    GetLastSequence(year int) (int, error)
    NumberExists(number string) (bool, error)
}

type invoiceRepository struct {
    db *sqlx.DB
}

func NewInvoiceRepository(db *sqlx.DB) InvoiceRepository {
    return &invoiceRepository{db: db}
}

func (r *invoiceRepository) Create(inv *domain.Invoice) error {
    query := `
        INSERT INTO invoices (
            number, client_id, issue_date, due_date, delivery_date,
            state, total_ht, total_ttc, vat_amount, vat_applicable,
            vat_exemption_text, payment_deadline, late_penalty_rate,
            recovery_fee, operation_category, notes
        ) VALUES (
            ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
        )
    `
    
    result, err := r.db.Exec(query,
        inv.Number, inv.ClientID, inv.IssueDate, inv.DueDate, inv.DeliveryDate,
        inv.State, inv.TotalHT, inv.TotalTTC, inv.VATAmount, inv.VATApplicable,
        inv.VATExemptionText, inv.PaymentDeadline, inv.LatePenaltyRate,
        inv.RecoveryFee, inv.OperationCategory, inv.Notes,
    )
    if err != nil {
        return err
    }
    
    id, err := result.LastInsertId()
    if err != nil {
        return err
    }
    inv.ID = int(id)
    
    // Insérer lignes
    for i := range inv.Lines {
        inv.Lines[i].InvoiceID = inv.ID
        if err := r.createLine(&inv.Lines[i]); err != nil {
            return err
        }
    }
    
    return nil
}

func (r *invoiceRepository) GetLastSequence(year int) (int, error) {
    var seq sql.NullInt64
    
    query := `
        SELECT MAX(CAST(SUBSTR(number, INSTR(number, '-') + 1) AS INTEGER))
        FROM invoices
        WHERE number LIKE ?
    `
    
    err := r.db.Get(&seq, query, fmt.Sprintf("%d-%%", year))
    if err != nil && err != sql.ErrNoRows {
        return 0, err
    }
    
    if !seq.Valid {
        return 0, nil
    }
    
    return int(seq.Int64), nil
}

func (r *invoiceRepository) NumberExists(number string) (bool, error) {
    var count int
    query := `SELECT COUNT(*) FROM invoices WHERE number = ?`
    
    err := r.db.Get(&count, query, number)
    return count > 0, err
}
```

### `/internal/tui`

**Responsabilité** : Interface utilisateur Bubbletea

#### `app.go`

```go
package tui

import (
    "github.com/charmbracelet/bubbletea"
    "github.com/yourname/autogest/internal/config"
    "github.com/yourname/autogest/internal/service"
    "github.com/yourname/autogest/internal/tui/views"
)

// ViewType représente le type de vue active
type ViewType int

const (
    ViewClients ViewType = iota
    ViewInvoices
    ViewQuotes
    ViewDashboard
)

// App est le modèle principal de l'application TUI
type App struct {
    services *service.Services
    config   *config.Config
    
    currentView ViewType
    views       map[ViewType]tea.Model
    
    width  int
    height int
}

func NewApp(services *service.Services, cfg *config.Config) *App {
    app := &App{
        services: services,
        config:   cfg,
        views:    make(map[ViewType]tea.Model),
    }
    
    // Initialiser vues
    app.views[ViewClients] = views.NewClientsView(services.ClientService)
    app.views[ViewInvoices] = views.NewInvoicesView(services.InvoiceService)
    app.views[ViewQuotes] = views.NewQuotesView(services.QuoteService)
    app.views[ViewDashboard] = views.NewDashboardView(services)
    
    app.currentView = ViewDashboard
    
    return app
}

func (m *App) Init() tea.Cmd {
    return m.getCurrentViewModel().Init()
}

func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        return m, nil
    
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        
        // Navigation entre vues
        case ":":
            // Activer command mode
            return m, nil
        }
    }
    
    // Déléguer au modèle de vue actif
    updatedView, cmd := m.getCurrentViewModel().Update(msg)
    m.views[m.currentView] = updatedView
    
    return m, cmd
}

func (m *App) View() string {
    return m.getCurrentViewModel().View()
}

func (m *App) getCurrentViewModel() tea.Model {
    return m.views[m.currentView]
}
```

---

## 🗄️ Schéma de Base de Données

Voir `migrations/001_initial.sql` pour le schéma SQL complet.

**Tables Principales** :
- `clients` : Clients
- `invoices` : Factures
- `invoice_lines` : Lignes de factures
- `quotes` : Devis
- `quote_lines` : Lignes de devis
- `credit_notes` : Avoirs
- `audit_log` : Journal d'audit

---

## 🔌 Dépendances et Justifications

```go
require (
    // TUI Framework - Bubbletea & family
    github.com/charmbracelet/bubbletea v0.25.0  // Framework TUI
    github.com/charmbracelet/bubbles v0.18.0    // Composants (table, list, etc.)
    github.com/charmbracelet/lipgloss v0.9.1    // Styling
    
    // Database
    github.com/mattn/go-sqlite3 v1.14.22        // SQLite driver
    github.com/jmoiron/sqlx v1.3.5              // SQL helpers
    
    // PDF
    github.com/johnfercher/maroto/v2 v2.0.0     // Génération PDF
    
    // Types sûrs pour argent (CRUCIAL)
    github.com/shopspring/decimal v1.3.1        // Précision décimale
    
    // Validation
    github.com/go-playground/validator/v10 v10.19.0
    
    // Configuration
    github.com/spf13/viper v1.18.2
)
```

---

## 🧪 Stratégie de Tests

### Tests Unitaires

```go
// domain/invoice_test.go
func TestInvoice_CanEdit(t *testing.T) {
    tests := []struct {
        name  string
        state domain.InvoiceState
        want  bool
    }{
        {"draft can edit", domain.InvoiceStateDraft, true},
        {"issued can edit", domain.InvoiceStateIssued, true},
        {"paid cannot edit", domain.InvoiceStatePaid, false},  // CRITIQUE
        {"canceled cannot edit", domain.InvoiceStateCanceled, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            inv := &domain.Invoice{State: tt.state}
            if got := inv.CanEdit(); got != tt.want {
                t.Errorf("CanEdit() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Tests d'Intégration

```go
// service/invoice_service_test.go
func TestInvoiceService_Update_ImmutableWhenPaid(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer db.Close()
    
    service := setupInvoiceService(db)
    
    // Créer facture et la marquer comme payée
    invoice := createTestInvoice()
    service.Create(invoice)
    service.MarkAsPaid(invoice.ID, time.Now())
    
    // Tenter modification
    invoice.TotalHT = decimal.NewFromInt(9999)
    err := service.Update(invoice.ID, invoice)
    
    // Vérifier que modification est refusée
    assert.Error(t, err)
    assert.IsType(t, &ErrImmutableInvoice{}, err)
}
```

---

**Dernière Mise à Jour** : 2026-03-04
