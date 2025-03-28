-- +goose Up
CREATE TABLE IF NOT EXISTS raffleWinners(
    wallet_address CHAR(42) PRIMARY KEY,
    nonce SMALLINT UNIQUE NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS ticketBuyers;