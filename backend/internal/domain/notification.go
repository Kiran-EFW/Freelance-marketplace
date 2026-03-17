package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// NotificationType identifies the category of a notification.
type NotificationType string

const (
	NotifJobNew           NotificationType = "job_new"
	NotifJobAccepted      NotificationType = "job_accepted"
	NotifJobCompleted     NotificationType = "job_completed"
	NotifPaymentReceived  NotificationType = "payment_received"
	NotifReviewReceived   NotificationType = "review_received"
	NotifDisputeUpdate    NotificationType = "dispute_update"
	NotifPointsEarned     NotificationType = "points_earned"
	NotifLevelUp          NotificationType = "level_up"
	NotifRouteVisit       NotificationType = "route_visit"
	NotifSeasonalReminder NotificationType = "seasonal_reminder"
)

// NotificationChannel is the delivery channel for a notification.
type NotificationChannel string

const (
	ChannelPush  NotificationChannel = "push"
	ChannelSMS   NotificationChannel = "sms"
	ChannelEmail NotificationChannel = "email"
	ChannelInApp NotificationChannel = "in_app"
)

// Notification is a message delivered to a user through one or more channels.
type Notification struct {
	ID        uuid.UUID           `json:"id" db:"id"`
	UserID    uuid.UUID           `json:"user_id" db:"user_id"`
	Type      NotificationType    `json:"type" db:"type"`
	Title     string              `json:"title" db:"title"`
	Body      string              `json:"body,omitempty" db:"body"`
	Data      json.RawMessage     `json:"data,omitempty" db:"data"`
	Channel   NotificationChannel `json:"channel" db:"channel"`
	SentAt    *time.Time          `json:"sent_at,omitempty" db:"sent_at"`
	ReadAt    *time.Time          `json:"read_at,omitempty" db:"read_at"`
	CreatedAt time.Time           `json:"created_at" db:"created_at"`
}

// NotificationPreference stores a user's per-channel notification settings.
type NotificationPreference struct {
	UserID  uuid.UUID           `json:"user_id" db:"user_id"`
	Channel NotificationChannel `json:"channel" db:"channel"`
	Enabled bool                `json:"enabled" db:"enabled"`
}

// NotificationRepository defines persistence operations for notifications and
// notification preferences.
type NotificationRepository interface {
	Create(ctx context.Context, n *Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Notification, error)
	ListUnread(ctx context.Context, userID uuid.UUID) ([]Notification, error)
	MarkRead(ctx context.Context, id uuid.UUID, readAt time.Time) error
	MarkAllRead(ctx context.Context, userID uuid.UUID, readAt time.Time) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int, error)

	GetPreferences(ctx context.Context, userID uuid.UUID) ([]NotificationPreference, error)
	UpsertPreference(ctx context.Context, pref *NotificationPreference) error
}
