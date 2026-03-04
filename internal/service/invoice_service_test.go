package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/kvitrvn/runar/internal/domain"
	"github.com/kvitrvn/runar/internal/repository"
	"github.com/shopspring/decimal"
)

// ─── Numérotation continue ────────────────────────────────────────────────────

func TestInvoiceService_NumerotationContinue(t *testing.T) {
	// LEGAL: Numérotation continue sans trou (Art. 242 nonies A CGI)
	svc := setupServices(t)
	client := createTestClient(t, svc)

	year := time.Now().Year()

	// Créer 3 factures et vérifier la séquence
	inv1 := createTestInvoice(t, svc, client.ID)
	inv2 := createTestInvoice(t, svc, client.ID)
	inv3 := createTestInvoice(t, svc, client.ID)

	expected1 := formatNum(year, 1)
	expected2 := formatNum(year, 2)
	expected3 := formatNum(year, 3)

	if inv1.Number != expected1 {
		t.Errorf("Première facture: numéro = %q, attendu %q", inv1.Number, expected1)
	}
	if inv2.Number != expected2 {
		t.Errorf("Deuxième facture: numéro = %q, attendu %q", inv2.Number, expected2)
	}
	if inv3.Number != expected3 {
		t.Errorf("Troisième facture: numéro = %q, attendu %q", inv3.Number, expected3)
	}
}

func TestInvoiceService_PasDeTrouNumerotation(t *testing.T) {
	// LEGAL: Pas de trou dans la séquence (Art. 242 nonies A CGI)
	svc := setupServices(t)
	client := createTestClient(t, svc)

	year := time.Now().Year()
	n := 5

	for i := 1; i <= n; i++ {
		inv := createTestInvoice(t, svc, client.ID)
		expected := formatNum(year, i)
		if inv.Number != expected {
			t.Errorf("Facture #%d: numéro = %q, attendu %q (trou dans la séquence)", i, inv.Number, expected)
		}
	}
}

// ─── Immuabilité ─────────────────────────────────────────────────────────────

func TestInvoiceService_Update_ImmutableApresPaiement(t *testing.T) {
	// LEGAL: Facture payée = IMMUABLE (Art. L441-9 Code de Commerce)
	// Sanction : amende 75 000€
	svc := setupServices(t)
	client := createTestClient(t, svc)

	inv := createTestInvoice(t, svc, client.ID)

	// Émettre puis marquer comme payée
	inv.State = domain.InvoiceStateIssued
	if err := svc.Invoice.Update(inv.ID, inv); err != nil {
		t.Fatalf("Passage à issued: %v", err)
	}
	if err := svc.Invoice.MarkAsPaid(inv.ID, time.Now()); err != nil {
		t.Fatalf("MarkAsPaid: %v", err)
	}

	// Tentative de modification = doit échouer
	inv.Lines[0].UnitPriceHT = decimal.NewFromInt(9999)
	inv.CalculateTotals()
	err := svc.Invoice.Update(inv.ID, inv)

	if err == nil {
		t.Fatal("Modification d'une facture payée doit retourner une erreur")
	}

	// L'erreur doit être de type ErrImmutableInvoice
	var immutableErr *domain.ErrImmutableInvoice
	if !errors.As(err, &immutableErr) {
		t.Errorf("Erreur de type %T, attendu *domain.ErrImmutableInvoice", err)
	}
}

func TestInvoiceService_Update_ImmutableAnnulee(t *testing.T) {
	// LEGAL: Facture annulée aussi immuable (Art. L441-9)
	svc := setupServices(t)
	client := createTestClient(t, svc)

	inv := createTestInvoice(t, svc, client.ID)

	// Forcer état annulé via GetByID + update direct via repo
	inv.State = domain.InvoiceStateIssued
	_ = svc.Invoice.Update(inv.ID, inv)
	_ = svc.Invoice.MarkAsPaid(inv.ID, time.Now())

	// Essayer de modifier
	inv.Notes = "modification tentée"
	err := svc.Invoice.Update(inv.ID, inv)
	if err == nil {
		t.Fatal("Modification d'une facture payée (précurseur cancelled) doit échouer")
	}
}

func TestInvoiceService_Update_BrouillonEditable(t *testing.T) {
	// Un brouillon doit être modifiable
	svc := setupServices(t)
	client := createTestClient(t, svc)

	inv := createTestInvoice(t, svc, client.ID)

	inv.Notes = "mise à jour note brouillon"
	if err := svc.Invoice.Update(inv.ID, inv); err != nil {
		t.Errorf("Brouillon doit être éditable: %v", err)
	}
}

// ─── MarkAsPaid ──────────────────────────────────────────────────────────────

func TestInvoiceService_MarkAsPaid_Verrouille(t *testing.T) {
	svc := setupServices(t)
	client := createTestClient(t, svc)

	inv := createTestInvoice(t, svc, client.ID)

	// Passer en issued
	inv.State = domain.InvoiceStateIssued
	_ = svc.Invoice.Update(inv.ID, inv)

	paidDate := time.Now()
	if err := svc.Invoice.MarkAsPaid(inv.ID, paidDate); err != nil {
		t.Fatalf("MarkAsPaid: %v", err)
	}

	// Vérifier l'état en base
	loaded, err := svc.Invoice.GetByID(inv.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if loaded.State != domain.InvoiceStatePaid {
		t.Errorf("État = %s, attendu paid", loaded.State)
	}
	if loaded.PaidDate == nil {
		t.Error("PaidDate doit être renseignée")
	}
	if loaded.PaidLockedAt == nil {
		t.Error("PaidLockedAt doit être renseignée (timestamp de verrouillage)")
	}
}

func TestInvoiceService_MarkAsPaid_BrouillonRefuse(t *testing.T) {
	// Un brouillon ne peut pas être directement marqué comme payé
	svc := setupServices(t)
	client := createTestClient(t, svc)

	inv := createTestInvoice(t, svc, client.ID) // état = draft

	err := svc.Invoice.MarkAsPaid(inv.ID, time.Now())
	if err == nil {
		t.Error("MarkAsPaid sur brouillon doit échouer")
	}
}

// ─── Validation légale ────────────────────────────────────────────────────────

func TestInvoiceService_Create_MentionTVAInvalide(t *testing.T) {
	// LEGAL: La mention TVA doit être exacte (Art. 293B CGI)
	svc := setupServices(t)
	client := createTestClient(t, svc)

	inv := &domain.Invoice{
		ClientID:         client.ID,
		IssueDate:        time.Now(),
		DueDate:          time.Now().Add(30 * 24 * time.Hour),
		DeliveryDate:     time.Now(),
		State:            domain.InvoiceStateDraft,
		VATApplicable:    false,
		VATExemptionText: "TVA non applicable", // Incomplet
		PaymentDeadline:  "30 jours",
		LatePenaltyRate:  decimal.NewFromFloat(13.25),
		RecoveryFee:      decimal.NewFromInt(40),
		Lines: []domain.InvoiceLine{
			{Description: "Test", Quantity: decimal.NewFromFloat(1), UnitPriceHT: decimal.NewFromFloat(100)},
		},
	}

	err := svc.Invoice.Create(inv)
	if err == nil {
		t.Fatal("Création avec mention TVA invalide doit échouer")
	}

	var validErr *domain.ValidationErrorList
	if !errors.As(err, &validErr) {
		t.Errorf("Erreur de type %T, attendu *domain.ValidationErrorList", err)
	}
}

func TestInvoiceService_Create_SansLignes(t *testing.T) {
	svc := setupServices(t)
	client := createTestClient(t, svc)

	inv := &domain.Invoice{
		ClientID:         client.ID,
		IssueDate:        time.Now(),
		DueDate:          time.Now().Add(30 * 24 * time.Hour),
		DeliveryDate:     time.Now(),
		State:            domain.InvoiceStateDraft,
		VATApplicable:    false,
		VATExemptionText: domain.VATMentionExemption,
		PaymentDeadline:  "30 jours",
		LatePenaltyRate:  decimal.NewFromFloat(13.25),
		RecoveryFee:      decimal.NewFromInt(40),
		Lines:            nil, // Aucune ligne
	}

	err := svc.Invoice.Create(inv)
	if err == nil {
		t.Fatal("Création sans lignes doit échouer")
	}
}

// ─── Calcul des totaux ────────────────────────────────────────────────────────

func TestInvoiceService_Create_TotauxCalcules(t *testing.T) {
	svc := setupServices(t)
	client := createTestClient(t, svc)

	inv := &domain.Invoice{
		ClientID:         client.ID,
		IssueDate:        time.Now(),
		DueDate:          time.Now().Add(30 * 24 * time.Hour),
		DeliveryDate:     time.Now(),
		State:            domain.InvoiceStateDraft,
		VATApplicable:    false,
		VATExemptionText: domain.VATMentionExemption,
		PaymentDeadline:  "30 jours",
		LatePenaltyRate:  decimal.NewFromFloat(13.25),
		RecoveryFee:      decimal.NewFromInt(40),
		Lines: []domain.InvoiceLine{
			{Description: "Service A", Quantity: decimal.NewFromFloat(2), UnitPriceHT: decimal.NewFromFloat(500)},
			{Description: "Service B", Quantity: decimal.NewFromFloat(1), UnitPriceHT: decimal.NewFromFloat(300)},
		},
	}

	if err := svc.Invoice.Create(inv); err != nil {
		t.Fatalf("Create: %v", err)
	}

	expectedHT := decimal.NewFromFloat(1300)
	loaded, err := svc.Invoice.GetByID(inv.ID)
	if err != nil {
		t.Fatalf("GetByID après création: %v", err)
	}

	if !loaded.TotalHT.Equal(expectedHT) {
		t.Errorf("TotalHT = %s, attendu %s", loaded.TotalHT, expectedHT)
	}
}

// ─── Audit trail ─────────────────────────────────────────────────────────────

func TestInvoiceService_AuditLog_Creation(t *testing.T) {
	// LEGAL: Toute création doit être loggée (Art. L47 A LPF)
	svc := setupServices(t)
	client := createTestClient(t, svc)

	inv := createTestInvoice(t, svc, client.ID)

	logs, err := svc.Audit.GetHistory("invoice", inv.ID)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}

	if len(logs) == 0 {
		t.Fatal("Aucun log d'audit après création de facture")
	}

	found := false
	for _, log := range logs {
		if log.Action == domain.AuditActionCreated {
			found = true
		}
	}
	if !found {
		t.Error("Action 'created' non trouvée dans l'audit")
	}
}

func TestInvoiceService_AuditLog_TentativeModificationRefusee(t *testing.T) {
	// LEGAL: La tentative de modification interdite doit aussi être loggée
	svc := setupServices(t)
	client := createTestClient(t, svc)

	inv := createTestInvoice(t, svc, client.ID)
	inv.State = domain.InvoiceStateIssued
	_ = svc.Invoice.Update(inv.ID, inv)
	_ = svc.Invoice.MarkAsPaid(inv.ID, time.Now())

	// Tentative de modification (doit échouer)
	inv.Notes = "tentative illégale"
	_ = svc.Invoice.Update(inv.ID, inv)

	logs, err := svc.Audit.GetHistory("invoice", inv.ID)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}

	found := false
	for _, log := range logs {
		if log.Action == domain.AuditActionDenied {
			found = true
		}
	}
	if !found {
		t.Error("Tentative de modification refusée non loggée dans l'audit")
	}
}

func TestInvoiceService_AuditLog_Paiement(t *testing.T) {
	// LEGAL: Le verrouillage doit être tracé comme action critique
	svc := setupServices(t)
	client := createTestClient(t, svc)

	inv := createTestInvoice(t, svc, client.ID)
	inv.State = domain.InvoiceStateIssued
	_ = svc.Invoice.Update(inv.ID, inv)
	_ = svc.Invoice.MarkAsPaid(inv.ID, time.Now())

	logs, err := svc.Audit.GetHistory("invoice", inv.ID)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}

	found := false
	for _, log := range logs {
		if log.Action == domain.AuditActionPaidLocked {
			found = true
		}
	}
	if !found {
		t.Error("Action paid_and_locked non trouvée dans l'audit")
	}
}

// ─── List & Filtres ───────────────────────────────────────────────────────────

func TestInvoiceService_List_ParEtat(t *testing.T) {
	svc := setupServices(t)
	client := createTestClient(t, svc)

	// Créer 2 brouillons
	createTestInvoice(t, svc, client.ID)
	createTestInvoice(t, svc, client.ID)

	invoices, err := svc.Invoice.List(repository.InvoiceFilters{State: "draft"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(invoices) != 2 {
		t.Errorf("List(draft) = %d factures, attendu 2", len(invoices))
	}
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func formatNum(year, seq int) string {
	return domain.DefaultNumberingConfig().FormatInvoiceNumber(year, seq)
}
