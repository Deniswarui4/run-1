-- name: GetUserByID :one
SELECT id, email, password, first_name, last_name, phone, role, is_active, is_verified,
       verification_token, verification_expiry, password_reset_token, password_reset_expiry,
       two_factor_secret, two_factor_enabled, created_at, updated_at, deleted_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT id, email, password, first_name, last_name, phone, role, is_active, is_verified,
       verification_token, verification_expiry, password_reset_token, password_reset_expiry,
       two_factor_secret, two_factor_enabled, created_at, updated_at, deleted_at
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: CreateUser :one
INSERT INTO users (
    email, password, first_name, last_name, phone, role,
    is_active, is_verified, two_factor_enabled
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING id, email, password, first_name, last_name, phone, role, is_active, is_verified,
          verification_token, verification_expiry, password_reset_token, password_reset_expiry,
          two_factor_secret, two_factor_enabled, created_at, updated_at, deleted_at;

-- name: UpdateUser :one
UPDATE users
SET
    first_name           = COALESCE(sqlc.narg(first_name), first_name),
    last_name            = COALESCE(sqlc.narg(last_name), last_name),
    phone                = COALESCE(sqlc.narg(phone), phone),
    is_active            = COALESCE(sqlc.narg(is_active), is_active),
    is_verified          = COALESCE(sqlc.narg(is_verified), is_verified),
    verification_token   = COALESCE(sqlc.narg(verification_token), verification_token),
    verification_expiry  = COALESCE(sqlc.narg(verification_expiry), verification_expiry),
    password_reset_token = COALESCE(sqlc.narg(password_reset_token), password_reset_token),
    password_reset_expiry = COALESCE(sqlc.narg(password_reset_expiry), password_reset_expiry),
    two_factor_secret    = COALESCE(sqlc.narg(two_factor_secret), two_factor_secret),
    two_factor_enabled   = COALESCE(sqlc.narg(two_factor_enabled), two_factor_enabled),
    updated_at           = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, email, password, first_name, last_name, phone, role, is_active, is_verified,
          verification_token, verification_expiry, password_reset_token, password_reset_expiry,
          two_factor_secret, two_factor_enabled, created_at, updated_at, deleted_at;
