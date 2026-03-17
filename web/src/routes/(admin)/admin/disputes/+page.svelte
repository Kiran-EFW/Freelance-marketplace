<script lang="ts">
	import { AlertTriangle, Search, Eye, UserCheck, Clock, CheckCircle, XCircle, MessageSquare } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import Pagination from '$lib/components/ui/Pagination.svelte';
	import { toastSuccess } from '$lib/stores/toast';

	let searchQuery = $state('');
	let statusFilter = $state('all');
	let severityFilter = $state('all');
	let currentPage = $state(1);
	let selectedDispute = $state<typeof disputes[0] | null>(null);
	let showDetailModal = $state(false);
	let showAssignModal = $state(false);
	let selectedMediator = $state('');

	const disputes = [
		{
			id: 'D-1045', status: 'open', severity: 'high', type: 'quality',
			title: 'Incomplete work - kitchen plumbing',
			description: 'The provider left without finishing the sink installation. Water is still leaking from under the basin.',
			customer: { name: 'Amit Verma', id: 'c1' },
			provider: { name: 'Suresh Nair', id: 'p1' },
			jobId: 'J-2045', jobTitle: 'Kitchen plumbing repair',
			amount: 3500, createdAt: '2026-03-16', updatedAt: '2026-03-16',
			mediator: null, messages: 3
		},
		{
			id: 'D-1044', status: 'open', severity: 'high', type: 'payment',
			title: 'Payment not received for completed job',
			description: 'I completed the electrical work 5 days ago but have not received payment despite the customer confirming completion.',
			customer: { name: 'Priya Menon', id: 'c2' },
			provider: { name: 'Deepak Kumar', id: 'p2' },
			jobId: 'J-2038', jobTitle: 'Electrical panel upgrade',
			amount: 8500, createdAt: '2026-03-15', updatedAt: '2026-03-16',
			mediator: null, messages: 5
		},
		{
			id: 'D-1043', status: 'in_review', severity: 'medium', type: 'behavior',
			title: 'Late arrival and unprofessional conduct',
			description: 'The provider arrived 2 hours late without notice and was dismissive when asked about the delay.',
			customer: { name: 'Arjun Das', id: 'c3' },
			provider: { name: 'Lakshmi Bai', id: 'p3' },
			jobId: 'J-2030', jobTitle: 'Home deep cleaning',
			amount: 2000, createdAt: '2026-03-14', updatedAt: '2026-03-15',
			mediator: 'Admin Priya', messages: 8
		},
		{
			id: 'D-1042', status: 'resolved', severity: 'medium', type: 'quality',
			title: 'Paint quality below expectations',
			description: 'The paint used was of lower quality than agreed upon. Visible streaks and uneven coverage.',
			customer: { name: 'Meera Reddy', id: 'c4' },
			provider: { name: 'Kiran Rao', id: 'p4' },
			jobId: 'J-2025', jobTitle: 'Living room painting',
			amount: 5000, createdAt: '2026-03-10', updatedAt: '2026-03-14',
			mediator: 'Admin Raj', messages: 12, resolution: 'Partial refund of Rs. 2,000 issued to customer. Provider agreed to redo the work.'
		},
		{
			id: 'D-1041', status: 'resolved', severity: 'low', type: 'scheduling',
			title: 'Rescheduling conflict',
			description: 'Provider rescheduled three times without valid reason.',
			customer: { name: 'Anita Gupta', id: 'c5' },
			provider: { name: 'Ravi Shankar', id: 'p5' },
			jobId: 'J-2020', jobTitle: 'AC installation',
			amount: 4500, createdAt: '2026-03-08', updatedAt: '2026-03-12',
			mediator: 'Admin Priya', messages: 6, resolution: 'Provider warned. Job reassigned to different provider.'
		},
		{
			id: 'D-1040', status: 'closed', severity: 'low', type: 'other',
			title: 'Miscommunication about service scope',
			description: 'Customer expected additional services not mentioned in the original job description.',
			customer: { name: 'Rajesh Kumar', id: 'c6' },
			provider: { name: 'Mohan Rao', id: 'p6' },
			jobId: 'J-2015', jobTitle: 'Electrical wiring',
			amount: 3000, createdAt: '2026-03-05', updatedAt: '2026-03-09',
			mediator: 'Admin Raj', messages: 4, resolution: 'Misunderstanding clarified. No action taken.'
		}
	];

	const mediators = ['Admin Priya', 'Admin Raj', 'Admin Sunil', 'Admin Meera'];

	let filteredDisputes = $derived(
		disputes.filter((d) => {
			const matchesSearch = !searchQuery || d.title.toLowerCase().includes(searchQuery.toLowerCase()) || d.id.toLowerCase().includes(searchQuery.toLowerCase());
			const matchesStatus = statusFilter === 'all' || d.status === statusFilter;
			const matchesSeverity = severityFilter === 'all' || d.severity === severityFilter;
			return matchesSearch && matchesStatus && matchesSeverity;
		})
	);

	const severityBadge: Record<string, 'danger' | 'warning' | 'info'> = {
		high: 'danger', medium: 'warning', low: 'info'
	};

	const statusBadge: Record<string, 'danger' | 'warning' | 'info' | 'success' | 'neutral'> = {
		open: 'danger', in_review: 'warning', resolved: 'success', closed: 'neutral'
	};

	function viewDispute(dispute: typeof disputes[0]) {
		selectedDispute = dispute;
		showDetailModal = true;
	}

	function openAssignModal(dispute: typeof disputes[0]) {
		selectedDispute = dispute;
		selectedMediator = '';
		showAssignModal = true;
	}

	function assignMediator() {
		if (!selectedMediator) return;
		toastSuccess(`Mediator ${selectedMediator} assigned to ${selectedDispute?.id}`);
		showAssignModal = false;
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
			<Badge variant="danger">{disputes.filter((d) => d.status === 'open').length} open</Badge>
			<Badge variant="warning">{disputes.filter((d) => d.status === 'in_review').length} in review</Badge>
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
		<Pagination {currentPage} totalPages={2} onPageChange={(p) => (currentPage = p)} />
	</div>
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
