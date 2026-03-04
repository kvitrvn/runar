package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kvitrvn/runar/internal/domain"
	"github.com/shopspring/decimal"
)

// QuoteRepository définit les opérations sur les devis.
type QuoteRepository interface {
	Create(quote *domain.Quote) error
	Update(id int, quote *domain.Quote) error
	GetByID(id int) (*domain.Quote, error)
	List(search string) ([]domain.Quote, error)
	GetLastSequence(year int) (int, error)
	NumberExists(number string) (bool, error)
}

// CreditNoteRepository définit les opérations sur les avoirs.
// LEGAL: Conservation 10 ans obligatoire, comme les factures.
type CreditNoteRepository interface {
	Create(cn *domain.CreditNote) error
	GetByID(id int) (*domain.CreditNote, error)
	GetByInvoiceID(invoiceID int) ([]domain.CreditNote, error)
	List() ([]domain.CreditNote, error)
	GetLastSequence(year int) (int, error)
	UpdatePDFPath(id int, path string) error
}

type quoteRow struct {
	ID         int       `db:"id"`
	Number     string    `db:"number"`
	ClientID   int       `db:"client_id"`
	IssueDate  time.Time `db:"issue_date"`
	ExpiryDate time.Time `db:"expiry_date"`
	State      string    `db:"state"`
	TotalHT    string    `db:"total_ht"`
	TotalTTC   string    `db:"total_ttc"`
	VATAmount  string    `db:"vat_amount"`
	Notes      sql.NullString `db:"notes"`
	PDFPath    sql.NullString `db:"pdf_path"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type quoteRepository struct {
	db *sqlx.DB
}

// NewQuoteRepository crée un repository devis.
func NewQuoteRepository(db *sqlx.DB) QuoteRepository {
	return &quoteRepository{db: db}
}

func (r *quoteRepository) Create(q *domain.Quote) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	query := `
		INSERT INTO quotes (number, client_id, issue_date, expiry_date, state, total_ht, total_ttc, vat_amount, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := tx.Exec(query,
		q.Number, q.ClientID, q.IssueDate, q.ExpiryDate, string(q.State),
		q.TotalHT.String(), q.TotalTTC.String(), q.VATAmount.String(), nilIfEmpty(q.Notes),
	)
	if err != nil {
		return fmt.Errorf("création devis: %w", err)
	}
	id, _ := result.LastInsertId()
	q.ID = int(id)

	for i := range q.Lines {
		q.Lines[i].QuoteID = q.ID
		q.Lines[i].LineOrder = i + 1
		if err := createQuoteLine(tx, &q.Lines[i]); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func createQuoteLine(tx *sqlx.Tx, l *domain.QuoteLine) error {
	query := `
		INSERT INTO quote_lines (quote_id, line_order, description, quantity, unit_price_ht, vat_rate, total_ht, total_ttc)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := tx.Exec(query,
		l.QuoteID, l.LineOrder, l.Description,
		l.Quantity.String(), l.UnitPriceHT.String(), l.VATRate.String(),
		l.TotalHT.String(), l.TotalTTC.String(),
	)
	if err != nil {
		return fmt.Errorf("création ligne devis: %w", err)
	}
	lid, _ := result.LastInsertId()
	l.ID = int(lid)
	return nil
}

func (r *quoteRepository) getLines(quoteID int) ([]domain.QuoteLine, error) {
	type lineRow struct {
		ID          int    `db:"id"`
		QuoteID     int    `db:"quote_id"`
		LineOrder   int    `db:"line_order"`
		Description string `db:"description"`
		Quantity    string `db:"quantity"`
		UnitPriceHT string `db:"unit_price_ht"`
		VATRate     string `db:"vat_rate"`
		TotalHT     string `db:"total_ht"`
		TotalTTC    string `db:"total_ttc"`
	}
	var rows []lineRow
	if err := r.db.Select(&rows,
		"SELECT * FROM quote_lines WHERE quote_id = ? ORDER BY line_order", quoteID); err != nil {
		return nil, err
	}
	lines := make([]domain.QuoteLine, len(rows))
	for i, row := range rows {
		lines[i] = domain.QuoteLine{
			ID: row.ID, QuoteID: row.QuoteID, LineOrder: row.LineOrder,
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

func (r *quoteRepository) Update(id int, q *domain.Quote) error {
	_, err := r.db.Exec(`
		UPDATE quotes SET
			state = ?, total_ht = ?, total_ttc = ?, vat_amount = ?,
			notes = ?, pdf_path = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, string(q.State), q.TotalHT.String(), q.TotalTTC.String(), q.VATAmount.String(),
		nilIfEmpty(q.Notes), nilIfEmpty(q.PDFPath), id)
	return err
}

func (r *quoteRepository) GetByID(id int) (*domain.Quote, error) {
	var row quoteRow
	if err := r.db.Get(&row, "SELECT * FROM quotes WHERE id = ?", id); err != nil {
		return nil, fmt.Errorf("devis %d introuvable: %w", id, err)
	}
	q := domain.Quote{
		ID:         row.ID,
		Number:     row.Number,
		ClientID:   row.ClientID,
		IssueDate:  row.IssueDate,
		ExpiryDate: row.ExpiryDate,
		State:      domain.QuoteState(row.State),
		Notes:      row.Notes.String,
		PDFPath:    row.PDFPath.String,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
	q.TotalHT, _ = decimal.NewFromString(row.TotalHT)
	q.TotalTTC, _ = decimal.NewFromString(row.TotalTTC)
	q.VATAmount, _ = decimal.NewFromString(row.VATAmount)
	lines, err := r.getLines(q.ID)
	if err != nil {
		return nil, fmt.Errorf("chargement lignes devis %d: %w", id, err)
	}
	q.Lines = lines
	return &q, nil
}

func (r *quoteRepository) List(search string) ([]domain.Quote, error) {
	var rows []quoteRow
	var err error
	if search == "" {
		err = r.db.Select(&rows, "SELECT * FROM quotes ORDER BY issue_date DESC")
	} else {
		err = r.db.Select(&rows, "SELECT * FROM quotes WHERE number LIKE ? ORDER BY issue_date DESC", "%"+search+"%")
	}
	if err != nil {
		return nil, err
	}
	quotes := make([]domain.Quote, len(rows))
	for i, row := range rows {
		quotes[i] = domain.Quote{
			ID:         row.ID,
			Number:     row.Number,
			ClientID:   row.ClientID,
			IssueDate:  row.IssueDate,
			ExpiryDate: row.ExpiryDate,
			State:      domain.QuoteState(row.State),
		}
		quotes[i].TotalHT, _ = decimal.NewFromString(row.TotalHT)
		quotes[i].TotalTTC, _ = decimal.NewFromString(row.TotalTTC)
	}
	return quotes, nil
}

func (r *quoteRepository) GetLastSequence(year int) (int, error) {
	// Format: "DEV-2026-0001" → cherche après le deuxième '-'
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM quotes WHERE number LIKE ?", fmt.Sprintf("DEV-%d-%%", year))
	if err != nil || count == 0 {
		return 0, err
	}
	var maxNum string
	r.db.Get(&maxNum, "SELECT MAX(number) FROM quotes WHERE number LIKE ?", fmt.Sprintf("DEV-%d-%%", year)) //nolint:errcheck
	if maxNum == "" {
		return 0, nil
	}
	// Extraire le séquence du dernier numéro
	var seq int
	fmt.Sscanf(maxNum, "DEV-%d-%d", &year, &seq)
	return seq, nil
}

func (r *quoteRepository) NumberExists(number string) (bool, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM quotes WHERE number = ?", number)
	return count > 0, err
}

// creditNoteRepository implémente CreditNoteRepository.
type creditNoteRepository struct {
	db *sqlx.DB
}

// NewCreditNoteRepository crée un repository avoir.
func NewCreditNoteRepository(db *sqlx.DB) CreditNoteRepository {
	return &creditNoteRepository{db: db}
}

func (r *creditNoteRepository) Create(cn *domain.CreditNote) error {
	query := `
		INSERT INTO credit_notes (number, invoice_id, invoice_reference, issue_date, reason, total_ht, total_ttc, vat_amount)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query,
		cn.Number, cn.InvoiceID, cn.InvoiceReference,
		cn.IssueDate, cn.Reason,
		cn.TotalHT.String(), cn.TotalTTC.String(), cn.VATAmount.String(),
	)
	if err != nil {
		return fmt.Errorf("création avoir: %w", err)
	}
	id, _ := result.LastInsertId()
	cn.ID = int(id)
	return nil
}

func (r *creditNoteRepository) GetByID(id int) (*domain.CreditNote, error) {
	type cnRow struct {
		ID               int            `db:"id"`
		Number           string         `db:"number"`
		InvoiceID        int            `db:"invoice_id"`
		InvoiceReference string         `db:"invoice_reference"`
		IssueDate        time.Time      `db:"issue_date"`
		Reason           string         `db:"reason"`
		TotalHT          string         `db:"total_ht"`
		TotalTTC         string         `db:"total_ttc"`
		VATAmount        string         `db:"vat_amount"`
		PDFPath          sql.NullString `db:"pdf_path"`
		CreatedAt        time.Time      `db:"created_at"`
	}
	var row cnRow
	if err := r.db.Get(&row, "SELECT * FROM credit_notes WHERE id = ?", id); err != nil {
		return nil, fmt.Errorf("avoir %d introuvable: %w", id, err)
	}
	cn := &domain.CreditNote{
		ID:               row.ID,
		Number:           row.Number,
		InvoiceID:        row.InvoiceID,
		InvoiceReference: row.InvoiceReference,
		IssueDate:        row.IssueDate,
		Reason:           row.Reason,
		PDFPath:          row.PDFPath.String,
		CreatedAt:        row.CreatedAt,
	}
	cn.TotalHT, _ = decimal.NewFromString(row.TotalHT)
	cn.TotalTTC, _ = decimal.NewFromString(row.TotalTTC)
	cn.VATAmount, _ = decimal.NewFromString(row.VATAmount)
	return cn, nil
}

func (r *creditNoteRepository) GetByInvoiceID(invoiceID int) ([]domain.CreditNote, error) {
	type cnRow struct {
		ID               int            `db:"id"`
		Number           string         `db:"number"`
		InvoiceID        int            `db:"invoice_id"`
		InvoiceReference string         `db:"invoice_reference"`
		IssueDate        time.Time      `db:"issue_date"`
		Reason           string         `db:"reason"`
		TotalHT          string         `db:"total_ht"`
		TotalTTC         string         `db:"total_ttc"`
		VATAmount        string         `db:"vat_amount"`
		PDFPath          sql.NullString `db:"pdf_path"`
		CreatedAt        time.Time      `db:"created_at"`
	}
	var rows []cnRow
	if err := r.db.Select(&rows,
		"SELECT * FROM credit_notes WHERE invoice_id = ? ORDER BY issue_date DESC", invoiceID); err != nil {
		return nil, err
	}
	cns := make([]domain.CreditNote, len(rows))
	for i, row := range rows {
		cns[i] = domain.CreditNote{
			ID: row.ID, Number: row.Number, InvoiceID: row.InvoiceID,
			InvoiceReference: row.InvoiceReference, IssueDate: row.IssueDate,
			Reason: row.Reason, PDFPath: row.PDFPath.String, CreatedAt: row.CreatedAt,
		}
		cns[i].TotalHT, _ = decimal.NewFromString(row.TotalHT)
		cns[i].TotalTTC, _ = decimal.NewFromString(row.TotalTTC)
	}
	return cns, nil
}

func (r *creditNoteRepository) List() ([]domain.CreditNote, error) {
	type cnRow struct {
		ID               int            `db:"id"`
		Number           string         `db:"number"`
		InvoiceID        int            `db:"invoice_id"`
		InvoiceReference string         `db:"invoice_reference"`
		IssueDate        time.Time      `db:"issue_date"`
		Reason           string         `db:"reason"`
		TotalHT          string         `db:"total_ht"`
		TotalTTC         string         `db:"total_ttc"`
		VATAmount        string         `db:"vat_amount"`
		PDFPath          sql.NullString `db:"pdf_path"`
		CreatedAt        time.Time      `db:"created_at"`
	}
	var rows []cnRow
	if err := r.db.Select(&rows, "SELECT * FROM credit_notes ORDER BY issue_date DESC"); err != nil {
		return nil, err
	}
	cns := make([]domain.CreditNote, len(rows))
	for i, row := range rows {
		cns[i] = domain.CreditNote{
			ID: row.ID, Number: row.Number, InvoiceID: row.InvoiceID,
			InvoiceReference: row.InvoiceReference, IssueDate: row.IssueDate,
			Reason: row.Reason, PDFPath: row.PDFPath.String, CreatedAt: row.CreatedAt,
		}
		cns[i].TotalHT, _ = decimal.NewFromString(row.TotalHT)
		cns[i].TotalTTC, _ = decimal.NewFromString(row.TotalTTC)
	}
	return cns, nil
}

func (r *creditNoteRepository) UpdatePDFPath(id int, path string) error {
	_, err := r.db.Exec("UPDATE credit_notes SET pdf_path = ? WHERE id = ?", path, id)
	return err
}

func (r *creditNoteRepository) GetLastSequence(year int) (int, error) {
	var seq int
	err := r.db.Get(&seq, `
		SELECT COALESCE(MAX(CAST(SUBSTR(number, INSTR(SUBSTR(number, 3), '-') + 3) AS INTEGER)), 0)
		FROM credit_notes WHERE number LIKE ?
	`, fmt.Sprintf("A-%d-%%", year))
	return seq, err
}
