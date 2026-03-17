<script lang="ts">
	import { onMount } from 'svelte';
	import { Star, MessageSquare, Loader2 } from 'lucide-svelte';
	import Tabs from '$lib/components/ui/Tabs.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import StarRating from '$lib/components/ui/StarRating.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import api from '$lib/api/client';

	let activeTab = $state('given');
	let loading = $state(true);
	let error = $state('');

	let reviewsGiven = $state<any[]>([]);
	let reviewsReceived = $state<any[]>([]);

	const tabs = $derived([
		{ id: 'given', label: 'Reviews I Wrote', count: reviewsGiven.length },
		{ id: 'received', label: 'Reviews About Me', count: reviewsReceived.length }
	]);

	onMount(async () => {
		try {
			const [givenRes, receivedRes] = await Promise.all([
				api.reviews.listMyReviews({ type: 'given', per_page: 20 }),
				api.reviews.listMyReviews({ type: 'received', per_page: 20 })
			]);

			reviewsGiven = (givenRes.data || []).map((r: any) => ({
				id: r.id,
				providerName: r.provider?.user?.name || r.reviewee?.name || 'Provider',
				jobTitle: r.job?.title || r.job_title || '',
				rating: r.rating,
				comment: r.comment || '',
				date: r.created_at?.split('T')[0] || '',
				response: r.response || null
			}));

			reviewsReceived = (receivedRes.data || []).map((r: any) => ({
				id: r.id,
				reviewerName: r.reviewer?.name || r.customer?.name || 'Customer',
				jobTitle: r.job?.title || r.job_title || '',
				rating: r.rating,
				comment: r.comment || '',
				date: r.created_at?.split('T')[0] || '',
				response: r.response || null
			}));
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load reviews';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Reviews - Seva</title>
</svelte:head>

{#if loading}
<div class="flex items-center justify-center py-20">
	<Loader2 class="h-8 w-8 animate-spin text-primary-600" />
</div>
{:else if error}
<div class="mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8">
	<div class="rounded-lg border border-red-200 bg-red-50 p-6 text-center dark:border-red-800 dark:bg-red-900/20">
		<p class="text-red-600 dark:text-red-400">{error}</p>
	</div>
</div>
{:else}
<div class="mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8">
	<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Reviews</h1>
	<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">See reviews you've written and reviews about you.</p>

	<div class="mt-6">
		<Tabs {tabs} bind:activeTab />
	</div>

	<div class="mt-6 space-y-4">
		{#if activeTab === 'given'}
			{#each reviewsGiven as review}
				<Card>
					<div class="flex items-start gap-4">
						<Avatar name={review.providerName} size="md" />
						<div class="flex-1 min-w-0">
							<div class="flex items-start justify-between gap-2">
								<div>
									<h3 class="font-medium text-gray-900 dark:text-white">{review.providerName}</h3>
									<span class="text-xs text-primary-600">{review.jobTitle}</span>
								</div>
								<span class="text-xs text-gray-500 dark:text-gray-400 shrink-0">{review.date}</span>
							</div>
							<StarRating rating={review.rating} size="sm" class="mt-1" />
							<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">{review.comment}</p>
							{#if review.response}
								<div class="mt-3 rounded-lg bg-gray-50 p-3 dark:bg-gray-700/50">
									<p class="text-xs font-medium text-gray-500 dark:text-gray-400">Provider response:</p>
									<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{review.response}</p>
								</div>
							{/if}
						</div>
					</div>
				</Card>
			{/each}
		{:else}
			{#each reviewsReceived as review}
				<Card>
					<div class="flex items-start gap-4">
						<Avatar name={review.reviewerName} size="md" />
						<div class="flex-1 min-w-0">
							<div class="flex items-start justify-between gap-2">
								<div>
									<h3 class="font-medium text-gray-900 dark:text-white">{review.reviewerName}</h3>
									<span class="text-xs text-primary-600">{review.jobTitle}</span>
								</div>
								<span class="text-xs text-gray-500 dark:text-gray-400 shrink-0">{review.date}</span>
							</div>
							<StarRating rating={review.rating} size="sm" class="mt-1" />
							<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">{review.comment}</p>
							{#if review.response}
								<div class="mt-3 rounded-lg bg-gray-50 p-3 dark:bg-gray-700/50">
									<p class="text-xs font-medium text-gray-500 dark:text-gray-400">Your response:</p>
									<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{review.response}</p>
								</div>
							{:else}
								<button class="mt-2 flex items-center gap-1 text-xs text-primary-600 hover:text-primary-700">
									<MessageSquare class="h-3 w-3" />
									Respond
								</button>
							{/if}
						</div>
					</div>
				</Card>
			{/each}
		{/if}
	</div>
</div>
{/if}
