package service

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

type timedGrantInvalidatorSpy struct {
	userIDs []int64
}

func (s *timedGrantInvalidatorSpy) InvalidateAuthCacheByKey(_ context.Context, _ string) {}

func (s *timedGrantInvalidatorSpy) InvalidateAuthCacheByUserID(_ context.Context, userID int64) {
	s.userIDs = append(s.userIDs, userID)
}

func (s *timedGrantInvalidatorSpy) InvalidateAuthCacheByGroupID(_ context.Context, _ int64) {}

func newTimedGrantSQLMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func TestTimedUserGrantService_CreateRejectsFractionalConcurrency(t *testing.T) {
	db, mock := newTimedGrantSQLMock(t)
	svc := NewTimedUserGrantService(db, nil)

	_, err := svc.Create(context.Background(), 123, CreateTimedUserGrantInput{
		GrantType:       TimedGrantTypeConcurrency,
		Amount:          1.5,
		DurationSeconds: 3600,
	})

	require.ErrorIs(t, err, ErrTimedGrantInvalid)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTimedUserGrantService_ActivatePendingUserGrantsAddsBalanceAndConcurrency(t *testing.T) {
	db, mock := newTimedGrantSQLMock(t)
	invalidator := &timedGrantInvalidatorSpy{}
	svc := NewTimedUserGrantService(db, invalidator)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, grant_type, amount, duration_seconds
		FROM timed_user_grants
		WHERE user_id = $1 AND status = 'pending'
		ORDER BY id
		FOR UPDATE`)).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "grant_type", "amount", "duration_seconds"}).
			AddRow(int64(11), TimedGrantTypeBalance, 25.0, 3600).
			AddRow(int64(12), TimedGrantTypeConcurrency, 2.0, 7200))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE timed_user_grants
			SET status = 'active', activated_at = $2, expires_at = $3, updated_at = NOW()
			WHERE id = $1`)).
		WithArgs(int64(11), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO redeem_codes (code, type, value, status, used_by, used_at, notes, created_at)
		VALUES ($1, $2, $3, 'used', $4, NOW(), $5, NOW())`)).
		WithArgs(sqlmock.AnyArg(), AdjustmentTypeAdminBalance, 25.0, int64(7), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE timed_user_grants
			SET status = 'active', activated_at = $2, expires_at = $3, updated_at = NOW()
			WHERE id = $1`)).
		WithArgs(int64(12), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO redeem_codes (code, type, value, status, used_by, used_at, notes, created_at)
		VALUES ($1, $2, $3, 'used', $4, NOW(), $5, NOW())`)).
		WithArgs(sqlmock.AnyArg(), AdjustmentTypeAdminConcurrency, 2.0, int64(7), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET balance = balance + $1, updated_at = NOW() WHERE id = $2`)).
		WithArgs(25.0, int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET concurrency = concurrency + $1, updated_at = NOW() WHERE id = $2`)).
		WithArgs(2, int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	activated, err := svc.ActivatePendingUserGrants(context.Background(), 7)

	require.NoError(t, err)
	require.Equal(t, 2, activated)
	require.Equal(t, []int64{7}, invalidator.userIDs)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTimedUserGrantService_ExpireDueGrantsDeductsAtMostCurrentBalanceAndConcurrency(t *testing.T) {
	db, mock := newTimedGrantSQLMock(t)
	invalidator := &timedGrantInvalidatorSpy{}
	svc := NewTimedUserGrantService(db, invalidator)
	now := time.Date(2026, 4, 29, 8, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, grant_type, amount
		FROM timed_user_grants
		WHERE status = 'active' AND expires_at <= $1
		ORDER BY expires_at, id
		LIMIT $2
		FOR UPDATE SKIP LOCKED`)).
		WithArgs(now, 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "grant_type", "amount"}).
			AddRow(int64(21), int64(7), TimedGrantTypeBalance, 50.0).
			AddRow(int64(22), int64(7), TimedGrantTypeConcurrency, 5.0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT balance FROM users WHERE id = $1 FOR UPDATE`)).
		WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(12.5))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET balance = GREATEST(balance - $1, 0), updated_at = NOW() WHERE id = $2`)).
		WithArgs(12.5, int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO redeem_codes (code, type, value, status, used_by, used_at, notes, created_at)
		VALUES ($1, $2, $3, 'used', $4, NOW(), $5, NOW())`)).
		WithArgs(sqlmock.AnyArg(), AdjustmentTypeAdminBalance, -12.5, int64(7), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE timed_user_grants SET status = 'expired', expired_at = $2, deducted_amount = $3, updated_at = NOW() WHERE id = $1`)).
		WithArgs(int64(21), now, 12.5).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT concurrency FROM users WHERE id = $1 FOR UPDATE`)).
		WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"concurrency"}).AddRow(3))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET concurrency = concurrency - $1, updated_at = NOW() WHERE id = $2`)).
		WithArgs(3, int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO redeem_codes (code, type, value, status, used_by, used_at, notes, created_at)
		VALUES ($1, $2, $3, 'used', $4, NOW(), $5, NOW())`)).
		WithArgs(sqlmock.AnyArg(), AdjustmentTypeAdminConcurrency, -3.0, int64(7), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE timed_user_grants SET status = 'expired', expired_at = $2, deducted_amount = $3, updated_at = NOW() WHERE id = $1`)).
		WithArgs(int64(22), now, float64(3)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	expired, err := svc.ExpireDueGrants(context.Background(), now, 10)

	require.NoError(t, err)
	require.Equal(t, 2, expired)
	require.Equal(t, []int64{7, 7}, invalidator.userIDs)
	require.NoError(t, mock.ExpectationsWereMet())
}
