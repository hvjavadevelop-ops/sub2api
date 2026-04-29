package service

import "time"

const (
	TimedGrantTypeBalance     = "balance"
	TimedGrantTypeConcurrency = "concurrency"

	TimedGrantStatusPending   = "pending"
	TimedGrantStatusActive    = "active"
	TimedGrantStatusExpired   = "expired"
	TimedGrantStatusCancelled = "cancelled"
)

type TimedUserGrant struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"user_id"`
	GrantType       string     `json:"grant_type"`
	Amount          float64    `json:"amount"`
	DurationSeconds int        `json:"duration_seconds"`
	Status          string     `json:"status"`
	ActivatedAt     *time.Time `json:"activated_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	ExpiredAt       *time.Time `json:"expired_at,omitempty"`
	DeductedAmount  float64    `json:"deducted_amount"`
	CreatedBy       *int64     `json:"created_by,omitempty"`
	Notes           string     `json:"notes"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateTimedUserGrantInput struct {
	GrantType       string  `json:"grant_type"`
	Amount          float64 `json:"amount"`
	DurationSeconds int     `json:"duration_seconds"`
	Notes           string  `json:"notes"`
	CreatedBy       *int64  `json:"-"`
}
