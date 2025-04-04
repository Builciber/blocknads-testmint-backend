-- +goose Up
CREATE TABLE IF NOT EXISTS totalMinted(
    total_nfts_minted SMALLINT UNIQUE NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS totalMinted;