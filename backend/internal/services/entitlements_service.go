package services

import (
	"context"
	"time"

	"github.com/XATAB1CH/v2b/internal/store/postgres"
)

type EntitlementsService struct {
	repo *postgres.EntitlementsRepo
}

func NewEntitlementsService(repo *postgres.EntitlementsRepo) *EntitlementsService {
	return &EntitlementsService{repo: repo}
}

type Me struct {
	PaidUntil          *time.Time `json:"paid_until"`
	SubscriptionActive bool       `json:"subscription_active"`
	FreeAttemptsLeft   int        `json:"free_attempts_left"`
}

func (s *EntitlementsService) GetMe(ctx context.Context, userID string) (Me, error) {
	e, err := s.repo.Get(ctx, userID)
	if err != nil {
		return Me{}, err
	}

	active := e.PaidUntil != nil && e.PaidUntil.After(time.Now().UTC())
	return Me{
		PaidUntil:          e.PaidUntil,
		SubscriptionActive: active,
		FreeAttemptsLeft:   e.FreeAttemptsLeft,
	}, nil
}

// Возвращает allowed=false и reason="free_limit_reached", если лимит исчерпан.
func (s *EntitlementsService) ConsumeAttemptOrAllow(ctx context.Context, userID string) (attemptsLeft int, allowed bool, reason string, err error) {
	active, _, err := s.repo.HasPaidAccess(ctx, userID)
	if err != nil {
		return 0, false, "", err
	}
	if active {
		return 0, true, "", nil
	}

	left, ok, err := s.repo.ConsumeFreeAttempt(ctx, userID)
	if err != nil {
		return 0, false, "", err
	}
	if !ok {
		return 0, false, "free_limit_reached", nil
	}

	return left, true, "", nil
}
