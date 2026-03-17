package geo

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const earthRadiusKM = 6371.0

// nominatimBaseURL is the base URL for the OpenStreetMap Nominatim API.
const nominatimBaseURL = "https://nominatim.openstreetmap.org"

// userAgent is required by the Nominatim usage policy.
const userAgent = "SevaApp/1.0 (seva-platform geocoding)"

// rateLimitInterval enforces the Nominatim rate limit of 1 request per second.
var rateLimitInterval = time.Second

// httpClient is a shared HTTP client with a reasonable timeout.
var httpClient = &http.Client{Timeout: 10 * time.Second}

// rateLimiter ensures at most one Nominatim request per second globally.
var (
	rateMu   sync.Mutex
	lastCall time.Time
)

// geocodeCache is an in-memory cache for forward geocoding results.
var geocodeCache sync.Map // key: "postcode:country" -> value: geocodeCacheEntry

// reverseCache is an in-memory cache for reverse geocoding results.
var reverseCache sync.Map // key: "lat:lng" -> value: reverseCacheEntry

type geocodeCacheEntry struct {
	Lat float64
	Lng float64
}

type reverseCacheEntry struct {
	Postcode string
	City     string
	Country  string
}

// nominatimSearchResult represents a single result from the Nominatim search API.
type nominatimSearchResult struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

// nominatimReverseResult represents the result from the Nominatim reverse API.
type nominatimReverseResult struct {
	Address struct {
		Postcode    string `json:"postcode"`
		City        string `json:"city"`
		Town        string `json:"town"`
		Village     string `json:"village"`
		County      string `json:"county"`
		State       string `json:"state"`
		Country     string `json:"country"`
		CountryCode string `json:"country_code"`
	} `json:"address"`
}

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

// PostcodeToCoords converts a postcode and country code (ISO 3166-1 alpha-2,
// e.g. "IN", "GB", "US") to latitude and longitude using the OpenStreetMap
// Nominatim API. Results are cached in memory to reduce API calls.
func PostcodeToCoords(postcode, countryCode string) (lat, lng float64, err error) {
	if postcode == "" {
		return 0, 0, fmt.Errorf("postcode is required")
	}
	if countryCode == "" {
		return 0, 0, fmt.Errorf("country code is required")
	}

	// Check the cache first.
	cacheKey := postcode + ":" + countryCode
	if cached, ok := geocodeCache.Load(cacheKey); ok {
		entry := cached.(geocodeCacheEntry)
		return entry.Lat, entry.Lng, nil
	}

	// Respect Nominatim rate limit.
	throttle()

	// Build the request URL.
	params := url.Values{}
	params.Set("postalcode", postcode)
	params.Set("country", countryCode)
	params.Set("format", "json")
	params.Set("limit", "1")

	reqURL := fmt.Sprintf("%s/search?%s", nominatimBaseURL, params.Encode())

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("geocode: failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("geocode: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("geocode: nominatim returned status %d", resp.StatusCode)
	}

	var results []nominatimSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return 0, 0, fmt.Errorf("geocode: failed to decode response: %w", err)
	}

	if len(results) == 0 {
		return 0, 0, fmt.Errorf("geocode: no results found for postcode %q in country %q", postcode, countryCode)
	}

	lat, err = strconv.ParseFloat(results[0].Lat, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("geocode: failed to parse latitude %q: %w", results[0].Lat, err)
	}

	lng, err = strconv.ParseFloat(results[0].Lon, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("geocode: failed to parse longitude %q: %w", results[0].Lon, err)
	}

	// Cache the result.
	geocodeCache.Store(cacheKey, geocodeCacheEntry{Lat: lat, Lng: lng})

	return lat, lng, nil
}

// ReverseGeocode converts latitude and longitude to address components using
// the OpenStreetMap Nominatim reverse geocoding API. Results are cached in
// memory to reduce API calls.
func ReverseGeocode(lat, lng float64) (postcode, city, country string, err error) {
	// Check the cache first. We round to 4 decimal places (~11m precision)
	// for the cache key to group nearby lookups.
	cacheKey := fmt.Sprintf("%.4f:%.4f", lat, lng)
	if cached, ok := reverseCache.Load(cacheKey); ok {
		entry := cached.(reverseCacheEntry)
		return entry.Postcode, entry.City, entry.Country, nil
	}

	// Respect Nominatim rate limit.
	throttle()

	// Build the request URL.
	params := url.Values{}
	params.Set("lat", strconv.FormatFloat(lat, 'f', 6, 64))
	params.Set("lon", strconv.FormatFloat(lng, 'f', 6, 64))
	params.Set("format", "json")
	params.Set("addressdetails", "1")

	reqURL := fmt.Sprintf("%s/reverse?%s", nominatimBaseURL, params.Encode())

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("reverse geocode: failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("reverse geocode: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("reverse geocode: nominatim returned status %d", resp.StatusCode)
	}

	var result nominatimReverseResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", "", fmt.Errorf("reverse geocode: failed to decode response: %w", err)
	}

	postcode = result.Address.Postcode
	country = result.Address.Country

	// Nominatim returns city, town, or village depending on the location size.
	// Prefer city > town > village > county as the locality name.
	city = result.Address.City
	if city == "" {
		city = result.Address.Town
	}
	if city == "" {
		city = result.Address.Village
	}
	if city == "" {
		city = result.Address.County
	}

	// Cache the result.
	reverseCache.Store(cacheKey, reverseCacheEntry{
		Postcode: postcode,
		City:     city,
		Country:  country,
	})

	return postcode, city, country, nil
}

// throttle enforces the Nominatim rate limit of 1 request per second.
func throttle() {
	rateMu.Lock()
	defer rateMu.Unlock()

	elapsed := time.Since(lastCall)
	if elapsed < rateLimitInterval {
		time.Sleep(rateLimitInterval - elapsed)
	}
	lastCall = time.Now()
}

func degreesToRadians(deg float64) float64 {
	return deg * math.Pi / 180
}

func radiansToDegrees(rad float64) float64 {
	return rad * 180 / math.Pi
}
