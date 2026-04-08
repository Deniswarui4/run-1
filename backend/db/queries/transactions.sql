-- name: GetTransactionByReference :one
SELECT id, user_id, event_id, type, status, amount, currency, platform_fee, net_amount,
       payment_gateway, payment_reference, payment_metadata, description, failure_reason,
       created_at, updated_at, deleted_at
FROM transactions
WHERE payment_reference = $1 AND deleted_at IS NULL;

-- name: CreateTransaction :one
INSERT INTO transactions (
    user_id, event_id, type, status, amount, currency, platform_fee, net_amount,
    payment_gateway, payment_reference, payment_metadata, description
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING id, user_id, event_id, type, status, amount, currency, platform_fee, net_amount,
          payment_gateway, payment_reference, payment_metadata, description, failure_reason,
          created_at, updated_at, deleted_at;

-- name: UpdateTransactionStatus :one
UPDATE transactions
SET
    status         = $2,
    failure_reason = COALESCE(sqlc.narg(failure_reason), failure_reason),
    updated_at     = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, user_id, event_id, type, status, amount, currency, platform_fee, net_amount,
          payment_gateway, payment_reference, payment_metadata, description, failure_reason,
          created_at, updated_at, deleted_at;
