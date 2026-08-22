-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

-- Organization — the optional/collapsible grouping layer (design §3.1). Intentionally thin.
CREATE TABLE IF NOT EXISTS organizations ();
ALTER TABLE IF EXISTS organizations ADD COLUMN IF NOT EXISTS id VARCHAR PRIMARY KEY;
ALTER TABLE IF EXISTS organizations ADD COLUMN IF NOT EXISTS name VARCHAR NOT NULL;
ALTER TABLE IF EXISTS organizations ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL CHECK (EXTRACT(TIMEZONE FROM updated_at) = 0);
ALTER TABLE IF EXISTS organizations ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL CHECK (EXTRACT(TIMEZONE FROM created_at) = 0);

-- Business — the aggregate root / legal entity (design §3.2). Holds legal identity,
-- lifecycle, and references-out only; never money or rail state (design §8.3).
CREATE TABLE IF NOT EXISTS businesses ();
ALTER TABLE IF EXISTS businesses ADD COLUMN IF NOT EXISTS id VARCHAR PRIMARY KEY;
ALTER TABLE IF EXISTS businesses ADD COLUMN IF NOT EXISTS correlation_id VARCHAR UNIQUE NOT NULL CHECK (correlation_id <> '');
ALTER TABLE IF EXISTS businesses ADD COLUMN IF NOT EXISTS legal_name VARCHAR NOT NULL;
ALTER TABLE IF EXISTS businesses ADD COLUMN IF NOT EXISTS trade_name VARCHAR NOT NULL;
ALTER TABLE IF EXISTS businesses ADD COLUMN IF NOT EXISTS cr_number VARCHAR NOT NULL;
ALTER TABLE IF EXISTS businesses ADD COLUMN IF NOT EXISTS legal_form VARCHAR NOT NULL;
ALTER TABLE IF EXISTS businesses ADD COLUMN IF NOT EXISTS incorporation_date timestamptz CHECK (EXTRACT(TIMEZONE FROM incorporation_date) = 0);
ALTER TABLE IF EXISTS businesses ADD COLUMN IF NOT EXISTS status VARCHAR NOT NULL;
ALTER TABLE IF EXISTS businesses ADD COLUMN IF NOT EXISTS organization_id VARCHAR REFERENCES organizations (id);
ALTER TABLE IF EXISTS businesses ADD COLUMN IF NOT EXISTS keycloak_org_id VARCHAR;
ALTER TABLE IF EXISTS businesses ADD COLUMN IF NOT EXISTS ledger_ref VARCHAR;
ALTER TABLE IF EXISTS businesses ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL CHECK (EXTRACT(TIMEZONE FROM updated_at) = 0);
ALTER TABLE IF EXISTS businesses ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL CHECK (EXTRACT(TIMEZONE FROM created_at) = 0);
CREATE INDEX IF NOT EXISTS businesses_status_idx ON businesses (status);
CREATE INDEX IF NOT EXISTS businesses_organization_id_idx ON businesses (organization_id);
CREATE INDEX IF NOT EXISTS businesses_cr_number_idx ON businesses (cr_number);

-- Classification — labeled (axis, value) records, not new types (design §3.3). The
-- closed vocabulary is enforced in core.ValidateClassification; one value per axis.
CREATE TABLE IF NOT EXISTS classifications ();
ALTER TABLE IF EXISTS classifications ADD COLUMN IF NOT EXISTS id VARCHAR PRIMARY KEY;
ALTER TABLE IF EXISTS classifications ADD COLUMN IF NOT EXISTS business_id VARCHAR NOT NULL REFERENCES businesses (id);
ALTER TABLE IF EXISTS classifications ADD COLUMN IF NOT EXISTS axis VARCHAR NOT NULL;
ALTER TABLE IF EXISTS classifications ADD COLUMN IF NOT EXISTS value VARCHAR NOT NULL;
ALTER TABLE IF EXISTS classifications ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL CHECK (EXTRACT(TIMEZONE FROM updated_at) = 0);
ALTER TABLE IF EXISTS classifications ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL CHECK (EXTRACT(TIMEZONE FROM created_at) = 0);
CREATE UNIQUE INDEX IF NOT EXISTS classifications_business_axis_uq ON classifications (business_id, axis);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP INDEX IF EXISTS classifications_business_axis_uq;
DROP TABLE IF EXISTS classifications;

DROP INDEX IF EXISTS businesses_cr_number_idx;
DROP INDEX IF EXISTS businesses_organization_id_idx;
DROP INDEX IF EXISTS businesses_status_idx;
DROP TABLE IF EXISTS businesses;

DROP TABLE IF EXISTS organizations;
-- +goose StatementEnd
