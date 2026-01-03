package billing

import (
	"context"
	"time"
)

type PaymentProvider interface {
	CreatePayment(ctx context.Context, userID string, amount string, currency string, returnURL string, description string) (providerPaymentID string, confirmationURL string, err error)
}

type PaymentsStore interface {
	Create(ctx context.Context, p Payment) error
	MarkSucceeded(ctx context.Context, paymentID string, paidAt time.Time) error
	Get(ctx context.Context, paymentID string) (Payment, bool, error)
}

type EntitlementsStore interface {
	ExtendPaidUntil(ctx context.Context, userID string, days int) error
}

type Payment struct {
	ID                string
	UserID            string
	Provider          string // "stub" сейчас, позже "yookassa"
	ProviderPaymentID string
	Status            string // created|pending|succeeded|canceled
	Amount            string
	Currency          string
	CreatedAt         time.Time
	SucceededAt       *time.Time
}

const (
	PaymentStatusCreated   = "created"
	PaymentStatusPending   = "pending"
	PaymentStatusSucceeded = "succeeded"
	PaymentStatusCanceled  = "canceled"
)

type CreatePaymentResult struct {
	PaymentID       string `json:"payment_id"`
	ConfirmationURL string `json:"confirmation_url"`
}
