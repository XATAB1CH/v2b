package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Entitlements struct {
	UserID           string
	FreeAttemptsLeft int
	PaidUntil        *time.Time
	Email            string
}

type EntitlementsRepo struct {
	db *pgxpool.Pool
}

func NewEntitlementsRepo(db *pgxpool.Pool) *EntitlementsRepo {
	return &EntitlementsRepo{db: db}
}

func (r *EntitlementsRepo) Ensure(ctx context.Context, userID string, freeLimit int) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO entitlements(user_id, free_attempts_left)
		VALUES($1, $2)
		ON CONFLICT(user_id) DO NOTHING
	`, userID, freeLimit)
	return err
}

func (r *EntitlementsRepo) Get(ctx context.Context, userID string) (Entitlements, error) {
	var e Entitlements
	var paidUntil *time.Time

	err := r.db.QueryRow(ctx, `
		SELECT
			u.id AS user_id,
			e.free_attempts_left,
			e.paid_until,
			u.email
		FROM public.users u
		LEFT JOIN public.entitlements e
		ON e.user_id = u.id
		WHERE u.id = $1
	`, userID).Scan(&e.UserID, &e.FreeAttemptsLeft, &paidUntil, &e.Email)

	if err != nil {
		return Entitlements{}, err
	}

	e.PaidUntil = paidUntil
	return e, nil
}

func (r *EntitlementsRepo) HasPaidAccess(ctx context.Context, userID string) (bool, *time.Time, error) {
	var paidUntil *time.Time
	if err := r.db.QueryRow(ctx, `SELECT paid_until FROM entitlements WHERE user_id=$1`, userID).Scan(&paidUntil); err != nil {
		return false, nil, err
	}

	now := time.Now().UTC()
	if paidUntil != nil && paidUntil.After(now) {
		return true, paidUntil, nil
	}
	return false, paidUntil, nil
}

// Списываем попытку только если нет активного доступа и попытки > 0.
// Возвращаем attemptsLeft после списания.
func (r *EntitlementsRepo) ConsumeFreeAttempt(ctx context.Context, userID string) (attemptsLeft int, ok bool, err error) {
	row := r.db.QueryRow(ctx, `
		UPDATE entitlements
		SET free_attempts_left = free_attempts_left - 1,
		    updated_at = now()
		WHERE user_id=$1
		  AND (paid_until IS NULL OR paid_until <= now())
		  AND free_attempts_left > 0
		RETURNING free_attempts_left
	`, userID)

	if scanErr := row.Scan(&attemptsLeft); scanErr != nil {
		// Не списалось: либо попыток нет, либо доступ активен
		return 0, false, nil
	}

	return attemptsLeft, true, nil
}

// Начислить paid_until на N дней с "stacking":
// если paid_until в будущем — добавляем сверху, иначе ставим now+days.
func (r *EntitlementsRepo) ExtendPaidDays(ctx context.Context, userID string, days int) (*time.Time, error) {
	var paidUntil time.Time
	err := r.db.QueryRow(ctx, `
		UPDATE entitlements
		SET paid_until = CASE
		  WHEN paid_until IS NOT NULL AND paid_until > now()
		    THEN paid_until + ($2 * interval '1 day')
		  ELSE now() + ($2 * interval '1 day')
		END,
		updated_at = now()
		WHERE user_id=$1
		RETURNING paid_until
	`, userID, days).Scan(&paidUntil)

	if err != nil {
		return nil, err
	}
	return &paidUntil, nil
}
