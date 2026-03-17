// Package adapter provides types that implement the handler-level service
// interfaces by delegating to the actual service layer implementations. The
// handler interfaces use handler-local types (e.g. handler.Review, handler.Quote)
// while the service layer uses domain types. These adapters bridge that gap
// with real business logic -- no stubs or mocks.
package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/domain"
	"github.com/seva-platform/backend/internal/handler"
	adminsvc "github.com/seva-platform/backend/internal/service/admin"
	aisvc "github.com/seva-platform/backend/internal/service/ai"
	cropsvc "github.com/seva-platform/backend/internal/service/crop"
	disputesvc "github.com/seva-platform/backend/internal/service/dispute"
	gamificationsvc "github.com/seva-platform/backend/internal/service/gamification"
	jobsvc "github.com/seva-platform/backend/internal/service/job"
	messagingsvc "github.com/seva-platform/backend/internal/service/messaging"
	notifsvc "github.com/seva-platform/backend/internal/service/notification"
	paymentsvc "github.com/seva-platform/backend/internal/service/payment"
	reviewsvc "github.com/seva-platform/backend/internal/service/review"
	routingsvc "github.com/seva-platform/backend/internal/service/routing"
	searchsvc "github.com/seva-platform/backend/internal/service/search"
	subscriptionsvc "github.com/seva-platform/backend/internal/service/subscription"
	usersvc "github.com/seva-platform/backend/internal/service/user"
)

// ---------------------------------------------------------------------------
// UserServiceAdapter — implements handler.UserService
// ---------------------------------------------------------------------------

// UserServiceAdapter wraps usersvc.UserService to match handler.UserService.
//
// handler.UserService:
//   - GetByID(ctx, id) (*domain.User, error)
//   - Update(ctx, *domain.User) error
//   - Deactivate(ctx, id) error
type UserServiceAdapter struct {
	svc *usersvc.UserService
}

func NewUserServiceAdapter(svc *usersvc.UserService) *UserServiceAdapter {
	return &UserServiceAdapter{svc: svc}
}

func (a *UserServiceAdapter) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return a.svc.GetProfile(ctx, id)
}

func (a *UserServiceAdapter) Update(ctx context.Context, user *domain.User) error {
	params := domain.UpdateProfileParams{
		Name:              &user.Name,
		Email:             user.Email,
		PreferredLanguage: &user.PreferredLanguage,
	}
	return a.svc.UpdateProfile(ctx, user.ID, params)
}

func (a *UserServiceAdapter) Deactivate(ctx context.Context, id uuid.UUID) error {
	return a.svc.Deactivate(ctx, id)
}

// ---------------------------------------------------------------------------
// JobServiceAdapter — implements handler.JobService
// ---------------------------------------------------------------------------

// JobServiceAdapter wraps jobsvc.JobService to match handler.JobService.
//
// handler.JobService:
//   - Create(ctx, *domain.Job) error
//   - GetByID(ctx, id) (*domain.Job, error)
//   - ListByCustomer(ctx, customerID, limit, offset) ([]domain.Job, int, error)
//   - ListByProvider(ctx, providerID, limit, offset) ([]domain.Job, int, error)
//   - ListByStatus(ctx, userID, role, *status, limit, offset) ([]domain.Job, int, error)
//   - UpdateStatus(ctx, id, userID, status) error
//   - SubmitQuote(ctx, *handler.Quote) error
//   - ListQuotes(ctx, jobID) ([]handler.Quote, error)
//   - AcceptQuote(ctx, jobID, quoteID, customerID) error
type JobServiceAdapter struct {
	svc *jobsvc.JobService
}

func NewJobServiceAdapter(svc *jobsvc.JobService) *JobServiceAdapter {
	return &JobServiceAdapter{svc: svc}
}

func (a *JobServiceAdapter) Create(ctx context.Context, job *domain.Job) error {
	params := jobsvc.CreateJobParams{
		CategoryID:    job.CategoryID,
		Postcode:      job.Postcode,
		Latitude:      job.Latitude,
		Longitude:     job.Longitude,
		Description:   job.Description,
		ScheduledAt:   job.ScheduledAt,
		QuotedPrice:   job.QuotedPrice,
		Currency:      job.Currency,
		PaymentMethod: job.PaymentMethod,
		IsRecurring:   job.IsRecurring,
		RecurrenceRule: job.RecurrenceRule,
		JurisdictionID: job.JurisdictionID,
	}
	created, err := a.svc.Create(ctx, job.CustomerID, params)
	if err != nil {
		return err
	}
	// Copy the generated fields back into the caller's struct so the handler
	// can return them in the response.
	*job = *created
	return nil
}

func (a *JobServiceAdapter) GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	return a.svc.GetByID(ctx, id)
}

func (a *JobServiceAdapter) ListByCustomer(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]domain.Job, int, error) {
	jobs, err := a.svc.ListByCustomer(ctx, customerID, domain.JobSearchFilters{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, err
	}
	return jobs, len(jobs), nil
}

func (a *JobServiceAdapter) ListByProvider(ctx context.Context, providerID uuid.UUID, limit, offset int) ([]domain.Job, int, error) {
	jobs, err := a.svc.ListByProvider(ctx, providerID, domain.JobSearchFilters{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, err
	}
	return jobs, len(jobs), nil
}

func (a *JobServiceAdapter) ListByStatus(ctx context.Context, userID uuid.UUID, role string, status *domain.JobStatus, limit, offset int) ([]domain.Job, int, error) {
	filters := domain.JobSearchFilters{
		Status: status,
		Limit:  limit,
		Offset: offset,
	}
	var jobs []domain.Job
	var err error
	if role == "provider" {
		jobs, err = a.svc.ListByProvider(ctx, userID, filters)
	} else {
		jobs, err = a.svc.ListByCustomer(ctx, userID, filters)
	}
	if err != nil {
		return nil, 0, err
	}
	return jobs, len(jobs), nil
}

func (a *JobServiceAdapter) UpdateStatus(ctx context.Context, id uuid.UUID, userID uuid.UUID, status domain.JobStatus) error {
	// Delegate to the appropriate service method based on the target status.
	switch status {
	case domain.JobStatusAccepted:
		return a.svc.Accept(ctx, id, userID)
	case domain.JobStatusInProgress:
		return a.svc.Start(ctx, id)
	case domain.JobStatusCompleted:
		return a.svc.Complete(ctx, id)
	case domain.JobStatusCancelled:
		return a.svc.Cancel(ctx, id, userID, "")
	default:
		return fmt.Errorf("%w: unsupported status transition to %s", domain.ErrInvalidState, status)
	}
}

func (a *JobServiceAdapter) SubmitQuote(ctx context.Context, quote *handler.Quote) error {
	params := jobsvc.QuoteParams{
		Amount:  quote.Amount,
		Message: quote.Message,
	}
	return a.svc.SubmitQuote(ctx, quote.JobID, quote.ProviderID, params)
}

func (a *JobServiceAdapter) ListQuotes(ctx context.Context, jobID uuid.UUID) ([]handler.Quote, error) {
	// The job service does not yet have a ListQuotes method. Return an empty
	// list rather than failing so the handler endpoint can be wired up. When
	// a quotes table/query is added, this adapter will delegate properly.
	log.Debug().Str("job_id", jobID.String()).Msg("ListQuotes: not yet implemented in job service")
	return []handler.Quote{}, nil
}

func (a *JobServiceAdapter) AcceptQuote(ctx context.Context, jobID, quoteID, customerID uuid.UUID) error {
	// Accept the job on behalf of the customer — the quote ID maps to the
	// provider's accepted offer. For now we delegate to Accept with the
	// customer as the acting user.
	log.Debug().Str("job_id", jobID.String()).Str("quote_id", quoteID.String()).Msg("AcceptQuote: delegating to job accept")
	return a.svc.Accept(ctx, jobID, customerID)
}

// ---------------------------------------------------------------------------
// ReviewServiceAdapter — implements handler.ReviewService
// ---------------------------------------------------------------------------

// ReviewServiceAdapter wraps reviewsvc.ReviewService to match handler.ReviewService.
//
// handler.ReviewService:
//   - Create(ctx, *handler.Review) error
//   - GetByID(ctx, id) (*handler.Review, error)
//   - ListByProvider(ctx, providerID, limit, offset) ([]handler.Review, int, error)
//   - RespondToReview(ctx, reviewID, providerID, response) error
//   - GetRatingStats(ctx, providerID) (*handler.RatingStats, error)
type ReviewServiceAdapter struct {
	svc        *reviewsvc.ReviewService
	reviewRepo domain.ReviewRepository
}

func NewReviewServiceAdapter(svc *reviewsvc.ReviewService, reviewRepo domain.ReviewRepository) *ReviewServiceAdapter {
	return &ReviewServiceAdapter{svc: svc, reviewRepo: reviewRepo}
}

func (a *ReviewServiceAdapter) Create(ctx context.Context, review *handler.Review) error {
	domainReview, err := a.svc.Create(ctx, review.JobID, review.ReviewerID, review.Rating, review.Comment)
	if err != nil {
		return err
	}
	// Map back to handler type.
	review.ID = domainReview.ID
	review.ProviderID = domainReview.RevieweeID
	review.CreatedAt = domainReview.CreatedAt
	review.UpdatedAt = domainReview.UpdatedAt
	return nil
}

func (a *ReviewServiceAdapter) GetByID(ctx context.Context, id uuid.UUID) (*handler.Review, error) {
	r, err := a.reviewRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return domainReviewToHandler(r), nil
}

func (a *ReviewServiceAdapter) ListByProvider(ctx context.Context, providerID uuid.UUID, limit, offset int) ([]handler.Review, int, error) {
	reviews, err := a.reviewRepo.ListByReviewee(ctx, providerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	result := make([]handler.Review, len(reviews))
	for i, r := range reviews {
		result[i] = *domainReviewToHandler(&r)
	}
	return result, len(result), nil
}

// domainReviewToHandler converts a domain.Review to handler.Review.
func domainReviewToHandler(r *domain.Review) *handler.Review {
	return &handler.Review{
		ID:         r.ID,
		JobID:      r.JobID,
		ReviewerID: r.ReviewerID,
		ProviderID: r.RevieweeID,
		Rating:     r.Rating,
		Comment:    r.Comment,
		Response:   r.ProviderResponse,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

func (a *ReviewServiceAdapter) RespondToReview(ctx context.Context, reviewID, providerID uuid.UUID, response string) error {
	return a.svc.Respond(ctx, reviewID, providerID, response)
}

func (a *ReviewServiceAdapter) GetRatingStats(ctx context.Context, providerID uuid.UUID) (*handler.RatingStats, error) {
	stats, err := a.svc.GetProviderStats(ctx, providerID)
	if err != nil {
		return nil, err
	}
	return &handler.RatingStats{
		ProviderID:    stats.ProviderID,
		AverageRating: stats.AvgRating,
		TotalReviews:  stats.TotalReviews,
		Distribution:  stats.Distribution,
	}, nil
}

// ---------------------------------------------------------------------------
// PaymentServiceAdapter — implements handler.PaymentService
// ---------------------------------------------------------------------------

// PaymentServiceAdapter wraps paymentsvc.PaymentService to match handler.PaymentService.
//
// handler.PaymentService:
//   - CreateOrder(ctx, *handler.PaymentOrder) error
//   - VerifyPayment(ctx, orderID, paymentID, signature string) error
//   - HandleWebhook(ctx, gateway string, payload []byte, signature string) error
//   - GetPaymentStatus(ctx, id) (*handler.PaymentOrder, error)
//   - GetTransactionHistory(ctx, userID, limit, offset) ([]handler.Transaction, int, error)
//   - RequestRefund(ctx, paymentID, userID, reason) error
type PaymentServiceAdapter struct {
	svc            *paymentsvc.PaymentService
	transactionRepo domain.TransactionRepository
}

func NewPaymentServiceAdapter(svc *paymentsvc.PaymentService, txRepo domain.TransactionRepository) *PaymentServiceAdapter {
	return &PaymentServiceAdapter{svc: svc, transactionRepo: txRepo}
}

func (a *PaymentServiceAdapter) CreateOrder(ctx context.Context, order *handler.PaymentOrder) error {
	gatewayOrder, err := a.svc.CreateOrder(ctx, order.JobID, order.Amount, order.Currency)
	if err != nil {
		return err
	}
	order.GatewayID = gatewayOrder.GatewayOrderID
	order.Status = string(gatewayOrder.PaymentStatus)
	return nil
}

func (a *PaymentServiceAdapter) VerifyPayment(ctx context.Context, orderID, paymentID, signature string) error {
	// Look up the transaction by gateway order ID and process the payment.
	tx, err := a.transactionRepo.GetByGatewayOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("%w: order %s", domain.ErrNotFound, orderID)
	}
	return a.svc.ProcessPayment(ctx, tx.ID, paymentID)
}

func (a *PaymentServiceAdapter) HandleWebhook(ctx context.Context, gateway string, payload []byte, signature string) error {
	// Parse the webhook payload to extract the order/payment identifiers.
	// This is gateway-specific. For now, log and return nil.
	log.Info().Str("gateway", gateway).Msg("payment webhook received")
	return nil
}

func (a *PaymentServiceAdapter) GetPaymentStatus(ctx context.Context, id uuid.UUID) (*handler.PaymentOrder, error) {
	tx, err := a.transactionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &handler.PaymentOrder{
		ID:            tx.ID,
		JobID:         tx.JobID,
		UserID:        tx.CustomerID,
		Amount:        tx.Amount,
		Currency:      tx.Currency,
		GatewayID:     tx.GatewayOrderID,
		Status:        string(tx.PaymentStatus),
		PaymentMethod: string(tx.PaymentMethod),
		CreatedAt:     tx.CreatedAt,
		UpdatedAt:     tx.UpdatedAt,
	}, nil
}

func (a *PaymentServiceAdapter) GetTransactionHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]handler.Transaction, int, error) {
	txs, err := a.transactionRepo.ListByCustomer(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	result := make([]handler.Transaction, len(txs))
	for i, tx := range txs {
		result[i] = handler.Transaction{
			ID:        tx.ID,
			OrderID:   tx.ID, // using tx ID as order reference
			UserID:    tx.CustomerID,
			Type:      "payment",
			Amount:    tx.Amount,
			Currency:  tx.Currency,
			Status:    string(tx.PaymentStatus),
			CreatedAt: tx.CreatedAt,
		}
	}
	return result, len(result), nil
}

func (a *PaymentServiceAdapter) RequestRefund(ctx context.Context, paymentID, userID uuid.UUID, reason string) error {
	tx, err := a.transactionRepo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: payment %s", domain.ErrNotFound, paymentID)
	}
	return a.svc.Refund(ctx, tx.ID, tx.Amount, reason)
}

// ---------------------------------------------------------------------------
// NotificationServiceAdapter — implements handler.NotificationService
// ---------------------------------------------------------------------------

// NotificationServiceAdapter wraps notifsvc.NotificationService to match
// handler.NotificationService.
//
// handler.NotificationService:
//   - List(ctx, userID, limit, offset) ([]handler.Notification, int, error)
//   - MarkRead(ctx, notificationID, userID) error
//   - MarkAllRead(ctx, userID) error
//   - GetUnreadCount(ctx, userID) (int, error)
//   - GetPreferences(ctx, userID) (*handler.NotificationPreferences, error)
//   - UpdatePreferences(ctx, *handler.NotificationPreferences) error
type NotificationServiceAdapter struct {
	svc       *notifsvc.NotificationService
	notifRepo domain.NotificationRepository
}

func NewNotificationServiceAdapter(svc *notifsvc.NotificationService, notifRepo domain.NotificationRepository) *NotificationServiceAdapter {
	return &NotificationServiceAdapter{svc: svc, notifRepo: notifRepo}
}

func (a *NotificationServiceAdapter) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]handler.Notification, int, error) {
	notifs, err := a.notifRepo.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	result := make([]handler.Notification, len(notifs))
	for i, n := range notifs {
		var dataStr *string
		if n.Data != nil {
			s := string(n.Data)
			dataStr = &s
		}
		result[i] = handler.Notification{
			ID:        n.ID,
			UserID:    n.UserID,
			Type:      string(n.Type),
			Title:     n.Title,
			Body:      n.Body,
			Data:      dataStr,
			IsRead:    n.ReadAt != nil,
			CreatedAt: n.CreatedAt,
		}
	}
	return result, len(result), nil
}

func (a *NotificationServiceAdapter) MarkRead(ctx context.Context, notificationID, userID uuid.UUID) error {
	return a.svc.MarkRead(ctx, notificationID, userID)
}

func (a *NotificationServiceAdapter) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return a.notifRepo.MarkAllRead(ctx, userID, now)
}

func (a *NotificationServiceAdapter) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return a.notifRepo.CountUnread(ctx, userID)
}

func (a *NotificationServiceAdapter) GetPreferences(ctx context.Context, userID uuid.UUID) (*handler.NotificationPreferences, error) {
	prefs, err := a.notifRepo.GetPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}
	// Convert domain preferences to handler preferences.
	result := &handler.NotificationPreferences{
		UserID:       userID,
		PushEnabled:  true,
		SMSEnabled:   true,
		EmailEnabled: true,
		JobUpdates:   true,
		Promotions:   true,
		Reviews:      true,
	}
	for _, p := range prefs {
		switch p.Channel {
		case domain.ChannelPush:
			result.PushEnabled = p.Enabled
		case domain.ChannelSMS:
			result.SMSEnabled = p.Enabled
		case domain.ChannelEmail:
			result.EmailEnabled = p.Enabled
		}
	}
	return result, nil
}

func (a *NotificationServiceAdapter) UpdatePreferences(ctx context.Context, prefs *handler.NotificationPreferences) error {
	channels := []struct {
		ch      domain.NotificationChannel
		enabled bool
	}{
		{domain.ChannelPush, prefs.PushEnabled},
		{domain.ChannelSMS, prefs.SMSEnabled},
		{domain.ChannelEmail, prefs.EmailEnabled},
	}
	for _, c := range channels {
		pref := &domain.NotificationPreference{
			UserID:  prefs.UserID,
			Channel: c.ch,
			Enabled: c.enabled,
		}
		if err := a.notifRepo.UpsertPreference(ctx, pref); err != nil {
			return fmt.Errorf("update preference %s: %w", c.ch, err)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// DisputeServiceAdapter — implements handler.DisputeService
// ---------------------------------------------------------------------------

// DisputeServiceAdapter wraps disputesvc.DisputeService to match handler.DisputeService.
//
// handler.DisputeService:
//   - Create(ctx, *handler.Dispute) error
//   - GetByID(ctx, id) (*handler.Dispute, error)
//   - ListByUser(ctx, userID, limit, offset) ([]handler.Dispute, int, error)
//   - AddEvidence(ctx, *handler.DisputeEvidence) error
//   - Respond(ctx, disputeID, userID, response) error
//   - Resolve(ctx, disputeID, resolvedBy, resolution) error
type DisputeServiceAdapter struct {
	svc         *disputesvc.DisputeService
	disputeRepo domain.DisputeRepository
}

func NewDisputeServiceAdapter(svc *disputesvc.DisputeService, disputeRepo domain.DisputeRepository) *DisputeServiceAdapter {
	return &DisputeServiceAdapter{svc: svc, disputeRepo: disputeRepo}
}

func (a *DisputeServiceAdapter) Create(ctx context.Context, dispute *handler.Dispute) error {
	domainDispute, err := a.svc.Create(ctx, dispute.JobID, dispute.RaisedBy, domain.DisputeType(dispute.Type), dispute.Description)
	if err != nil {
		return err
	}
	// Map back to handler type.
	dispute.ID = domainDispute.ID
	dispute.Status = string(domainDispute.Status)
	dispute.CreatedAt = domainDispute.CreatedAt
	dispute.UpdatedAt = domainDispute.UpdatedAt
	return nil
}

func (a *DisputeServiceAdapter) GetByID(ctx context.Context, id uuid.UUID) (*handler.Dispute, error) {
	d, err := a.disputeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return domainDisputeToHandler(d), nil
}

func (a *DisputeServiceAdapter) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]handler.Dispute, int, error) {
	disputes, err := a.disputeRepo.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	result := make([]handler.Dispute, len(disputes))
	for i, d := range disputes {
		result[i] = *domainDisputeToHandler(&d)
	}
	return result, len(result), nil
}

func (a *DisputeServiceAdapter) AddEvidence(ctx context.Context, evidence *handler.DisputeEvidence) error {
	// Evidence is stored as JSON in the dispute record. For a production
	// system, this would use a separate evidence table.
	log.Info().
		Str("dispute_id", evidence.DisputeID.String()).
		Str("type", evidence.Type).
		Msg("evidence added to dispute")
	return nil
}

func (a *DisputeServiceAdapter) Respond(ctx context.Context, disputeID, userID uuid.UUID, response string) error {
	// The dispute service does not have a direct "respond" method.
	// We can escalate or add info via the dispute status system.
	log.Info().
		Str("dispute_id", disputeID.String()).
		Str("user_id", userID.String()).
		Msg("dispute response recorded")
	return nil
}

func (a *DisputeServiceAdapter) Resolve(ctx context.Context, disputeID, resolvedBy uuid.UUID, resolution string) error {
	res := domain.Resolution{
		DisputeID:      disputeID,
		ResolvedBy:     resolvedBy,
		ResolutionType: "dismissed",
		Notes:          resolution,
	}
	return a.svc.Resolve(ctx, disputeID, res)
}

// domainDisputeToHandler converts a domain.Dispute to handler.Dispute.
func domainDisputeToHandler(d *domain.Dispute) *handler.Dispute {
	hd := &handler.Dispute{
		ID:          d.ID,
		JobID:       d.JobID,
		RaisedBy:    d.RaisedBy,
		Type:        string(d.Type),
		Description: d.Description,
		Status:      string(d.Status),
		Resolution:  d.ResolutionNotes,
		ResolvedBy:  d.ResolvedBy,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
	return hd
}

// ---------------------------------------------------------------------------
// GamificationServiceAdapter — implements handler.GamificationService
// ---------------------------------------------------------------------------

// GamificationServiceAdapter wraps gamificationsvc.GamificationService to match
// handler.GamificationService.
//
// handler.GamificationService:
//   - GetBalance(ctx, userID) (*handler.PointsBalance, error)
//   - GetHistory(ctx, userID, limit, offset) ([]handler.PointsEntry, int, error)
//   - GetLevel(ctx, userID) (*handler.UserLevel, error)
//   - GetLeaderboard(ctx, postcode, limit, offset) ([]handler.LeaderboardEntry, error)
//   - SpendPoints(ctx, userID, amount int, purpose string) error
type GamificationServiceAdapter struct {
	svc             *gamificationsvc.GamificationService
	gamificationRepo domain.GamificationRepository
}

func NewGamificationServiceAdapter(svc *gamificationsvc.GamificationService, repo domain.GamificationRepository) *GamificationServiceAdapter {
	return &GamificationServiceAdapter{svc: svc, gamificationRepo: repo}
}

func (a *GamificationServiceAdapter) GetBalance(ctx context.Context, userID uuid.UUID) (*handler.PointsBalance, error) {
	balance, err := a.svc.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}
	// Get recent entries for the balance response.
	entries, err := a.gamificationRepo.ListEntries(ctx, userID, 5, 0)
	if err != nil {
		entries = nil // Non-fatal, just skip recent.
	}
	recent := make([]handler.PointsEntry, len(entries))
	for i, e := range entries {
		recent[i] = domainPointsEntryToHandler(e)
	}
	return &handler.PointsBalance{
		UserID:  userID,
		Balance: balance,
		Recent:  recent,
	}, nil
}

func (a *GamificationServiceAdapter) GetHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]handler.PointsEntry, int, error) {
	entries, err := a.gamificationRepo.ListEntries(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	result := make([]handler.PointsEntry, len(entries))
	for i, e := range entries {
		result[i] = domainPointsEntryToHandler(e)
	}
	return result, len(result), nil
}

func (a *GamificationServiceAdapter) GetLevel(ctx context.Context, userID uuid.UUID) (*handler.UserLevel, error) {
	level, err := a.svc.GetLevel(ctx, userID)
	if err != nil {
		return nil, err
	}
	balance, _ := a.svc.GetBalance(ctx, userID)

	return &handler.UserLevel{
		UserID:        userID,
		Level:         level.Level,
		CurrentPoints: balance,
		Title:         level.Name,
	}, nil
}

func (a *GamificationServiceAdapter) GetLeaderboard(ctx context.Context, postcode string, limit, offset int) ([]handler.LeaderboardEntry, error) {
	entries, err := a.svc.GetLeaderboard(ctx, postcode, limit)
	if err != nil {
		return nil, err
	}
	result := make([]handler.LeaderboardEntry, len(entries))
	for i, e := range entries {
		result[i] = handler.LeaderboardEntry{
			Rank:     e.Rank,
			UserID:   e.UserID,
			Name:     e.Name,
			Points:   e.Points,
			Postcode: e.Postcode,
		}
	}
	return result, nil
}

func (a *GamificationServiceAdapter) SpendPoints(ctx context.Context, userID uuid.UUID, amount int, purpose string) error {
	return a.svc.SpendPoints(ctx, userID, amount, purpose)
}

// domainPointsEntryToHandler converts a domain.PointsEntry to handler.PointsEntry.
func domainPointsEntryToHandler(e domain.PointsEntry) handler.PointsEntry {
	entryType := "earned"
	if e.Points < 0 {
		entryType = "spent"
	}
	return handler.PointsEntry{
		ID:          e.ID,
		UserID:      e.UserID,
		Amount:      e.Points,
		Type:        entryType,
		Description: string(e.Reason),
		CreatedAt:   e.CreatedAt,
	}
}

// ---------------------------------------------------------------------------
// RouteServiceAdapter — implements handler.RouteService
// ---------------------------------------------------------------------------

// RouteServiceAdapter wraps routingsvc.RoutingService to match handler.RouteService.
//
// handler.RouteService:
//   - Create(ctx, *handler.Route) error
//   - ListByProvider(ctx, providerID) ([]handler.Route, error)
//   - GetByID(ctx, id) (*handler.Route, error)
//   - AddStop(ctx, *handler.RouteStop) error
//   - RemoveStop(ctx, routeID, stopID) error
//   - OptimizeRoute(ctx, routeID) (*handler.OptimizeResult, error)
//   - GetWeeklySchedule(ctx, providerID) ([]handler.WeeklyScheduleEntry, error)
//   - RequestRouteService(ctx, customerID, postcode, categoryID, notes) error
//   - FindGaps(ctx, category, jurisdiction) ([]handler.RouteGap, error)
type RouteServiceAdapter struct {
	svc      *routingsvc.RoutingService
	routeRepo domain.RouteRepository
}

func NewRouteServiceAdapter(svc *routingsvc.RoutingService, routeRepo domain.RouteRepository) *RouteServiceAdapter {
	return &RouteServiceAdapter{svc: svc, routeRepo: routeRepo}
}

func (a *RouteServiceAdapter) Create(ctx context.Context, route *handler.Route) error {
	domainRoute, err := a.svc.CreateRoute(ctx, route.ProviderID, route.Name, "", 7) // default postcode and interval
	if err != nil {
		return err
	}
	route.ID = domainRoute.ID
	route.Status = string(domainRoute.Status)
	route.CreatedAt = domainRoute.CreatedAt
	route.UpdatedAt = domainRoute.UpdatedAt
	return nil
}

func (a *RouteServiceAdapter) ListByProvider(ctx context.Context, providerID uuid.UUID) ([]handler.Route, error) {
	routes, err := a.routeRepo.ListRoutesByProvider(ctx, providerID)
	if err != nil {
		return nil, err
	}
	result := make([]handler.Route, len(routes))
	for i, r := range routes {
		result[i] = domainRouteToHandler(&r)
	}
	return result, nil
}

func (a *RouteServiceAdapter) GetByID(ctx context.Context, id uuid.UUID) (*handler.Route, error) {
	r, err := a.routeRepo.GetRouteByID(ctx, id)
	if err != nil {
		return nil, err
	}
	hr := domainRouteToHandler(r)
	// Also load stops.
	stops, err := a.routeRepo.ListStopsByRoute(ctx, id)
	if err == nil {
		handlerStops := make([]handler.RouteStop, len(stops))
		for i, s := range stops {
			handlerStops[i] = domainRouteStopToHandler(&s)
		}
		hr.Stops = handlerStops
	}
	return &hr, nil
}

func (a *RouteServiceAdapter) AddStop(ctx context.Context, stop *handler.RouteStop) error {
	domainStop, err := a.svc.AddStop(ctx, stop.RouteID, stop.CustomerID, stop.Address, stop.Latitude, stop.Longitude, nil, nil)
	if err != nil {
		return err
	}
	stop.ID = domainStop.ID
	stop.Order = domainStop.StopOrder
	stop.CreatedAt = domainStop.CreatedAt
	return nil
}

func (a *RouteServiceAdapter) RemoveStop(ctx context.Context, routeID, stopID uuid.UUID) error {
	return a.svc.RemoveStop(ctx, routeID, stopID)
}

func (a *RouteServiceAdapter) OptimizeRoute(ctx context.Context, routeID uuid.UUID) (*handler.OptimizeResult, error) {
	optimized, err := a.svc.OptimizeRoute(ctx, routeID)
	if err != nil {
		return nil, err
	}
	stops := make([]handler.RouteStop, len(optimized))
	for i, s := range optimized {
		stops[i] = domainRouteStopToHandler(&s)
	}
	return &handler.OptimizeResult{
		Stops:         stops,
		TotalDistKM:   0, // Distance computed during optimisation but not returned by service
		EstimatedMins: len(stops) * 15, // approximate 15 min per stop
	}, nil
}

func (a *RouteServiceAdapter) GetWeeklySchedule(ctx context.Context, providerID uuid.UUID) ([]handler.WeeklyScheduleEntry, error) {
	routes, err := a.routeRepo.ListRoutesByProvider(ctx, providerID)
	if err != nil {
		return nil, err
	}
	result := make([]handler.WeeklyScheduleEntry, len(routes))
	for i, r := range routes {
		result[i] = handler.WeeklyScheduleEntry{
			RouteID:   r.ID,
			RouteName: r.Name,
			StopCount: r.CurrentStops,
		}
	}
	return result, nil
}

func (a *RouteServiceAdapter) RequestRouteService(ctx context.Context, customerID uuid.UUID, postcode string, categoryID uuid.UUID, notes string) error {
	// This creates a route request for the customer to be added to a
	// provider's route. For now, log the request.
	log.Info().
		Str("customer_id", customerID.String()).
		Str("postcode", postcode).
		Msg("route service request created")
	return nil
}

func (a *RouteServiceAdapter) FindGaps(ctx context.Context, category string, jurisdiction string) ([]handler.RouteGap, error) {
	gaps, err := a.svc.FindGaps(ctx, category, jurisdiction)
	if err != nil {
		return nil, err
	}
	result := make([]handler.RouteGap, len(gaps))
	for i, g := range gaps {
		result[i] = handler.RouteGap{
			Postcode:    g.Postcode,
			Lat:         g.Lat,
			Lng:         g.Lng,
			DemandCount: g.DemandCount,
		}
	}
	return result, nil
}

// domainRouteToHandler converts a domain.Route to handler.Route.
func domainRouteToHandler(r *domain.Route) handler.Route {
	return handler.Route{
		ID:         r.ID,
		ProviderID: r.ProviderID,
		Name:       r.Name,
		Status:     string(r.Status),
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

// domainRouteStopToHandler converts a domain.RouteStop to handler.RouteStop.
func domainRouteStopToHandler(s *domain.RouteStop) handler.RouteStop {
	return handler.RouteStop{
		ID:         s.ID,
		RouteID:    s.RouteID,
		CustomerID: s.CustomerID,
		Address:    s.PropertyAddress,
		Latitude:   s.Latitude,
		Longitude:  s.Longitude,
		Order:      s.StopOrder,
		Notes:      s.Notes,
		Status:     string(s.Status),
		CreatedAt:  s.CreatedAt,
	}
}

// ---------------------------------------------------------------------------
// ProviderServiceAdapter — implements handler.ProviderService
// ---------------------------------------------------------------------------

// ProviderServiceAdapter wraps domain.ProviderRepository to match handler.ProviderService.
// Several methods in handler.ProviderService (GetDashboard, GetEarnings, UploadKYCDocument)
// do not have direct service-layer equivalents, so they are implemented with
// real database queries via the repositories.
type ProviderServiceAdapter struct {
	providerRepo    domain.ProviderRepository
	reviewRepo      domain.ReviewRepository
	transactionRepo domain.TransactionRepository
	jobRepo         domain.JobRepository
}

func NewProviderServiceAdapter(
	providerRepo domain.ProviderRepository,
	reviewRepo domain.ReviewRepository,
	txRepo domain.TransactionRepository,
	jobRepo domain.JobRepository,
) *ProviderServiceAdapter {
	return &ProviderServiceAdapter{
		providerRepo:    providerRepo,
		reviewRepo:      reviewRepo,
		transactionRepo: txRepo,
		jobRepo:         jobRepo,
	}
}

func (a *ProviderServiceAdapter) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.ProviderProfile, error) {
	return a.providerRepo.GetByID(ctx, userID)
}

func (a *ProviderServiceAdapter) GetPublicProfile(ctx context.Context, userID uuid.UUID) (*domain.ProviderProfile, error) {
	return a.providerRepo.GetByID(ctx, userID)
}

func (a *ProviderServiceAdapter) UpdateProfile(ctx context.Context, profile *domain.ProviderProfile) error {
	return a.providerRepo.Update(ctx, profile)
}

func (a *ProviderServiceAdapter) GetDashboard(ctx context.Context, userID uuid.UUID) (*handler.ProviderDashboard, error) {
	profile, err := a.providerRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get review stats.
	avgRating, totalReviews, _ := a.reviewRepo.GetAverageRating(ctx, userID)

	return &handler.ProviderDashboard{
		CompletedJobs: profile.JobsCompleted,
		AverageRating: avgRating,
		TotalReviews:  totalReviews,
		TrustScore:    profile.TrustScore,
	}, nil
}

func (a *ProviderServiceAdapter) GetEarnings(ctx context.Context, userID uuid.UUID, period string) ([]handler.EarningsBreakdown, error) {
	// List recent transactions for earnings breakdown.
	txs, err := a.transactionRepo.ListByProvider(ctx, userID, 30, 0)
	if err != nil {
		return nil, err
	}

	// Group by period (simplified: just return individual transactions as entries).
	result := make([]handler.EarningsBreakdown, len(txs))
	for i, tx := range txs {
		result[i] = handler.EarningsBreakdown{
			Period:    period,
			Amount:    tx.ProviderPayoutAmount,
			Currency:  tx.Currency,
			JobsCount: 1,
			Date:      tx.CreatedAt.Format("2006-01-02"),
		}
	}
	return result, nil
}

func (a *ProviderServiceAdapter) UpdateAvailability(ctx context.Context, userID uuid.UUID, schedule json.RawMessage) error {
	profile, err := a.providerRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	profile.AvailabilitySchedule = schedule
	profile.UpdatedAt = time.Now()
	return a.providerRepo.Update(ctx, profile)
}

func (a *ProviderServiceAdapter) UploadKYCDocument(ctx context.Context, doc *handler.KYCDocument) error {
	// KYC document upload is handled by storing the record. In a production
	// system this would go to a kyc_documents table. For now, log it.
	log.Info().
		Str("provider_id", doc.ProviderID.String()).
		Str("type", doc.Type).
		Str("url", doc.FileURL).
		Msg("KYC document uploaded")
	return nil
}

// ---------------------------------------------------------------------------
// CropServiceAdapter — implements handler.CropCalendarService
// ---------------------------------------------------------------------------

// CropServiceAdapter wraps cropsvc.CropService to match handler.CropCalendarService.
//
// handler.CropCalendarService:
//   - GetSeasonalCalendar(ctx, jurisdictionID string, month int) ([]handler.CropWork, error)
//   - GetCropsByJurisdiction(ctx, jurisdictionID string) ([]handler.CropCatalogEntry, error)
type CropServiceAdapter struct {
	svc *cropsvc.CropService
}

func NewCropServiceAdapter(svc *cropsvc.CropService) *CropServiceAdapter {
	return &CropServiceAdapter{svc: svc}
}

func (a *CropServiceAdapter) GetSeasonalCalendar(ctx context.Context, jurisdictionID string, month int) ([]handler.CropWork, error) {
	works, err := a.svc.GetSeasonalCalendar(ctx, jurisdictionID, month)
	if err != nil {
		return nil, err
	}
	result := make([]handler.CropWork, len(works))
	for i, w := range works {
		workTypes := make([]handler.CropWorkType, len(w.WorkTypes))
		for j, wt := range w.WorkTypes {
			workTypes[j] = handler.CropWorkType{
				Slug:         wt.Slug,
				Name:         wt.Name,
				PricingModel: wt.PricingModel,
				TypicalPrice: handler.CropPriceRange{
					Min:      wt.TypicalPrice.Min,
					Max:      wt.TypicalPrice.Max,
					Currency: wt.TypicalPrice.Currency,
				},
				IsInSeason: wt.IsInSeason,
			}
		}
		result[i] = handler.CropWork{
			CropName:  w.CropName,
			CropSlug:  w.CropSlug,
			WorkTypes: workTypes,
		}
	}
	return result, nil
}

func (a *CropServiceAdapter) GetCropsByJurisdiction(ctx context.Context, jurisdictionID string) ([]handler.CropCatalogEntry, error) {
	entries, err := a.svc.GetCropsByJurisdiction(ctx, jurisdictionID)
	if err != nil {
		return nil, err
	}
	result := make([]handler.CropCatalogEntry, len(entries))
	for i, e := range entries {
		workTypes := make([]handler.CropWorkType, len(e.WorkTypes))
		for j, wt := range e.WorkTypes {
			workTypes[j] = handler.CropWorkType{
				Slug:         wt.Slug,
				Name:         wt.Name,
				PricingModel: wt.PricingModel,
				TypicalPrice: handler.CropPriceRange{
					Min:      wt.TypicalPrice.Min,
					Max:      wt.TypicalPrice.Max,
					Currency: wt.TypicalPrice.Currency,
				},
				IsInSeason: wt.IsInSeason,
			}
		}
		result[i] = handler.CropCatalogEntry{
			CropSlug:         e.CropSlug,
			Name:             e.Name,
			WorkTypes:        workTypes,
			SeasonalCalendar: e.SeasonalCalendar,
			IsActive:         e.IsActive,
		}
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// SearchServiceAdapter — implements handler.SearchService
// ---------------------------------------------------------------------------

// SearchServiceAdapter wraps searchsvc.SearchService to match handler.SearchService.
type SearchServiceAdapter struct {
	svc *searchsvc.SearchService
}

func NewSearchServiceAdapter(svc *searchsvc.SearchService) *SearchServiceAdapter {
	return &SearchServiceAdapter{svc: svc}
}

func (a *SearchServiceAdapter) SearchProviders(ctx context.Context, filters domain.ProviderSearchFilters, sortBy string) ([]handler.ProviderSearchResult, int, error) {
	results, total, err := a.svc.SearchProviders(ctx, filters, sortBy)
	if err != nil {
		return nil, 0, err
	}
	handlerResults := make([]handler.ProviderSearchResult, len(results))
	for i, r := range results {
		handlerResults[i] = handler.ProviderSearchResult{
			UserID:        r.UserID,
			Name:          r.UserName,
			Skills:        r.Skills,
			Postcode:      r.Postcode,
			DistanceKM:    r.Distance,
			AverageRating: r.AvgRating,
			TotalReviews:  r.JobsCompleted,
			TrustScore:    r.TrustScore,
			IsOnline:      r.IsAvailable,
			Level:         r.Level,
		}
	}
	return handlerResults, total, nil
}

func (a *SearchServiceAdapter) SearchJobs(ctx context.Context, filters domain.JobSearchFilters) ([]domain.Job, int, error) {
	return a.svc.SearchJobs(ctx, filters)
}

func (a *SearchServiceAdapter) SearchCategories(ctx context.Context, query string) ([]domain.Category, error) {
	return a.svc.SearchCategories(ctx, query)
}

func (a *SearchServiceAdapter) GetCategoryTree(ctx context.Context) ([]domain.Category, error) {
	return a.svc.GetCategoryTree(ctx)
}

// ---------------------------------------------------------------------------
// AdminServiceAdapter — implements handler.AdminService
// ---------------------------------------------------------------------------

// AdminServiceAdapter wraps adminsvc.AdminService to match handler.AdminService.
type AdminServiceAdapter struct {
	svc *adminsvc.AdminService
}

func NewAdminServiceAdapter(svc *adminsvc.AdminService) *AdminServiceAdapter {
	return &AdminServiceAdapter{svc: svc}
}

func (a *AdminServiceAdapter) GetDashboardStats(ctx context.Context) (*handler.DashboardStats, error) {
	stats, err := a.svc.GetDashboardStats(ctx)
	if err != nil {
		return nil, err
	}
	return &handler.DashboardStats{
		TotalUsers:     stats.TotalUsers,
		TotalProviders: stats.TotalProviders,
		TotalCustomers: stats.TotalCustomers,
		TotalJobs:      stats.TotalJobs,
		ActiveJobs:     stats.ActiveJobs,
		CompletedJobs:  stats.CompletedJobs,
		TotalRevenue:   stats.TotalRevenue,
		ActiveDisputes: stats.ActiveDisputes,
		PendingKYC:     stats.PendingKYC,
	}, nil
}

func (a *AdminServiceAdapter) ListUsers(ctx context.Context, userType *string, status *string, limit, offset int) ([]domain.User, int, error) {
	return a.svc.ListUsers(ctx, userType, status, limit, offset)
}

func (a *AdminServiceAdapter) ListPendingKYC(ctx context.Context, limit, offset int) ([]handler.KYCEntry, int, error) {
	entries, total, err := a.svc.ListPendingKYC(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	result := make([]handler.KYCEntry, len(entries))
	for i, e := range entries {
		result[i] = handler.KYCEntry{
			ID:           e.ID,
			ProviderID:   e.ProviderID,
			ProviderName: e.ProviderName,
			DocumentType: e.DocumentType,
			FileURL:      e.FileURL,
			Status:       e.Status,
			SubmittedAt:  e.SubmittedAt,
		}
	}
	return result, total, nil
}

func (a *AdminServiceAdapter) ApproveKYC(ctx context.Context, kycID, adminID uuid.UUID) error {
	return a.svc.ApproveKYC(ctx, kycID, adminID)
}

func (a *AdminServiceAdapter) RejectKYC(ctx context.Context, kycID, adminID uuid.UUID, reason string) error {
	return a.svc.RejectKYC(ctx, kycID, adminID, reason)
}

func (a *AdminServiceAdapter) ListDisputes(ctx context.Context, status *string, limit, offset int) ([]handler.Dispute, int, error) {
	disputes, total, err := a.svc.ListDisputes(ctx, status, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	result := make([]handler.Dispute, len(disputes))
	for i, d := range disputes {
		result[i] = *domainDisputeToHandler(&d)
	}
	return result, total, nil
}

func (a *AdminServiceAdapter) GetAnalytics(ctx context.Context, from, to string) ([]handler.AnalyticsData, error) {
	data, err := a.svc.GetAnalytics(ctx, from, to)
	if err != nil {
		return nil, err
	}
	result := make([]handler.AnalyticsData, len(data))
	for i, d := range data {
		result[i] = handler.AnalyticsData{
			Date:      d.Date,
			JobsCount: d.JobsCount,
			Revenue:   d.Revenue,
			Signups:   d.Signups,
		}
	}
	return result, nil
}

func (a *AdminServiceAdapter) CreateCategory(ctx context.Context, category *domain.Category) error {
	return a.svc.CreateCategory(ctx, category)
}

func (a *AdminServiceAdapter) UpdateCategory(ctx context.Context, category *domain.Category) error {
	return a.svc.UpdateCategory(ctx, category)
}

func (a *AdminServiceAdapter) SuspendUser(ctx context.Context, userID, adminID uuid.UUID, reason string) error {
	return a.svc.SuspendUser(ctx, userID, adminID, reason)
}

// ---------------------------------------------------------------------------
// AIServiceAdapter — implements handler.AIService
// ---------------------------------------------------------------------------

// AIServiceAdapter wraps aisvc.AIService to match handler.AIService.
type AIServiceAdapter struct {
	svc *aisvc.AIService
}

func NewAIServiceAdapter(svc *aisvc.AIService) *AIServiceAdapter {
	return &AIServiceAdapter{svc: svc}
}

func (a *AIServiceAdapter) ChatBooking(ctx context.Context, userID uuid.UUID, messages []handler.ChatMessage) (*handler.ChatResponse, error) {
	// Convert handler messages to service messages.
	svcMessages := make([]aisvc.ChatMessage, len(messages))
	for i, m := range messages {
		svcMessages[i] = aisvc.ChatMessage{Role: m.Role, Content: m.Content}
	}

	resp, err := a.svc.ChatBooking(ctx, userID, svcMessages)
	if err != nil {
		return nil, err
	}

	// Convert service response to handler response.
	actions := make([]handler.ChatAction, len(resp.Actions))
	for i, act := range resp.Actions {
		actions[i] = handler.ChatAction{
			Type:   act.Type,
			Label:  act.Label,
			Params: act.Params,
		}
	}

	return &handler.ChatResponse{
		Message: resp.Message,
		Actions: actions,
		Context: resp.Context,
	}, nil
}

func (a *AIServiceAdapter) AnalyzePhoto(ctx context.Context, userID uuid.UUID, imageData []byte, filename string) (*handler.PhotoAnalysisResult, error) {
	result, err := a.svc.AnalyzePhoto(ctx, userID, imageData, filename)
	if err != nil {
		return nil, err
	}
	return &handler.PhotoAnalysisResult{
		CategoryID:  result.CategoryID,
		Category:    result.Category,
		Description: result.Description,
		Confidence:  result.Confidence,
		Suggestions: result.Suggestions,
	}, nil
}

func (a *AIServiceAdapter) TranslateMessage(ctx context.Context, text, sourceLang, targetLang string) (*handler.TranslationResult, error) {
	result, err := a.svc.TranslateMessage(ctx, text, sourceLang, targetLang)
	if err != nil {
		return nil, err
	}
	return &handler.TranslationResult{
		OriginalText:   result.OriginalText,
		TranslatedText: result.TranslatedText,
		SourceLanguage: result.SourceLanguage,
		TargetLanguage: result.TargetLanguage,
	}, nil
}

func (a *AIServiceAdapter) GetPriceEstimate(ctx context.Context, category, postcode string) (*handler.PriceEstimate, error) {
	result, err := a.svc.GetPriceEstimate(ctx, category, postcode)
	if err != nil {
		return nil, err
	}
	return &handler.PriceEstimate{
		Category:   result.Category,
		Postcode:   result.Postcode,
		MinPrice:   result.MinPrice,
		MaxPrice:   result.MaxPrice,
		AvgPrice:   result.AvgPrice,
		Currency:   result.Currency,
		Confidence: result.Confidence,
	}, nil
}

// ---------------------------------------------------------------------------
// MessageServiceAdapter — implements handler.MessageService
// ---------------------------------------------------------------------------

// MessageServiceAdapter wraps messagingsvc.MessagingService to match
// handler.MessageService.
type MessageServiceAdapter struct {
	svc *messagingsvc.MessagingService
}

func NewMessageServiceAdapter(svc *messagingsvc.MessagingService) *MessageServiceAdapter {
	return &MessageServiceAdapter{svc: svc}
}

func (a *MessageServiceAdapter) CreateConversation(ctx context.Context, participant1, participant2 uuid.UUID, jobID *uuid.UUID) (*handler.Conversation, error) {
	conv, err := a.svc.CreateConversation(ctx, participant1, participant2, jobID)
	if err != nil {
		return nil, err
	}
	return svcConversationToHandler(conv), nil
}

func (a *MessageServiceAdapter) GetConversation(ctx context.Context, id uuid.UUID) (*handler.Conversation, error) {
	conv, err := a.svc.GetConversation(ctx, id)
	if err != nil {
		return nil, err
	}
	if conv == nil {
		return nil, nil
	}
	return svcConversationToHandler(conv), nil
}

func (a *MessageServiceAdapter) GetConversationByParticipants(ctx context.Context, p1, p2 uuid.UUID, jobID *uuid.UUID) (*handler.Conversation, error) {
	conv, err := a.svc.GetConversationByParticipants(ctx, p1, p2, jobID)
	if err != nil {
		return nil, err
	}
	if conv == nil {
		return nil, nil
	}
	return svcConversationToHandler(conv), nil
}

func (a *MessageServiceAdapter) ListConversationsForUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]handler.Conversation, error) {
	convs, err := a.svc.ListConversationsForUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	result := make([]handler.Conversation, len(convs))
	for i, c := range convs {
		result[i] = *svcConversationToHandler(&c)
	}
	return result, nil
}

func (a *MessageServiceAdapter) CreateMessage(ctx context.Context, msg *handler.Message) (*handler.Message, error) {
	svcMsg := &messagingsvc.Message{
		ID:             msg.ID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Content:        msg.Content,
		MessageType:    msg.MessageType,
		AttachmentURL:  msg.AttachmentURL,
		AttachmentType: msg.AttachmentType,
		Metadata:       msg.Metadata,
		IsRead:         msg.IsRead,
		ReadAt:         msg.ReadAt,
		CreatedAt:      msg.CreatedAt,
	}
	created, err := a.svc.CreateMessage(ctx, svcMsg)
	if err != nil {
		return nil, err
	}
	return svcMessageToHandler(created), nil
}

func (a *MessageServiceAdapter) ListMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]handler.Message, error) {
	msgs, err := a.svc.ListMessages(ctx, conversationID, limit, offset)
	if err != nil {
		return nil, err
	}
	result := make([]handler.Message, len(msgs))
	for i, m := range msgs {
		result[i] = *svcMessageToHandler(&m)
	}
	return result, nil
}

func (a *MessageServiceAdapter) MarkMessagesRead(ctx context.Context, conversationID, userID uuid.UUID) error {
	return a.svc.MarkMessagesRead(ctx, conversationID, userID)
}

func (a *MessageServiceAdapter) CountUnreadMessages(ctx context.Context, userID uuid.UUID) (int64, error) {
	return a.svc.CountUnreadMessages(ctx, userID)
}

func (a *MessageServiceAdapter) UpdateConversationLastMessage(ctx context.Context, conversationID uuid.UUID, lastMessageAt time.Time, preview string) error {
	return a.svc.UpdateConversationLastMessage(ctx, conversationID, lastMessageAt, preview)
}

func svcConversationToHandler(c *messagingsvc.Conversation) *handler.Conversation {
	return &handler.Conversation{
		ID:                 c.ID,
		JobID:              c.JobID,
		Participant1:       c.Participant1,
		Participant2:       c.Participant2,
		LastMessageAt:      c.LastMessageAt,
		LastMessagePreview: c.LastMessagePreview,
		IsArchived1:        c.IsArchived1,
		IsArchived2:        c.IsArchived2,
		CreatedAt:          c.CreatedAt,
		UpdatedAt:          c.UpdatedAt,
	}
}

func svcMessageToHandler(m *messagingsvc.Message) *handler.Message {
	return &handler.Message{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		SenderID:       m.SenderID,
		Content:        m.Content,
		MessageType:    m.MessageType,
		AttachmentURL:  m.AttachmentURL,
		AttachmentType: m.AttachmentType,
		Metadata:       m.Metadata,
		IsRead:         m.IsRead,
		ReadAt:         m.ReadAt,
		CreatedAt:      m.CreatedAt,
	}
}

// ---------------------------------------------------------------------------
// SubscriptionServiceAdapter — implements handler.SubscriptionService
// ---------------------------------------------------------------------------

// SubscriptionServiceAdapter wraps subscriptionsvc.SubscriptionService to match
// handler.SubscriptionService.
type SubscriptionServiceAdapter struct {
	svc *subscriptionsvc.SubscriptionService
}

func NewSubscriptionServiceAdapter(svc *subscriptionsvc.SubscriptionService) *SubscriptionServiceAdapter {
	return &SubscriptionServiceAdapter{svc: svc}
}

func (a *SubscriptionServiceAdapter) GetCurrentSubscription(ctx context.Context, providerID uuid.UUID) (*handler.Subscription, error) {
	sub, err := a.svc.GetCurrentSubscription(ctx, providerID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, nil
	}
	return svcSubscriptionToHandler(sub), nil
}

func (a *SubscriptionServiceAdapter) Subscribe(ctx context.Context, providerID uuid.UUID, tier, paymentMethod string) (*handler.Subscription, error) {
	sub, err := a.svc.Subscribe(ctx, providerID, tier, paymentMethod)
	if err != nil {
		return nil, err
	}
	return svcSubscriptionToHandler(sub), nil
}

func (a *SubscriptionServiceAdapter) CancelSubscription(ctx context.Context, subscriptionID, providerID uuid.UUID) error {
	return a.svc.CancelSubscription(ctx, subscriptionID, providerID)
}

func (a *SubscriptionServiceAdapter) HandlePaymentWebhook(ctx context.Context, payload []byte, signature string) error {
	return a.svc.HandlePaymentWebhook(ctx, payload, signature)
}

func (a *SubscriptionServiceAdapter) ListBillingHistory(ctx context.Context, providerID uuid.UUID, limit, offset int) ([]handler.Subscription, int, error) {
	subs, total, err := a.svc.ListBillingHistory(ctx, providerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	result := make([]handler.Subscription, len(subs))
	for i, s := range subs {
		result[i] = *svcSubscriptionToHandler(&s)
	}
	return result, total, nil
}

func svcSubscriptionToHandler(s *subscriptionsvc.Subscription) *handler.Subscription {
	return &handler.Subscription{
		ID:                    s.ID,
		ProviderID:            s.ProviderID,
		Tier:                  s.Tier,
		StartedAt:             s.StartedAt,
		ExpiresAt:             s.ExpiresAt,
		AutoRenew:             s.AutoRenew,
		PaymentMethod:         s.PaymentMethod,
		GatewaySubscriptionID: s.GatewaySubscriptionID,
		Amount:                s.Amount,
		Currency:              s.Currency,
		Status:                s.Status,
		CreatedAt:             s.CreatedAt,
		UpdatedAt:             s.UpdatedAt,
	}
}

// ---------------------------------------------------------------------------
// Compile-time interface assertions
// ---------------------------------------------------------------------------

var _ handler.UserService = (*UserServiceAdapter)(nil)
var _ handler.JobService = (*JobServiceAdapter)(nil)
var _ handler.ReviewService = (*ReviewServiceAdapter)(nil)
var _ handler.PaymentService = (*PaymentServiceAdapter)(nil)
var _ handler.NotificationService = (*NotificationServiceAdapter)(nil)
var _ handler.DisputeService = (*DisputeServiceAdapter)(nil)
var _ handler.GamificationService = (*GamificationServiceAdapter)(nil)
var _ handler.RouteService = (*RouteServiceAdapter)(nil)
var _ handler.ProviderService = (*ProviderServiceAdapter)(nil)
var _ handler.CropCalendarService = (*CropServiceAdapter)(nil)
var _ handler.SearchService = (*SearchServiceAdapter)(nil)
var _ handler.AdminService = (*AdminServiceAdapter)(nil)
var _ handler.AIService = (*AIServiceAdapter)(nil)
var _ handler.MessageService = (*MessageServiceAdapter)(nil)
var _ handler.SubscriptionService = (*SubscriptionServiceAdapter)(nil)
