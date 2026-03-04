package domain

import "time"

// Client représente un client de l'auto-entreprise.
type Client struct {
	ID         int
	Name       string
	SIRET      string // 14 chiffres pour entreprises
	SIREN      string // 9 chiffres (obligatoire B2B depuis 2024)
	VATNumber  string // N° TVA intracommunautaire si UE
	Address    string
	PostalCode string
	City       string
	Country    string
	Email      string
	Phone      string
	Notes      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// IsCompany retourne true si le client est une entreprise (B2B).
func (c *Client) IsCompany() bool {
	return c.SIRET != "" || c.SIREN != ""
}

// Validate vérifie la validité d'un client.
func (c *Client) Validate() []ValidationError {
	var errors []ValidationError

	if c.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Le nom du client est obligatoire",
			Fine:    fineAmount(),
		})
	}

	if c.SIRET != "" && !ValidateSIRET(c.SIRET) {
		errors = append(errors, ValidationError{
			Field:   "siret",
			Message: "SIRET invalide (14 chiffres avec algorithme Luhn)",
		})
	}

	// LEGAL: SIREN obligatoire si B2B depuis 2024 (Décret n° 2022-1299)
	if c.IsCompany() && c.SIREN == "" && c.SIRET == "" {
		errors = append(errors, ValidationError{
			Field:   "siren",
			Message: "SIREN obligatoire pour les clients professionnels (depuis 2024)",
			Fine:    fineAmount(),
		})
	}

	if c.SIREN != "" && !ValidateSIREN(c.SIREN) {
		errors = append(errors, ValidationError{
			Field:   "siren",
			Message: "SIREN invalide (9 chiffres avec algorithme Luhn)",
		})
	}

	if c.Email != "" && !ValidateEmail(c.Email) {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "Format email invalide",
		})
	}

	return errors
}
