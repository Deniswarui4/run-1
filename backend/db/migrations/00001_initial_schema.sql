-- +goose Up
-- +goose StatementBegin

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. users
CREATE TABLE users (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email                 VARCHAR NOT NULL,
    password              VARCHAR NOT NULL,
    first_name            VARCHAR NOT NULL,
    last_name             VARCHAR NOT NULL,
    phone                 VARCHAR,
    role                  VARCHAR(20) NOT NULL DEFAULT 'attendee' CHECK (role IN ('admin', 'moderator', 'organizer', 'attendee')),
    is_active             BOOLEAN NOT NULL DEFAULT TRUE,
    is_verified           BOOLEAN NOT NULL DEFAULT FALSE,
    verification_token    VARCHAR,
    verification_expiry   TIMESTAMPTZ,
    password_reset_token  VARCHAR,
    password_reset_expiry TIMESTAMPTZ,
    two_factor_secret     VARCHAR,
    two_factor_enabled    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_users_email ON users (email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_verification_token ON users (verification_token);
CREATE INDEX idx_users_password_reset_token ON users (password_reset_token);
CREATE INDEX idx_users_deleted_at ON users (deleted_at);

-- 2. categories
CREATE TABLE categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR NOT NULL,
    description TEXT,
    color       VARCHAR NOT NULL DEFAULT '#3B82F6',
    icon        VARCHAR,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_categories_name ON categories (name);

-- 3. events (FK → users)
CREATE TABLE events (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title               VARCHAR NOT NULL,
    description         TEXT,
    category            VARCHAR,
    venue               VARCHAR NOT NULL,
    address             VARCHAR,
    city                VARCHAR,
    country             VARCHAR,
    image_url           VARCHAR,
    start_date          TIMESTAMPTZ NOT NULL,
    end_date            TIMESTAMPTZ NOT NULL,
    status              VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'pending', 'approved', 'rejected', 'published', 'cancelled', 'completed')),
    is_featured         BOOLEAN NOT NULL DEFAULT FALSE,
    organizer_id        UUID NOT NULL REFERENCES users(id),
    moderator_id        UUID REFERENCES users(id),
    moderation_comment  TEXT,
    moderated_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX idx_events_organizer_id ON events (organizer_id);
CREATE INDEX idx_events_moderator_id ON events (moderator_id);
CREATE INDEX idx_events_deleted_at ON events (deleted_at);

-- 4. ticket_types (FK → events)
CREATE TABLE ticket_types (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id      UUID NOT NULL REFERENCES events(id),
    name          VARCHAR NOT NULL,
    description   TEXT,
    price         NUMERIC NOT NULL,
    quantity      INTEGER NOT NULL,
    sold          INTEGER NOT NULL DEFAULT 0,
    max_per_order INTEGER NOT NULL DEFAULT 10,
    sale_start    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sale_end      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ticket_types_event_id ON ticket_types (event_id);

-- 5. tickets (FK → events, ticket_types, users) — transaction_id FK added later
CREATE TABLE tickets (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_number  VARCHAR NOT NULL,
    event_id       UUID NOT NULL REFERENCES events(id),
    ticket_type_id UUID NOT NULL REFERENCES ticket_types(id),
    attendee_id    UUID NOT NULL REFERENCES users(id),
    transaction_id UUID NOT NULL,
    status         VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'cancelled', 'used')),
    price          NUMERIC NOT NULL,
    qr_code_url    VARCHAR,
    pdf_url        VARCHAR,
    checked_in_at  TIMESTAMPTZ,
    checked_in_by  UUID REFERENCES users(id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_tickets_ticket_number ON tickets (ticket_number);
CREATE INDEX idx_tickets_event_id ON tickets (event_id);
CREATE INDEX idx_tickets_ticket_type_id ON tickets (ticket_type_id);
CREATE INDEX idx_tickets_attendee_id ON tickets (attendee_id);
CREATE INDEX idx_tickets_transaction_id ON tickets (transaction_id);
CREATE INDEX idx_tickets_deleted_at ON tickets (deleted_at);

-- 6. transactions (FK → users, events)
CREATE TABLE transactions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id),
    event_id          UUID REFERENCES events(id),
    type              VARCHAR(30) NOT NULL CHECK (type IN ('ticket_purchase', 'refund', 'withdrawal')),
    status            VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed', 'refunded')),
    amount            NUMERIC NOT NULL,
    currency          VARCHAR NOT NULL DEFAULT 'KES',
    platform_fee      NUMERIC NOT NULL DEFAULT 0,
    net_amount        NUMERIC NOT NULL,
    payment_gateway   VARCHAR,
    payment_reference VARCHAR,
    payment_metadata  JSONB,
    description       VARCHAR,
    failure_reason    VARCHAR,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_transactions_payment_reference ON transactions (payment_reference) WHERE payment_reference IS NOT NULL;
CREATE INDEX idx_transactions_user_id ON transactions (user_id);
CREATE INDEX idx_transactions_event_id ON transactions (event_id);
CREATE INDEX idx_transactions_deleted_at ON transactions (deleted_at);

-- 7. platform_settings
CREATE TABLE platform_settings (
    id                        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform_fee_percentage   NUMERIC NOT NULL DEFAULT 5.0,
    withdrawal_fee_percentage NUMERIC NOT NULL DEFAULT 2.5,
    min_withdrawal_amount     NUMERIC NOT NULL DEFAULT 1000,
    currency                  VARCHAR NOT NULL DEFAULT 'KES',
    updated_by                UUID,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 8. withdrawal_requests (FK → users)
CREATE TABLE withdrawal_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organizer_id    UUID NOT NULL REFERENCES users(id),
    amount          NUMERIC NOT NULL,
    withdrawal_fee  NUMERIC NOT NULL,
    net_amount      NUMERIC NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'processed')),
    bank_name       VARCHAR NOT NULL,
    account_number  VARCHAR NOT NULL,
    account_name    VARCHAR NOT NULL,
    reviewed_by     UUID REFERENCES users(id),
    reviewed_at     TIMESTAMPTZ,
    review_comment  TEXT,
    processed_at    TIMESTAMPTZ,
    transaction_ref VARCHAR,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_withdrawal_requests_organizer_id ON withdrawal_requests (organizer_id);
CREATE INDEX idx_withdrawal_requests_deleted_at ON withdrawal_requests (deleted_at);

-- 9. organizer_balances (FK → users)
CREATE TABLE organizer_balances (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organizer_id      UUID NOT NULL REFERENCES users(id),
    total_earnings    NUMERIC NOT NULL DEFAULT 0,
    available_balance NUMERIC NOT NULL DEFAULT 0,
    pending_balance   NUMERIC NOT NULL DEFAULT 0,
    withdrawn_amount  NUMERIC NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_organizer_balances_organizer_id ON organizer_balances (organizer_id);

-- Resolve circular FK: tickets.transaction_id → transactions
ALTER TABLE tickets
    ADD CONSTRAINT fk_tickets_transaction
    FOREIGN KEY (transaction_id) REFERENCES transactions(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop circular FK first
ALTER TABLE tickets DROP CONSTRAINT fk_tickets_transaction;

-- Drop tables in reverse FK dependency order
DROP TABLE IF EXISTS organizer_balances;
DROP TABLE IF EXISTS withdrawal_requests;
DROP TABLE IF EXISTS platform_settings;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS ticket_types;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "uuid-ossp";

-- +goose StatementEnd
