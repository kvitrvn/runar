package config

// Seller contient les informations de l'auto-entrepreneur (vendeur).
// Ces informations sont obligatoires sur toutes les factures.
type Seller struct {
	Name       string `mapstructure:"name"`
	SIRET      string `mapstructure:"siret"`
	Address    string `mapstructure:"address"`
	PostalCode string `mapstructure:"postal_code"`
	City       string `mapstructure:"city"`
	Country    string `mapstructure:"country"`
	Email      string `mapstructure:"email"`
	Phone      string `mapstructure:"phone"`
	// VATNumber est le numéro TVA intracommunautaire (si assujetti TVA)
	VATNumber string `mapstructure:"vat_number"`
}
