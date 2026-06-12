-- name: CreateRoom :one
INSERT INTO rooms (name, owner_id, is_public, max_participants)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetRoomById :one
SELECT * FROM rooms WHERE id = $1;

-- name: ListRooms :many
SELECT * FROM rooms ORDER BY created_at DESC;

-- name: ListRoomsByUser :many
SELECT r.* FROM rooms r
JOIN room_members rm ON r.id = rm.room_id
WHERE rm.user_id = $1
ORDER BY r.created_at DESC;

-- name: UpdateRoom :one
UPDATE rooms
SET name = COALESCE($2, name),
    is_public = COALESCE($3, is_public),
    max_participants = COALESCE($4, max_participants),
    settings = COALESCE($5, settings)
WHERE id = $1
RETURNING *;

-- name: DeleteRoom :exec
DELETE FROM rooms WHERE id = $1;

-- name: AddRoomMember :exec
INSERT INTO room_members (room_id, user_id, role)
VALUES ($1, $2, $3);

-- name: RemoveRoomMember :exec
DELETE FROM room_members WHERE room_id = $1 AND user_id = $2;

-- name: GetRoomMembers :many
SELECT u.*, rm.role, rm.joined_at FROM users u
JOIN room_members rm ON u.id = rm.user_id
WHERE rm.room_id = $1
ORDER BY rm.joined_at ASC;

-- name: IsRoomMember :one
SELECT EXISTS(SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2);
