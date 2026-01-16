package audit

import (
	"context"
)

type Event struct {
	TS     int64  `json:"ts"`
	Action string `json:"action"`
	UserID string `json:"user_id,omitempty"`
	URL    string `json:"url"`
}

type Observer interface {
	OnEvent(ctx context.Context, e Event) error
}
