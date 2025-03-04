-- name: CreateWhitelistMinter :exec
INSERT INTO whitelistMinters(discord_id, wallet_address, nonce, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetWhitelistMinterById :one
SELECT * FROM whitelistMinters
WHERE discord_id = $1;