<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { ArrowLeft, MapPin, Clock, User, IndianRupee, MessageSquare, Camera, Star, Phone, Shield, Loader2 } from 'lucide-svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import StarRating from '$lib/components/ui/StarRating.svelte';
	import StatusTimeline from '$lib/components/job/StatusTimeline.svelte';
	import { toastSuccess, toastError } from '$lib/stores/toast';
	import api from '$lib/api/client';
	import type { Job, Quote, JobStatus } from '$lib/types';

	let jobId = $derived($page.params.id);
	let loading = $state(true);
	let error = $state('');
	let job = $state<Job | null>(null);
	let jobQuotes = $state<Quote[]>([]);

	let reviewRating = $state(0);
	let reviewComment = $state('');
	let submittingReview = $state(false);

	onMount(async () => {
		try {
			const [jobRes, quotesRes] = await Promise.all([
				api.jobs.get(jobId),
				api.quotes.listForJob(jobId).catch(() => ({ data: [] }))
			]);
			job = jobRes.data;
			jobQuotes = quotesRes.data || [];
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load job details';
		} finally {
			loading = false;
		}
	});

	async function acceptQuote(quoteId: string) {
		try {
			await api.quotes.accept(jobId, quoteId);
			toastSuccess('Quote accepted!');
			// Refresh job data
			const jobRes = await api.jobs.get(jobId);
			job = jobRes.data;
		} catch (err) {
			toastError(err instanceof Error ? err.message : 'Failed to accept quote');
		}
	}

	async function rejectQuote(quoteId: string) {
		try {
			await api.quotes.reject(jobId, quoteId);
			toastSuccess('Quote declined');
			jobQuotes = jobQuotes.filter(q => q.id !== quoteId);
		} catch (err) {
			toastError(err instanceof Error ? err.message : 'Failed to decline quote');
		}
	}

	async function cancelJob() {
		try {
			await api.jobs.cancel(jobId);
			toastSuccess('Job cancelled');
			const jobRes = await api.jobs.get(jobId);
			job = jobRes.data;
		} catch (err) {
			toastError(err instanceof Error ? err.message : 'Failed to cancel job');
		}
	}

	async function submitReview() {
		if (!reviewRating) return;
		submittingReview = true;
		try {
			await api.reviews.create(jobId, {
				rating: reviewRating,
				comment: reviewComment
			});
			toastSuccess('Review submitted!');
			reviewRating = 0;
			reviewComment = '';
		} catch (err) {
			toastError(err instanceof Error ? err.message : 'Failed to submit review');
		} finally {
			submittingReview = false;
		}
	}

	const statusVariant: Record<string, 'success' | 'warning' | 'danger' | 'info' | 'neutral'> = {
		draft: 'neutral', open: 'info', quoted: 'warning', accepted: 'info',
		in_progress: 'warning', completed: 'success', cancelled: 'danger', disputed: 'danger'
	};

	const statusLabel: Record<string, string> = {
		draft: 'Draft', open: 'Open', quoted: 'Quoted', accepted: 'Accepted',
		in_progress: 'In Progress', completed: 'Completed', cancelled: 'Cancelled', disputed: 'Disputed'
	};
</script>

<svelte:head>
	<title>{job?.title || 'Job Details'} - Seva</title>
</svelte:head>

{#if loading}
<div class="flex items-center justify-center py-20">
	<Loader2 class="h-8 w-8 animate-spin text-primary-600" />
</div>
{:else if error || !job}
<div class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
	<a href="/jobs" class="inline-flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700">
		<ArrowLeft class="h-4 w-4" />
		Back to Jobs
	</a>
	<div class="mt-6 rounded-lg border border-red-200 bg-red-50 p-6 text-center dark:border-red-800 dark:bg-red-900/20">
		<p class="text-red-600 dark:text-red-400">{error || 'Job not found'}</p>
	</div>
</div>
{:else}
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
					<Badge variant={statusVariant[job.status] || 'neutral'}>{statusLabel[job.status] || job.status}</Badge>
					<span class="text-sm text-gray-500 dark:text-gray-400">{job.category?.name}</span>
				</div>
				<h1 class="mt-2 text-2xl font-bold text-gray-900 dark:text-white">{job.title}</h1>
			</div>
			{#if job.status === 'open' || job.status === 'quoted'}
				<Button variant="danger" size="sm" onclick={cancelJob}>Cancel Job</Button>
			{/if}
		</div>

		<div class="mt-4 grid gap-3 border-t border-gray-200 pt-4 dark:border-gray-700 sm:grid-cols-2 lg:grid-cols-4">
			<div class="flex items-center gap-2">
				<MapPin class="h-5 w-5 text-gray-400" />
				<div>
					<p class="text-xs text-gray-500 dark:text-gray-400">Location</p>
					<p class="text-sm font-medium text-gray-900 dark:text-white">{job.location?.postcode || job.location?.address || 'N/A'}</p>
				</div>
			</div>
			<div class="flex items-center gap-2">
				<Clock class="h-5 w-5 text-gray-400" />
				<div>
					<p class="text-xs text-gray-500 dark:text-gray-400">Preferred Date</p>
					<p class="text-sm font-medium text-gray-900 dark:text-white">{job.preferred_date || 'Flexible'}</p>
				</div>
			</div>
			<div class="flex items-center gap-2">
				<User class="h-5 w-5 text-gray-400" />
				<div>
					<p class="text-xs text-gray-500 dark:text-gray-400">Customer</p>
					<p class="text-sm font-medium text-gray-900 dark:text-white">{job.customer?.name || 'N/A'}</p>
				</div>
			</div>
			<div class="flex items-center gap-2">
				<IndianRupee class="h-5 w-5 text-gray-400" />
				<div>
					<p class="text-xs text-gray-500 dark:text-gray-400">Budget</p>
					<p class="text-sm font-medium text-gray-900 dark:text-white">Rs. {job.budget_min}{job.budget_max ? ` - ${job.budget_max}` : ''}</p>
				</div>
			</div>
		</div>
	</div>

	<!-- Status Timeline -->
	<Card class="mt-6">
		<h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">Status</h2>
		<StatusTimeline currentStatus={job.status} />
	</Card>

	<!-- Description -->
	<Card class="mt-6">
		<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Description</h2>
		<p class="mt-2 text-sm text-gray-600 dark:text-gray-400 whitespace-pre-line">{job.description}</p>
	</Card>

	<!-- Quotes Section (visible when status is quoted) -->
	{#if (job.status === 'quoted' || job.status === 'open') && jobQuotes.length > 0}
		<div class="mt-6">
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">
				Quotes ({jobQuotes.length})
			</h2>
			<div class="mt-4 space-y-4">
				{#each jobQuotes as quote}
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
									<span class="text-gray-500 dark:text-gray-400">({quote.provider?.rating_count || 0} reviews)</span>
								</div>
								<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">{quote.message}</p>
								<div class="mt-2 flex flex-wrap gap-4 text-sm text-gray-500 dark:text-gray-400">
									{#if quote.estimated_duration_hours}
										<span>Est. {quote.estimated_duration_hours}h</span>
									{/if}
									{#if quote.provider?.response_time_minutes}
										<span>Response: {quote.provider.response_time_minutes}min avg</span>
									{/if}
								</div>
							</div>
							<div class="text-right">
								<p class="text-xl font-bold text-gray-900 dark:text-white">Rs. {quote.amount}</p>
								<div class="mt-3 flex flex-col gap-2">
									<Button variant="primary" size="sm" onclick={() => acceptQuote(quote.id)}>Accept</Button>
									<Button variant="ghost" size="sm" onclick={() => rejectQuote(quote.id)}>Decline</Button>
								</div>
							</div>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}

	<!-- Assigned Provider (when accepted/in_progress) -->
	{#if job.status === 'accepted' || job.status === 'in_progress'}
		<Card class="mt-6">
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Assigned Provider</h2>
			<div class="mt-4 flex items-center gap-4">
				<Avatar name={job.provider?.user?.name || 'Provider'} size="lg" />
				<div class="flex-1">
					<h3 class="font-semibold text-gray-900 dark:text-white">{job.provider?.user?.name || 'Provider'}</h3>
					<StarRating rating={job.provider?.rating_average || 0} size="sm" showValue />
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
	{#if job.status === 'completed'}
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
				<Button variant="primary" onclick={submitReview} loading={submittingReview}>Submit Review</Button>
			</div>
		</Card>
	{/if}
</div>
{/if}
