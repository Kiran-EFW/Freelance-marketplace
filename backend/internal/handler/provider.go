package handler

import (
	"context"
	"encoding/json"
	"fmt"
	stdmime "mime"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/adapter/storage"
	"github.com/seva-platform/backend/internal/domain"
	"github.com/seva-platform/backend/internal/middleware"
)

// ProviderDashboard holds summary data for the provider dashboard.
type ProviderDashboard struct {
	TotalEarnings   float64 `json:"total_earnings"`
	MonthlyEarnings float64 `json:"monthly_earnings"`
	CompletedJobs   int     `json:"completed_jobs"`
	ActiveJobs      int     `json:"active_jobs"`
	UpcomingJobs    int     `json:"upcoming_jobs"`
	AverageRating   float64 `json:"average_rating"`
	TotalReviews    int     `json:"total_reviews"`
	TrustScore      float64 `json:"trust_score"`
}

// EarningsBreakdown holds earnings data for a specific period.
type EarningsBreakdown struct {
	Period    string  `json:"period"` // daily, weekly, monthly
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	JobsCount int     `json:"jobs_count"`
	Date      string  `json:"date"`
}

// KYCDocument represents an uploaded KYC verification document.
type KYCDocument struct {
	ID         uuid.UUID `json:"id"`
	ProviderID uuid.UUID `json:"provider_id"`
	Type       string    `json:"type"` // aadhaar, pan, license, etc.
	FileURL    string    `json:"file_url"`
	Status     string    `json:"status"` // pending, approved, rejected
	CreatedAt  time.Time `json:"created_at"`
}

// ProviderService defines the business operations required by ProviderHandler.
type ProviderService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*domain.ProviderProfile, error)
	GetPublicProfile(ctx context.Context, userID uuid.UUID) (*domain.ProviderProfile, error)
	UpdateProfile(ctx context.Context, profile *domain.ProviderProfile) error
	GetDashboard(ctx context.Context, userID uuid.UUID) (*ProviderDashboard, error)
	GetEarnings(ctx context.Context, userID uuid.UUID, period string) ([]EarningsBreakdown, error)
	UpdateAvailability(ctx context.Context, userID uuid.UUID, schedule json.RawMessage) error
	UploadKYCDocument(ctx context.Context, doc *KYCDocument) error
}

// ProviderHandler handles provider-specific endpoints.
type ProviderHandler struct {
	service ProviderService
	storage storage.StorageProvider
}

// NewProviderHandler creates a new ProviderHandler.
func NewProviderHandler(svc ProviderService, store storage.StorageProvider) *ProviderHandler {
	return &ProviderHandler{service: svc, storage: store}
}

// RegisterRoutes mounts provider routes on the given Fiber router group.
func (h *ProviderHandler) RegisterRoutes(rg fiber.Router) {
	rg.Get("/me", h.GetMyProfile)
	rg.Put("/me", h.UpdateProviderProfile)
	rg.Get("/me/dashboard", h.GetProviderDashboard)
	rg.Get("/me/earnings", h.GetProviderEarnings)
	rg.Patch("/me/availability", h.UpdateAvailability)
	rg.Post("/me/kyc", h.UploadKYCDocument)
	rg.Get("/:id", h.GetProviderProfile)
}

// GetProviderProfile returns a public provider profile.
// GET /api/v1/providers/:id
func (h *ProviderHandler) GetProviderProfile(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid provider ID format",
			},
		})
	}

	profile, err := h.service.GetPublicProfile(c.UserContext(), id)
	if err != nil {
		log.Error().Err(err).Str("provider_id", idStr).Msg("failed to get provider profile")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve provider profile",
			},
		})
	}

	if profile == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "provider not found",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": profile,
	})
}

// GetMyProfile returns the authenticated provider's own profile.
// GET /api/v1/providers/me
func (h *ProviderHandler) GetMyProfile(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	profile, err := h.service.GetProfile(c.UserContext(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get own provider profile")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve provider profile",
			},
		})
	}

	if profile == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "provider profile not found",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": profile,
	})
}

// updateProviderProfileRequest is the payload for PUT /api/v1/providers/me.
type updateProviderProfileRequest struct {
	Skills          []string         `json:"skills"`
	ServiceRadiusKM *float64         `json:"service_radius_km"`
	Postcode        *string          `json:"postcode"`
	Latitude        *float64         `json:"latitude"`
	Longitude       *float64         `json:"longitude"`
	Bio             *string          `json:"bio"`
	Availability    *json.RawMessage `json:"availability_schedule"`
}

// UpdateProviderProfile updates the authenticated provider's profile.
// PUT /api/v1/providers/me
func (h *ProviderHandler) UpdateProviderProfile(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req updateProviderProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	// Fetch existing profile to apply partial updates.
	profile, err := h.service.GetProfile(c.UserContext(), userID)
	if err != nil || profile == nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get provider profile for update")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "provider profile not found",
			},
		})
	}

	if req.Skills != nil {
		profile.Skills = req.Skills
	}
	if req.ServiceRadiusKM != nil {
		profile.ServiceRadiusKM = *req.ServiceRadiusKM
	}
	if req.Postcode != nil {
		profile.Postcode = *req.Postcode
	}
	if req.Latitude != nil {
		profile.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		profile.Longitude = *req.Longitude
	}
	if req.Availability != nil {
		profile.AvailabilitySchedule = *req.Availability
	}
	profile.UpdatedAt = time.Now().UTC()

	if err := h.service.UpdateProfile(c.UserContext(), profile); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to update provider profile")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to update provider profile",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": profile,
	})
}

// GetProviderDashboard returns the provider's dashboard summary.
// GET /api/v1/providers/me/dashboard
func (h *ProviderHandler) GetProviderDashboard(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	dashboard, err := h.service.GetDashboard(c.UserContext(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get provider dashboard")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve dashboard",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": dashboard,
	})
}

// GetProviderEarnings returns the provider's earnings breakdown.
// GET /api/v1/providers/me/earnings?period=daily|weekly|monthly
func (h *ProviderHandler) GetProviderEarnings(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	period := c.Query("period", "monthly")
	validPeriods := map[string]bool{"daily": true, "weekly": true, "monthly": true}
	if !validPeriods[period] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "period must be one of: daily, weekly, monthly",
			},
		})
	}

	earnings, err := h.service.GetEarnings(c.UserContext(), userID, period)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get provider earnings")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve earnings",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": earnings,
	})
}

// updateAvailabilityRequest is the payload for PATCH /api/v1/providers/me/availability.
type updateAvailabilityRequest struct {
	Online   *bool            `json:"online"`
	Schedule *json.RawMessage `json:"schedule"`
}

// UpdateAvailability toggles the provider online/offline and sets their schedule.
// PATCH /api/v1/providers/me/availability
func (h *ProviderHandler) UpdateAvailability(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req updateAvailabilityRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	var schedule json.RawMessage
	if req.Schedule != nil {
		schedule = *req.Schedule
	} else if req.Online != nil {
		// Build a simple schedule indicating online/offline status.
		s, _ := json.Marshal(map[string]interface{}{"online": *req.Online})
		schedule = s
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "either online or schedule must be provided",
			},
		})
	}

	if err := h.service.UpdateAvailability(c.UserContext(), userID, schedule); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to update availability")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to update availability",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "availability updated successfully",
		},
	})
}

// UploadKYCDocument handles multipart file upload for KYC verification.
// POST /api/v1/providers/me/kyc
func (h *ProviderHandler) UploadKYCDocument(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	docType := c.FormValue("type")
	if docType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "document type is required (e.g., aadhaar, pan, license)",
			},
		})
	}

	file, err := c.FormFile("document")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "document file is required",
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

	// Open the uploaded file for reading.
	src, err := file.Open()
	if err != nil {
		log.Error().Err(err).Msg("failed to open uploaded file")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to process uploaded file",
			},
		})
	}
	defer src.Close()

	// Detect content type from the file extension, falling back to
	// application/octet-stream if unknown.
	ext := filepath.Ext(file.Filename)
	contentType := stdmime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Generate a unique storage key to avoid filename collisions.
	uniqueName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	key := storage.KYCKey(userID.String(), uniqueName)

	fileURL, err := h.storage.Upload(c.UserContext(), key, src, contentType)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Str("key", key).Msg("failed to upload KYC document to storage")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "STORAGE_ERROR",
				"message": "failed to upload document to storage",
			},
		})
	}

	doc := &KYCDocument{
		ID:         uuid.New(),
		ProviderID: userID,
		Type:       docType,
		FileURL:    fileURL,
		Status:     "pending",
		CreatedAt:  time.Now().UTC(),
	}

	if err := h.service.UploadKYCDocument(c.UserContext(), doc); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to upload KYC document")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to upload KYC document",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": doc,
	})
}
