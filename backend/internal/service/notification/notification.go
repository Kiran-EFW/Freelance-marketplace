package notification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/adapter/push"
	"github.com/seva-platform/backend/internal/adapter/sms"
	"github.com/seva-platform/backend/internal/domain"
	"github.com/seva-platform/backend/internal/repository/postgres"
)

// Service defines the notification service interface.
type Service interface {
	Send(ctx context.Context, userID uuid.UUID, notifType domain.NotificationType, title, body string, data map[string]interface{}) error
	SendSMS(ctx context.Context, phone, message string) error
	SendBulk(ctx context.Context, userIDs []uuid.UUID, notifType domain.NotificationType, title, body string) error
	MarkRead(ctx context.Context, notificationID, userID uuid.UUID) error
	GetUnread(ctx context.Context, userID uuid.UUID) ([]domain.Notification, error)
	SendSeasonalReminder(ctx context.Context, cropID uuid.UUID, jurisdictionID string) error
}

// NotificationService delivers notifications via push, SMS, email, or in-app.
type NotificationService struct {
	notifications domain.NotificationRepository
	users         domain.UserRepository
	smsProvider   sms.SMSProvider
	pushProvider  push.Provider
	queries       *postgres.Queries
}

// NewNotificationService returns a ready-to-use NotificationService.
func NewNotificationService(
	notifications domain.NotificationRepository,
	users domain.UserRepository,
	smsProvider sms.SMSProvider,
	pushProvider push.Provider,
	queries *postgres.Queries,
) *NotificationService {
	return &NotificationService{
		notifications: notifications,
		users:         users,
		smsProvider:   smsProvider,
		pushProvider:  pushProvider,
		queries:       queries,
	}
}

// Send creates and delivers a notification to a single user.
func (s *NotificationService) Send(ctx context.Context, userID uuid.UUID, notifType domain.NotificationType, title, body string, data map[string]interface{}) error {
	rawData, err := json.Marshal(data)
	if err != nil {
		rawData = []byte("{}")
	}

	now := time.Now()
	notif := &domain.Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      notifType,
		Title:     title,
		Body:      body,
		Data:      rawData,
		Channel:   domain.ChannelPush,
		SentAt:    &now,
	}

	// Determine the best channel for this user.
	user, err := s.users.GetByID(ctx, userID)
	if err == nil && user.DeviceType == "basic_phone" {
		notif.Channel = domain.ChannelSMS
	}

	if err := s.notifications.Create(ctx, notif); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to persist notification")
		return fmt.Errorf("create notification: %w", err)
	}

	// Deliver via the appropriate channel.
	switch notif.Channel {
	case domain.ChannelSMS:
		if user != nil {
			if err := s.smsProvider.SendSMS(user.Phone, body); err != nil {
				log.Warn().Err(err).Str("phone", user.Phone).Msg("SMS delivery failed")
			}
		}
	case domain.ChannelPush:
		s.deliverPush(ctx, userID, title, body, data)
	case domain.ChannelEmail:
		// TODO: integrate with email service.
		log.Debug().Str("user_id", userID.String()).Msg("email notification queued")
	}

	log.Info().
		Str("notification_id", notif.ID.String()).
		Str("user_id", userID.String()).
		Str("type", string(notifType)).
		Str("channel", string(notif.Channel)).
		Msg("notification sent")

	return nil
}

// SendSMS sends a raw SMS message to a phone number.
func (s *NotificationService) SendSMS(ctx context.Context, phone, message string) error {
	if phone == "" {
		return fmt.Errorf("%w: phone is required", domain.ErrInvalidInput)
	}

	if err := s.smsProvider.SendSMS(phone, message); err != nil {
		log.Error().Err(err).Str("phone", phone).Msg("SMS send failed")
		return fmt.Errorf("send SMS: %w", err)
	}

	log.Info().Str("phone", phone).Msg("SMS sent")
	return nil
}

// SendBulk delivers the same notification to multiple users.
func (s *NotificationService) SendBulk(ctx context.Context, userIDs []uuid.UUID, notifType domain.NotificationType, title, body string) error {
	var lastErr error
	for _, userID := range userIDs {
		if err := s.Send(ctx, userID, notifType, title, body, nil); err != nil {
			log.Warn().Err(err).Str("user_id", userID.String()).Msg("bulk notification failed for user")
			lastErr = err
		}
	}
	if lastErr != nil {
		return fmt.Errorf("some bulk notifications failed: %w", lastErr)
	}
	return nil
}

// MarkRead marks a notification as read by the user.
func (s *NotificationService) MarkRead(ctx context.Context, notificationID, userID uuid.UUID) error {
	notif, err := s.notifications.GetByID(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("%w: notification %s", domain.ErrNotFound, notificationID)
	}

	if notif.UserID != userID {
		return fmt.Errorf("%w: notification belongs to another user", domain.ErrUnauthorized)
	}

	now := time.Now()
	if err := s.notifications.MarkRead(ctx, notificationID, now); err != nil {
		return fmt.Errorf("mark read: %w", err)
	}
	return nil
}

// GetUnread returns all unread notifications for a user.
func (s *NotificationService) GetUnread(ctx context.Context, userID uuid.UUID) ([]domain.Notification, error) {
	notifs, err := s.notifications.ListUnread(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get unread: %w", err)
	}
	return notifs, nil
}

// SendSeasonalReminder sends seasonal work reminders to customers in a
// jurisdiction who have crops/land matching the given crop catalogue entry.
func (s *NotificationService) SendSeasonalReminder(ctx context.Context, cropID uuid.UUID, jurisdictionID string) error {
	// In a full implementation this would:
	// 1. Look up the crop catalog entry to get the crop name and seasonal work types.
	// 2. Query customers in the jurisdiction who have previously booked crop-related jobs.
	// 3. Send each customer a reminder about upcoming seasonal work.

	title := "Seasonal reminder"
	body := "It's time for seasonal maintenance! Book a service provider now."

	log.Info().
		Str("crop_id", cropID.String()).
		Str("jurisdiction", jurisdictionID).
		Msg("seasonal reminder campaign triggered")

	// Placeholder: in production, fetch the list of target user IDs and call SendBulk.
	_ = title
	_ = body

	return nil
}

// deliverPush sends a push notification to all active device tokens for a user.
// If FCM reports a token as invalid/unregistered, the token is deactivated.
func (s *NotificationService) deliverPush(ctx context.Context, userID uuid.UUID, title, body string, data map[string]interface{}) {
	if s.pushProvider == nil || s.queries == nil {
		log.Debug().Str("user_id", userID.String()).Msg("push provider or queries not configured, skipping push delivery")
		return
	}

	// Look up active device tokens for this user.
	tokens, err := s.queries.GetDeviceTokensForUser(ctx, userID)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to fetch device tokens for push delivery")
		return
	}

	if len(tokens) == 0 {
		log.Debug().Str("user_id", userID.String()).Msg("no active device tokens for user, skipping push delivery")
		return
	}

	// Convert data to string map for push notification.
	pushData := make(map[string]string)
	for k, v := range data {
		pushData[k] = fmt.Sprintf("%v", v)
	}

	notification := push.Notification{
		Title: title,
		Body:  body,
		Data:  pushData,
	}

	for _, dt := range tokens {
		if err := s.pushProvider.SendToDevice(ctx, dt.Token, notification); err != nil {
			if errors.Is(err, push.ErrInvalidToken) {
				// Deactivate the invalid token.
				log.Info().
					Str("token", dt.Token).
					Str("user_id", userID.String()).
					Msg("deactivating invalid device token")
				if deactivateErr := s.queries.DeactivateDeviceToken(ctx, dt.Token); deactivateErr != nil {
					log.Error().Err(deactivateErr).Str("token", dt.Token).Msg("failed to deactivate device token")
				}
			} else {
				log.Warn().Err(err).Str("token", dt.Token).Str("user_id", userID.String()).Msg("push delivery failed")
			}
		}
	}
}
