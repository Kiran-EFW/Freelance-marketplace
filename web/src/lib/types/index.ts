// ============================================================
// Core domain types for the Seva platform.
// These mirror the API response schemas.
// ============================================================

// --- Enums / Union Types ---

export type UserRole = 'customer' | 'provider' | 'admin';

export type JobStatus =
	| 'draft'
	| 'open'
	| 'quoted'
	| 'accepted'
	| 'in_progress'
	| 'completed'
	| 'cancelled'
	| 'disputed';

export type QuoteStatus = 'pending' | 'accepted' | 'rejected' | 'withdrawn';

export type PaymentStatus = 'pending' | 'held' | 'released' | 'refunded' | 'failed';

export type DisputeStatus = 'open' | 'under_review' | 'resolved' | 'escalated' | 'closed';

export type VerificationStatus = 'pending' | 'approved' | 'rejected';

export type TransactionType = 'payment' | 'refund' | 'payout' | 'fee';

export type DisputeType = 'quality' | 'no_show' | 'payment' | 'damage' | 'other';

export type NotificationType =
	| 'job_created'
	| 'quote_received'
	| 'quote_accepted'
	| 'quote_rejected'
	| 'job_started'
	| 'job_completed'
	| 'review_received'
	| 'payment_received'
	| 'dispute_opened'
	| 'dispute_resolved'
	| 'kyc_approved'
	| 'kyc_rejected'
	| 'points_earned'
	| 'level_up'
	| 'system';

// --- User ---

export interface User {
	id: string;
	phone: string;
	email?: string;
	name: string;
	role: UserRole;
	avatar_url?: string;
	is_verified: boolean;
	is_active: boolean;
	address?: string;
	postcode?: string;
	language?: string;
	preferred_language?: string;
	points_balance?: number;
	level?: UserLevel;
	notification_preferences?: {
		job_updates?: boolean;
		quotes?: boolean;
		marketing?: boolean;
		sms?: boolean;
	};
	provider_profile?: ProviderProfile;
	created_at: string;
	updated_at: string;
}

// --- Provider Profile ---

export interface ProviderProfile {
	id: string;
	user_id: string;
	user?: User;
	business_name?: string;
	bio?: string;
	categories: Category[];
	service_areas: ServiceArea[];
	hourly_rate?: number;
	rating_average: number;
	rating_count: number;
	response_time_minutes?: number;
	completion_rate: number;
	trust_score?: number;
	verification_status: VerificationStatus;
	is_featured: boolean;
	is_online?: boolean;
	portfolio_images: string[];
	certifications: Certification[];
	total_jobs_completed?: number;
	completed_jobs_count?: number;
	languages?: string[];
	working_hours?: string;
	created_at: string;
	updated_at: string;
}

export interface Certification {
	id: string;
	name: string;
	issuer: string;
	issued_at: string;
	expires_at?: string;
	document_url?: string;
	verified: boolean;
}

export interface ServiceArea {
	id: string;
	postcode: string;
	radius_km: number;
	latitude: number;
	longitude: number;
}

// --- Category ---

export interface Category {
	id: string;
	name: string;
	slug: string;
	description?: string;
	icon?: string;
	parent_id?: string;
	children?: Category[];
	is_active: boolean;
}

// --- Job ---

export interface Job {
	id: string;
	customer_id: string;
	customer?: User;
	title: string;
	description: string;
	category_id: string;
	category?: Category;
	status: JobStatus;
	budget_min?: number;
	budget_max?: number;
	location: JobLocation;
	preferred_date?: string;
	preferred_time_slot?: string;
	payment_method?: string;
	images: string[];
	quotes_count: number;
	accepted_quote_id?: string;
	assigned_provider_id?: string;
	assigned_provider?: ProviderProfile;
	provider?: ProviderProfile;
	created_at: string;
	updated_at: string;
}

export interface JobLocation {
	postcode: string;
	address?: string;
	latitude?: number;
	longitude?: number;
}

// --- Quote ---

export interface Quote {
	id: string;
	job_id: string;
	provider_id: string;
	provider?: ProviderProfile;
	amount: number;
	currency: string;
	message?: string;
	estimated_duration_hours?: number;
	status: QuoteStatus;
	valid_until?: string;
	created_at: string;
	updated_at: string;
}

// --- Review ---

export interface Review {
	id: string;
	job_id: string;
	job?: Job;
	reviewer_id: string;
	reviewer?: User;
	reviewee_id: string;
	reviewee?: User;
	rating: number;
	comment?: string;
	images: string[];
	response?: string;
	response_at?: string;
	created_at: string;
	updated_at: string;
}

// --- Dispute ---

export interface Dispute {
	id: string;
	job_id: string;
	job?: Job;
	raised_by_id: string;
	raised_by?: User;
	against_id: string;
	against?: User;
	type: DisputeType;
	reason: string;
	description: string;
	status: DisputeStatus;
	evidence: DisputeEvidence[];
	messages: DisputeMessage[];
	resolution?: string;
	resolved_by_id?: string;
	resolved_at?: string;
	mediator_id?: string;
	mediator?: User;
	severity?: 'low' | 'medium' | 'high' | 'critical';
	created_at: string;
	updated_at: string;
}

export interface DisputeEvidence {
	id: string;
	dispute_id: string;
	uploaded_by_id: string;
	uploaded_by?: User;
	file_url: string;
	file_type: string;
	description?: string;
	created_at: string;
}

export interface DisputeMessage {
	id: string;
	dispute_id: string;
	sender_id: string;
	sender?: User;
	message: string;
	is_internal: boolean;
	created_at: string;
}

// --- Payment / Transaction ---

export interface Payment {
	id: string;
	job_id: string;
	quote_id: string;
	payer_id: string;
	payee_id: string;
	amount: number;
	currency: string;
	platform_fee: number;
	provider_amount: number;
	status: PaymentStatus;
	payment_method?: string;
	stripe_payment_intent_id?: string;
	paid_at?: string;
	released_at?: string;
	refunded_at?: string;
	created_at: string;
	updated_at: string;
}

export interface Transaction {
	id: string;
	user_id: string;
	job_id?: string;
	job?: Job;
	type: TransactionType;
	amount: number;
	currency: string;
	description: string;
	status: PaymentStatus;
	payment_method?: string;
	created_at: string;
}

// --- Notification ---

export interface Notification {
	id: string;
	user_id: string;
	type: NotificationType | string;
	title: string;
	message: string;
	data?: Record<string, unknown>;
	action_url?: string;
	read_at?: string;
	created_at: string;
}

// --- Points / Gamification ---

export interface PointsEntry {
	id: string;
	user_id: string;
	points: number;
	type: 'earned' | 'spent' | 'bonus' | 'penalty';
	reason: string;
	reference_type?: string;
	reference_id?: string;
	created_at: string;
}

export interface UserLevel {
	level: number;
	name: string;
	min_points: number;
	max_points: number;
	benefits: string[];
}

export interface LeaderboardEntry {
	rank: number;
	user_id: string;
	user?: User;
	provider?: ProviderProfile;
	total_points: number;
	level: UserLevel;
}

// --- Routes (Provider) ---

export interface Route {
	id: string;
	provider_id: string;
	name: string;
	description?: string;
	stops: RouteStop[];
	next_visit?: string;
	next_visit_date?: string;
	recurrence?: 'daily' | 'weekly' | 'biweekly' | 'monthly';
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

export interface RouteStop {
	id: string;
	route_id: string;
	customer_id: string;
	customer?: User;
	address: string;
	postcode: string;
	latitude?: number;
	longitude?: number;
	order: number;
	notes?: string;
	tree_count?: number;
	last_visit_date?: string;
	next_visit_date?: string;
	estimated_duration_minutes?: number;
}

// --- Schedule ---

export interface ScheduleEntry {
	id: string;
	provider_id: string;
	date: string;
	start_time: string;
	end_time: string;
	type: 'job' | 'route_visit' | 'blocked';
	job_id?: string;
	job?: Job;
	route_stop_id?: string;
	route_stop?: RouteStop;
	notes?: string;
}

// --- Search Filters ---

export interface SearchFilters {
	query?: string;
	category?: string;
	postcode?: string;
	radius_km?: number;
	min_rating?: number;
	max_budget?: number;
	min_budget?: number;
	sort_by?: 'rating' | 'distance' | 'price' | 'reviews' | 'newest';
	sort_order?: 'asc' | 'desc';
	page?: number;
	per_page?: number;
}

// --- Admin ---

export interface AdminStats {
	total_users: number;
	total_providers: number;
	total_customers: number;
	active_providers: number;
	total_jobs: number;
	jobs_today: number;
	jobs_this_week: number;
	jobs_this_month: number;
	revenue_today: number;
	revenue_this_week: number;
	revenue_this_month: number;
	revenue_total: number;
	open_disputes: number;
	pending_kyc: number;
	user_growth_rate: number;
	job_growth_rate: number;
	users_growth?: number;
	providers_growth?: number;
	jobs_growth?: number;
	revenue?: number;
	revenue_growth?: number;
	active_jobs?: number;
	completion_rate?: number;
	revenue_chart?: { month?: string; label?: string; amount?: number; revenue?: number }[];
	revenue_trend?: { month?: string; label?: string; amount?: number; revenue?: number }[];
	recent_activity?: { id: string; type: string; action?: string; title?: string; detail?: string; description?: string; time?: string; created_at?: string }[];
}

export interface KYCApplication {
	id: string;
	user_id: string;
	user?: User;
	provider?: ProviderProfile;
	document_type: string;
	document_url: string;
	selfie_url?: string;
	extracted_data?: Record<string, string>;
	status: VerificationStatus;
	reviewer_id?: string;
	reviewer_notes?: string;
	submitted_at: string;
	reviewed_at?: string;
}

// --- Provider Dashboard ---

export interface ProviderDashboard {
	earnings_today: number;
	earnings_this_week: number;
	earnings_this_month: number;
	active_jobs: number;
	completed_this_month: number;
	trust_score: number;
	rating_average: number;
	rating_count: number;
	rating_distribution: Record<number, number>;
	upcoming_schedule: ScheduleEntry[];
	is_online: boolean;
}

export interface ProviderEarnings {
	total: number;
	pending: number;
	total_earnings?: number;
	pending_payout?: number;
	chart_data?: { label?: string; date?: string; period?: string; amount?: number; earnings?: number }[];
	breakdown?: { label?: string; date?: string; period?: string; amount?: number; earnings?: number }[];
	daily: { date: string; amount: number }[];
	weekly: { week: string; amount: number }[];
	monthly: { month: string; amount: number }[];
	payouts: Payout[];
}

export interface Payout {
	id: string;
	amount: number;
	status: 'pending' | 'processing' | 'completed' | 'failed';
	bank_account_last4?: string;
	created_at: string;
	completed_at?: string;
}

// --- SEO Landing ---

export interface SEOLandingData {
	service: string;
	city: string;
	area: string;
	providers: ProviderProfile[];
	average_price: number;
	min_price: number;
	max_price: number;
	total_providers: number;
	recent_reviews: Review[];
	faqs: { question: string; answer: string }[];
}

// --- Messaging ---

export type MessageType = 'text' | 'image' | 'quote' | 'system';

export interface Conversation {
	id: string;
	job_id?: string;
	participant_1: string;
	participant_2: string;
	last_message_at?: string;
	last_message_preview?: string;
	is_archived_1: boolean;
	is_archived_2: boolean;
	created_at: string;
	updated_at: string;
	// Enriched fields from API joins or client-side resolution
	other_user?: User;
	unread_count?: number;
}

export interface ChatMessage {
	id: string;
	conversation_id: string;
	sender_id: string;
	content: string;
	message_type: MessageType;
	attachment_url?: string;
	attachment_type?: string;
	metadata?: Record<string, unknown>;
	is_read: boolean;
	read_at?: string;
	created_at: string;
}

// --- AI ---

export interface AIChatMessage {
	role: 'user' | 'assistant';
	content: string;
}

export interface AIPriceEstimate {
	estimated_min: number;
	estimated_max: number;
	confidence: number;
	factors: string[];
}

// --- API Response Wrappers ---

export interface ApiResponse<T> {
	data: T;
	message?: string;
}

export interface PaginatedResponse<T> {
	data: T[];
	total: number;
	page: number;
	per_page: number;
	total_pages: number;
	meta?: {
		total?: number;
		total_pages?: number;
		page?: number;
		per_page?: number;
	};
}

export interface ApiError {
	message: string;
	code?: string;
	details?: Record<string, string[]>;
}

// --- Auth ---

export interface AuthTokens {
	access_token: string;
	refresh_token: string;
	token_type: string;
	expires_in: number;
}

export interface LoginRequest {
	phone: string;
	otp: string;
}

export interface RegisterRequest {
	name: string;
	phone: string;
	email?: string;
	role: UserRole;
	postcode?: string;
	bio?: string;
	categories?: string[];
	service_radius_km?: number;
}

// --- Request Bodies ---

export interface CreateJobRequest {
	title: string;
	description: string;
	category_id: string;
	location: JobLocation;
	budget_min?: number;
	budget_max?: number;
	preferred_date?: string;
	preferred_time_slot?: string;
	payment_method?: string;
	images?: string[];
}

export interface CreateQuoteRequest {
	amount: number;
	currency?: string;
	message?: string;
	estimated_duration_hours?: number;
	valid_until?: string;
}

export interface CreateReviewRequest {
	rating: number;
	comment?: string;
	images?: string[];
}

export interface CreateDisputeRequest {
	type: DisputeType;
	reason: string;
	description: string;
}

export interface CreateRouteRequest {
	name: string;
	description?: string;
	recurrence?: 'daily' | 'weekly' | 'biweekly' | 'monthly';
}

export interface AddRouteStopRequest {
	customer_id?: string;
	customer_name?: string;
	address: string;
	postcode: string;
	latitude?: number;
	longitude?: number;
	notes?: string;
	tree_count?: number;
}

// --- Organization (B2B) ---

export type OrgType = 'housing_society' | 'company' | 'institution';
export type OrgStatus = 'active' | 'suspended' | 'pending';
export type OrgRole = 'admin' | 'manager' | 'member';
export type MemberStatus = 'active' | 'inactive' | 'invited';
export type RequestPriority = 'low' | 'medium' | 'high' | 'urgent';
export type ServiceRequestStatus = 'pending' | 'assigned' | 'in_progress' | 'completed' | 'cancelled';

export interface Organization {
	id: string;
	name: string;
	type: OrgType;
	address?: string;
	postcode?: string;
	city?: string;
	state?: string;
	country: string;
	contact_phone?: string;
	contact_email?: string;
	logo_url?: string;
	settings?: Record<string, unknown>;
	status: OrgStatus;
	created_at: string;
	updated_at: string;
}

export interface OrganizationMember {
	id: string;
	org_id: string;
	user_id: string;
	role: OrgRole;
	joined_at: string;
	status: MemberStatus;
	user_name?: string;
	user_phone?: string;
	user_email?: string;
}

export interface OrganizationServiceRequest {
	id: string;
	org_id: string;
	requested_by: string;
	category_id: string;
	title: string;
	description?: string;
	priority: RequestPriority;
	status: ServiceRequestStatus;
	assigned_provider_id?: string;
	scheduled_at?: string;
	completed_at?: string;
	notes?: string;
	created_at: string;
	updated_at: string;
	requester_name?: string;
	category_slug?: string;
	category_name?: string;
	provider_name?: string;
}

export interface OrgStats {
	total_requests: number;
	pending_requests: number;
	completed_requests: number;
	in_progress_requests: number;
	assigned_requests: number;
	active_members: number;
}

export interface CreateOrganizationRequest {
	name: string;
	type: OrgType;
	address?: string;
	postcode?: string;
	city?: string;
	state?: string;
	country?: string;
	contact_phone?: string;
	contact_email?: string;
}

export interface CreateServiceRequestPayload {
	category_id: string;
	title: string;
	description?: string;
	priority?: RequestPriority;
	scheduled_at?: string;
	notes?: string;
}

// --- Safety ---

export type SOSAlertStatus = 'active' | 'responded' | 'resolved' | 'false_alarm';

export interface SOSAlert {
	id: string;
	user_id: string;
	job_id?: string;
	latitude: number;
	longitude: number;
	status: SOSAlertStatus;
	emergency_contacts_notified: boolean;
	notes?: string;
	created_at: string;
	resolved_at?: string;
}

export interface LocationShare {
	id: string;
	job_id: string;
	user_id: string;
	latitude: number;
	longitude: number;
	accuracy?: number;
	shared_at: string;
}

export interface EmergencyContact {
	id: string;
	user_id: string;
	name: string;
	phone: string;
	relationship?: string;
	created_at: string;
}
