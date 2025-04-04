// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package database

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Rafflewinner struct {
	WalletAddress string
	Nonce         int16
}

type Rafflingstate struct {
	ID         int16
	HasRaffled bool
}

type Ticketbuyer struct {
	WalletAddress string
	Nonce         pgtype.Int2
	NumTickets    int16
	CreatedAt     pgtype.Timestamp
	UpdatedAt     pgtype.Timestamp
}

type Totalminted struct {
	ID              int16
	TotalNftsMinted int16
}

type Whitelistminter struct {
	ID              pgtype.UUID
	DiscordID       pgtype.Text
	DiscordUsername pgtype.Text
	WalletAddress   pgtype.Text
	AvatarHash      pgtype.Text
	Nonce           int16
	CreatedAt       pgtype.Timestamp
	UpdatedAt       pgtype.Timestamp
}
