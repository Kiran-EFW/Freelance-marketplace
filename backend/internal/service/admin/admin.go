// Package admin provides the admin service implementation with real database
// queries for platform management: user listing, KYC verification,
// dispute management, analytics, and category CRUD.
package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/domain"
	"github.com/seva-platform/backend/internal/repository/postgres"
)

// AdminService implements platform administration operations using direct
// database queries.
type AdminService struct {
	queries    *postgres.Queries
	db         *pgxpool.Pool
	disputeRepo domain.DisputeRepository
}

// NewAdminService returns a ready-to-use AdminService.
func NewAdminService(queries *postgres.Queries, db *pgxpool.Pool, disputeRepo domain.DisputeRepository) *AdminService {
	return &AdminService{queries: queries, db: db, disputeRepo: disputeRepo}
}

// DashboardStats holds aggregate statistics for the admin dashboard.
type DashboardStats struct {
	TotalUsers     int     `json:"total_users"`
	TotalProviders int     `json:"total_providers"`
	TotalCustomers int     `json:"total_customers"`
	TotalJobs      int     `json:"total_jobs"`
	ActiveJobs     int     `json:"active_jobs"`
	CompletedJobs  int     `json:"completed_jobs"`
	TotalRevenue   float64 `json:"total_revenue"`
	ActiveDisputes int     `json:"active_disputes"`
	PendingKYC     int     `json:"pending_kyc"`
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
	Date      string  `json:"date"`
	JobsCount int     `json:"jobs_count"`
	Revenue   float64 `json:"revenue"`
	Signups   int     `json:"signups"`
}

// GetDashboardStats returns aggregate platform statistics by querying
// multiple tables.
func (s *AdminService) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Count users by type.
	customerCount, err := s.queries.CountUsersByType(ctx, postgres.UserTypeCustomer)
	if err != nil {
		log.Warn().Err(err).Msg("failed to count customers")
	}
	stats.TotalCustomers = int(customerCount)

	providerCount, err := s.queries.CountUsersByType(ctx, postgres.UserTypeProvider)
	if err != nil {
		log.Warn().Err(err).Msg("failed to count providers")
	}
	stats.TotalProviders = int(providerCount)
	stats.TotalUsers = stats.TotalCustomers + stats.TotalProviders

	// Count jobs by status using raw queries.
	var totalJobs, activeJobs, completedJobs int64
	row := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM jobs`)
	row.Scan(&totalJobs)
	stats.TotalJobs = int(totalJobs)

	row = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM jobs WHERE status IN ('posted', 'matched', 'accepted', 'in_progress')`)
	row.Scan(&activeJobs)
	stats.ActiveJobs = int(activeJobs)

	row = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM jobs WHERE status = 'completed'`)
	row.Scan(&completedJobs)
	stats.CompletedJobs = int(completedJobs)

	// Total revenue from completed transactions.
	var totalRevenue pgtype.Numeric
	row = s.db.QueryRow(ctx, `SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE payment_status = 'captured'`)
	if err := row.Scan(&totalRevenue); err == nil && totalRevenue.Valid {
		f, _ := totalRevenue.Float64Value()
		if f.Valid {
			stats.TotalRevenue = f.Float64
		}
	}

	// Active disputes.
	var activeDisputes int64
	row = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM disputes WHERE status IN ('open', 'under_review', 'mediation', 'escalated')`)
	row.Scan(&activeDisputes)
	stats.ActiveDisputes = int(activeDisputes)

	// Pending KYC.
	var pendingKYC int64
	row = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM provider_profiles WHERE verification_status = 'pending'`)
	row.Scan(&pendingKYC)
	stats.PendingKYC = int(pendingKYC)

	log.Info().
		Int("total_users", stats.TotalUsers).
		Int("total_jobs", stats.TotalJobs).
		Msg("admin dashboard stats retrieved")

	return stats, nil
}

// ListUsers returns a paginated list of users with optional type and status filters.
func (s *AdminService) ListUsers(ctx context.Context, userType *string, status *string, limit, offset int) ([]domain.User, int, error) {
	query := `SELECT id, type, phone, email, name, jurisdiction_id, preferred_language, device_type, is_active, created_at, updated_at
		FROM users WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if userType != nil && *userType != "" {
		query += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, *userType)
		argIdx++
	}
	if status != nil && *status != "" {
		if *status == "active" {
			query += " AND is_active = true"
		} else if *status == "suspended" || *status == "inactive" {
			query += " AND is_active = false"
		}
	}

	// Count total matching users.
	countQuery := `SELECT COUNT(*) FROM users WHERE 1=1`
	if userType != nil && *userType != "" {
		countQuery += fmt.Sprintf(" AND type = $1")
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		var email pgtype.Text
		var name pgtype.Text
		if err := rows.Scan(
			&u.ID, &u.Type, &u.Phone, &email, &name,
			&u.JurisdictionID, &u.PreferredLanguage, &u.DeviceType,
			new(bool), // is_active
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			log.Warn().Err(err).Msg("failed to scan user row, skipping")
			continue
		}
		if email.Valid {
			u.Email = &email.String
		}
		if name.Valid {
			u.Name = name.String
		}
		users = append(users, u)
	}

	return users, len(users), nil
}

// ListPendingKYC returns provider profiles with pending verification status.
func (s *AdminService) ListPendingKYC(ctx context.Context, limit, offset int) ([]KYCEntry, int, error) {
	query := `SELECT p.id, p.user_id, u.name, p.verification_status, p.created_at
		FROM provider_profiles p
		JOIN users u ON u.id = p.user_id
		WHERE p.verification_status = 'pending'
		ORDER BY p.created_at ASC
		LIMIT $1 OFFSET $2`

	rows, err := s.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list pending KYC: %w", err)
	}
	defer rows.Close()

	var entries []KYCEntry
	for rows.Next() {
		var e KYCEntry
		var name pgtype.Text
		var status string
		if err := rows.Scan(&e.ID, &e.ProviderID, &name, &status, &e.SubmittedAt); err != nil {
			log.Warn().Err(err).Msg("failed to scan KYC row, skipping")
			continue
		}
		if name.Valid {
			e.ProviderName = name.String
		}
		e.Status = status
		e.DocumentType = "identity" // default; would come from kyc_documents table in production
		entries = append(entries, e)
	}

	return entries, len(entries), nil
}

// ApproveKYC approves a provider's KYC verification.
func (s *AdminService) ApproveKYC(ctx context.Context, kycID, adminID uuid.UUID) error {
	err := s.queries.UpdateVerificationStatus(ctx, postgres.UpdateVerificationStatusParams{
		ID:                 kycID,
		VerificationStatus: postgres.VerificationStatusVerified,
	})
	if err != nil {
		return fmt.Errorf("approve KYC: %w", err)
	}

	log.Info().
		Str("kyc_id", kycID.String()).
		Str("admin_id", adminID.String()).
		Msg("KYC approved")

	return nil
}

// RejectKYC rejects a provider's KYC verification.
func (s *AdminService) RejectKYC(ctx context.Context, kycID, adminID uuid.UUID, reason string) error {
	err := s.queries.UpdateVerificationStatus(ctx, postgres.UpdateVerificationStatusParams{
		ID:                 kycID,
		VerificationStatus: postgres.VerificationStatusRejected,
	})
	if err != nil {
		return fmt.Errorf("reject KYC: %w", err)
	}

	log.Info().
		Str("kyc_id", kycID.String()).
		Str("admin_id", adminID.String()).
		Str("reason", reason).
		Msg("KYC rejected")

	return nil
}

// ListDisputes returns disputes with optional status filter. Uses the
// handler.Dispute type to match the handler interface.
func (s *AdminService) ListDisputes(ctx context.Context, status *string, limit, offset int) ([]domain.Dispute, int, error) {
	if status != nil && *status != "" {
		disputes, err := s.disputeRepo.ListByStatus(ctx, domain.DisputeStatus(*status), limit, offset)
		if err != nil {
			return nil, 0, fmt.Errorf("list disputes: %w", err)
		}
		return disputes, len(disputes), nil
	}

	// List all disputes (no status filter) via raw query.
	query := `SELECT id, job_id, raised_by, against, type, severity, status, description,
		evidence, resolution, resolved_by, resolution_amount, escalated_at, resolved_at,
		created_at, updated_at FROM disputes
		ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := s.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list all disputes: %w", err)
	}
	defer rows.Close()

	var disputes []domain.Dispute
	for rows.Next() {
		var d domain.Dispute
		var evidence json.RawMessage
		var resolution pgtype.Text
		var resolutionAmount pgtype.Numeric
		var escalatedAt, resolvedAt pgtype.Timestamptz

		if err := rows.Scan(
			&d.ID, &d.JobID, &d.RaisedBy, &d.Against,
			&d.Type, &d.Severity, &d.Status, &d.Description,
			&evidence, &resolution, &d.ResolvedBy, &resolutionAmount,
			&escalatedAt, &resolvedAt,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			log.Warn().Err(err).Msg("failed to scan dispute row, skipping")
			continue
		}
		d.Evidence = evidence
		if resolution.Valid {
			d.ResolutionNotes = &resolution.String
		}
		if resolutionAmount.Valid {
			f, _ := resolutionAmount.Float64Value()
			if f.Valid {
				v := f.Float64
				d.ResolutionAmount = &v
			}
		}
		if escalatedAt.Valid {
			t := escalatedAt.Time
			d.EscalatedAt = &t
		}
		if resolvedAt.Valid {
			t := resolvedAt.Time
			d.ResolvedAt = &t
		}
		disputes = append(disputes, d)
	}

	return disputes, len(disputes), nil
}

// GetAnalytics returns time-series analytics data for the given date range.
func (s *AdminService) GetAnalytics(ctx context.Context, from, to string) ([]AnalyticsData, error) {
	if from == "" {
		from = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}
	if to == "" {
		to = time.Now().Format("2006-01-02")
	}

	// Job counts per day.
	query := `SELECT DATE(created_at) as day, COUNT(*) as job_count
		FROM jobs
		WHERE DATE(created_at) BETWEEN $1 AND $2
		GROUP BY DATE(created_at)
		ORDER BY day`

	rows, err := s.db.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("get analytics: %w", err)
	}
	defer rows.Close()

	dataMap := make(map[string]*AnalyticsData)
	for rows.Next() {
		var day time.Time
		var count int64
		if err := rows.Scan(&day, &count); err != nil {
			continue
		}
		dateStr := day.Format("2006-01-02")
		dataMap[dateStr] = &AnalyticsData{
			Date:      dateStr,
			JobsCount: int(count),
		}
	}

	// Revenue per day.
	revenueQuery := `SELECT DATE(created_at) as day, COALESCE(SUM(amount), 0) as revenue
		FROM transactions
		WHERE payment_status = 'captured' AND DATE(created_at) BETWEEN $1 AND $2
		GROUP BY DATE(created_at)
		ORDER BY day`

	rows2, err := s.db.Query(ctx, revenueQuery, from, to)
	if err != nil {
		log.Warn().Err(err).Msg("failed to get revenue analytics")
	} else {
		defer rows2.Close()
		for rows2.Next() {
			var day time.Time
			var revenue pgtype.Numeric
			if err := rows2.Scan(&day, &revenue); err != nil {
				continue
			}
			dateStr := day.Format("2006-01-02")
			if _, ok := dataMap[dateStr]; !ok {
				dataMap[dateStr] = &AnalyticsData{Date: dateStr}
			}
			if revenue.Valid {
				f, _ := revenue.Float64Value()
				if f.Valid {
					dataMap[dateStr].Revenue = f.Float64
				}
			}
		}
	}

	// Convert map to sorted slice.
	var result []AnalyticsData
	for _, v := range dataMap {
		result = append(result, *v)
	}

	return result, nil
}

// CreateCategory creates a new service category.
func (s *AdminService) CreateCategory(ctx context.Context, category *domain.Category) error {
	nameJSON, err := json.Marshal(category.Name)
	if err != nil {
		return fmt.Errorf("marshal category name: %w", err)
	}

	_, err = s.queries.CreateCategory(ctx, postgres.CreateCategoryParams{
		Slug:            category.Slug,
		Name:            nameJSON,
		ParentID:        category.ParentID,
		Icon:            pgtype.Text{String: category.Icon, Valid: category.Icon != ""},
		SortOrder:       0,
		IsActive:        category.IsActive,
		RequiresLicense: category.RequiresLicense,
		PricingModel:    postgres.PricingModelFixed,
		Metadata:        []byte("{}"),
	})
	if err != nil {
		return fmt.Errorf("create category: %w", err)
	}

	log.Info().Str("slug", category.Slug).Msg("category created")
	return nil
}

// UpdateCategory updates an existing service category.
func (s *AdminService) UpdateCategory(ctx context.Context, category *domain.Category) error {
	var nameJSON json.RawMessage
	if len(category.Name) > 0 {
		var err error
		nameJSON, err = json.Marshal(category.Name)
		if err != nil {
			return fmt.Errorf("marshal category name: %w", err)
		}
	}

	_, err := s.queries.UpdateCategory(ctx, postgres.UpdateCategoryParams{
		ID:   category.ID,
		Name: nameJSON,
		Icon: pgtype.Text{String: category.Icon, Valid: category.Icon != ""},
		IsActive: pgtype.Bool{Bool: category.IsActive, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}

	log.Info().Str("category_id", category.ID.String()).Msg("category updated")
	return nil
}

// SuspendUser deactivates a user account.
func (s *AdminService) SuspendUser(ctx context.Context, userID, adminID uuid.UUID, reason string) error {
	if err := s.queries.DeactivateUser(ctx, userID); err != nil {
		return fmt.Errorf("suspend user: %w", err)
	}

	log.Info().
		Str("user_id", userID.String()).
		Str("admin_id", adminID.String()).
		Str("reason", reason).
		Msg("user suspended")

	return nil
}
