package service

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	dailyCheckinMinReward = 10
	dailyCheckinMaxReward = 29
)

var (
	ErrDailyCheckinNotFound    = infraerrors.NotFound("DAILY_CHECKIN_NOT_FOUND", "daily check-in record not found")
	ErrDailyCheckinAlreadyDone = infraerrors.Conflict("DAILY_CHECKIN_ALREADY_DONE", "already checked in today")
)

type DailyCheckinRecord struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	UserEmail    string    `json:"user_email,omitempty"`
	Username     string    `json:"username,omitempty"`
	CheckinDate  time.Time `json:"checkin_date"`
	Reward       float64   `json:"reward"`
	BalanceAfter float64   `json:"balance_after"`
	CreatedAt    time.Time `json:"created_at"`
}

type DailyCheckinListResult struct {
	Records []DailyCheckinRecord `json:"records"`
	Total   int                  `json:"total"`
	Page    int                  `json:"page"`
	Limit   int                  `json:"limit"`
}

type DailyCheckinResult struct {
	CheckedInToday bool                `json:"checked_in_today"`
	MinReward      float64             `json:"min_reward"`
	MaxReward      float64             `json:"max_reward"`
	Reward         float64             `json:"reward,omitempty"`
	BalanceAfter   float64             `json:"balance_after,omitempty"`
	Today          *DailyCheckinRecord `json:"today,omitempty"`
}

type DailyCheckinRepository interface {
	GetToday(ctx context.Context, userID int64, now time.Time) (*DailyCheckinRecord, error)
	CreateToday(ctx context.Context, userID int64, reward float64, now time.Time) (*DailyCheckinRecord, error)
	List(ctx context.Context, page, limit int) (*DailyCheckinListResult, error)
}

type randomIntFunc func(max int) (int, error)

type DailyCheckinService struct {
	repo      DailyCheckinRepository
	randomInt randomIntFunc
}

func NewDailyCheckinService(repo DailyCheckinRepository, randomInt randomIntFunc) *DailyCheckinService {
	if randomInt == nil {
		randomInt = secureRandomInt
	}
	return &DailyCheckinService{repo: repo, randomInt: randomInt}
}

func (s *DailyCheckinService) Status(ctx context.Context, userID int64, now time.Time) (*DailyCheckinResult, error) {
	record, err := s.repo.GetToday(ctx, userID, now)
	if err != nil {
		if errors.Is(err, ErrDailyCheckinNotFound) {
			return &DailyCheckinResult{CheckedInToday: false, MinReward: dailyCheckinMinReward, MaxReward: dailyCheckinMaxReward}, nil
		}
		return nil, err
	}
	return &DailyCheckinResult{
		CheckedInToday: true,
		MinReward:      dailyCheckinMinReward,
		MaxReward:      dailyCheckinMaxReward,
		Reward:         record.Reward,
		BalanceAfter:   record.BalanceAfter,
		Today:          record,
	}, nil
}

func (s *DailyCheckinService) Checkin(ctx context.Context, userID int64, now time.Time) (*DailyCheckinResult, error) {
	if existing, err := s.repo.GetToday(ctx, userID, now); err == nil && existing != nil {
		return nil, ErrDailyCheckinAlreadyDone
	} else if err != nil && !errors.Is(err, ErrDailyCheckinNotFound) {
		return nil, err
	}

	rewardOffset, err := s.randomInt(dailyCheckinMaxReward - dailyCheckinMinReward + 1)
	if err != nil {
		return nil, err
	}
	reward := float64(dailyCheckinMinReward + rewardOffset)
	record, err := s.repo.CreateToday(ctx, userID, reward, now)
	if err != nil {
		return nil, err
	}
	return &DailyCheckinResult{
		CheckedInToday: true,
		MinReward:      dailyCheckinMinReward,
		MaxReward:      dailyCheckinMaxReward,
		Reward:         record.Reward,
		BalanceAfter:   record.BalanceAfter,
		Today:          record,
	}, nil
}

func (s *DailyCheckinService) List(ctx context.Context, page, limit int) (*DailyCheckinListResult, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.List(ctx, page, limit)
}

func secureRandomInt(max int) (int, error) {
	if max <= 0 {
		return 0, nil
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}
