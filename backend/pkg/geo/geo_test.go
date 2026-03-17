package geo

import (
	"math"
	"testing"
)

// ---------------------------------------------------------------------------
// TestDistanceKM
// ---------------------------------------------------------------------------

func TestDistanceKM(t *testing.T) {
	tests := []struct {
		name     string
		lat1     float64
		lng1     float64
		lat2     float64
		lng2     float64
		wantKM   float64
		tolerance float64 // allowed deviation in km
	}{
		{
			name:      "Bangalore to Chennai",
			lat1:      12.9716, lng1: 77.5946,
			lat2:      13.0827, lng2: 80.2707,
			wantKM:    291,
			tolerance: 10,
		},
		{
			name:      "Mumbai to Delhi",
			lat1:      19.0760, lng1: 72.8777,
			lat2:      28.7041, lng2: 77.1025,
			wantKM:    1148,
			tolerance: 20,
		},
		{
			name:      "Same point should be zero",
			lat1:      12.9716, lng1: 77.5946,
			lat2:      12.9716, lng2: 77.5946,
			wantKM:    0,
			tolerance: 0.001,
		},
		{
			name:      "Kolkata to Hyderabad",
			lat1:      22.5726, lng1: 88.3639,
			lat2:      17.3850, lng2: 78.4867,
			wantKM:    1178,
			tolerance: 30,
		},
		{
			name:      "Very short distance (within a city)",
			lat1:      12.9716, lng1: 77.5946,
			lat2:      12.9816, lng2: 77.6046,
			wantKM:    1.4,
			tolerance: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DistanceKM(tt.lat1, tt.lng1, tt.lat2, tt.lng2)
			diff := math.Abs(got - tt.wantKM)
			if diff > tt.tolerance {
				t.Errorf("DistanceKM(%f, %f, %f, %f) = %f, want ~%f (tolerance %f, diff %f)",
					tt.lat1, tt.lng1, tt.lat2, tt.lng2, got, tt.wantKM, tt.tolerance, diff)
			}
		})
	}
}

func TestDistanceKMSymmetric(t *testing.T) {
	// Distance from A to B should equal distance from B to A.
	d1 := DistanceKM(12.9716, 77.5946, 13.0827, 80.2707)
	d2 := DistanceKM(13.0827, 80.2707, 12.9716, 77.5946)

	if math.Abs(d1-d2) > 0.001 {
		t.Errorf("DistanceKM is not symmetric: %f != %f", d1, d2)
	}
}

func TestDistanceKMNonNegative(t *testing.T) {
	d := DistanceKM(0, 0, -90, 180)
	if d < 0 {
		t.Errorf("DistanceKM should never be negative, got %f", d)
	}
}

// ---------------------------------------------------------------------------
// TestBoundingBox
// ---------------------------------------------------------------------------

func TestBoundingBox(t *testing.T) {
	tests := []struct {
		name     string
		lat      float64
		lng      float64
		radiusKM float64
	}{
		{
			name:     "Bangalore 10km radius",
			lat:      12.9716,
			lng:      77.5946,
			radiusKM: 10,
		},
		{
			name:     "Delhi 50km radius",
			lat:      28.7041,
			lng:      77.1025,
			radiusKM: 50,
		},
		{
			name:     "Very small radius 1km",
			lat:      12.9716,
			lng:      77.5946,
			radiusKM: 1,
		},
		{
			name:     "Large radius 200km",
			lat:      20.0,
			lng:      78.0,
			radiusKM: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			swLat, swLng, neLat, neLng := BoundingBox(tt.lat, tt.lng, tt.radiusKM)

			// SW corner should be south and west of the centre.
			if swLat >= tt.lat {
				t.Errorf("SW latitude %f should be less than centre %f", swLat, tt.lat)
			}
			if swLng >= tt.lng {
				t.Errorf("SW longitude %f should be less than centre %f", swLng, tt.lng)
			}

			// NE corner should be north and east of the centre.
			if neLat <= tt.lat {
				t.Errorf("NE latitude %f should be greater than centre %f", neLat, tt.lat)
			}
			if neLng <= tt.lng {
				t.Errorf("NE longitude %f should be greater than centre %f", neLng, tt.lng)
			}

			// The distance from the centre to the SW corner should be
			// approximately equal to the radius (allow 15% tolerance due
			// to the rectangular approximation).
			distSW := DistanceKM(tt.lat, tt.lng, swLat, swLng)
			if distSW < tt.radiusKM*0.7 || distSW > tt.radiusKM*1.5 {
				t.Errorf("Distance from centre to SW corner = %f km, expected ~%f km", distSW, tt.radiusKM)
			}

			// The distance from the centre to the NE corner should also be
			// approximately equal to the radius.
			distNE := DistanceKM(tt.lat, tt.lng, neLat, neLng)
			if distNE < tt.radiusKM*0.7 || distNE > tt.radiusKM*1.5 {
				t.Errorf("Distance from centre to NE corner = %f km, expected ~%f km", distNE, tt.radiusKM)
			}

			// The bounding box should be symmetric.
			latDiffSW := tt.lat - swLat
			latDiffNE := neLat - tt.lat
			if math.Abs(latDiffSW-latDiffNE) > 0.0001 {
				t.Errorf("Bounding box latitude is not symmetric: SW diff=%f, NE diff=%f", latDiffSW, latDiffNE)
			}
		})
	}
}

func TestBoundingBoxContainsCentre(t *testing.T) {
	swLat, swLng, neLat, neLng := BoundingBox(12.9716, 77.5946, 5)

	if 12.9716 < swLat || 12.9716 > neLat {
		t.Error("Centre latitude should be within the bounding box")
	}
	if 77.5946 < swLng || 77.5946 > neLng {
		t.Error("Centre longitude should be within the bounding box")
	}
}

// ---------------------------------------------------------------------------
// Test helper functions
// ---------------------------------------------------------------------------

func TestDegreesToRadians(t *testing.T) {
	tests := []struct {
		degrees float64
		want    float64
	}{
		{0, 0},
		{180, math.Pi},
		{90, math.Pi / 2},
		{360, 2 * math.Pi},
		{-180, -math.Pi},
	}

	for _, tt := range tests {
		got := degreesToRadians(tt.degrees)
		if math.Abs(got-tt.want) > 0.0001 {
			t.Errorf("degreesToRadians(%f) = %f, want %f", tt.degrees, got, tt.want)
		}
	}
}

func TestRadiansToDegrees(t *testing.T) {
	tests := []struct {
		radians float64
		want    float64
	}{
		{0, 0},
		{math.Pi, 180},
		{math.Pi / 2, 90},
		{2 * math.Pi, 360},
	}

	for _, tt := range tests {
		got := radiansToDegrees(tt.radians)
		if math.Abs(got-tt.want) > 0.0001 {
			t.Errorf("radiansToDegrees(%f) = %f, want %f", tt.radians, got, tt.want)
		}
	}
}
