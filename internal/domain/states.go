package domain

// InvoiceState représente l'état d'une facture.
// LEGAL: Une facture payée ou annulée est IMMUABLE (Art. L441-9 Code de Commerce).
type InvoiceState string

const (
	InvoiceStateDraft    InvoiceState = "draft"    // Brouillon - éditable
	InvoiceStateIssued   InvoiceState = "issued"   // Émise - éditable sous conditions
	InvoiceStateSent     InvoiceState = "sent"     // Envoyée - éditable sous conditions
	InvoiceStatePaid     InvoiceState = "paid"     // Payée - IMMUABLE
	InvoiceStateCanceled InvoiceState = "canceled" // Annulée via avoir - IMMUABLE
)

// QuoteState représente l'état d'un devis.
type QuoteState string

const (
	QuoteStateDraft    QuoteState = "draft"    // Brouillon
	QuoteStateSent     QuoteState = "sent"     // Envoyé
	QuoteStateAccepted QuoteState = "accepted" // Accepté
	QuoteStateRefused  QuoteState = "refused"  // Refusé
	QuoteStateExpired  QuoteState = "expired"  // Expiré
)

// OperationCategory pour les obligations de facturation électronique 2027.
// LEGAL: Catégorie obligatoire sur les factures électroniques (à partir de 2027).
type OperationCategory string

const (
	OperationService OperationCategory = "service" // Prestation de service
	OperationGoods   OperationCategory = "goods"   // Vente de biens
	OperationMixed   OperationCategory = "mixed"   // Mixte
)

// VATMentionExemption est la mention légale EXACTE à utiliser en franchise TVA.
// LEGAL: Texte exact obligatoire, toute variation est une infraction (Art. 293B CGI).
const VATMentionExemption = "TVA non applicable, article 293B du CGI"

// DefaultRecoveryFee est l'indemnité forfaitaire de recouvrement légale.
// LEGAL: 40€ fixe, obligatoire sur toute facture B2B (Art. L441-6 Code de Commerce).
const DefaultRecoveryFee = "40"

// DefaultPaymentDeadline est le délai de paiement par défaut.
// LEGAL: Maximum 60 jours nets ou 45 jours fin de mois (Art. L441-6 Code de Commerce).
const DefaultPaymentDeadline = "30 jours"
