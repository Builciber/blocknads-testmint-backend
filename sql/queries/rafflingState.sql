-- name: GetRaffleState :one
SELECT has_raffled FROM rafflingState;

-- name: UpdateRaffleState :exec
UPDATE rafflingState SET has_raffled = $1;