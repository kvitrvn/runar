-- Migration 006: ajout du libellé de virement unique pour les acomptes sur devis
-- Le deposit_payment_ref est généré uniquement si un acompte est demandé et que
-- les coordonnées bancaires de configuration sont valides.
ALTER TABLE quotes ADD COLUMN deposit_payment_ref TEXT NOT NULL DEFAULT '';
