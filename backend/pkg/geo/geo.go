package geo

import "math"

const earthRadiusKM = 6371.0

// DistanceKM calculates the great-circle distance in kilometres between two
// points on Earth using the Haversine formula.
func DistanceKM(lat1, lng1, lat2, lng2 float64) float64 {
	dLat := degreesToRadians(lat2 - lat1)
	dLng := degreesToRadians(lng2 - lng1)

	rLat1 := degreesToRadians(lat1)
	rLat2 := degreesToRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rLat1)*math.Cos(rLat2)*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKM * c
}

// BoundingBox returns the south-west and north-east coordinates of a bounding
// box centred on (lat, lng) with the given radius in kilometres.
// This is useful for fast spatial pre-filtering in SQL queries before applying
// a precise Haversine or PostGIS distance check.
func BoundingBox(lat, lng, radiusKM float64) (swLat, swLng, neLat, neLng float64) {
	// Angular distance in radians on a great circle
	angular := radiusKM / earthRadiusKM

	rLat := degreesToRadians(lat)
	rLng := degreesToRadians(lng)

	minLat := rLat - angular
	maxLat := rLat + angular

	// Longitude boundaries need to account for the latitude
	deltaLng := math.Asin(math.Sin(angular) / math.Cos(rLat))

	minLng := rLng - deltaLng
	maxLng := rLng + deltaLng

	swLat = radiansToDegrees(minLat)
	swLng = radiansToDegrees(minLng)
	neLat = radiansToDegrees(maxLat)
	neLng = radiansToDegrees(maxLng)

	return swLat, swLng, neLat, neLng
}

// PostcodeToCoords converts a postcode string to latitude and longitude.
// This is a stub that should be backed by a geocoding service or a local
// lookup table.
func PostcodeToCoords(postcode string) (lat, lng float64, err error) {
	// TODO: integrate with a geocoding API (Google Maps, Mapbox, etc.)
	// or use a local postcode-to-coordinate lookup table
	_ = postcode
	return 0, 0, nil
}

func degreesToRadians(deg float64) float64 {
	return deg * math.Pi / 180
}

func radiansToDegrees(rad float64) float64 {
	return rad * 180 / math.Pi
}
