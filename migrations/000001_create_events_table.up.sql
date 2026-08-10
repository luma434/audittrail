CREATE TABLE events (
    id         BIGSERIAL PRIMARY KEY,
    payload    JSONB NOT NULL,
    "timestamp" TIMESTAMPTZ NOT NULL,
    prev_hash  TEXT NOT NULL,
    hash       TEXT NOT NULL UNIQUE
);

CREATE INDEX events_timestamp_idx ON events ("timestamp");

CREATE FUNCTION prevent_events_mutation() RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'events table is append-only: % not allowed', TG_OP;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER events_no_update
    BEFORE UPDATE ON events
    FOR EACH ROW EXECUTE FUNCTION prevent_events_mutation();

CREATE TRIGGER events_no_delete
    BEFORE DELETE ON events
    FOR EACH ROW EXECUTE FUNCTION prevent_events_mutation();
