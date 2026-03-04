package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kvitrvn/runar/internal/domain"
	"github.com/shopspring/decimal"
)

// InvoiceFilters définit les filtres pour la liste des factures.
type InvoiceFilters struct {
	State    string
	ClientID int
	Year     int
	Search   string
}

// InvoiceRepository définit les opérations sur les factures.
type InvoiceRepository interface {
	Create(invoice *domain.Invoice) error
	Update(id int, invoice *domain.Invoice) error
	GetByID(id int) (*domain.Invoice, error)
	List(filters InvoiceFilters) ([]domain.Invoice, error)
	GetLastSequence(year int) (int, error)
	NumberExists(number string) (bool, error)
	// LEGAL: Pas de suppression physique avant 10 ans
	SoftDelete(id int) error
}

// invoiceRow représente une facture en SQL.
type invoiceRow struct {
	ID                  int            `db:"id"`
	Number              string         `db:"number"`
	ClientID            int            `db:"client_id"`
	QuoteID             sql.NullInt64  `db:"quote_id"`
	IssueDate           time.Time      `db:"issue_date"`
	DueDate             time.Time      `db:"due_date"`
	DeliveryDate        time.Time      `db:"delivery_date"`
	PaidDate            sql.NullTime   `db:"paid_date"`
	PaidLockedAt        sql.NullTime   `db:"paid_locked_at"`
	State               string         `db:"state"`
	TotalHT             string         `db:"total_ht"`
	TotalTTC            string         `db:"total_ttc"`
	VATAmount           string         `db:"vat_amount"`
	VATApplicable       bool           `db:"vat_applicable"`
	VATExemptionText    string         `db:"vat_exemption_text"`
	PaymentDeadline     string         `db:"payment_deadline"`
	LatePenaltyRate     string         `db:"late_penalty_rate"`
	RecoveryFee         string         `db:"recovery_fee"`
	EarlyPaymentDisc    sql.NullString `db:"early_payment_disc"`
	OperationCategory   sql.NullString `db:"operation_category"`
	DeliveryAddress     sql.NullString `db:"delivery_address"`
	Notes               sql.NullString `db:"notes"`
	PDFPath             sql.NullString `db:"pdf_path"`
	CreatedAt           time.Time      `db:"created_at"`
	UpdatedAt           time.Time      `db:"updated_at"`
	// Colonnes migration 003 (e-facturation 2027)
	EInvoiceFormat   sql.NullString `db:"e_invoice_format"`
	EInvoiceXML      sql.NullString `db:"e_invoice_xml"`
	EInvoiceSentAt   sql.NullTime   `db:"e_invoice_sent_at"`
	EInvoicePDPRef   sql.NullString `db:"e_invoice_pdp_ref"`
	VATPaymentOption sql.NullString `db:"vat_payment_option"`
	// Colonnes migration 004 (libellé virement)
	PaymentRef string `db:"payment_ref"`
	// Chargé par LEFT JOIN dans List() uniquement
	ClientName sql.NullString `db:"client_name"`
}

func (r invoiceRow) toDomain() domain.Invoice {
	inv := domain.Invoice{
		ID:               r.ID,
		Number:           r.Number,
		ClientID:         r.ClientID,
		IssueDate:        r.IssueDate,
		DueDate:          r.DueDate,
		DeliveryDate:     r.DeliveryDate,
		State:            domain.InvoiceState(r.State),
		VATApplicable:    r.VATApplicable,
		VATExemptionText: r.VATExemptionText,
		PaymentDeadline:  r.PaymentDeadline,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
	if r.QuoteID.Valid {
		id := int(r.QuoteID.Int64)
		inv.QuoteID = &id
	}
	if r.PaidDate.Valid {
		inv.PaidDate = &r.PaidDate.Time
	}
	if r.PaidLockedAt.Valid {
		inv.PaidLockedAt = &r.PaidLockedAt.Time
	}
	if r.EarlyPaymentDisc.Valid {
		inv.EarlyPaymentDiscount = r.EarlyPaymentDisc.String
	}
	if r.OperationCategory.Valid {
		inv.OperationCategory = domain.OperationCategory(r.OperationCategory.String)
	}
	if r.DeliveryAddress.Valid {
		inv.DeliveryAddress = r.DeliveryAddress.String
	}
	if r.Notes.Valid {
		inv.Notes = r.Notes.String
	}
	if r.PDFPath.Valid {
		inv.PDFPath = r.PDFPath.String
	}
	inv.TotalHT, _ = decimal.NewFromString(r.TotalHT)
	inv.TotalTTC, _ = decimal.NewFromString(r.TotalTTC)
	inv.VATAmount, _ = decimal.NewFromString(r.VATAmount)
	inv.LatePenaltyRate, _ = decimal.NewFromString(r.LatePenaltyRate)
	inv.RecoveryFee, _ = decimal.NewFromString(r.RecoveryFee)
	inv.PaymentRef = r.PaymentRef
	if r.ClientName.Valid && r.ClientName.String != "" {
		inv.Client = &domain.Client{Name: r.ClientName.String}
	}
	return inv
}

type invoiceRepository struct {
	db *sqlx.DB
}

// NewInvoiceRepository crée un repository facture.
func NewInvoiceRepository(db *sqlx.DB) InvoiceRepository {
	return &invoiceRepository{db: db}
}

func (r *invoiceRepository) Create(inv *domain.Invoice) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	query := `
		INSERT INTO invoices (
			number, client_id, quote_id, issue_date, due_date, delivery_date,
			state, total_ht, total_ttc, vat_amount, vat_applicable,
			vat_exemption_text, payment_deadline, late_penalty_rate,
			recovery_fee, early_payment_disc, operation_category,
			delivery_address, notes, payment_ref
		) VALUES (
			:number, :client_id, :quote_id, :issue_date, :due_date, :delivery_date,
			:state, :total_ht, :total_ttc, :vat_amount, :vat_applicable,
			:vat_exemption_text, :payment_deadline, :late_penalty_rate,
			:recovery_fee, :early_payment_disc, :operation_category,
			:delivery_address, :notes, :payment_ref
		)
	`

	var quoteID interface{}
	if inv.QuoteID != nil {
		quoteID = *inv.QuoteID
	}

	row := map[string]interface{}{
		"number":             inv.Number,
		"client_id":          inv.ClientID,
		"quote_id":           quoteID,
		"issue_date":         inv.IssueDate,
		"due_date":           inv.DueDate,
		"delivery_date":      inv.DeliveryDate,
		"state":              string(inv.State),
		"total_ht":           inv.TotalHT.String(),
		"total_ttc":          inv.TotalTTC.String(),
		"vat_amount":         inv.VATAmount.String(),
		"vat_applicable":     inv.VATApplicable,
		"vat_exemption_text": inv.VATExemptionText,
		"payment_deadline":   inv.PaymentDeadline,
		"late_penalty_rate":  inv.LatePenaltyRate.String(),
		"recovery_fee":       inv.RecoveryFee.String(),
		"early_payment_disc": nilIfEmpty(inv.EarlyPaymentDiscount),
		"operation_category": nilIfEmpty(string(inv.OperationCategory)),
		"delivery_address":   nilIfEmpty(inv.DeliveryAddress),
		"notes":              nilIfEmpty(inv.Notes),
		"payment_ref":        inv.PaymentRef,
	}

	result, err := tx.NamedExec(query, row)
	if err != nil {
		return fmt.Errorf("création facture: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	inv.ID = int(id)

	// Insérer les lignes
	for i := range inv.Lines {
		inv.Lines[i].InvoiceID = inv.ID
		if err := r.createLine(tx, &inv.Lines[i]); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *invoiceRepository) createLine(tx *sqlx.Tx, line *domain.InvoiceLine) error {
	query := `
		INSERT INTO invoice_lines (invoice_id, line_order, description, quantity, unit_price_ht, vat_rate, total_ht, total_ttc)
		VALUES (:invoice_id, :line_order, :description, :quantity, :unit_price_ht, :vat_rate, :total_ht, :total_ttc)
	`
	row := map[string]interface{}{
		"invoice_id":   line.InvoiceID,
		"line_order":   line.LineOrder,
		"description":  line.Description,
		"quantity":     line.Quantity.String(),
		"unit_price_ht": line.UnitPriceHT.String(),
		"vat_rate":     line.VATRate.String(),
		"total_ht":     line.TotalHT.String(),
		"total_ttc":    line.TotalTTC.String(),
	}
	result, err := tx.NamedExec(query, row)
	if err != nil {
		return fmt.Errorf("création ligne facture: %w", err)
	}
	id, _ := result.LastInsertId()
	line.ID = int(id)
	return nil
}

func (r *invoiceRepository) Update(id int, inv *domain.Invoice) error {
	query := `
		UPDATE invoices SET
			client_id = :client_id, due_date = :due_date, delivery_date = :delivery_date,
			paid_date = :paid_date, paid_locked_at = :paid_locked_at,
			state = :state, total_ht = :total_ht, total_ttc = :total_ttc,
			vat_amount = :vat_amount, vat_applicable = :vat_applicable,
			vat_exemption_text = :vat_exemption_text, payment_deadline = :payment_deadline,
			late_penalty_rate = :late_penalty_rate, recovery_fee = :recovery_fee,
			early_payment_disc = :early_payment_disc, operation_category = :operation_category,
			delivery_address = :delivery_address, notes = :notes, pdf_path = :pdf_path,
			payment_ref = :payment_ref, updated_at = CURRENT_TIMESTAMP
		WHERE id = :id
	`
	row := map[string]interface{}{
		"id":                 id,
		"client_id":          inv.ClientID,
		"due_date":           inv.DueDate,
		"delivery_date":      inv.DeliveryDate,
		"paid_date":          inv.PaidDate,
		"paid_locked_at":     inv.PaidLockedAt,
		"state":              string(inv.State),
		"total_ht":           inv.TotalHT.String(),
		"total_ttc":          inv.TotalTTC.String(),
		"vat_amount":         inv.VATAmount.String(),
		"vat_applicable":     inv.VATApplicable,
		"vat_exemption_text": inv.VATExemptionText,
		"payment_deadline":   inv.PaymentDeadline,
		"late_penalty_rate":  inv.LatePenaltyRate.String(),
		"recovery_fee":       inv.RecoveryFee.String(),
		"early_payment_disc": nilIfEmpty(inv.EarlyPaymentDiscount),
		"operation_category": nilIfEmpty(string(inv.OperationCategory)),
		"delivery_address":   nilIfEmpty(inv.DeliveryAddress),
		"notes":              nilIfEmpty(inv.Notes),
		"pdf_path":           nilIfEmpty(inv.PDFPath),
		"payment_ref":        inv.PaymentRef,
	}
	_, err := r.db.NamedExec(query, row)
	return err
}

func (r *invoiceRepository) GetByID(id int) (*domain.Invoice, error) {
	var row invoiceRow
	err := r.db.Get(&row, "SELECT * FROM invoices WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("facture %d introuvable: %w", id, err)
	}
	inv := row.toDomain()
	lines, err := r.getLines(inv.ID)
	if err != nil {
		return nil, err
	}
	inv.Lines = lines
	return &inv, nil
}

func (r *invoiceRepository) getLines(invoiceID int) ([]domain.InvoiceLine, error) {
	type lineRow struct {
		ID          int    `db:"id"`
		InvoiceID   int    `db:"invoice_id"`
		LineOrder   int    `db:"line_order"`
		Description string `db:"description"`
		Quantity    string `db:"quantity"`
		UnitPriceHT string `db:"unit_price_ht"`
		VATRate     string `db:"vat_rate"`
		TotalHT     string `db:"total_ht"`
		TotalTTC    string `db:"total_ttc"`
	}

	var rows []lineRow
	err := r.db.Select(&rows, "SELECT * FROM invoice_lines WHERE invoice_id = ? ORDER BY line_order", invoiceID)
	if err != nil {
		return nil, err
	}

	lines := make([]domain.InvoiceLine, len(rows))
	for i, row := range rows {
		lines[i] = domain.InvoiceLine{
			ID:          row.ID,
			InvoiceID:   row.InvoiceID,
			LineOrder:   row.LineOrder,
			Description: row.Description,
		}
		lines[i].Quantity, _ = decimal.NewFromString(row.Quantity)
		lines[i].UnitPriceHT, _ = decimal.NewFromString(row.UnitPriceHT)
		lines[i].VATRate, _ = decimal.NewFromString(row.VATRate)
		lines[i].TotalHT, _ = decimal.NewFromString(row.TotalHT)
		lines[i].TotalTTC, _ = decimal.NewFromString(row.TotalTTC)
	}
	return lines, nil
}

func (r *invoiceRepository) List(filters InvoiceFilters) ([]domain.Invoice, error) {
	query := `
		SELECT invoices.*, clients.name AS client_name
		FROM invoices
		LEFT JOIN clients ON invoices.client_id = clients.id
		WHERE 1=1`
	args := []interface{}{}

	if filters.State != "" {
		query += " AND invoices.state = ?"
		args = append(args, filters.State)
	}
	if filters.ClientID > 0 {
		query += " AND invoices.client_id = ?"
		args = append(args, filters.ClientID)
	}
	if filters.Year > 0 {
		query += " AND strftime('%Y', invoices.issue_date) = ?"
		args = append(args, fmt.Sprintf("%d", filters.Year))
	}
	if filters.Search != "" {
		query += " AND (invoices.number LIKE ? OR clients.name LIKE ?)"
		args = append(args, "%"+filters.Search+"%", "%"+filters.Search+"%")
	}
	query += " ORDER BY invoices.issue_date DESC, invoices.number DESC"

	var rows []invoiceRow
	if err := r.db.Select(&rows, query, args...); err != nil {
		return nil, err
	}

	invoices := make([]domain.Invoice, len(rows))
	for i, row := range rows {
		invoices[i] = row.toDomain()
	}
	return invoices, nil
}

// GetLastSequence retourne le dernier numéro de séquence pour une année donnée.
// LEGAL: Utilisé pour garantir la continuité de la numérotation (Art. 242 nonies A CGI).
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

// NumberExists vérifie si un numéro de facture existe déjà.
// LEGAL: Empêche les doublons de numéro (Art. 242 nonies A CGI).
func (r *invoiceRepository) NumberExists(number string) (bool, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM invoices WHERE number = ?", number)
	return count > 0, err
}

// SoftDelete effectue une suppression logique.
// LEGAL: Pas de suppression physique avant 10 ans (Art. L123-22 Code de Commerce).
func (r *invoiceRepository) SoftDelete(id int) error {
	// TODO: Implémenter la suppression logique avec vérification de la date d'expiration
	return fmt.Errorf("suppression non implémentée: les factures doivent être conservées 10 ans")
}

// nilIfEmpty retourne nil si la string est vide, sinon la valeur.
func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
