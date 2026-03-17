package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/domain"
	"github.com/seva-platform/backend/internal/middleware"
)

// DashboardStats holds aggregate statistics for the admin dashboard.
type DashboardStats struct {
	TotalUsers       int     `json:"total_users"`
	TotalProviders   int     `json:"total_providers"`
	TotalCustomers   int     `json:"total_customers"`
	TotalJobs        int     `json:"total_jobs"`
	ActiveJobs       int     `json:"active_jobs"`
	CompletedJobs    int     `json:"completed_jobs"`
	TotalRevenue     float64 `json:"total_revenue"`
	ActiveDisputes   int     `json:"active_disputes"`
	PendingKYC       int     `json:"pending_kyc"`
}

// KYCEntry represents a KYC verification entry in the admin queue.
type KYCEntry struct {
	ID           uuid.UUID `json:"id"`
	ProviderID   uuid.UUID `json:"provider_id"`
	ProviderName string    `json:"provider_name"`
	DocumentType string    `json:"document_type"`
	FileURL      string    `json:"file_url"`
	Status       string    `json:"status"`
	SubmittedAt  time.Time `json:"submitted_at"`
}

// AnalyticsData holds time-series data for admin analytics charts.
type AnalyticsData struct {
	Date       string  `json:"date"`
	JobsCount  int     `json:"jobs_count"`
	Revenue    float64 `json:"revenue"`
	Signups    int     `json:"signups"`
}

// AdminService defines the business operations required by AdminHandler.
type AdminService interface {
	GetDashboardStats(ctx context.Context) (*DashboardStats, error)
	ListUsers(ctx context.Context, userType *string, status *string, limit, offset int) ([]domain.User, int, error)
	ListPendingKYC(ctx context.Context, limit, offset int) ([]KYCEntry, int, error)
	ApproveKYC(ctx context.Context, kycID, adminID uuid.UUID) error
	RejectKYC(ctx context.Context, kycID, adminID uuid.UUID, reason string) error
	ListDisputes(ctx context.Context, status *string, limit, offset int) ([]Dispute, int, error)
	GetAnalytics(ctx context.Context, from, to string) ([]AnalyticsData, error)
	CreateCategory(ctx context.Context, category *domain.Category) error
	UpdateCategory(ctx context.Context, category *domain.Category) error
	SuspendUser(ctx context.Context, userID, adminID uuid.UUID, reason string) error
}

// AdminHandler handles admin dashboard endpoints.
type AdminHandler struct {
	service AdminService
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(svc AdminService) *AdminHandler {
	return &AdminHandler{service: svc}
}

// RegisterRoutes mounts admin routes on the given Fiber router group.
func (h *AdminHandler) RegisterRoutes(rg fiber.Router) {
	rg.Get("/stats", h.GetDashboardStats)
	rg.Get("/users", h.ListUsers)
	rg.Post("/users/:id/suspend", h.SuspendUser)
	rg.Get("/kyc/pending", h.ListPendingKYC)
	rg.Post("/kyc/:id/approve", h.ApproveKYC)
	rg.Post("/kyc/:id/reject", h.RejectKYC)
	rg.Get("/disputes", h.ListDisputes)
	rg.Get("/analytics", h.GetAnalytics)
	rg.Post("/categories", h.CreateCategory)
	rg.Put("/categories", h.UpdateCategory)
}

// GetDashboardStats returns aggregate platform statistics.
// GET /api/v1/admin/stats
func (h *AdminHandler) GetDashboardStats(c *fiber.Ctx) error {
	stats, err := h.service.GetDashboardStats(c.UserContext())
	if err != nil {
		log.Error().Err(err).Msg("failed to get dashboard stats")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve dashboard statistics",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": stats,
	})
}

// ListUsers returns a paginated list of users with optional filters.
// GET /api/v1/admin/users?type=customer|provider&status=active&page=1&limit=20
func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	page, limit := parsePagination(c)
	offset := (page - 1) * limit

	var userType *string
	if v := c.Query("type"); v != "" {
		userType = &v
	}

	var status *string
	if v := c.Query("status"); v != "" {
		status = &v
	}

	users, total, err := h.service.ListUsers(c.UserContext(), userType, status, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("failed to list users")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list users",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": users,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// ListPendingKYC returns the KYC verification queue.
// GET /api/v1/admin/kyc/pending?page=1&limit=20
func (h *AdminHandler) ListPendingKYC(c *fiber.Ctx) error {
	page, limit := parsePagination(c)
	offset := (page - 1) * limit

	entries, total, err := h.service.ListPendingKYC(c.UserContext(), limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("failed to list pending KYC")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list pending KYC entries",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": entries,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// ApproveKYC approves a KYC verification entry.
// POST /api/v1/admin/kyc/:id/approve
func (h *AdminHandler) ApproveKYC(c *fiber.Ctx) error {
	adminID := middleware.GetUserID(c)
	if adminID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	kycID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid KYC entry ID format",
			},
		})
	}

	if err := h.service.ApproveKYC(c.UserContext(), kycID, adminID); err != nil {
		log.Error().Err(err).Str("kyc_id", kycID.String()).Msg("failed to approve KYC")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to approve KYC",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"kyc_id":  kycID,
			"status":  "approved",
			"message": "KYC approved successfully",
		},
	})
}

// rejectKYCRequest is the payload for POST /api/v1/admin/kyc/:id/reject.
type rejectKYCRequest struct {
	Reason string `json:"reason"`
}

// RejectKYC rejects a KYC verification entry.
// POST /api/v1/admin/kyc/:id/reject
func (h *AdminHandler) RejectKYC(c *fiber.Ctx) error {
	adminID := middleware.GetUserID(c)
	if adminID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	kycID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid KYC entry ID format",
			},
		})
	}

	var req rejectKYCRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	if req.Reason == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "reason is required",
			},
		})
	}

	if err := h.service.RejectKYC(c.UserContext(), kycID, adminID, req.Reason); err != nil {
		log.Error().Err(err).Str("kyc_id", kycID.String()).Msg("failed to reject KYC")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to reject KYC",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"kyc_id":  kycID,
			"status":  "rejected",
			"message": "KYC rejected",
		},
	})
}

// ListDisputes lists all disputes with optional status filter.
// GET /api/v1/admin/disputes?status=open&page=1&limit=20
func (h *AdminHandler) ListDisputes(c *fiber.Ctx) error {
	page, limit := parsePagination(c)
	offset := (page - 1) * limit

	var status *string
	if v := c.Query("status"); v != "" {
		status = &v
	}

	disputes, total, err := h.service.ListDisputes(c.UserContext(), status, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("failed to list disputes")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list disputes",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": disputes,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetAnalytics returns time-series analytics data.
// GET /api/v1/admin/analytics?from=2024-01-01&to=2024-12-31
func (h *AdminHandler) GetAnalytics(c *fiber.Ctx) error {
	from := c.Query("from", "")
	to := c.Query("to", "")

	analytics, err := h.service.GetAnalytics(c.UserContext(), from, to)
	if err != nil {
		log.Error().Err(err).Msg("failed to get analytics")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve analytics",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": analytics,
	})
}

// createCategoryRequest is the payload for POST /api/v1/admin/categories.
type createCategoryRequest struct {
	Slug            string            `json:"slug"`
	Name            map[string]string `json:"name"`
	ParentID        *string           `json:"parent_id"`
	Icon            string            `json:"icon"`
	RequiresLicense bool              `json:"requires_license"`
}

func (r *createCategoryRequest) validate() error {
	if r.Slug == "" {
		return fiber.NewError(fiber.StatusBadRequest, "slug is required")
	}
	if len(r.Name) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "name is required (at least one language)")
	}
	return nil
}

// CreateCategory creates a new service category.
// POST /api/v1/admin/categories
func (h *AdminHandler) CreateCategory(c *fiber.Ctx) error {
	var req createCategoryRequest
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

	var parentID *uuid.UUID
	if req.ParentID != nil {
		id, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "INVALID_ID",
					"message": "invalid parent_id format",
				},
			})
		}
		parentID = &id
	}

	category := &domain.Category{
		ID:              uuid.New(),
		Slug:            req.Slug,
		Name:            req.Name,
		ParentID:        parentID,
		Icon:            req.Icon,
		IsActive:        true,
		RequiresLicense: req.RequiresLicense,
	}

	if err := h.service.CreateCategory(c.UserContext(), category); err != nil {
		log.Error().Err(err).Str("slug", req.Slug).Msg("failed to create category")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to create category",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": category,
	})
}

// updateCategoryRequest is the payload for PUT /api/v1/admin/categories.
type updateCategoryRequest struct {
	ID              string            `json:"id"`
	Slug            *string           `json:"slug"`
	Name            map[string]string `json:"name"`
	Icon            *string           `json:"icon"`
	IsActive        *bool             `json:"is_active"`
	RequiresLicense *bool             `json:"requires_license"`
}

// UpdateCategory updates an existing service category.
// PUT /api/v1/admin/categories
func (h *AdminHandler) UpdateCategory(c *fiber.Ctx) error {
	var req updateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	if req.ID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "id is required",
			},
		})
	}

	catID, err := uuid.Parse(req.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid category id format",
			},
		})
	}

	category := &domain.Category{
		ID:   catID,
		Name: req.Name,
	}

	if req.Slug != nil {
		category.Slug = *req.Slug
	}
	if req.Icon != nil {
		category.Icon = *req.Icon
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}
	if req.RequiresLicense != nil {
		category.RequiresLicense = *req.RequiresLicense
	}

	if err := h.service.UpdateCategory(c.UserContext(), category); err != nil {
		log.Error().Err(err).Str("category_id", catID.String()).Msg("failed to update category")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to update category",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": category,
	})
}

// suspendUserRequest is the payload for POST /api/v1/admin/users/:id/suspend.
type suspendUserRequest struct {
	Reason string `json:"reason"`
}

// SuspendUser suspends a user account.
// POST /api/v1/admin/users/:id/suspend
func (h *AdminHandler) SuspendUser(c *fiber.Ctx) error {
	adminID := middleware.GetUserID(c)
	if adminID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid user ID format",
			},
		})
	}

	var req suspendUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	if req.Reason == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "reason is required",
			},
		})
	}

	if err := h.service.SuspendUser(c.UserContext(), userID, adminID, req.Reason); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to suspend user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to suspend user",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"user_id": userID,
			"status":  "suspended",
			"message": "user suspended successfully",
		},
	})
}
