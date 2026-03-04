# 📜 LEGAL.md - Règles Juridiques Auto-Entrepreneur France

## ⚖️ Document de Référence Légale

Ce document liste TOUTES les obligations légales pour la facturation en auto-entreprise en France (2026-2027).
Toute implémentation DOIT respecter ces règles sous peine de sanctions.

**Sources** :
- Code de Commerce (articles L441-3, L441-9)
- Code Général des Impôts (article 293B)
- Décret n° 2022-1299 du 7 octobre 2022
- Ordonnance n° 2021-1190 (facturation électronique)

---

## 🚨 RÈGLES CRITIQUES - PRIORITÉ ABSOLUE

### 1. IMMUABILITÉ DES FACTURES PAYÉES

**Base Légale** : Article L441-9 du Code de Commerce

**Règle** : Une facture NE PEUT JAMAIS être modifiée ou supprimée après émission.

**Implémentation** :
```go
// État de la facture
type InvoiceState int

const (
    InvoiceDraft    InvoiceState = iota // Brouillon - éditable
    InvoiceIssued                       // Émise - éditable sous conditions
    InvoiceSent                         // Envoyée - éditable sous conditions
    InvoicePaid                         // PAYÉE - IMMUABLE ❌
    InvoiceCanceled                     // Annulée via avoir - IMMUABLE ❌
)

// Protection absolue
func (i *Invoice) CanEdit() bool {
    return i.State == InvoiceDraft
}

func (i *Invoice) CanDelete() bool {
    return i.State == InvoiceDraft // Seul le brouillon peut être supprimé
}
```

**Actions Autorisées sur Facture Payée** :
- ✅ Consultation
- ✅ Export PDF
- ✅ Création d'un avoir (annulation partielle ou totale)
- ❌ Modification quelconque
- ❌ Suppression

**Sanction** : Amende administrative jusqu'à 75 000 € (150 000 € en cas de récidive)

### 2. NUMÉROTATION CONTINUE OBLIGATOIRE

**Base Légale** : Article 242 nonies A de l'annexe II du CGI

**Règle** : Les factures doivent être numérotées de façon continue, chronologique, sans trou ni doublon.

**Format Accepté** :
- `2026-0001`, `2026-0002`, `2026-0003` ✅
- `FAC-2026-001`, `FAC-2026-002` ✅
- `2026-01-001` pour numérotation mensuelle ⚠️ (mais continuité annuelle obligatoire)

**Format Interdit** :
- Redémarrage à 0 chaque mois : `2026-01-001`, `2026-02-001` ❌
- Trous dans la séquence : `001, 002, 004` ❌
- Doublons : deux factures avec même numéro ❌

**Implémentation** :
```go
func (s *InvoiceService) GenerateNextNumber(year int) (string, error) {
    // Récupérer le dernier numéro de l'année
    lastSeq, err := s.repo.GetLastSequence(year)
    if err != nil {
        return "", err
    }
    
    nextSeq := lastSeq + 1
    number := fmt.Sprintf("%d-%04d", year, nextSeq)
    
    // CRITICAL: Vérifier qu'aucune facture n'existe avec ce numéro
    exists, err := s.repo.NumberExists(number)
    if err != nil {
        return "", err
    }
    if exists {
        return "", ErrDuplicateInvoiceNumber
    }
    
    return number, nil
}
```

**Sanction** : Amende de 15 € par facture non conforme

### 3. CONSERVATION 10 ANS

**Base Légale** : Article L123-22 du Code de Commerce

**Règle** : Toutes les factures (et leurs PDFs) doivent être conservées pendant 10 ans à compter de la clôture de l'exercice.

**Exemple** : Facture émise le 15 mars 2026 → Conservation jusqu'au 31 décembre 2036 minimum.

**Implémentation** :
```go
// Pas de suppression physique avant expiration
func (r *InvoiceRepository) Delete(id int) error {
    invoice, err := r.Get(id)
    if err != nil {
        return err
    }
    
    // Calcul date expiration
    expiryYear := invoice.IssueDate.Year() + 10
    expiryDate := time.Date(expiryYear, 12, 31, 23, 59, 59, 0, time.UTC)
    
    if time.Now().Before(expiryDate) {
        return &ErrCannotDeleteBeforeExpiry{
            InvoiceNumber: invoice.Number,
            ExpiryDate:    expiryDate,
        }
    }
    
    // Soft delete seulement
    return r.SoftDelete(id)
}
```

**Sanction** : Amende fiscale, rejet de comptabilité en cas de contrôle

---

## 📋 MENTIONS OBLIGATOIRES - TOUTES FACTURES

### Mentions de Base (Toujours Obligatoires)

| Mention | Champ | Validation | Amende |
|---------|-------|------------|--------|
| Mot "FACTURE" | `document_type` | `== "FACTURE"` | 15 € |
| Numéro unique | `number` | Regex + continuité | 15 € |
| Date d'émission | `issue_date` | Format date valide | 15 € |
| Date de livraison/fin prestation | `delivery_date` | `<= issue_date` | 15 € |
| Identité vendeur | `seller_name` | Non vide | 15 € |
| Adresse vendeur | `seller_address` | Non vide | 15 € |
| SIRET | `seller_siret` | 14 chiffres valides | 15 € |
| Email ou téléphone | `seller_contact` | Format valide | 15 € |
| Identité client | `customer_name` | Non vide | 15 € |
| Adresse client | `customer_address` | Non vide | 15 € |
| Description produits/services | `lines[].description` | Non vide par ligne | 15 € |
| Quantité | `lines[].quantity` | > 0 | 15 € |
| Prix unitaire HT | `lines[].unit_price_ht` | >= 0 | 15 € |
| Montant total HT | `total_ht` | Somme lignes | 15 € |
| Montant total TTC | `total_ttc` | Cohérent avec TVA | 15 € |

**Total Amende Potentielle** : jusqu'à 25% du montant de la facture (plafond)

### Mentions Franchise TVA (Auto-Entrepreneurs)

Si vous NE facturez PAS la TVA (franchise en base) :

```
✅ OBLIGATOIRE : "TVA non applicable, article 293B du CGI"
```

**Position** : Sur la facture, clairement visible
**Champ** : `vat_exemption_mention`
**Validation** : Texte exact (sensible à la casse)

**Erreur Fréquente** : ❌ "TVA non applicable" → Incomplet, amende 15 €

### Mentions si Assujetti TVA

Si vous dépassez les seuils et facturez la TVA :

| Mention | Champ | Obligation |
|---------|-------|------------|
| Numéro TVA intracommunautaire | `seller_vat_number` | Si facture > 150 € |
| Taux de TVA | `lines[].vat_rate` | Par ligne |
| Montant TVA | `vat_amount` | Calculé |
| Total TTC | `total_ttc` | HT + TVA |

**Seuils TVA 2026** :
- Vente de biens : 91 900 € (seuil majoré)
- Prestations de services : 39 100 € (seuil majoré)

### Mentions Paiement (Obligatoires)

```go
type PaymentTerms struct {
    DueDate         time.Time       // Date d'échéance
    PaymentDeadline string          // Ex: "30 jours" ou "45 jours fin de mois"
    LatePenaltyRate decimal.Decimal // Taux BCE + 10 points
    RecoveryFee     decimal.Decimal // 40 € forfaitaire
    EarlyPayment    string          // Escompte si paiement anticipé (optionnel)
}
```

**Délais Légaux** :
- Par défaut : 30 jours à compter de la date d'émission
- Maximum autorisé : 60 jours nets ou 45 jours fin de mois
- Dépassement = amende jusqu'à 75 000 €

**Pénalités de Retard** :
- Taux = Taux BCE (au 1er janvier ou 1er juillet) + 10 points
- Exemple (2026) : 3,25% + 10 = 13,25% annuel
- Application : Automatique dès le premier jour de retard

**Indemnité Forfaitaire** :
- Montant fixe : 40 € par facture impayée
- S'ajoute aux pénalités de retard

### Mentions Clients Professionnels (depuis 2024)

**Base Légale** : Décret n° 2022-1299 du 7 octobre 2022

Si votre client est un professionnel :

| Mention | Champ | Format | Depuis |
|---------|-------|--------|--------|
| SIREN client | `customer_siren` | 9 chiffres | 2024 |
| OU N° TVA intracommunautaire | `customer_vat_number` | Format UE | 2024 |
| Adresse de livraison | `delivery_address` | Si ≠ facturation | 2027 |
| Catégorie opération | `operation_category` | "service"/"goods"/"mixed" | 2027 |

**Validation SIREN** :
```go
func ValidateSIREN(siren string) error {
    // Retirer espaces
    siren = strings.ReplaceAll(siren, " ", "")
    
    // Vérifier 9 chiffres
    if len(siren) != 9 {
        return ErrInvalidSIRENLength
    }
    
    // Vérifier que ce sont des chiffres
    if _, err := strconv.Atoi(siren); err != nil {
        return ErrSIRENNotNumeric
    }
    
    // Algorithme Luhn (optionnel mais recommandé)
    if !luhnCheck(siren) {
        return ErrSIRENInvalidChecksum
    }
    
    return nil
}
```

### Mentions Assurance (Si Obligatoire)

Pour activités réglementées (BTP, santé, etc.) :

```go
type Insurance struct {
    CompanyName       string // Nom assureur
    PolicyNumber      string // N° contrat
    GeographicCoverage string // Ex: "France et UE"
}
```

**Activités Concernées** :
- BTP : Garantie décennale obligatoire
- Professions libérales de santé : RC Pro
- Agents immobiliers : Garantie financière

---

## 📅 NOUVELLES OBLIGATIONS 2027

### Facturation Électronique Obligatoire

**Calendrier** :
- 1er septembre 2026 : Réception de factures électroniques obligatoire
- 1er septembre 2027 : Émission de factures électroniques obligatoire

**Format Obligatoire** : Factur-X (PDF hybride avec données XML structurées)

**Plateforme** :
- Via PDP (Plateforme de Dématérialisation Partenaire) agréée
- OU via PPF (Portail Public de Facturation - Chorus Pro étendu)

**Implémentation Future** :
```go
type ElectronicInvoice struct {
    PDF        []byte // PDF visuel
    StructuredData XML  // Données structurées EN 16931
    Signature  []byte // Signature électronique
}
```

**E-Reporting** :
- Transmission automatique des données à la DGFiP
- Même pour ventes B2C (particuliers)
- Traçabilité complète du CA

### Nouvelles Mentions 2027

À partir du 1er juillet 2027 :

```go
type Invoice2027 struct {
    // ... mentions existantes
    
    // NOUVELLES mentions obligatoires
    CustomerSIREN      string // Déjà obligatoire depuis 2024
    DeliveryAddress    string // Si différente de facturation
    OperationCategory  string // "service", "goods", "mixed"
    VATPaymentOption   string // Si option débit au lieu encaissement
}
```

---

## 🔍 AVOIRS (CREDIT NOTES)

### Quand Émettre un Avoir ?

Un avoir doit être émis dans ces cas :

1. **Erreur sur facture payée** → Avoir + nouvelle facture
2. **Annulation totale** → Avoir à 100%
3. **Annulation partielle** → Avoir pour montant concerné
4. **Retour de marchandise**
5. **Remise accordée après facturation**

### Règles des Avoirs

**Base Légale** : Article 272 du CGI

```go
type CreditNote struct {
    Number           string    // Numérotation séparée ou continue
    InvoiceReference string    // OBLIGATOIRE : référence facture d'origine
    IssueDate        time.Time
    Reason           string    // Motif de l'avoir
    
    // Montants (NÉGATIFS)
    TotalHT          decimal.Decimal // Négatif
    TotalTTC         decimal.Decimal // Négatif
    
    Lines            []CreditNoteLine
}
```

**Mentions Obligatoires Avoir** :
- Mot "AVOIR" clairement visible
- Référence à la facture d'origine (numéro et date)
- Motif de l'avoir
- Mêmes mentions qu'une facture classique
- Montants négatifs ou indication "à déduire"

**Conservation** : 10 ans, comme les factures

### Numérotation des Avoirs

Deux options acceptées :

**Option 1** : Séquence séparée
```
Factures : 2026-0001, 2026-0002, 2026-0003
Avoirs   : A-2026-0001, A-2026-0002
```

**Option 2** : Séquence unique (préfixe)
```
2026-0001 (facture)
2026-0002 (facture)
A-2026-0003 (avoir)
2026-0004 (facture)
```

**Implémentation Recommandée** : Option 1 (séquences séparées) pour clarté

---

## 📊 CALCULS ET ARRONDIS

### Précision Décimale

**RÈGLE ABSOLUE** : Ne JAMAIS utiliser `float32` ou `float64` pour les montants.

```go
// ❌ INTERDIT
type Invoice struct {
    TotalHT float64  // Erreurs d'arrondi !
}

// ✅ OBLIGATOIRE
import "github.com/shopspring/decimal"

type Invoice struct {
    TotalHT decimal.Decimal
}
```

**Pourquoi ?** : Les float peuvent causer des erreurs d'arrondi : 0.1 + 0.2 ≠ 0.3

### Règles d'Arrondi

**TVA et totaux** : Arrondi au centime le plus proche (0.005 → 0.01)

```go
func RoundToNearestCent(amount decimal.Decimal) decimal.Decimal {
    return amount.Round(2) // 2 décimales
}
```

**Prix unitaires** : Peuvent avoir plus de décimales, arrondi final sur total ligne

---

## 🔒 AUDIT TRAIL - TRAÇABILITÉ

### Obligation de Traçabilité

**Base Légale** : Article L47 A du Livre des procédures fiscales

Toute modification d'une facture (avant paiement) doit être tracée :

```go
type AuditLog struct {
    ID          int
    EntityType  string    // "invoice", "quote", "client"
    EntityID    int       // ID de l'entité
    Action      string    // "created", "updated", "paid", "canceled"
    UserID      string    // Qui (toujours l'auto-entrepreneur ici)
    OldValue    string    // JSON de l'état avant
    NewValue    string    // JSON de l'état après
    Timestamp   time.Time
    IPAddress   string    // Optionnel mais recommandé
}
```

**Actions à Logger** :
- Création facture/devis/avoir
- Modification (si autorisée)
- Marquage comme payée
- Tentative de modification interdite (avec rejet)
- Génération PDF
- Export comptable

**Conservation** : 6 ans minimum (durée contrôle fiscal)

---

## ⚠️ SANCTIONS ET AMENDES

### Tableau Récapitulatif

| Infraction | Base Légale | Sanction | Récidive |
|------------|-------------|----------|----------|
| Mention manquante/erronée | Art. 1737 CGI | 15 € par mention | Plafond 25% montant facture |
| Absence de facturation | Art. L441-3 | 75 000 € | 150 000 € |
| Modification facture payée | Art. L441-9 | 75 000 € | 150 000 € |
| Numérotation non continue | Art. 242 nonies A | 15 € par facture | - |
| Délai paiement dépassé | Art. L441-6 | 75 000 € | 150 000 € |
| Non-conservation | Art. L123-22 | Rejet comptabilité | Redressement |

### Calcul Amende Mentions

```go
func CalculatePotentialFine(invoice *Invoice) decimal.Decimal {
    errors := ValidateInvoice(invoice)
    
    finePerError := decimal.NewFromInt(15)
    totalFine := finePerError.Mul(decimal.NewFromInt(int64(len(errors))))
    
    // Plafond : 25% du montant facture
    maxFine := invoice.TotalTTC.Mul(decimal.NewFromFloat(0.25))
    
    if totalFine.GreaterThan(maxFine) {
        return maxFine
    }
    
    return totalFine
}
```

---

## 📖 RÉFÉRENCES LÉGALES

### Textes Principaux

1. **Code de Commerce**
   - Article L441-3 : Obligation de facturation
   - Article L441-9 : Immuabilité et sanctions
   - Article L123-22 : Conservation

2. **Code Général des Impôts**
   - Article 293B : Franchise en base de TVA
   - Article 1737 : Amendes mentions
   - Article 242 nonies A annexe II : Numérotation

3. **Décrets et Ordonnances**
   - Décret n° 2022-1299 (7 oct 2022) : Nouvelles mentions B2B
   - Ordonnance n° 2021-1190 : Facturation électronique

### Sites Officiels

- Légifrance : https://www.legifrance.gouv.fr
- Economie.gouv.fr : Facturation électronique
- URSSAF : Obligations auto-entrepreneur
- Impots.gouv.fr : FAQ facturation

---

## ✅ CHECKLIST CONFORMITÉ

Avant chaque release, vérifier :

### Factures
- [ ] Immuabilité après paiement testée
- [ ] Numérotation continue vérifiée
- [ ] Toutes mentions obligatoires présentes
- [ ] Calculs décimaux corrects (decimal.Decimal)
- [ ] Validation SIREN clients B2B
- [ ] Génération PDF conforme
- [ ] Conservation 10 ans implémentée
- [ ] Audit trail fonctionnel

### Avoirs
- [ ] Référence facture obligatoire
- [ ] Numérotation séparée
- [ ] Montants négatifs corrects
- [ ] Raison documentée

### Devis
- [ ] Numérotation distincte factures
- [ ] Date d'expiration présente
- [ ] Conversion en facture tracée

### Système
- [ ] Backup automatique DB
- [ ] Logs audit accessibles
- [ ] Export comptable conforme
- [ ] Tests règles légales OK

---

**Dernière Mise à Jour** : 2026-03-04
**Prochaine Révision** : 2027-01-01 (nouvelles obligations e-invoicing)
