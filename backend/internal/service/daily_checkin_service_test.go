package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeDailyCheckinRepo struct {
	balance      float64
	records      map[string]*DailyCheckinRecord
	recent       []DailyCheckinRecord
	createdCount int
}

func newFakeDailyCheckinRepo(balance float64) *fakeDailyCheckinRepo {
	return &fakeDailyCheckinRepo{balance: balance, records: make(map[string]*DailyCheckinRecord)}
}

func (r *fakeDailyCheckinRepo) key(userID int64, date time.Time) string {
	return date.Format("2006-01-02")
}

func (r *fakeDailyCheckinRepo) GetToday(ctx context.Context, userID int64, now time.Time) (*DailyCheckinRecord, error) {
	if record := r.records[r.key(userID, now)]; record != nil {
		copy := *record
		return &copy, nil
	}
	return nil, ErrDailyCheckinNotFound
}

func (r *fakeDailyCheckinRepo) CreateToday(ctx context.Context, userID int64, reward float64, now time.Time) (*DailyCheckinRecord, error) {
	key := r.key(userID, now)
	if _, exists := r.records[key]; exists {
		return nil, ErrDailyCheckinAlreadyDone
	}
	r.balance += reward
	r.createdCount++
	record := &DailyCheckinRecord{
		ID:           int64(r.createdCount),
		UserID:       userID,
		CheckinDate:  time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
		Reward:       reward,
		BalanceAfter: r.balance,
		CreatedAt:    now,
	}
	r.records[key] = record
	copy := *record
	return &copy, nil
}

func (r *fakeDailyCheckinRepo) ListUserRecent(ctx context.Context, userID int64, limit int) ([]DailyCheckinRecord, error) {
	items := make([]DailyCheckinRecord, 0)
	for _, record := range r.recent {
		if record.UserID == userID {
			items = append(items, record)
		}
	}
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func TestDailyCheckinService_Checkin_GrantsRewardAndOnlyOncePerDay(t *testing.T) {
	now := time.Date(2026, 4, 28, 10, 0, 0, 0, time.Local)
	repo := newFakeDailyCheckinRepo(100)
	svc := NewDailyCheckinService(repo, func(max int) (int, error) { return 7, nil })

	result, err := svc.Checkin(context.Background(), 42, now)
	if err != nil {
		t.Fatalf("first checkin failed: %v", err)
	}
	if result.Reward != 17 {
		t.Fatalf("reward = %v, want 17", result.Reward)
	}
	if result.BalanceAfter != 117 {
		t.Fatalf("balance after = %v, want 117", result.BalanceAfter)
	}
	if !result.CheckedInToday {
		t.Fatalf("checked_in_today should be true")
	}

	_, err = svc.Checkin(context.Background(), 42, now.Add(2*time.Hour))
	if !errors.Is(err, ErrDailyCheckinAlreadyDone) {
		t.Fatalf("second checkin err = %v, want ErrDailyCheckinAlreadyDone", err)
	}
	if repo.createdCount != 1 {
		t.Fatalf("createdCount = %d, want 1", repo.createdCount)
	}
}

func TestDailyCheckinService_Status(t *testing.T) {
	now := time.Date(2026, 4, 28, 10, 0, 0, 0, time.Local)
	repo := newFakeDailyCheckinRepo(0)
	svc := NewDailyCheckinService(repo, func(max int) (int, error) { return 0, nil })

	status, err := svc.Status(context.Background(), 1, now)
	if err != nil {
		t.Fatalf("status before checkin failed: %v", err)
	}
	if status.CheckedInToday {
		t.Fatalf("expected not checked in before checkin")
	}
	if status.MinReward != 10 || status.MaxReward != 29 {
		t.Fatalf("reward range = %v-%v, want 10-29", status.MinReward, status.MaxReward)
	}

	if _, err := svc.Checkin(context.Background(), 1, now); err != nil {
		t.Fatalf("checkin failed: %v", err)
	}
	status, err = svc.Status(context.Background(), 1, now)
	if err != nil {
		t.Fatalf("status after checkin failed: %v", err)
	}
	if !status.CheckedInToday || status.Today == nil {
		t.Fatalf("expected checked in today with record")
	}
}

func TestDailyCheckinService_ListUserRecent(t *testing.T) {
	now := time.Date(2026, 4, 28, 10, 0, 0, 0, time.UTC)
	repo := newFakeDailyCheckinRepo(0)
	repo.recent = append(repo.recent,
		DailyCheckinRecord{ID: 1, UserID: 1, Reward: 20, CreatedAt: now.Add(-time.Hour)},
		DailyCheckinRecord{ID: 2, UserID: 2, Reward: 25, CreatedAt: now},
		DailyCheckinRecord{ID: 3, UserID: 1, Reward: 29, CreatedAt: now},
	)
	svc := NewDailyCheckinService(repo, nil)

	records, err := svc.ListUserRecent(context.Background(), 1, 25)

	if err != nil {
		t.Fatalf("ListUserRecent returned error: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 user records, got %d", len(records))
	}
	for _, record := range records {
		if record.UserID != 1 {
			t.Fatalf("expected only user 1 records, got user %d", record.UserID)
		}
	}
}
