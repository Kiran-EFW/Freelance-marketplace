<script lang="ts">
	import { AlertTriangle, Clock, CheckCircle, ArrowRight } from 'lucide-svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import type { DisputeStatus } from '$lib/types';

	const mockDisputes = [
		{ id: '1', jobTitle: 'Painting bedroom', type: 'quality', status: 'open' as DisputeStatus, date: '2026-03-12', description: 'The paint job was uneven and patches were left uncovered.' },
		{ id: '2', jobTitle: 'AC repair', type: 'no_show', status: 'resolved' as DisputeStatus, date: '2026-02-20', description: 'Provider did not show up for the scheduled appointment.' },
		{ id: '3', jobTitle: 'Garden maintenance', type: 'payment', status: 'under_review' as DisputeStatus, date: '2026-01-15', description: 'Was charged more than the agreed quote amount.' }
	];

	const statusVariant: Record<DisputeStatus, 'success' | 'warning' | 'danger' | 'info' | 'neutral'> = {
		open: 'danger', under_review: 'warning', resolved: 'success', escalated: 'danger', closed: 'neutral'
	};

	const statusLabel: Record<DisputeStatus, string> = {
		open: 'Open', under_review: 'Under Review', resolved: 'Resolved', escalated: 'Escalated', closed: 'Closed'
	};
</script>

<svelte:head>
	<title>Disputes - Seva</title>
</svelte:head>

<div class="mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8">
	<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Disputes</h1>
	<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">Track and manage your service disputes.</p>

	<div class="mt-6 space-y-4">
		{#each mockDisputes as dispute}
			<a href="/disputes/{dispute.id}" class="block">
				<Card hover>
					<div class="flex items-start justify-between gap-3">
						<div class="flex items-start gap-3">
							<div class="mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-red-100 text-red-600 dark:bg-red-900/20 dark:text-red-400">
								<AlertTriangle class="h-5 w-5" />
							</div>
							<div>
								<div class="flex items-center gap-2">
									<h3 class="font-medium text-gray-900 dark:text-white">{dispute.jobTitle}</h3>
									<Badge variant={statusVariant[dispute.status]} size="sm">{statusLabel[dispute.status]}</Badge>
								</div>
								<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{dispute.description}</p>
								<div class="mt-2 flex items-center gap-3 text-xs text-gray-400">
									<span>Type: {dispute.type}</span>
									<span>Filed: {dispute.date}</span>
								</div>
							</div>
						</div>
						<ArrowRight class="h-5 w-5 shrink-0 text-gray-400" />
					</div>
				</Card>
			</a>
		{/each}

		{#if mockDisputes.length === 0}
			<Card>
				<div class="py-8 text-center">
					<CheckCircle class="mx-auto h-12 w-12 text-gray-300 dark:text-gray-600" />
					<h3 class="mt-4 text-lg font-medium text-gray-900 dark:text-white">No disputes</h3>
					<p class="mt-2 text-sm text-gray-500 dark:text-gray-400">You have no active disputes. Great!</p>
				</div>
			</Card>
		{/if}
	</div>
</div>
