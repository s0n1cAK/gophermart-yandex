CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext; 

CREATE TABLE users (
    id          BIGSERIAL PRIMARY KEY,
    login_name  CITEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE users_credentials (
    user_id         BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    password_hash   TEXT    NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_orders (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_number    VARCHAR(255) NOT NULL,
    status          TEXT NOT NULL DEFAULT 'NEW' CHECK (status IN ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED')),
    points_awarded  REAL NOT NULL DEFAULT 0 CHECK (points_awarded >= 0),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    processing_started_at TIMESTAMPTZ,
    UNIQUE (user_id, order_number)
);

CREATE INDEX idx_user_orders_user_time ON user_orders(user_id, created_at DESC);
CREATE INDEX idx_user_orders_user_status ON user_orders(user_id, status);

CREATE TYPE balance_entry_type AS ENUM ('accrual', 'withdrawal', 'adjustment');

CREATE TABLE user_balance_entries (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    entry_type      balance_entry_type NOT NULL,
    amount_points   REAL NOT NULL CHECK (amount_points >= 0),
    order_id        BIGINT REFERENCES user_orders(id),
    withdrawal_ref  TEXT UNIQUE, 
    posted_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT user_balance_entries_accrual_unique UNIQUE (order_id, entry_type)
);

CREATE INDEX idx_user_balance_entries_user_time
    ON user_balance_entries(user_id, posted_at DESC);

CREATE TABLE user_point_balances (
    user_id     BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    balance     REAL NOT NULL DEFAULT 0 CHECK (balance >= 0),
    withdrawal  REAL NOT NULL DEFAULT 0 CHECK (withdrawal >= 0),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

--https://github.com/softwareNuggets/PostgreSQL_shorts_resources/blob/main/password_salt_and_hash.sql
CREATE OR REPLACE FUNCTION public.sfn_hash_password(p_password text)
RETURNS text
LANGUAGE sql
AS $$
    SELECT crypt($1, gen_salt('bf', 12));
$$;