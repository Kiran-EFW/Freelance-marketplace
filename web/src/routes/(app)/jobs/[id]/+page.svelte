<script lang="ts">
	import { page } from '$app/stores';
	import { ArrowLeft, MapPin, Clock, User, IndianRupee, MessageSquare, Camera, Star, Phone, Shield } from 'lucide-svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import StarRating from '$lib/components/ui/StarRating.svelte';
	import StatusTimeline from '$lib/components/job/StatusTimeline.svelte';
	import type { Job, Quote, JobStatus } from '$lib/types';

	let jobId = $derived($page.params.id);

	// Mock data for demo
	const mockJob: Job = {
		id: '1',
		customer_id: 'c1',
		customer: { id: 'c1', name: 'Amit Verma', phone: '+919876543210', role: 'customer', is_verified: true, is_active: true, created_at: '2026-01-15', updated_at: '2026-01-15' },
		title: 'Fix leaking kitchen tap and replace washers',
		description: 'The kitchen tap has been dripping for a week. Need a plumber to fix the leak and replace worn out washers. The tap is a standard mixer tap. There may also be an issue with the water pressure.',
		category_id: 'plumbing',
		category: { id: 'plumbing', name: 'Plumbing', slug: 'plumbing', is_active: true },
		status: 'quoted' as JobStatus,
		budget_min: 500,
		budget_max: 2000,
		location: { postcode: '560001', address: 'Koramangala, Bangalore' },
		preferred_date: '2026-03-20',
		preferred_time_slot: 'morning',
		images: [],
		quotes_count: 3,
		created_at: '2026-03-15T10:00:00Z',
		updated_at: '2026-03-15T10:00:00Z'
	};

	const mockQuotes: Quote[] = [
		{
			id: 'q1', job_id: '1', provider_id: 'p1', amount: 800, currency: 'INR', message: 'I can fix this quickly. Have all the parts needed. Available tomorrow morning.', estimated_duration_hours: 1, status: 'pending',
			provider: { id: 'p1', user_id: 'u1', user: { id: 'u1', name: 'Suresh Nair', phone: '+91', role: 'provider', is_verified: true, is_active: true, created_at: '', updated_at: '' }, rating_average: 4.8, rating_count: 124, completion_rate: 0.96, verification_status: 'approved', is_featured: false, categories: [], service_areas: [], portfolio_images: [], certifications: [], created_at: '', updated_at: '', response_time_minutes: 15 },
			created_at: '2026-03-15T11:00:00Z', updated_at: '2026-03-15T11:00:00Z'
		},
		{
			id: 'q2', job_id: '1', provider_id: 'p2', amount: 1200, currency: 'INR', message: 'Will replace the entire tap cartridge for long-lasting results. Warranty included.', estimated_duration_hours: 2, status: 'pending',
			provider: { id: 'p2', user_id: 'u2', user: { id: 'u2', name: 'Ravi Kumar', phone: '+91', role: 'provider', is_verified: true, is_active: true, created_at: '', updated_at: '' }, rating_average: 4.5, rating_count: 87, completion_rate: 0.93, verification_status: 'approved', is_featured: true, categories: [], service_areas: [], portfolio_images: [], certifications: [], created_at: '', updated_at: '', response_time_minutes: 30 },
			created_at: '2026-03-15T12:00:00Z', updated_at: '2026-03-15T12:00:00Z'
		},
		{
			id: 'q3', job_id: '1', provider_id: 'p3', amount: 650, currency: 'INR', message: 'Simple repair job. Can come today evening.', estimated_duration_hours: 1, status: 'pending',
			provider: { id: 'p3', user_id: 'u3', user: { id: 'u3', name: 'Deepak Sharma', phone: '+91', role: 'provider', is_verified: true, is_active: true, created_at: '', updated_at: '' }, rating_average: 4.2, rating_count: 45, completion_rate: 0.89, verification_status: 'approved', is_featured: false, categories: [], service_areas: [], portfolio_images: [], certifications: [], created_at: '', updated_at: '', response_time_minutes: 45 },
			created_at: '2026-03-15T14:00:00Z', updated_at: '2026-03-15T14:00:00Z'
		}
	];

	let reviewRating = $state(0);
	let reviewComment = $state('');

	const statusVariant: Record<JobStatus, 'success' | 'warning' | 'danger' | 'info' | 'neutral'> = {
		draft: 'neutral', open: 'info', quoted: 'warning', accepted: 'info',
		in_progress: 'warning', completed: 'success', cancelled: 'danger', disputed: 'danger'
	};

	const statusLabel: Record<JobStatus, string> = {
		draft: 'Draft', open: 'Open', quoted: 'Quoted', accepted: 'Accepted',
		in_progress: 'In Progress', completed: 'Completed', cancelled: 'Cancelled', disputed: 'Disputed'
	};
</script>

<svelte:head>
	<title>{mockJob.title} - Seva</title>
</svelte:head>

<div class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
	<a href="/jobs" class="inline-flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700">
		<ArrowLeft class="h-4 w-4" />
		Back to Jobs
	</a>

	<!-- Job Header -->
	<div class="mt-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
		<div class="flex flex-wrap items-start justify-between gap-3">
			<div>
				<div class="flex items-center gap-2">
					<Badge variant={statusVariant[mockJob.status]}>{statusLabel[mockJob.status]}</Badge>
					<span class="text-sm text-gray-500 dark:text-gray-400">{mockJob.category?.name}</span>
				</div>
				<h1 class="mt-2 text-2xl font-bold text-gray-900 dark:text-white">{mockJob.title}</h1>
			</div>
			{#if mockJob.status === 'open' || mockJob.status === 'quoted'}
				<Button variant="danger" size="sm">Cancel Job</Button>
			{/if}
		</div>

		<div class="mt-4 grid gap-3 border-t border-gray-200 pt-4 dark:border-gray-700 sm:grid-cols-2 lg:grid-cols-4">
			<div class="flex items-center gap-2">
				<MapPin class="h-5 w-5 text-gray-400" />
				<div>
					<p class="text-xs text-gray-500 dark:text-gray-400">Location</p>
					<p class="text-sm font-medium text-gray-900 dark:text-white">{mockJob.location.postcode}</p>
				</div>
			</div>
			<div class="flex items-center gap-2">
				<Clock class="h-5 w-5 text-gray-400" />
				<div>
					<p class="text-xs text-gray-500 dark:text-gray-400">Preferred Date</p>
					<p class="text-sm font-medium text-gray-900 dark:text-white">{mockJob.preferred_date || 'Flexible'}</p>
				</div>
			</div>
			<div class="flex items-center gap-2">
				<User class="h-5 w-5 text-gray-400" />
				<div>
					<p class="text-xs text-gray-500 dark:text-gray-400">Customer</p>
					<p class="text-sm font-medium text-gray-900 dark:text-white">{mockJob.customer?.name}</p>
				</div>
			</div>
			<div class="flex items-center gap-2">
				<IndianRupee class="h-5 w-5 text-gray-400" />
				<div>
					<p class="text-xs text-gray-500 dark:text-gray-400">Budget</p>
					<p class="text-sm font-medium text-gray-900 dark:text-white">Rs. {mockJob.budget_min} - {mockJob.budget_max}</p>
				</div>
			</div>
		</div>
	</div>

	<!-- Status Timeline -->
	<Card class="mt-6">
		<h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">Status</h2>
		<StatusTimeline currentStatus={mockJob.status} />
	</Card>

	<!-- Description -->
	<Card class="mt-6">
		<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Description</h2>
		<p class="mt-2 text-sm text-gray-600 dark:text-gray-400 whitespace-pre-line">{mockJob.description}</p>
	</Card>

	<!-- Quotes Section (visible when status is quoted) -->
	{#if mockJob.status === 'quoted' || mockJob.status === 'open'}
		<div class="mt-6">
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">
				Quotes ({mockQuotes.length})
			</h2>
			<div class="mt-4 space-y-4">
				{#each mockQuotes as quote}
					<div class="rounded-xl border border-gray-200 bg-white p-5 dark:border-gray-700 dark:bg-gray-800">
						<div class="flex items-start gap-4">
							<Avatar name={quote.provider?.user?.name || 'Provider'} size="lg" />
							<div class="flex-1 min-w-0">
								<div class="flex items-center gap-2">
									<h3 class="font-semibold text-gray-900 dark:text-white">{quote.provider?.user?.name}</h3>
									{#if quote.provider?.verification_status === 'approved'}
										<Shield class="h-4 w-4 text-secondary-500" />
									{/if}
								</div>
								<div class="mt-1 flex items-center gap-3 text-sm">
									<StarRating rating={quote.provider?.rating_average || 0} size="sm" />
									<span class="text-gray-500 dark:text-gray-400">({quote.provider?.rating_count} reviews)</span>
								</div>
								<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">{quote.message}</p>
								<div class="mt-2 flex flex-wrap gap-4 text-sm text-gray-500 dark:text-gray-400">
									{#if quote.estimated_duration_hours}
										<span>Est. {quote.estimated_duration_hours}h</span>
									{/if}
									<span>Response: {quote.provider?.response_time_minutes}min avg</span>
								</div>
							</div>
							<div class="text-right">
								<p class="text-xl font-bold text-gray-900 dark:text-white">Rs. {quote.amount}</p>
								<div class="mt-3 flex flex-col gap-2">
									<Button variant="primary" size="sm">Accept</Button>
									<Button variant="ghost" size="sm">Decline</Button>
								</div>
							</div>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}

	<!-- Assigned Provider (when accepted/in_progress) -->
	{#if mockJob.status === 'accepted' || mockJob.status === 'in_progress'}
		<Card class="mt-6">
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Assigned Provider</h2>
			<div class="mt-4 flex items-center gap-4">
				<Avatar name="Suresh Nair" size="lg" />
				<div class="flex-1">
					<h3 class="font-semibold text-gray-900 dark:text-white">Suresh Nair</h3>
					<StarRating rating={4.8} size="sm" showValue />
				</div>
				<div class="flex gap-2">
					<Button variant="outline" size="sm">
						<Phone class="h-4 w-4" />
						Call
					</Button>
					<Button variant="primary" size="sm">
						<MessageSquare class="h-4 w-4" />
						Chat
					</Button>
				</div>
			</div>
		</Card>
	{/if}

	<!-- Review Form (when completed) -->
	{#if mockJob.status === 'completed'}
		<Card class="mt-6">
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Leave a Review</h2>
			<div class="mt-4 space-y-4">
				<div>
					<label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">Rating</label>
					<StarRating bind:rating={reviewRating} editable size="lg" />
				</div>
				<div>
					<label for="review" class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">Comment</label>
					<textarea
						id="review"
						bind:value={reviewComment}
						rows="3"
						placeholder="How was your experience?"
						class="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
					></textarea>
				</div>
				<Button variant="primary">Submit Review</Button>
			</div>
		</Card>
	{/if}
</div>
