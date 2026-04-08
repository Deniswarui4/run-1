-- +goose Up

-- Guest sessions table
CREATE TABLE guest_sessions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token      VARCHAR NOT NULL UNIQUE,
    email      VARCHAR NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_guest_sessions_token ON guest_sessions (token);
CREATE INDEX idx_guest_sessions_expires_at ON guest_sessions (expires_at);

-- Add guest_email to tickets (null for authenticated purchases)
ALTER TABLE tickets ADD COLUMN guest_email VARCHAR;
ALTER TABLE tickets ALTER COLUMN attendee_id DROP NOT NULL;

-- Draft carts
CREATE TABLE draft_carts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID REFERENCES users(id),
    guest_session   VARCHAR,
    event_id        UUID NOT NULL REFERENCES events(id),
    status          VARCHAR(20) NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'expired', 'completed')),
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_draft_carts_user_id ON draft_carts (user_id);
CREATE INDEX idx_draft_carts_guest_session ON draft_carts (guest_session);
CREATE INDEX idx_draft_carts_expires_at ON draft_carts (expires_at);
-- Partial unique indexes for active carts only
CREATE UNIQUE INDEX uq_draft_cart_user_event ON draft_carts (user_id, event_id) 
    WHERE status = 'active' AND user_id IS NOT NULL;
CREATE UNIQUE INDEX uq_draft_cart_guest_event ON draft_carts (guest_session, event_id) 
    WHERE status = 'active' AND guest_session IS NOT NULL;

-- Draft cart items
CREATE TABLE draft_cart_items (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id        UUID NOT NULL REFERENCES draft_carts(id) ON DELETE CASCADE,
    ticket_type_id UUID NOT NULL REFERENCES ticket_types(id),
    quantity       INTEGER NOT NULL CHECK (quantity > 0),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (cart_id, ticket_type_id)
);
CREATE INDEX idx_draft_cart_items_cart_id ON draft_cart_items (cart_id);

-- Add featured_type to events
ALTER TABLE events ADD COLUMN featured_type VARCHAR(10) NOT NULL DEFAULT 'none'
    CHECK (featured_type IN ('none', 'manual', 'auto'));

-- Add auto_feature_threshold to platform_settings
ALTER TABLE platform_settings ADD COLUMN auto_feature_threshold INTEGER NOT NULL DEFAULT 0;

-- Add consecutive_below_threshold counter for auto-unfeature logic
CREATE TABLE event_feature_evaluations (
    event_id                    UUID PRIMARY KEY REFERENCES events(id),
    consecutive_below_threshold INTEGER NOT NULL DEFAULT 0,
    last_evaluated_at           TIMESTAMPTZ
);

-- +goose Down
DROP TABLE IF EXISTS event_feature_evaluations;
ALTER TABLE platform_settings DROP COLUMN IF EXISTS auto_feature_threshold;
ALTER TABLE events DROP COLUMN IF EXISTS featured_type;
DROP TABLE IF EXISTS draft_cart_items;
DROP TABLE IF EXISTS draft_carts;
ALTER TABLE tickets DROP COLUMN IF EXISTS guest_email;
ALTER TABLE tickets ALTER COLUMN attendee_id SET NOT NULL;
DROP TABLE IF EXISTS guest_sessions;
