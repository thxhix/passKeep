CREATE TABLE IF NOT EXISTS keychain (
    id SERIAL PRIMARY KEY,
    key_uuid UUID UNIQUE NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(16) NOT NULL CHECK (type IN ('credential','text','file','card')),
    title VARCHAR(128) NOT NULL,
    data BYTEA NOT NULL,
    nonce BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    soft_deleted BOOL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_keychain_user_id ON keychain(user_id);