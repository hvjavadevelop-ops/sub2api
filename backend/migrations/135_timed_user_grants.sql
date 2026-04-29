CREATE TABLE IF NOT EXISTS timed_user_grants (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    grant_type VARCHAR(32) NOT NULL CHECK (grant_type IN ('balance', 'concurrency')),
    amount DECIMAL(20,8) NOT NULL CHECK (amount > 0),
    duration_seconds INTEGER NOT NULL CHECK (duration_seconds > 0),
    status VARCHAR(32) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'expired', 'cancelled')),
    activated_at TIMESTAMPTZ NULL,
    expires_at TIMESTAMPTZ NULL,
    expired_at TIMESTAMPTZ NULL,
    deducted_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    created_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_timed_user_grants_user_status ON timed_user_grants(user_id, status);
CREATE INDEX IF NOT EXISTS idx_timed_user_grants_active_expires ON timed_user_grants(expires_at) WHERE status = 'active';
