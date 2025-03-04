-- +goose Up
CREATE TABLE IF NOT EXISTS ticketBuyers (
    wallet_address CHAR(42) PRIMARY KEY,
    nonce SMALLINT UNIQUE,
    num_tickets SMALLINT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS ticketBuyers;