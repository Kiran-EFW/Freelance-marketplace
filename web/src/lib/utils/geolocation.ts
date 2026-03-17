// Browser geolocation utility for auto-detecting user location
// and reverse geocoding to a postcode/address via OpenStreetMap Nominatim.
// Pure TypeScript -- no Svelte dependencies. SSR-safe.

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface DetectedLocation {
	latitude: number;
	longitude: number;
	postcode: string;
	area: string; // neighborhood / locality name
	city: string;
	state: string;
	country: string;
	displayName: string; // formatted like "Koramangala, Bangalore"
}

export interface GeolocationError {
	code:
		| 'PERMISSION_DENIED'
		| 'POSITION_UNAVAILABLE'
		| 'TIMEOUT'
		| 'GEOCODE_FAILED'
		| 'NOT_SUPPORTED';
	message: string;
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const STORAGE_KEY = 'seva-location';
const CACHE_MAX_AGE_MS = 24 * 60 * 60 * 1000; // 24 hours
const NOMINATIM_BASE = 'https://nominatim.openstreetmap.org/reverse';
const USER_AGENT = 'Seva-Platform';

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

interface StoredEntry {
	location: DetectedLocation;
	timestamp: number;
}

/** Return true when running in a browser context. */
function isBrowser(): boolean {
	return typeof window !== 'undefined' && typeof navigator !== 'undefined';
}

/**
 * Create and throw a typed GeolocationError.
 */
function throwGeolocationError(
	code: GeolocationError['code'],
	message: string
): never {
	const error: GeolocationError & Error = Object.assign(new Error(message), {
		code,
		message
	});
	throw error;
}

/**
 * Map the native GeolocationPositionError code to our error code union.
 */
function mapPositionErrorCode(
	nativeCode: number
): GeolocationError['code'] {
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

/**
 * Pick the first defined, non-empty string from a list of candidates.
 */
function firstOf(...candidates: (string | undefined | null)[]): string {
	for (const c of candidates) {
		if (c && c.trim().length > 0) return c.trim();
	}
	return '';
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/**
 * Check whether the browser supports the Geolocation API.
 * Always returns `false` during SSR.
 */
export function isGeolocationSupported(): boolean {
	return isBrowser() && 'geolocation' in navigator;
}

/**
 * Read the cached location from localStorage.
 * Returns `null` if no entry exists, the entry is malformed, or the cache
 * is older than 24 hours.  SSR-safe -- always returns `null` on the server.
 */
export function getStoredLocation(): DetectedLocation | null {
	if (!isBrowser()) return null;

	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		if (!raw) return null;

		const entry: StoredEntry = JSON.parse(raw);

		if (
			!entry ||
			typeof entry.timestamp !== 'number' ||
			!entry.location
		) {
			return null;
		}

		const age = Date.now() - entry.timestamp;
		if (age > CACHE_MAX_AGE_MS) {
			return null;
		}

		return entry.location;
	} catch {
		return null;
	}
}

/**
 * Persist a DetectedLocation to localStorage with the current timestamp.
 * No-op during SSR.
 */
export function saveLocation(location: DetectedLocation): void {
	if (!isBrowser()) return;

	try {
		const entry: StoredEntry = {
			location,
			timestamp: Date.now()
		};
		localStorage.setItem(STORAGE_KEY, JSON.stringify(entry));
	} catch {
		// localStorage may be full or disabled -- silently ignore.
	}
}

/**
 * Remove the cached location from localStorage.
 * No-op during SSR.
 */
export function clearLocation(): void {
	if (!isBrowser()) return;

	try {
		localStorage.removeItem(STORAGE_KEY);
	} catch {
		// Ignore errors from restricted storage contexts.
	}
}

/**
 * Reverse-geocode a latitude/longitude pair via the OpenStreetMap Nominatim
 * API and return a DetectedLocation.
 *
 * Throws a GeolocationError with code `'GEOCODE_FAILED'` if the network
 * request fails or the response cannot be parsed.
 */
export async function reverseGeocode(
	lat: number,
	lng: number
): Promise<DetectedLocation> {
	const url = `${NOMINATIM_BASE}?format=json&lat=${lat}&lon=${lng}&zoom=18&addressdetails=1`;

	// Determine Accept-Language from the browser when available.
	const acceptLanguage =
		isBrowser() && navigator.languages?.length
			? navigator.languages.join(',')
			: 'en';

	let response: Response;
	try {
		response = await fetch(url, {
			headers: {
				'Accept-Language': acceptLanguage,
				'User-Agent': USER_AGENT
			}
		});
	} catch (err) {
		throwGeolocationError(
			'GEOCODE_FAILED',
			`Network error while reverse-geocoding: ${err instanceof Error ? err.message : String(err)}`
		);
	}

	if (!response.ok) {
		throwGeolocationError(
			'GEOCODE_FAILED',
			`Nominatim returned HTTP ${response.status}`
		);
	}

	let data: Record<string, unknown>;
	try {
		data = (await response.json()) as Record<string, unknown>;
	} catch {
		throwGeolocationError(
			'GEOCODE_FAILED',
			'Failed to parse Nominatim JSON response'
		);
	}

	const address = data.address as Record<string, string> | undefined;

	if (!address) {
		throwGeolocationError(
			'GEOCODE_FAILED',
			'Nominatim response did not include address details'
		);
	}

	const postcode = firstOf(address.postcode);
	const area = firstOf(
		address.suburb,
		address.neighbourhood,
		address.village,
		address.hamlet,
		address.town
	);
	const city = firstOf(
		address.city,
		address.town,
		address.county,
		address.state_district
	);
	const state = firstOf(address.state);
	const country = firstOf(address.country);

	// Build a human-readable display name.
	let displayName: string;
	if (area && city) {
		displayName = `${area}, ${city}`;
	} else if (city && country) {
		displayName = `${city}, ${country}`;
	} else {
		displayName = firstOf(
			data.display_name as string | undefined,
			`${lat.toFixed(4)}, ${lng.toFixed(4)}`
		);
	}

	return {
		latitude: lat,
		longitude: lng,
		postcode,
		area,
		city,
		state,
		country,
		displayName
	};
}

/**
 * Detect the user's current location using the browser Geolocation API,
 * reverse-geocode it, cache the result, and return a DetectedLocation.
 *
 * @param useCached  When `true` (the default) a previously cached result
 *                   that is less than 24 hours old will be returned
 *                   immediately without querying the GPS.
 *
 * Throws a GeolocationError when:
 * - the browser does not support geolocation (`NOT_SUPPORTED`)
 * - the user denied permission (`PERMISSION_DENIED`)
 * - the position could not be determined (`POSITION_UNAVAILABLE`)
 * - the request timed out (`TIMEOUT`)
 * - reverse geocoding failed (`GEOCODE_FAILED`)
 */
export async function detectLocation(
	useCached: boolean = true
): Promise<DetectedLocation> {
	// --- SSR guard ---
	if (!isBrowser()) {
		throwGeolocationError(
			'NOT_SUPPORTED',
			'Geolocation is not available during server-side rendering'
		);
	}

	// --- Check browser support ---
	if (!isGeolocationSupported()) {
		throwGeolocationError(
			'NOT_SUPPORTED',
			'Geolocation API is not supported by this browser'
		);
	}

	// --- Return cached result when allowed ---
	if (useCached) {
		const cached = getStoredLocation();
		if (cached) return cached;
	}

	// --- Acquire position from the browser ---
	const position = await new Promise<GeolocationPosition>(
		(resolve, reject) => {
			navigator.geolocation.getCurrentPosition(resolve, reject, {
				enableHighAccuracy: true,
				timeout: 10_000,
				maximumAge: 0
			});
		}
	).catch((err: GeolocationPositionError) => {
		throwGeolocationError(
			mapPositionErrorCode(err.code),
			err.message || 'Failed to retrieve the current position'
		);
	});

	const { latitude, longitude } = position.coords;

	// --- Reverse-geocode ---
	const location = await reverseGeocode(latitude, longitude);

	// --- Cache the result ---
	saveLocation(location);

	return location;
}
