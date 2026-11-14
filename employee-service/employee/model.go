package employee

import (
	"time"
)

type AccessRequest struct {
	ID             string    `json:"id"`
	UserID         string    `json:"userId" db:"user_id"`
	System         string    `json:"system"`
	Status         string    `json:"status"`
	ExternalID     string    `json:"externalId" db:"external_id"`
	IdempotencyKey string    `json:"idempotencyKey" db:"idempotency_key"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
}
