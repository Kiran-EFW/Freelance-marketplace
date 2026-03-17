// Package fraud provides a fraud detection engine for the Seva platform.
// It implements velocity checks, review graph analysis, refund rate monitoring,
// provider no-show detection, and aggregate risk scoring using direct SQL
// queries against the platform's existing tables.
package fraud

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ---------------------------------------------------------------------------
// Result types
// ---------------------------------------------------------------------------

// VelocityResult holds the outcome of an account velocity check.
type VelocityResult struct {
	IsBlocked     bool    `json:"is_blocked"`
	AccountsFound int     `json:"accounts_found"`
	RiskScore     float64 `json:"risk_score"`
	Reason        string  `json:"reason,omitempty"`
}

// ReviewRingResult holds the outcome of a fake review ring analysis.
type ReviewRingResult struct {
	IsSuspicious      bool      `json:"is_suspicious"`
	MutualReviewCount int       `json:"mutual_review_count"`
	SuspiciousPairs   []UUIDPair `json:"suspicious_pairs,omitempty"`
	RiskScore         float64   `json:"risk_score"`
}

// UUIDPair represents a pair of user IDs involved in suspicious activity.
type UUIDPair struct {
	A uuid.UUID `json:"a"`
	B uuid.UUID `json:"b"`
}

// RefundRateResult holds the outcome of a refund rate analysis.
type RefundRateResult struct {
	TotalJobs    int     `json:"total_jobs"`
	RefundedJobs int     `json:"refunded_jobs"`
	DisputedJobs int     `json:"disputed_jobs"`
	RefundRate   float64 `json:"refund_rate"`
	IsAbnormal   bool    `json:"is_abnormal"`
	RiskScore    float64 `json:"risk_score"`
}

// NoShowResult holds the outcome of a provider no-show analysis.
type NoShowResult struct {
	TotalAccepted int     `json:"total_accepted"`
	NoShows       int     `json:"no_shows"`
	NoShowRate    float64 `json:"no_show_rate"`
	IsAbnormal    bool    `json:"is_abnormal"`
	RiskScore     float64 `json:"risk_score"`
}

// FlaggedAccount represents a user account that has been flagged for review.
type FlaggedAccount struct {
	UserID    uuid.UUID `json:"user_id"`
	UserName  string    `json:"user_name"`
	UserPhone string    `json:"user_phone"`
	UserRole  string    `json:"user_role"`
	RiskScore float64   `json:"risk_score"`
	Reason    string    `json:"reason"`
	FlaggedAt time.Time `json:"flagged_at"`
	Status    string    `json:"status"` // flagged, reviewed, cleared, suspended
}

// ---------------------------------------------------------------------------
// Thresholds
// ---------------------------------------------------------------------------

const (
	// velocityWindowHours is the look-back window for account creation velocity.
	velocityWindowHours = 24
	// maxAccountsPerPhone is the threshold for phone-based velocity.
	maxAccountsPerPhone = 2
	// maxAccountsPerDevice is the threshold for device-based velocity (reserved for future use).
	maxAccountsPerDevice = 3
	// maxAccountsPerIP is the threshold for IP-based velocity (reserved for future use).
	maxAccountsPerIP = 5

	// refundRateThreshold: refund rate above this is considered abnormal.
	refundRateThreshold = 0.30
	// noShowRateThreshold: no-show rate above this is considered abnormal.
	noShowRateThreshold = 0.25
	// mutualReviewThreshold: number of mutual review pairs before flagging.
	mutualReviewThreshold = 2

	// minJobsForRefundAnalysis: minimum completed jobs before refund rate is meaningful.
	minJobsForRefundAnalysis = 3
	// minAcceptedForNoShowAnalysis: minimum accepted jobs before no-show rate is meaningful.
	minAcceptedForNoShowAnalysis = 3
)

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// FraudService implements fraud detection operations using direct database
// queries against the platform's existing PostgreSQL tables.
type FraudService struct {
	db *pgxpool.Pool
}

// NewFraudService returns a ready-to-use FraudService. It ensures the
// flagged_accounts table exists on first use.
func NewFraudService(db *pgxpool.Pool) *FraudService {
	svc := &FraudService{db: db}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := svc.ensureTable(ctx); err != nil {
		log.Error().Err(err).Msg("fraud: failed to create flagged_accounts table")
	}

	return svc
}

// ensureTable creates the flagged_accounts table if it does not already exist.
func (s *FraudService) ensureTable(ctx context.Context) error {
	ddl := `CREATE TABLE IF NOT EXISTS flagged_accounts (
		id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		user_id     UUID NOT NULL REFERENCES users(id),
		risk_score  DECIMAL(5,2) NOT NULL DEFAULT 0.00,
		reason      TEXT NOT NULL,
		status      VARCHAR(20) NOT NULL DEFAULT 'flagged',
		flagged_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		reviewed_at TIMESTAMPTZ,
		reviewed_by UUID REFERENCES users(id),
		created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_flagged_accounts_user ON flagged_accounts(user_id);
	CREATE INDEX IF NOT EXISTS idx_flagged_accounts_status ON flagged_accounts(status);
	CREATE INDEX IF NOT EXISTS idx_flagged_accounts_score ON flagged_accounts(risk_score DESC);`

	_, err := s.db.Exec(ctx, ddl)
	return err
}

// ---------------------------------------------------------------------------
// CheckAccountVelocity
// ---------------------------------------------------------------------------

// CheckAccountVelocity determines whether the given phone, device ID, or IP
// address has been used to create an abnormal number of accounts within the
// recent velocity window. The phone check uses the users table directly.
// Device and IP checks are reserved for future implementation (when an
// account_metadata table is available) and currently contribute zero risk.
func (s *FraudService) CheckAccountVelocity(ctx context.Context, phone, deviceID, ip string) (*VelocityResult, error) {
	result := &VelocityResult{}

	// Count accounts sharing the same phone number created within the window.
	windowStart := time.Now().UTC().Add(-velocityWindowHours * time.Hour)

	var phoneCount int64
	err := s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE phone = $1 AND created_at >= $2`,
		phone, windowStart,
	).Scan(&phoneCount)
	if err != nil {
		return nil, fmt.Errorf("velocity check phone count: %w", err)
	}

	result.AccountsFound = int(phoneCount)

	if phoneCount > int64(maxAccountsPerPhone) {
		result.IsBlocked = true
		result.RiskScore = math.Min(100, float64(phoneCount)*25)
		result.Reason = fmt.Sprintf("phone %s used for %d accounts in last %d hours (threshold: %d)",
			phone, phoneCount, velocityWindowHours, maxAccountsPerPhone)
		log.Warn().
			Str("phone", phone).
			Int64("accounts", phoneCount).
			Float64("risk_score", result.RiskScore).
			Msg("fraud: velocity check blocked")
		return result, nil
	}

	// Risk score proportional to proximity to the threshold.
	if phoneCount > 1 {
		result.RiskScore = float64(phoneCount) / float64(maxAccountsPerPhone) * 50
		result.Reason = fmt.Sprintf("phone used for %d accounts (near threshold)", phoneCount)
	}

	return result, nil
}

// ---------------------------------------------------------------------------
// DetectFakeReviewRing
// ---------------------------------------------------------------------------

// DetectFakeReviewRing analyses the review graph around a provider to identify
// pairs of users who have reviewed each other (mutual reviews). A ring of
// providers mutually rating each other is a strong signal of collusion.
//
// The query self-joins the reviews table to find cases where user A reviewed
// user B AND user B reviewed user A (across different jobs).
func (s *FraudService) DetectFakeReviewRing(ctx context.Context, providerID uuid.UUID) (*ReviewRingResult, error) {
	result := &ReviewRingResult{}

	// Find mutual review pairs involving the given provider.
	query := `
		SELECT r1.reviewer_id AS a, r1.reviewee_id AS b, COUNT(*) AS pair_count
		FROM reviews r1
		JOIN reviews r2
			ON r1.reviewer_id = r2.reviewee_id
			AND r1.reviewee_id = r2.reviewer_id
			AND r1.id <> r2.id
		WHERE r1.reviewer_id = $1 OR r1.reviewee_id = $1
		GROUP BY r1.reviewer_id, r1.reviewee_id
		HAVING COUNT(*) >= 1
		ORDER BY pair_count DESC`

	rows, err := s.db.Query(ctx, query, providerID)
	if err != nil {
		return nil, fmt.Errorf("detect review ring: %w", err)
	}
	defer rows.Close()

	seen := make(map[string]bool)
	for rows.Next() {
		var a, b uuid.UUID
		var count int64
		if err := rows.Scan(&a, &b, &count); err != nil {
			log.Warn().Err(err).Msg("fraud: failed to scan review ring row")
			continue
		}

		// Deduplicate reversed pairs (A->B is the same ring as B->A).
		key := a.String() + "|" + b.String()
		reverseKey := b.String() + "|" + a.String()
		if seen[key] || seen[reverseKey] {
			continue
		}
		seen[key] = true

		result.MutualReviewCount++
		result.SuspiciousPairs = append(result.SuspiciousPairs, UUIDPair{A: a, B: b})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("review ring rows: %w", err)
	}

	if result.MutualReviewCount >= mutualReviewThreshold {
		result.IsSuspicious = true
		result.RiskScore = math.Min(100, float64(result.MutualReviewCount)*30)
	} else if result.MutualReviewCount > 0 {
		result.RiskScore = float64(result.MutualReviewCount) * 15
	}

	log.Info().
		Str("provider_id", providerID.String()).
		Int("mutual_reviews", result.MutualReviewCount).
		Bool("suspicious", result.IsSuspicious).
		Msg("fraud: review ring analysis complete")

	return result, nil
}

// ---------------------------------------------------------------------------
// MonitorRefundRate
// ---------------------------------------------------------------------------

// MonitorRefundRate checks whether a user has an abnormally high rate of
// refunded transactions or filed disputes relative to their total completed
// jobs. This catches both fake-complaint customers and bait-and-switch
// providers.
func (s *FraudService) MonitorRefundRate(ctx context.Context, userID uuid.UUID) (*RefundRateResult, error) {
	result := &RefundRateResult{}

	// Count total completed jobs for this user (as either customer or provider).
	var totalJobs int64
	err := s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM jobs
		 WHERE (customer_id = $1 OR provider_id = $1)
		   AND status IN ('completed', 'cancelled', 'disputed')`,
		userID,
	).Scan(&totalJobs)
	if err != nil {
		return nil, fmt.Errorf("refund rate total jobs: %w", err)
	}
	result.TotalJobs = int(totalJobs)

	// Count refunded transactions.
	var refundedJobs int64
	err = s.db.QueryRow(ctx,
		`SELECT COUNT(DISTINCT t.job_id) FROM transactions t
		 JOIN jobs j ON j.id = t.job_id
		 WHERE (j.customer_id = $1 OR j.provider_id = $1)
		   AND t.payment_status IN ('refunded', 'partially_refunded')`,
		userID,
	).Scan(&refundedJobs)
	if err != nil {
		return nil, fmt.Errorf("refund rate refunded count: %w", err)
	}
	result.RefundedJobs = int(refundedJobs)

	// Count disputes raised by this user.
	var disputedJobs int64
	err = s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM disputes WHERE raised_by = $1`,
		userID,
	).Scan(&disputedJobs)
	if err != nil {
		return nil, fmt.Errorf("refund rate dispute count: %w", err)
	}
	result.DisputedJobs = int(disputedJobs)

	// Calculate refund rate.
	if result.TotalJobs > 0 {
		result.RefundRate = float64(result.RefundedJobs+result.DisputedJobs) / float64(result.TotalJobs)
	}

	// Only flag as abnormal if the user has enough history.
	if result.TotalJobs >= minJobsForRefundAnalysis && result.RefundRate > refundRateThreshold {
		result.IsAbnormal = true
		result.RiskScore = math.Min(100, result.RefundRate*100)
	} else if result.RefundRate > 0 {
		result.RiskScore = result.RefundRate / refundRateThreshold * 40
	}

	log.Info().
		Str("user_id", userID.String()).
		Int("total_jobs", result.TotalJobs).
		Int("refunded", result.RefundedJobs).
		Int("disputed", result.DisputedJobs).
		Float64("rate", result.RefundRate).
		Bool("abnormal", result.IsAbnormal).
		Msg("fraud: refund rate analysis complete")

	return result, nil
}

// ---------------------------------------------------------------------------
// CheckProviderNoShowRate
// ---------------------------------------------------------------------------

// CheckProviderNoShowRate detects providers who accept jobs but fail to show
// up. An accepted job that is subsequently cancelled or disputed as "no_show"
// counts as a no-show.
func (s *FraudService) CheckProviderNoShowRate(ctx context.Context, providerID uuid.UUID) (*NoShowResult, error) {
	result := &NoShowResult{}

	// Count jobs accepted by this provider (status progressed past 'accepted').
	var totalAccepted int64
	err := s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM jobs
		 WHERE provider_id = $1
		   AND status IN ('accepted', 'in_progress', 'completed', 'cancelled', 'disputed')`,
		providerID,
	).Scan(&totalAccepted)
	if err != nil {
		return nil, fmt.Errorf("no-show total accepted: %w", err)
	}
	result.TotalAccepted = int(totalAccepted)

	// Count no-shows: cancelled jobs where a no_show dispute was raised, or
	// disputes of type 'no_show' against this provider.
	var noShows int64
	err = s.db.QueryRow(ctx,
		`SELECT COUNT(DISTINCT d.job_id) FROM disputes d
		 JOIN jobs j ON j.id = d.job_id
		 WHERE j.provider_id = $1
		   AND d.type = 'no_show'`,
		providerID,
	).Scan(&noShows)
	if err != nil {
		return nil, fmt.Errorf("no-show dispute count: %w", err)
	}

	// Also count jobs cancelled by the provider after acceptance (heuristic:
	// provider was assigned but job was cancelled and no dispute was raised).
	var cancelledAfterAccept int64
	err = s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM jobs
		 WHERE provider_id = $1
		   AND status = 'cancelled'
		   AND scheduled_at IS NOT NULL
		   AND scheduled_at < NOW()`,
		providerID,
	).Scan(&cancelledAfterAccept)
	if err != nil {
		log.Warn().Err(err).Msg("fraud: failed to count cancelled-after-accept jobs")
	}

	result.NoShows = int(noShows) + int(cancelledAfterAccept)

	if result.TotalAccepted > 0 {
		result.NoShowRate = float64(result.NoShows) / float64(result.TotalAccepted)
	}

	if result.TotalAccepted >= minAcceptedForNoShowAnalysis && result.NoShowRate > noShowRateThreshold {
		result.IsAbnormal = true
		result.RiskScore = math.Min(100, result.NoShowRate*120)
	} else if result.NoShowRate > 0 {
		result.RiskScore = result.NoShowRate / noShowRateThreshold * 35
	}

	log.Info().
		Str("provider_id", providerID.String()).
		Int("accepted", result.TotalAccepted).
		Int("no_shows", result.NoShows).
		Float64("rate", result.NoShowRate).
		Bool("abnormal", result.IsAbnormal).
		Msg("fraud: no-show analysis complete")

	return result, nil
}

// ---------------------------------------------------------------------------
// CalculateRiskScore
// ---------------------------------------------------------------------------

// CalculateRiskScore produces an aggregate risk score (0-100) for a user by
// combining signals from refund rate monitoring, review ring detection, and
// no-show analysis. The score is a weighted average of individual signals.
func (s *FraudService) CalculateRiskScore(ctx context.Context, userID uuid.UUID) (float64, error) {
	var totalScore float64
	var weights float64

	// 1. Refund rate signal (weight 0.35).
	refund, err := s.MonitorRefundRate(ctx, userID)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("fraud: refund rate check failed during risk score")
	} else {
		totalScore += refund.RiskScore * 0.35
		weights += 0.35
	}

	// 2. Review ring signal (weight 0.30).
	ring, err := s.DetectFakeReviewRing(ctx, userID)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("fraud: review ring check failed during risk score")
	} else {
		totalScore += ring.RiskScore * 0.30
		weights += 0.30
	}

	// 3. No-show signal (weight 0.25).
	noShow, err := s.CheckProviderNoShowRate(ctx, userID)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("fraud: no-show check failed during risk score")
	} else {
		totalScore += noShow.RiskScore * 0.25
		weights += 0.25
	}

	// 4. Dispute severity signal (weight 0.10) — counts escalated/critical disputes.
	var escalatedCount int64
	err = s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM disputes
		 WHERE (raised_by = $1 OR against = $1)
		   AND (severity IN ('high', 'critical') OR status = 'escalated')`,
		userID,
	).Scan(&escalatedCount)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("fraud: escalated dispute count failed")
	} else {
		disputeScore := math.Min(100, float64(escalatedCount)*25)
		totalScore += disputeScore * 0.10
		weights += 0.10
	}

	// Normalize if some signals failed.
	if weights > 0 {
		totalScore = totalScore / weights
	}

	return math.Round(totalScore*100) / 100, nil
}

// ---------------------------------------------------------------------------
// FlagAccount
// ---------------------------------------------------------------------------

// FlagAccount inserts or updates a flagged_accounts record for the given user.
// If the user is already flagged, the risk score and reason are updated.
func (s *FraudService) FlagAccount(ctx context.Context, userID uuid.UUID, reason string, riskScore float64) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO flagged_accounts (user_id, risk_score, reason, status, flagged_at)
		 VALUES ($1, $2, $3, 'flagged', NOW())
		 ON CONFLICT (user_id) WHERE status = 'flagged'
		 DO UPDATE SET risk_score = EXCLUDED.risk_score,
		              reason = EXCLUDED.reason,
		              updated_at = NOW()`,
		userID, riskScore, reason,
	)
	if err != nil {
		// If ON CONFLICT fails (no partial unique index), fall back to a simple insert.
		_, err = s.db.Exec(ctx,
			`INSERT INTO flagged_accounts (user_id, risk_score, reason, status, flagged_at)
			 VALUES ($1, $2, $3, 'flagged', NOW())`,
			userID, riskScore, reason,
		)
		if err != nil {
			return fmt.Errorf("flag account: %w", err)
		}
	}

	log.Info().
		Str("user_id", userID.String()).
		Float64("risk_score", riskScore).
		Str("reason", reason).
		Msg("fraud: account flagged")

	return nil
}

// ---------------------------------------------------------------------------
// GetFlaggedAccounts
// ---------------------------------------------------------------------------

// GetFlaggedAccounts returns a paginated list of flagged accounts joined with
// user information, sorted by risk score descending.
func (s *FraudService) GetFlaggedAccounts(ctx context.Context, limit, offset int) ([]FlaggedAccount, int, error) {
	// Count total flagged accounts.
	var total int64
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM flagged_accounts`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count flagged accounts: %w", err)
	}

	query := `
		SELECT fa.user_id, u.name, u.phone, u.type,
		       fa.risk_score, fa.reason, fa.flagged_at, fa.status
		FROM flagged_accounts fa
		JOIN users u ON u.id = fa.user_id
		ORDER BY fa.risk_score DESC, fa.flagged_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := s.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list flagged accounts: %w", err)
	}
	defer rows.Close()

	var accounts []FlaggedAccount
	for rows.Next() {
		var fa FlaggedAccount
		var name pgtype.Text
		var riskScore pgtype.Numeric
		if err := rows.Scan(
			&fa.UserID, &name, &fa.UserPhone, &fa.UserRole,
			&riskScore, &fa.Reason, &fa.FlaggedAt, &fa.Status,
		); err != nil {
			log.Warn().Err(err).Msg("fraud: failed to scan flagged account row")
			continue
		}
		if name.Valid {
			fa.UserName = name.String
		}
		if riskScore.Valid {
			f, _ := riskScore.Float64Value()
			if f.Valid {
				fa.RiskScore = f.Float64
			}
		}
		accounts = append(accounts, fa)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("flagged accounts rows: %w", err)
	}

	return accounts, int(total), nil
}

// ---------------------------------------------------------------------------
// UpdateFlagStatus
// ---------------------------------------------------------------------------

// UpdateFlagStatus updates the status of a flagged account (e.g. cleared,
// suspended, reviewed). It also optionally deactivates the user when
// suspended.
func (s *FraudService) UpdateFlagStatus(ctx context.Context, userID uuid.UUID, newStatus string, adminID uuid.UUID) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE flagged_accounts
		 SET status = $1, reviewed_at = NOW(), reviewed_by = $2, updated_at = NOW()
		 WHERE user_id = $3 AND status = 'flagged'`,
		newStatus, adminID, userID,
	)
	if err != nil {
		return fmt.Errorf("update flag status: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no active flagged account found for user %s", userID)
	}

	// If suspending the user, also deactivate them in the users table.
	if newStatus == "suspended" {
		_, err = s.db.Exec(ctx,
			`UPDATE users SET is_active = false, updated_at = NOW() WHERE id = $1`,
			userID,
		)
		if err != nil {
			return fmt.Errorf("deactivate user on suspend: %w", err)
		}
	}

	// If clearing, ensure the user is active.
	if newStatus == "cleared" {
		_, err = s.db.Exec(ctx,
			`UPDATE users SET is_active = true, updated_at = NOW() WHERE id = $1`,
			userID,
		)
		if err != nil {
			log.Warn().Err(err).Str("user_id", userID.String()).Msg("fraud: failed to reactivate cleared user")
		}
	}

	log.Info().
		Str("user_id", userID.String()).
		Str("new_status", newStatus).
		Str("admin_id", adminID.String()).
		Msg("fraud: flag status updated")

	return nil
}

// ---------------------------------------------------------------------------
// GetUserRiskProfile
// ---------------------------------------------------------------------------

// GetUserRiskProfile returns the computed risk score plus any existing flag
// information for a user. This is a convenience method for the admin UI.
func (s *FraudService) GetUserRiskProfile(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	riskScore, err := s.CalculateRiskScore(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("calculate risk score: %w", err)
	}

	profile := map[string]interface{}{
		"user_id":    userID,
		"risk_score": riskScore,
	}

	// Check if the user is already flagged.
	var flagStatus pgtype.Text
	var flagReason pgtype.Text
	var flaggedAt pgtype.Timestamptz
	err = s.db.QueryRow(ctx,
		`SELECT status, reason, flagged_at FROM flagged_accounts
		 WHERE user_id = $1 ORDER BY flagged_at DESC LIMIT 1`,
		userID,
	).Scan(&flagStatus, &flagReason, &flaggedAt)
	if err != nil && err != pgx.ErrNoRows {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("fraud: failed to check flag status")
	}

	if flagStatus.Valid {
		profile["flag_status"] = flagStatus.String
		profile["flag_reason"] = flagReason.String
		if flaggedAt.Valid {
			profile["flagged_at"] = flaggedAt.Time
		}
	} else {
		profile["flag_status"] = "none"
	}

	// Fetch individual signal details.
	refundResult, err := s.MonitorRefundRate(ctx, userID)
	if err == nil {
		profile["refund_rate"] = refundResult
	}

	noShowResult, err := s.CheckProviderNoShowRate(ctx, userID)
	if err == nil {
		profile["no_show"] = noShowResult
	}

	ringResult, err := s.DetectFakeReviewRing(ctx, userID)
	if err == nil {
		profile["review_ring"] = ringResult
	}

	return profile, nil
}
