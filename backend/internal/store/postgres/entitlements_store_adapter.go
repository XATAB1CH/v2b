package postgres

import "context"

type EntitlementsStoreAdapter struct {
	repo *EntitlementsRepo
}

func NewEntitlementsStoreAdapter(repo *EntitlementsRepo) *EntitlementsStoreAdapter {
	return &EntitlementsStoreAdapter{repo: repo}
}

func (a *EntitlementsStoreAdapter) ExtendPaidUntil(ctx context.Context, userID string, days int) error {
	_, err := a.repo.ExtendPaidDays(ctx, userID, days)
	return err
}
