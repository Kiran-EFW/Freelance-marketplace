<script lang="ts">
	import { Search, Filter, MoreVertical, Shield, ShieldOff, Eye, Ban, CheckCircle, Mail, Phone, Loader2 } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Pagination from '$lib/components/ui/Pagination.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import { toastSuccess, toastError } from '$lib/stores/toast';
	import api from '$lib/api/client';

	let searchQuery = $state('');
	let roleFilter = $state('all');
	let statusFilter = $state('all');
	let currentPage = $state(1);
	let loading = $state(true);
	let error = $state('');
	let usersList = $state<any[]>([]);
	let totalPages = $state(1);
	let selectedUser = $state<any | null>(null);
	let showUserModal = $state(false);

	async function fetchUsers() {
		loading = true;
		error = '';
		try {
			const params: any = {
				page: currentPage,
				per_page: 10
			};
			if (roleFilter !== 'all') params.role = roleFilter;
			if (statusFilter !== 'all') params.status = statusFilter;
			if (searchQuery.trim()) params.search = searchQuery.trim();

			const res = await api.admin.listUsers(params);
			usersList = (res.data || []).map((u: any) => ({
				id: u.id,
				name: u.name || '',
				email: u.email || '',
				phone: u.phone || '',
				role: u.role || 'customer',
				status: u.status || 'active',
				joinedAt: u.created_at?.split('T')[0] || u.joined_at || '',
				jobsPosted: u.jobs_posted || u.jobsPosted || 0,
				totalSpent: u.total_spent || u.totalSpent || 0,
				avatar: u.avatar_url || u.avatar || null,
				verified: u.verified || false,
				rating: u.rating || 0,
				completedJobs: u.completed_jobs || u.completedJobs || 0
			}));
			totalPages = res.meta?.total_pages || Math.ceil((res.meta?.total || 1) / 10);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load users';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		const _q = searchQuery;
		const _r = roleFilter;
		const _s = statusFilter;
		const _p = currentPage;
		fetchUsers();
	});

	let filteredUsers = $derived(usersList);

	const roleBadge: Record<string, 'info' | 'success' | 'warning'> = {
		customer: 'info', provider: 'success', admin: 'warning'
	};

	const statusBadge: Record<string, 'success' | 'danger' | 'neutral'> = {
		active: 'success', suspended: 'danger', inactive: 'neutral'
	};

	function viewUser(user: any) {
		selectedUser = user;
		showUserModal = true;
	}

	async function suspendUser(id: string) {
		try {
			await api.users.suspend(id);
			toastSuccess('User suspended');
			showUserModal = false;
			fetchUsers();
		} catch (err) {
			toastError(err instanceof Error ? err.message : 'Failed to suspend user');
		}
	}

	async function activateUser(id: string) {
		try {
			await api.users.activate(id);
			toastSuccess('User activated');
			showUserModal = false;
			fetchUsers();
		} catch (err) {
			toastError(err instanceof Error ? err.message : 'Failed to activate user');
		}
	}
</script>

<svelte:head>
	<title>Users - Admin - Seva</title>
</svelte:head>

<div class="px-6 py-8 lg:px-8">
	<div class="flex flex-wrap items-center justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Users</h1>
			<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">Manage all platform users.</p>
		</div>
		<p class="text-sm text-gray-500 dark:text-gray-400">{filteredUsers.length} users</p>
	</div>

	<!-- Filters -->
	<div class="mt-6 flex flex-col gap-4 sm:flex-row">
		<div class="relative flex-1">
			<Search class="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-gray-400" />
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Search by name or email..."
				class="w-full rounded-lg border border-gray-300 py-2.5 pl-10 pr-4 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
			/>
		</div>
		<select
			bind:value={roleFilter}
			class="rounded-lg border border-gray-300 px-3 py-2.5 text-sm focus:border-primary-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-white"
		>
			<option value="all">All Roles</option>
			<option value="customer">Customer</option>
			<option value="provider">Provider</option>
			<option value="admin">Admin</option>
		</select>
		<select
			bind:value={statusFilter}
			class="rounded-lg border border-gray-300 px-3 py-2.5 text-sm focus:border-primary-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-white"
		>
			<option value="all">All Status</option>
			<option value="active">Active</option>
			<option value="suspended">Suspended</option>
			<option value="inactive">Inactive</option>
		</select>
	</div>

	{#if loading}
		<div class="mt-8 flex items-center justify-center py-12">
			<Loader2 class="h-8 w-8 animate-spin text-primary-600" />
		</div>
	{:else if error}
		<div class="mt-4 rounded-lg border border-red-200 bg-red-50 p-6 text-center dark:border-red-800 dark:bg-red-900/20">
			<p class="text-red-600 dark:text-red-400">{error}</p>
		</div>
	{:else}
	<!-- Users Table -->
	<Card class="mt-6" padding="none">
		<div class="overflow-x-auto">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b border-gray-200 dark:border-gray-700">
						<th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">User</th>
						<th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">Role</th>
						<th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">Status</th>
						<th class="px-4 py-3 text-left font-medium text-gray-500 dark:text-gray-400">Joined</th>
						<th class="px-4 py-3 text-right font-medium text-gray-500 dark:text-gray-400">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-700">
					{#each filteredUsers as user}
						<tr class="hover:bg-gray-50 dark:hover:bg-gray-800/50">
							<td class="px-4 py-3">
								<div class="flex items-center gap-3">
									<Avatar name={user.name} size="sm" />
									<div>
										<p class="font-medium text-gray-900 dark:text-white">{user.name}</p>
										<p class="text-xs text-gray-500 dark:text-gray-400">{user.email}</p>
									</div>
								</div>
							</td>
							<td class="px-4 py-3">
								<Badge variant={roleBadge[user.role] || 'neutral'} size="sm">{user.role}</Badge>
							</td>
							<td class="px-4 py-3">
								<Badge variant={statusBadge[user.status] || 'neutral'} size="sm">{user.status}</Badge>
							</td>
							<td class="px-4 py-3 text-gray-500 dark:text-gray-400">{user.joinedAt}</td>
							<td class="px-4 py-3 text-right">
								<button
									onclick={() => viewUser(user)}
									class="rounded p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-700 dark:hover:text-gray-300"
								>
									<Eye class="h-4 w-4" />
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</Card>

	<div class="mt-6">
		<Pagination {currentPage} {totalPages} onPageChange={(p) => (currentPage = p)} />
	</div>
	{/if}
</div>

<!-- User Detail Modal -->
<Modal bind:open={showUserModal} title="User Details" size="md">
	{#if selectedUser}
		<div class="space-y-6">
			<div class="flex items-center gap-4">
				<Avatar name={selectedUser.name} size="lg" />
				<div>
					<h3 class="text-lg font-semibold text-gray-900 dark:text-white">{selectedUser.name}</h3>
					<Badge variant={roleBadge[selectedUser.role] || 'neutral'} size="sm">{selectedUser.role}</Badge>
					<Badge variant={statusBadge[selectedUser.status] || 'neutral'} size="sm" class="ml-1">{selectedUser.status}</Badge>
				</div>
			</div>

			<div class="grid grid-cols-2 gap-4">
				<div>
					<p class="text-xs text-gray-500 dark:text-gray-400">Email</p>
					<div class="mt-1 flex items-center gap-1">
						<Mail class="h-3.5 w-3.5 text-gray-400" />
						<p class="text-sm text-gray-900 dark:text-white">{selectedUser.email}</p>
					</div>
				</div>
				<div>
					<p class="text-xs text-gray-500 dark:text-gray-400">Phone</p>
					<div class="mt-1 flex items-center gap-1">
						<Phone class="h-3.5 w-3.5 text-gray-400" />
						<p class="text-sm text-gray-900 dark:text-white">{selectedUser.phone}</p>
					</div>
				</div>
				<div>
					<p class="text-xs text-gray-500 dark:text-gray-400">Joined</p>
					<p class="mt-1 text-sm text-gray-900 dark:text-white">{selectedUser.joinedAt}</p>
				</div>
				{#if selectedUser.role === 'customer'}
					<div>
						<p class="text-xs text-gray-500 dark:text-gray-400">Jobs Posted</p>
						<p class="mt-1 text-sm text-gray-900 dark:text-white">{selectedUser.jobsPosted}</p>
					</div>
					<div>
						<p class="text-xs text-gray-500 dark:text-gray-400">Total Spent</p>
						<p class="mt-1 text-sm text-gray-900 dark:text-white">Rs. {selectedUser.totalSpent.toLocaleString()}</p>
					</div>
				{/if}
				{#if selectedUser.role === 'provider' && 'completedJobs' in selectedUser}
					<div>
						<p class="text-xs text-gray-500 dark:text-gray-400">Completed Jobs</p>
						<p class="mt-1 text-sm text-gray-900 dark:text-white">{selectedUser.completedJobs}</p>
					</div>
					<div>
						<p class="text-xs text-gray-500 dark:text-gray-400">Rating</p>
						<p class="mt-1 text-sm text-gray-900 dark:text-white">{selectedUser.rating || 'N/A'}</p>
					</div>
					<div>
						<p class="text-xs text-gray-500 dark:text-gray-400">Verified</p>
						<p class="mt-1 text-sm text-gray-900 dark:text-white">{selectedUser.verified ? 'Yes' : 'No'}</p>
					</div>
				{/if}
			</div>
		</div>
	{/if}
	{#snippet footer()}
		<Button variant="outline" onclick={() => (showUserModal = false)}>Close</Button>
		{#if selectedUser}
			{#if selectedUser.status === 'active'}
				<Button variant="danger" onclick={() => suspendUser(selectedUser!.id)}>
					<Ban class="h-4 w-4" />
					Suspend User
				</Button>
			{:else}
				<Button variant="primary" onclick={() => activateUser(selectedUser!.id)}>
					<CheckCircle class="h-4 w-4" />
					Activate User
				</Button>
			{/if}
		{/if}
	{/snippet}
</Modal>
