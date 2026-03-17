// Package messaging provides the messaging service implementation with real
// database operations for conversations and messages between users.
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Conversation represents a messaging conversation between two users.
type Conversation struct {
	ID                 uuid.UUID  `json:"id"`
	JobID              *uuid.UUID `json:"job_id,omitempty"`
	Participant1       uuid.UUID  `json:"participant_1"`
	Participant2       uuid.UUID  `json:"participant_2"`
	LastMessageAt      *time.Time `json:"last_message_at,omitempty"`
	LastMessagePreview *string    `json:"last_message_preview,omitempty"`
	IsArchived1        bool       `json:"is_archived_1"`
	IsArchived2        bool       `json:"is_archived_2"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// Message represents a single message within a conversation.
type Message struct {
	ID             uuid.UUID        `json:"id"`
	ConversationID uuid.UUID        `json:"conversation_id"`
	SenderID       uuid.UUID        `json:"sender_id"`
	Content        string           `json:"content"`
	MessageType    string           `json:"message_type"`
	AttachmentURL  *string          `json:"attachment_url,omitempty"`
	AttachmentType *string          `json:"attachment_type,omitempty"`
	Metadata       *json.RawMessage `json:"metadata,omitempty"`
	IsRead         bool             `json:"is_read"`
	ReadAt         *time.Time       `json:"read_at,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
}

// MessagingService implements messaging operations using direct database
// queries against the conversations and messages tables.
type MessagingService struct {
	db *pgxpool.Pool
}

// NewMessagingService returns a ready-to-use MessagingService.
func NewMessagingService(db *pgxpool.Pool) *MessagingService {
	return &MessagingService{db: db}
}

// ensureTablesExist creates the conversations and messages tables if they
// do not already exist. This is called lazily on first use.
func (s *MessagingService) ensureTablesExist(ctx context.Context) error {
	ddl := `
	CREATE TABLE IF NOT EXISTS conversations (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		job_id UUID REFERENCES jobs(id),
		participant_1 UUID NOT NULL REFERENCES users(id),
		participant_2 UUID NOT NULL REFERENCES users(id),
		last_message_at TIMESTAMPTZ,
		last_message_preview TEXT,
		is_archived_1 BOOLEAN NOT NULL DEFAULT FALSE,
		is_archived_2 BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE (participant_1, participant_2, job_id)
	);

	CREATE TABLE IF NOT EXISTS messages (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		conversation_id UUID NOT NULL REFERENCES conversations(id),
		sender_id UUID NOT NULL REFERENCES users(id),
		content TEXT NOT NULL,
		message_type TEXT NOT NULL DEFAULT 'text',
		attachment_url TEXT,
		attachment_type TEXT,
		metadata JSONB,
		is_read BOOLEAN NOT NULL DEFAULT FALSE,
		read_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id, created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_conversations_participants ON conversations(participant_1, participant_2);
	`
	_, err := s.db.Exec(ctx, ddl)
	return err
}

// CreateConversation creates a new conversation between two participants.
func (s *MessagingService) CreateConversation(ctx context.Context, participant1, participant2 uuid.UUID, jobID *uuid.UUID) (*Conversation, error) {
	if err := s.ensureTablesExist(ctx); err != nil {
		log.Warn().Err(err).Msg("failed to ensure messaging tables exist")
	}

	// Normalize participant order for consistent deduplication.
	p1, p2 := participant1, participant2
	if p1.String() > p2.String() {
		p1, p2 = p2, p1
	}

	now := time.Now().UTC()
	conv := &Conversation{
		ID:           uuid.New(),
		JobID:        jobID,
		Participant1: p1,
		Participant2: p2,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	query := `INSERT INTO conversations (id, job_id, participant_1, participant_2, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (participant_1, participant_2, job_id) DO NOTHING
		RETURNING id, created_at, updated_at`

	err := s.db.QueryRow(ctx, query, conv.ID, conv.JobID, conv.Participant1, conv.Participant2, conv.CreatedAt, conv.UpdatedAt).
		Scan(&conv.ID, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		// If ON CONFLICT DO NOTHING returned no row, look up existing.
		existing, err2 := s.GetConversationByParticipants(ctx, p1, p2, jobID)
		if err2 == nil && existing != nil {
			return existing, nil
		}
		return nil, fmt.Errorf("create conversation: %w", err)
	}

	log.Info().
		Str("conversation_id", conv.ID.String()).
		Str("p1", p1.String()).
		Str("p2", p2.String()).
		Msg("conversation created")

	return conv, nil
}

// GetConversation retrieves a conversation by ID.
func (s *MessagingService) GetConversation(ctx context.Context, id uuid.UUID) (*Conversation, error) {
	query := `SELECT id, job_id, participant_1, participant_2, last_message_at, last_message_preview,
		is_archived_1, is_archived_2, created_at, updated_at
		FROM conversations WHERE id = $1`

	var conv Conversation
	var lastMsgAt pgtype.Timestamptz
	var lastMsgPreview pgtype.Text

	err := s.db.QueryRow(ctx, query, id).Scan(
		&conv.ID, &conv.JobID, &conv.Participant1, &conv.Participant2,
		&lastMsgAt, &lastMsgPreview,
		&conv.IsArchived1, &conv.IsArchived2, &conv.CreatedAt, &conv.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get conversation: %w", err)
	}

	if lastMsgAt.Valid {
		t := lastMsgAt.Time
		conv.LastMessageAt = &t
	}
	if lastMsgPreview.Valid {
		conv.LastMessagePreview = &lastMsgPreview.String
	}

	return &conv, nil
}

// GetConversationByParticipants looks up an existing conversation between
// two users, optionally scoped to a job.
func (s *MessagingService) GetConversationByParticipants(ctx context.Context, p1, p2 uuid.UUID, jobID *uuid.UUID) (*Conversation, error) {
	// Normalize order.
	if p1.String() > p2.String() {
		p1, p2 = p2, p1
	}

	var query string
	var args []interface{}
	if jobID != nil {
		query = `SELECT id, job_id, participant_1, participant_2, last_message_at, last_message_preview,
			is_archived_1, is_archived_2, created_at, updated_at
			FROM conversations WHERE participant_1 = $1 AND participant_2 = $2 AND job_id = $3`
		args = []interface{}{p1, p2, *jobID}
	} else {
		query = `SELECT id, job_id, participant_1, participant_2, last_message_at, last_message_preview,
			is_archived_1, is_archived_2, created_at, updated_at
			FROM conversations WHERE participant_1 = $1 AND participant_2 = $2 AND job_id IS NULL`
		args = []interface{}{p1, p2}
	}

	var conv Conversation
	var lastMsgAt pgtype.Timestamptz
	var lastMsgPreview pgtype.Text

	err := s.db.QueryRow(ctx, query, args...).Scan(
		&conv.ID, &conv.JobID, &conv.Participant1, &conv.Participant2,
		&lastMsgAt, &lastMsgPreview,
		&conv.IsArchived1, &conv.IsArchived2, &conv.CreatedAt, &conv.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get conversation by participants: %w", err)
	}

	if lastMsgAt.Valid {
		t := lastMsgAt.Time
		conv.LastMessageAt = &t
	}
	if lastMsgPreview.Valid {
		conv.LastMessagePreview = &lastMsgPreview.String
	}

	return &conv, nil
}

// ListConversationsForUser returns all conversations involving the given user.
func (s *MessagingService) ListConversationsForUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Conversation, error) {
	query := `SELECT id, job_id, participant_1, participant_2, last_message_at, last_message_preview,
		is_archived_1, is_archived_2, created_at, updated_at
		FROM conversations
		WHERE (participant_1 = $1 OR participant_2 = $1)
		ORDER BY COALESCE(last_message_at, created_at) DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	defer rows.Close()

	var convs []Conversation
	for rows.Next() {
		var conv Conversation
		var lastMsgAt pgtype.Timestamptz
		var lastMsgPreview pgtype.Text

		if err := rows.Scan(
			&conv.ID, &conv.JobID, &conv.Participant1, &conv.Participant2,
			&lastMsgAt, &lastMsgPreview,
			&conv.IsArchived1, &conv.IsArchived2, &conv.CreatedAt, &conv.UpdatedAt,
		); err != nil {
			log.Warn().Err(err).Msg("failed to scan conversation row, skipping")
			continue
		}

		if lastMsgAt.Valid {
			t := lastMsgAt.Time
			conv.LastMessageAt = &t
		}
		if lastMsgPreview.Valid {
			conv.LastMessagePreview = &lastMsgPreview.String
		}

		convs = append(convs, conv)
	}

	return convs, nil
}

// CreateMessage creates a new message in a conversation.
func (s *MessagingService) CreateMessage(ctx context.Context, msg *Message) (*Message, error) {
	query := `INSERT INTO messages (id, conversation_id, sender_id, content, message_type, attachment_url, attachment_type, metadata, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at`

	err := s.db.QueryRow(ctx, query,
		msg.ID, msg.ConversationID, msg.SenderID, msg.Content,
		msg.MessageType, msg.AttachmentURL, msg.AttachmentType,
		msg.Metadata, msg.IsRead, msg.CreatedAt,
	).Scan(&msg.ID, &msg.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	log.Info().
		Str("message_id", msg.ID.String()).
		Str("conversation_id", msg.ConversationID.String()).
		Str("sender_id", msg.SenderID.String()).
		Msg("message created")

	return msg, nil
}

// ListMessages returns messages in a conversation ordered by creation time.
func (s *MessagingService) ListMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]Message, error) {
	query := `SELECT id, conversation_id, sender_id, content, message_type,
		attachment_url, attachment_type, metadata, is_read, read_at, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(ctx, query, conversationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var msg Message
		var attachURL, attachType pgtype.Text
		var metadata json.RawMessage
		var readAt pgtype.Timestamptz

		if err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Content,
			&msg.MessageType, &attachURL, &attachType, &metadata,
			&msg.IsRead, &readAt, &msg.CreatedAt,
		); err != nil {
			log.Warn().Err(err).Msg("failed to scan message row, skipping")
			continue
		}

		if attachURL.Valid {
			msg.AttachmentURL = &attachURL.String
		}
		if attachType.Valid {
			msg.AttachmentType = &attachType.String
		}
		if len(metadata) > 0 {
			m := json.RawMessage(metadata)
			msg.Metadata = &m
		}
		if readAt.Valid {
			t := readAt.Time
			msg.ReadAt = &t
		}

		msgs = append(msgs, msg)
	}

	return msgs, nil
}

// MarkMessagesRead marks all unread messages in a conversation as read
// for the given user (i.e., messages NOT sent by the user).
func (s *MessagingService) MarkMessagesRead(ctx context.Context, conversationID, userID uuid.UUID) error {
	query := `UPDATE messages
		SET is_read = true, read_at = NOW()
		WHERE conversation_id = $1 AND sender_id != $2 AND is_read = false`

	_, err := s.db.Exec(ctx, query, conversationID, userID)
	if err != nil {
		return fmt.Errorf("mark messages read: %w", err)
	}

	return nil
}

// CountUnreadMessages returns the total count of unread messages for a user
// across all conversations.
func (s *MessagingService) CountUnreadMessages(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM messages m
		JOIN conversations c ON c.id = m.conversation_id
		WHERE (c.participant_1 = $1 OR c.participant_2 = $1)
		  AND m.sender_id != $1
		  AND m.is_read = false`

	var count int64
	err := s.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count unread messages: %w", err)
	}

	return count, nil
}

// UpdateConversationLastMessage updates the last_message_at and preview
// fields on a conversation.
func (s *MessagingService) UpdateConversationLastMessage(ctx context.Context, conversationID uuid.UUID, lastMessageAt time.Time, preview string) error {
	query := `UPDATE conversations
		SET last_message_at = $2, last_message_preview = $3, updated_at = NOW()
		WHERE id = $1`

	_, err := s.db.Exec(ctx, query, conversationID, lastMessageAt, preview)
	if err != nil {
		return fmt.Errorf("update conversation last message: %w", err)
	}

	return nil
}
