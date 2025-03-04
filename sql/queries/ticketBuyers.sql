-- name: CreateTicketBuyer :exec
INSERT INTO ticketBuyers(wallet_address, nonce, num_tickets, created_at)
VALUES ($1, $2, $3, $4);

-- name: GetRaffleMinter :one
SELECT * FROM ticketBuyers
WHERE wallet_address = $1;