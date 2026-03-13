package domain

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

// ValidationError représente une erreur de validation avec amende potentielle.
// LEGAL: Chaque mention manquante = 15€ d'amende (Art. 1737 CGI).
type ValidationError struct {
	Field   string
	Message string
	Fine    decimal.Decimal // Amende potentielle en euros
}

func (e ValidationError) Error() string {
	if e.Fine.GreaterThan(decimal.Zero) {
		return fmt.Sprintf("%s: %s (Amende potentielle: %s€)", e.Field, e.Message, e.Fine.String())
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrorList regroupe plusieurs erreurs de validation.
type ValidationErrorList struct {
	Errors []ValidationError
}

func (e *ValidationErrorList) Error() string {
	if len(e.Errors) == 0 {
		return "aucune erreur"
	}
	msgs := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		msgs[i] = err.Error()
	}
	return strings.Join(msgs, "; ")
}

// TotalFine calcule l'amende totale potentielle, plafonnée à 25% du montant TTC.
// LEGAL: Plafond 25% du montant de la facture (Art. 1737 CGI).
func (e *ValidationErrorList) TotalFine(invoiceTTC decimal.Decimal) decimal.Decimal {
	total := decimal.Zero
	for _, err := range e.Errors {
		total = total.Add(err.Fine)
	}
	maxFine := invoiceTTC.Mul(decimal.NewFromFloat(0.25))
	if total.GreaterThan(maxFine) {
		return maxFine
	}
	return total
}

// ValidateSIRET vérifie la validité d'un SIRET (14 chiffres, algorithme Luhn).
// LEGAL: SIRET obligatoire pour le vendeur (Art. L441-3 Code de Commerce).
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

// ValidateSIREN vérifie la validité d'un SIREN (9 chiffres, algorithme Luhn).
// LEGAL: SIREN client obligatoire pour B2B depuis 2024 (Décret n° 2022-1299).
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

// ValidateEmail vérifie le format d'une adresse email.
func ValidateEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// NormalizeIBAN met l'IBAN en majuscules sans espaces.
func NormalizeIBAN(iban string) string {
	return strings.ToUpper(strings.Join(strings.Fields(iban), ""))
}

// ValidateIBAN vérifie la structure et la clé modulo 97 d'un IBAN.
func ValidateIBAN(iban string) bool {
	iban = NormalizeIBAN(iban)
	if len(iban) < 15 || len(iban) > 34 {
		return false
	}

	re := regexp.MustCompile(`^[A-Z]{2}[0-9]{2}[A-Z0-9]+$`)
	if !re.MatchString(iban) {
		return false
	}

	rearranged := iban[4:] + iban[:4]
	mod := 0
	for _, r := range rearranged {
		switch {
		case r >= '0' && r <= '9':
			mod = (mod*10 + int(r-'0')) % 97
		case r >= 'A' && r <= 'Z':
			value := int(r-'A') + 10
			mod = (mod*10 + value/10) % 97
			mod = (mod*10 + value%10) % 97
		default:
			return false
		}
	}

	return mod == 1
}

// NormalizeBIC met le BIC en majuscules sans espaces.
func NormalizeBIC(bic string) string {
	return strings.ToUpper(strings.Join(strings.Fields(bic), ""))
}

// ValidateBIC vérifie le format ISO 9362 (8 ou 11 caractères).
func ValidateBIC(bic string) bool {
	bic = NormalizeBIC(bic)
	re := regexp.MustCompile(`^[A-Z]{4}[A-Z]{2}[A-Z0-9]{2}([A-Z0-9]{3})?$`)
	return re.MatchString(bic)
}

// luhnCheck implémente l'algorithme de Luhn pour valider SIRET/SIREN.
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

// fineAmount retourne un montant d'amende standard de 15€.
func fineAmount() decimal.Decimal {
	return decimal.NewFromInt(15)
}
