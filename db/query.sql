-- name: InsertSnippet :one
INSERT INTO 
snippets(title,content,created,expires)
VALUES($1,$2,now(),now() + $3 *INTERVAL '1 day') 
RETURNING id;

-- name: GetSnippet :one
SELECT * FROM 
snippets
WHERE expires > now() and id=$1;

-- name: GetLatestSnippets :many
SELECT * FROM
snippets
WHERE expires > now()
ORDER BY id DESC LIMIT 10;

-- name: InsertUser :exec
INSERT INTO
users(name,email,hashed_password,created)
VALUES($1,$2,$3,$4);

-- name: GetUserByEmail :one
SELECT * FROM 
users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM
users
WHERE id = $1;