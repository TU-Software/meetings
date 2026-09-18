-- name: ListRooms :many
SELECT id, name
FROM rooms
ORDER BY name;

-- name: GetRoom :one
SELECT id, name
FROM rooms
WHERE id = @id;

-- name: CreateRoom :one
INSERT INTO rooms (name)
VALUES (@name)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING id, name;

-- name: CreateMessage :one
INSERT INTO messages (room_id, sender, content)
VALUES (@room_id, @sender, @content)
RETURNING id, room_id, sender, content;

-- name: GetRecentMessages :many
SELECT id, room_id, sender, content
FROM messages
WHERE room_id = @room_id
ORDER BY id DESC
LIMIT 20;
