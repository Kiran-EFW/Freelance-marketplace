package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/domain"
	gateway "github.com/seva-platform/backend/internal/adapter/payment"
)

// Loyalty-based commission rates. The platform takes a decreasing cut as
// providers bring more bookings through a single customer relationship.
const (
	CommissionFirst    = 0.08 // 1st booking: 8%
	CommissionSecond   = 0.06 // 2nd booking: 6%
	CommissionThird    = 0.04 // 3rd booking: 4%
	CommissionFourthOn = 0.03 // 4th+ booking: 3%
	CommissionRecurring = 0.02 // recurring/route: 2%
)

// Service defines the payment service interface.
type Service interface {
	CreateOrder(ctx context.Context, jobID uuid.UUID, amount float64, currency string) (*domain.Transaction, error)
	ProcessPayment(ctx context.Context, txID uuid.UUID, gatewayPaymentID string) error
	HoldEscrow(ctx context.Context, txID uuid.UUID) error
	ReleaseEscrow(ctx context.Context, txID uuid.UUID) error
	Refund(ctx context.Context, txID uuid.UUID, amount float64, reason string) error
	CalculateCommission(ctx context.Context, jobAmount float64, providerLevel int, bookingCount int) domain.CommissionBreakdown
	CalculateTax(ctx context.Context, amount float64, jurisdictionID string) domain.TaxBreakdown
}

// PaymentService implements payment processing, escrow, and commission logic.
type PaymentService struct {
	transactions domain.TransactionRepository
	jobs         domain.JobRepository
	gateway      gateway.PaymentGateway
}

// NewPaymentService returns a ready-to-use PaymentService.
func NewPaymentService(
	transactions domain.TransactionRepository,
	jobs domain.JobRepository,
	gw gateway.PaymentGateway,
) *PaymentService {
	return &PaymentService{
		transactions: transactions,
		jobs:         jobs,
		gateway:      gw,
	}
}

// CreateOrder creates a gateway order and a local transaction record.
func (s *PaymentService) CreateOrder(ctx context.Context, jobID uuid.UUID, amount float64, currency string) (*domain.Transaction, error) {
	job, err := s.jobs.GetByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("%w: job %s", domain.ErrNotFound, jobID)
	}

	if job.ProviderID == nil {
		return nil, fmt.Errorf("%w: job has no assigned provider", domain.ErrInvalidState)
	}

	if currency == "" {
		currency = job.Currency
	}

	// Create gateway order.
	metadata := map[string]string{
		"job_id":      jobID.String(),
		"customer_id": job.CustomerID.String(),
	}
	order, err := s.gateway.CreateOrder(ctx, amount, currency, metadata)
	if err != nil {
		log.Error().Err(err).Str("job_id", jobID.String()).Msg("gateway order creation failed")
		return nil, fmt.Errorf("create gateway order: %w", err)
	}

	// Calculate commission (default to first booking rate).
	commission := s.CalculateCommission(ctx, amount, 0, 1)
	tax := s.CalculateTax(ctx, commission.CommissionAmount, job.JurisdictionID)

	tx := &domain.Transaction{
		ID:                   uuid.New(),
		JobID:                jobID,
		CustomerID:           job.CustomerID,
		ProviderID:           *job.ProviderID,
		Amount:               amount,
		Currency:             currency,
		CommissionRate:       commission.CommissionRate,
		CommissionAmount:     commission.CommissionAmount,
		TaxAmount:            tax.TaxAmount,
		ProviderPayoutAmount: commission.ProviderPayout,
		PaymentStatus:        domain.PaymentStatusPending,
		PaymentGateway:       order.Gateway,
		GatewayOrderID:       order.ID,
	}

	if err := s.transactions.Create(ctx, tx); err != nil {
		log.Error().Err(err).Str("job_id", jobID.String()).Msg("failed to create transaction")
		return nil, fmt.Errorf("create transaction: %w", err)
	}

	log.Info().
		Str("tx_id", tx.ID.String()).
		Str("job_id", jobID.String()).
		Float64("amount", amount).
		Msg("payment order created")

	return tx, nil
}

// ProcessPayment verifies the payment with the gateway and updates the
// transaction status.
func (s *PaymentService) ProcessPayment(ctx context.Context, txID uuid.UUID, gatewayPaymentID string) error {
	tx, err := s.transactions.GetByID(ctx, txID)
	if err != nil {
		return fmt.Errorf("%w: transaction %s", domain.ErrNotFound, txID)
	}

	if tx.PaymentStatus != domain.PaymentStatusPending {
		return fmt.Errorf("%w: transaction is already %s", domain.ErrInvalidState, tx.PaymentStatus)
	}

	result, err := s.gateway.VerifyPayment(ctx, tx.GatewayOrderID, gatewayPaymentID, "")
	if err != nil {
		log.Error().Err(err).Str("tx_id", txID.String()).Msg("payment verification failed")
		if updateErr := s.transactions.UpdateStatus(ctx, txID, domain.PaymentStatusFailed); updateErr != nil {
			log.Error().Err(updateErr).Msg("failed to mark transaction as failed")
		}
		return fmt.Errorf("verify payment: %w", err)
	}

	if !result.Verified {
		if err := s.transactions.UpdateStatus(ctx, txID, domain.PaymentStatusFailed); err != nil {
			log.Error().Err(err).Msg("failed to mark transaction as failed")
		}
		return fmt.Errorf("%w: payment verification failed", domain.ErrInvalidInput)
	}

	if err := s.transactions.UpdateGatewayPaymentID(ctx, txID, gatewayPaymentID, ""); err != nil {
		return fmt.Errorf("update gateway payment ID: %w", err)
	}

	if err := s.transactions.UpdateStatus(ctx, txID, domain.PaymentStatusCompleted); err != nil {
		return fmt.Errorf("update transaction status: %w", err)
	}

	log.Info().Str("tx_id", txID.String()).Msg("payment processed")
	return nil
}

// HoldEscrow places funds in escrow until the job is completed.
func (s *PaymentService) HoldEscrow(ctx context.Context, txID uuid.UUID) error {
	tx, err := s.transactions.GetByID(ctx, txID)
	if err != nil {
		return fmt.Errorf("%w: transaction %s", domain.ErrNotFound, txID)
	}

	if tx.PaymentStatus != domain.PaymentStatusCompleted {
		return fmt.Errorf("%w: can only hold escrow on completed payments", domain.ErrInvalidState)
	}

	if err := s.transactions.UpdateEscrowStatus(ctx, txID, domain.EscrowHeld); err != nil {
		return fmt.Errorf("hold escrow: %w", err)
	}

	log.Info().Str("tx_id", txID.String()).Msg("funds held in escrow")
	return nil
}

// ReleaseEscrow releases held funds to the provider minus commission.
func (s *PaymentService) ReleaseEscrow(ctx context.Context, txID uuid.UUID) error {
	tx, err := s.transactions.GetByID(ctx, txID)
	if err != nil {
		return fmt.Errorf("%w: transaction %s", domain.ErrNotFound, txID)
	}

	if tx.EscrowStatus == nil || *tx.EscrowStatus != domain.EscrowHeld {
		return fmt.Errorf("%w: escrow is not in held state", domain.ErrInvalidState)
	}

	if err := s.transactions.UpdateEscrowStatus(ctx, txID, domain.EscrowReleased); err != nil {
		return fmt.Errorf("release escrow: %w", err)
	}

	now := time.Now()
	if err := s.transactions.SetSettled(ctx, txID, now); err != nil {
		return fmt.Errorf("set settled: %w", err)
	}

	log.Info().
		Str("tx_id", txID.String()).
		Float64("payout", tx.ProviderPayoutAmount).
		Msg("escrow released to provider")

	return nil
}

// Refund processes a full or partial refund.
func (s *PaymentService) Refund(ctx context.Context, txID uuid.UUID, amount float64, reason string) error {
	tx, err := s.transactions.GetByID(ctx, txID)
	if err != nil {
		return fmt.Errorf("%w: transaction %s", domain.ErrNotFound, txID)
	}

	if amount <= 0 || amount > tx.Amount {
		return fmt.Errorf("%w: refund amount must be between 0 and %.2f", domain.ErrInvalidInput, tx.Amount)
	}

	_, err = s.gateway.Refund(ctx, tx.GatewayPaymentID, amount)
	if err != nil {
		log.Error().Err(err).Str("tx_id", txID.String()).Msg("gateway refund failed")
		return fmt.Errorf("gateway refund: %w", err)
	}

	now := time.Now()
	if err := s.transactions.SetRefunded(ctx, txID, amount, now); err != nil {
		return fmt.Errorf("set refunded: %w", err)
	}

	status := domain.PaymentStatusRefunded
	if amount < tx.Amount {
		status = domain.PaymentStatusPartiallyRefunded
	}
	if err := s.transactions.UpdateStatus(ctx, txID, status); err != nil {
		return fmt.Errorf("update refund status: %w", err)
	}

	if tx.EscrowStatus != nil {
		if err := s.transactions.UpdateEscrowStatus(ctx, txID, domain.EscrowRefunded); err != nil {
			log.Warn().Err(err).Msg("failed to update escrow status after refund")
		}
	}

	log.Info().
		Str("tx_id", txID.String()).
		Float64("amount", amount).
		Str("reason", reason).
		Msg("refund processed")

	return nil
}

// CalculateCommission implements the loyalty pricing model.
// bookingCount is the number of bookings this customer has made with this
// provider (including the current one).
func (s *PaymentService) CalculateCommission(_ context.Context, jobAmount float64, providerLevel int, bookingCount int) domain.CommissionBreakdown {
	var rate float64

	switch {
	case bookingCount <= 0:
		rate = CommissionFirst
	case bookingCount == 1:
		rate = CommissionFirst
	case bookingCount == 2:
		rate = CommissionSecond
	case bookingCount == 3:
		rate = CommissionThird
	default:
		rate = CommissionFourthOn
	}

	// Recurring jobs always get the lowest rate.
	if bookingCount > 10 {
		rate = CommissionRecurring
	}

	// Premium providers get 1% discount on commission.
	if providerLevel >= 4 {
		rate -= 0.01
		if rate < CommissionRecurring {
			rate = CommissionRecurring
		}
	}

	commissionAmount := jobAmount * rate
	payout := jobAmount - commissionAmount

	return domain.CommissionBreakdown{
		GrossAmount:      jobAmount,
		CommissionRate:   rate,
		CommissionAmount: commissionAmount,
		ProviderPayout:   payout,
	}
}

// CalculateTax computes the applicable tax for a given jurisdiction.
func (s *PaymentService) CalculateTax(_ context.Context, amount float64, jurisdictionID string) domain.TaxBreakdown {
	// Jurisdiction-specific tax rates.
	var rate float64
	var taxType string

	switch jurisdictionID {
	case "in":
		rate = 0.18 // GST 18%
		taxType = "GST"
	case "uk":
		rate = 0.20 // VAT 20%
		taxType = "VAT"
	case "us":
		rate = 0.0 // varies by state; handled separately
		taxType = "SALES_TAX"
	default:
		rate = 0.0
		taxType = "NONE"
	}

	return domain.TaxBreakdown{
		TaxableAmount: amount,
		TaxRate:       rate,
		TaxAmount:     amount * rate,
		TaxType:       taxType,
	}
}
