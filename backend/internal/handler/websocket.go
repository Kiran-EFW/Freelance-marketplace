package handler

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/config"
)

// ---------------------------------------------------------------------------
// WebSocket message types
// ---------------------------------------------------------------------------

// WSMessageType defines the types of WebSocket messages.
type WSMessageType string

const (
	WSTypeNewMessage      WSMessageType = "new_message"
	WSTypeMessageRead     WSMessageType = "message_read"
	WSTypeTypingIndicator WSMessageType = "typing_indicator"
	WSTypeUserOnline      WSMessageType = "user_online"
	WSTypeUserOffline     WSMessageType = "user_offline"
	WSTypePong            WSMessageType = "pong"
	WSTypePing            WSMessageType = "ping"
)

// WSMessage is the envelope for all WebSocket messages.
type WSMessage struct {
	Type    WSMessageType   `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// WSNewMessagePayload is sent when a new chat message arrives.
type WSNewMessagePayload struct {
	ID             uuid.UUID `json:"id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	SenderID       uuid.UUID `json:"sender_id"`
	Content        string    `json:"content"`
	MessageType    string    `json:"message_type"`
	CreatedAt      time.Time `json:"created_at"`
}

// WSMessageReadPayload is sent when messages in a conversation are read.
type WSMessageReadPayload struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	ReaderID       uuid.UUID `json:"reader_id"`
	ReadAt         time.Time `json:"read_at"`
}

// WSTypingPayload is sent when a user starts or stops typing.
type WSTypingPayload struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	UserID         uuid.UUID `json:"user_id"`
	IsTyping       bool      `json:"is_typing"`
}

// WSUserStatusPayload is sent when a user comes online or goes offline.
type WSUserStatusPayload struct {
	UserID uuid.UUID `json:"user_id"`
}

// ---------------------------------------------------------------------------
// Hub — manages connected WebSocket clients
// ---------------------------------------------------------------------------

// Hub maintains the set of active WebSocket connections.
type Hub struct {
	mu          sync.RWMutex
	connections map[uuid.UUID][]*websocket.Conn // userID -> slice of connections (multi-tab)
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		connections: make(map[uuid.UUID][]*websocket.Conn),
	}
}

// Register adds a connection for a user.
func (h *Hub) Register(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connections[userID] = append(h.connections[userID], conn)
	log.Info().Str("user_id", userID.String()).Int("connections", len(h.connections[userID])).Msg("websocket client connected")
}

// Unregister removes a connection for a user.
func (h *Hub) Unregister(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns := h.connections[userID]
	filtered := make([]*websocket.Conn, 0, len(conns))
	for _, c := range conns {
		if c != conn {
			filtered = append(filtered, c)
		}
	}

	if len(filtered) == 0 {
		delete(h.connections, userID)
		log.Info().Str("user_id", userID.String()).Msg("websocket client fully disconnected")
	} else {
		h.connections[userID] = filtered
		log.Info().Str("user_id", userID.String()).Int("remaining", len(filtered)).Msg("websocket tab disconnected")
	}
}

// IsOnline checks if a user has any active connections.
func (h *Hub) IsOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections[userID]) > 0
}

// GetOnlineUserIDs returns a list of all online user IDs.
func (h *Hub) GetOnlineUserIDs() []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ids := make([]uuid.UUID, 0, len(h.connections))
	for id := range h.connections {
		ids = append(ids, id)
	}
	return ids
}

// SendToUser sends a message to all connections of a specific user.
func (h *Hub) SendToUser(userID uuid.UUID, msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to marshal websocket message")
		return
	}

	h.mu.RLock()
	conns := h.connections[userID]
	h.mu.RUnlock()

	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to send websocket message, will be cleaned up")
		}
	}
}

// BroadcastNewMessage notifies the recipient of a new message.
func (h *Hub) BroadcastNewMessage(recipientID uuid.UUID, payload WSNewMessagePayload) {
	data, _ := json.Marshal(payload)
	h.SendToUser(recipientID, WSMessage{
		Type:    WSTypeNewMessage,
		Payload: json.RawMessage(data),
	})
}

// BroadcastMessageRead notifies the sender that their messages were read.
func (h *Hub) BroadcastMessageRead(recipientID uuid.UUID, payload WSMessageReadPayload) {
	data, _ := json.Marshal(payload)
	h.SendToUser(recipientID, WSMessage{
		Type:    WSTypeMessageRead,
		Payload: json.RawMessage(data),
	})
}

// BroadcastTyping notifies the other participant about typing status.
func (h *Hub) BroadcastTyping(recipientID uuid.UUID, payload WSTypingPayload) {
	data, _ := json.Marshal(payload)
	h.SendToUser(recipientID, WSMessage{
		Type:    WSTypeTypingIndicator,
		Payload: json.RawMessage(data),
	})
}

// BroadcastUserOnline notifies relevant users that a user came online.
func (h *Hub) BroadcastUserOnline(userID uuid.UUID) {
	payload := WSUserStatusPayload{UserID: userID}
	data, _ := json.Marshal(payload)
	msg := WSMessage{Type: WSTypeUserOnline, Payload: json.RawMessage(data)}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for id := range h.connections {
		if id != userID {
			h.SendToUser(id, msg)
		}
	}
}

// BroadcastUserOffline notifies relevant users that a user went offline.
func (h *Hub) BroadcastUserOffline(userID uuid.UUID) {
	payload := WSUserStatusPayload{UserID: userID}
	data, _ := json.Marshal(payload)
	msg := WSMessage{Type: WSTypeUserOffline, Payload: json.RawMessage(data)}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for id := range h.connections {
		if id != userID {
			h.SendToUser(id, msg)
		}
	}
}

// ---------------------------------------------------------------------------
// WebSocket Handler
// ---------------------------------------------------------------------------

// WebSocketHandler manages WebSocket connections.
type WebSocketHandler struct {
	hub *Hub
	cfg *config.Config
}

// NewWebSocketHandler creates a new WebSocketHandler.
func NewWebSocketHandler(hub *Hub, cfg *config.Config) *WebSocketHandler {
	return &WebSocketHandler{hub: hub, cfg: cfg}
}

// UpgradeMiddleware is the Fiber middleware that checks WebSocket upgrade requests.
// It must be applied before the websocket.New() handler.
func (h *WebSocketHandler) UpgradeMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			// Validate JWT from query parameter
			tokenStr := c.Query("token")
			if tokenStr == "" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "missing token query parameter",
				})
			}

			userID, err := h.validateToken(tokenStr)
			if err != nil {
				log.Warn().Err(err).Msg("websocket: invalid JWT token")
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "invalid or expired token",
				})
			}

			c.Locals("ws_user_id", userID)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}

// HandleWebSocket is the WebSocket handler function.
func (h *WebSocketHandler) HandleWebSocket() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		userID, ok := c.Locals("ws_user_id").(uuid.UUID)
		if !ok || userID == uuid.Nil {
			log.Warn().Msg("websocket: no user ID in locals")
			c.Close()
			return
		}

		// Register the connection
		h.hub.Register(userID, c)
		h.hub.BroadcastUserOnline(userID)

		// Set up ping/pong for heartbeat
		c.SetPongHandler(func(appData string) error {
			return nil
		})

		// Start heartbeat ticker
		done := make(chan struct{})
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
						return
					}
				case <-done:
					return
				}
			}
		}()

		// Read loop
		for {
			_, msgBytes, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					log.Warn().Err(err).Str("user_id", userID.String()).Msg("websocket: unexpected close")
				}
				break
			}

			// Parse incoming message
			var msg WSMessage
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Warn().Err(err).Str("user_id", userID.String()).Msg("websocket: invalid message format")
				continue
			}

			h.handleIncomingMessage(userID, msg)
		}

		// Cleanup
		close(done)
		h.hub.Unregister(userID, c)

		// Only broadcast offline if user has no more connections
		if !h.hub.IsOnline(userID) {
			h.hub.BroadcastUserOffline(userID)
		}
	})
}

// handleIncomingMessage processes messages received from clients.
func (h *WebSocketHandler) handleIncomingMessage(userID uuid.UUID, msg WSMessage) {
	switch msg.Type {
	case WSTypePing:
		// Respond with pong
		pongData, _ := json.Marshal(map[string]string{"ts": time.Now().UTC().Format(time.RFC3339)})
		h.hub.SendToUser(userID, WSMessage{
			Type:    WSTypePong,
			Payload: json.RawMessage(pongData),
		})

	case WSTypeTypingIndicator:
		// Forward typing indicator to the other participant
		var payload WSTypingPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Warn().Err(err).Msg("websocket: invalid typing payload")
			return
		}
		payload.UserID = userID // Ensure the sender is correct
		// The recipient needs to be resolved from the conversation
		// For now, we broadcast to all online users (in a production system,
		// you'd look up the conversation participants)
		h.hub.BroadcastTyping(payload.UserID, payload)

	default:
		log.Debug().Str("type", string(msg.Type)).Str("user_id", userID.String()).Msg("websocket: unhandled message type")
	}
}

// validateToken extracts and validates a JWT token, returning the user ID.
func (h *WebSocketHandler) validateToken(tokenStr string) (uuid.UUID, error) {
	// Strip "Bearer " prefix if present
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(h.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, err
	}

	return uuid.Parse(claims.Subject)
}

// GetHub returns the Hub instance for use by other handlers.
func (h *WebSocketHandler) GetHub() *Hub {
	return h.hub
}
