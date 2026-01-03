package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/XATAB1CH/v2b/internal/domain/billing"
)

type BillingService struct {
	provider     billing.PaymentProvider
	payments     billing.PaymentsStore
	entitlements billing.EntitlementsStore
	durationDays int
}

func NewBillingService(provider billing.PaymentProvider, payments billing.PaymentsStore, ent billing.EntitlementsStore, durationDays int) *BillingService {
	return &BillingService{
		provider:     provider,
		payments:     payments,
		entitlements: ent,
		durationDays: durationDays,
	}
}

func (s *BillingService) Create(ctx context.Context, userID string, amount string, currency string, returnURL string, description string) (billing.CreatePaymentResult, error) {
	if userID == "" {
		return billing.CreatePaymentResult{}, errors.New("empty userID")
	}
	if amount == "" {
		return billing.CreatePaymentResult{}, errors.New("empty amount")
	}
	if currency == "" {
		currency = "RUB"
	}

	paymentID := fmt.Sprintf("pay_%d", time.Now().UnixNano())

	providerPaymentID, confirmationURL, err := s.provider.CreatePayment(ctx, userID, amount, currency, returnURL, description)
	if err != nil {
		return billing.CreatePaymentResult{}, err
	}

	now := time.Now().UTC()
	p := billing.Payment{
		ID:                paymentID,
		UserID:            userID,
		Provider:          "stub",
		ProviderPaymentID: providerPaymentID,
		Status:            billing.PaymentStatusPending,
		Amount:            amount,
		Currency:          currency,
		CreatedAt:         now,
	}

	if err := s.payments.Create(ctx, p); err != nil {
		return billing.CreatePaymentResult{}, err
	}

	return billing.CreatePaymentResult{
		PaymentID:       paymentID,
		ConfirmationURL: confirmationURL,
	}, nil
}

func (s *BillingService) MarkSucceeded(ctx context.Context, paymentID string) error {
	if paymentID == "" {
		return errors.New("empty paymentID")
	}

	p, ok, err := s.payments.Get(ctx, paymentID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("payment not found")
	}
	if p.Status == billing.PaymentStatusSucceeded {
		return nil
	}

	now := time.Now().UTC()
	if err := s.payments.MarkSucceeded(ctx, paymentID, now); err != nil {
		return err
	}
	log.Printf("BILLING MarkSucceeded paymentID=%s paymentUserID=%s status=%s", paymentID, p.UserID, p.Status)

	if err := s.entitlements.ExtendPaidUntil(ctx, p.UserID, s.durationDays); err != nil {
		log.Printf("BILLING ExtendPaidUntil error: %v", err)
		return err
	}
	log.Printf("BILLING ExtendPaidUntil OK userID=%s days=%d", p.UserID, s.durationDays)

	return nil
}
