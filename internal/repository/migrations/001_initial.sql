-- Migration 001 : Tables principales
-- LEGAL: Conservation obligatoire 10 ans (Art. L123-22 Code de Commerce)

PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;

-- Table clients
CREATE TABLE IF NOT EXISTS clients (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL,
    siret       TEXT,               -- 14 chiffres pour entreprises
    siren       TEXT,               -- 9 chiffres (obligatoire B2B depuis 2024)
    vat_number  TEXT,               -- N° TVA intracommunautaire
    address     TEXT    NOT NULL,
    postal_code TEXT    NOT NULL,
    city        TEXT    NOT NULL,
    country     TEXT    NOT NULL DEFAULT 'France',
    email       TEXT,
    phone       TEXT,
    notes       TEXT,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Table factures
-- LEGAL: Numérotation continue sans trou (Art. 242 nonies A CGI)
-- LEGAL: Immuabilité après paiement (Art. L441-9 Code de Commerce)
CREATE TABLE IF NOT EXISTS invoices (
    id                   INTEGER PRIMARY KEY AUTOINCREMENT,
    number               TEXT    NOT NULL UNIQUE, -- Ex: "2026-0001"
    client_id            INTEGER NOT NULL REFERENCES clients(id),
    quote_id             INTEGER REFERENCES quotes(id), -- Si issu d'un devis

    -- Dates
    issue_date           DATE    NOT NULL,
    due_date             DATE    NOT NULL,
    delivery_date        DATE    NOT NULL,
    paid_date            DATE,
    paid_locked_at       DATETIME, -- Timestamp du verrouillage immuable

    -- État
    -- LEGAL: États : draft, issued, sent, paid (immuable), canceled (immuable)
    state                TEXT    NOT NULL DEFAULT 'draft'
                         CHECK(state IN ('draft','issued','sent','paid','canceled')),

    -- Montants (stockés en centimes sur TEXT pour éviter float)
    total_ht             TEXT    NOT NULL DEFAULT '0',
    total_ttc            TEXT    NOT NULL DEFAULT '0',
    vat_amount           TEXT    NOT NULL DEFAULT '0',

    -- TVA
    vat_applicable       INTEGER NOT NULL DEFAULT 0, -- 0=franchise, 1=assujetti
    -- LEGAL: Mention exacte obligatoire si franchise en base (Art. 293B CGI)
    vat_exemption_text   TEXT    NOT NULL DEFAULT 'TVA non applicable, article 293B du CGI',

    -- Paiement
    -- LEGAL: Délai max 60 jours nets ou 45 jours fin de mois
    payment_deadline     TEXT    NOT NULL DEFAULT '30 jours',
    -- LEGAL: Taux BCE + 10 points
    late_penalty_rate    TEXT    NOT NULL DEFAULT '0',
    -- LEGAL: Indemnité forfaitaire 40€ obligatoire
    recovery_fee         TEXT    NOT NULL DEFAULT '40',
    early_payment_disc   TEXT,   -- Escompte si paiement anticipé (optionnel)

    -- Mentions 2027
    operation_category   TEXT    CHECK(operation_category IN ('service','goods','mixed')),
    delivery_address     TEXT,   -- Si différente adresse client

    -- Métadonnées
    notes                TEXT,
    pdf_path             TEXT,   -- LEGAL: PDF jamais supprimé
    created_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Table lignes de facture
CREATE TABLE IF NOT EXISTS invoice_lines (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    invoice_id    INTEGER NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    line_order    INTEGER NOT NULL DEFAULT 0,
    description   TEXT    NOT NULL,
    quantity      TEXT    NOT NULL,
    unit_price_ht TEXT    NOT NULL,
    vat_rate      TEXT    NOT NULL DEFAULT '0',
    total_ht      TEXT    NOT NULL,
    total_ttc     TEXT    NOT NULL
);

-- Table devis
CREATE TABLE IF NOT EXISTS quotes (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    number      TEXT    NOT NULL UNIQUE, -- Ex: "DEV-2026-0001"
    client_id   INTEGER NOT NULL REFERENCES clients(id),
    issue_date  DATE    NOT NULL,
    expiry_date DATE    NOT NULL,
    state       TEXT    NOT NULL DEFAULT 'draft'
                CHECK(state IN ('draft','sent','accepted','refused','expired')),
    total_ht    TEXT    NOT NULL DEFAULT '0',
    total_ttc   TEXT    NOT NULL DEFAULT '0',
    vat_amount  TEXT    NOT NULL DEFAULT '0',
    notes       TEXT,
    pdf_path    TEXT,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Table lignes de devis
CREATE TABLE IF NOT EXISTS quote_lines (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    quote_id      INTEGER NOT NULL REFERENCES quotes(id) ON DELETE CASCADE,
    line_order    INTEGER NOT NULL DEFAULT 0,
    description   TEXT    NOT NULL,
    quantity      TEXT    NOT NULL,
    unit_price_ht TEXT    NOT NULL,
    vat_rate      TEXT    NOT NULL DEFAULT '0',
    total_ht      TEXT    NOT NULL,
    total_ttc     TEXT    NOT NULL
);

-- Table avoirs (credit notes)
-- LEGAL: Référence obligatoire à la facture d'origine (Art. 272 CGI)
CREATE TABLE IF NOT EXISTS credit_notes (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    number            TEXT    NOT NULL UNIQUE, -- Ex: "A-2026-0001"
    invoice_id        INTEGER NOT NULL REFERENCES invoices(id),
    invoice_reference TEXT    NOT NULL, -- Numéro et date de la facture d'origine
    issue_date        DATE    NOT NULL,
    reason            TEXT    NOT NULL,
    total_ht          TEXT    NOT NULL, -- Négatif
    total_ttc         TEXT    NOT NULL, -- Négatif
    vat_amount        TEXT    NOT NULL DEFAULT '0',
    pdf_path          TEXT,
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Table lignes d'avoir
CREATE TABLE IF NOT EXISTS credit_note_lines (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    credit_note_id  INTEGER NOT NULL REFERENCES credit_notes(id) ON DELETE CASCADE,
    line_order      INTEGER NOT NULL DEFAULT 0,
    description     TEXT    NOT NULL,
    quantity        TEXT    NOT NULL,
    unit_price_ht   TEXT    NOT NULL,
    vat_rate        TEXT    NOT NULL DEFAULT '0',
    total_ht        TEXT    NOT NULL,
    total_ttc       TEXT    NOT NULL
);

-- Index pour performances
CREATE INDEX IF NOT EXISTS idx_invoices_number   ON invoices(number);
CREATE INDEX IF NOT EXISTS idx_invoices_client   ON invoices(client_id);
CREATE INDEX IF NOT EXISTS idx_invoices_state    ON invoices(state);
CREATE INDEX IF NOT EXISTS idx_invoices_year     ON invoices(strftime('%Y', issue_date));
CREATE INDEX IF NOT EXISTS idx_quotes_number     ON quotes(number);
CREATE INDEX IF NOT EXISTS idx_quotes_client     ON quotes(client_id);
CREATE INDEX IF NOT EXISTS idx_clients_siret     ON clients(siret);
