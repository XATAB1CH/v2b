package memory

import (
	"context"
	"sync"
	"time"

	"github.com/XATAB1CH/v2b/internal/domain/billing"
)

type PaymentsStore struct {
	mu sync.RWMutex
	m  map[string]billing.Payment
}

func NewPaymentsStore() *PaymentsStore {
	return &PaymentsStore{m: make(map[string]billing.Payment)}
}

func (s *PaymentsStore) Create(ctx context.Context, p billing.Payment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[p.ID] = p
	return nil
}

func (s *PaymentsStore) Get(ctx context.Context, paymentID string) (billing.Payment, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.m[paymentID]
	return p, ok, nil
}

func (s *PaymentsStore) MarkSucceeded(ctx context.Context, paymentID string, paidAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.m[paymentID]
	if !ok {
		return nil
	}

	p.Status = billing.PaymentStatusSucceeded
	p.SucceededAt = &paidAt
	s.m[paymentID] = p
	return nil
}
