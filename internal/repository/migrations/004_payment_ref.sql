-- Migration 004: ajout du libellé de virement unique par facture
-- Le payment_ref est un code de 8 caractères alphanumériques (A-Z0-9) généré automatiquement
-- Il sert de libellé de virement lors des paiements par virement bancaire.
ALTER TABLE invoices ADD COLUMN payment_ref TEXT NOT NULL DEFAULT '';
