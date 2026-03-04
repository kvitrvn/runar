package repository

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kvitrvn/runar/internal/domain"
)

// ClientRepository définit les opérations sur les clients.
type ClientRepository interface {
	Create(client *domain.Client) error
	Update(id int, client *domain.Client) error
	GetByID(id int) (*domain.Client, error)
	List(search string) ([]domain.Client, error)
	Delete(id int) error
}

// clientRow est la représentation SQL d'un client.
type clientRow struct {
	ID         int       `db:"id"`
	Name       string    `db:"name"`
	SIRET      string    `db:"siret"`
	SIREN      string    `db:"siren"`
	VATNumber  string    `db:"vat_number"`
	Address    string    `db:"address"`
	PostalCode string    `db:"postal_code"`
	City       string    `db:"city"`
	Country    string    `db:"country"`
	Email      string    `db:"email"`
	Phone      string    `db:"phone"`
	Notes      string    `db:"notes"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

func (r clientRow) toDomain() domain.Client {
	return domain.Client{
		ID:         r.ID,
		Name:       r.Name,
		SIRET:      r.SIRET,
		SIREN:      r.SIREN,
		VATNumber:  r.VATNumber,
		Address:    r.Address,
		PostalCode: r.PostalCode,
		City:       r.City,
		Country:    r.Country,
		Email:      r.Email,
		Phone:      r.Phone,
		Notes:      r.Notes,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

type clientRepository struct {
	db *sqlx.DB
}

// NewClientRepository crée un repository client.
func NewClientRepository(db *sqlx.DB) ClientRepository {
	return &clientRepository{db: db}
}

func (r *clientRepository) Create(c *domain.Client) error {
	query := `
		INSERT INTO clients (name, siret, siren, vat_number, address, postal_code, city, country, email, phone, notes)
		VALUES (:name, :siret, :siren, :vat_number, :address, :postal_code, :city, :country, :email, :phone, :notes)
	`
	row := map[string]interface{}{
		"name":        c.Name,
		"siret":       c.SIRET,
		"siren":       c.SIREN,
		"vat_number":  c.VATNumber,
		"address":     c.Address,
		"postal_code": c.PostalCode,
		"city":        c.City,
		"country":     c.Country,
		"email":       c.Email,
		"phone":       c.Phone,
		"notes":       c.Notes,
	}
	result, err := r.db.NamedExec(query, row)
	if err != nil {
		return fmt.Errorf("création client: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	c.ID = int(id)
	return nil
}

func (r *clientRepository) Update(id int, c *domain.Client) error {
	query := `
		UPDATE clients SET
			name = :name, siret = :siret, siren = :siren, vat_number = :vat_number,
			address = :address, postal_code = :postal_code, city = :city, country = :country,
			email = :email, phone = :phone, notes = :notes,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = :id
	`
	row := map[string]interface{}{
		"id":          id,
		"name":        c.Name,
		"siret":       c.SIRET,
		"siren":       c.SIREN,
		"vat_number":  c.VATNumber,
		"address":     c.Address,
		"postal_code": c.PostalCode,
		"city":        c.City,
		"country":     c.Country,
		"email":       c.Email,
		"phone":       c.Phone,
		"notes":       c.Notes,
	}
	_, err := r.db.NamedExec(query, row)
	return err
}

func (r *clientRepository) GetByID(id int) (*domain.Client, error) {
	var row clientRow
	err := r.db.Get(&row, "SELECT * FROM clients WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("client %d introuvable: %w", id, err)
	}
	c := row.toDomain()
	return &c, nil
}

func (r *clientRepository) List(search string) ([]domain.Client, error) {
	var rows []clientRow
	var err error
	if search == "" {
		err = r.db.Select(&rows, "SELECT * FROM clients ORDER BY name")
	} else {
		like := "%" + search + "%"
		err = r.db.Select(&rows, `
			SELECT * FROM clients
			WHERE name LIKE ? OR siret LIKE ? OR email LIKE ?
			ORDER BY name
		`, like, like, like)
	}
	if err != nil {
		return nil, err
	}
	clients := make([]domain.Client, len(rows))
	for i, row := range rows {
		clients[i] = row.toDomain()
	}
	return clients, nil
}

func (r *clientRepository) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM clients WHERE id = ?", id)
	return err
}
