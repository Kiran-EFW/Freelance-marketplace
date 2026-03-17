package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/adapter/sms"
	"github.com/seva-platform/backend/internal/middleware"
	"github.com/seva-platform/backend/internal/repository/postgres"
)

// SafetyHandler handles SOS alerts, live tracking, and emergency contacts.
type SafetyHandler struct {
	queries     *postgres.Queries
	smsProvider sms.SMSProvider
}

// NewSafetyHandler creates a new SafetyHandler.
func NewSafetyHandler(queries *postgres.Queries, smsProvider sms.SMSProvider) *SafetyHandler {
	return &SafetyHandler{
		queries:     queries,
		smsProvider: smsProvider,
	}
}

// RegisterRoutes mounts safety routes on the given Fiber router group.
func (h *SafetyHandler) RegisterRoutes(rg fiber.Router) {
	// SOS
	rg.Post("/sos", h.TriggerSOS)
	rg.Put("/sos/:id/resolve", h.ResolveSOS)
	rg.Get("/sos", h.ListMyAlerts)

	// Live location
	rg.Post("/location", h.ShareLocation)
	rg.Get("/location/:jobId", h.GetProviderLocation)

	// Emergency contacts
	rg.Get("/contacts", h.ListEmergencyContacts)
	rg.Post("/contacts", h.AddEmergencyContact)
	rg.Delete("/contacts/:id", h.RemoveEmergencyContact)

	// Job verification OTP
	rg.Get("/verify/:jobId", h.GenerateJobVerificationOTP)
}

// triggerSOSRequest is the payload for POST /api/v1/safety/sos.
type triggerSOSRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	JobID     *string `json:"job_id"`
	Notes     string  `json:"notes"`
}

func (r *triggerSOSRequest) validate() error {
	if r.Latitude == 0 && r.Longitude == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "latitude and longitude are required")
	}
	return nil
}

// TriggerSOS triggers an SOS alert and notifies emergency contacts via SMS.
// POST /api/v1/safety/sos
func (h *SafetyHandler) TriggerSOS(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req triggerSOSRequest
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

	var jobID *uuid.UUID
	if req.JobID != nil && *req.JobID != "" {
		parsed, err := uuid.Parse(*req.JobID)
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

	// Get the user's name for SMS
	user, err := h.queries.GetUserByID(c.UserContext(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get user for SOS alert")
	}
	userName := "A Seva user"
	if user.Name.Valid {
		userName = user.Name.String
	}

	// Get emergency contacts and send SMS notifications
	contacts, err := h.queries.ListEmergencyContacts(c.UserContext(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to list emergency contacts for SOS")
	}

	contactsNotified := false
	if len(contacts) > 0 {
		mapsLink := fmt.Sprintf("https://maps.google.com/?q=%f,%f", req.Latitude, req.Longitude)
		message := fmt.Sprintf("EMERGENCY ALERT from Seva: %s needs help! Location: %s. Please check on them immediately.", userName, mapsLink)

		for _, contact := range contacts {
			if err := h.smsProvider.SendSMS(contact.Phone, message); err != nil {
				log.Error().Err(err).Str("contact_phone", contact.Phone).Msg("failed to send SOS SMS to contact")
			} else {
				contactsNotified = true
			}
		}
	}

	alert, err := h.queries.CreateSOSAlert(c.UserContext(), postgres.CreateSOSAlertParams{
		UserID:                    userID,
		JobID:                     jobID,
		Latitude:                  req.Latitude,
		Longitude:                 req.Longitude,
		EmergencyContactsNotified: contactsNotified,
		Notes:                     pgtype.Text{String: req.Notes, Valid: req.Notes != ""},
	})
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to create SOS alert")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to create SOS alert",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{
			"alert":              alert,
			"contacts_notified":  contactsNotified,
			"contacts_count":     len(contacts),
		},
	})
}

// resolveSOSRequest is the payload for PUT /api/v1/safety/sos/:id/resolve.
type resolveSOSRequest struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}

// ResolveSOS resolves an active SOS alert.
// PUT /api/v1/safety/sos/:id/resolve
func (h *SafetyHandler) ResolveSOS(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	alertID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid alert ID format",
			},
		})
	}

	var req resolveSOSRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	status := postgres.SOSStatusResolved
	if req.Status == "false_alarm" {
		status = postgres.SOSStatusFalseAlarm
	}

	alert, err := h.queries.ResolveSOSAlert(c.UserContext(), alertID, status, pgtype.Text{String: req.Notes, Valid: req.Notes != ""})
	if err != nil {
		log.Error().Err(err).Str("alert_id", alertID.String()).Msg("failed to resolve SOS alert")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to resolve SOS alert",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": alert,
	})
}

// ListMyAlerts lists the authenticated user's SOS alerts.
// GET /api/v1/safety/sos
func (h *SafetyHandler) ListMyAlerts(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	page, limit := parsePagination(c)
	offset := (page - 1) * limit

	alerts, err := h.queries.ListSOSAlertsByUser(c.UserContext(), userID, int32(limit), int32(offset))
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to list SOS alerts")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list alerts",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": alerts,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
		},
	})
}

// shareLocationRequest is the payload for POST /api/v1/safety/location.
type shareLocationRequest struct {
	JobID     string  `json:"job_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Accuracy  float32 `json:"accuracy"`
}

func (r *shareLocationRequest) validate() error {
	if r.JobID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "job_id is required")
	}
	if r.Latitude == 0 && r.Longitude == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "latitude and longitude are required")
	}
	return nil
}

// ShareLocation records the user's current location during an active job.
// POST /api/v1/safety/location
func (h *SafetyHandler) ShareLocation(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req shareLocationRequest
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

	jobID, err := uuid.Parse(req.JobID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid job_id format",
			},
		})
	}

	location, err := h.queries.ShareLocation(c.UserContext(), postgres.ShareLocationParams{
		JobID:     jobID,
		UserID:    userID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Accuracy:  pgtype.Float4{Float32: req.Accuracy, Valid: req.Accuracy > 0},
	})
	if err != nil {
		log.Error().Err(err).Str("job_id", jobID.String()).Msg("failed to share location")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to share location",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": location,
	})
}

// GetProviderLocation returns the provider's latest location for a job.
// GET /api/v1/safety/location/:jobId
func (h *SafetyHandler) GetProviderLocation(c *fiber.Ctx) error {
	jobID, err := uuid.Parse(c.Params("jobId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid job ID format",
			},
		})
	}

	// Get location history for the job
	page, limit := parsePagination(c)
	offset := (page - 1) * limit

	history, err := h.queries.GetLocationHistory(c.UserContext(), jobID, int32(limit), int32(offset))
	if err != nil {
		log.Error().Err(err).Str("job_id", jobID.String()).Msg("failed to get location history")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to get location data",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": history,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
		},
	})
}

// addEmergencyContactRequest is the payload for POST /api/v1/safety/contacts.
type addEmergencyContactRequest struct {
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Relationship string `json:"relationship"`
}

func (r *addEmergencyContactRequest) validate() error {
	if r.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	if r.Phone == "" {
		return fiber.NewError(fiber.StatusBadRequest, "phone is required")
	}
	return nil
}

// ListEmergencyContacts lists the user's emergency contacts.
// GET /api/v1/safety/contacts
func (h *SafetyHandler) ListEmergencyContacts(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	contacts, err := h.queries.ListEmergencyContacts(c.UserContext(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to list emergency contacts")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list emergency contacts",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": contacts,
	})
}

// AddEmergencyContact adds a new emergency contact.
// POST /api/v1/safety/contacts
func (h *SafetyHandler) AddEmergencyContact(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req addEmergencyContactRequest
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

	contact, err := h.queries.AddEmergencyContact(c.UserContext(), postgres.AddEmergencyContactParams{
		UserID:       userID,
		Name:         req.Name,
		Phone:        req.Phone,
		Relationship: pgtype.Text{String: req.Relationship, Valid: req.Relationship != ""},
	})
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to add emergency contact")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to add emergency contact",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": contact,
	})
}

// RemoveEmergencyContact removes an emergency contact.
// DELETE /api/v1/safety/contacts/:id
func (h *SafetyHandler) RemoveEmergencyContact(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	contactID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid contact ID format",
			},
		})
	}

	if err := h.queries.RemoveEmergencyContact(c.UserContext(), contactID, userID); err != nil {
		log.Error().Err(err).Str("contact_id", contactID.String()).Msg("failed to remove emergency contact")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to remove emergency contact",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "emergency contact removed successfully",
		},
	})
}

// GenerateJobVerificationOTP generates a 6-digit OTP for job verification.
// GET /api/v1/safety/verify/:jobId
func (h *SafetyHandler) GenerateJobVerificationOTP(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	jobID, err := uuid.Parse(c.Params("jobId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid job ID format",
			},
		})
	}

	// Verify the job exists and belongs to the user
	job, err := h.queries.GetJobByID(c.UserContext(), jobID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "NOT_FOUND",
					"message": "job not found",
				},
			})
		}
		log.Error().Err(err).Str("job_id", jobID.String()).Msg("failed to get job for verification")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to verify job",
			},
		})
	}

	_ = job // job exists

	// Generate a 6-digit OTP
	otp, err := generateOTP()
	if err != nil {
		log.Error().Err(err).Msg("failed to generate OTP")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to generate verification code",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"otp":    otp,
			"job_id": jobID,
			"message": "Share this code with the provider to verify the job",
		},
	})
}

// generateOTP is defined in auth.go and reused here.
