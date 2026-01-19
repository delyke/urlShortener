package audit

import (
	"context"
)

// Event describes a single audit action emitted by the service
type Event struct {
	TS     int64  `json:"ts"`
	Action string `json:"action"`
	UserID string `json:"user_id,omitempty"`
	URL    string `json:"url"`
}

// Observer receives audit events from the service
type Observer interface {
	OnEvent(ctx context.Context, e Event) error
}
