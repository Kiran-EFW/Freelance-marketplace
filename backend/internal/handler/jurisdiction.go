package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// JurisdictionServiceAPI defines the business operations required by JurisdictionHandler.
type JurisdictionServiceAPI interface {
	ListActive(ctx context.Context) ([]JurisdictionInfoResponse, error)
	GetJurisdiction(ctx context.Context, id string) (*JurisdictionInfoResponse, error)
	GetEnabledCategories(ctx context.Context, id string) ([]string, error)
	DetectJurisdiction(phone string) string
}

// JurisdictionInfoResponse is the API response for a jurisdiction.
type JurisdictionInfoResponse struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	DefaultLanguage string      `json:"default_language"`
	Currency        string      `json:"currency"`
	CurrencySymbol  string      `json:"currency_symbol"`
	PhonePrefix     string      `json:"phone_prefix"`
	Timezone        string      `json:"timezone"`
	IsActive        bool        `json:"is_active"`
	Config          interface{} `json:"config,omitempty"`
}

// JurisdictionHandler handles jurisdiction endpoints.
type JurisdictionHandler struct {
	service *JurisdictionHandlerService
}

// JurisdictionHandlerService wraps the jurisdiction service for the handler layer.
// This avoids a circular import by adapting the service layer types.
type JurisdictionHandlerService struct {
	listActive         func(ctx context.Context) (interface{}, error)
	getJurisdiction    func(ctx context.Context, id string) (interface{}, error)
	getCategories      func(ctx context.Context, id string) ([]string, error)
	detectJurisdiction func(phone string) string
}

// NewJurisdictionHandler creates a new JurisdictionHandler using function adapters
// to avoid coupling to the service package types.
func NewJurisdictionHandler(
	listActive func(ctx context.Context) (interface{}, error),
	getJurisdiction func(ctx context.Context, id string) (interface{}, error),
	getCategories func(ctx context.Context, id string) ([]string, error),
	detectJurisdiction func(phone string) string,
) *JurisdictionHandler {
	return &JurisdictionHandler{
		service: &JurisdictionHandlerService{
			listActive:         listActive,
			getJurisdiction:    getJurisdiction,
			getCategories:      getCategories,
			detectJurisdiction: detectJurisdiction,
		},
	}
}

// RegisterRoutes mounts jurisdiction routes on the given Fiber router group.
func (h *JurisdictionHandler) RegisterRoutes(rg fiber.Router) {
	rg.Get("/", h.ListJurisdictions)
	rg.Get("/detect", h.DetectJurisdiction)
	rg.Get("/:id", h.GetJurisdiction)
	rg.Get("/:id/categories", h.GetCategories)
}

// ListJurisdictions returns all active jurisdictions.
// GET /api/v1/jurisdictions
func (h *JurisdictionHandler) ListJurisdictions(c *fiber.Ctx) error {
	jurisdictions, err := h.service.listActive(c.UserContext())
	if err != nil {
		log.Error().Err(err).Msg("failed to list jurisdictions")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list jurisdictions",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": jurisdictions,
	})
}

// GetJurisdiction returns the config for a specific jurisdiction.
// GET /api/v1/jurisdictions/:id
func (h *JurisdictionHandler) GetJurisdiction(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "jurisdiction ID is required",
			},
		})
	}

	jurisdiction, err := h.service.getJurisdiction(c.UserContext(), id)
	if err != nil {
		log.Error().Err(err).Str("jurisdiction_id", id).Msg("failed to get jurisdiction")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "jurisdiction not found",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": jurisdiction,
	})
}

// GetCategories returns the enabled service categories for a jurisdiction.
// GET /api/v1/jurisdictions/:id/categories
func (h *JurisdictionHandler) GetCategories(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "jurisdiction ID is required",
			},
		})
	}

	categories, err := h.service.getCategories(c.UserContext(), id)
	if err != nil {
		log.Error().Err(err).Str("jurisdiction_id", id).Msg("failed to get categories")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "jurisdiction not found",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"jurisdiction_id": id,
			"categories":      categories,
		},
	})
}

// DetectJurisdiction auto-detects jurisdiction from a phone number prefix.
// GET /api/v1/jurisdictions/detect?phone=+91...
func (h *JurisdictionHandler) DetectJurisdiction(c *fiber.Ctx) error {
	phone := c.Query("phone")
	if phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "phone query parameter is required",
			},
		})
	}

	jurisdictionID := h.service.detectJurisdiction(phone)
	if jurisdictionID == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "could not detect jurisdiction from phone number",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"jurisdiction_id": jurisdictionID,
			"phone":           phone,
		},
	})
}
