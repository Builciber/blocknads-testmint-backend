// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: whitelistMinters.sql

package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const addWhitelistMintWallet = `-- name: AddWhitelistMintWallet :exec
UPDATE whitelistMinters SET wallet_address = $2
WHERE discord_id = $1
`

type AddWhitelistMintWalletParams struct {
	DiscordID     pgtype.Text
	WalletAddress pgtype.Text
}

func (q *Queries) AddWhitelistMintWallet(ctx context.Context, arg AddWhitelistMintWalletParams) error {
	_, err := q.db.Exec(ctx, addWhitelistMintWallet, arg.DiscordID, arg.WalletAddress)
	return err
}

const canMint = `-- name: CanMint :one
SELECT EXISTS (SELECT 1 FROM whitelistMinters WHERE whitelistMinters.discord_id = $1 AND (whitelistMinters.wallet_address IS NULL OR whitelistMinters.wallet_address = $2))
`

type CanMintParams struct {
	DiscordID     pgtype.Text
	WalletAddress pgtype.Text
}

func (q *Queries) CanMint(ctx context.Context, arg CanMintParams) (bool, error) {
	row := q.db.QueryRow(ctx, canMint, arg.DiscordID, arg.WalletAddress)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

type CreateNoncesParams struct {
	ID        pgtype.UUID
	Nonce     int16
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
}

const getWhitelistMinterById = `-- name: GetWhitelistMinterById :one
SELECT id, discord_id, discord_username, wallet_address, avatar_hash, nonce, nonce_used, created_at, updated_at FROM whitelistMinters
WHERE discord_id = $1
`

func (q *Queries) GetWhitelistMinterById(ctx context.Context, discordID pgtype.Text) (Whitelistminter, error) {
	row := q.db.QueryRow(ctx, getWhitelistMinterById, discordID)
	var i Whitelistminter
	err := row.Scan(
		&i.ID,
		&i.DiscordID,
		&i.DiscordUsername,
		&i.WalletAddress,
		&i.AvatarHash,
		&i.Nonce,
		&i.NonceUsed,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const isExistingUser = `-- name: IsExistingUser :one
SELECT EXISTS (SELECT 1 FROM whitelistMinters WHERE whitelistMinters.discord_id = $1)
`

func (q *Queries) IsExistingUser(ctx context.Context, discordID pgtype.Text) (bool, error) {
	row := q.db.QueryRow(ctx, isExistingUser, discordID)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const isNonceColumnFilled = `-- name: IsNonceColumnFilled :one
SELECT EXISTS (SELECT 1 FROM whitelistMinters WHERE whitelistMinters.nonce IS NOT NULL LIMIT 1)
`

func (q *Queries) IsNonceColumnFilled(ctx context.Context) (bool, error) {
	row := q.db.QueryRow(ctx, isNonceColumnFilled)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const updateWhitelistMinterAfterAuth = `-- name: UpdateWhitelistMinterAfterAuth :exec
UPDATE whitelistMinters SET discord_id = $1, discord_username = $2, avatar_hash = $3, updated_at = $4
WHERE nonce = (SELECT nonce FROM whitelistMinters WHERE whitelistMinters.discord_id IS NULL LIMIT 1)
`

type UpdateWhitelistMinterAfterAuthParams struct {
	DiscordID       pgtype.Text
	DiscordUsername pgtype.Text
	AvatarHash      pgtype.Text
	UpdatedAt       pgtype.Timestamp
}

func (q *Queries) UpdateWhitelistMinterAfterAuth(ctx context.Context, arg UpdateWhitelistMinterAfterAuthParams) error {
	_, err := q.db.Exec(ctx, updateWhitelistMinterAfterAuth,
		arg.DiscordID,
		arg.DiscordUsername,
		arg.AvatarHash,
		arg.UpdatedAt,
	)
	return err
}
