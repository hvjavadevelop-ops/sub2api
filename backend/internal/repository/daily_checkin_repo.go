package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
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

func (r *dailyCheckinRepository) ListUserRecent(ctx context.Context, userID int64, limit int) ([]service.DailyCheckinRecord, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("daily checkin repository is not initialized")
	}
	if limit <= 0 || limit > 25 {
		limit = 25
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT id, user_id, checkin_date, reward, balance_after, created_at
FROM user_daily_checkins
WHERE user_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2
`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := make([]service.DailyCheckinRecord, 0)
	for rows.Next() {
		var record service.DailyCheckinRecord
		if err := rows.Scan(&record.ID, &record.UserID, &record.CheckinDate, &record.Reward, &record.BalanceAfter, &record.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return records, nil
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

func (r *dailyCheckinRepository) ListAdmin(ctx context.Context, params service.DailyCheckinListParams) ([]service.DailyCheckinRecord, int, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("daily checkin repository is not initialized")
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 || params.PageSize > 100 {
		params.PageSize = 20
	}
	where := []string{"1=1"}
	args := []any{}
	add := func(v any) string { args = append(args, v); return fmt.Sprintf("$%d", len(args)) }
	if params.UserID > 0 {
		where = append(where, "c.user_id = "+add(params.UserID))
	}
	if q := strings.TrimSpace(params.Search); q != "" {
		ph := add("%" + strings.ToLower(q) + "%")
		clauses := []string{"LOWER(u.email) LIKE " + ph, "LOWER(u.username) LIKE " + ph}
		if id, err := strconv.ParseInt(q, 10, 64); err == nil && id > 0 {
			clauses = append(clauses, "c.user_id = "+add(id))
		}
		where = append(where, "("+strings.Join(clauses, " OR ")+")")
	}
	if params.StartDate != "" {
		where = append(where, "c.checkin_date >= "+add(params.StartDate))
	}
	if params.EndDate != "" {
		where = append(where, "c.checkin_date <= "+add(params.EndDate))
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	countSQL := `SELECT COUNT(*) FROM user_daily_checkins c LEFT JOIN users u ON u.id = c.user_id WHERE ` + whereSQL
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	limitPh := add(params.PageSize)
	offsetPh := add((params.Page - 1) * params.PageSize)
	query := `
SELECT c.id, c.user_id, COALESCE(u.email, ''), COALESCE(u.username, ''), c.checkin_date, c.reward, c.balance_after, c.created_at
FROM user_daily_checkins c
LEFT JOIN users u ON u.id = c.user_id
WHERE ` + whereSQL + `
ORDER BY c.created_at DESC, c.id DESC
LIMIT ` + limitPh + ` OFFSET ` + offsetPh
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	records := make([]service.DailyCheckinRecord, 0)
	for rows.Next() {
		var record service.DailyCheckinRecord
		if err := rows.Scan(&record.ID, &record.UserID, &record.UserEmail, &record.Username, &record.CheckinDate, &record.Reward, &record.BalanceAfter, &record.CreatedAt); err != nil {
			return nil, 0, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return records, total, nil
}
