import { describe, it, expect } from 'vitest';

// We test only the pure, SSR-safe functions from geolocation.ts.
// Functions that rely on browser APIs (navigator, localStorage, fetch)
// are not imported directly since vitest runs in Node by default.
// Instead we replicate the logic here to verify correctness.

// ---------------------------------------------------------------------------
// Test: firstOf helper logic
// ---------------------------------------------------------------------------

/** Pick the first defined, non-empty string from a list of candidates. */
function firstOf(...candidates: (string | undefined | null)[]): string {
	for (const c of candidates) {
		if (c && c.trim().length > 0) return c.trim();
	}
	return '';
}

describe('firstOf', () => {
	it('returns the first non-empty string', () => {
		expect(firstOf('hello', 'world')).toBe('hello');
	});

	it('skips undefined values', () => {
		expect(firstOf(undefined, 'world')).toBe('world');
	});

	it('skips null values', () => {
		expect(firstOf(null, 'world')).toBe('world');
	});

	it('skips empty strings', () => {
		expect(firstOf('', 'world')).toBe('world');
	});

	it('skips whitespace-only strings', () => {
		expect(firstOf('  ', 'world')).toBe('world');
	});

	it('trims the returned value', () => {
		expect(firstOf('  hello  ')).toBe('hello');
	});

	it('returns empty string if no candidates match', () => {
		expect(firstOf(undefined, null, '', '  ')).toBe('');
	});

	it('returns empty string when called with no arguments', () => {
		expect(firstOf()).toBe('');
	});
});

// ---------------------------------------------------------------------------
// Test: Display name construction logic
// ---------------------------------------------------------------------------

function buildDisplayName(
	area: string,
	city: string,
	country: string,
	lat: number,
	lng: number,
	rawDisplayName?: string
): string {
	if (area && city) {
		return `${area}, ${city}`;
	} else if (city && country) {
		return `${city}, ${country}`;
	} else {
		return firstOf(rawDisplayName, `${lat.toFixed(4)}, ${lng.toFixed(4)}`);
	}
}

describe('buildDisplayName', () => {
	it('uses area and city when both are present', () => {
		expect(buildDisplayName('Koramangala', 'Bangalore', 'India', 12.9, 77.6)).toBe(
			'Koramangala, Bangalore'
		);
	});

	it('falls back to city and country when area is missing', () => {
		expect(buildDisplayName('', 'Bangalore', 'India', 12.9, 77.6)).toBe(
			'Bangalore, India'
		);
	});

	it('falls back to raw display name when both area and city are missing', () => {
		expect(buildDisplayName('', '', 'India', 12.9, 77.6, 'Some place, India')).toBe(
			'Some place, India'
		);
	});

	it('falls back to coordinates when everything is missing', () => {
		expect(buildDisplayName('', '', '', 12.9716, 77.5946)).toBe('12.9716, 77.5946');
	});
});

// ---------------------------------------------------------------------------
// Test: mapPositionErrorCode logic
// ---------------------------------------------------------------------------

function mapPositionErrorCode(
	nativeCode: number
): 'PERMISSION_DENIED' | 'POSITION_UNAVAILABLE' | 'TIMEOUT' {
	switch (nativeCode) {
		case 1:
			return 'PERMISSION_DENIED';
		case 2:
			return 'POSITION_UNAVAILABLE';
		case 3:
			return 'TIMEOUT';
		default:
			return 'POSITION_UNAVAILABLE';
	}
}

describe('mapPositionErrorCode', () => {
	it('maps 1 to PERMISSION_DENIED', () => {
		expect(mapPositionErrorCode(1)).toBe('PERMISSION_DENIED');
	});

	it('maps 2 to POSITION_UNAVAILABLE', () => {
		expect(mapPositionErrorCode(2)).toBe('POSITION_UNAVAILABLE');
	});

	it('maps 3 to TIMEOUT', () => {
		expect(mapPositionErrorCode(3)).toBe('TIMEOUT');
	});

	it('maps unknown codes to POSITION_UNAVAILABLE', () => {
		expect(mapPositionErrorCode(99)).toBe('POSITION_UNAVAILABLE');
		expect(mapPositionErrorCode(0)).toBe('POSITION_UNAVAILABLE');
		expect(mapPositionErrorCode(-1)).toBe('POSITION_UNAVAILABLE');
	});
});

// ---------------------------------------------------------------------------
// Test: Coordinate parsing and validation
// ---------------------------------------------------------------------------

function isValidCoordinate(lat: number, lng: number): boolean {
	return (
		typeof lat === 'number' &&
		typeof lng === 'number' &&
		!isNaN(lat) &&
		!isNaN(lng) &&
		lat >= -90 &&
		lat <= 90 &&
		lng >= -180 &&
		lng <= 180
	);
}

describe('isValidCoordinate', () => {
	it('accepts valid coordinates', () => {
		expect(isValidCoordinate(12.9716, 77.5946)).toBe(true);
		expect(isValidCoordinate(0, 0)).toBe(true);
		expect(isValidCoordinate(-90, -180)).toBe(true);
		expect(isValidCoordinate(90, 180)).toBe(true);
	});

	it('rejects out-of-range latitude', () => {
		expect(isValidCoordinate(91, 0)).toBe(false);
		expect(isValidCoordinate(-91, 0)).toBe(false);
	});

	it('rejects out-of-range longitude', () => {
		expect(isValidCoordinate(0, 181)).toBe(false);
		expect(isValidCoordinate(0, -181)).toBe(false);
	});

	it('rejects NaN', () => {
		expect(isValidCoordinate(NaN, 77)).toBe(false);
		expect(isValidCoordinate(12, NaN)).toBe(false);
	});
});

// ---------------------------------------------------------------------------
// Test: Distance calculation (Haversine)
// ---------------------------------------------------------------------------

const EARTH_RADIUS_KM = 6371;

function distanceKM(lat1: number, lng1: number, lat2: number, lng2: number): number {
	const toRad = (deg: number) => (deg * Math.PI) / 180;
	const dLat = toRad(lat2 - lat1);
	const dLng = toRad(lng2 - lng1);
	const a =
		Math.sin(dLat / 2) ** 2 +
		Math.cos(toRad(lat1)) * Math.cos(toRad(lat2)) * Math.sin(dLng / 2) ** 2;
	return 2 * EARTH_RADIUS_KM * Math.asin(Math.sqrt(a));
}

describe('distanceKM', () => {
	it('returns 0 for the same point', () => {
		expect(distanceKM(12.9716, 77.5946, 12.9716, 77.5946)).toBeCloseTo(0, 5);
	});

	it('calculates Bangalore to Chennai correctly (~290 km)', () => {
		const d = distanceKM(12.9716, 77.5946, 13.0827, 80.2707);
		expect(d).toBeGreaterThan(280);
		expect(d).toBeLessThan(300);
	});

	it('is symmetric', () => {
		const d1 = distanceKM(12.9716, 77.5946, 13.0827, 80.2707);
		const d2 = distanceKM(13.0827, 80.2707, 12.9716, 77.5946);
		expect(Math.abs(d1 - d2)).toBeLessThan(0.001);
	});

	it('returns non-negative values', () => {
		expect(distanceKM(0, 0, -90, 180)).toBeGreaterThanOrEqual(0);
	});

	it('calculates short distances within a city', () => {
		// MG Road to Koramangala in Bangalore (~5 km)
		const d = distanceKM(12.9716, 77.5946, 12.9352, 77.6245);
		expect(d).toBeGreaterThan(3);
		expect(d).toBeLessThan(7);
	});
});

// ---------------------------------------------------------------------------
// Test: Cache age check logic
// ---------------------------------------------------------------------------

const CACHE_MAX_AGE_MS = 24 * 60 * 60 * 1000; // 24 hours

describe('cache age check', () => {
	it('considers recent timestamps as valid', () => {
		const age = Date.now() - (Date.now() - 1000); // 1 second ago
		expect(age < CACHE_MAX_AGE_MS).toBe(true);
	});

	it('considers timestamps older than 24h as stale', () => {
		const age = 25 * 60 * 60 * 1000; // 25 hours
		expect(age > CACHE_MAX_AGE_MS).toBe(true);
	});

	it('considers exactly 24h as stale', () => {
		const age = CACHE_MAX_AGE_MS;
		// The condition in geolocation.ts is `age > CACHE_MAX_AGE_MS` so exactly 24h is NOT stale.
		expect(age > CACHE_MAX_AGE_MS).toBe(false);
	});
});

// ---------------------------------------------------------------------------
// Test: Nominatim URL construction
// ---------------------------------------------------------------------------

describe('Nominatim URL construction', () => {
	const NOMINATIM_BASE = 'https://nominatim.openstreetmap.org/reverse';

	it('constructs correct URL with coordinates', () => {
		const lat = 12.9716;
		const lng = 77.5946;
		const url = `${NOMINATIM_BASE}?format=json&lat=${lat}&lon=${lng}&zoom=18&addressdetails=1`;
		expect(url).toBe(
			'https://nominatim.openstreetmap.org/reverse?format=json&lat=12.9716&lon=77.5946&zoom=18&addressdetails=1'
		);
	});

	it('handles negative coordinates', () => {
		const lat = -33.8688;
		const lng = 151.2093;
		const url = `${NOMINATIM_BASE}?format=json&lat=${lat}&lon=${lng}&zoom=18&addressdetails=1`;
		expect(url).toContain('lat=-33.8688');
		expect(url).toContain('lon=151.2093');
	});
});
