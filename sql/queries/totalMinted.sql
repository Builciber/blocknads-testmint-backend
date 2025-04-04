-- name: UpdateTotalNftsMinted :exec
UPDATE totalMinted SET total_nfts_minted = $1;

-- name: GetTotalMinted :one
SELECT total_nfts_minted FROM totalMinted;