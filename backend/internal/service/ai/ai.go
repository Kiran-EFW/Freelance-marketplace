// Package ai provides the AI service implementation. It uses real AI
// providers (Claude, Google Vision, Google Translate) when configured,
// and falls back to rule-based / passthrough behaviour otherwise.
package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	aiadapter "github.com/seva-platform/backend/internal/adapter/ai"
	"github.com/seva-platform/backend/internal/repository/postgres"
)

// ---------------------------------------------------------------------------
// Provider interfaces – these are satisfied by the adapter package types.
// Defining them locally follows the Go idiom of consumers owning interfaces.
// ---------------------------------------------------------------------------

// ClaudeProvider abstracts the Claude AI client for chat and image analysis.
type ClaudeProvider interface {
	Chat(ctx context.Context, messages []aiadapter.Message, tools []aiadapter.Tool) (*aiadapter.Response, error)
	AnalyzeImage(ctx context.Context, imageBase64 string, prompt string) (*aiadapter.Response, error)
}

// VisionProvider abstracts the Google Vision client for OCR / image analysis.
type VisionProvider interface {
	DetectText(ctx context.Context, imageData []byte) (*aiadapter.OCRResult, error)
}

// TranslateProvider abstracts the Google Translate client for text translation.
type TranslateProvider interface {
	Translate(ctx context.Context, text, sourceLang, targetLang string) (string, error)
}

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
// and optional external AI provider integration. Each provider field may
// be nil; the methods gracefully fall back to rule-based / passthrough
// behaviour when the corresponding provider is not configured.
type AIService struct {
	queries    *postgres.Queries
	db         *pgxpool.Pool
	claude     ClaudeProvider
	vision     VisionProvider
	translator TranslateProvider
}

// NewAIService returns a ready-to-use AIService. Any of the provider
// arguments may be nil; the service will degrade gracefully.
func NewAIService(
	queries *postgres.Queries,
	db *pgxpool.Pool,
	claude ClaudeProvider,
	vision VisionProvider,
	translator TranslateProvider,
) *AIService {
	return &AIService{
		queries:    queries,
		db:         db,
		claude:     claude,
		vision:     vision,
		translator: translator,
	}
}

// chatBookingSystemPrompt is the system prompt used when Claude is available
// to guide the AI in the service booking conversation.
const chatBookingSystemPrompt = `You are Seva, an AI assistant that helps users book home services in India.
Your job is to understand what service the user needs, gather relevant details
(location, urgency, budget preferences), and guide them toward creating a
booking. Be friendly, concise, and helpful. Respond in the same language the
user writes in. When you have enough information to suggest a service category,
include a JSON block at the end of your message in the following format:
{"action":"select_category","category":"<category_slug>"}
Available service categories include: plumbing, electrical, carpentry,
painting, cleaning, gardening, pest_control, appliance_repair, ac_service,
general_maintenance.`

// ChatBooking processes a conversational booking request. When Claude is
// configured it produces a real AI-driven response; otherwise it falls back
// to a rule-based reply that guides the user through the booking flow.
func (s *AIService) ChatBooking(ctx context.Context, userID uuid.UUID, messages []ChatMessage) (*ChatResponse, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("at least one message is required")
	}

	response := &ChatResponse{
		Context: map[string]interface{}{
			"session_id": uuid.New().String(),
			"user_id":    userID.String(),
		},
	}

	// --- Real AI path ---
	if s.claude != nil {
		// Convert service ChatMessages to adapter Messages.
		adapterMsgs := make([]aiadapter.Message, 0, len(messages)+1)
		// Prepend a system-level instruction as the first user turn so that
		// Claude has context. (The Claude Messages API accepts a top-level
		// "system" param, but our adapter's Chat method uses only Messages;
		// we emulate it with an initial assistant turn.)
		adapterMsgs = append(adapterMsgs, aiadapter.Message{
			Role:    "user",
			Content: chatBookingSystemPrompt,
		})
		adapterMsgs = append(adapterMsgs, aiadapter.Message{
			Role:    "assistant",
			Content: "Understood. I will help the user book a service on Seva.",
		})
		for _, m := range messages {
			adapterMsgs = append(adapterMsgs, aiadapter.Message{
				Role:    m.Role,
				Content: m.Content,
			})
		}

		resp, err := s.claude.Chat(ctx, adapterMsgs, nil)
		if err != nil {
			log.Warn().Err(err).Msg("Claude chat failed, falling back to rule-based response")
			// Fall through to the rule-based path below.
		} else {
			response.Message = resp.Content

			// Try to extract a structured action from the response.
			if action := extractChatAction(resp.Content); action != nil {
				response.Actions = []ChatAction{*action}
			}

			log.Info().
				Str("user_id", userID.String()).
				Int("message_count", len(messages)).
				Msg("AI chat booking processed via Claude")

			return response, nil
		}
	}

	// --- Rule-based fallback ---
	lastMessage := messages[len(messages)-1].Content

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
		Msg("AI chat booking processed (rule-based fallback)")

	return response, nil
}

// extractChatAction tries to find a JSON action block in the Claude response
// text, e.g. {"action":"select_category","category":"plumbing"}.
func extractChatAction(text string) *ChatAction {
	// Look for a JSON object containing "action" somewhere in the response.
	start := strings.LastIndex(text, "{")
	end := strings.LastIndex(text, "}")
	if start == -1 || end == -1 || end <= start {
		return nil
	}

	raw := text[start : end+1]
	var parsed struct {
		Action   string `json:"action"`
		Category string `json:"category"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil || parsed.Action == "" {
		return nil
	}

	return &ChatAction{
		Type:  parsed.Action,
		Label: fmt.Sprintf("Select %s", parsed.Category),
		Params: map[string]interface{}{
			"category": parsed.Category,
		},
	}
}

// analyzePhotoPrompt is sent to Claude when both Vision OCR results and
// Claude are available, asking Claude to interpret the detected text.
const analyzePhotoPrompt = `Analyze this image of a home service issue. Determine:
1. What service category is needed (one of: plumbing, electrical, carpentry,
   painting, cleaning, gardening, pest_control, appliance_repair, ac_service,
   general_maintenance)
2. A brief description of the issue visible in the image
3. Any suggestions for the user

Respond in JSON format:
{"category":"<slug>","description":"<text>","confidence":<0-1>,"suggestions":["<s1>","<s2>"]}`

// AnalyzePhoto analyzes an uploaded photo to determine the service category.
// It uses Google Vision (OCR) and/or Claude (image understanding) when
// available, falling back to a generic response when neither is configured.
func (s *AIService) AnalyzePhoto(ctx context.Context, userID uuid.UUID, imageData []byte, filename string) (*PhotoAnalysisResult, error) {
	if len(imageData) == 0 {
		return nil, fmt.Errorf("empty image data")
	}

	log.Info().
		Str("user_id", userID.String()).
		Str("filename", filename).
		Int("size_bytes", len(imageData)).
		Msg("AI photo analysis requested")

	// --- Try Google Vision OCR ---
	var ocrText string
	if s.vision != nil {
		ocrResult, err := s.vision.DetectText(ctx, imageData)
		if err != nil {
			log.Warn().Err(err).Msg("Google Vision OCR failed, continuing without OCR")
		} else if ocrResult != nil && ocrResult.Text != "" {
			ocrText = ocrResult.Text
			log.Debug().Str("ocr_text", truncate(ocrText, 200)).Msg("OCR text detected")
		}
	}

	// --- Try Claude image analysis ---
	if s.claude != nil {
		imageB64 := base64.StdEncoding.EncodeToString(imageData)

		prompt := analyzePhotoPrompt
		if ocrText != "" {
			prompt += fmt.Sprintf("\n\nOCR text detected in the image:\n%s", ocrText)
		}

		resp, err := s.claude.AnalyzeImage(ctx, imageB64, prompt)
		if err != nil {
			log.Warn().Err(err).Msg("Claude image analysis failed, falling back")
		} else {
			result := parsePhotoAnalysisResponse(resp.Content)
			log.Info().
				Str("user_id", userID.String()).
				Str("category", result.Category).
				Float64("confidence", result.Confidence).
				Msg("AI photo analysis completed via Claude")
			return result, nil
		}
	}

	// --- Vision-only fallback (OCR text available but no Claude) ---
	if ocrText != "" {
		return &PhotoAnalysisResult{
			Category:    "general_maintenance",
			Description: fmt.Sprintf("Text detected in the image: %s", truncate(ocrText, 500)),
			Confidence:  0.5,
			Suggestions: []string{
				"Describe the problem you see in the photo",
				"Select a service category manually",
			},
		}, nil
	}

	// --- Generic fallback (no providers available) ---
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

// parsePhotoAnalysisResponse tries to parse Claude's JSON response into a
// PhotoAnalysisResult, falling back to using the raw text as a description.
func parsePhotoAnalysisResponse(content string) *PhotoAnalysisResult {
	// Try to extract JSON from the response.
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start != -1 && end != -1 && end > start {
		raw := content[start : end+1]
		var parsed struct {
			Category    string   `json:"category"`
			Description string   `json:"description"`
			Confidence  float64  `json:"confidence"`
			Suggestions []string `json:"suggestions"`
		}
		if err := json.Unmarshal([]byte(raw), &parsed); err == nil && parsed.Category != "" {
			return &PhotoAnalysisResult{
				Category:    parsed.Category,
				Description: parsed.Description,
				Confidence:  parsed.Confidence,
				Suggestions: parsed.Suggestions,
			}
		}
	}

	// If JSON parsing fails, use the entire response as the description.
	return &PhotoAnalysisResult{
		Category:    "general_maintenance",
		Description: content,
		Confidence:  0.6,
		Suggestions: []string{
			"Select a service category manually",
		},
	}
}

// TranslateMessage translates text between languages. When Google Translate
// is configured it performs a real translation; otherwise it returns the
// original text as a passthrough.
func (s *AIService) TranslateMessage(ctx context.Context, text, sourceLang, targetLang string) (*TranslationResult, error) {
	if text == "" {
		return nil, fmt.Errorf("text is required for translation")
	}

	log.Info().
		Str("source_lang", sourceLang).
		Str("target_lang", targetLang).
		Int("text_length", len(text)).
		Msg("AI translation requested")

	// --- Real translation path ---
	if s.translator != nil {
		translated, err := s.translator.Translate(ctx, text, sourceLang, targetLang)
		if err != nil {
			log.Warn().Err(err).Msg("Google Translate failed, falling back to passthrough")
			// Fall through to the passthrough below.
		} else {
			log.Info().
				Int("input_len", len(text)).
				Int("output_len", len(translated)).
				Msg("AI translation completed via Google Translate")

			return &TranslationResult{
				OriginalText:   text,
				TranslatedText: translated,
				SourceLanguage: sourceLang,
				TargetLanguage: targetLang,
			}, nil
		}
	}

	// --- Passthrough fallback ---
	return &TranslationResult{
		OriginalText:   text,
		TranslatedText: text,
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
