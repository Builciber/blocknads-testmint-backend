-- name: CreateTicketBuyer :exec
INSERT INTO ticketBuyers(wallet_address, nonce, num_tickets, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetRaffleMinter :one
SELECT * FROM ticketBuyers
WHERE wallet_address = $1;

-- name: AddRaffleMinterNonce :exec
UPDATE ticketBuyers SET wallet_address = $2
WHERE wallet_address = $1;