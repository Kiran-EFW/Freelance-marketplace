package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/middleware"
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
	Type   string                 `json:"type"` // create_job, select_category, confirm_booking, set_schedule
	Label  string                 `json:"label"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// PhotoAnalysisResult holds the AI photo analysis output.
type PhotoAnalysisResult struct {
	CategoryID  *string `json:"category_id,omitempty"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// TranslationResult holds the AI translation output.
type TranslationResult struct {
	OriginalText   string `json:"original_text"`
	TranslatedText string `json:"translated_text"`
	SourceLanguage string `json:"source_language"`
	TargetLanguage string `json:"target_language"`
}

// PriceEstimate holds the AI price estimation result.
type PriceEstimate struct {
	Category   string  `json:"category"`
	Postcode   string  `json:"postcode"`
	MinPrice   float64 `json:"min_price"`
	MaxPrice   float64 `json:"max_price"`
	AvgPrice   float64 `json:"avg_price"`
	Currency   string  `json:"currency"`
	Confidence float64 `json:"confidence"`
}

// AIService defines the business operations required by AIHandler.
type AIService interface {
	ChatBooking(ctx context.Context, userID uuid.UUID, messages []ChatMessage) (*ChatResponse, error)
	AnalyzePhoto(ctx context.Context, userID uuid.UUID, imageData []byte, filename string) (*PhotoAnalysisResult, error)
	TranslateMessage(ctx context.Context, text, sourceLang, targetLang string) (*TranslationResult, error)
	GetPriceEstimate(ctx context.Context, category, postcode string) (*PriceEstimate, error)
}

// AIHandler handles AI-powered endpoints.
type AIHandler struct {
	service AIService
}

// NewAIHandler creates a new AIHandler.
func NewAIHandler(svc AIService) *AIHandler {
	return &AIHandler{service: svc}
}

// RegisterRoutes mounts AI routes on the given Fiber router group.
func (h *AIHandler) RegisterRoutes(rg fiber.Router) {
	rg.Post("/chat", h.ChatBooking)
	rg.Post("/photo", h.AnalyzePhoto)
	rg.Post("/translate", h.TranslateMessage)
	rg.Post("/price-estimate", h.GetPriceEstimate)
}

// chatBookingRequest is the payload for POST /api/v1/ai/chat.
type chatBookingRequest struct {
	Messages []ChatMessage `json:"messages"`
}

func (r *chatBookingRequest) validate() error {
	if len(r.Messages) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "at least one message is required")
	}
	return nil
}

// ChatBooking handles conversational booking via AI.
// POST /api/v1/ai/chat
func (h *AIHandler) ChatBooking(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req chatBookingRequest
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

	response, err := h.service.ChatBooking(c.UserContext(), userID, req.Messages)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("AI chat booking failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to process chat message",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": response,
	})
}

// AnalyzePhoto handles photo analysis for job categorization.
// POST /api/v1/ai/photo (multipart form)
func (h *AIHandler) AnalyzePhoto(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	file, err := c.FormFile("photo")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "photo file is required",
			},
		})
	}

	// Validate file size (max 10MB).
	if file.Size > 10*1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "file size must not exceed 10MB",
			},
		})
	}

	// Read file contents.
	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to read uploaded file",
			},
		})
	}
	defer f.Close()

	imageData := make([]byte, file.Size)
	if _, err := f.Read(imageData); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to read image data",
			},
		})
	}

	result, err := h.service.AnalyzePhoto(c.UserContext(), userID, imageData, file.Filename)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("AI photo analysis failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to analyze photo",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": result,
	})
}

// translateMessageRequest is the payload for POST /api/v1/ai/translate.
type translateMessageRequest struct {
	Text           string `json:"text"`
	SourceLanguage string `json:"source_language"`
	TargetLanguage string `json:"target_language"`
}

func (r *translateMessageRequest) validate() error {
	if r.Text == "" {
		return fiber.NewError(fiber.StatusBadRequest, "text is required")
	}
	if r.TargetLanguage == "" {
		return fiber.NewError(fiber.StatusBadRequest, "target_language is required")
	}
	return nil
}

// TranslateMessage translates text between languages.
// POST /api/v1/ai/translate
func (h *AIHandler) TranslateMessage(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req translateMessageRequest
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

	result, err := h.service.TranslateMessage(c.UserContext(), req.Text, req.SourceLanguage, req.TargetLanguage)
	if err != nil {
		log.Error().Err(err).Msg("AI translation failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to translate message",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": result,
	})
}

// priceEstimateRequest is the payload for POST /api/v1/ai/price-estimate.
type priceEstimateRequest struct {
	Category string `json:"category"`
	Postcode string `json:"postcode"`
}

func (r *priceEstimateRequest) validate() error {
	if r.Category == "" {
		return fiber.NewError(fiber.StatusBadRequest, "category is required")
	}
	if r.Postcode == "" {
		return fiber.NewError(fiber.StatusBadRequest, "postcode is required")
	}
	return nil
}

// GetPriceEstimate returns a fair price range estimate for a service category.
// POST /api/v1/ai/price-estimate
func (h *AIHandler) GetPriceEstimate(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req priceEstimateRequest
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

	estimate, err := h.service.GetPriceEstimate(c.UserContext(), req.Category, req.Postcode)
	if err != nil {
		log.Error().Err(err).Str("category", req.Category).Str("postcode", req.Postcode).Msg("AI price estimation failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to estimate price",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": estimate,
	})
}
