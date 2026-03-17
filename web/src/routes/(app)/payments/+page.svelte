<script lang="ts">
	import { IndianRupee, Filter, ArrowUpRight, ArrowDownLeft, Search } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Pagination from '$lib/components/ui/Pagination.svelte';

	let statusFilter = $state('all');
	let searchQuery = $state('');
	let currentPage = $state(1);

	const mockTransactions = [
		{ id: '1', jobTitle: 'Fix leaking tap', amount: 800, type: 'payment', status: 'released', method: 'UPI', date: '2026-03-15' },
		{ id: '2', jobTitle: 'Electrical wiring', amount: 1500, type: 'payment', status: 'held', method: 'Card', date: '2026-03-10' },
		{ id: '3', jobTitle: 'Deep cleaning', amount: 2000, type: 'payment', status: 'released', method: 'UPI', date: '2026-02-28' },
		{ id: '4', jobTitle: 'Garden maintenance', amount: 600, type: 'refund', status: 'refunded', method: 'UPI', date: '2026-02-15' },
		{ id: '5', jobTitle: 'Painting bedroom', amount: 3500, type: 'payment', status: 'released', method: 'Cash', date: '2026-01-20' },
		{ id: '6', jobTitle: 'AC repair', amount: 1200, type: 'payment', status: 'pending', method: 'Card', date: '2026-01-15' }
	];

	const statusBadge: Record<string, 'success' | 'warning' | 'danger' | 'info' | 'neutral'> = {
		released: 'success', held: 'warning', pending: 'info', refunded: 'danger', failed: 'danger'
	};

	const totalSpent = $derived(
		mockTransactions.filter((t) => t.type === 'payment' && t.status !== 'refunded').reduce((sum, t) => sum + t.amount, 0)
	);

	const totalRefunded = $derived(
		mockTransactions.filter((t) => t.status === 'refunded').reduce((sum, t) => sum + t.amount, 0)
	);

	const pendingAmount = $derived(
		mockTransactions.filter((t) => t.status === 'pending' || t.status === 'held').reduce((sum, t) => sum + t.amount, 0)
	);
</script>

<svelte:head>
	<title>Payments - Seva</title>
</svelte:head>

<div class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
	<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Payments</h1>
	<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">View your transaction history and payment details.</p>

	<!-- Summary Cards -->
	<div class="mt-6 grid gap-4 sm:grid-cols-3">
		<Card>
			<p class="text-sm text-gray-500 dark:text-gray-400">Total Spent</p>
			<p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">Rs. {totalSpent.toLocaleString()}</p>
		</Card>
		<Card>
			<p class="text-sm text-gray-500 dark:text-gray-400">Pending</p>
			<p class="mt-1 text-2xl font-bold text-yellow-600 dark:text-yellow-400">Rs. {pendingAmount.toLocaleString()}</p>
		</Card>
		<Card>
			<p class="text-sm text-gray-500 dark:text-gray-400">Refunded</p>
			<p class="mt-1 text-2xl font-bold text-red-600 dark:text-red-400">Rs. {totalRefunded.toLocaleString()}</p>
		</Card>
	</div>

	<!-- Filters -->
	<div class="mt-6 flex flex-col gap-3 sm:flex-row">
		<div class="relative flex-1">
			<Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Search transactions..."
				class="w-full rounded-lg border border-gray-300 py-2 pl-9 pr-4 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
			/>
		</div>
		<select
			bind:value={statusFilter}
			class="rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-white"
		>
			<option value="all">All Status</option>
			<option value="released">Completed</option>
			<option value="pending">Pending</option>
			<option value="held">Held</option>
			<option value="refunded">Refunded</option>
		</select>
	</div>

	<!-- Transaction List -->
	<div class="mt-4 space-y-2">
		{#each mockTransactions as txn}
			<div class="flex items-center gap-4 rounded-xl border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-800">
				<div class="flex h-10 w-10 items-center justify-center rounded-full
					{txn.type === 'refund'
						? 'bg-red-100 text-red-600 dark:bg-red-900/20 dark:text-red-400'
						: 'bg-green-100 text-green-600 dark:bg-green-900/20 dark:text-green-400'}">
					{#if txn.type === 'refund'}
						<ArrowDownLeft class="h-5 w-5" />
					{:else}
						<ArrowUpRight class="h-5 w-5" />
					{/if}
				</div>
				<div class="flex-1 min-w-0">
					<p class="text-sm font-medium text-gray-900 dark:text-white truncate">{txn.jobTitle}</p>
					<p class="text-xs text-gray-500 dark:text-gray-400">{txn.date} -- {txn.method}</p>
				</div>
				<div class="text-right">
					<p class="text-sm font-semibold {txn.type === 'refund' ? 'text-red-600 dark:text-red-400' : 'text-gray-900 dark:text-white'}">
						{txn.type === 'refund' ? '-' : ''}Rs. {txn.amount.toLocaleString()}
					</p>
					<Badge variant={statusBadge[txn.status]} size="sm">{txn.status}</Badge>
				</div>
			</div>
		{/each}
	</div>

	<Pagination {currentPage} totalPages={3} onPageChange={(p) => (currentPage = p)} class="mt-6" />
</div>
