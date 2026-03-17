package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/domain"
)

// Service defines the user service interface.
type Service interface {
	Register(ctx context.Context, phone string, userType domain.UserType, jurisdictionID string) (*domain.User, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, params domain.UpdateProfileParams) error
	Deactivate(ctx context.Context, userID uuid.UUID) error
	GetOrCreateByPhone(ctx context.Context, phone string) (*domain.User, error)
}

// UserService implements user-related business logic.
type UserService struct {
	users domain.UserRepository
	cache domain.CacheStore
}

// NewUserService returns a ready-to-use UserService.
func NewUserService(users domain.UserRepository, cache domain.CacheStore) *UserService {
	return &UserService{
		users: users,
		cache: cache,
	}
}

// Register creates a new user account and triggers OTP delivery.
func (s *UserService) Register(ctx context.Context, phone string, userType domain.UserType, jurisdictionID string) (*domain.User, error) {
	if phone == "" {
		return nil, fmt.Errorf("%w: phone is required", domain.ErrInvalidInput)
	}

	// Check whether the phone is already registered.
	existing, err := s.users.GetByPhone(ctx, phone)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("%w: phone already registered", domain.ErrAlreadyExists)
	}

	user := &domain.User{
		ID:                uuid.New(),
		Type:              userType,
		Phone:             phone,
		JurisdictionID:    jurisdictionID,
		PreferredLanguage: "en",
	}

	if err := s.users.Create(ctx, user); err != nil {
		log.Error().Err(err).Str("phone", phone).Msg("failed to create user")
		return nil, fmt.Errorf("create user: %w", err)
	}

	log.Info().
		Str("user_id", user.ID.String()).
		Str("phone", phone).
		Str("type", string(userType)).
		Msg("user registered")

	// TODO: trigger OTP send via notification service

	return user, nil
}

// GetProfile returns the full user record.
func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: user %s", domain.ErrNotFound, userID)
	}
	return user, nil
}

// UpdateProfile applies partial updates to a user's profile.
func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, params domain.UpdateProfileParams) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("%w: user %s", domain.ErrNotFound, userID)
	}

	if params.Name != nil {
		user.Name = *params.Name
	}
	if params.Email != nil {
		user.Email = params.Email
	}
	if params.PreferredLanguage != nil {
		user.PreferredLanguage = *params.PreferredLanguage
	}
	if params.DeviceType != nil {
		user.DeviceType = *params.DeviceType
	}

	if err := s.users.Update(ctx, user); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to update user")
		return fmt.Errorf("update user: %w", err)
	}

	// Invalidate cache.
	_ = s.cache.Delete(fmt.Sprintf("user:%s", userID))

	log.Info().Str("user_id", userID.String()).Msg("user profile updated")
	return nil
}

// Deactivate soft-deletes a user account.
func (s *UserService) Deactivate(ctx context.Context, userID uuid.UUID) error {
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		return fmt.Errorf("%w: user %s", domain.ErrNotFound, userID)
	}

	if err := s.users.Delete(ctx, userID); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to deactivate user")
		return fmt.Errorf("deactivate user: %w", err)
	}

	_ = s.cache.Delete(fmt.Sprintf("user:%s", userID))

	log.Info().Str("user_id", userID.String()).Msg("user deactivated")
	return nil
}

// GetOrCreateByPhone finds an existing user by phone number or creates a new
// one. This is the primary entry-point for SMS-based authentication flows.
func (s *UserService) GetOrCreateByPhone(ctx context.Context, phone string) (*domain.User, error) {
	if phone == "" {
		return nil, fmt.Errorf("%w: phone is required", domain.ErrInvalidInput)
	}

	existing, err := s.users.GetByPhone(ctx, phone)
	if err == nil && existing != nil {
		return existing, nil
	}

	// Create a new customer by default for SMS flows.
	user := &domain.User{
		ID:                uuid.New(),
		Type:              domain.UserTypeCustomer,
		Phone:             phone,
		JurisdictionID:    "in", // default jurisdiction
		PreferredLanguage: "en",
	}

	if err := s.users.Create(ctx, user); err != nil {
		log.Error().Err(err).Str("phone", phone).Msg("failed to create user via SMS flow")
		return nil, fmt.Errorf("create user: %w", err)
	}

	log.Info().
		Str("user_id", user.ID.String()).
		Str("phone", phone).
		Msg("user created via SMS flow")

	return user, nil
}
