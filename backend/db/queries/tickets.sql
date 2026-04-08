-- name: GetTicketByID :one
SELECT id, ticket_number, event_id, ticket_type_id, attendee_id, transaction_id,
       status, price, qr_code_url, pdf_url, checked_in_at, checked_in_by,
       created_at, updated_at, deleted_at
FROM tickets
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetTicketsByEvent :many
SELECT id, ticket_number, event_id, ticket_type_id, attendee_id, transaction_id,
       status, price, qr_code_url, pdf_url, checked_in_at, checked_in_by,
       created_at, updated_at, deleted_at
FROM tickets
WHERE event_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CreateTicket :one
INSERT INTO tickets (
    ticket_number, event_id, ticket_type_id, attendee_id, transaction_id,
    status, price, qr_code_url, pdf_url
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING id, ticket_number, event_id, ticket_type_id, attendee_id, transaction_id,
          status, price, qr_code_url, pdf_url, checked_in_at, checked_in_by,
          created_at, updated_at, deleted_at;
