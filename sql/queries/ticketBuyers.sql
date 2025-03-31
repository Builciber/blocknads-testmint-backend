-- name: CreateTicketBuyer :exec
INSERT INTO ticketBuyers(wallet_address, nonce, num_tickets, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: CreateFakeTicketBuyers :copyfrom
INSERT INTO ticketBuyers(wallet_address, nonce, num_tickets, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetTicketBuyer :one
SELECT * FROM ticketBuyers
WHERE wallet_address = $1;

-- name: AddRaffleMinterNonce :exec
UPDATE ticketBuyers SET nonce = $2, updated_at = $3
WHERE wallet_address = $1;

-- name: GetAllTicketBuyers :many
SELECT wallet_address, num_tickets FROM ticketBuyers
ORDER BY num_tickets ASC;

-- name: GetUniqueWeights :many
SELECT DISTINCT num_tickets FROM ticketBuyers
ORDER BY num_tickets ASC;

-- name: GetNumTickets :one
SELECT num_tickets FROM ticketBuyers
WHERE wallet_address = $1;

-- name: UpdateNumTickets :exec
UPDATE ticketBuyers SET num_tickets = $2, updated_at = $3
WHERE wallet_address = $1;

-- name: CreateRaffleWinnersForTx :copyfrom
INSERT INTO raffleWinners(wallet_address, nonce)
VALUES ($1, $2);

-- name: UpdateTicketBuyersNonceForTx :exec
UPDATE ticketBuyers SET nonce = raffleWinners.nonce, updated_at = $1
FROM raffleWinners
WHERE ticketBuyers.wallet_address = raffleWinners.wallet_address;

-- name: IsRaffleWinner :one
SELECT EXISTS (SELECT nonce FROM ticketBuyers WHERE wallet_address = $1 AND nonce IS NOT NULL);