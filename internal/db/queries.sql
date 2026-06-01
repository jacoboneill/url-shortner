-- name: GetURL :one
SELECT url FROM urls WHERE token = ? LIMIT 1;

-- name: CreateURL :exec
INSERT INTO urls (token, url) VALUES (?, ?);

-- name: TokenExists :one
SELECT EXISTS(SELECT 1 FROM urls WHERE token = ?) AS found;

-- name: AddTimestamp :exec
INSERT INTO timestamps (token) VALUES (?);
