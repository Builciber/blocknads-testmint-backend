-- name: CreateWhitelistMinter :exec
INSERT INTO whitelistMinters(discord_id, discord_username, wallet_address, avatar_hash, nonce, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetWhitelistMinterById :one
SELECT * FROM whitelistMinters
WHERE discord_id = $1;

-- name: UpdateWhitelistMinterAfterAuth :exec
UPDATE whitelistMinters SET discord_username = $2, avatar_hash = $3, updated_at = $4
WHERE discord_id = $1;

-- name: AddWhitelistMintWallet :exec
UPDATE whitelistMinters SET wallet_address = $2
WHERE discord_id = $1;