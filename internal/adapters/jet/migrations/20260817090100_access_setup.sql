-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

-- Actor — the initiation abstraction (design §4.1). Thin: an id + a discriminator of
-- which concrete form (Membership | ServiceAccount) backs it. Downstream references
-- actor_id, never a user_id or service_account_id.
CREATE TABLE IF NOT EXISTS actors ();
ALTER TABLE IF EXISTS actors ADD COLUMN IF NOT EXISTS id VARCHAR PRIMARY KEY;
ALTER TABLE IF EXISTS actors ADD COLUMN IF NOT EXISTS type VARCHAR NOT NULL;
ALTER TABLE IF EXISTS actors ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL CHECK (EXTRACT(TIMEZONE FROM updated_at) = 0);
ALTER TABLE IF EXISTS actors ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL CHECK (EXTRACT(TIMEZONE FROM created_at) = 0);

-- Membership — the human operator grant (design §4.2). person_id is the nullable
-- reconciliation seam: NO foreign key, because the Person table is deferred (design §5).
CREATE TABLE IF NOT EXISTS memberships ();
ALTER TABLE IF EXISTS memberships ADD COLUMN IF NOT EXISTS id VARCHAR PRIMARY KEY;
ALTER TABLE IF EXISTS memberships ADD COLUMN IF NOT EXISTS correlation_id VARCHAR UNIQUE NOT NULL CHECK (correlation_id <> '');
ALTER TABLE IF EXISTS memberships ADD COLUMN IF NOT EXISTS actor_id VARCHAR NOT NULL REFERENCES actors (id);
ALTER TABLE IF EXISTS memberships ADD COLUMN IF NOT EXISTS business_id VARCHAR NOT NULL REFERENCES businesses (id);
ALTER TABLE IF EXISTS memberships ADD COLUMN IF NOT EXISTS keycloak_sub VARCHAR NOT NULL;
ALTER TABLE IF EXISTS memberships ADD COLUMN IF NOT EXISTS person_id VARCHAR;
ALTER TABLE IF EXISTS memberships ADD COLUMN IF NOT EXISTS role VARCHAR NOT NULL;
ALTER TABLE IF EXISTS memberships ADD COLUMN IF NOT EXISTS status VARCHAR NOT NULL;
ALTER TABLE IF EXISTS memberships ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL CHECK (EXTRACT(TIMEZONE FROM updated_at) = 0);
ALTER TABLE IF EXISTS memberships ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL CHECK (EXTRACT(TIMEZONE FROM created_at) = 0);
-- A human operates a given business via at most one membership (design §6.1).
CREATE UNIQUE INDEX IF NOT EXISTS memberships_business_sub_uq ON memberships (business_id, keycloak_sub);
CREATE INDEX IF NOT EXISTS memberships_business_id_idx ON memberships (business_id);
CREATE INDEX IF NOT EXISTS memberships_status_idx ON memberships (status);

-- Invitation — the state before a User exists (design §4.4).
CREATE TABLE IF NOT EXISTS invitations ();
ALTER TABLE IF EXISTS invitations ADD COLUMN IF NOT EXISTS id VARCHAR PRIMARY KEY;
ALTER TABLE IF EXISTS invitations ADD COLUMN IF NOT EXISTS correlation_id VARCHAR UNIQUE NOT NULL CHECK (correlation_id <> '');
ALTER TABLE IF EXISTS invitations ADD COLUMN IF NOT EXISTS business_id VARCHAR NOT NULL REFERENCES businesses (id);
ALTER TABLE IF EXISTS invitations ADD COLUMN IF NOT EXISTS invited_email VARCHAR NOT NULL;
ALTER TABLE IF EXISTS invitations ADD COLUMN IF NOT EXISTS intended_role VARCHAR NOT NULL;
ALTER TABLE IF EXISTS invitations ADD COLUMN IF NOT EXISTS token VARCHAR UNIQUE NOT NULL;
ALTER TABLE IF EXISTS invitations ADD COLUMN IF NOT EXISTS status VARCHAR NOT NULL;
ALTER TABLE IF EXISTS invitations ADD COLUMN IF NOT EXISTS invited_by VARCHAR NOT NULL;
ALTER TABLE IF EXISTS invitations ADD COLUMN IF NOT EXISTS expires_at timestamptz NOT NULL CHECK (EXTRACT(TIMEZONE FROM expires_at) = 0);
ALTER TABLE IF EXISTS invitations ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL CHECK (EXTRACT(TIMEZONE FROM updated_at) = 0);
ALTER TABLE IF EXISTS invitations ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL CHECK (EXTRACT(TIMEZONE FROM created_at) = 0);
CREATE INDEX IF NOT EXISTS invitations_business_id_idx ON invitations (business_id);
CREATE INDEX IF NOT EXISTS invitations_status_idx ON invitations (status);

-- ServiceAccount — the machine operator (design §4.5), topology 1: scoped to one Business.
CREATE TABLE IF NOT EXISTS service_accounts ();
ALTER TABLE IF EXISTS service_accounts ADD COLUMN IF NOT EXISTS id VARCHAR PRIMARY KEY;
ALTER TABLE IF EXISTS service_accounts ADD COLUMN IF NOT EXISTS correlation_id VARCHAR UNIQUE NOT NULL CHECK (correlation_id <> '');
ALTER TABLE IF EXISTS service_accounts ADD COLUMN IF NOT EXISTS actor_id VARCHAR NOT NULL REFERENCES actors (id);
ALTER TABLE IF EXISTS service_accounts ADD COLUMN IF NOT EXISTS business_id VARCHAR NOT NULL REFERENCES businesses (id);
ALTER TABLE IF EXISTS service_accounts ADD COLUMN IF NOT EXISTS label VARCHAR NOT NULL;
ALTER TABLE IF EXISTS service_accounts ADD COLUMN IF NOT EXISTS status VARCHAR NOT NULL;
ALTER TABLE IF EXISTS service_accounts ADD COLUMN IF NOT EXISTS credential_ref VARCHAR;
ALTER TABLE IF EXISTS service_accounts ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL CHECK (EXTRACT(TIMEZONE FROM updated_at) = 0);
ALTER TABLE IF EXISTS service_accounts ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL CHECK (EXTRACT(TIMEZONE FROM created_at) = 0);
CREATE INDEX IF NOT EXISTS service_accounts_business_id_idx ON service_accounts (business_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP INDEX IF EXISTS service_accounts_business_id_idx;
DROP TABLE IF EXISTS service_accounts;

DROP INDEX IF EXISTS invitations_status_idx;
DROP INDEX IF EXISTS invitations_business_id_idx;
DROP TABLE IF EXISTS invitations;

DROP INDEX IF EXISTS memberships_status_idx;
DROP INDEX IF EXISTS memberships_business_id_idx;
DROP INDEX IF EXISTS memberships_business_sub_uq;
DROP TABLE IF EXISTS memberships;

DROP TABLE IF EXISTS actors;
-- +goose StatementEnd
