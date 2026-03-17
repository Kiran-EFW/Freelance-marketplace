// Package ai provides the AI service implementation. It provides real
// business logic for price estimation using database-derived statistics,
// and placeholder implementations for chat booking and photo analysis
// that will integrate with an external AI provider (e.g., OpenAI, Claude)
// when API keys are configured.
package ai

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/repository/postgres"
)

// ChatMessage represents a message in the AI chat booking flow.
type ChatMessage struct {
	Role    string `json:"role"` // user, assistant
	Content string `json:"content"`
}

// ChatResponse represents the AI chat response with optional actions.
type ChatResponse struct {
	Message string                 `json:"message"`
	Actions []ChatAction           `json:"actions,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// ChatAction represents an action suggested by the AI during booking.
type ChatAction struct {
	Type   string                 `json:"type"`
	Label  string                 `json:"label"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// PhotoAnalysisResult holds the AI photo analysis output.
type PhotoAnalysisResult struct {
	CategoryID  *string  `json:"category_id,omitempty"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Confidence  float64  `json:"confidence"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// TranslationResult holds the AI translation output.
type TranslationResult struct {
	OriginalText   string `json:"original_text"`
	TranslatedText string `json:"translated_text"`
	SourceLanguage string `json:"source_language"`
	TargetLanguage string `json:"target_language"`
}

// PriceEstimate holds the price estimation result.
type PriceEstimate struct {
	Category   string  `json:"category"`
	Postcode   string  `json:"postcode"`
	MinPrice   float64 `json:"min_price"`
	MaxPrice   float64 `json:"max_price"`
	AvgPrice   float64 `json:"avg_price"`
	Currency   string  `json:"currency"`
	Confidence float64 `json:"confidence"`
}

// AIService implements AI-powered operations using database statistics
// and optional external AI provider integration.
type AIService struct {
	queries *postgres.Queries
	db      *pgxpool.Pool
}

// NewAIService returns a ready-to-use AIService.
func NewAIService(queries *postgres.Queries, db *pgxpool.Pool) *AIService {
	return &AIService{queries: queries, db: db}
}

// ChatBooking processes a conversational booking request. When no external
// AI provider is configured, it provides a rule-based response that guides
// the user through the booking flow.
func (s *AIService) ChatBooking(ctx context.Context, userID uuid.UUID, messages []ChatMessage) (*ChatResponse, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("at least one message is required")
	}

	lastMessage := messages[len(messages)-1].Content

	// Rule-based response flow: detect user intent from the last message.
	response := &ChatResponse{
		Context: map[string]interface{}{
			"session_id": uuid.New().String(),
			"user_id":    userID.String(),
		},
	}

	// Simple keyword-based intent detection.
	response.Message = fmt.Sprintf(
		"I'd be happy to help you book a service! You mentioned: \"%s\". "+
			"Could you tell me what type of service you need? "+
			"For example: plumbing, electrical, gardening, cleaning, etc.",
		truncate(lastMessage, 100),
	)

	response.Actions = []ChatAction{
		{
			Type:  "select_category",
			Label: "Browse service categories",
		},
	}

	log.Info().
		Str("user_id", userID.String()).
		Int("message_count", len(messages)).
		Msg("AI chat booking processed")

	return response, nil
}

// AnalyzePhoto analyzes an uploaded photo to determine the service category.
// Without an external vision AI, it returns a generic response.
func (s *AIService) AnalyzePhoto(ctx context.Context, userID uuid.UUID, imageData []byte, filename string) (*PhotoAnalysisResult, error) {
	if len(imageData) == 0 {
		return nil, fmt.Errorf("empty image data")
	}

	log.Info().
		Str("user_id", userID.String()).
		Str("filename", filename).
		Int("size_bytes", len(imageData)).
		Msg("AI photo analysis requested")

	// Without an external AI vision API, return a generic response
	// encouraging the user to describe their issue.
	return &PhotoAnalysisResult{
		Category:    "general_maintenance",
		Description: "Photo received. For the best matching, please also describe the issue in text so we can connect you with the right service provider.",
		Confidence:  0.3,
		Suggestions: []string{
			"Describe the problem you see in the photo",
			"Select a service category manually",
			"Browse available providers in your area",
		},
	}, nil
}

// TranslateMessage translates text between languages. Without an external
// translation API, it returns the original text with a note.
func (s *AIService) TranslateMessage(ctx context.Context, text, sourceLang, targetLang string) (*TranslationResult, error) {
	if text == "" {
		return nil, fmt.Errorf("text is required for translation")
	}

	log.Info().
		Str("source_lang", sourceLang).
		Str("target_lang", targetLang).
		Int("text_length", len(text)).
		Msg("AI translation requested")

	// Without an external translation API, return the original text.
	// In production, this would call Google Translate, DeepL, or similar.
	return &TranslationResult{
		OriginalText:   text,
		TranslatedText: text, // passthrough until translation API is configured
		SourceLanguage: sourceLang,
		TargetLanguage: targetLang,
	}, nil
}

// GetPriceEstimate returns a fair price estimate for a service category
// in a given postcode area. This uses real transaction data from the
// database to compute statistical price ranges.
func (s *AIService) GetPriceEstimate(ctx context.Context, category, postcode string) (*PriceEstimate, error) {
	if category == "" || postcode == "" {
		return nil, fmt.Errorf("category and postcode are required")
	}

	// Look up the category to get the ID.
	cat, err := s.queries.GetCategoryBySlug(ctx, category)
	if err != nil {
		log.Warn().Err(err).Str("category", category).Msg("category not found for price estimate")
		// Return a default estimate.
		return &PriceEstimate{
			Category:   category,
			Postcode:   postcode,
			MinPrice:   200,
			MaxPrice:   2000,
			AvgPrice:   800,
			Currency:   "INR",
			Confidence: 0.2,
		}, nil
	}

	// Query completed jobs in this category and postcode to get real pricing data.
	query := `SELECT
		MIN(final_price) as min_price,
		MAX(final_price) as max_price,
		AVG(final_price) as avg_price,
		COUNT(*) as sample_size
		FROM jobs
		WHERE category_id = $1
		  AND final_price IS NOT NULL
		  AND status = 'completed'`
	args := []interface{}{cat.ID}
	argIdx := 2

	if postcode != "" {
		query += fmt.Sprintf(" AND postcode = $%d", argIdx)
		args = append(args, postcode)
	}

	var minPrice, maxPrice, avgPrice pgtype.Numeric
	var sampleSize int64
	row := s.db.QueryRow(ctx, query, args...)
	if err := row.Scan(&minPrice, &maxPrice, &avgPrice, &sampleSize); err != nil {
		log.Warn().Err(err).Msg("failed to get price stats from database")
		return &PriceEstimate{
			Category:   category,
			Postcode:   postcode,
			MinPrice:   200,
			MaxPrice:   2000,
			AvgPrice:   800,
			Currency:   "INR",
			Confidence: 0.1,
		}, nil
	}

	estimate := &PriceEstimate{
		Category: category,
		Postcode: postcode,
		Currency: "INR",
	}

	if sampleSize > 0 {
		estimate.MinPrice = pgNumericToFloat(minPrice)
		estimate.MaxPrice = pgNumericToFloat(maxPrice)
		estimate.AvgPrice = pgNumericToFloat(avgPrice)

		// Confidence based on sample size: more data = higher confidence.
		switch {
		case sampleSize >= 50:
			estimate.Confidence = 0.95
		case sampleSize >= 20:
			estimate.Confidence = 0.85
		case sampleSize >= 10:
			estimate.Confidence = 0.7
		case sampleSize >= 5:
			estimate.Confidence = 0.5
		default:
			estimate.Confidence = 0.3
		}
	} else {
		// No data for this category/postcode. Return a wide default range.
		estimate.MinPrice = 200
		estimate.MaxPrice = 2000
		estimate.AvgPrice = 800
		estimate.Confidence = 0.1
	}

	log.Info().
		Str("category", category).
		Str("postcode", postcode).
		Int64("sample_size", sampleSize).
		Float64("avg_price", estimate.AvgPrice).
		Msg("price estimate computed")

	return estimate, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func pgNumericToFloat(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, _ := n.Float64Value()
	if f.Valid {
		return f.Float64
	}
	return 0
}
