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

func (r *fakeDailyCheckinRepo) List(ctx context.Context, page, limit int) (*DailyCheckinListResult, error) {
	return &DailyCheckinListResult{Records: []DailyCheckinRecord{}, Total: 0, Page: page, Limit: limit}, nil
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
