-- name: CreateNonces :copyfrom
INSERT INTO whitelistMinters(id, nonce, created_at, updated_at)
VALUES ($1, $2, $3, $4);

-- name: GetWhitelistMinterById :one
SELECT * FROM whitelistMinters
WHERE discord_id = $1;

-- name: AddWhitelistMintWallet :exec
UPDATE whitelistMinters SET wallet_address = $2
WHERE discord_id = $1;

-- name: UpdateWhitelistMinterAfterAuth :exec
UPDATE whitelistMinters SET discord_id = $1, discord_username = $2, avatar_hash = $3, updated_at = $4
WHERE nonce = (SELECT nonce FROM whitelistMinters WHERE whitelistMinters.discord_id IS NULL LIMIT 1);

-- name: IsNonceColumnFilled :one
SELECT EXISTS (SELECT 1 FROM whitelistMinters WHERE whitelistMinters.nonce IS NOT NULL LIMIT 1);

-- name: CanMint :one
SELECT EXISTS (SELECT 1 FROM whitelistMinters WHERE whitelistMinters.discord_id = $1 AND (whitelistMinters.wallet_address IS NULL OR whitelistMinters.wallet_address = $2));

-- name: IsExistingUser :one
SELECT EXISTS (SELECT 1 FROM whitelistMinters WHERE whitelistMinters.discord_id = $1);