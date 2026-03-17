-- name: CreateConversation :one
INSERT INTO conversations (participant_1, participant_2, job_id)
VALUES ($1, $2, $3) RETURNING *;

-- name: GetConversation :one
SELECT * FROM conversations WHERE id = $1;

-- name: GetConversationByParticipants :one
SELECT * FROM conversations
WHERE LEAST(participant_1, participant_2) = LEAST($1, $2)
  AND GREATEST(participant_1, participant_2) = GREATEST($1, $2)
  AND COALESCE(job_id, '00000000-0000-0000-0000-000000000000') = COALESCE($3, '00000000-0000-0000-0000-000000000000');

-- name: ListConversationsForUser :many
SELECT * FROM conversations
WHERE (participant_1 = $1 AND is_archived_1 = false)
   OR (participant_2 = $1 AND is_archived_2 = false)
ORDER BY last_message_at DESC NULLS LAST
LIMIT $2 OFFSET $3;

-- name: UpdateConversationLastMessage :exec
UPDATE conversations SET last_message_at = $2, last_message_preview = $3 WHERE id = $1;

-- name: CreateMessage :one
INSERT INTO messages (conversation_id, sender_id, content, message_type, attachment_url, attachment_type, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: ListMessages :many
SELECT * FROM messages WHERE conversation_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: MarkMessagesRead :exec
UPDATE messages SET is_read = true, read_at = NOW()
WHERE conversation_id = $1 AND sender_id != $2 AND is_read = false;

-- name: CountUnreadMessages :one
SELECT COUNT(*) FROM messages m
JOIN conversations c ON m.conversation_id = c.id
WHERE (c.participant_1 = $1 OR c.participant_2 = $1)
  AND m.sender_id != $1 AND m.is_read = false;
