-- Phase 5: authentication + parent -> bus link
BEGIN;

ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT;

CREATE TABLE IF NOT EXISTS parent_bus_subscriptions (
    parent_id   BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bus_id      BIGINT NOT NULL REFERENCES buses(id) ON DELETE CASCADE,
    stop_id     BIGINT REFERENCES bus_stops(id) ON DELETE SET NULL,  -- optional: the child's pickup stop
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (parent_id, bus_id)
);
CREATE INDEX IF NOT EXISTS pbs_bus_idx ON parent_bus_subscriptions (bus_id);

COMMIT;