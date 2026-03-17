import type {
	ApiResponse,
	PaginatedResponse,
	ApiError,
	AuthTokens,
	User,
	ProviderProfile,
	Job,
	Quote,
	Review,
	Category,
	Dispute,
	Payment,
	Notification,
	Transaction,
	PointsEntry,
	UserLevel,
	LeaderboardEntry,
	Route,
	RouteStop,
	ScheduleEntry,
	AdminStats,
	KYCApplication,
	ProviderDashboard,
	ProviderEarnings,
	SEOLandingData,
	AIChatMessage,
	AIPriceEstimate,
	LoginRequest,
	RegisterRequest,
	CreateJobRequest,
	CreateQuoteRequest,
	CreateReviewRequest,
	CreateDisputeRequest,
	CreateRouteRequest,
	AddRouteStopRequest
} from '$lib/types';

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8000/api/v1';

// ---------------------------------------------------------------------------
// Token management (client-side only)
// ---------------------------------------------------------------------------

let accessToken: string | null = null;
let refreshToken: string | null = null;

function isBrowser(): boolean {
	return typeof window !== 'undefined';
}

export function setTokens(tokens: AuthTokens): void {
	accessToken = tokens.access_token;
	refreshToken = tokens.refresh_token;
	if (isBrowser()) {
		localStorage.setItem('access_token', tokens.access_token);
		localStorage.setItem('refresh_token', tokens.refresh_token);
	}
}

export function getAccessToken(): string | null {
	if (accessToken) return accessToken;
	if (isBrowser()) {
		accessToken = localStorage.getItem('access_token');
	}
	return accessToken;
}

export function clearTokens(): void {
	accessToken = null;
	refreshToken = null;
	if (isBrowser()) {
		localStorage.removeItem('access_token');
		localStorage.removeItem('refresh_token');
	}
}

function getRefreshToken(): string | null {
	if (refreshToken) return refreshToken;
	if (isBrowser()) {
		refreshToken = localStorage.getItem('refresh_token');
	}
	return refreshToken;
}

// ---------------------------------------------------------------------------
// Core fetch wrapper
// ---------------------------------------------------------------------------

export class ApiClientError extends Error {
	status: number;
	code?: string;
	details?: Record<string, string[]>;

	constructor(message: string, status: number, code?: string, details?: Record<string, string[]>) {
		super(message);
		this.name = 'ApiClientError';
		this.status = status;
		this.code = code;
		this.details = details;
	}
}

interface RequestOptions {
	method?: string;
	body?: unknown;
	headers?: Record<string, string>;
	params?: Record<string, string | number | boolean | undefined>;
	auth?: boolean;
}

async function refreshAccessToken(): Promise<boolean> {
	const token = getRefreshToken();
	if (!token) return false;

	try {
		const response = await fetch(`${BASE_URL}/auth/refresh`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ refresh_token: token })
		});

		if (!response.ok) {
			clearTokens();
			return false;
		}

		const data: ApiResponse<AuthTokens> = await response.json();
		setTokens(data.data);
		return true;
	} catch {
		clearTokens();
		return false;
	}
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
	const { method = 'GET', body, headers = {}, params, auth = true } = options;

	// Build URL with query params
	const url = new URL(`${BASE_URL}${path}`);
	if (params) {
		for (const [key, value] of Object.entries(params)) {
			if (value !== undefined) {
				url.searchParams.set(key, String(value));
			}
		}
	}

	// Build headers
	const reqHeaders: Record<string, string> = {
		'Content-Type': 'application/json',
		...headers
	};

	if (auth) {
		const token = getAccessToken();
		if (token) {
			reqHeaders['Authorization'] = `Bearer ${token}`;
		}
	}

	let response = await fetch(url.toString(), {
		method,
		headers: reqHeaders,
		body: body ? JSON.stringify(body) : undefined
	});

	// Attempt token refresh on 401
	if (response.status === 401 && auth) {
		const refreshed = await refreshAccessToken();
		if (refreshed) {
			reqHeaders['Authorization'] = `Bearer ${getAccessToken()}`;
			response = await fetch(url.toString(), {
				method,
				headers: reqHeaders,
				body: body ? JSON.stringify(body) : undefined
			});
		}
	}

	if (!response.ok) {
		let errorBody: ApiError = { message: 'An unexpected error occurred' };
		try {
			errorBody = await response.json();
		} catch {
			// Could not parse JSON error body
		}
		throw new ApiClientError(
			errorBody.message,
			response.status,
			errorBody.code,
			errorBody.details
		);
	}

	// Handle 204 No Content
	if (response.status === 204) {
		return undefined as T;
	}

	return response.json();
}

// Multipart form request for file uploads
async function requestMultipart<T>(path: string, formData: FormData): Promise<T> {
	const url = new URL(`${BASE_URL}${path}`);
	const reqHeaders: Record<string, string> = {};

	const token = getAccessToken();
	if (token) {
		reqHeaders['Authorization'] = `Bearer ${token}`;
	}

	const response = await fetch(url.toString(), {
		method: 'POST',
		headers: reqHeaders,
		body: formData
	});

	if (!response.ok) {
		let errorBody: ApiError = { message: 'Upload failed' };
		try {
			errorBody = await response.json();
		} catch {
			// ignore
		}
		throw new ApiClientError(errorBody.message, response.status, errorBody.code, errorBody.details);
	}

	return response.json();
}

// ---------------------------------------------------------------------------
// Auth API
// ---------------------------------------------------------------------------

export const auth = {
	requestOtp(phone: string): Promise<ApiResponse<{ expires_in: number }>> {
		return request('/auth/otp/request', {
			method: 'POST',
			body: { phone },
			auth: false
		});
	},

	verifyOtp(data: LoginRequest): Promise<ApiResponse<AuthTokens>> {
		return request('/auth/otp/verify', {
			method: 'POST',
			body: data,
			auth: false
		});
	},

	register(data: RegisterRequest): Promise<ApiResponse<AuthTokens>> {
		return request('/auth/register', {
			method: 'POST',
			body: data,
			auth: false
		});
	},

	me(): Promise<ApiResponse<User>> {
		return request('/auth/me');
	},

	logout(): Promise<void> {
		return request('/auth/logout', { method: 'POST' });
	}
};

// ---------------------------------------------------------------------------
// Users API
// ---------------------------------------------------------------------------

export const users = {
	get(id: string): Promise<ApiResponse<User>> {
		return request(`/users/${id}`);
	},

	update(id: string, data: Partial<User>): Promise<ApiResponse<User>> {
		return request(`/users/${id}`, { method: 'PATCH', body: data });
	},

	list(params?: {
		page?: number;
		per_page?: number;
		role?: string;
		search?: string;
		status?: string;
	}): Promise<PaginatedResponse<User>> {
		return request('/users', { params });
	},

	suspend(id: string): Promise<ApiResponse<User>> {
		return request(`/users/${id}/suspend`, { method: 'POST' });
	},

	activate(id: string): Promise<ApiResponse<User>> {
		return request(`/users/${id}/activate`, { method: 'POST' });
	},

	uploadAvatar(file: File): Promise<ApiResponse<{ url: string }>> {
		const formData = new FormData();
		formData.append('avatar', file);
		return requestMultipart('/users/me/avatar', formData);
	}
};

// ---------------------------------------------------------------------------
// Providers API
// ---------------------------------------------------------------------------

export const providers = {
	get(id: string): Promise<ApiResponse<ProviderProfile>> {
		return request(`/providers/${id}`);
	},

	search(params?: {
		page?: number;
		per_page?: number;
		category?: string;
		postcode?: string;
		radius_km?: number;
		min_rating?: number;
		sort_by?: string;
		query?: string;
	}): Promise<PaginatedResponse<ProviderProfile>> {
		return request('/providers', { params });
	},

	updateProfile(data: Partial<ProviderProfile>): Promise<ApiResponse<ProviderProfile>> {
		return request('/providers/me', { method: 'PATCH', body: data });
	},

	getDashboard(): Promise<ApiResponse<ProviderDashboard>> {
		return request('/providers/me/dashboard');
	},

	getEarnings(params?: {
		period?: 'daily' | 'weekly' | 'monthly';
		from?: string;
		to?: string;
	}): Promise<ApiResponse<ProviderEarnings>> {
		return request('/providers/me/earnings', { params });
	},

	toggleOnline(online: boolean): Promise<ApiResponse<{ is_online: boolean }>> {
		return request('/providers/me/online', { method: 'POST', body: { is_online: online } });
	}
};

// ---------------------------------------------------------------------------
// Jobs API
// ---------------------------------------------------------------------------

export const jobs = {
	get(id: string): Promise<ApiResponse<Job>> {
		return request(`/jobs/${id}`);
	},

	list(params?: {
		page?: number;
		per_page?: number;
		status?: string;
		category?: string;
		search?: string;
	}): Promise<PaginatedResponse<Job>> {
		return request('/jobs', { params });
	},

	create(data: CreateJobRequest): Promise<ApiResponse<Job>> {
		return request('/jobs', { method: 'POST', body: data });
	},

	update(id: string, data: Partial<Job>): Promise<ApiResponse<Job>> {
		return request(`/jobs/${id}`, { method: 'PATCH', body: data });
	},

	updateStatus(id: string, status: string): Promise<ApiResponse<Job>> {
		return request(`/jobs/${id}/status`, { method: 'PATCH', body: { status } });
	},

	cancel(id: string): Promise<ApiResponse<Job>> {
		return request(`/jobs/${id}/cancel`, { method: 'POST' });
	},

	uploadImages(id: string, files: File[]): Promise<ApiResponse<{ urls: string[] }>> {
		const formData = new FormData();
		files.forEach((f) => formData.append('images', f));
		return requestMultipart(`/jobs/${id}/images`, formData);
	}
};

// ---------------------------------------------------------------------------
// Quotes API
// ---------------------------------------------------------------------------

export const quotes = {
	listForJob(jobId: string): Promise<ApiResponse<Quote[]>> {
		return request(`/jobs/${jobId}/quotes`);
	},

	create(jobId: string, data: CreateQuoteRequest): Promise<ApiResponse<Quote>> {
		return request(`/jobs/${jobId}/quotes`, { method: 'POST', body: data });
	},

	accept(jobId: string, quoteId: string): Promise<ApiResponse<Quote>> {
		return request(`/jobs/${jobId}/quotes/${quoteId}/accept`, { method: 'POST' });
	},

	reject(jobId: string, quoteId: string): Promise<ApiResponse<Quote>> {
		return request(`/jobs/${jobId}/quotes/${quoteId}/reject`, { method: 'POST' });
	}
};

// ---------------------------------------------------------------------------
// Reviews API
// ---------------------------------------------------------------------------

export const reviews = {
	listForProvider(providerId: string, params?: {
		page?: number;
		per_page?: number;
	}): Promise<PaginatedResponse<Review>> {
		return request(`/providers/${providerId}/reviews`, { params });
	},

	create(jobId: string, data: CreateReviewRequest): Promise<ApiResponse<Review>> {
		return request(`/jobs/${jobId}/review`, { method: 'POST', body: data });
	},

	listMyReviews(params?: {
		page?: number;
		per_page?: number;
		type?: 'given' | 'received';
	}): Promise<PaginatedResponse<Review>> {
		return request('/reviews/me', { params });
	},

	respond(reviewId: string, response: string): Promise<ApiResponse<Review>> {
		return request(`/reviews/${reviewId}/respond`, { method: 'POST', body: { response } });
	}
};

// ---------------------------------------------------------------------------
// Search API
// ---------------------------------------------------------------------------

export const search = {
	providers(params: {
		query?: string;
		category?: string;
		postcode?: string;
		radius_km?: number;
		min_rating?: number;
		sort_by?: string;
		page?: number;
		per_page?: number;
	}): Promise<PaginatedResponse<ProviderProfile>> {
		return request('/search/providers', { params, auth: false });
	},

	jobs(params: {
		query?: string;
		category?: string;
		status?: string;
		postcode?: string;
		page?: number;
		per_page?: number;
	}): Promise<PaginatedResponse<Job>> {
		return request('/search/jobs', { params });
	},

	categories(query: string): Promise<ApiResponse<Category[]>> {
		return request('/search/categories', { params: { query }, auth: false });
	}
};

// ---------------------------------------------------------------------------
// Categories API
// ---------------------------------------------------------------------------

export const categories = {
	list(): Promise<ApiResponse<Category[]>> {
		return request('/categories', { auth: false });
	},

	get(id: string): Promise<ApiResponse<Category>> {
		return request(`/categories/${id}`, { auth: false });
	}
};

// ---------------------------------------------------------------------------
// Payments API
// ---------------------------------------------------------------------------

export const payments = {
	get(id: string): Promise<ApiResponse<Payment>> {
		return request(`/payments/${id}`);
	},

	listForJob(jobId: string): Promise<ApiResponse<Payment[]>> {
		return request(`/jobs/${jobId}/payments`);
	},

	createOrder(jobId: string, quoteId: string): Promise<ApiResponse<{ client_secret: string; order_id: string }>> {
		return request(`/jobs/${jobId}/pay`, {
			method: 'POST',
			body: { quote_id: quoteId }
		});
	},

	verifyPayment(orderId: string, paymentData: Record<string, string>): Promise<ApiResponse<Payment>> {
		return request(`/payments/${orderId}/verify`, {
			method: 'POST',
			body: paymentData
		});
	},

	getHistory(params?: {
		page?: number;
		per_page?: number;
		status?: string;
		from?: string;
		to?: string;
	}): Promise<PaginatedResponse<Transaction>> {
		return request('/payments/history', { params });
	}
};

// ---------------------------------------------------------------------------
// Disputes API
// ---------------------------------------------------------------------------

export const disputes = {
	get(id: string): Promise<ApiResponse<Dispute>> {
		return request(`/disputes/${id}`);
	},

	create(jobId: string, data: CreateDisputeRequest): Promise<ApiResponse<Dispute>> {
		return request(`/jobs/${jobId}/dispute`, { method: 'POST', body: data });
	},

	list(params?: {
		page?: number;
		per_page?: number;
		status?: string;
		severity?: string;
		type?: string;
	}): Promise<PaginatedResponse<Dispute>> {
		return request('/disputes', { params });
	},

	addEvidence(id: string, files: File[], description?: string): Promise<ApiResponse<Dispute>> {
		const formData = new FormData();
		files.forEach((f) => formData.append('files', f));
		if (description) formData.append('description', description);
		return requestMultipart(`/disputes/${id}/evidence`, formData);
	},

	addMessage(id: string, message: string): Promise<ApiResponse<Dispute>> {
		return request(`/disputes/${id}/messages`, { method: 'POST', body: { message } });
	},

	resolve(id: string, resolution: string): Promise<ApiResponse<Dispute>> {
		return request(`/disputes/${id}/resolve`, {
			method: 'POST',
			body: { resolution }
		});
	},

	assignMediator(id: string, mediatorId: string): Promise<ApiResponse<Dispute>> {
		return request(`/disputes/${id}/assign`, {
			method: 'POST',
			body: { mediator_id: mediatorId }
		});
	}
};

// ---------------------------------------------------------------------------
// Points / Gamification API
// ---------------------------------------------------------------------------

export const points = {
	getBalance(): Promise<ApiResponse<{ balance: number; level: UserLevel }>> {
		return request('/points/balance');
	},

	getHistory(params?: {
		page?: number;
		per_page?: number;
		type?: string;
	}): Promise<PaginatedResponse<PointsEntry>> {
		return request('/points/history', { params });
	},

	getLevel(): Promise<ApiResponse<{ current: UserLevel; next: UserLevel | null; progress: number }>> {
		return request('/points/level');
	},

	getLeaderboard(params?: {
		postcode?: string;
		limit?: number;
	}): Promise<ApiResponse<LeaderboardEntry[]>> {
		return request('/points/leaderboard', { params });
	},

	spend(amount: number, reason: string): Promise<ApiResponse<{ balance: number }>> {
		return request('/points/spend', { method: 'POST', body: { amount, reason } });
	}
};

// ---------------------------------------------------------------------------
// Routes API (Provider)
// ---------------------------------------------------------------------------

export const routes = {
	list(): Promise<ApiResponse<Route[]>> {
		return request('/routes');
	},

	get(id: string): Promise<ApiResponse<Route>> {
		return request(`/routes/${id}`);
	},

	create(data: CreateRouteRequest): Promise<ApiResponse<Route>> {
		return request('/routes', { method: 'POST', body: data });
	},

	update(id: string, data: Partial<Route>): Promise<ApiResponse<Route>> {
		return request(`/routes/${id}`, { method: 'PATCH', body: data });
	},

	delete(id: string): Promise<void> {
		return request(`/routes/${id}`, { method: 'DELETE' });
	},

	addStop(routeId: string, data: AddRouteStopRequest): Promise<ApiResponse<RouteStop>> {
		return request(`/routes/${routeId}/stops`, { method: 'POST', body: data });
	},

	removeStop(routeId: string, stopId: string): Promise<void> {
		return request(`/routes/${routeId}/stops/${stopId}`, { method: 'DELETE' });
	},

	reorderStops(routeId: string, stopIds: string[]): Promise<ApiResponse<Route>> {
		return request(`/routes/${routeId}/reorder`, { method: 'POST', body: { stop_ids: stopIds } });
	},

	optimize(routeId: string): Promise<ApiResponse<Route>> {
		return request(`/routes/${routeId}/optimize`, { method: 'POST' });
	},

	getSchedule(params?: {
		from?: string;
		to?: string;
	}): Promise<ApiResponse<ScheduleEntry[]>> {
		return request('/schedule', { params });
	}
};

// ---------------------------------------------------------------------------
// Notifications API
// ---------------------------------------------------------------------------

export const notifications = {
	list(params?: {
		page?: number;
		per_page?: number;
		unread_only?: boolean;
	}): Promise<PaginatedResponse<Notification>> {
		return request('/notifications', { params });
	},

	markRead(id: string): Promise<void> {
		return request(`/notifications/${id}/read`, { method: 'POST' });
	},

	markAllRead(): Promise<void> {
		return request('/notifications/read-all', { method: 'POST' });
	},

	getUnreadCount(): Promise<ApiResponse<{ count: number }>> {
		return request('/notifications/unread-count');
	}
};

// ---------------------------------------------------------------------------
// Admin API
// ---------------------------------------------------------------------------

export const admin = {
	getStats(): Promise<ApiResponse<AdminStats>> {
		return request('/admin/stats');
	},

	listUsers(params?: {
		page?: number;
		per_page?: number;
		role?: string;
		search?: string;
		status?: string;
	}): Promise<PaginatedResponse<User>> {
		return request('/admin/users', { params });
	},

	pendingKYC(params?: {
		page?: number;
		per_page?: number;
	}): Promise<PaginatedResponse<KYCApplication>> {
		return request('/admin/kyc/pending', { params });
	},

	getKYC(id: string): Promise<ApiResponse<KYCApplication>> {
		return request(`/admin/kyc/${id}`);
	},

	approveKYC(id: string, notes?: string): Promise<ApiResponse<KYCApplication>> {
		return request(`/admin/kyc/${id}/approve`, { method: 'POST', body: { notes } });
	},

	rejectKYC(id: string, reason: string): Promise<ApiResponse<KYCApplication>> {
		return request(`/admin/kyc/${id}/reject`, { method: 'POST', body: { reason } });
	},

	getAnalytics(params?: {
		from?: string;
		to?: string;
		metric?: string;
	}): Promise<ApiResponse<Record<string, unknown>>> {
		return request('/admin/analytics', { params });
	}
};

// ---------------------------------------------------------------------------
// AI API
// ---------------------------------------------------------------------------

export const ai = {
	chat(messages: AIChatMessage[]): Promise<ApiResponse<{ reply: string }>> {
		return request('/ai/chat', { method: 'POST', body: { messages } });
	},

	analyzePhoto(file: File): Promise<ApiResponse<{ description: string; suggestions: string[] }>> {
		const formData = new FormData();
		formData.append('image', file);
		return requestMultipart('/ai/analyze-photo', formData);
	},

	translate(text: string, targetLanguage: string): Promise<ApiResponse<{ translated: string }>> {
		return request('/ai/translate', {
			method: 'POST',
			body: { text, target_language: targetLanguage }
		});
	},

	priceEstimate(params: {
		category_id: string;
		postcode: string;
		description?: string;
	}): Promise<ApiResponse<AIPriceEstimate>> {
		return request('/ai/price-estimate', { method: 'POST', body: params });
	}
};

// ---------------------------------------------------------------------------
// SEO API
// ---------------------------------------------------------------------------

export const seo = {
	getLandingData(service: string, city: string, area: string): Promise<ApiResponse<SEOLandingData>> {
		return request(`/seo/${service}/${city}/${area}`, { auth: false });
	}
};

// ---------------------------------------------------------------------------
// Default export for convenience
// ---------------------------------------------------------------------------

const api = {
	auth,
	users,
	providers,
	jobs,
	quotes,
	reviews,
	search,
	categories,
	payments,
	disputes,
	points,
	routes,
	notifications,
	admin,
	ai,
	seo,
	setTokens,
	clearTokens,
	getAccessToken
};

export default api;
