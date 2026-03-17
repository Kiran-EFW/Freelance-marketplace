<script lang="ts">
	import { AlertTriangle, Search, Eye, UserCheck, Clock, CheckCircle, XCircle, MessageSquare, Loader2 } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import Pagination from '$lib/components/ui/Pagination.svelte';
	import { toastSuccess, toastError } from '$lib/stores/toast';
	import api from '$lib/api/client';

	let searchQuery = $state('');
	let statusFilter = $state('all');
	let severityFilter = $state('all');
	let currentPage = $state(1);
	let loading = $state(true);
	let error = $state('');
	let disputesList = $state<any[]>([]);
	let totalPages = $state(1);
	let selectedDispute = $state<any | null>(null);
	let showDetailModal = $state(false);
	let showAssignModal = $state(false);
	let selectedMediator = $state('');

	const mediators = ['Admin Priya', 'Admin Raj', 'Admin Sunil', 'Admin Meera'];

	async function fetchDisputes() {
		loading = true;
		error = '';
		try {
			const params: any = {
				page: currentPage,
				per_page: 10
			};
			if (statusFilter !== 'all') params.status = statusFilter;
			if (severityFilter !== 'all') params.severity = severityFilter;

			const res = await api.disputes.list(params);
			disputesList = (res.data || []).map((d: any) => ({
				id: d.id,
				status: d.status || 'open',
				severity: d.severity || 'medium',
				type: d.type || 'other',
				title: d.title || '',
				description: d.description || '',
				customer: { name: d.customer?.name || d.customer_name || '', id: d.customer?.id || d.customer_id || '' },
				provider: { name: d.provider?.name || d.provider_name || '', id: d.provider?.id || d.provider_id || '' },
				jobId: d.job_id || '',
				jobTitle: d.job?.title || d.job_title || '',
				amount: d.amount || 0,
				createdAt: d.created_at?.split('T')[0] || '',
				updatedAt: d.updated_at?.split('T')[0] || '',
				mediator: d.mediator?.name || d.mediator_name || null,
				mediatorId: d.mediator?.id || d.mediator_id || null,
				messages: d.message_count || d.messages || 0,
				resolution: d.resolution || null
			}));
			totalPages = res.meta?.total_pages || Math.ceil((res.meta?.total || 1) / 10);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load disputes';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		const _s = statusFilter;
		const _sv = severityFilter;
		const _p = currentPage;
		fetchDisputes();
	});

	let filteredDisputes = $derived(
		disputesList.filter((d) => {
			const matchesSearch = !searchQuery || d.title.toLowerCase().includes(searchQuery.toLowerCase()) || d.id.toLowerCase().includes(searchQuery.toLowerCase());
			return matchesSearch;
		})
	);

	let openCount = $derived(disputesList.filter((d) => d.status === 'open').length);
	let inReviewCount = $derived(disputesList.filter((d) => d.status === 'in_review').length);

	const severityBadge: Record<string, 'danger' | 'warning' | 'info'> = {
		high: 'danger', medium: 'warning', low: 'info'
	};

	const statusBadge: Record<string, 'danger' | 'warning' | 'info' | 'success' | 'neutral'> = {
		open: 'danger', in_review: 'warning', resolved: 'success', closed: 'neutral'
	};

	function viewDispute(dispute: any) {
		selectedDispute = dispute;
		showDetailModal = true;
	}

	function openAssignModal(dispute: any) {
		selectedDispute = dispute;
		selectedMediator = '';
		showAssignModal = true;
	}

	async function assignMediator() {
		if (!selectedMediator || !selectedDispute) return;
		try {
			await api.disputes.assignMediator(selectedDispute.id, selectedMediator);
			toastSuccess(`Mediator ${selectedMediator} assigned to ${selectedDispute.id}`);
			showAssignModal = false;
			fetchDisputes();
		} catch (err) {
			toastError(err instanceof Error ? err.message : 'Failed to assign mediator');
		}
	}
</script>

<svelte:head>
	<title>Disputes - Admin - Seva</title>
</svelte:head>

<div class="px-6 py-8 lg:px-8">
	<div class="flex flex-wrap items-center justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Disputes</h1>
			<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">Manage and resolve platform disputes.</p>
		</div>
		<div class="flex gap-2">
			<Badge variant="danger">{openCount} open</Badge>
			<Badge variant="warning">{inReviewCount} in review</Badge>
		</div>
	</div>

	<!-- Filters -->
	<div class="mt-6 flex flex-col gap-4 sm:flex-row">
		<div class="relative flex-1">
			<Search class="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-gray-400" />
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Search disputes..."
				class="w-full rounded-lg border border-gray-300 py-2.5 pl-10 pr-4 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
			/>
		</div>
		<select
			bind:value={statusFilter}
			class="rounded-lg border border-gray-300 px-3 py-2.5 text-sm focus:border-primary-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-white"
		>
			<option value="all">All Status</option>
			<option value="open">Open</option>
			<option value="in_review">In Review</option>
			<option value="resolved">Resolved</option>
			<option value="closed">Closed</option>
		</select>
		<select
			bind:value={severityFilter}
			class="rounded-lg border border-gray-300 px-3 py-2.5 text-sm focus:border-primary-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-white"
		>
			<option value="all">All Severity</option>
			<option value="high">High</option>
			<option value="medium">Medium</option>
			<option value="low">Low</option>
		</select>
	</div>

	<!-- Disputes List -->
	{#if loading}
		<div class="mt-8 flex items-center justify-center py-12">
			<Loader2 class="h-8 w-8 animate-spin text-primary-600" />
		</div>
	{:else if error}
		<div class="mt-4 rounded-lg border border-red-200 bg-red-50 p-6 text-center dark:border-red-800 dark:bg-red-900/20">
			<p class="text-red-600 dark:text-red-400">{error}</p>
		</div>
	{:else}
	<div class="mt-6 space-y-4">
		{#each filteredDisputes as dispute}
			<Card>
				<div class="flex items-start gap-4">
					<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full
						{dispute.severity === 'high' ? 'bg-red-100 text-red-500 dark:bg-red-900/20' : dispute.severity === 'medium' ? 'bg-yellow-100 text-yellow-500 dark:bg-yellow-900/20' : 'bg-blue-100 text-blue-500 dark:bg-blue-900/20'}">
						<AlertTriangle class="h-5 w-5" />
					</div>
					<div class="flex-1 min-w-0">
						<div class="flex items-center gap-2">
							<span class="text-xs font-mono text-gray-400">{dispute.id}</span>
							<Badge variant={statusBadge[dispute.status] || 'neutral'} size="sm">{dispute.status.replace('_', ' ')}</Badge>
							<Badge variant={severityBadge[dispute.severity] || 'info'} size="sm">{dispute.severity}</Badge>
						</div>
						<h3 class="mt-1 font-medium text-gray-900 dark:text-white">{dispute.title}</h3>
						<p class="mt-0.5 text-sm text-gray-500 dark:text-gray-400 line-clamp-1">{dispute.description}</p>
						<div class="mt-2 flex flex-wrap items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
							<span>Customer: <strong class="text-gray-700 dark:text-gray-300">{dispute.customer.name}</strong></span>
							<span>Provider: <strong class="text-gray-700 dark:text-gray-300">{dispute.provider.name}</strong></span>
							<span>Rs. {dispute.amount.toLocaleString()}</span>
							<span class="flex items-center gap-1">
								<MessageSquare class="h-3 w-3" />
								{dispute.messages} messages
							</span>
							{#if dispute.mediator}
								<span class="flex items-center gap-1">
									<UserCheck class="h-3 w-3" />
									{dispute.mediator}
								</span>
							{/if}
						</div>
					</div>
					<div class="flex flex-col gap-2">
						<Button variant="outline" size="sm" onclick={() => viewDispute(dispute)}>
							<Eye class="h-4 w-4" />
							View
						</Button>
						{#if dispute.status === 'open' && !dispute.mediator}
							<Button variant="primary" size="sm" onclick={() => openAssignModal(dispute)}>
								<UserCheck class="h-4 w-4" />
								Assign
							</Button>
						{/if}
					</div>
				</div>
			</Card>
		{/each}

		{#if filteredDisputes.length === 0}
			<Card>
				<div class="py-8 text-center">
					<CheckCircle class="mx-auto h-10 w-10 text-gray-300 dark:text-gray-600" />
					<p class="mt-3 text-sm text-gray-500 dark:text-gray-400">No disputes found matching your filters.</p>
				</div>
			</Card>
		{/if}
	</div>

	<div class="mt-6">
		<Pagination {currentPage} {totalPages} onPageChange={(p) => (currentPage = p)} />
	</div>
	{/if}
</div>

<!-- Dispute Detail Modal -->
<Modal bind:open={showDetailModal} title="Dispute Details" size="lg">
	{#if selectedDispute}
		<div class="space-y-6">
			<div class="flex items-center gap-3">
				<span class="text-sm font-mono text-gray-400">{selectedDispute.id}</span>
				<Badge variant={statusBadge[selectedDispute.status] || 'neutral'} size="sm">{selectedDispute.status.replace('_', ' ')}</Badge>
				<Badge variant={severityBadge[selectedDispute.severity] || 'info'} size="sm">{selectedDispute.severity} severity</Badge>
			</div>

			<div>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white">{selectedDispute.title}</h3>
				<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">{selectedDispute.description}</p>
			</div>

			<div class="grid grid-cols-2 gap-4">
				<div class="rounded-lg border border-gray-200 p-3 dark:border-gray-700">
					<p class="text-xs text-gray-500 dark:text-gray-400">Customer</p>
					<div class="mt-1 flex items-center gap-2">
						<Avatar name={selectedDispute.customer.name} size="sm" />
						<p class="text-sm font-medium text-gray-900 dark:text-white">{selectedDispute.customer.name}</p>
					</div>
				</div>
				<div class="rounded-lg border border-gray-200 p-3 dark:border-gray-700">
					<p class="text-xs text-gray-500 dark:text-gray-400">Provider</p>
					<div class="mt-1 flex items-center gap-2">
						<Avatar name={selectedDispute.provider.name} size="sm" />
						<p class="text-sm font-medium text-gray-900 dark:text-white">{selectedDispute.provider.name}</p>
					</div>
				</div>
			</div>

			<div class="grid grid-cols-3 gap-4">
				<div>
					<p class="text-xs text-gray-500 dark:text-gray-400">Job</p>
					<p class="mt-1 text-sm text-gray-900 dark:text-white">{selectedDispute.jobTitle}</p>
				</div>
				<div>
					<p class="text-xs text-gray-500 dark:text-gray-400">Amount</p>
					<p class="mt-1 text-sm text-gray-900 dark:text-white">Rs. {selectedDispute.amount.toLocaleString()}</p>
				</div>
				<div>
					<p class="text-xs text-gray-500 dark:text-gray-400">Created</p>
					<p class="mt-1 text-sm text-gray-900 dark:text-white">{selectedDispute.createdAt}</p>
				</div>
			</div>

			{#if selectedDispute.mediator}
				<div>
					<p class="text-xs text-gray-500 dark:text-gray-400">Assigned Mediator</p>
					<p class="mt-1 text-sm font-medium text-gray-900 dark:text-white">{selectedDispute.mediator}</p>
				</div>
			{/if}

			{#if 'resolution' in selectedDispute && selectedDispute.resolution}
				<div class="rounded-lg bg-green-50 p-4 dark:bg-green-900/10">
					<h4 class="text-sm font-medium text-green-800 dark:text-green-400">Resolution</h4>
					<p class="mt-1 text-sm text-green-700 dark:text-green-300">{selectedDispute.resolution}</p>
				</div>
			{/if}
		</div>
	{/if}
	{#snippet footer()}
		<Button variant="outline" onclick={() => (showDetailModal = false)}>Close</Button>
		{#if selectedDispute && selectedDispute.status === 'open'}
			<Button variant="primary" onclick={() => { showDetailModal = false; openAssignModal(selectedDispute!); }}>
				<UserCheck class="h-4 w-4" />
				Assign Mediator
			</Button>
		{/if}
	{/snippet}
</Modal>

<!-- Assign Mediator Modal -->
<Modal bind:open={showAssignModal} title="Assign Mediator" size="sm">
	<div class="space-y-4">
		<p class="text-sm text-gray-600 dark:text-gray-400">
			Assign a mediator to dispute <strong>{selectedDispute?.id}</strong>.
		</p>
		<div>
			<label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">Select Mediator</label>
			<select
				bind:value={selectedMediator}
				class="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm focus:border-primary-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-white"
			>
				<option value="">Choose a mediator...</option>
				{#each mediators as mediator}
					<option value={mediator}>{mediator}</option>
				{/each}
			</select>
		</div>
	</div>
	{#snippet footer()}
		<Button variant="outline" onclick={() => (showAssignModal = false)}>Cancel</Button>
		<Button variant="primary" onclick={assignMediator} disabled={!selectedMediator}>Assign</Button>
	{/snippet}
</Modal>
