package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type dailyCheckinRepository struct {
	db *sql.DB
}

func NewDailyCheckinRepository(db *sql.DB) service.DailyCheckinRepository {
	return &dailyCheckinRepository{db: db}
}

func (r *dailyCheckinRepository) GetToday(ctx context.Context, userID int64, now time.Time) (*service.DailyCheckinRecord, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("daily checkin repository is not initialized")
	}
	var record service.DailyCheckinRecord
	err := r.db.QueryRowContext(ctx, `
SELECT id, user_id, checkin_date, reward, balance_after, created_at
FROM user_daily_checkins
WHERE user_id = $1 AND checkin_date = CURRENT_DATE
`, userID).Scan(&record.ID, &record.UserID, &record.CheckinDate, &record.Reward, &record.BalanceAfter, &record.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrDailyCheckinNotFound
		}
		return nil, err
	}
	return &record, nil
}

func (r *dailyCheckinRepository) List(ctx context.Context, page, limit int) (*service.DailyCheckinListResult, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("daily checkin repository is not initialized")
	}
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_daily_checkins`).Scan(&total); err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT c.id, c.user_id, COALESCE(u.email, ''), COALESCE(u.username, ''), c.checkin_date, c.reward, c.balance_after, c.created_at
FROM user_daily_checkins c
LEFT JOIN users u ON u.id = c.user_id
ORDER BY c.created_at DESC, c.id DESC
LIMIT $1 OFFSET $2
`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := make([]service.DailyCheckinRecord, 0)
	for rows.Next() {
		var record service.DailyCheckinRecord
		if err := rows.Scan(&record.ID, &record.UserID, &record.UserEmail, &record.Username, &record.CheckinDate, &record.Reward, &record.BalanceAfter, &record.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &service.DailyCheckinListResult{Records: records, Total: total, Page: page, Limit: limit}, nil
}

func (r *dailyCheckinRepository) CreateToday(ctx context.Context, userID int64, reward float64, now time.Time) (*service.DailyCheckinRecord, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("daily checkin repository is not initialized")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var record service.DailyCheckinRecord
	err = tx.QueryRowContext(ctx, `
WITH updated_user AS (
    UPDATE users
    SET balance = balance + $2,
        total_recharged = total_recharged + $2,
        updated_at = NOW()
    WHERE id = $1
    RETURNING balance
), inserted AS (
    INSERT INTO user_daily_checkins (user_id, checkin_date, reward, balance_after, created_at, updated_at)
    SELECT $1, CURRENT_DATE, $2, balance, NOW(), NOW()
    FROM updated_user
    ON CONFLICT (user_id, checkin_date) DO NOTHING
    RETURNING id, user_id, checkin_date, reward, balance_after, created_at
)
SELECT id, user_id, checkin_date, reward, balance_after, created_at FROM inserted
`, userID, reward).Scan(&record.ID, &record.UserID, &record.CheckinDate, &record.Reward, &record.BalanceAfter, &record.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrDailyCheckinAlreadyDone
		}
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &record, nil
}
