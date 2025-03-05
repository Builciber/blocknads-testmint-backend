-- +goose Up
CREATE TABLE IF NOT EXISTS whitelistMinters (
    discord_id VARCHAR(64) PRIMARY KEY,
    discord_username VARCHAR(64),
    wallet_address CHAR(42) UNIQUE,
    avatar_hash char(64) UNIQUE,
    nonce SMALLINT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS whitelistMinters;