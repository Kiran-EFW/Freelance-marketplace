package handler

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/middleware"
	"github.com/seva-platform/backend/internal/repository/postgres"
)

// OrganizationHandler handles B2B organization endpoints.
type OrganizationHandler struct {
	queries *postgres.Queries
}

// NewOrganizationHandler creates a new OrganizationHandler.
func NewOrganizationHandler(queries *postgres.Queries) *OrganizationHandler {
	return &OrganizationHandler{queries: queries}
}

// RegisterRoutes mounts organization routes on the given Fiber router group.
func (h *OrganizationHandler) RegisterRoutes(rg fiber.Router) {
	rg.Post("/", h.CreateOrganization)
	rg.Get("/:id", h.GetOrganization)
	rg.Post("/:id/members", h.requireOrgAdmin, h.AddMember)
	rg.Get("/:id/members", h.requireOrgMember, h.ListMembers)
	rg.Delete("/:id/members/:userId", h.requireOrgAdmin, h.RemoveMember)
	rg.Post("/:id/requests", h.requireOrgMember, h.CreateServiceRequest)
	rg.Get("/:id/requests", h.requireOrgMember, h.ListServiceRequests)
	rg.Put("/:id/requests/:reqId/assign", h.requireOrgAdmin, h.AssignProvider)
	rg.Put("/:id/requests/:reqId/status", h.requireOrgAdmin, h.UpdateRequestStatus)
	rg.Get("/:id/stats", h.requireOrgMember, h.GetOrgStats)
}

// requireOrgMember is middleware that checks the user is a member of the organization.
func (h *OrganizationHandler) requireOrgMember(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid organization ID format",
			},
		})
	}

	memberRole, err := h.queries.GetMemberRole(c.UserContext(), orgID, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "FORBIDDEN",
					"message": "you are not a member of this organization",
				},
			})
		}
		log.Error().Err(err).Str("org_id", orgID.String()).Str("user_id", userID.String()).Msg("failed to check membership")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to verify membership",
			},
		})
	}

	if memberRole.Status != postgres.MemberStatusActive {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "FORBIDDEN",
				"message": "your membership is not active",
			},
		})
	}

	c.Locals("org_role", string(memberRole.Role))
	return c.Next()
}

// requireOrgAdmin is middleware that checks the user is an admin of the organization.
func (h *OrganizationHandler) requireOrgAdmin(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid organization ID format",
			},
		})
	}

	memberRole, err := h.queries.GetMemberRole(c.UserContext(), orgID, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "FORBIDDEN",
					"message": "you are not a member of this organization",
				},
			})
		}
		log.Error().Err(err).Str("org_id", orgID.String()).Str("user_id", userID.String()).Msg("failed to check admin role")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to verify admin role",
			},
		})
	}

	if memberRole.Role != postgres.OrgRoleAdmin && memberRole.Role != postgres.OrgRoleManager {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "FORBIDDEN",
				"message": "admin or manager role required",
			},
		})
	}

	c.Locals("org_role", string(memberRole.Role))
	return c.Next()
}

// createOrganizationRequest is the payload for POST /api/v1/organizations.
type createOrganizationRequest struct {
	Name         string          `json:"name"`
	Type         string          `json:"type"`
	Address      string          `json:"address"`
	Postcode     string          `json:"postcode"`
	City         string          `json:"city"`
	State        string          `json:"state"`
	Country      string          `json:"country"`
	ContactPhone string          `json:"contact_phone"`
	ContactEmail string          `json:"contact_email"`
	Settings     json.RawMessage `json:"settings"`
}

func (r *createOrganizationRequest) validate() error {
	if r.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	validTypes := map[string]bool{
		"housing_society": true,
		"company":         true,
		"institution":     true,
	}
	if !validTypes[r.Type] {
		return fiber.NewError(fiber.StatusBadRequest, "type must be one of: housing_society, company, institution")
	}
	return nil
}

// CreateOrganization creates a new organization and sets the creating user as admin.
// POST /api/v1/organizations
func (h *OrganizationHandler) CreateOrganization(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req createOrganizationRequest
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

	settings := req.Settings
	if settings == nil {
		settings = json.RawMessage(`{}`)
	}

	country := req.Country
	if country == "" {
		country = "India"
	}

	org, err := h.queries.CreateOrganization(c.UserContext(), postgres.CreateOrganizationParams{
		Name:         req.Name,
		Type:         postgres.OrgType(req.Type),
		Address:      pgtype.Text{String: req.Address, Valid: req.Address != ""},
		Postcode:     pgtype.Text{String: req.Postcode, Valid: req.Postcode != ""},
		City:         pgtype.Text{String: req.City, Valid: req.City != ""},
		State:        pgtype.Text{String: req.State, Valid: req.State != ""},
		Country:      country,
		ContactPhone: pgtype.Text{String: req.ContactPhone, Valid: req.ContactPhone != ""},
		ContactEmail: pgtype.Text{String: req.ContactEmail, Valid: req.ContactEmail != ""},
		LogoUrl:      pgtype.Text{},
		Settings:     settings,
		Status:       postgres.OrgStatusActive,
	})
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to create organization")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to create organization",
			},
		})
	}

	// Add the creator as admin
	_, err = h.queries.AddOrganizationMember(c.UserContext(), postgres.AddOrganizationMemberParams{
		OrgID:  org.ID,
		UserID: userID,
		Role:   postgres.OrgRoleAdmin,
		Status: postgres.MemberStatusActive,
	})
	if err != nil {
		log.Error().Err(err).Str("org_id", org.ID.String()).Str("user_id", userID.String()).Msg("failed to add creator as admin")
		// The org was created; log the error but still return success
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": org,
	})
}

// GetOrganization returns a single organization by ID.
// GET /api/v1/organizations/:id
func (h *OrganizationHandler) GetOrganization(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid organization ID format",
			},
		})
	}

	org, err := h.queries.GetOrganizationByID(c.UserContext(), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "NOT_FOUND",
					"message": "organization not found",
				},
			})
		}
		log.Error().Err(err).Str("org_id", id.String()).Msg("failed to get organization")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve organization",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": org,
	})
}

// addMemberRequest is the payload for POST /api/v1/organizations/:id/members.
type addMemberRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

func (r *addMemberRequest) validate() error {
	if r.UserID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "user_id is required")
	}
	validRoles := map[string]bool{
		"admin":   true,
		"manager": true,
		"member":  true,
	}
	if r.Role != "" && !validRoles[r.Role] {
		return fiber.NewError(fiber.StatusBadRequest, "role must be one of: admin, manager, member")
	}
	return nil
}

// AddMember adds a member to the organization.
// POST /api/v1/organizations/:id/members
func (h *OrganizationHandler) AddMember(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid organization ID format",
			},
		})
	}

	var req addMemberRequest
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

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid user_id format",
			},
		})
	}

	role := postgres.OrgRoleMember
	if req.Role != "" {
		role = postgres.OrgRole(req.Role)
	}

	member, err := h.queries.AddOrganizationMember(c.UserContext(), postgres.AddOrganizationMemberParams{
		OrgID:  orgID,
		UserID: userID,
		Role:   role,
		Status: postgres.MemberStatusInvited,
	})
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID.String()).Str("user_id", userID.String()).Msg("failed to add member")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to add member",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": member,
	})
}

// ListMembers lists all members of an organization.
// GET /api/v1/organizations/:id/members
func (h *OrganizationHandler) ListMembers(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid organization ID format",
			},
		})
	}

	page, limit := parsePagination(c)
	offset := (page - 1) * limit

	members, err := h.queries.ListOrganizationMembers(c.UserContext(), orgID, int32(limit), int32(offset))
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID.String()).Msg("failed to list members")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list members",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": members,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
		},
	})
}

// RemoveMember removes a member from the organization.
// DELETE /api/v1/organizations/:id/members/:userId
func (h *OrganizationHandler) RemoveMember(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid organization ID format",
			},
		})
	}

	memberUserID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid user ID format",
			},
		})
	}

	if err := h.queries.RemoveOrganizationMember(c.UserContext(), orgID, memberUserID); err != nil {
		log.Error().Err(err).Str("org_id", orgID.String()).Str("user_id", memberUserID.String()).Msg("failed to remove member")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to remove member",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "member removed successfully",
		},
	})
}

// createServiceRequestRequest is the payload for POST /api/v1/organizations/:id/requests.
type createServiceRequestRequest struct {
	CategoryID  string  `json:"category_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Priority    string  `json:"priority"`
	ScheduledAt *string `json:"scheduled_at"`
	Notes       string  `json:"notes"`
}

func (r *createServiceRequestRequest) validate() error {
	if r.CategoryID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "category_id is required")
	}
	if r.Title == "" {
		return fiber.NewError(fiber.StatusBadRequest, "title is required")
	}
	validPriorities := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
		"urgent": true,
	}
	if r.Priority != "" && !validPriorities[r.Priority] {
		return fiber.NewError(fiber.StatusBadRequest, "priority must be one of: low, medium, high, urgent")
	}
	return nil
}

// CreateServiceRequest creates a service request for the organization.
// POST /api/v1/organizations/:id/requests
func (h *OrganizationHandler) CreateServiceRequest(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid organization ID format",
			},
		})
	}

	var req createServiceRequestRequest
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

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid category_id format",
			},
		})
	}

	priority := postgres.RequestPriorityMedium
	if req.Priority != "" {
		priority = postgres.RequestPriority(req.Priority)
	}

	var scheduledAt pgtype.Timestamptz
	if req.ScheduledAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ScheduledAt)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "INVALID_DATE",
					"message": "scheduled_at must be RFC3339 format",
				},
			})
		}
		scheduledAt = pgtype.Timestamptz{Time: t, Valid: true}
	}

	serviceReq, err := h.queries.CreateOrganizationServiceRequest(c.UserContext(), postgres.CreateOrganizationServiceRequestParams{
		OrgID:       orgID,
		RequestedBy: userID,
		CategoryID:  categoryID,
		Title:       req.Title,
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
		Priority:    priority,
		ScheduledAt: scheduledAt,
		Notes:       pgtype.Text{String: req.Notes, Valid: req.Notes != ""},
	})
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID.String()).Msg("failed to create service request")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to create service request",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": serviceReq,
	})
}

// ListServiceRequests lists service requests for an organization.
// GET /api/v1/organizations/:id/requests?status=pending&priority=high&page=1&limit=20
func (h *OrganizationHandler) ListServiceRequests(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid organization ID format",
			},
		})
	}

	page, limit := parsePagination(c)
	offset := (page - 1) * limit

	statusFilter := pgtype.Text{}
	if s := c.Query("status"); s != "" {
		statusFilter = pgtype.Text{String: s, Valid: true}
	}

	priorityFilter := pgtype.Text{}
	if p := c.Query("priority"); p != "" {
		priorityFilter = pgtype.Text{String: p, Valid: true}
	}

	requests, err := h.queries.ListOrganizationServiceRequests(c.UserContext(), postgres.ListOrganizationServiceRequestsParams{
		OrgID:    orgID,
		Limit:    int32(limit),
		Offset:   int32(offset),
		Status:   statusFilter,
		Priority: priorityFilter,
	})
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID.String()).Msg("failed to list service requests")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list service requests",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": requests,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
		},
	})
}

// assignProviderRequest is the payload for PUT /api/v1/organizations/:id/requests/:reqId/assign.
type assignProviderRequest struct {
	ProviderID string `json:"provider_id"`
}

// AssignProvider assigns a provider to a service request.
// PUT /api/v1/organizations/:id/requests/:reqId/assign
func (h *OrganizationHandler) AssignProvider(c *fiber.Ctx) error {
	reqID, err := uuid.Parse(c.Params("reqId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid request ID format",
			},
		})
	}

	var req assignProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	if req.ProviderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "provider_id is required",
			},
		})
	}

	providerID, err := uuid.Parse(req.ProviderID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid provider_id format",
			},
		})
	}

	updated, err := h.queries.AssignOrganizationServiceRequestProvider(c.UserContext(), reqID, providerID)
	if err != nil {
		log.Error().Err(err).Str("request_id", reqID.String()).Msg("failed to assign provider")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to assign provider",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": updated,
	})
}

// updateRequestStatusRequest is the payload for PUT /api/v1/organizations/:id/requests/:reqId/status.
type updateRequestStatusRequest struct {
	Status string `json:"status"`
}

// UpdateRequestStatus updates the status of a service request.
// PUT /api/v1/organizations/:id/requests/:reqId/status
func (h *OrganizationHandler) UpdateRequestStatus(c *fiber.Ctx) error {
	reqID, err := uuid.Parse(c.Params("reqId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid request ID format",
			},
		})
	}

	var req updateRequestStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	validStatuses := map[string]bool{
		"pending":     true,
		"assigned":    true,
		"in_progress": true,
		"completed":   true,
		"cancelled":   true,
	}
	if !validStatuses[req.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "status must be one of: pending, assigned, in_progress, completed, cancelled",
			},
		})
	}

	updated, err := h.queries.UpdateOrganizationServiceRequestStatus(c.UserContext(), reqID, postgres.RequestStatus(req.Status))
	if err != nil {
		log.Error().Err(err).Str("request_id", reqID.String()).Msg("failed to update request status")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to update request status",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": updated,
	})
}

// GetOrgStats returns organization dashboard statistics.
// GET /api/v1/organizations/:id/stats
func (h *OrganizationHandler) GetOrgStats(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid organization ID format",
			},
		})
	}

	pendingCount, err := h.queries.CountOrganizationServiceRequestsByStatus(c.UserContext(), orgID, postgres.RequestStatusPending)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID.String()).Msg("failed to count pending requests")
		pendingCount = 0
	}

	completedCount, err := h.queries.CountOrganizationServiceRequestsByStatus(c.UserContext(), orgID, postgres.RequestStatusCompleted)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID.String()).Msg("failed to count completed requests")
		completedCount = 0
	}

	inProgressCount, err := h.queries.CountOrganizationServiceRequestsByStatus(c.UserContext(), orgID, postgres.RequestStatusInProgress)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID.String()).Msg("failed to count in-progress requests")
		inProgressCount = 0
	}

	assignedCount, err := h.queries.CountOrganizationServiceRequestsByStatus(c.UserContext(), orgID, postgres.RequestStatusAssigned)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID.String()).Msg("failed to count assigned requests")
		assignedCount = 0
	}

	memberCount, err := h.queries.CountOrganizationMembers(c.UserContext(), orgID)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID.String()).Msg("failed to count members")
		memberCount = 0
	}

	totalRequests := pendingCount + completedCount + inProgressCount + assignedCount

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"total_requests":    totalRequests,
			"pending_requests":  pendingCount,
			"completed_requests": completedCount,
			"in_progress_requests": inProgressCount,
			"assigned_requests": assignedCount,
			"active_members":    memberCount,
		},
	})
}
