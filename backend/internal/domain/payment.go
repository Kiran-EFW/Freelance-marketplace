package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PaymentStatus represents the lifecycle state of a payment transaction.
type PaymentStatus string

const (
	PaymentStatusPending            PaymentStatus = "pending"
	PaymentStatusProcessing         PaymentStatus = "processing"
	PaymentStatusCompleted          PaymentStatus = "completed"
	PaymentStatusFailed             PaymentStatus = "failed"
	PaymentStatusRefunded           PaymentStatus = "refunded"
	PaymentStatusPartiallyRefunded  PaymentStatus = "partially_refunded"
)

// EscrowStatus represents the state of funds held in escrow.
type EscrowStatus string

const (
	EscrowHeld     EscrowStatus = "held"
	EscrowReleased EscrowStatus = "released"
	EscrowRefunded EscrowStatus = "refunded"
	EscrowDisputed EscrowStatus = "disputed"
)

// TransactionPaymentMethod represents how the customer is paying.
type TransactionPaymentMethod string

const (
	TxPaymentUPI          TransactionPaymentMethod = "upi"
	TxPaymentCard         TransactionPaymentMethod = "card"
	TxPaymentBankTransfer TransactionPaymentMethod = "bank_transfer"
	TxPaymentWallet       TransactionPaymentMethod = "wallet"
	TxPaymentCash         TransactionPaymentMethod = "cash"
)

// Transaction represents a financial transaction tied to a job.
type Transaction struct {
	ID                  uuid.UUID                `json:"id" db:"id"`
	JobID               uuid.UUID                `json:"job_id" db:"job_id"`
	CustomerID          uuid.UUID                `json:"customer_id" db:"customer_id"`
	ProviderID          uuid.UUID                `json:"provider_id" db:"provider_id"`
	Amount              float64                  `json:"amount" db:"amount"`
	Currency            string                   `json:"currency" db:"currency"`
	CommissionRate      float64                  `json:"commission_rate" db:"commission_rate"`
	CommissionAmount    float64                  `json:"commission_amount" db:"commission_amount"`
	TaxAmount           float64                  `json:"tax_amount" db:"tax_amount"`
	ProviderPayoutAmount float64                 `json:"provider_payout_amount" db:"provider_payout_amount"`
	PaymentMethod       TransactionPaymentMethod `json:"payment_method" db:"payment_method"`
	PaymentStatus       PaymentStatus            `json:"payment_status" db:"payment_status"`
	EscrowStatus        *EscrowStatus            `json:"escrow_status,omitempty" db:"escrow_status"`
	PaymentGateway      string                   `json:"payment_gateway,omitempty" db:"payment_gateway"`
	GatewayOrderID      string                   `json:"gateway_order_id,omitempty" db:"gateway_order_id"`
	GatewayPaymentID    string                   `json:"gateway_payment_id,omitempty" db:"gateway_payment_id"`
	GatewaySignature    string                   `json:"gateway_signature,omitempty" db:"gateway_signature"`
	PaidAt              *time.Time               `json:"paid_at,omitempty" db:"paid_at"`
	SettledAt           *time.Time               `json:"settled_at,omitempty" db:"settled_at"`
	RefundAmount        *float64                 `json:"refund_amount,omitempty" db:"refund_amount"`
	RefundedAt          *time.Time               `json:"refunded_at,omitempty" db:"refunded_at"`
	CreatedAt           time.Time                `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time                `json:"updated_at" db:"updated_at"`
}

// CreatePaymentParams holds the inputs needed to create a new transaction.
type CreatePaymentParams struct {
	JobID          uuid.UUID
	CustomerID     uuid.UUID
	ProviderID     uuid.UUID
	Amount         float64
	Currency       string
	PaymentMethod  TransactionPaymentMethod
	PaymentGateway string
}

// CommissionBreakdown describes the platform's take on a transaction.
type CommissionBreakdown struct {
	GrossAmount       float64 `json:"gross_amount"`
	CommissionRate    float64 `json:"commission_rate"`
	CommissionAmount  float64 `json:"commission_amount"`
	TaxAmount         float64 `json:"tax_amount"`
	ProviderPayout    float64 `json:"provider_payout"`
}

// TaxBreakdown describes taxes applied to a transaction.
type TaxBreakdown struct {
	TaxableAmount float64 `json:"taxable_amount"`
	TaxRate       float64 `json:"tax_rate"`
	TaxAmount     float64 `json:"tax_amount"`
	TaxType       string  `json:"tax_type"` // e.g. "GST", "VAT"
}

// TransactionRepository defines persistence operations for transactions.
type TransactionRepository interface {
	Create(ctx context.Context, tx *Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	GetByJobID(ctx context.Context, jobID uuid.UUID) (*Transaction, error)
	GetByGatewayOrderID(ctx context.Context, orderID string) (*Transaction, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status PaymentStatus) error
	UpdateEscrowStatus(ctx context.Context, id uuid.UUID, status EscrowStatus) error
	UpdateGatewayPaymentID(ctx context.Context, id uuid.UUID, paymentID, signature string) error
	SetSettled(ctx context.Context, id uuid.UUID, settledAt time.Time) error
	SetRefunded(ctx context.Context, id uuid.UUID, amount float64, refundedAt time.Time) error
	ListByCustomer(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]Transaction, error)
	ListByProvider(ctx context.Context, providerID uuid.UUID, limit, offset int) ([]Transaction, error)
}
