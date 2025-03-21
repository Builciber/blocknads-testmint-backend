// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: ticketBuyers.sql

package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const addRaffleMinterNonce = `-- name: AddRaffleMinterNonce :exec
UPDATE ticketBuyers SET wallet_address = $2
WHERE wallet_address = $1
`

type AddRaffleMinterNonceParams struct {
	WalletAddress   string
	WalletAddress_2 string
}

func (q *Queries) AddRaffleMinterNonce(ctx context.Context, arg AddRaffleMinterNonceParams) error {
	_, err := q.db.Exec(ctx, addRaffleMinterNonce, arg.WalletAddress, arg.WalletAddress_2)
	return err
}

const createTicketBuyer = `-- name: CreateTicketBuyer :exec
INSERT INTO ticketBuyers(wallet_address, nonce, num_tickets, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
`

type CreateTicketBuyerParams struct {
	WalletAddress string
	Nonce         pgtype.Int2
	NumTickets    int16
	CreatedAt     pgtype.Timestamp
	UpdatedAt     pgtype.Timestamp
}

func (q *Queries) CreateTicketBuyer(ctx context.Context, arg CreateTicketBuyerParams) error {
	_, err := q.db.Exec(ctx, createTicketBuyer,
		arg.WalletAddress,
		arg.Nonce,
		arg.NumTickets,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	return err
}

const getRaffleMinter = `-- name: GetRaffleMinter :one
SELECT wallet_address, nonce, num_tickets, created_at, updated_at FROM ticketBuyers
WHERE wallet_address = $1
`

func (q *Queries) GetRaffleMinter(ctx context.Context, walletAddress string) (Ticketbuyer, error) {
	row := q.db.QueryRow(ctx, getRaffleMinter, walletAddress)
	var i Ticketbuyer
	err := row.Scan(
		&i.WalletAddress,
		&i.Nonce,
		&i.NumTickets,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
