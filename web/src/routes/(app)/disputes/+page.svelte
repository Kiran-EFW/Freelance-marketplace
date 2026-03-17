<script lang="ts">
	import { onMount } from 'svelte';
	import { AlertTriangle, CheckCircle, ArrowRight, Loader2 } from 'lucide-svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import api from '$lib/api/client';
	import type { DisputeStatus } from '$lib/types';

	let loading = $state(true);
	let error = $state('');
	let disputesList = $state<any[]>([]);

	onMount(async () => {
		try {
			const res = await api.disputes.list({ per_page: 20 });
			disputesList = (res.data || []).map((d: any) => ({
				id: d.id,
				jobTitle: d.job?.title || d.title || 'Dispute',
				type: d.type || d.dispute_type || 'other',
				status: d.status as DisputeStatus,
				date: d.created_at?.split('T')[0] || '',
				description: d.description || ''
			}));
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load disputes';
		} finally {
			loading = false;
		}
	});

	const statusVariant: Record<string, 'success' | 'warning' | 'danger' | 'info' | 'neutral'> = {
		open: 'danger', under_review: 'warning', resolved: 'success', escalated: 'danger', closed: 'neutral'
	};

	const statusLabel: Record<string, string> = {
		open: 'Open', under_review: 'Under Review', resolved: 'Resolved', escalated: 'Escalated', closed: 'Closed'
	};
</script>

<svelte:head>
	<title>Disputes - Seva</title>
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
	<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Disputes</h1>
	<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">Track and manage your service disputes.</p>

	<div class="mt-6 space-y-4">
		{#each disputesList as dispute}
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
									<Badge variant={statusVariant[dispute.status] || 'neutral'} size="sm">{statusLabel[dispute.status] || dispute.status}</Badge>
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

		{#if disputesList.length === 0}
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
{/if}
