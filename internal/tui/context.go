package tui

import "time"

// ViewType identifie la vue courante.
type ViewType int

const (
	ViewPulse ViewType = iota
	ViewClients
	ViewInvoices
	ViewQuotes
	ViewCreditNotes
)

// String retourne le nom lisible de la vue.
func (v ViewType) String() string {
	switch v {
	case ViewPulse:
		return "pulse"
	case ViewClients:
		return "clients"
	case ViewInvoices:
		return "factures"
	case ViewQuotes:
		return "devis"
	case ViewCreditNotes:
		return "avoirs"
	default:
		return "inconnu"
	}
}

// AppMode représente le mode d'interaction courant.
type AppMode int

const (
	ModeNormal  AppMode = iota // Navigation standard (j/k/g/G)
	ModeCommand                // Saisie commande (après ':')
	ModeSearch                 // Saisie filtre (après '/')
	ModeHelp                   // Panneau d'aide affiché
)

// ToastType définit la sévérité d'un toast.
type ToastType int

const (
	ToastInfo ToastType = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// Toast est une notification temporaire non-bloquante.
type Toast struct {
	Message  string
	Type     ToastType
	Duration time.Duration
	ShowTime time.Time
}

// IsVisible indique si le toast doit encore être affiché.
func (t Toast) IsVisible() bool {
	return time.Since(t.ShowTime) < t.Duration
}

// NewToast crée un toast avec durée par défaut de 3 secondes.
func NewToast(msg string, kind ToastType) Toast {
	return Toast{
		Message:  msg,
		Type:     kind,
		Duration: 3 * time.Second,
		ShowTime: time.Now(),
	}
}
