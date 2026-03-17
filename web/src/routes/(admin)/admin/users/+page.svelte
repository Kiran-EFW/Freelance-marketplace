<script lang="ts">
	import { Search, Filter, MoreVertical, Shield, ShieldOff, Eye, Ban, CheckCircle, Mail, Phone } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Pagination from '$lib/components/ui/Pagination.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import { toastSuccess } from '$lib/stores/toast';

	let searchQuery = $state('');
	let roleFilter = $state('all');
	let statusFilter = $state('all');
	let currentPage = $state(1);
	let selectedUser = $state<typeof mockUsers[0] | null>(null);
	let showUserModal = $state(false);

	const mockUsers = [
		{ id: '1', name: 'Amit Verma', email: 'amit@example.com', phone: '+91 98765 43210', role: 'customer', status: 'active', joinedAt: '2025-08-15', jobsPosted: 12, totalSpent: 45000, avatar: null },
		{ id: '2', name: 'Suresh Nair', email: 'suresh@example.com', phone: '+91 98765 43211', role: 'provider', status: 'active', joinedAt: '2025-07-20', jobsPosted: 0, totalSpent: 0, avatar: null, verified: true, rating: 4.8, completedJobs: 156 },
		{ id: '3', name: 'Priya Menon', email: 'priya@example.com', phone: '+91 98765 43212', role: 'customer', status: 'active', joinedAt: '2025-09-01', jobsPosted: 8, totalSpent: 32000, avatar: null },
		{ id: '4', name: 'Deepak Kumar', email: 'deepak@example.com', phone: '+91 98765 43213', role: 'provider', status: 'suspended', joinedAt: '2025-06-10', jobsPosted: 0, totalSpent: 0, avatar: null, verified: true, rating: 3.2, completedJobs: 45 },
		{ id: '5', name: 'Anita Gupta', email: 'anita@example.com', phone: '+91 98765 43214', role: 'customer', status: 'active', joinedAt: '2025-10-05', jobsPosted: 5, totalSpent: 18500, avatar: null },
		{ id: '6', name: 'Ravi Shankar', email: 'ravi@example.com', phone: '+91 98765 43215', role: 'provider', status: 'active', joinedAt: '2025-11-12', jobsPosted: 0, totalSpent: 0, avatar: null, verified: false, rating: 0, completedJobs: 0 },
		{ id: '7', name: 'Meera Reddy', email: 'meera@example.com', phone: '+91 98765 43216', role: 'customer', status: 'active', joinedAt: '2026-01-08', jobsPosted: 3, totalSpent: 12000, avatar: null },
		{ id: '8', name: 'Kiran Rao', email: 'kiran@example.com', phone: '+91 98765 43217', role: 'provider', status: 'active', joinedAt: '2025-12-20', jobsPosted: 0, totalSpent: 0, avatar: null, verified: true, rating: 4.5, completedJobs: 78 },
		{ id: '9', name: 'Lakshmi Bai', email: 'lakshmi@example.com', phone: '+91 98765 43218', role: 'provider', status: 'active', joinedAt: '2026-02-01', jobsPosted: 0, totalSpent: 0, avatar: null, verified: true, rating: 4.9, completedJobs: 210 },
		{ id: '10', name: 'Arjun Das', email: 'arjun@example.com', phone: '+91 98765 43219', role: 'customer', status: 'inactive', joinedAt: '2025-05-22', jobsPosted: 1, totalSpent: 3500, avatar: null }
	];

	let filteredUsers = $derived(
		mockUsers.filter((u) => {
			const matchesSearch = !searchQuery || u.name.toLowerCase().includes(searchQuery.toLowerCase()) || u.email.toLowerCase().includes(searchQuery.toLowerCase());
			const matchesRole = roleFilter === 'all' || u.role === roleFilter;
			const matchesStatus = statusFilter === 'all' || u.status === statusFilter;
			return matchesSearch && matchesRole && matchesStatus;
		})
	);

	const roleBadge: Record<string, 'info' | 'success' | 'warning'> = {
		customer: 'info', provider: 'success', admin: 'warning'
	};

	const statusBadge: Record<string, 'success' | 'danger' | 'neutral'> = {
		active: 'success', suspended: 'danger', inactive: 'neutral'
	};

	function viewUser(user: typeof mockUsers[0]) {
		selectedUser = user;
		showUserModal = true;
	}

	function suspendUser(id: string) {
		toastSuccess('User suspended');
		showUserModal = false;
	}

	function activateUser(id: string) {
		toastSuccess('User activated');
		showUserModal = false;
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
		<Pagination {currentPage} totalPages={3} onPageChange={(p) => (currentPage = p)} />
	</div>
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
