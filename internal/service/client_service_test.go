package service_test

import (
	"testing"

	"github.com/kvitrvn/runar/internal/domain"
)

func TestClientService_Create_Valide(t *testing.T) {
	svc := setupServices(t)
	c := createTestClient(t, svc)

	if c.ID == 0 {
		t.Error("ID du client doit être > 0 après création")
	}
}

func TestClientService_Create_NomManquant(t *testing.T) {
	svc := setupServices(t)
	c := &domain.Client{Name: ""} // Invalide
	err := svc.Client.Create(c)
	if err == nil {
		t.Fatal("Création client sans nom doit échouer")
	}
}

func TestClientService_GetByID(t *testing.T) {
	svc := setupServices(t)
	created := createTestClient(t, svc)

	loaded, err := svc.Client.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if loaded.Name != created.Name {
		t.Errorf("Name = %q, attendu %q", loaded.Name, created.Name)
	}
	if loaded.SIRET != created.SIRET {
		t.Errorf("SIRET = %q, attendu %q", loaded.SIRET, created.SIRET)
	}
}

func TestClientService_Update(t *testing.T) {
	svc := setupServices(t)
	c := createTestClient(t, svc)

	c.Email = "nouveau@email.fr"
	if err := svc.Client.Update(c.ID, c); err != nil {
		t.Fatalf("Update: %v", err)
	}

	loaded, _ := svc.Client.GetByID(c.ID)
	if loaded.Email != "nouveau@email.fr" {
		t.Errorf("Email = %q, attendu %q", loaded.Email, "nouveau@email.fr")
	}
}

func TestClientService_List(t *testing.T) {
	svc := setupServices(t)
	createTestClient(t, svc)
	createTestClient(t, svc)

	clients, err := svc.Client.List("")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(clients) < 2 {
		t.Errorf("List = %d clients, attendu au moins 2", len(clients))
	}
}

func TestClientService_AuditLog_Creation(t *testing.T) {
	// LEGAL: Toute création doit être loggée (Art. L47 A LPF)
	svc := setupServices(t)
	c := createTestClient(t, svc)

	logs, err := svc.Audit.GetHistory("client", c.ID)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if len(logs) == 0 {
		t.Fatal("Aucun log d'audit après création client")
	}
	if logs[0].Action != domain.AuditActionCreated {
		t.Errorf("Action = %q, attendu %q", logs[0].Action, domain.AuditActionCreated)
	}
}
