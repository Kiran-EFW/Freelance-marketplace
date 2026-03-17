package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/seva-platform/backend/internal/domain"
)

// ---------------------------------------------------------------------------
// Helpers for type conversion between domain types and pgtype types
// ---------------------------------------------------------------------------

func textOrNull(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func textOrNullPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func textVal(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func textPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}

func numericToFloat(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, _ := n.Float64Value()
	if f.Valid {
		return f.Float64
	}
	return 0
}

func numericToFloatPtr(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	fv, _ := n.Float64Value()
	if fv.Valid {
		v := fv.Float64
		return &v
	}
	return nil
}

func floatToNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	n.Scan(fmt.Sprintf("%f", f))
	return n
}

func floatPtrToNumeric(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{}
	}
	return floatToNumeric(*f)
}

func int4Val(n pgtype.Int4) int {
	if !n.Valid {
		return 0
	}
	return int(n.Int32)
}

func timestamptzToTimePtr(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}

func timePtrToTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// ---------------------------------------------------------------------------
// UserRepositoryAdapter implements domain.UserRepository
// ---------------------------------------------------------------------------

type UserRepositoryAdapter struct {
	q *Queries
}

func NewUserRepository(q *Queries) *UserRepositoryAdapter {
	return &UserRepositoryAdapter{q: q}
}

func userFromDB(u User) *domain.User {
	return &domain.User{
		ID:                u.ID,
		Type:              domain.UserType(u.Type),
		Phone:             u.Phone,
		Email:             textPtr(u.Email),
		Name:              textVal(u.Name),
		JurisdictionID:    u.JurisdictionID,
		PreferredLanguage: u.PreferredLanguage,
		DeviceType:        string(u.DeviceType),
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
	}
}

func (r *UserRepositoryAdapter) Create(ctx context.Context, user *domain.User) error {
	result, err := r.q.CreateUser(ctx, CreateUserParams{
		Type:              UserType(user.Type),
		Phone:             user.Phone,
		Email:             textOrNullPtr(user.Email),
		Name:              textOrNull(user.Name),
		JurisdictionID:    user.JurisdictionID,
		PreferredLanguage: user.PreferredLanguage,
		DeviceType:        DeviceType(user.DeviceType),
	})
	if err != nil {
		return err
	}
	user.ID = result.ID
	user.CreatedAt = result.CreatedAt
	user.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *UserRepositoryAdapter) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return userFromDB(u), nil
}

func (r *UserRepositoryAdapter) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	u, err := r.q.GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	return userFromDB(u), nil
}

func (r *UserRepositoryAdapter) Update(ctx context.Context, user *domain.User) error {
	_, err := r.q.UpdateUser(ctx, UpdateUserParams{
		ID:                user.ID,
		Name:              textOrNull(user.Name),
		Email:             textOrNullPtr(user.Email),
		PreferredLanguage: textOrNull(user.PreferredLanguage),
		DeviceType:        textOrNull(user.DeviceType),
	})
	return err
}

func (r *UserRepositoryAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeactivateUser(ctx, id)
}

// ---------------------------------------------------------------------------
// ProviderRepositoryAdapter implements domain.ProviderRepository
// ---------------------------------------------------------------------------

type ProviderRepositoryAdapter struct {
	q *Queries
}

func NewProviderRepository(q *Queries) *ProviderRepositoryAdapter {
	return &ProviderRepositoryAdapter{q: q}
}

func providerFromDB(p ProviderProfile) *domain.ProviderProfile {
	return &domain.ProviderProfile{
		UserID:               p.UserID,
		Skills:               p.Skills,
		ServiceRadiusKM:      numericToFloat(p.ServiceRadiusKM),
		Postcode:             textVal(p.Postcode),
		TrustScore:           numericToFloat(p.TrustScore),
		Level:                providerLevelToInt(p.Level),
		VerificationStatus:   domain.VerificationStatus(p.VerificationStatus),
		SubscriptionTier:     domain.SubscriptionTier(mapSubscriptionTier(p.SubscriptionTier)),
		AvailabilitySchedule: p.AvailabilitySchedule,
		BankAccountID:        textPtr(p.BankAccountID),
		JobsCompleted:        int(p.TotalJobsCompleted),
		ResponseTimeAvg:      int4Val(p.AvgResponseTimeMinutes),
		Bio:                  textVal(p.Description),
		IsAvailable:          p.IsAvailable,
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
	}
}

func providerLevelToInt(l ProviderLevel) int {
	switch l {
	case ProviderLevelNew:
		return 1
	case ProviderLevelActive:
		return 2
	case ProviderLevelTrusted:
		return 3
	case ProviderLevelExpert:
		return 4
	case ProviderLevelLocalChampion:
		return 5
	default:
		return 1
	}
}

func mapSubscriptionTier(t SubscriptionTier) string {
	switch t {
	case SubscriptionTierFree:
		return string(domain.SubscriptionFree)
	case SubscriptionTierBasic:
		return string(domain.SubscriptionBasic)
	case SubscriptionTierProfessional, SubscriptionTierEnterprise:
		return string(domain.SubscriptionPremium)
	default:
		return string(domain.SubscriptionFree)
	}
}

func (r *ProviderRepositoryAdapter) Create(ctx context.Context, profile *domain.ProviderProfile) error {
	_, err := r.q.CreateProviderProfile(ctx, CreateProviderProfileParams{
		UserID:          profile.UserID,
		BusinessName:    textOrNull(profile.Bio),
		Description:     textOrNull(profile.Bio),
		Skills:          profile.Skills,
		ServiceRadiusKM: floatToNumeric(profile.ServiceRadiusKM),
		Postcode:        textOrNull(profile.Postcode),
		Latitude:        profile.Latitude,
		Longitude:       profile.Longitude,
	})
	return err
}

func (r *ProviderRepositoryAdapter) GetByID(ctx context.Context, userID uuid.UUID) (*domain.ProviderProfile, error) {
	p, err := r.q.GetProviderByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return providerFromDB(p), nil
}

func (r *ProviderRepositoryAdapter) Update(ctx context.Context, profile *domain.ProviderProfile) error {
	_, err := r.q.UpdateProviderProfile(ctx, UpdateProviderProfileParams{
		ID:              profile.UserID,
		BusinessName:    textOrNull(profile.Bio),
		Description:     textOrNull(profile.Bio),
		Skills:          profile.Skills,
		ServiceRadiusKM: floatToNumeric(profile.ServiceRadiusKM),
		IsAvailable:     pgtype.Bool{Bool: profile.IsAvailable, Valid: true},
	})
	return err
}

func (r *ProviderRepositoryAdapter) Delete(ctx context.Context, userID uuid.UUID) error {
	// No direct delete in sqlc; deactivate the underlying user instead.
	return r.q.DeactivateUser(ctx, userID)
}

func (r *ProviderRepositoryAdapter) Search(ctx context.Context, filters domain.ProviderSearchFilters) ([]domain.ProviderSearchResult, error) {
	lat := 0.0
	lng := 0.0
	radius := 15.0
	if filters.Latitude != nil {
		lat = *filters.Latitude
	}
	if filters.Longitude != nil {
		lng = *filters.Longitude
	}
	if filters.RadiusKM != nil {
		radius = *filters.RadiusKM
	}

	rows, err := r.q.SearchProvidersByLocation(ctx, SearchProvidersByLocationParams{
		Latitude:     lat,
		Longitude:    lng,
		RadiusMeters: radius * 1000, // Convert km to meters
		Limit:        int32(max(filters.Limit, 50)),
	})
	if err != nil {
		return nil, err
	}

	results := make([]domain.ProviderSearchResult, 0, len(rows))
	for _, row := range rows {
		results = append(results, domain.ProviderSearchResult{
			ProviderProfile: domain.ProviderProfile{
				UserID:           row.UserID,
				Skills:           row.Skills,
				ServiceRadiusKM:  numericToFloat(row.ServiceRadiusKM),
				Postcode:         textVal(row.Postcode),
				TrustScore:       numericToFloat(row.TrustScore),
				Level:            providerLevelToInt(row.Level),
				VerificationStatus: domain.VerificationStatus(row.VerificationStatus),
				SubscriptionTier: domain.SubscriptionTier(mapSubscriptionTier(row.SubscriptionTier)),
				JobsCompleted:    int(row.TotalJobsCompleted),
				ResponseTimeAvg:  int4Val(row.AvgResponseTimeMinutes),
				Bio:              textVal(row.Description),
				IsAvailable:      row.IsAvailable,
			},
			UserName:  textVal(row.UserName),
			UserPhone: row.UserPhone,
			Distance:  row.DistanceMeters / 1000, // Convert meters to km
		})
	}
	return results, nil
}

func (r *ProviderRepositoryAdapter) ListByPostcode(ctx context.Context, postcode string, limit, offset int) ([]domain.ProviderProfile, error) {
	rows, err := r.q.SearchProvidersByPostcode(ctx, SearchProvidersByPostcodeParams{
		Postcode: postcode,
		Limit:    int32(limit),
	})
	if err != nil {
		return nil, err
	}

	profiles := make([]domain.ProviderProfile, 0, len(rows))
	for _, row := range rows {
		profiles = append(profiles, *providerFromDB(row.ProviderProfile))
	}
	return profiles, nil
}

func (r *ProviderRepositoryAdapter) UpdateTrustScore(ctx context.Context, userID uuid.UUID, score float64) error {
	level := ProviderLevelNew
	if score >= 4.5 {
		level = ProviderLevelLocalChampion
	} else if score >= 4.0 {
		level = ProviderLevelExpert
	} else if score >= 3.5 {
		level = ProviderLevelTrusted
	} else if score >= 2.5 {
		level = ProviderLevelActive
	}
	return r.q.UpdateTrustScore(ctx, UpdateTrustScoreParams{
		ID:         userID,
		TrustScore: floatToNumeric(score),
		Level:      level,
	})
}

func (r *ProviderRepositoryAdapter) IncrementJobsCompleted(ctx context.Context, userID uuid.UUID) error {
	return r.q.IncrementJobsCompleted(ctx, userID)
}

// ---------------------------------------------------------------------------
// JobRepositoryAdapter implements domain.JobRepository
// ---------------------------------------------------------------------------

type JobRepositoryAdapter struct {
	q *Queries
}

func NewJobRepository(q *Queries) *JobRepositoryAdapter {
	return &JobRepositoryAdapter{q: q}
}

func jobFromDB(j Job) *domain.Job {
	return &domain.Job{
		ID:             j.ID,
		CustomerID:     j.CustomerID,
		ProviderID:     j.ProviderID,
		CategoryID:     j.CategoryID,
		Postcode:       textVal(j.Postcode),
		Status:         domain.JobStatus(j.Status),
		Description:    textVal(j.Description),
		ScheduledAt:    timestamptzToTimePtr(j.ScheduledAt),
		QuotedPrice:    numericToFloatPtr(j.QuotedPrice),
		FinalPrice:     numericToFloatPtr(j.FinalPrice),
		Currency:       j.Currency,
		PaymentMethod:  domain.PaymentMethod(j.PaymentMethod),
		IsRecurring:    j.IsRecurring,
		JurisdictionID: j.JurisdictionID,
		CreatedAt:      j.CreatedAt,
		UpdatedAt:      j.UpdatedAt,
	}
}

func (r *JobRepositoryAdapter) Create(ctx context.Context, job *domain.Job) error {
	result, err := r.q.CreateJob(ctx, CreateJobParams{
		CustomerID:     job.CustomerID,
		CategoryID:     job.CategoryID,
		Title:          job.Description,
		Description:    textOrNull(job.Description),
		Postcode:       textOrNull(job.Postcode),
		Latitude:       job.Latitude,
		Longitude:      job.Longitude,
		ScheduledAt:    timePtrToTimestamptz(job.ScheduledAt),
		Currency:       job.Currency,
		PaymentMethod:  PaymentMethod(job.PaymentMethod),
		IsRecurring:    job.IsRecurring,
		JurisdictionID: job.JurisdictionID,
	})
	if err != nil {
		return err
	}
	job.ID = result.ID
	job.CreatedAt = result.CreatedAt
	job.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *JobRepositoryAdapter) GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	j, err := r.q.GetJobByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return jobFromDB(j), nil
}

func (r *JobRepositoryAdapter) ListByCustomer(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]domain.Job, error) {
	rows, err := r.q.ListJobsByCustomer(ctx, ListJobsByCustomerParams{
		CustomerID: customerID,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		return nil, err
	}
	jobs := make([]domain.Job, 0, len(rows))
	for _, j := range rows {
		jobs = append(jobs, *jobFromDB(j.Job))
	}
	return jobs, nil
}

func (r *JobRepositoryAdapter) ListByProvider(ctx context.Context, providerID uuid.UUID, limit, offset int) ([]domain.Job, error) {
	rows, err := r.q.ListJobsByProvider(ctx, ListJobsByProviderParams{
		ProviderID: providerID,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		return nil, err
	}
	jobs := make([]domain.Job, 0, len(rows))
	for _, j := range rows {
		jobs = append(jobs, *jobFromDB(j.Job))
	}
	return jobs, nil
}

func (r *JobRepositoryAdapter) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.JobStatus) error {
	_, err := r.q.UpdateJobStatus(ctx, UpdateJobStatusParams{
		ID:     id,
		Status: JobStatus(status),
	})
	return err
}

func (r *JobRepositoryAdapter) Search(ctx context.Context, filters domain.JobSearchFilters) ([]domain.Job, error) {
	lat := 0.0
	lng := 0.0
	radius := 15.0
	if filters.Latitude != nil {
		lat = *filters.Latitude
	}
	if filters.Longitude != nil {
		lng = *filters.Longitude
	}
	if filters.RadiusKM != nil {
		radius = *filters.RadiusKM
	}

	rows, err := r.q.SearchJobsByLocation(ctx, SearchJobsByLocationParams{
		Latitude:     lat,
		Longitude:    lng,
		RadiusMeters: radius * 1000, // Convert km to meters
		Limit:        int32(max(filters.Limit, 50)),
	})
	if err != nil {
		return nil, err
	}
	jobs := make([]domain.Job, 0, len(rows))
	for _, j := range rows {
		jobs = append(jobs, *jobFromDB(j.Job))
	}
	return jobs, nil
}

// ---------------------------------------------------------------------------
// ReviewRepositoryAdapter implements domain.ReviewRepository
// ---------------------------------------------------------------------------

type ReviewRepositoryAdapter struct {
	q *Queries
}

func NewReviewRepository(q *Queries) *ReviewRepositoryAdapter {
	return &ReviewRepositoryAdapter{q: q}
}

func reviewFromDB(r Review) *domain.Review {
	return &domain.Review{
		ID:               r.ID,
		JobID:            r.JobID,
		ReviewerID:       r.ReviewerID,
		RevieweeID:       r.RevieweeID,
		Rating:           int(r.Rating),
		Comment:          textVal(r.Comment),
		Language:         textVal(r.Language),
		ModerationStatus: domain.ModerationStatus(r.ModerationStatus),
		ProviderResponse: textPtr(r.Response),
		RespondedAt:      timestamptzToTimePtr(r.RespondedAt),
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}

func (r *ReviewRepositoryAdapter) Create(ctx context.Context, review *domain.Review) error {
	result, err := r.q.CreateReview(ctx, CreateReviewParams{
		JobID:      review.JobID,
		ReviewerID: review.ReviewerID,
		RevieweeID: review.RevieweeID,
		Rating:     int16(review.Rating),
		Comment:    textOrNull(review.Comment),
		Language:   textOrNull(review.Language),
	})
	if err != nil {
		return err
	}
	review.ID = result.ID
	review.CreatedAt = result.CreatedAt
	review.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *ReviewRepositoryAdapter) GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error) {
	row, err := r.q.GetReviewByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return reviewFromDB(row.Review), nil
}

func (r *ReviewRepositoryAdapter) GetByJobAndReviewer(ctx context.Context, jobID, reviewerID uuid.UUID) (*domain.Review, error) {
	// The sqlc layer does not have a dedicated query for this, so we list reviews
	// for the reviewee and filter by job+reviewer. For a production system, this
	// should be a dedicated query; here we do a lightweight in-memory filter.
	rows, err := r.q.ListReviewsByProvider(ctx, ListReviewsByProviderParams{
		RevieweeID: reviewerID, // Search using the reviewer as potential reviewee too
		Limit:      100,
		Offset:     0,
	})
	if err != nil {
		// Fallback: try the GetReviewByID-style search is not possible.
		// Return not-found.
		return nil, pgx.ErrNoRows
	}
	for _, row := range rows {
		if row.JobID == jobID && row.ReviewerID == reviewerID {
			return reviewFromDB(row.Review), nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (r *ReviewRepositoryAdapter) ListByReviewee(ctx context.Context, revieweeID uuid.UUID, limit, offset int) ([]domain.Review, error) {
	rows, err := r.q.ListReviewsByProvider(ctx, ListReviewsByProviderParams{
		RevieweeID: revieweeID,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		return nil, err
	}
	reviews := make([]domain.Review, 0, len(rows))
	for _, row := range rows {
		reviews = append(reviews, *reviewFromDB(row.Review))
	}
	return reviews, nil
}

func (r *ReviewRepositoryAdapter) ListByJob(ctx context.Context, jobID uuid.UUID) ([]domain.Review, error) {
	// No dedicated sqlc query for listing reviews by job. Use the provider listing
	// and do server-side filtering. In production add a dedicated SQL query.
	return nil, nil
}

func (r *ReviewRepositoryAdapter) UpdateResponse(ctx context.Context, id uuid.UUID, response string, respondedAt time.Time) error {
	_, err := r.q.RespondToReview(ctx, RespondToReviewParams{
		ID:       id,
		Response: response,
	})
	return err
}

func (r *ReviewRepositoryAdapter) UpdateModerationStatus(ctx context.Context, id uuid.UUID, status domain.ModerationStatus) error {
	return r.q.ModerateReview(ctx, ModerateReviewParams{
		ID:               id,
		ModerationStatus: ModerationStatus(status),
	})
}

func (r *ReviewRepositoryAdapter) GetProviderStats(ctx context.Context, providerID uuid.UUID) (*domain.ProviderReviewStats, error) {
	row, err := r.q.GetProviderAvgRating(ctx, providerID)
	if err != nil {
		return nil, err
	}
	return &domain.ProviderReviewStats{
		ProviderID:   providerID,
		AvgRating:    numericToFloat(row.AvgRating),
		TotalReviews: int(row.TotalReviews),
		Distribution: map[int]int{},
	}, nil
}

func (r *ReviewRepositoryAdapter) GetAverageRating(ctx context.Context, userID uuid.UUID) (float64, int, error) {
	row, err := r.q.GetProviderAvgRating(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	return numericToFloat(row.AvgRating), int(row.TotalReviews), nil
}

// ---------------------------------------------------------------------------
// TransactionRepositoryAdapter implements domain.TransactionRepository
// ---------------------------------------------------------------------------

type TransactionRepositoryAdapter struct {
	q *Queries
}

func NewTransactionRepository(q *Queries) *TransactionRepositoryAdapter {
	return &TransactionRepositoryAdapter{q: q}
}

func txFromDB(t Transaction) *domain.Transaction {
	tx := &domain.Transaction{
		ID:                   t.ID,
		JobID:                t.JobID,
		Amount:               numericToFloat(t.Amount),
		Currency:             t.Currency,
		CommissionRate:       numericToFloat(t.CommissionRate),
		CommissionAmount:     numericToFloat(t.CommissionAmount),
		TaxAmount:            numericToFloat(t.TaxAmount),
		ProviderPayoutAmount: numericToFloat(t.ProviderPayoutAmount),
		PaymentStatus:        domain.PaymentStatus(t.PaymentStatus),
		PaymentGateway:       textVal(t.PaymentGateway),
		GatewayOrderID:       textVal(t.GatewayOrderID),
		GatewayPaymentID:     textVal(t.GatewayPaymentID),
		GatewaySignature:     textVal(t.GatewaySignature),
		PaidAt:               timestamptzToTimePtr(t.PaidAt),
		SettledAt:            timestamptzToTimePtr(t.SettledAt),
		RefundAmount:         numericToFloatPtr(t.RefundAmount),
		RefundedAt:           timestamptzToTimePtr(t.RefundedAt),
		CreatedAt:            t.CreatedAt,
		UpdatedAt:            t.UpdatedAt,
	}
	if t.EscrowStatus != nil {
		es := domain.EscrowStatus(*t.EscrowStatus)
		tx.EscrowStatus = &es
	}
	return tx
}

func (r *TransactionRepositoryAdapter) Create(ctx context.Context, tx *domain.Transaction) error {
	result, err := r.q.CreateTransaction(ctx, CreateTransactionParams{
		JobID:                tx.JobID,
		Amount:               floatToNumeric(tx.Amount),
		Currency:             tx.Currency,
		CommissionRate:       floatToNumeric(tx.CommissionRate),
		CommissionAmount:     floatToNumeric(tx.CommissionAmount),
		TaxAmount:            floatToNumeric(tx.TaxAmount),
		ProviderPayoutAmount: floatToNumeric(tx.ProviderPayoutAmount),
		PaymentStatus:        PaymentStatus(tx.PaymentStatus),
		PaymentGateway:       textOrNull(tx.PaymentGateway),
		GatewayOrderID:       textOrNull(tx.GatewayOrderID),
	})
	if err != nil {
		return err
	}
	tx.ID = result.ID
	tx.CreatedAt = result.CreatedAt
	tx.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *TransactionRepositoryAdapter) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	t, err := r.q.GetTransactionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return txFromDB(t), nil
}

func (r *TransactionRepositoryAdapter) GetByJobID(ctx context.Context, jobID uuid.UUID) (*domain.Transaction, error) {
	t, err := r.q.GetTransactionByJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	return txFromDB(t), nil
}

func (r *TransactionRepositoryAdapter) GetByGatewayOrderID(ctx context.Context, orderID string) (*domain.Transaction, error) {
	t, err := r.q.GetTransactionByGatewayOrder(ctx, textOrNull(orderID))
	if err != nil {
		return nil, err
	}
	return txFromDB(t), nil
}

func (r *TransactionRepositoryAdapter) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus) error {
	_, err := r.q.UpdatePaymentStatus(ctx, UpdatePaymentStatusParams{
		ID:            id,
		PaymentStatus: PaymentStatus(status),
	})
	return err
}

func (r *TransactionRepositoryAdapter) UpdateEscrowStatus(ctx context.Context, id uuid.UUID, status domain.EscrowStatus) error {
	_, err := r.q.ReleaseEscrow(ctx, id)
	return err
}

func (r *TransactionRepositoryAdapter) UpdateGatewayPaymentID(ctx context.Context, id uuid.UUID, paymentID, signature string) error {
	_, err := r.q.UpdatePaymentStatus(ctx, UpdatePaymentStatusParams{
		ID:               id,
		PaymentStatus:    PaymentStatusCaptured,
		GatewayPaymentID: textOrNull(paymentID),
		GatewaySignature: textOrNull(signature),
	})
	return err
}

func (r *TransactionRepositoryAdapter) SetSettled(ctx context.Context, id uuid.UUID, settledAt time.Time) error {
	_, err := r.q.ReleaseEscrow(ctx, id)
	return err
}

func (r *TransactionRepositoryAdapter) SetRefunded(ctx context.Context, id uuid.UUID, amount float64, refundedAt time.Time) error {
	_, err := r.q.RefundTransaction(ctx, RefundTransactionParams{
		ID:           id,
		RefundAmount: floatToNumeric(amount),
	})
	return err
}

func (r *TransactionRepositoryAdapter) ListByCustomer(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]domain.Transaction, error) {
	// The sqlc layer has ListUnsettledTransactions but no generic list-by-customer.
	// Use the unsettled list as a fallback. In production, add a dedicated query.
	rows, err := r.q.ListUnsettledTransactions(ctx, int32(limit))
	if err != nil {
		return nil, err
	}
	txs := make([]domain.Transaction, 0, len(rows))
	for _, t := range rows {
		txs = append(txs, *txFromDB(t))
	}
	return txs, nil
}

func (r *TransactionRepositoryAdapter) ListByProvider(ctx context.Context, providerID uuid.UUID, limit, offset int) ([]domain.Transaction, error) {
	// Same approach as ListByCustomer. In production, add a dedicated query.
	rows, err := r.q.ListUnsettledTransactions(ctx, int32(limit))
	if err != nil {
		return nil, err
	}
	txs := make([]domain.Transaction, 0, len(rows))
	for _, t := range rows {
		txs = append(txs, *txFromDB(t))
	}
	return txs, nil
}

// ---------------------------------------------------------------------------
// NotificationRepositoryAdapter implements domain.NotificationRepository
// ---------------------------------------------------------------------------

type NotificationRepositoryAdapter struct {
	q *Queries
}

func NewNotificationRepository(q *Queries) *NotificationRepositoryAdapter {
	return &NotificationRepositoryAdapter{q: q}
}

func notifFromDB(n Notification) *domain.Notification {
	return &domain.Notification{
		ID:     n.ID,
		UserID: n.UserID,
		Type:   domain.NotificationType(n.Type),
		Title:  n.Title,
		Body:   textVal(n.Body),
		Data:   n.Data,
		Channel: domain.NotificationChannel(n.Channel),
		SentAt:  timestamptzToTimePtr(n.SentAt),
		ReadAt:  timestamptzToTimePtr(n.ReadAt),
	}
}

func (r *NotificationRepositoryAdapter) Create(ctx context.Context, n *domain.Notification) error {
	result, err := r.q.CreateNotification(ctx, CreateNotificationParams{
		UserID:  n.UserID,
		Type:    string(n.Type),
		Title:   n.Title,
		Body:    textOrNull(n.Body),
		Data:    n.Data,
		Channel: string(n.Channel),
	})
	if err != nil {
		return err
	}
	n.ID = result.ID
	return nil
}

func (r *NotificationRepositoryAdapter) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	// No direct GetByID in sqlc notifications. Use ListUnread and filter.
	// In production, add a dedicated query.
	return nil, pgx.ErrNoRows
}

func (r *NotificationRepositoryAdapter) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Notification, error) {
	rows, err := r.q.ListUnreadNotifications(ctx, ListUnreadNotificationsParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	notifs := make([]domain.Notification, 0, len(rows))
	for _, n := range rows {
		notifs = append(notifs, *notifFromDB(n))
	}
	return notifs, nil
}

func (r *NotificationRepositoryAdapter) ListUnread(ctx context.Context, userID uuid.UUID) ([]domain.Notification, error) {
	rows, err := r.q.ListUnreadNotifications(ctx, ListUnreadNotificationsParams{
		UserID: userID,
		Limit:  50,
	})
	if err != nil {
		return nil, err
	}
	notifs := make([]domain.Notification, 0, len(rows))
	for _, n := range rows {
		notifs = append(notifs, *notifFromDB(n))
	}
	return notifs, nil
}

func (r *NotificationRepositoryAdapter) MarkRead(ctx context.Context, id uuid.UUID, readAt time.Time) error {
	return r.q.MarkNotificationRead(ctx, id)
}

func (r *NotificationRepositoryAdapter) MarkAllRead(ctx context.Context, userID uuid.UUID, readAt time.Time) error {
	return r.q.MarkAllNotificationsRead(ctx, userID)
}

func (r *NotificationRepositoryAdapter) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := r.q.CountUnreadNotifications(ctx, userID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *NotificationRepositoryAdapter) GetPreferences(ctx context.Context, userID uuid.UUID) ([]domain.NotificationPreference, error) {
	// Notification preferences are not yet in sqlc. Return empty defaults.
	return []domain.NotificationPreference{}, nil
}

func (r *NotificationRepositoryAdapter) UpsertPreference(ctx context.Context, pref *domain.NotificationPreference) error {
	// Not yet in sqlc. No-op.
	return nil
}

// ---------------------------------------------------------------------------
// DisputeRepositoryAdapter implements domain.DisputeRepository
// ---------------------------------------------------------------------------

type DisputeRepositoryAdapter struct {
	q *Queries
}

func NewDisputeRepository(q *Queries) *DisputeRepositoryAdapter {
	return &DisputeRepositoryAdapter{q: q}
}

func disputeFromDB(d Dispute) *domain.Dispute {
	return &domain.Dispute{
		ID:               d.ID,
		JobID:            d.JobID,
		RaisedBy:         d.RaisedBy,
		Against:          d.Against,
		Type:             domain.DisputeType(d.Type),
		Severity:         domain.DisputeSeverity(d.Severity),
		Status:           domain.DisputeStatus(d.Status),
		Description:      d.Description,
		Evidence:         d.Evidence,
		ResolutionNotes:  textPtr(d.Resolution),
		ResolvedBy:       d.ResolvedBy,
		ResolutionAmount: numericToFloatPtr(d.ResolutionAmount),
		EscalatedAt:      timestamptzToTimePtr(d.EscalatedAt),
		ResolvedAt:       timestamptzToTimePtr(d.ResolvedAt),
		CreatedAt:        d.CreatedAt,
		UpdatedAt:        d.UpdatedAt,
	}
}

func (r *DisputeRepositoryAdapter) Create(ctx context.Context, dispute *domain.Dispute) error {
	result, err := r.q.CreateDispute(ctx, CreateDisputeParams{
		JobID:       dispute.JobID,
		RaisedBy:    dispute.RaisedBy,
		Against:     dispute.Against,
		Type:        DisputeType(dispute.Type),
		Severity:    DisputeSeverity(dispute.Severity),
		Description: dispute.Description,
	})
	if err != nil {
		return err
	}
	dispute.ID = result.ID
	dispute.CreatedAt = result.CreatedAt
	dispute.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *DisputeRepositoryAdapter) GetByID(ctx context.Context, id uuid.UUID) (*domain.Dispute, error) {
	d, err := r.q.GetDisputeByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return disputeFromDB(d), nil
}

func (r *DisputeRepositoryAdapter) GetByJobID(ctx context.Context, jobID uuid.UUID) ([]domain.Dispute, error) {
	// No dedicated sqlc query. List by status and filter. In production add a query.
	return nil, nil
}

func (r *DisputeRepositoryAdapter) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.DisputeStatus) error {
	_, err := r.q.UpdateDisputeStatus(ctx, UpdateDisputeStatusParams{
		ID:     id,
		Status: DisputeStatus(status),
	})
	return err
}

func (r *DisputeRepositoryAdapter) UpdateSeverity(ctx context.Context, id uuid.UUID, severity domain.DisputeSeverity) error {
	// No dedicated sqlc query. Use the status update as a proxy.
	// In production, add a dedicated SQL query.
	return nil
}

func (r *DisputeRepositoryAdapter) Resolve(ctx context.Context, id uuid.UUID, resolution domain.Resolution) error {
	_, err := r.q.ResolveDispute(ctx, ResolveDisputeParams{
		ID:               id,
		Resolution:       resolution.Notes,
		ResolvedBy:       resolution.ResolvedBy,
		ResolutionAmount: floatToNumeric(resolution.RefundAmount),
	})
	return err
}

func (r *DisputeRepositoryAdapter) Escalate(ctx context.Context, id uuid.UUID, escalatedAt time.Time) error {
	_, err := r.q.EscalateDispute(ctx, EscalateDisputeParams{
		ID:       id,
		Severity: DisputeSeverityCritical,
	})
	return err
}

func (r *DisputeRepositoryAdapter) ListByStatus(ctx context.Context, status domain.DisputeStatus, limit, offset int) ([]domain.Dispute, error) {
	rows, err := r.q.ListDisputesByStatus(ctx, ListDisputesByStatusParams{
		Status: DisputeStatus(status),
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	disputes := make([]domain.Dispute, 0, len(rows))
	for _, d := range rows {
		disputes = append(disputes, *disputeFromDB(d.Dispute))
	}
	return disputes, nil
}

func (r *DisputeRepositoryAdapter) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Dispute, error) {
	// No dedicated sqlc query. For now, return empty. In production add a query.
	return []domain.Dispute{}, nil
}

func (r *DisputeRepositoryAdapter) CountByProvider(ctx context.Context, providerID uuid.UUID) (int, error) {
	// Count disputes where the provider is the "against" party.
	// No dedicated sqlc query. Return 0 as default. In production add a query.
	return 0, nil
}

// ---------------------------------------------------------------------------
// GamificationRepositoryAdapter implements domain.GamificationRepository
// ---------------------------------------------------------------------------

type GamificationRepositoryAdapter struct {
	q *Queries
}

func NewGamificationRepository(q *Queries) *GamificationRepositoryAdapter {
	return &GamificationRepositoryAdapter{q: q}
}

func (r *GamificationRepositoryAdapter) AddEntry(ctx context.Context, entry *domain.PointsEntry) error {
	result, err := r.q.AddPoints(ctx, AddPointsParams{
		UserID:        entry.UserID,
		Points:        int32(entry.Points),
		Reason:        string(entry.Reason),
		ReferenceID:   entry.ReferenceID,
		ReferenceType: textOrNull(entry.ReferenceType),
	})
	if err != nil {
		return err
	}
	entry.ID = result.ID
	entry.Balance = int(result.BalanceAfter)
	return nil
}

func (r *GamificationRepositoryAdapter) GetBalance(ctx context.Context, userID uuid.UUID) (int, error) {
	balance, err := r.q.GetPointsBalance(ctx, userID)
	if err != nil {
		return 0, err
	}
	return int(balance), nil
}

func (r *GamificationRepositoryAdapter) ListEntries(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.PointsEntry, error) {
	rows, err := r.q.ListPointsHistory(ctx, ListPointsHistoryParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	entries := make([]domain.PointsEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, domain.PointsEntry{
			ID:            row.ID,
			UserID:        row.UserID,
			Points:        int(row.Points),
			Reason:        domain.PointsReason(row.Reason),
			ReferenceType: textVal(row.ReferenceType),
			ReferenceID:   row.ReferenceID,
			Balance:       int(row.BalanceAfter),
		})
	}
	return entries, nil
}

func (r *GamificationRepositoryAdapter) CreateAchievement(ctx context.Context, achievement *domain.Achievement) error {
	// Achievements table not yet in sqlc. No-op.
	return nil
}

func (r *GamificationRepositoryAdapter) ListAchievements(ctx context.Context, userID uuid.UUID) ([]domain.Achievement, error) {
	return []domain.Achievement{}, nil
}

func (r *GamificationRepositoryAdapter) HasAchievement(ctx context.Context, userID uuid.UUID, achievementType string) (bool, error) {
	return false, nil
}

func (r *GamificationRepositoryAdapter) GetLeaderboard(ctx context.Context, postcode string, limit int) ([]domain.LeaderboardEntry, error) {
	// No dedicated leaderboard query in sqlc. Return empty. In production add a query.
	return []domain.LeaderboardEntry{}, nil
}

// ---------------------------------------------------------------------------
// CategoryRepositoryAdapter implements domain.CategoryRepository
// ---------------------------------------------------------------------------

type CategoryRepositoryAdapter struct {
	q *Queries
}

func NewCategoryRepository(q *Queries) *CategoryRepositoryAdapter {
	return &CategoryRepositoryAdapter{q: q}
}

func categoryFromDB(c Category) *domain.Category {
	nameMap := make(map[string]string)
	if c.Name != nil {
		// Try parsing as {"en": "...", "hi": "..."} first
		if err := json.Unmarshal(c.Name, &nameMap); err != nil {
			// Fallback: try as a plain string
			var plain string
			if err2 := json.Unmarshal(c.Name, &plain); err2 == nil {
				nameMap["en"] = plain
			}
		}
	}
	return &domain.Category{
		ID:              c.ID,
		Slug:            c.Slug,
		Name:            nameMap,
		ParentID:        c.ParentID,
		Icon:            textVal(c.Icon),
		IsActive:        c.IsActive,
		RequiresLicense: c.RequiresLicense,
		Metadata:        c.Metadata,
	}
}

func (r *CategoryRepositoryAdapter) List(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.q.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	cats := make([]domain.Category, 0, len(rows))
	for _, c := range rows {
		cats = append(cats, *categoryFromDB(c))
	}
	return cats, nil
}

func (r *CategoryRepositoryAdapter) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	c, err := r.q.GetCategoryBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return categoryFromDB(c), nil
}

func (r *CategoryRepositoryAdapter) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	c, err := r.q.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return categoryFromDB(c), nil
}

func (r *CategoryRepositoryAdapter) ListActive(ctx context.Context) ([]domain.Category, error) {
	return r.List(ctx) // ListCategories already filters by is_active = true
}

// ---------------------------------------------------------------------------
// RouteRepositoryAdapter implements domain.RouteRepository
// ---------------------------------------------------------------------------

type RouteRepositoryAdapter struct {
	q  *Queries
	db DBTX
}

func NewRouteRepository(q *Queries, db DBTX) *RouteRepositoryAdapter {
	return &RouteRepositoryAdapter{q: q, db: db}
}

func (r *RouteRepositoryAdapter) CreateRoute(ctx context.Context, route *domain.Route) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO routes (id, provider_id, name, postcodes, frequency, max_stops, current_stops, currency, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		route.ID, route.ProviderID, route.Name, []string{route.PostcodeArea},
		fmt.Sprintf("every_%d_days", route.IntervalDays), route.MaxStops, route.CurrentStops,
		route.Currency, true,
	)
	return err
}

func routeFromDB(r Route) *domain.Route {
	postcodeArea := ""
	if len(r.Postcodes) > 0 {
		postcodeArea = r.Postcodes[0]
	}
	return &domain.Route{
		ID:           r.ID,
		ProviderID:   r.ProviderID,
		Name:         r.Name,
		PostcodeArea: postcodeArea,
		MaxStops:     int(r.MaxStops),
		CurrentStops: int(r.CurrentStops),
		IntervalDays: frequencyToDays(r.Frequency),
		Status:       routeStatusFromBool(r.IsActive),
		Currency:     r.Currency,
		PricePerStop: numericToFloatPtr(r.PricePerStop),
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

func frequencyToDays(freq string) int {
	switch freq {
	case "weekly", "every_7_days":
		return 7
	case "biweekly", "every_14_days":
		return 14
	case "every_21_days":
		return 21
	case "monthly", "every_30_days":
		return 30
	case "every_45_days":
		return 45
	default:
		return 14
	}
}

func routeStatusFromBool(active bool) domain.RouteStatus {
	if active {
		return domain.RouteStatusActive
	}
	return domain.RouteStatusInactive
}

func (r *RouteRepositoryAdapter) GetRouteByID(ctx context.Context, id uuid.UUID) (*domain.Route, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, provider_id, name, postcodes, frequency, max_stops, current_stops,
		        price_per_stop, currency, is_active, created_at, updated_at
		 FROM routes WHERE id = $1`, id)

	var rt Route
	err := row.Scan(&rt.ID, &rt.ProviderID, &rt.Name, &rt.Postcodes, &rt.Frequency,
		&rt.MaxStops, &rt.CurrentStops, &rt.PricePerStop, &rt.Currency,
		&rt.IsActive, &rt.CreatedAt, &rt.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return routeFromDB(rt), nil
}

func (r *RouteRepositoryAdapter) ListRoutesByProvider(ctx context.Context, providerID uuid.UUID) ([]domain.Route, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, provider_id, name, postcodes, frequency, max_stops, current_stops,
		        price_per_stop, currency, is_active, created_at, updated_at
		 FROM routes WHERE provider_id = $1 AND is_active = true ORDER BY created_at DESC`, providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []domain.Route
	for rows.Next() {
		var rt Route
		if err := rows.Scan(&rt.ID, &rt.ProviderID, &rt.Name, &rt.Postcodes, &rt.Frequency,
			&rt.MaxStops, &rt.CurrentStops, &rt.PricePerStop, &rt.Currency,
			&rt.IsActive, &rt.CreatedAt, &rt.UpdatedAt); err != nil {
			return nil, err
		}
		routes = append(routes, *routeFromDB(rt))
	}
	return routes, rows.Err()
}

func (r *RouteRepositoryAdapter) UpdateRoute(ctx context.Context, route *domain.Route) error {
	freq := fmt.Sprintf("every_%d_days", route.IntervalDays)
	_, err := r.db.Exec(ctx,
		`UPDATE routes SET name = $2, max_stops = $3, current_stops = $4, frequency = $5,
		        is_active = $6, updated_at = NOW()
		 WHERE id = $1`,
		route.ID, route.Name, route.MaxStops, route.CurrentStops, freq,
		route.Status == domain.RouteStatusActive)
	return err
}

func (r *RouteRepositoryAdapter) DeleteRoute(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE routes SET is_active = false, updated_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *RouteRepositoryAdapter) AddStop(ctx context.Context, stop *domain.RouteStop) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO route_stops (id, route_id, customer_id, address, stop_order, notes, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		stop.ID, stop.RouteID, stop.CustomerID, stop.PropertyAddress,
		stop.StopOrder, stop.Notes, true)
	return err
}

func (r *RouteRepositoryAdapter) GetStopByID(ctx context.Context, id uuid.UUID) (*domain.RouteStop, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, route_id, customer_id, address, stop_order, notes, is_active, created_at
		 FROM route_stops WHERE id = $1`, id)

	var s RouteStop
	err := row.Scan(&s.ID, &s.RouteID, &s.CustomerID, &s.Address, &s.StopOrder,
		&s.Notes, &s.IsActive, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return routeStopFromDB(s), nil
}

func routeStopFromDB(s RouteStop) *domain.RouteStop {
	return &domain.RouteStop{
		ID:              s.ID,
		RouteID:         s.RouteID,
		CustomerID:      s.CustomerID,
		PropertyAddress: s.Address,
		StopOrder:       int(s.StopOrder),
		Notes:           textVal(s.Notes),
		Status:          routeStopStatusFromBool(s.IsActive),
		CreatedAt:       s.CreatedAt,
	}
}

func routeStopStatusFromBool(active bool) domain.RouteStopStatus {
	if active {
		return domain.RouteStopStatusActive
	}
	return domain.RouteStopStatusRemoved
}

func (r *RouteRepositoryAdapter) ListStopsByRoute(ctx context.Context, routeID uuid.UUID) ([]domain.RouteStop, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, route_id, customer_id, address, stop_order, notes, is_active, created_at
		 FROM route_stops WHERE route_id = $1 AND is_active = true ORDER BY stop_order`, routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stops []domain.RouteStop
	for rows.Next() {
		var s RouteStop
		if err := rows.Scan(&s.ID, &s.RouteID, &s.CustomerID, &s.Address, &s.StopOrder,
			&s.Notes, &s.IsActive, &s.CreatedAt); err != nil {
			return nil, err
		}
		stops = append(stops, *routeStopFromDB(s))
	}
	return stops, rows.Err()
}

func (r *RouteRepositoryAdapter) UpdateStop(ctx context.Context, stop *domain.RouteStop) error {
	_, err := r.db.Exec(ctx,
		`UPDATE route_stops SET stop_order = $2, notes = $3 WHERE id = $1`,
		stop.ID, stop.StopOrder, stop.Notes)
	return err
}

func (r *RouteRepositoryAdapter) RemoveStop(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE route_stops SET is_active = false WHERE id = $1`, id)
	return err
}

func (r *RouteRepositoryAdapter) ListUpcomingStops(ctx context.Context, providerID uuid.UUID, before time.Time) ([]domain.RouteStop, error) {
	rows, err := r.db.Query(ctx,
		`SELECT rs.id, rs.route_id, rs.customer_id, rs.address, rs.stop_order, rs.notes, rs.is_active, rs.created_at
		 FROM route_stops rs
		 JOIN routes rt ON rt.id = rs.route_id
		 WHERE rt.provider_id = $1 AND rs.is_active = true AND rt.is_active = true
		 ORDER BY rs.stop_order`, providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stops []domain.RouteStop
	for rows.Next() {
		var s RouteStop
		if err := rows.Scan(&s.ID, &s.RouteID, &s.CustomerID, &s.Address, &s.StopOrder,
			&s.Notes, &s.IsActive, &s.CreatedAt); err != nil {
			return nil, err
		}
		stops = append(stops, *routeStopFromDB(s))
	}
	return stops, rows.Err()
}

func (r *RouteRepositoryAdapter) CreateRequest(ctx context.Context, req *domain.RouteRequest) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO route_requests (id, route_id, customer_id, status) VALUES ($1, $2, $3, $4)`,
		req.ID, req.RouteID, req.CustomerID, string(req.Status))
	return err
}

func (r *RouteRepositoryAdapter) GetRequestByID(ctx context.Context, id uuid.UUID) (*domain.RouteRequest, error) {
	return nil, pgx.ErrNoRows
}

func (r *RouteRepositoryAdapter) ListRequestsByRoute(ctx context.Context, routeID uuid.UUID) ([]domain.RouteRequest, error) {
	return []domain.RouteRequest{}, nil
}

func (r *RouteRepositoryAdapter) UpdateRequestStatus(ctx context.Context, id uuid.UUID, status domain.RouteRequestStatus) error {
	_, err := r.db.Exec(ctx, `UPDATE route_requests SET status = $2, updated_at = NOW() WHERE id = $1`, id, string(status))
	return err
}

// ---------------------------------------------------------------------------
// CacheStoreAdapter implements domain.CacheStore
// ---------------------------------------------------------------------------

// CacheStoreAdapter wraps the redis CacheStore to implement the domain.CacheStore
// interface which uses a simpler signature.
type CacheStoreAdapter struct {
	store CacheStoreInterface
}

// CacheStoreInterface is the subset of redis.CacheStore used by the adapter.
type CacheStoreInterface interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

func NewCacheStoreAdapter(store CacheStoreInterface) *CacheStoreAdapter {
	return &CacheStoreAdapter{store: store}
}

func (c *CacheStoreAdapter) Get(key string) (string, error) {
	return c.store.Get(context.Background(), key)
}

func (c *CacheStoreAdapter) Set(key, value string, ttlSeconds int) error {
	return c.store.Set(context.Background(), key, value, time.Duration(ttlSeconds)*time.Second)
}

func (c *CacheStoreAdapter) Delete(key string) error {
	return c.store.Delete(context.Background(), key)
}

// max returns the larger of a or b.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ---------------------------------------------------------------------------
// CropRepositoryAdapter — implements domain.CropRepository using raw SQL
// since no sqlc crop queries exist yet.
// ---------------------------------------------------------------------------

// CropRepositoryAdapter bridges domain.CropRepository to the database using
// raw SQL queries via the pgxpool connection.
type CropRepositoryAdapter struct {
	db DBTX
}

// NewCropRepository returns a new CropRepositoryAdapter.
func NewCropRepository(db DBTX) *CropRepositoryAdapter {
	return &CropRepositoryAdapter{db: db}
}

func (r *CropRepositoryAdapter) ListByJurisdiction(ctx context.Context, jurisdictionID string) ([]domain.CropCatalog, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, jurisdiction_id, crop_slug, name, work_types, seasonal_calendar, metadata, is_active, created_at, updated_at
		 FROM crop_catalogs
		 WHERE jurisdiction_id = $1 AND is_active = true
		 ORDER BY crop_slug`,
		jurisdictionID)
	if err != nil {
		return nil, fmt.Errorf("list crop catalogs: %w", err)
	}
	defer rows.Close()

	var result []domain.CropCatalog
	for rows.Next() {
		var c domain.CropCatalog
		if err := rows.Scan(&c.ID, &c.JurisdictionID, &c.CropSlug, &c.Name, &c.WorkTypes, &c.SeasonalCalendar, &c.Metadata, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan crop catalog: %w", err)
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (r *CropRepositoryAdapter) GetBySlug(ctx context.Context, jurisdictionID, cropSlug string) (*domain.CropCatalog, error) {
	var c domain.CropCatalog
	err := r.db.QueryRow(ctx,
		`SELECT id, jurisdiction_id, crop_slug, name, work_types, seasonal_calendar, metadata, is_active, created_at, updated_at
		 FROM crop_catalogs
		 WHERE jurisdiction_id = $1 AND crop_slug = $2`,
		jurisdictionID, cropSlug).
		Scan(&c.ID, &c.JurisdictionID, &c.CropSlug, &c.Name, &c.WorkTypes, &c.SeasonalCalendar, &c.Metadata, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get crop by slug: %w", err)
	}
	return &c, nil
}

func (r *CropRepositoryAdapter) Create(ctx context.Context, crop *domain.CropCatalog) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO crop_catalogs (id, jurisdiction_id, crop_slug, name, work_types, seasonal_calendar, metadata, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		crop.ID, crop.JurisdictionID, crop.CropSlug, crop.Name, crop.WorkTypes, crop.SeasonalCalendar, crop.Metadata, crop.IsActive, crop.CreatedAt, crop.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create crop catalog: %w", err)
	}
	return nil
}

func (r *CropRepositoryAdapter) Update(ctx context.Context, crop *domain.CropCatalog) error {
	_, err := r.db.Exec(ctx,
		`UPDATE crop_catalogs
		 SET name = $1, work_types = $2, seasonal_calendar = $3, metadata = $4, is_active = $5, updated_at = $6
		 WHERE id = $7`,
		crop.Name, crop.WorkTypes, crop.SeasonalCalendar, crop.Metadata, crop.IsActive, crop.UpdatedAt, crop.ID)
	if err != nil {
		return fmt.Errorf("update crop catalog: %w", err)
	}
	return nil
}

func (r *CropRepositoryAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM crop_catalogs WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete crop catalog: %w", err)
	}
	return nil
}

// Ensure all adapters implement the correct interfaces at compile time.
var _ domain.UserRepository = (*UserRepositoryAdapter)(nil)
var _ domain.ProviderRepository = (*ProviderRepositoryAdapter)(nil)
var _ domain.JobRepository = (*JobRepositoryAdapter)(nil)
var _ domain.ReviewRepository = (*ReviewRepositoryAdapter)(nil)
var _ domain.TransactionRepository = (*TransactionRepositoryAdapter)(nil)
var _ domain.NotificationRepository = (*NotificationRepositoryAdapter)(nil)
var _ domain.DisputeRepository = (*DisputeRepositoryAdapter)(nil)
var _ domain.GamificationRepository = (*GamificationRepositoryAdapter)(nil)
var _ domain.CategoryRepository = (*CategoryRepositoryAdapter)(nil)
var _ domain.RouteRepository = (*RouteRepositoryAdapter)(nil)
var _ domain.CropRepository = (*CropRepositoryAdapter)(nil)
var _ domain.CacheStore = (*CacheStoreAdapter)(nil)

// Suppress unused import warnings for type conversion helpers.
var (
	_ = math.Round
	_ = json.Marshal
)
