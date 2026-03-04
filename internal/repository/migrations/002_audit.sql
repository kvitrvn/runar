-- Migration 002 : Table d'audit
-- LEGAL: Traçabilité obligatoire (Art. L47 A Livre des procédures fiscales)
-- LEGAL: Conservation 6 ans minimum (durée contrôle fiscal)

CREATE TABLE IF NOT EXISTS audit_log (
    id          INTEGER  PRIMARY KEY AUTOINCREMENT,
    entity_type TEXT     NOT NULL, -- "invoice", "quote", "client", "credit_note"
    entity_id   INTEGER  NOT NULL,
    action      TEXT     NOT NULL, -- "created", "updated", "paid", "pdf_generated", etc.
    user_id     TEXT     NOT NULL DEFAULT 'owner', -- Auto-entrepreneur = toujours owner
    old_value   TEXT,              -- JSON de l'état avant (NULL si création)
    new_value   TEXT,              -- JSON de l'état après (NULL si suppression)
    ip_address  TEXT,              -- Optionnel
    timestamp   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index pour recherche rapide dans l'audit
CREATE INDEX IF NOT EXISTS idx_audit_entity  ON audit_log(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_time    ON audit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_action  ON audit_log(action);
