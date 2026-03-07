package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config est la configuration complète de l'application.
type Config struct {
	Seller   Seller  `mapstructure:"seller"`
	VAT      VATConf `mapstructure:"vat"`
	Payment  Payment `mapstructure:"payment"`
	Database DBConf  `mapstructure:"database"`
	PDF      PDFConf `mapstructure:"pdf"`
}

// VATConf configure le régime de TVA.
type VATConf struct {
	// Applicable indique si la TVA est facturée (false = franchise en base).
	// LEGAL: Si false, la mention "TVA non applicable, article 293B du CGI" est obligatoire.
	Applicable    bool    `mapstructure:"applicable"`
	ExemptionText string  `mapstructure:"exemption_text"`
	DefaultRate   float64 `mapstructure:"default_rate"`
}

// Payment configure les conditions de paiement par défaut.
// LEGAL: Délai max 60 jours nets ou 45 jours fin de mois (Art. L441-6 Code de Commerce).
type Payment struct {
	DefaultDeadline string  `mapstructure:"default_deadline"`
	LatePenaltyRate float64 `mapstructure:"late_penalty_rate"`
	// LEGAL: Indemnité forfaitaire 40€ obligatoire (Art. L441-6 Code de Commerce).
	RecoveryFee float64 `mapstructure:"recovery_fee"`
	// Coordonnées bancaires pour paiement par virement.
	IBAN string `mapstructure:"iban"`
	BIC  string `mapstructure:"bic"`
	// DefaultDepositRate est le taux d'acompte par défaut (0 = pas d'acompte).
	DefaultDepositRate float64 `mapstructure:"default_deposit_rate"`
}

// DBConf configure la base de données.
type DBConf struct {
	Path string `mapstructure:"path"`
}

// PDFConf configure la génération PDF.
type PDFConf struct {
	// OutputDir est le répertoire de stockage des PDFs.
	// LEGAL: Les PDFs ne doivent JAMAIS être supprimés (conservation 10 ans).
	OutputDir string `mapstructure:"output_dir"`
}

// Load charge la configuration depuis le fichier config.yaml.
func Load() (*Config, error) {
	v := viper.New()

	// Valeurs par défaut
	v.SetDefault("vat.applicable", false)
	v.SetDefault("vat.exemption_text", "TVA non applicable, article 293B du CGI")
	v.SetDefault("vat.default_rate", 0.0)
	v.SetDefault("payment.default_deadline", "30 jours")
	v.SetDefault("payment.late_penalty_rate", 13.25)
	v.SetDefault("payment.recovery_fee", 40.0)
	v.SetDefault("database.path", "./runar.db")
	v.SetDefault("pdf.output_dir", "./invoices")
	v.SetDefault("seller.name", "")
	v.SetDefault("seller.siret", "")
	v.SetDefault("seller.country", "France")
	v.SetDefault("payment.iban", "")
	v.SetDefault("payment.bic", "")
	v.SetDefault("payment.default_deposit_rate", 0.0)

	// Chercher config dans répertoire courant et ~/.config/runar/
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	homeDir, err := os.UserHomeDir()
	if err == nil {
		v.AddConfigPath(filepath.Join(homeDir, ".config", "runar"))
	}

	// Ignorer si fichier absent (on utilisera les valeurs par défaut)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("lecture config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

// WriteDefault écrit un fichier config.yaml d'exemple avec des données fictives.
func WriteDefault(path string) error {
	content := `# Configuration AutoGest
seller:
  name: "Jean Dupont"
  siret: "00000000000000"
  address: "1 rue de la Paix"
  postal_code: "75001"
  city: "Paris"
  country: "France"
  email: "contact@example.com"
  phone: "0600000000"

vat:
  applicable: false
  exemption_text: "TVA non applicable, article 293B du CGI"

payment:
  default_deadline: "30 jours"
  late_penalty_rate: 13.25
  recovery_fee: 40
  iban: "FR76 XXXX XXXX XXXX XXXX XXXX XXX"
  bic: "XXXXXXXX"
  default_deposit_rate: 30

database:
  path: "./runar.db"

pdf:
  output_dir: "./invoices"
`
	return os.WriteFile(path, []byte(content), 0o600)
}
