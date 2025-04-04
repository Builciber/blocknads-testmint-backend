-- +goose Up
CREATE TABLE IF NOT EXISTS rafflingState(
    id SMALLSERIAL PRIMARY KEY,
    has_raffled BOOLEAN DEFAULT FALSE NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS rafflingState;