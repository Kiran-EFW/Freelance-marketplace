<script lang="ts">
	import { Shield, CheckCircle, XCircle, Eye, FileText, Image, Calendar, User, ChevronDown, Loader2 } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import { toastSuccess, toastError } from '$lib/stores/toast';
	import api from '$lib/api/client';

	let filterStatus = $state('pending');
	let loading = $state(true);
	let error = $state('');
	let kycApplications = $state<any[]>([]);
	let selectedApplication = $state<any | null>(null);
	let showDetailModal = $state(false);
	let rejectReason = $state('');
	let showRejectModal = $state(false);

	async function fetchKYC() {
		loading = true;
		error = '';
		try {
			const res = await api.admin.pendingKYC({ per_page: 50 });
			kycApplications = (res.data || []).map((app: any) => ({
				id: app.id,
				name: app.name || app.provider_name || '',
				category: app.category || '',
				phone: app.phone || '',
				submittedAt: app.submitted_at?.split('T')[0] || app.created_at?.split('T')[0] || '',
				status: app.status || 'pending',
				documents: (app.documents || []).map((doc: any) => ({
					type: doc.type || doc.document_type || '',
					name: doc.name || doc.file_name || '',
					status: doc.status || 'uploaded'
				})),
				experience: app.experience || '',
				bio: app.bio || '',
				serviceArea: app.service_area || app.serviceArea || '',
				rejectionReason: app.rejection_reason || app.rejectionReason || null
			}));
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load KYC applications';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		const _f = filterStatus;
		fetchKYC();
	});

	let filteredApplications = $derived(
		kycApplications.filter((a) => filterStatus === 'all' || a.status === filterStatus)
	);

	let pendingCount = $derived(
		kycApplications.filter((a) => a.status === 'pending').length
	);

	function viewApplication(app: any) {
		selectedApplication = app;
		showDetailModal = true;
	}

	async function approveApplication() {
		if (!selectedApplication) return;
		try {
			await api.admin.approveKYC(selectedApplication.id);
			toastSuccess(`KYC approved for ${selectedApplication.name}`);
			showDetailModal = false;
			fetchKYC();
		} catch (err) {
			toastError(err instanceof Error ? err.message : 'Failed to approve KYC');
		}
	}

	function openRejectModal() {
		showDetailModal = false;
		showRejectModal = true;
		rejectReason = '';
	}

	async function rejectApplication() {
		if (!rejectReason.trim() || !selectedApplication) return;
		try {
			await api.admin.rejectKYC(selectedApplication.id, rejectReason);
			toastError(`KYC rejected for ${selectedApplication.name}`);
			showRejectModal = false;
			fetchKYC();
		} catch (err) {
			toastError(err instanceof Error ? err.message : 'Failed to reject KYC');
		}
	}

	const statusColors: Record<string, 'warning' | 'success' | 'danger'> = {
		pending: 'warning', approved: 'success', rejected: 'danger'
	};
</script>

<svelte:head>
	<title>KYC Verification - Admin - Seva</title>
</svelte:head>

<div class="px-6 py-8 lg:px-8">
	<div class="flex flex-wrap items-center justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">KYC Verification</h1>
			<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">Review and verify provider applications.</p>
		</div>
		<Badge variant="warning">{pendingCount} pending reviews</Badge>
	</div>

	<!-- Filter Tabs -->
	<div class="mt-6 flex gap-2">
		{#each [{ id: 'pending', label: 'Pending' }, { id: 'approved', label: 'Approved' }, { id: 'rejected', label: 'Rejected' }, { id: 'all', label: 'All' }] as tab}
			<button
				onclick={() => (filterStatus = tab.id)}
				class="rounded-lg px-4 py-2 text-sm font-medium transition
					{filterStatus === tab.id
						? 'bg-primary-600 text-white'
						: 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600'}"
			>
				{tab.label}
			</button>
		{/each}
	</div>

	<!-- Applications List -->
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
		{#each filteredApplications as app}
			<Card>
				<div class="flex items-center gap-4">
					<Avatar name={app.name} size="md" />
					<div class="flex-1 min-w-0">
						<div class="flex items-center gap-2">
							<h3 class="font-semibold text-gray-900 dark:text-white">{app.name}</h3>
							<Badge variant={statusColors[app.status] || 'neutral'} size="sm">{app.status}</Badge>
						</div>
						<p class="mt-0.5 text-sm text-gray-500 dark:text-gray-400">{app.category} -- {app.experience} experience</p>
						<div class="mt-1 flex flex-wrap items-center gap-3 text-xs text-gray-500 dark:text-gray-400">
							<span class="flex items-center gap-1">
								<FileText class="h-3 w-3" />
								{app.documents.length} documents
							</span>
							<span class="flex items-center gap-1">
								<Calendar class="h-3 w-3" />
								{app.submittedAt}
							</span>
						</div>
						{#if app.status === 'rejected' && app.rejectionReason}
							<p class="mt-2 text-xs text-red-600 dark:text-red-400">Reason: {app.rejectionReason}</p>
						{/if}
					</div>
					<div class="flex gap-2">
						<Button variant="outline" size="sm" onclick={() => viewApplication(app)}>
							<Eye class="h-4 w-4" />
							Review
						</Button>
						{#if app.status === 'pending'}
							<Button variant="primary" size="sm" onclick={() => { selectedApplication = app; approveApplication(); }}>
								<CheckCircle class="h-4 w-4" />
								Approve
							</Button>
						{/if}
					</div>
				</div>
			</Card>
		{/each}

		{#if filteredApplications.length === 0}
			<Card>
				<div class="py-8 text-center">
					<Shield class="mx-auto h-10 w-10 text-gray-300 dark:text-gray-600" />
					<p class="mt-3 text-sm text-gray-500 dark:text-gray-400">No applications in this category.</p>
				</div>
			</Card>
		{/if}
	</div>
	{/if}
</div>

<!-- Application Detail Modal -->
<Modal bind:open={showDetailModal} title="KYC Application" size="lg">
	{#if selectedApplication}
		<div class="space-y-6">
			<!-- Applicant Info -->
			<div class="flex items-center gap-4">
				<Avatar name={selectedApplication.name} size="lg" />
				<div>
					<h3 class="text-lg font-semibold text-gray-900 dark:text-white">{selectedApplication.name}</h3>
					<p class="text-sm text-gray-500 dark:text-gray-400">{selectedApplication.category} -- {selectedApplication.experience}</p>
					<p class="text-sm text-gray-500 dark:text-gray-400">{selectedApplication.phone}</p>
				</div>
			</div>

			<!-- Bio -->
			<div>
				<h4 class="text-sm font-medium text-gray-700 dark:text-gray-300">Bio</h4>
				<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{selectedApplication.bio}</p>
			</div>

			<!-- Service Area -->
			<div>
				<h4 class="text-sm font-medium text-gray-700 dark:text-gray-300">Service Area</h4>
				<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{selectedApplication.serviceArea}</p>
			</div>

			<!-- Documents -->
			<div>
				<h4 class="text-sm font-medium text-gray-700 dark:text-gray-300">Documents ({selectedApplication.documents.length})</h4>
				<div class="mt-2 space-y-2">
					{#each selectedApplication.documents as doc}
						<div class="flex items-center gap-3 rounded-lg border border-gray-200 p-3 dark:border-gray-700">
							<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-gray-100 text-gray-500 dark:bg-gray-700">
								{#if doc.name.endsWith('.pdf')}
									<FileText class="h-5 w-5" />
								{:else}
									<Image class="h-5 w-5" />
								{/if}
							</div>
							<div class="flex-1">
								<p class="text-sm font-medium text-gray-900 dark:text-white">{doc.type}</p>
								<p class="text-xs text-gray-500 dark:text-gray-400">{doc.name}</p>
							</div>
							<Badge variant={doc.status === 'verified' ? 'success' : doc.status === 'rejected' ? 'danger' : 'info'} size="sm">
								{doc.status}
							</Badge>
						</div>
					{/each}
				</div>
			</div>
		</div>
	{/if}
	{#snippet footer()}
		<Button variant="outline" onclick={() => (showDetailModal = false)}>Close</Button>
		{#if selectedApplication?.status === 'pending'}
			<Button variant="danger" onclick={openRejectModal}>
				<XCircle class="h-4 w-4" />
				Reject
			</Button>
			<Button variant="primary" onclick={approveApplication}>
				<CheckCircle class="h-4 w-4" />
				Approve
			</Button>
		{/if}
	{/snippet}
</Modal>

<!-- Reject Modal -->
<Modal bind:open={showRejectModal} title="Reject Application" size="sm">
	<div class="space-y-4">
		<p class="text-sm text-gray-600 dark:text-gray-400">
			Please provide a reason for rejecting {selectedApplication?.name}'s application. This will be shared with the applicant.
		</p>
		<div>
			<label for="reject-reason" class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">Reason</label>
			<textarea
				id="reject-reason"
				bind:value={rejectReason}
				rows="3"
				placeholder="Explain why the application is being rejected..."
				class="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
			></textarea>
		</div>
	</div>
	{#snippet footer()}
		<Button variant="outline" onclick={() => (showRejectModal = false)}>Cancel</Button>
		<Button variant="danger" onclick={rejectApplication} disabled={!rejectReason.trim()}>Reject Application</Button>
	{/snippet}
</Modal>
