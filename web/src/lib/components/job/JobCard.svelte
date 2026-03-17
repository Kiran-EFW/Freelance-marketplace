<script lang="ts">
	import { MapPin, Clock, IndianRupee, MessageSquare } from 'lucide-svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import type { Job, JobStatus } from '$lib/types';

	interface Props {
		job: Job;
		class?: string;
	}

	let { job, class: className = '' }: Props = $props();

	const statusVariant = $derived<Record<JobStatus, 'success' | 'warning' | 'danger' | 'info' | 'neutral'>>({
		draft: 'neutral',
		open: 'info',
		quoted: 'warning',
		accepted: 'info',
		in_progress: 'warning',
		completed: 'success',
		cancelled: 'danger',
		disputed: 'danger'
	});

	const statusLabel = $derived<Record<JobStatus, string>>({
		draft: 'Draft',
		open: 'Open',
		quoted: 'Quoted',
		accepted: 'Accepted',
		in_progress: 'In Progress',
		completed: 'Completed',
		cancelled: 'Cancelled',
		disputed: 'Disputed'
	});

	const budgetDisplay = $derived(() => {
		if (job.budget_min && job.budget_max) {
			return `Rs. ${job.budget_min} - Rs. ${job.budget_max}`;
		}
		if (job.budget_min) return `From Rs. ${job.budget_min}`;
		if (job.budget_max) return `Up to Rs. ${job.budget_max}`;
		return 'Open budget';
	});

	const timeAgo = $derived(() => {
		const diff = Date.now() - new Date(job.created_at).getTime();
		const hours = Math.floor(diff / (1000 * 60 * 60));
		if (hours < 1) return 'Just now';
		if (hours < 24) return `${hours}h ago`;
		const days = Math.floor(hours / 24);
		if (days === 1) return 'Yesterday';
		return `${days}d ago`;
	});
</script>

<a
	href="/jobs/{job.id}"
	class="block rounded-xl border border-gray-200 bg-white p-5 transition-shadow hover:shadow-md dark:border-gray-700 dark:bg-gray-800 {className}"
>
	<div class="flex items-start justify-between gap-3">
		<div class="min-w-0 flex-1">
			<div class="flex items-center gap-2">
				<Badge variant={statusVariant[job.status]} size="sm">{statusLabel[job.status]}</Badge>
				{#if job.category}
					<span class="text-xs text-gray-500 dark:text-gray-400">{job.category.name}</span>
				{/if}
			</div>
			<h3 class="mt-1.5 font-semibold text-gray-900 dark:text-white truncate">{job.title}</h3>
			<p class="mt-1 text-sm text-gray-500 dark:text-gray-400 line-clamp-2">{job.description}</p>
		</div>
	</div>

	<div class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1.5 text-sm text-gray-500 dark:text-gray-400">
		{#if job.location?.postcode}
			<div class="flex items-center gap-1">
				<MapPin class="h-3.5 w-3.5" />
				<span>{job.location.postcode}</span>
			</div>
		{/if}
		<div class="flex items-center gap-1">
			<Clock class="h-3.5 w-3.5" />
			<span>{timeAgo()}</span>
		</div>
		<div class="flex items-center gap-1">
			<IndianRupee class="h-3.5 w-3.5" />
			<span>{budgetDisplay()}</span>
		</div>
		{#if job.quotes_count > 0}
			<div class="flex items-center gap-1">
				<MessageSquare class="h-3.5 w-3.5" />
				<span>{job.quotes_count} quote{job.quotes_count > 1 ? 's' : ''}</span>
			</div>
		{/if}
	</div>

	{#if job.images && job.images.length > 0}
		<div class="mt-3 flex gap-1.5">
			{#each job.images.slice(0, 3) as image, i}
				<div class="h-16 w-16 overflow-hidden rounded-lg bg-gray-100 dark:bg-gray-700">
					<img src={image} alt="Job photo {i + 1}" class="h-full w-full object-cover" />
				</div>
			{/each}
			{#if job.images.length > 3}
				<div class="flex h-16 w-16 items-center justify-center rounded-lg bg-gray-100 text-sm text-gray-500 dark:bg-gray-700 dark:text-gray-400">
					+{job.images.length - 3}
				</div>
			{/if}
		</div>
	{/if}
</a>
