<script lang="ts">
	import { Shield, CheckCircle, XCircle, Eye, FileText, Image, Calendar, User, ChevronDown } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import { toastSuccess, toastError } from '$lib/stores/toast';

	let filterStatus = $state('pending');
	let selectedApplication = $state<typeof kycApplications[0] | null>(null);
	let showDetailModal = $state(false);
	let rejectReason = $state('');
	let showRejectModal = $state(false);

	const kycApplications = [
		{
			id: '1', name: 'Vikram Singh', category: 'Plumbing', phone: '+91 98765 11111',
			submittedAt: '2026-03-16', status: 'pending',
			documents: [
				{ type: 'Aadhaar Card', name: 'aadhaar_front.jpg', status: 'uploaded' },
				{ type: 'Aadhaar Card (Back)', name: 'aadhaar_back.jpg', status: 'uploaded' },
				{ type: 'Skill Certificate', name: 'plumbing_cert.pdf', status: 'uploaded' }
			],
			experience: '8 years', bio: 'Experienced plumber specializing in residential plumbing repairs and installations.',
			serviceArea: 'Koramangala, HSR Layout, BTM Layout'
		},
		{
			id: '2', name: 'Fatima Begum', category: 'Cleaning', phone: '+91 98765 22222',
			submittedAt: '2026-03-15', status: 'pending',
			documents: [
				{ type: 'Aadhaar Card', name: 'aadhaar_fatima.jpg', status: 'uploaded' },
				{ type: 'Aadhaar Card (Back)', name: 'aadhaar_fatima_back.jpg', status: 'uploaded' },
				{ type: 'Police Verification', name: 'police_verify.pdf', status: 'uploaded' },
				{ type: 'Training Certificate', name: 'cleaning_training.pdf', status: 'uploaded' }
			],
			experience: '5 years', bio: 'Professional deep cleaning and sanitization services for homes and offices.',
			serviceArea: 'Indiranagar, Whitefield, Marathahalli'
		},
		{
			id: '3', name: 'Ravi Shankar', category: 'Electrical', phone: '+91 98765 33333',
			submittedAt: '2026-03-15', status: 'pending',
			documents: [
				{ type: 'Aadhaar Card', name: 'aadhaar_ravi.jpg', status: 'uploaded' },
				{ type: 'Electrician License', name: 'elec_license.pdf', status: 'uploaded' }
			],
			experience: '12 years', bio: 'Licensed electrician with expertise in wiring, panel installation, and troubleshooting.',
			serviceArea: 'JP Nagar, Jayanagar, Bannerghatta'
		},
		{
			id: '4', name: 'Meena Devi', category: 'Gardening', phone: '+91 98765 44444',
			submittedAt: '2026-03-14', status: 'pending',
			documents: [
				{ type: 'Aadhaar Card', name: 'aadhaar_meena.jpg', status: 'uploaded' },
				{ type: 'Aadhaar Card (Back)', name: 'aadhaar_meena_back.jpg', status: 'uploaded' },
				{ type: 'Horticulture Certificate', name: 'horticulture.pdf', status: 'uploaded' }
			],
			experience: '6 years', bio: 'Gardening and landscaping expert. Tree trimming, lawn maintenance, and plant care.',
			serviceArea: 'Sarjapur Road, Electronic City, Bommanahalli'
		},
		{
			id: '5', name: 'Mohan Rao', category: 'Electrical', phone: '+91 98765 55555',
			submittedAt: '2026-03-10', status: 'approved',
			documents: [
				{ type: 'Aadhaar Card', name: 'aadhaar_mohan.jpg', status: 'verified' },
				{ type: 'Electrician License', name: 'elec_mohan.pdf', status: 'verified' },
				{ type: 'Experience Letter', name: 'exp_letter.pdf', status: 'verified' }
			],
			experience: '10 years', bio: 'Certified electrician for residential and commercial projects.',
			serviceArea: 'Koramangala, Indiranagar'
		},
		{
			id: '6', name: 'Sanjay Patel', category: 'Plumbing', phone: '+91 98765 66666',
			submittedAt: '2026-03-08', status: 'rejected',
			documents: [
				{ type: 'Aadhaar Card', name: 'aadhaar_sanjay.jpg', status: 'rejected' },
				{ type: 'Certificate', name: 'cert_blurry.jpg', status: 'rejected' }
			],
			experience: '3 years', bio: 'Plumber for home repairs.',
			serviceArea: 'BTM Layout',
			rejectionReason: 'Documents are blurry and unreadable. Please re-upload clear copies.'
		}
	];

	let filteredApplications = $derived(
		kycApplications.filter((a) => filterStatus === 'all' || a.status === filterStatus)
	);

	function viewApplication(app: typeof kycApplications[0]) {
		selectedApplication = app;
		showDetailModal = true;
	}

	function approveApplication() {
		toastSuccess(`KYC approved for ${selectedApplication?.name}`);
		showDetailModal = false;
	}

	function openRejectModal() {
		showDetailModal = false;
		showRejectModal = true;
		rejectReason = '';
	}

	function rejectApplication() {
		if (!rejectReason.trim()) return;
		toastError(`KYC rejected for ${selectedApplication?.name}`);
		showRejectModal = false;
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
		<Badge variant="warning">{kycApplications.filter((a) => a.status === 'pending').length} pending reviews</Badge>
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
			<label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">Reason</label>
			<textarea
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
