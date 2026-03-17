package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ServiceModel represents the type of service delivery model.
type ServiceModel string

const (
	ServiceModelRouteBased        ServiceModel = "route_based"
	ServiceModelSeasonalContract  ServiceModel = "seasonal_contract"
	ServiceModelOnDemand          ServiceModel = "on_demand"
	ServiceModelPeriodic          ServiceModel = "periodic"
)

// RouteStatus represents the current state of a route.
type RouteStatus string

const (
	RouteStatusActive   RouteStatus = "active"
	RouteStatusPaused   RouteStatus = "paused"
	RouteStatusInactive RouteStatus = "inactive"
)

// RouteStopStatus represents the state of a stop on a route.
type RouteStopStatus string

const (
	RouteStopStatusActive   RouteStopStatus = "active"
	RouteStopStatusPaused   RouteStopStatus = "paused"
	RouteStopStatusRemoved  RouteStopStatus = "removed"
)

// RouteRequestStatus represents the state of a customer's request to join a route.
type RouteRequestStatus string

const (
	RouteRequestPending  RouteRequestStatus = "pending"
	RouteRequestAccepted RouteRequestStatus = "accepted"
	RouteRequestDeclined RouteRequestStatus = "declined"
)

// Route represents a recurring service circuit for a provider (e.g., a coconut
// tree climber visiting houses on a fixed schedule).
type Route struct {
	ID           uuid.UUID   `json:"id" db:"id"`
	ProviderID   uuid.UUID   `json:"provider_id" db:"provider_id"`
	Name         string      `json:"name" db:"name"`
	PostcodeArea string      `json:"postcode_area" db:"postcode_area"`
	MaxStops     int         `json:"max_stops" db:"max_stops"`
	CurrentStops int         `json:"current_stops" db:"current_stops"`
	IntervalDays int         `json:"interval_days" db:"interval_days"`
	Status       RouteStatus `json:"status" db:"status"`
	Currency     string      `json:"currency" db:"currency"`
	PricePerStop *float64    `json:"price_per_stop,omitempty" db:"price_per_stop"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
}

// RouteStop is a single customer location visited on a route.
type RouteStop struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	RouteID         uuid.UUID       `json:"route_id" db:"route_id"`
	CustomerID      uuid.UUID       `json:"customer_id" db:"customer_id"`
	PropertyAddress string          `json:"property_address" db:"address"`
	Latitude        float64         `json:"latitude" db:"latitude"`
	Longitude       float64         `json:"longitude" db:"longitude"`
	StopOrder       int             `json:"stop_order" db:"stop_order"`
	TreeCount       *int            `json:"tree_count,omitempty" db:"tree_count"`
	AreaSize        *float64        `json:"area_size,omitempty" db:"area_size"`
	LastVisited     *time.Time      `json:"last_visited,omitempty" db:"last_visited"`
	NextVisit       *time.Time      `json:"next_visit,omitempty" db:"next_visit"`
	Notes           string          `json:"notes,omitempty" db:"notes"`
	Status          RouteStopStatus `json:"status" db:"status"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

// RouteRequest is a customer's request to be added to a provider's route.
type RouteRequest struct {
	ID         uuid.UUID          `json:"id" db:"id"`
	RouteID    uuid.UUID          `json:"route_id" db:"route_id"`
	CustomerID uuid.UUID          `json:"customer_id" db:"customer_id"`
	Status     RouteRequestStatus `json:"status" db:"status"`
	CreatedAt  time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" db:"updated_at"`
}

// GapSuggestion represents a potential new customer stop that fits within an
// existing route's geography.
type GapSuggestion struct {
	Postcode  string  `json:"postcode"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Distance  float64 `json:"distance_km"`
	Reason    string  `json:"reason"`
}

// RouteRepository defines persistence operations for routes and route stops.
type RouteRepository interface {
	CreateRoute(ctx context.Context, route *Route) error
	GetRouteByID(ctx context.Context, id uuid.UUID) (*Route, error)
	ListRoutesByProvider(ctx context.Context, providerID uuid.UUID) ([]Route, error)
	UpdateRoute(ctx context.Context, route *Route) error
	DeleteRoute(ctx context.Context, id uuid.UUID) error

	AddStop(ctx context.Context, stop *RouteStop) error
	GetStopByID(ctx context.Context, id uuid.UUID) (*RouteStop, error)
	ListStopsByRoute(ctx context.Context, routeID uuid.UUID) ([]RouteStop, error)
	UpdateStop(ctx context.Context, stop *RouteStop) error
	RemoveStop(ctx context.Context, id uuid.UUID) error
	ListUpcomingStops(ctx context.Context, providerID uuid.UUID, before time.Time) ([]RouteStop, error)

	CreateRequest(ctx context.Context, req *RouteRequest) error
	GetRequestByID(ctx context.Context, id uuid.UUID) (*RouteRequest, error)
	ListRequestsByRoute(ctx context.Context, routeID uuid.UUID) ([]RouteRequest, error)
	UpdateRequestStatus(ctx context.Context, id uuid.UUID, status RouteRequestStatus) error
}
