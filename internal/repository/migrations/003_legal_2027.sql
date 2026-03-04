-- Migration 003 : Préparation obligations légales 2027
-- LEGAL: Facturation électronique obligatoire dès le 1er septembre 2027
-- LEGAL: Nouvelles mentions obligatoires à partir du 1er juillet 2027

-- Champs déjà présents dans 001 (opération category, delivery_address)
-- Cette migration ajoute les champs spécifiques e-facturation

ALTER TABLE invoices ADD COLUMN IF NOT EXISTS e_invoice_format   TEXT;   -- "factur-x", "ubl", etc.
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS e_invoice_xml      TEXT;   -- Données XML structurées EN 16931
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS e_invoice_sent_at  DATETIME; -- Date envoi sur PDP/PPF
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS e_invoice_pdp_ref  TEXT;   -- Référence PDP partenaire
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS vat_payment_option TEXT    -- "debit" ou "encaissement"
                     CHECK(vat_payment_option IS NULL OR vat_payment_option IN ('debit','encaissement'));

-- Table séquences de numérotation pour garantir la continuité
-- LEGAL: Numérotation continue sans trou (Art. 242 nonies A CGI)
CREATE TABLE IF NOT EXISTS invoice_sequences (
    year        INTEGER PRIMARY KEY,
    last_seq    INTEGER NOT NULL DEFAULT 0,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS quote_sequences (
    year        INTEGER PRIMARY KEY,
    last_seq    INTEGER NOT NULL DEFAULT 0,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS credit_note_sequences (
    year        INTEGER PRIMARY KEY,
    last_seq    INTEGER NOT NULL DEFAULT 0,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
