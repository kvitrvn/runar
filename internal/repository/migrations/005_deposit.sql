-- Migration 005: Acompte sur devis
-- Permet de demander un acompte (dépôt) sur un devis.
-- LEGAL: L'acompte apparaît sur le devis et est déduit en ligne négative sur la facture finale.
ALTER TABLE quotes ADD COLUMN deposit_rate DECIMAL(5,2) NOT NULL DEFAULT 0;
ALTER TABLE quotes ADD COLUMN deposit_paid INTEGER NOT NULL DEFAULT 0;
ALTER TABLE quotes ADD COLUMN deposit_paid_at TEXT;
