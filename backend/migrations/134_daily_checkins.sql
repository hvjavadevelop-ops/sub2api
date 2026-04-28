CREATE TABLE IF NOT EXISTS user_daily_checkins (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    checkin_date DATE NOT NULL,
    reward DECIMAL(20,8) NOT NULL,
    balance_after DECIMAL(20,8) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uniq_user_daily_checkins_user_date UNIQUE (user_id, checkin_date)
);

CREATE INDEX IF NOT EXISTS idx_user_daily_checkins_user_created
    ON user_daily_checkins (user_id, created_at DESC);

COMMENT ON TABLE user_daily_checkins IS 'Daily sign-in reward records';
COMMENT ON COLUMN user_daily_checkins.reward IS 'Balance reward granted by daily sign-in';
