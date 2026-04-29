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
	DailyCheckinDefaultMinReward = 10
	DailyCheckinDefaultMaxReward = 29
	DailyCheckinMinAllowedReward = 0
	DailyCheckinMaxAllowedReward = 1000000
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
	ListUserRecent(ctx context.Context, userID int64, limit int) ([]DailyCheckinRecord, error)
	ListAdmin(ctx context.Context, params DailyCheckinListParams) ([]DailyCheckinRecord, int, error)
}

type DailyCheckinListParams struct {
	Page      int
	PageSize  int
	Search    string
	UserID    int64
	StartDate string
	EndDate   string
}

type DailyCheckinSettingsProvider interface {
	GetDailyCheckinRewardRange(ctx context.Context) (minReward, maxReward int)
}

type randomIntFunc func(max int) (int, error)

type DailyCheckinService struct {
	repo             DailyCheckinRepository
	settingsProvider DailyCheckinSettingsProvider
	randomInt        randomIntFunc
}

func NewDailyCheckinService(repo DailyCheckinRepository, settingsProvider DailyCheckinSettingsProvider, randomInt randomIntFunc) *DailyCheckinService {
	if randomInt == nil {
		randomInt = secureRandomInt
	}
	return &DailyCheckinService{repo: repo, settingsProvider: settingsProvider, randomInt: randomInt}
}

func normalizeDailyCheckinRewardRange(minReward, maxReward int) (int, int) {
	if minReward < DailyCheckinMinAllowedReward || maxReward < DailyCheckinMinAllowedReward || minReward > maxReward {
		return DailyCheckinDefaultMinReward, DailyCheckinDefaultMaxReward
	}
	if maxReward > DailyCheckinMaxAllowedReward {
		return DailyCheckinDefaultMinReward, DailyCheckinDefaultMaxReward
	}
	return minReward, maxReward
}

func (s *DailyCheckinService) rewardRange(ctx context.Context) (int, int) {
	if s.settingsProvider == nil {
		return DailyCheckinDefaultMinReward, DailyCheckinDefaultMaxReward
	}
	return normalizeDailyCheckinRewardRange(s.settingsProvider.GetDailyCheckinRewardRange(ctx))
}

func (s *DailyCheckinService) Status(ctx context.Context, userID int64, now time.Time) (*DailyCheckinResult, error) {
	record, err := s.repo.GetToday(ctx, userID, now)
	if err != nil {
		if errors.Is(err, ErrDailyCheckinNotFound) {
			minReward, maxReward := s.rewardRange(ctx)
			return &DailyCheckinResult{CheckedInToday: false, MinReward: float64(minReward), MaxReward: float64(maxReward)}, nil
		}
		return nil, err
	}
	minReward, maxReward := s.rewardRange(ctx)
	return &DailyCheckinResult{
		CheckedInToday: true,
		MinReward:      float64(minReward),
		MaxReward:      float64(maxReward),
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

	minReward, maxReward := s.rewardRange(ctx)
	rewardOffset, err := s.randomInt(maxReward - minReward + 1)
	if err != nil {
		return nil, err
	}
	reward := float64(minReward + rewardOffset)
	record, err := s.repo.CreateToday(ctx, userID, reward, now)
	if err != nil {
		return nil, err
	}
	return &DailyCheckinResult{
		CheckedInToday: true,
		MinReward:      float64(minReward),
		MaxReward:      float64(maxReward),
		Reward:         record.Reward,
		BalanceAfter:   record.BalanceAfter,
		Today:          record,
	}, nil
}

func (s *DailyCheckinService) ListUserRecent(ctx context.Context, userID int64, limit int) ([]DailyCheckinRecord, error) {
	if limit <= 0 || limit > 25 {
		limit = 25
	}
	return s.repo.ListUserRecent(ctx, userID, limit)
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

func (s *DailyCheckinService) ListAdmin(ctx context.Context, params DailyCheckinListParams) ([]DailyCheckinRecord, int, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 || params.PageSize > 100 {
		params.PageSize = 20
	}
	return s.repo.ListAdmin(ctx, params)
}
