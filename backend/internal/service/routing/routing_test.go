package routing

import (
	"math"
	"testing"

	"github.com/google/uuid"
	"github.com/seva-platform/backend/internal/domain"
)

// ---------------------------------------------------------------------------
// TestCalculateTotalDistance
// ---------------------------------------------------------------------------

func TestCalculateTotalDistance(t *testing.T) {
	tests := []struct {
		name  string
		stops []domain.RouteStop
		want  float64 // expected approximate total distance in km
		tol   float64 // tolerance
	}{
		{
			name:  "empty stops",
			stops: nil,
			want:  0,
			tol:   0.001,
		},
		{
			name: "single stop",
			stops: []domain.RouteStop{
				{Latitude: 12.9716, Longitude: 77.5946},
			},
			want: 0,
			tol:  0.001,
		},
		{
			name: "two stops in Bangalore",
			stops: []domain.RouteStop{
				{Latitude: 12.9716, Longitude: 77.5946}, // MG Road
				{Latitude: 12.9352, Longitude: 77.6245}, // Koramangala
			},
			want: 5, // roughly 5 km
			tol:  2,
		},
		{
			name: "three stops forming a path",
			stops: []domain.RouteStop{
				{Latitude: 12.9716, Longitude: 77.5946}, // MG Road
				{Latitude: 12.9352, Longitude: 77.6245}, // Koramangala
				{Latitude: 12.9698, Longitude: 77.7500}, // Whitefield
			},
			want: 18, // roughly MG Road -> Koramangala -> Whitefield
			tol:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateTotalDistance(tt.stops)
			diff := math.Abs(got - tt.want)
			if diff > tt.tol {
				t.Errorf("CalculateTotalDistance() = %f, want ~%f (diff %f)", got, tt.want, diff)
			}
		})
	}
}

func TestCalculateTotalDistanceNonNegative(t *testing.T) {
	stops := []domain.RouteStop{
		{Latitude: 0, Longitude: 0},
		{Latitude: -33.8688, Longitude: 151.2093}, // Sydney
		{Latitude: 40.7128, Longitude: -74.0060},   // New York
	}
	d := CalculateTotalDistance(stops)
	if d < 0 {
		t.Errorf("CalculateTotalDistance should never be negative, got %f", d)
	}
}

// ---------------------------------------------------------------------------
// TestNearestNeighbor (via twoOptImprove and ordering)
// ---------------------------------------------------------------------------

// We test the nearest-neighbour heuristic indirectly by verifying that
// twoOptImprove does not increase the total distance.

func TestTwoOptImprove(t *testing.T) {
	// Create a set of stops that form a known sub-optimal order.
	// A "zigzag" ordering should be improved by 2-opt.
	stops := []domain.RouteStop{
		{ID: uuid.New(), Latitude: 12.97, Longitude: 77.59, StopOrder: 1},  // A (north)
		{ID: uuid.New(), Latitude: 12.90, Longitude: 77.70, StopOrder: 2},  // C (south-east)
		{ID: uuid.New(), Latitude: 12.95, Longitude: 77.60, StopOrder: 3},  // B (middle)
		{ID: uuid.New(), Latitude: 12.88, Longitude: 77.72, StopOrder: 4},  // D (further south-east)
	}

	distBefore := CalculateTotalDistance(stops)

	improved := twoOptImprove(stops)

	distAfter := CalculateTotalDistance(improved)

	// The improved distance should be less than or equal to the original.
	if distAfter > distBefore+0.001 {
		t.Errorf("twoOptImprove increased distance: before=%f, after=%f", distBefore, distAfter)
	}

	// Length should be preserved.
	if len(improved) != 4 {
		t.Errorf("twoOptImprove changed stop count: got %d, want 4", len(improved))
	}
}

func TestTwoOptImproveFewStops(t *testing.T) {
	// With 0, 1, or 2 stops, twoOptImprove should return them unchanged.
	tests := []struct {
		name  string
		stops []domain.RouteStop
	}{
		{"empty", nil},
		{"single", []domain.RouteStop{{Latitude: 12.97, Longitude: 77.59}}},
		{"two", []domain.RouteStop{
			{Latitude: 12.97, Longitude: 77.59},
			{Latitude: 12.90, Longitude: 77.70},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := twoOptImprove(tt.stops)
			if len(result) != len(tt.stops) {
				t.Errorf("twoOptImprove(%s) changed length: got %d, want %d", tt.name, len(result), len(tt.stops))
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestReverseSegment
// ---------------------------------------------------------------------------

func TestReverseSegment(t *testing.T) {
	makeStops := func(orders ...int) []domain.RouteStop {
		stops := make([]domain.RouteStop, len(orders))
		for i, o := range orders {
			stops[i] = domain.RouteStop{StopOrder: o}
		}
		return stops
	}

	stops := makeStops(1, 2, 3, 4, 5)
	reverseSegment(stops, 1, 3) // reverse indices 1..3

	expected := []int{1, 4, 3, 2, 5}
	for i, want := range expected {
		if stops[i].StopOrder != want {
			t.Errorf("after reverseSegment, stops[%d].StopOrder = %d, want %d", i, stops[i].StopOrder, want)
		}
	}
}

func TestReverseSegmentFullReverse(t *testing.T) {
	makeStops := func(orders ...int) []domain.RouteStop {
		stops := make([]domain.RouteStop, len(orders))
		for i, o := range orders {
			stops[i] = domain.RouteStop{StopOrder: o}
		}
		return stops
	}

	stops := makeStops(1, 2, 3, 4)
	reverseSegment(stops, 0, 3) // reverse entire slice

	expected := []int{4, 3, 2, 1}
	for i, want := range expected {
		if stops[i].StopOrder != want {
			t.Errorf("stops[%d].StopOrder = %d, want %d", i, stops[i].StopOrder, want)
		}
	}
}

func TestReverseSegmentSingleElement(t *testing.T) {
	stops := []domain.RouteStop{{StopOrder: 1}}
	reverseSegment(stops, 0, 0) // no-op
	if stops[0].StopOrder != 1 {
		t.Error("reverseSegment on single element should be no-op")
	}
}
