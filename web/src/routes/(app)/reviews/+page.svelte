<script lang="ts">
	import { Star, MessageSquare } from 'lucide-svelte';
	import Tabs from '$lib/components/ui/Tabs.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import StarRating from '$lib/components/ui/StarRating.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';

	let activeTab = $state('given');

	const tabs = [
		{ id: 'given', label: 'Reviews I Wrote', count: 5 },
		{ id: 'received', label: 'Reviews About Me', count: 3 }
	];

	const mockReviewsGiven = [
		{ id: '1', providerName: 'Suresh Nair', jobTitle: 'Fix leaking tap', rating: 5, comment: 'Excellent work! Fixed the leak quickly and cleanly. Very professional.', date: '2026-03-10', response: null },
		{ id: '2', providerName: 'Ravi Kumar', jobTitle: 'Electrical wiring', rating: 4, comment: 'Good job overall. Slightly delayed but quality work.', date: '2026-02-28', response: 'Thank you for the feedback! Sorry about the delay, will ensure on-time service next time.' },
		{ id: '3', providerName: 'Deepak Sharma', jobTitle: 'Deep cleaning', rating: 5, comment: 'Thorough cleaning, very satisfied with the results.', date: '2026-02-15', response: null },
		{ id: '4', providerName: 'Anita Gupta', jobTitle: 'Garden maintenance', rating: 3, comment: 'Decent work but missed some areas. Had to follow up.', date: '2026-01-20', response: 'Apologies for the oversight. Will make sure to cover everything next visit.' },
		{ id: '5', providerName: 'Kumar Singh', jobTitle: 'Painting bedroom', rating: 5, comment: 'Beautiful finish, clean work, and completed on time.', date: '2026-01-05', response: null }
	];

	const mockReviewsReceived = [
		{ id: '6', reviewerName: 'Priya Menon', jobTitle: 'Tap installation', rating: 5, comment: 'Very knowledgeable and efficient. Highly recommend!', date: '2026-03-12', response: null },
		{ id: '7', reviewerName: 'Arjun Das', jobTitle: 'Pipe repair', rating: 4, comment: 'Fixed the pipe issue. Good communication throughout.', date: '2026-03-01', response: 'Thank you Arjun! Glad I could help.' },
		{ id: '8', reviewerName: 'Meera Reddy', jobTitle: 'Water heater install', rating: 5, comment: 'Professional installation, everything works perfectly.', date: '2026-02-20', response: null }
	];
</script>

<svelte:head>
	<title>Reviews - Seva</title>
</svelte:head>

<div class="mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8">
	<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Reviews</h1>
	<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">See reviews you've written and reviews about you.</p>

	<div class="mt-6">
		<Tabs {tabs} bind:activeTab />
	</div>

	<div class="mt-6 space-y-4">
		{#if activeTab === 'given'}
			{#each mockReviewsGiven as review}
				<Card>
					<div class="flex items-start gap-4">
						<Avatar name={review.providerName} size="md" />
						<div class="flex-1 min-w-0">
							<div class="flex items-start justify-between gap-2">
								<div>
									<h3 class="font-medium text-gray-900 dark:text-white">{review.providerName}</h3>
									<a href="#" class="text-xs text-primary-600 hover:text-primary-700">{review.jobTitle}</a>
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
			{#each mockReviewsReceived as review}
				<Card>
					<div class="flex items-start gap-4">
						<Avatar name={review.reviewerName} size="md" />
						<div class="flex-1 min-w-0">
							<div class="flex items-start justify-between gap-2">
								<div>
									<h3 class="font-medium text-gray-900 dark:text-white">{review.reviewerName}</h3>
									<a href="#" class="text-xs text-primary-600 hover:text-primary-700">{review.jobTitle}</a>
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
