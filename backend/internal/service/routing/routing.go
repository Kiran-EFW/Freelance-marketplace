package routing

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/domain"
	"github.com/seva-platform/backend/pkg/geo"
)

// Service defines the routing service interface.
type Service interface {
	CreateRoute(ctx context.Context, providerID uuid.UUID, name, postcode string, intervalDays int) (*domain.Route, error)
	AddStop(ctx context.Context, routeID, customerID uuid.UUID, address string, lat, lng float64, treeCount *int, areaSize *float64) (*domain.RouteStop, error)
	RemoveStop(ctx context.Context, routeID, stopID uuid.UUID) error
	OptimizeRoute(ctx context.Context, routeID uuid.UUID) ([]domain.RouteStop, error)
	GetNextVisits(ctx context.Context, providerID uuid.UUID) ([]domain.RouteStop, error)
	DetectRouteGaps(ctx context.Context, routeID uuid.UUID) ([]domain.GapSuggestion, error)
	GenerateWeeklySMS(ctx context.Context, providerID uuid.UUID) (string, error)
	AdjustSeasonalInterval(ctx context.Context, routeID uuid.UUID, season string) error
}

// RoutingService implements route management for circuit-based service workers.
type RoutingService struct {
	routes    domain.RouteRepository
	providers domain.ProviderRepository
}

// NewRoutingService returns a ready-to-use RoutingService.
func NewRoutingService(routes domain.RouteRepository, providers domain.ProviderRepository) *RoutingService {
	return &RoutingService{
		routes:    routes,
		providers: providers,
	}
}

// CreateRoute sets up a new service route for a provider.
func (s *RoutingService) CreateRoute(ctx context.Context, providerID uuid.UUID, name, postcode string, intervalDays int) (*domain.Route, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: route name is required", domain.ErrInvalidInput)
	}
	if intervalDays <= 0 {
		return nil, fmt.Errorf("%w: interval must be positive", domain.ErrInvalidInput)
	}

	route := &domain.Route{
		ID:           uuid.New(),
		ProviderID:   providerID,
		Name:         name,
		PostcodeArea: postcode,
		MaxStops:     20,
		CurrentStops: 0,
		IntervalDays: intervalDays,
		Status:       domain.RouteStatusActive,
		Currency:     "INR",
	}

	if err := s.routes.CreateRoute(ctx, route); err != nil {
		log.Error().Err(err).Str("provider_id", providerID.String()).Msg("failed to create route")
		return nil, fmt.Errorf("create route: %w", err)
	}

	log.Info().
		Str("route_id", route.ID.String()).
		Str("provider_id", providerID.String()).
		Str("name", name).
		Int("interval_days", intervalDays).
		Msg("route created")

	return route, nil
}

// AddStop adds a new customer stop to an existing route.
func (s *RoutingService) AddStop(ctx context.Context, routeID, customerID uuid.UUID, address string, lat, lng float64, treeCount *int, areaSize *float64) (*domain.RouteStop, error) {
	route, err := s.routes.GetRouteByID(ctx, routeID)
	if err != nil {
		return nil, fmt.Errorf("%w: route %s", domain.ErrNotFound, routeID)
	}

	if route.CurrentStops >= route.MaxStops {
		return nil, fmt.Errorf("%w: route has reached maximum stops (%d)", domain.ErrInvalidState, route.MaxStops)
	}

	now := time.Now()
	nextVisit := now.AddDate(0, 0, route.IntervalDays)

	stop := &domain.RouteStop{
		ID:              uuid.New(),
		RouteID:         routeID,
		CustomerID:      customerID,
		PropertyAddress: address,
		Latitude:        lat,
		Longitude:       lng,
		StopOrder:       route.CurrentStops + 1,
		TreeCount:       treeCount,
		AreaSize:        areaSize,
		NextVisit:       &nextVisit,
		Status:          domain.RouteStopStatusActive,
	}

	if err := s.routes.AddStop(ctx, stop); err != nil {
		log.Error().Err(err).Str("route_id", routeID.String()).Msg("failed to add stop")
		return nil, fmt.Errorf("add stop: %w", err)
	}

	// Update the route's current stop count.
	route.CurrentStops++
	if err := s.routes.UpdateRoute(ctx, route); err != nil {
		log.Warn().Err(err).Msg("failed to update route stop count")
	}

	log.Info().
		Str("stop_id", stop.ID.String()).
		Str("route_id", routeID.String()).
		Str("customer_id", customerID.String()).
		Msg("stop added to route")

	return stop, nil
}

// RemoveStop removes a stop from a route and decrements the count.
func (s *RoutingService) RemoveStop(ctx context.Context, routeID, stopID uuid.UUID) error {
	route, err := s.routes.GetRouteByID(ctx, routeID)
	if err != nil {
		return fmt.Errorf("%w: route %s", domain.ErrNotFound, routeID)
	}

	if err := s.routes.RemoveStop(ctx, stopID); err != nil {
		return fmt.Errorf("remove stop: %w", err)
	}

	if route.CurrentStops > 0 {
		route.CurrentStops--
		if err := s.routes.UpdateRoute(ctx, route); err != nil {
			log.Warn().Err(err).Msg("failed to update route stop count after removal")
		}
	}

	log.Info().
		Str("stop_id", stopID.String()).
		Str("route_id", routeID.String()).
		Msg("stop removed from route")

	return nil
}

// OptimizeRoute reorders the stops in a route by proximity using a nearest-
// neighbour heuristic, minimising the total travel distance.
func (s *RoutingService) OptimizeRoute(ctx context.Context, routeID uuid.UUID) ([]domain.RouteStop, error) {
	stops, err := s.routes.ListStopsByRoute(ctx, routeID)
	if err != nil {
		return nil, fmt.Errorf("list stops: %w", err)
	}

	if len(stops) <= 1 {
		return stops, nil
	}

	// Nearest-neighbour algorithm.
	optimized := make([]domain.RouteStop, 0, len(stops))
	visited := make(map[int]bool)

	// Start with the first stop.
	current := 0
	visited[current] = true
	optimized = append(optimized, stops[current])

	for len(optimized) < len(stops) {
		bestIdx := -1
		bestDist := float64(999999)

		for i := range stops {
			if visited[i] {
				continue
			}
			dist := geo.DistanceKM(
				stops[current].Latitude, stops[current].Longitude,
				stops[i].Latitude, stops[i].Longitude,
			)
			if dist < bestDist {
				bestDist = dist
				bestIdx = i
			}
		}

		if bestIdx == -1 {
			break
		}

		visited[bestIdx] = true
		current = bestIdx
		optimized = append(optimized, stops[current])
	}

	// Update stop order in the database.
	for i := range optimized {
		optimized[i].StopOrder = i + 1
		if err := s.routes.UpdateStop(ctx, &optimized[i]); err != nil {
			log.Warn().Err(err).
				Str("stop_id", optimized[i].ID.String()).
				Int("order", i+1).
				Msg("failed to update stop order")
		}
	}

	log.Info().
		Str("route_id", routeID.String()).
		Int("stops", len(optimized)).
		Msg("route optimized")

	return optimized, nil
}

// GetNextVisits returns the stops a provider should visit this week.
func (s *RoutingService) GetNextVisits(ctx context.Context, providerID uuid.UUID) ([]domain.RouteStop, error) {
	endOfWeek := time.Now().AddDate(0, 0, 7)
	stops, err := s.routes.ListUpcomingStops(ctx, providerID, endOfWeek)
	if err != nil {
		return nil, fmt.Errorf("list upcoming stops: %w", err)
	}

	// Sort by next visit time.
	sort.Slice(stops, func(i, j int) bool {
		if stops[i].NextVisit == nil {
			return false
		}
		if stops[j].NextVisit == nil {
			return true
		}
		return stops[i].NextVisit.Before(*stops[j].NextVisit)
	})

	return stops, nil
}

// DetectRouteGaps identifies areas along a route where new customers could be
// added to fill geographical gaps.
func (s *RoutingService) DetectRouteGaps(ctx context.Context, routeID uuid.UUID) ([]domain.GapSuggestion, error) {
	stops, err := s.routes.ListStopsByRoute(ctx, routeID)
	if err != nil {
		return nil, fmt.Errorf("list stops: %w", err)
	}

	if len(stops) < 2 {
		return nil, nil
	}

	// Sort stops by their current order.
	sort.Slice(stops, func(i, j int) bool {
		return stops[i].StopOrder < stops[j].StopOrder
	})

	var suggestions []domain.GapSuggestion
	const gapThresholdKM = 2.0

	for i := 0; i < len(stops)-1; i++ {
		dist := geo.DistanceKM(
			stops[i].Latitude, stops[i].Longitude,
			stops[i+1].Latitude, stops[i+1].Longitude,
		)

		if dist > gapThresholdKM {
			// Midpoint between the two stops.
			midLat := (stops[i].Latitude + stops[i+1].Latitude) / 2
			midLng := (stops[i].Longitude + stops[i+1].Longitude) / 2

			suggestions = append(suggestions, domain.GapSuggestion{
				Latitude:  midLat,
				Longitude: midLng,
				Distance:  dist,
				Reason:    fmt.Sprintf("%.1f km gap between stop %d and %d", dist, stops[i].StopOrder, stops[i+1].StopOrder),
			})
		}
	}

	log.Info().
		Str("route_id", routeID.String()).
		Int("gaps_found", len(suggestions)).
		Msg("route gaps detected")

	return suggestions, nil
}

// GenerateWeeklySMS formats a provider's upcoming weekly schedule as a plain-
// text SMS message suitable for feature phones.
func (s *RoutingService) GenerateWeeklySMS(ctx context.Context, providerID uuid.UUID) (string, error) {
	stops, err := s.GetNextVisits(ctx, providerID)
	if err != nil {
		return "", err
	}

	if len(stops) == 0 {
		return "Seva: No visits scheduled this week.", nil
	}

	var b strings.Builder
	b.WriteString("Seva Weekly Schedule:\n")

	for i, stop := range stops {
		dayStr := "TBD"
		if stop.NextVisit != nil {
			dayStr = stop.NextVisit.Format("Mon 02 Jan")
		}
		b.WriteString(fmt.Sprintf("%d. %s - %s\n", i+1, dayStr, stop.PropertyAddress))
	}

	b.WriteString(fmt.Sprintf("Total: %d stops", len(stops)))

	return b.String(), nil
}

// AdjustSeasonalInterval changes the visit interval for a route based on the
// current season. For example, coconut tree routes may need more frequent
// visits during the monsoon.
func (s *RoutingService) AdjustSeasonalInterval(ctx context.Context, routeID uuid.UUID, season string) error {
	route, err := s.routes.GetRouteByID(ctx, routeID)
	if err != nil {
		return fmt.Errorf("%w: route %s", domain.ErrNotFound, routeID)
	}

	switch season {
	case "monsoon":
		route.IntervalDays = 21 // less frequent during monsoon
	case "summer":
		route.IntervalDays = 30 // monthly in summer
	case "winter":
		route.IntervalDays = 45 // coconut harvesting slows
	case "harvest":
		route.IntervalDays = 14 // more frequent during harvest
	default:
		return fmt.Errorf("%w: unknown season %s", domain.ErrInvalidInput, season)
	}

	if err := s.routes.UpdateRoute(ctx, route); err != nil {
		return fmt.Errorf("update route interval: %w", err)
	}

	log.Info().
		Str("route_id", routeID.String()).
		Str("season", season).
		Int("new_interval", route.IntervalDays).
		Msg("seasonal interval adjusted")

	return nil
}
