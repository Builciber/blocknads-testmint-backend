-- +goose Up
CREATE TABLE IF NOT EXISTS totalMinted(
    id SMALLSERIAL PRIMARY KEY,
    total_nfts_minted SMALLINT UNIQUE NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS totalMinted;