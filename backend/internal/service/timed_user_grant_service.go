package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrTimedGrantInvalid = infraerrors.BadRequest("TIMED_GRANT_INVALID", "invalid timed grant")
)

type timedGrantSQLDB interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type TimedUserGrantService struct {
	db                   timedGrantSQLDB
	authCacheInvalidator APIKeyAuthCacheInvalidator
}

type dueGrant struct {
	id     int64
	userID int64
	kind   string
	amount float64
}

func NewTimedUserGrantService(sqlDB *sql.DB, authCacheInvalidator APIKeyAuthCacheInvalidator) *TimedUserGrantService {
	return &TimedUserGrantService{db: sqlDB, authCacheInvalidator: authCacheInvalidator}
}

func (s *TimedUserGrantService) Create(ctx context.Context, userID int64, input CreateTimedUserGrantInput) (*TimedUserGrant, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("timed grant service unavailable")
	}
	if userID <= 0 || !validTimedGrantType(input.GrantType) || input.Amount <= 0 || input.DurationSeconds <= 0 {
		return nil, ErrTimedGrantInvalid
	}
	if input.GrantType == TimedGrantTypeConcurrency && math.Trunc(input.Amount) != input.Amount {
		return nil, ErrTimedGrantInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer rollbackUnlessDone(tx)

	grant, err := scanTimedGrant(tx.QueryRowContext(ctx, `
		INSERT INTO timed_user_grants (user_id, grant_type, amount, duration_seconds, status, created_by, notes)
		VALUES ($1, $2, $3, $4, 'pending', $5, $6)
		RETURNING id, user_id, grant_type, amount, duration_seconds, status, activated_at, expires_at, expired_at, deducted_amount, created_by, notes, created_at, updated_at
	`, userID, input.GrantType, input.Amount, input.DurationSeconds, input.CreatedBy, input.Notes))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return grant, nil
}

func (s *TimedUserGrantService) ListByUser(ctx context.Context, userID int64) ([]TimedUserGrant, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("timed grant service unavailable")
	}
	rows, err := queryer(s.db).QueryContext(ctx, `
		SELECT id, user_id, grant_type, amount, duration_seconds, status, activated_at, expires_at, expired_at, deducted_amount, created_by, notes, created_at, updated_at
		FROM timed_user_grants
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var grants []TimedUserGrant
	for rows.Next() {
		grant, err := scanTimedGrantRows(rows)
		if err != nil {
			return nil, err
		}
		grants = append(grants, *grant)
	}
	return grants, rows.Err()
}

func (s *TimedUserGrantService) ActivatePendingUserGrants(ctx context.Context, userID int64) (int, error) {
	if s == nil || s.db == nil || userID <= 0 {
		return 0, nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer rollbackUnlessDone(tx)

	rows, err := tx.QueryContext(ctx, `
		SELECT id, grant_type, amount, duration_seconds
		FROM timed_user_grants
		WHERE user_id = $1 AND status = 'pending'
		ORDER BY id
		FOR UPDATE
	`, userID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type pendingGrant struct {
		id       int64
		kind     string
		amount   float64
		duration int
	}
	var grants []pendingGrant
	for rows.Next() {
		var g pendingGrant
		if err := rows.Scan(&g.id, &g.kind, &g.amount, &g.duration); err != nil {
			return 0, err
		}
		grants = append(grants, g)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if len(grants) == 0 {
		return 0, tx.Commit()
	}

	var balanceAdd float64
	var concurrencyAdd int
	now := time.Now().UTC()
	for _, g := range grants {
		expiresAt := now.Add(time.Duration(g.duration) * time.Second)
		if _, err := tx.ExecContext(ctx, `
			UPDATE timed_user_grants
			SET status = 'active', activated_at = $2, expires_at = $3, updated_at = NOW()
			WHERE id = $1
		`, g.id, now, expiresAt); err != nil {
			return 0, err
		}
		typeName := "限时余额"
		historyType := AdjustmentTypeAdminBalance
		if g.kind == TimedGrantTypeConcurrency {
			typeName = "限时并发"
			historyType = AdjustmentTypeAdminConcurrency
			concurrencyAdd += int(g.amount)
		} else {
			balanceAdd += g.amount
		}
		if err := insertTimedGrantHistory(ctx, tx, userID, historyType, g.amount, fmt.Sprintf("%s首次使用激活，数量 %.8g，有效期至 %s", typeName, g.amount, expiresAt.Format("2006-01-02 15:04:05")), g.id); err != nil {
			return 0, err
		}
	}
	if balanceAdd > 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE users SET balance = balance + $1, updated_at = NOW() WHERE id = $2`, balanceAdd, userID); err != nil {
			return 0, err
		}
	}
	if concurrencyAdd > 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE users SET concurrency = concurrency + $1, updated_at = NOW() WHERE id = $2`, concurrencyAdd, userID); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	if s.authCacheInvalidator != nil {
		s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, userID)
	}
	return len(grants), nil
}

func (s *TimedUserGrantService) ExpireDueGrants(ctx context.Context, now time.Time, limit int) (int, error) {
	if s == nil || s.db == nil {
		return 0, nil
	}
	if limit <= 0 {
		limit = 100
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer rollbackUnlessDone(tx)

	rows, err := tx.QueryContext(ctx, `
		SELECT id, user_id, grant_type, amount
		FROM timed_user_grants
		WHERE status = 'active' AND expires_at <= $1
		ORDER BY expires_at, id
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`, now, limit)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var grants []dueGrant
	for rows.Next() {
		var g dueGrant
		if err := rows.Scan(&g.id, &g.userID, &g.kind, &g.amount); err != nil {
			return 0, err
		}
		grants = append(grants, g)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	for _, g := range grants {
		deducted, err := expireOneGrant(ctx, tx, g, now)
		if err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE timed_user_grants SET status = 'expired', expired_at = $2, deducted_amount = $3, updated_at = NOW() WHERE id = $1
		`, g.id, now, deducted); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	for _, g := range grants {
		if s.authCacheInvalidator != nil {
			s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, g.userID)
		}
	}
	return len(grants), nil
}

func (s *TimedUserGrantService) StartExpiryLoop(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_, _ = s.ExpireDueGrants(ctx, time.Now().UTC(), 100)
			}
		}
	}()
}

func expireOneGrant(ctx context.Context, tx *sql.Tx, g dueGrant, now time.Time) (float64, error) {
	if g.kind == TimedGrantTypeConcurrency {
		var current int
		if err := tx.QueryRowContext(ctx, `SELECT concurrency FROM users WHERE id = $1 FOR UPDATE`, g.userID).Scan(&current); err != nil {
			return 0, err
		}
		deduct := int(g.amount)
		if deduct > current {
			deduct = current
		}
		if deduct < 0 {
			deduct = 0
		}
		if _, err := tx.ExecContext(ctx, `UPDATE users SET concurrency = concurrency - $1, updated_at = NOW() WHERE id = $2`, deduct, g.userID); err != nil {
			return 0, err
		}
		return float64(deduct), insertTimedGrantHistory(ctx, tx, g.userID, AdjustmentTypeAdminConcurrency, -float64(deduct), fmt.Sprintf("限时并发到期自动扣减，原发放 %.8g，实际扣减 %d", g.amount, deduct), g.id)
	}

	var current float64
	if err := tx.QueryRowContext(ctx, `SELECT balance FROM users WHERE id = $1 FOR UPDATE`, g.userID).Scan(&current); err != nil {
		return 0, err
	}
	deduct := g.amount
	if deduct > current {
		deduct = current
	}
	if deduct < 0 {
		deduct = 0
	}
	if _, err := tx.ExecContext(ctx, `UPDATE users SET balance = GREATEST(balance - $1, 0), updated_at = NOW() WHERE id = $2`, deduct, g.userID); err != nil {
		return 0, err
	}
	return deduct, insertTimedGrantHistory(ctx, tx, g.userID, AdjustmentTypeAdminBalance, -deduct, fmt.Sprintf("限时余额到期自动扣减，原发放 %.8g，实际扣减 %.8g", g.amount, deduct), g.id)
}

func insertTimedGrantHistory(ctx context.Context, tx *sql.Tx, userID int64, historyType string, value float64, notes string, grantID int64) error {
	code := fmt.Sprintf("TG%012d%08x", grantID, time.Now().UnixNano()&0xffffffff)
	_, err := tx.ExecContext(ctx, `
		INSERT INTO redeem_codes (code, type, value, status, used_by, used_at, notes, created_at)
		VALUES ($1, $2, $3, 'used', $4, NOW(), $5, NOW())
	`, code, historyType, value, userID, notes)
	if err == nil {
		return nil
	}
	// Very old installs may not have redeem_codes.notes yet during migration/bootstrap.
	if strings.Contains(err.Error(), "notes") {
		_, fallbackErr := tx.ExecContext(ctx, `
			INSERT INTO redeem_codes (code, type, value, status, used_by, used_at, created_at)
			VALUES ($1, $2, $3, 'used', $4, NOW(), NOW())
		`, code, historyType, value, userID)
		return fallbackErr
	}
	return err
}

type timedGrantScanner interface {
	Scan(dest ...any) error
}

type timedGrantRows interface {
	Scan(dest ...any) error
}

type queryContext interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func queryer(db timedGrantSQLDB) queryContext { return db.(*sql.DB) }

func scanTimedGrant(row timedGrantScanner) (*TimedUserGrant, error) {
	grant := &TimedUserGrant{}
	err := row.Scan(&grant.ID, &grant.UserID, &grant.GrantType, &grant.Amount, &grant.DurationSeconds, &grant.Status, &grant.ActivatedAt, &grant.ExpiresAt, &grant.ExpiredAt, &grant.DeductedAmount, &grant.CreatedBy, &grant.Notes, &grant.CreatedAt, &grant.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return grant, nil
}

func scanTimedGrantRows(row timedGrantRows) (*TimedUserGrant, error) {
	return scanTimedGrant(row)
}

func validTimedGrantType(t string) bool {
	return t == TimedGrantTypeBalance || t == TimedGrantTypeConcurrency
}

func rollbackUnlessDone(tx *sql.Tx) {
	_ = tx.Rollback()
}
