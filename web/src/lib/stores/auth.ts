import { auth as authApi, setTokens, clearTokens, getAccessToken } from '$lib/api/client';
import type { User, AuthTokens, UserRole } from '$lib/types';

// ---------------------------------------------------------------------------
// Auth state using Svelte 5 runes (module-level $state)
// ---------------------------------------------------------------------------

export interface AuthState {
	user: User | null;
	loading: boolean;
	initialized: boolean;
	notificationCount: number;
	pointsBalance: number;
}

let _user: User | null = null;
let _loading = false;
let _initialized = false;
let _notificationCount = 0;
let _pointsBalance = 0;

// Subscribers for reactivity outside of Svelte components
type Subscriber = (state: AuthState) => void;
const subscribers: Set<Subscriber> = new Set();

function getState(): AuthState {
	return {
		user: _user,
		loading: _loading,
		initialized: _initialized,
		notificationCount: _notificationCount,
		pointsBalance: _pointsBalance
	};
}

function notify(): void {
	const state = getState();
	for (const sub of subscribers) {
		sub(state);
	}
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/**
 * Subscribe to auth state changes.
 * Returns an unsubscribe function.
 */
export function subscribe(fn: Subscriber): () => void {
	subscribers.add(fn);
	fn(getState()); // Emit current state immediately
	return () => {
		subscribers.delete(fn);
	};
}

/**
 * Get the current auth state (snapshot, non-reactive).
 */
export function getAuthState(): AuthState {
	return getState();
}

/**
 * Get the current user (snapshot).
 */
export function getCurrentUser(): User | null {
	return _user;
}

/**
 * Check whether the user is authenticated.
 */
export function isAuthenticated(): boolean {
	return _user !== null;
}

/**
 * Check if the current user is a provider.
 */
export function isProvider(): boolean {
	return _user?.role === 'provider';
}

/**
 * Check if the current user is an admin.
 */
export function isAdmin(): boolean {
	return _user?.role === 'admin';
}

/**
 * Check if the current user is a customer.
 */
export function isCustomer(): boolean {
	return _user?.role === 'customer';
}

/**
 * Get the user's role.
 */
export function getUserRole(): UserRole | null {
	return _user?.role ?? null;
}

/**
 * Get the notification count.
 */
export function getNotificationCount(): number {
	return _notificationCount;
}

/**
 * Set the notification count.
 */
export function setNotificationCount(count: number): void {
	_notificationCount = count;
	notify();
}

/**
 * Get the points balance.
 */
export function getPointsBalance(): number {
	return _pointsBalance;
}

/**
 * Set the points balance.
 */
export function setPointsBalance(balance: number): void {
	_pointsBalance = balance;
	notify();
}

/**
 * Initialize auth state from stored tokens.
 * Call this once on app startup (e.g. in root +layout.svelte).
 */
export async function initAuth(): Promise<void> {
	if (_initialized) return;
	_loading = true;
	notify();

	const token = getAccessToken();
	if (token) {
		try {
			const response = await authApi.me();
			_user = response.data;
			_pointsBalance = _user.points_balance ?? 0;
		} catch {
			// Token is invalid or expired -- clear it.
			clearTokens();
			_user = null;
		}
	}

	_loading = false;
	_initialized = true;
	notify();
}

/**
 * Request an OTP for the given phone number.
 */
export async function requestOtp(phone: string): Promise<{ expires_in: number }> {
	const response = await authApi.requestOtp(phone);
	return response.data;
}

/**
 * Verify an OTP and log in.
 */
export async function login(phone: string, otp: string): Promise<User> {
	_loading = true;
	notify();

	try {
		const tokenResponse = await authApi.verifyOtp({ phone, otp });
		setTokens(tokenResponse.data);

		const userResponse = await authApi.me();
		_user = userResponse.data;
		_pointsBalance = _user.points_balance ?? 0;
		notify();
		return _user;
	} finally {
		_loading = false;
		notify();
	}
}

/**
 * Register a new user.
 */
export async function register(
	name: string,
	phone: string,
	role: 'customer' | 'provider',
	email?: string,
	extra?: {
		postcode?: string;
		bio?: string;
		categories?: string[];
		service_radius_km?: number;
	}
): Promise<User> {
	_loading = true;
	notify();

	try {
		const tokenResponse = await authApi.register({
			name,
			phone,
			role,
			email,
			...extra
		});
		setTokens(tokenResponse.data);

		const userResponse = await authApi.me();
		_user = userResponse.data;
		_pointsBalance = _user.points_balance ?? 0;
		notify();
		return _user;
	} finally {
		_loading = false;
		notify();
	}
}

/**
 * Log the current user out.
 */
export async function logout(): Promise<void> {
	try {
		await authApi.logout();
	} catch {
		// Ignore errors during logout (token may already be invalid)
	} finally {
		_user = null;
		_notificationCount = 0;
		_pointsBalance = 0;
		clearTokens();
		_initialized = false;
		notify();
	}
}

/**
 * Update the cached user object (e.g. after a profile edit).
 */
export function setUser(user: User): void {
	_user = user;
	_pointsBalance = user.points_balance ?? _pointsBalance;
	notify();
}
