package handler

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/middleware"
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

// MessageService defines the business operations required by MessageHandler.
type MessageService interface {
	CreateConversation(ctx context.Context, participant1, participant2 uuid.UUID, jobID *uuid.UUID) (*Conversation, error)
	GetConversation(ctx context.Context, id uuid.UUID) (*Conversation, error)
	GetConversationByParticipants(ctx context.Context, p1, p2 uuid.UUID, jobID *uuid.UUID) (*Conversation, error)
	ListConversationsForUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Conversation, error)
	CreateMessage(ctx context.Context, msg *Message) (*Message, error)
	ListMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]Message, error)
	MarkMessagesRead(ctx context.Context, conversationID, userID uuid.UUID) error
	CountUnreadMessages(ctx context.Context, userID uuid.UUID) (int64, error)
	UpdateConversationLastMessage(ctx context.Context, conversationID uuid.UUID, lastMessageAt time.Time, preview string) error
}

// MessageHandler handles messaging endpoints.
type MessageHandler struct {
	service MessageService
}

// NewMessageHandler creates a new MessageHandler.
func NewMessageHandler(svc MessageService) *MessageHandler {
	return &MessageHandler{service: svc}
}

// RegisterRoutes mounts messaging routes on the given Fiber router group.
func (h *MessageHandler) RegisterRoutes(rg fiber.Router) {
	rg.Get("/conversations", h.ListConversations)
	rg.Post("/conversations", h.CreateOrGetConversation)
	rg.Get("/conversations/:id", h.GetConversationWithMessages)
	rg.Post("/conversations/:id/messages", h.SendMessage)
	rg.Put("/conversations/:id/read", h.MarkRead)
	rg.Get("/unread-count", h.GetUnreadCount)
}

// createConversationRequest is the payload for POST /api/v1/messages/conversations.
type createConversationRequest struct {
	RecipientID string `json:"recipient_id"`
	JobID       string `json:"job_id,omitempty"`
}

func (r *createConversationRequest) validate() error {
	if r.RecipientID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "recipient_id is required")
	}
	return nil
}

// sendMessageRequest is the payload for POST /api/v1/messages/conversations/:id/messages.
type sendMessageRequest struct {
	Content     string `json:"content"`
	MessageType string `json:"message_type,omitempty"`
}

func (r *sendMessageRequest) validate() error {
	if r.Content == "" {
		return fiber.NewError(fiber.StatusBadRequest, "content is required")
	}
	return nil
}

// ListConversations returns all conversations for the authenticated user.
// GET /api/v1/messages/conversations
func (h *MessageHandler) ListConversations(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	page, limit := parseMessagePagination(c)
	offset := (page - 1) * limit

	conversations, err := h.service.ListConversationsForUser(c.UserContext(), userID, limit, offset)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to list conversations")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list conversations",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": conversations,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
		},
	})
}

// CreateOrGetConversation creates a new conversation or returns an existing one.
// POST /api/v1/messages/conversations
func (h *MessageHandler) CreateOrGetConversation(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req createConversationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	if err := req.validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": err.Error(),
			},
		})
	}

	recipientID, err := uuid.Parse(req.RecipientID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid recipient_id format",
			},
		})
	}

	if recipientID == userID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "cannot create conversation with yourself",
			},
		})
	}

	var jobID *uuid.UUID
	if req.JobID != "" {
		parsed, err := uuid.Parse(req.JobID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "INVALID_ID",
					"message": "invalid job_id format",
				},
			})
		}
		jobID = &parsed
	}

	// Try to find an existing conversation first.
	existing, err := h.service.GetConversationByParticipants(c.UserContext(), userID, recipientID, jobID)
	if err == nil && existing != nil {
		return c.JSON(fiber.Map{
			"data": existing,
		})
	}

	// Create a new conversation.
	conversation, err := h.service.CreateConversation(c.UserContext(), userID, recipientID, jobID)
	if err != nil {
		log.Error().Err(err).
			Str("user_id", userID.String()).
			Str("recipient_id", recipientID.String()).
			Msg("failed to create conversation")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to create conversation",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": conversation,
	})
}

// GetConversationWithMessages returns a conversation and its messages.
// GET /api/v1/messages/conversations/:id
func (h *MessageHandler) GetConversationWithMessages(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	conversationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid conversation ID format",
			},
		})
	}

	conversation, err := h.service.GetConversation(c.UserContext(), conversationID)
	if err != nil {
		log.Error().Err(err).Str("conversation_id", conversationID.String()).Msg("failed to get conversation")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve conversation",
			},
		})
	}

	if conversation == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "conversation not found",
			},
		})
	}

	// Verify the user is a participant.
	if conversation.Participant1 != userID && conversation.Participant2 != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "FORBIDDEN",
				"message": "you are not a participant of this conversation",
			},
		})
	}

	page, limit := parseMessagePagination(c)
	offset := (page - 1) * limit

	messages, err := h.service.ListMessages(c.UserContext(), conversationID, limit, offset)
	if err != nil {
		log.Error().Err(err).Str("conversation_id", conversationID.String()).Msg("failed to list messages")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list messages",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"conversation": conversation,
			"messages":     messages,
		},
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
		},
	})
}

// SendMessage sends a new message in a conversation.
// POST /api/v1/messages/conversations/:id/messages
func (h *MessageHandler) SendMessage(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	conversationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid conversation ID format",
			},
		})
	}

	// Verify the user is a participant.
	conversation, err := h.service.GetConversation(c.UserContext(), conversationID)
	if err != nil || conversation == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "conversation not found",
			},
		})
	}

	if conversation.Participant1 != userID && conversation.Participant2 != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "FORBIDDEN",
				"message": "you are not a participant of this conversation",
			},
		})
	}

	var req sendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	if err := req.validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": err.Error(),
			},
		})
	}

	msgType := req.MessageType
	if msgType == "" {
		msgType = "text"
	}

	now := time.Now().UTC()
	msg := &Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       userID,
		Content:        req.Content,
		MessageType:    msgType,
		IsRead:         false,
		CreatedAt:      now,
	}

	created, err := h.service.CreateMessage(c.UserContext(), msg)
	if err != nil {
		log.Error().Err(err).
			Str("conversation_id", conversationID.String()).
			Str("sender_id", userID.String()).
			Msg("failed to send message")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to send message",
			},
		})
	}

	// Update conversation's last message preview.
	preview := req.Content
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}
	if err := h.service.UpdateConversationLastMessage(c.UserContext(), conversationID, now, preview); err != nil {
		log.Warn().Err(err).Str("conversation_id", conversationID.String()).Msg("failed to update conversation last message")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": created,
	})
}

// MarkRead marks all messages in a conversation as read for the authenticated user.
// PUT /api/v1/messages/conversations/:id/read
func (h *MessageHandler) MarkRead(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	conversationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid conversation ID format",
			},
		})
	}

	// Verify the user is a participant.
	conversation, err := h.service.GetConversation(c.UserContext(), conversationID)
	if err != nil || conversation == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "conversation not found",
			},
		})
	}

	if conversation.Participant1 != userID && conversation.Participant2 != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "FORBIDDEN",
				"message": "you are not a participant of this conversation",
			},
		})
	}

	if err := h.service.MarkMessagesRead(c.UserContext(), conversationID, userID); err != nil {
		log.Error().Err(err).
			Str("conversation_id", conversationID.String()).
			Str("user_id", userID.String()).
			Msg("failed to mark messages as read")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to mark messages as read",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "messages marked as read",
		},
	})
}

// GetUnreadCount returns the total unread message count for the authenticated user.
// GET /api/v1/messages/unread-count
func (h *MessageHandler) GetUnreadCount(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	count, err := h.service.CountUnreadMessages(c.UserContext(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to count unread messages")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to count unread messages",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"count": count,
		},
	})
}

// parseMessagePagination extracts page and limit from query parameters with defaults.
func parseMessagePagination(c *fiber.Ctx) (page, limit int) {
	page = 1
	limit = 30

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}

	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	if limit > 100 {
		limit = 100
	}

	return page, limit
}
