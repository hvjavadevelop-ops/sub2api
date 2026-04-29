package service

import (
	"context"
	"errors"
)

func (s *adminServiceImpl) CreateTimedUserGrant(ctx context.Context, userID int64, input CreateTimedUserGrantInput) (*TimedUserGrant, error) {
	if s == nil || s.timedGrantService == nil {
		return nil, errors.New("timed grant service unavailable")
	}
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		return nil, err
	}
	return s.timedGrantService.Create(ctx, userID, input)
}

func (s *adminServiceImpl) ListTimedUserGrants(ctx context.Context, userID int64) ([]TimedUserGrant, error) {
	if s == nil || s.timedGrantService == nil {
		return nil, errors.New("timed grant service unavailable")
	}
	return s.timedGrantService.ListByUser(ctx, userID)
}
