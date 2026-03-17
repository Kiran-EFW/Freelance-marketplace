<script lang="ts">
	import { onMount } from 'svelte';
	import { Building2, Users, ClipboardList, Plus, Filter, Clock, CheckCircle, AlertCircle, Loader2, Trash2, UserPlus } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import api from '$lib/api/client';
	import { t } from '$lib/i18n/index.svelte';

	let loading = $state(true);
	let error = $state('');
	let activeTab = $state<'requests' | 'members'>('requests');

	// Organization data
	let orgId = $state('');
	let org = $state<any>(null);
	let stats = $state({ total_requests: 0, pending_requests: 0, completed_requests: 0, in_progress_requests: 0, assigned_requests: 0, active_members: 0 });
	let requests = $state<any[]>([]);
	let members = $state<any[]>([]);

	// Filters
	let statusFilter = $state('');
	let priorityFilter = $state('');

	// Modals
	let showCreateRequestModal = $state(false);
	let showAddMemberModal = $state(false);
	let showCreateOrgModal = $state(false);

	// Create request form
	let newRequest = $state({ category_id: '', title: '', description: '', priority: 'medium', scheduled_at: '', notes: '' });
	let creatingRequest = $state(false);

	// Add member form
	let newMember = $state({ user_id: '', role: 'member' });
	let addingMember = $state(false);

	// Create org form
	let newOrg = $state({ name: '', type: 'housing_society', address: '', postcode: '', city: '', state: '', contact_phone: '', contact_email: '' });
	let creatingOrg = $state(false);

	const statusOptions = [
		{ value: '', label: t('organization.all') },
		{ value: 'pending', label: t('organization.pending') },
		{ value: 'assigned', label: t('organization.assigned') },
		{ value: 'in_progress', label: t('organization.in_progress') },
		{ value: 'completed', label: t('organization.completed') },
		{ value: 'cancelled', label: t('organization.cancelled') }
	];

	const priorityOptions = [
		{ value: '', label: t('organization.all') },
		{ value: 'low', label: t('organization.low') },
		{ value: 'medium', label: t('organization.medium') },
		{ value: 'high', label: t('organization.high') },
		{ value: 'urgent', label: t('organization.urgent') }
	];

	const orgTypeOptions = [
		{ value: 'housing_society', label: t('organization.housing_society') },
		{ value: 'company', label: t('organization.company') },
		{ value: 'institution', label: t('organization.institution') }
	];

	const roleOptions = [
		{ value: 'member', label: t('organization.member') },
		{ value: 'manager', label: t('organization.manager') },
		{ value: 'admin', label: t('organization.admin') }
	];

	const priorityBadge: Record<string, 'warning' | 'info' | 'success' | 'neutral' | 'danger'> = {
		low: 'neutral',
		medium: 'info',
		high: 'warning',
		urgent: 'danger'
	};

	const statusBadge: Record<string, 'warning' | 'info' | 'success' | 'neutral' | 'danger'> = {
		pending: 'warning',
		assigned: 'info',
		in_progress: 'warning',
		completed: 'success',
		cancelled: 'danger'
	};

	const roleBadge: Record<string, 'warning' | 'info' | 'success' | 'neutral' | 'danger'> = {
		admin: 'danger',
		manager: 'warning',
		member: 'neutral'
	};

	onMount(async () => {
		// For now, try to load organization data (user may need to create one first)
		try {
			// Try to get the user's organization from localStorage or default
			const savedOrgId = typeof window !== 'undefined' ? localStorage.getItem('seva-org-id') : null;
			if (savedOrgId) {
				orgId = savedOrgId;
				await loadOrgData();
			} else {
				loading = false;
				showCreateOrgModal = true;
			}
		} catch (err) {
			loading = false;
			showCreateOrgModal = true;
		}
	});

	async function loadOrgData() {
		loading = true;
		error = '';
		try {
			const [orgRes, statsRes, requestsRes, membersRes] = await Promise.all([
				api.organizations.get(orgId),
				api.organizations.getStats(orgId),
				api.organizations.listServiceRequests(orgId, {
					status: statusFilter || undefined,
					priority: priorityFilter || undefined,
					limit: 20
				}),
				api.organizations.listMembers(orgId, { limit: 50 })
			]);

			org = orgRes.data;
			stats = statsRes.data;
			requests = requestsRes.data || [];
			members = membersRes.data || [];
		} catch (err) {
			error = err instanceof Error ? err.message : t('organization.error');
		} finally {
			loading = false;
		}
	}

	async function handleCreateOrg() {
		creatingOrg = true;
		try {
			const res = await api.organizations.create(newOrg);
			orgId = res.data.id;
			org = res.data;
			if (typeof window !== 'undefined') {
				localStorage.setItem('seva-org-id', orgId);
			}
			showCreateOrgModal = false;
			await loadOrgData();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create organization';
		} finally {
			creatingOrg = false;
		}
	}

	async function handleCreateRequest() {
		creatingRequest = true;
		try {
			await api.organizations.createServiceRequest(orgId, {
				category_id: newRequest.category_id,
				title: newRequest.title,
				description: newRequest.description,
				priority: newRequest.priority,
				scheduled_at: newRequest.scheduled_at || undefined,
				notes: newRequest.notes
			});
			showCreateRequestModal = false;
			newRequest = { category_id: '', title: '', description: '', priority: 'medium', scheduled_at: '', notes: '' };
			await loadOrgData();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create request';
		} finally {
			creatingRequest = false;
		}
	}

	async function handleAddMember() {
		addingMember = true;
		try {
			await api.organizations.addMember(orgId, {
				user_id: newMember.user_id,
				role: newMember.role
			});
			showAddMemberModal = false;
			newMember = { user_id: '', role: 'member' };
			await loadOrgData();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to add member';
		} finally {
			addingMember = false;
		}
	}

	async function handleRemoveMember(userId: string) {
		try {
			await api.organizations.removeMember(orgId, userId);
			await loadOrgData();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to remove member';
		}
	}

	async function applyFilters() {
		await loadOrgData();
	}
</script>

<svelte:head>
	<title>{t('organization.title')} - Seva</title>
</svelte:head>

{#if loading}
<div class="flex items-center justify-center py-20">
	<Loader2 class="h-8 w-8 animate-spin text-primary-600" />
</div>
{:else if !org && !showCreateOrgModal}
<div class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
	<div class="rounded-lg border border-gray-200 bg-white p-12 text-center dark:border-gray-700 dark:bg-gray-800">
		<Building2 class="mx-auto h-12 w-12 text-gray-400" />
		<h2 class="mt-4 text-lg font-semibold text-gray-900 dark:text-white">{t('organization.create_organization')}</h2>
		<p class="mt-2 text-sm text-gray-500 dark:text-gray-400">Create an organization to manage service requests for your housing society, company, or institution.</p>
		<Button variant="primary" class="mt-6" onclick={() => (showCreateOrgModal = true)}>
			<Plus class="h-4 w-4" />
			{t('organization.create_organization')}
		</Button>
	</div>
</div>
{:else if org}
<div class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
	<!-- Header -->
	<div class="flex flex-wrap items-center justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">{org.name}</h1>
			<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
				<Badge variant="info" size="sm">{org.type.replace('_', ' ')}</Badge>
				{#if org.city}<span class="ml-2">{org.city}{org.state ? `, ${org.state}` : ''}</span>{/if}
			</p>
		</div>
		<Button variant="primary" onclick={() => (showCreateRequestModal = true)}>
			<Plus class="h-4 w-4" />
			{t('organization.create_request')}
		</Button>
	</div>

	{#if error}
	<div class="mt-4 rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-600 dark:border-red-800 dark:bg-red-900/20 dark:text-red-400">
		{error}
	</div>
	{/if}

	<!-- Stats Grid -->
	<div class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
		<Card>
			<div class="flex items-center gap-3">
				<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-primary-100 text-primary-600 dark:bg-primary-900/30">
					<ClipboardList class="h-5 w-5" />
				</div>
				<span class="text-sm font-medium text-gray-600 dark:text-gray-400">{t('organization.total_requests')}</span>
			</div>
			<p class="mt-2 text-3xl font-bold text-gray-900 dark:text-white">{stats.total_requests}</p>
		</Card>
		<Card>
			<div class="flex items-center gap-3">
				<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-yellow-100 text-yellow-600 dark:bg-yellow-900/30">
					<Clock class="h-5 w-5" />
				</div>
				<span class="text-sm font-medium text-gray-600 dark:text-gray-400">{t('organization.pending_requests')}</span>
			</div>
			<p class="mt-2 text-3xl font-bold text-gray-900 dark:text-white">{stats.pending_requests}</p>
		</Card>
		<Card>
			<div class="flex items-center gap-3">
				<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-green-100 text-green-600 dark:bg-green-900/30">
					<CheckCircle class="h-5 w-5" />
				</div>
				<span class="text-sm font-medium text-gray-600 dark:text-gray-400">{t('organization.completed_requests')}</span>
			</div>
			<p class="mt-2 text-3xl font-bold text-gray-900 dark:text-white">{stats.completed_requests}</p>
		</Card>
		<Card>
			<div class="flex items-center gap-3">
				<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100 text-blue-600 dark:bg-blue-900/30">
					<Users class="h-5 w-5" />
				</div>
				<span class="text-sm font-medium text-gray-600 dark:text-gray-400">{t('organization.active_members')}</span>
			</div>
			<p class="mt-2 text-3xl font-bold text-gray-900 dark:text-white">{stats.active_members}</p>
		</Card>
	</div>

	<!-- Tabs -->
	<div class="mt-8 flex gap-1 border-b border-gray-200 dark:border-gray-700">
		<button
			onclick={() => (activeTab = 'requests')}
			class="flex items-center gap-2 border-b-2 px-4 py-3 text-sm font-medium transition-colors
				{activeTab === 'requests' ? 'border-primary-600 text-primary-600' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400'}"
		>
			<ClipboardList class="h-4 w-4" />
			{t('organization.service_requests')}
		</button>
		<button
			onclick={() => (activeTab = 'members')}
			class="flex items-center gap-2 border-b-2 px-4 py-3 text-sm font-medium transition-colors
				{activeTab === 'members' ? 'border-primary-600 text-primary-600' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400'}"
		>
			<Users class="h-4 w-4" />
			{t('organization.members')}
		</button>
	</div>

	<!-- Service Requests Tab -->
	{#if activeTab === 'requests'}
	<div class="mt-6">
		<!-- Filters -->
		<div class="flex flex-wrap items-end gap-3">
			<Select
				options={statusOptions}
				bind:value={statusFilter}
				label={t('organization.filter_status')}
				onchange={applyFilters}
				class="w-40"
			/>
			<Select
				options={priorityOptions}
				bind:value={priorityFilter}
				label={t('organization.filter_priority')}
				onchange={applyFilters}
				class="w-40"
			/>
		</div>

		<!-- Request List -->
		<div class="mt-4 space-y-3">
			{#if requests.length === 0}
			<Card>
				<div class="py-8 text-center">
					<ClipboardList class="mx-auto h-10 w-10 text-gray-400" />
					<p class="mt-3 text-sm text-gray-500 dark:text-gray-400">{t('organization.no_requests')}</p>
					<Button variant="primary" class="mt-4" onclick={() => (showCreateRequestModal = true)}>
						<Plus class="h-4 w-4" />
						{t('organization.create_request')}
					</Button>
				</div>
			</Card>
			{:else}
				{#each requests as req}
				<Card>
					<div class="flex flex-wrap items-start justify-between gap-3">
						<div class="min-w-0 flex-1">
							<div class="flex flex-wrap items-center gap-2">
								<h3 class="text-sm font-semibold text-gray-900 dark:text-white">{req.title}</h3>
								<Badge variant={statusBadge[req.status] || 'neutral'} size="sm">{req.status.replace('_', ' ')}</Badge>
								<Badge variant={priorityBadge[req.priority] || 'neutral'} size="sm">{req.priority}</Badge>
							</div>
							{#if req.description}
							<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{req.description}</p>
							{/if}
							<div class="mt-2 flex flex-wrap gap-4 text-xs text-gray-500 dark:text-gray-400">
								{#if req.requester_name}<span>{t('organization.requested_by')}: {req.requester_name}</span>{/if}
								{#if req.category_slug}<span>{t('organization.category')}: {req.category_slug}</span>{/if}
								{#if req.provider_name}<span>{t('organization.assigned_to')}: {req.provider_name}</span>{:else}<span class="text-yellow-600">{t('organization.unassigned')}</span>{/if}
								<span>{t('organization.created')}: {new Date(req.created_at).toLocaleDateString()}</span>
							</div>
						</div>
					</div>
				</Card>
				{/each}
			{/if}
		</div>
	</div>
	{/if}

	<!-- Members Tab -->
	{#if activeTab === 'members'}
	<div class="mt-6">
		<div class="flex justify-end">
			<Button variant="primary" onclick={() => (showAddMemberModal = true)}>
				<UserPlus class="h-4 w-4" />
				{t('organization.add_member')}
			</Button>
		</div>

		<div class="mt-4 space-y-3">
			{#if members.length === 0}
			<Card>
				<div class="py-8 text-center">
					<Users class="mx-auto h-10 w-10 text-gray-400" />
					<p class="mt-3 text-sm text-gray-500 dark:text-gray-400">{t('organization.no_members')}</p>
				</div>
			</Card>
			{:else}
				{#each members as member}
				<Card>
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-3">
							<div class="flex h-10 w-10 items-center justify-center rounded-full bg-primary-100 text-primary-600 dark:bg-primary-900/30">
								<Users class="h-5 w-5" />
							</div>
							<div>
								<p class="text-sm font-medium text-gray-900 dark:text-white">{member.user_name || member.user_id}</p>
								<div class="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
									{#if member.user_phone}<span>{member.user_phone}</span>{/if}
									<Badge variant={roleBadge[member.role] || 'neutral'} size="sm">{member.role}</Badge>
								</div>
							</div>
						</div>
						<Button variant="ghost" size="sm" onclick={() => handleRemoveMember(member.user_id)}>
							<Trash2 class="h-4 w-4 text-red-500" />
						</Button>
					</div>
				</Card>
				{/each}
			{/if}
		</div>
	</div>
	{/if}
</div>
{/if}

<!-- Create Organization Modal -->
<Modal bind:open={showCreateOrgModal} title={t('organization.create_organization')} size="lg">
	<div class="space-y-4">
		<Input label={t('organization.org_name')} bind:value={newOrg.name} required placeholder="e.g. Greenfield Apartments" />
		<Select options={orgTypeOptions} bind:value={newOrg.type} label={t('organization.org_type')} />
		<Input label={t('organization.address')} bind:value={newOrg.address} placeholder="Street address" />
		<div class="grid gap-4 sm:grid-cols-2">
			<Input label={t('organization.city')} bind:value={newOrg.city} placeholder="e.g. Bangalore" />
			<Input label={t('organization.state')} bind:value={newOrg.state} placeholder="e.g. Karnataka" />
		</div>
		<div class="grid gap-4 sm:grid-cols-2">
			<Input label={t('organization.contact_phone')} bind:value={newOrg.contact_phone} placeholder="+91..." />
			<Input label={t('organization.contact_email')} bind:value={newOrg.contact_email} placeholder="admin@org.com" />
		</div>
	</div>
	{#snippet footer()}
		<Button variant="outline" onclick={() => (showCreateOrgModal = false)}>{t('common.cancel')}</Button>
		<Button variant="primary" loading={creatingOrg} onclick={handleCreateOrg}>{t('organization.create')}</Button>
	{/snippet}
</Modal>

<!-- Create Service Request Modal -->
<Modal bind:open={showCreateRequestModal} title={t('organization.create_request')} size="lg">
	<div class="space-y-4">
		<Input label={t('organization.request_title')} bind:value={newRequest.title} required placeholder="e.g. Plumbing issue in Block A" />
		<Input label={t('organization.category')} bind:value={newRequest.category_id} required placeholder="Category ID" />
		<Input label={t('organization.description')} bind:value={newRequest.description} placeholder="Describe the issue..." />
		<Select
			options={priorityOptions.filter(o => o.value !== '')}
			bind:value={newRequest.priority}
			label={t('organization.priority')}
		/>
		<Input label={t('organization.scheduled_at')} type="datetime-local" bind:value={newRequest.scheduled_at} />
		<Input label={t('organization.notes')} bind:value={newRequest.notes} placeholder="Additional notes..." />
	</div>
	{#snippet footer()}
		<Button variant="outline" onclick={() => (showCreateRequestModal = false)}>{t('common.cancel')}</Button>
		<Button variant="primary" loading={creatingRequest} onclick={handleCreateRequest}>{t('organization.create')}</Button>
	{/snippet}
</Modal>

<!-- Add Member Modal -->
<Modal bind:open={showAddMemberModal} title={t('organization.add_member')}>
	<div class="space-y-4">
		<Input label={t('organization.user_id')} bind:value={newMember.user_id} required placeholder="User UUID" />
		<Select options={roleOptions} bind:value={newMember.role} label={t('organization.role')} />
	</div>
	{#snippet footer()}
		<Button variant="outline" onclick={() => (showAddMemberModal = false)}>{t('common.cancel')}</Button>
		<Button variant="primary" loading={addingMember} onclick={handleAddMember}>{t('organization.add_member')}</Button>
	{/snippet}
</Modal>
