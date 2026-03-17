package sms_session

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// SessionState represents the current state of an SMS conversation.
type SessionState string

const (
	StateIdle      SessionState = "idle"
	StateSearching SessionState = "searching"
	StateBooking   SessionState = "booking"
	StateRating    SessionState = "rating"
)

const (
	sessionKeyPrefix = "sms_session:"
	sessionTTL       = 30 * time.Minute
)

// SMSSession holds the conversation state for a phone number.
type SMSSession struct {
	Phone     string            `json:"phone"`
	State     SessionState      `json:"state"`
	Context   map[string]string `json:"context"`
	Language  string            `json:"language"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// SMSSessionManager manages SMS conversation state in Redis.
type SMSSessionManager struct {
	rdb *redis.Client
}

// NewSMSSessionManager creates a new SMSSessionManager backed by Redis.
func NewSMSSessionManager(rdb *redis.Client) *SMSSessionManager {
	return &SMSSessionManager{rdb: rdb}
}

// GetSession retrieves the session for a phone number. If no session exists,
// a new idle session is returned.
func (m *SMSSessionManager) GetSession(ctx context.Context, phone string) (*SMSSession, error) {
	key := sessionKeyPrefix + phone
	data, err := m.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return &SMSSession{
			Phone:     phone,
			State:     StateIdle,
			Context:   make(map[string]string),
			Language:  "en",
			UpdatedAt: time.Now(),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get session: %w", err)
	}

	var session SMSSession
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}
	return &session, nil
}

// SaveSession persists the session to Redis with the configured TTL.
func (m *SMSSessionManager) SaveSession(ctx context.Context, session *SMSSession) error {
	session.UpdatedAt = time.Now()
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	key := sessionKeyPrefix + session.Phone
	if err := m.rdb.Set(ctx, key, data, sessionTTL).Err(); err != nil {
		return fmt.Errorf("redis set session: %w", err)
	}
	return nil
}

// ClearSession removes the session for a phone number.
func (m *SMSSessionManager) ClearSession(ctx context.Context, phone string) error {
	key := sessionKeyPrefix + phone
	return m.rdb.Del(ctx, key).Err()
}

// ---------------------------------------------------------------------------
// Message templates by language
// ---------------------------------------------------------------------------

// MessageTemplates holds localised SMS reply templates.
type MessageTemplates struct {
	Help             string
	SearchResults    string
	NoResults        string
	BookingStarted   string
	BookingConfirmed string
	StatusReply      string
	NoActiveJobs     string
	CancelConfirmed  string
	CancelNotFound   string
	RatingPrompt     string
	RatingConfirmed  string
	InvalidCommand   string
	InvalidRating    string
}

// GetTemplates returns message templates for the given language code.
// Falls back to English if the language is not supported.
func GetTemplates(lang string) MessageTemplates {
	switch lang {
	case "hi":
		return hindiTemplates
	case "kn":
		return kannadaTemplates
	default:
		return englishTemplates
	}
}

var englishTemplates = MessageTemplates{
	Help: "Seva SMS Commands:\n" +
		"FIND <service> - Search providers\n" +
		"BOOK <number> - Book a provider\n" +
		"STATUS - Check active jobs\n" +
		"CANCEL <id> - Cancel a job\n" +
		"RATE <1-5> - Rate last job\n" +
		"HELP - Show this message",
	SearchResults:    "Found %d providers for '%s':\n%s\nReply BOOK <number> to book.",
	NoResults:        "No providers found for '%s'. Try a different search term.",
	BookingStarted:   "Booking with %s. They will contact you shortly. Reply STATUS to check.",
	BookingConfirmed: "Booking confirmed! Provider %s will contact you at your number.",
	StatusReply:      "Your active jobs:\n%s",
	NoActiveJobs:     "You have no active jobs. Reply FIND <service> to search.",
	CancelConfirmed:  "Job %s has been cancelled.",
	CancelNotFound:   "Job %s not found. Reply STATUS to see your active jobs.",
	RatingPrompt:     "Please rate your last service (1-5). Reply RATE <number>.",
	RatingConfirmed:  "Thank you! You rated the service %d/5.",
	InvalidCommand:   "Unknown command. Reply HELP for available commands.",
	InvalidRating:    "Rating must be between 1 and 5. Reply RATE <1-5>.",
}

var hindiTemplates = MessageTemplates{
	Help: "Seva SMS:\n" +
		"FIND <seva> - provider khojein\n" +
		"BOOK <number> - booking karein\n" +
		"STATUS - active jobs dekhein\n" +
		"CANCEL <id> - cancel karein\n" +
		"RATE <1-5> - rating dein\n" +
		"HELP - yeh message",
	SearchResults:    "'%2$s' ke liye %1$d provider mile:\n%3$s\nBOOK <number> bhejein booking ke liye.",
	NoResults:        "'%s' ke liye koi provider nahi mila. Kuch aur khojein.",
	BookingStarted:   "%s ke saath booking ho gayi. Woh jaldi sampark karenge.",
	BookingConfirmed: "Booking confirmed! Provider %s aapse sampark karenge.",
	StatusReply:      "Aapke active jobs:\n%s",
	NoActiveJobs:     "Koi active job nahi hai. FIND <seva> bhejein khojne ke liye.",
	CancelConfirmed:  "Job %s cancel ho gaya.",
	CancelNotFound:   "Job %s nahi mila. STATUS bhejein dekhne ke liye.",
	RatingPrompt:     "Apni akhiri seva ko rate karein (1-5). RATE <number> bhejein.",
	RatingConfirmed:  "Dhanyavaad! Aapne %d/5 rating di.",
	InvalidCommand:   "Galat command. HELP bhejein commands dekhne ke liye.",
	InvalidRating:    "Rating 1 se 5 ke beech honi chahiye. RATE <1-5> bhejein.",
}

var kannadaTemplates = MessageTemplates{
	Help: "Seva SMS:\n" +
		"FIND <seva> - provider hudukiri\n" +
		"BOOK <number> - booking maadi\n" +
		"STATUS - active jobs nodi\n" +
		"CANCEL <id> - cancel maadi\n" +
		"RATE <1-5> - rating kodi\n" +
		"HELP - ee message",
	SearchResults:    "'%2$s' ge %1$d provider sigiddaree:\n%3$s\nBOOK <number> kalisiri booking ge.",
	NoResults:        "'%s' ge yaaru provider sigalilla. Bere hesharu prayatnisi.",
	BookingStarted:   "%s jote booking aayitu. Avaru bega samparki suttaare.",
	BookingConfirmed: "Booking confirmed! Provider %s nimmannu samparki suttaare.",
	StatusReply:      "Nimma active jobs:\n%s",
	NoActiveJobs:     "Yaavu active job illa. FIND <seva> kalisiri hududalu.",
	CancelConfirmed:  "Job %s cancel aayitu.",
	CancelNotFound:   "Job %s sigalilla. STATUS kalisiri nodalu.",
	RatingPrompt:     "Nimma koneya sevege rate kodi (1-5). RATE <number> kalisiri.",
	RatingConfirmed:  "Dhanyavaadagalu! Neevu %d/5 rating kottiddeera.",
	InvalidCommand:   "Gotilla aada command. HELP kalisiri commands nodalu.",
	InvalidRating:    "Rating 1 rinda 5 ra madhye irabeku. RATE <1-5> kalisiri.",
}

// ---------------------------------------------------------------------------
// ProcessMessage - main entry point for SMS processing
// ---------------------------------------------------------------------------

// ProcessMessage handles an incoming SMS message and returns the reply text.
// It maintains conversation state through the Redis-backed session manager.
func (m *SMSSessionManager) ProcessMessage(ctx context.Context, phone, message string) string {
	session, err := m.GetSession(ctx, phone)
	if err != nil {
		log.Error().Err(err).Str("phone", phone).Msg("failed to get SMS session")
		return "Seva: An error occurred. Please try again."
	}

	templates := GetTemplates(session.Language)
	msg := strings.TrimSpace(message)
	upperMsg := strings.ToUpper(msg)

	var reply string

	switch {
	case upperMsg == "HELP":
		reply = templates.Help

	case strings.HasPrefix(upperMsg, "FIND ") || strings.HasPrefix(upperMsg, "SEARCH "):
		reply = m.handleSearch(ctx, session, msg, templates)

	case strings.HasPrefix(upperMsg, "BOOK "):
		reply = m.handleBook(ctx, session, msg, templates)

	case upperMsg == "STATUS":
		reply = m.handleStatus(ctx, session, templates)

	case strings.HasPrefix(upperMsg, "CANCEL "):
		reply = m.handleCancel(ctx, session, msg, templates)

	case strings.HasPrefix(upperMsg, "RATE "):
		reply = m.handleRate(ctx, session, msg, templates)

	default:
		// If in a specific state, try context-aware processing.
		switch session.State {
		case StateSearching:
			// Treat as a search refinement.
			reply = m.handleSearch(ctx, session, msg, templates)
		case StateRating:
			// Treat bare digit as a rating.
			reply = m.handleRate(ctx, session, "RATE "+msg, templates)
		default:
			reply = templates.InvalidCommand
		}
	}

	// Save updated session.
	if err := m.SaveSession(ctx, session); err != nil {
		log.Error().Err(err).Str("phone", phone).Msg("failed to save SMS session")
	}

	return reply
}

func (m *SMSSessionManager) handleSearch(ctx context.Context, session *SMSSession, msg string, t MessageTemplates) string {
	// Extract the search term after FIND/SEARCH prefix.
	term := msg
	upperMsg := strings.ToUpper(msg)
	if strings.HasPrefix(upperMsg, "FIND ") {
		term = strings.TrimSpace(msg[5:])
	} else if strings.HasPrefix(upperMsg, "SEARCH ") {
		term = strings.TrimSpace(msg[7:])
	}

	if term == "" {
		return t.InvalidCommand
	}

	session.State = StateSearching
	session.Context["search_term"] = term

	// In a real implementation, this would query the provider search service.
	// For now, we return a placeholder indicating the search was processed.
	log.Info().Str("phone", session.Phone).Str("term", term).Msg("SMS search request")

	// Simulate search results (in production, wire to searchSvc).
	session.Context["result_count"] = "0"
	return fmt.Sprintf(t.NoResults, term)
}

func (m *SMSSessionManager) handleBook(ctx context.Context, session *SMSSession, msg string, t MessageTemplates) string {
	parts := strings.Fields(msg)
	if len(parts) < 2 {
		return t.InvalidCommand
	}

	providerNum := parts[1]
	session.State = StateBooking
	session.Context["booking_provider"] = providerNum

	log.Info().Str("phone", session.Phone).Str("provider", providerNum).Msg("SMS booking request")

	return fmt.Sprintf(t.BookingStarted, providerNum)
}

func (m *SMSSessionManager) handleStatus(ctx context.Context, session *SMSSession, t MessageTemplates) string {
	// In production, query the job service for active jobs for this phone number.
	log.Info().Str("phone", session.Phone).Msg("SMS status request")
	return t.NoActiveJobs
}

func (m *SMSSessionManager) handleCancel(ctx context.Context, session *SMSSession, msg string, t MessageTemplates) string {
	parts := strings.Fields(msg)
	if len(parts) < 2 {
		return t.InvalidCommand
	}

	jobID := parts[1]
	log.Info().Str("phone", session.Phone).Str("job_id", jobID).Msg("SMS cancel request")

	// In production, attempt to cancel the job via the job service.
	return fmt.Sprintf(t.CancelNotFound, jobID)
}

func (m *SMSSessionManager) handleRate(ctx context.Context, session *SMSSession, msg string, t MessageTemplates) string {
	parts := strings.Fields(msg)
	if len(parts) < 2 {
		session.State = StateRating
		return t.RatingPrompt
	}

	ratingStr := parts[1]
	if len(ratingStr) != 1 || ratingStr[0] < '1' || ratingStr[0] > '5' {
		return t.InvalidRating
	}

	rating := int(ratingStr[0] - '0')
	session.State = StateIdle
	session.Context["last_rating"] = ratingStr

	log.Info().Str("phone", session.Phone).Int("rating", rating).Msg("SMS rating submitted")

	return fmt.Sprintf(t.RatingConfirmed, rating)
}
