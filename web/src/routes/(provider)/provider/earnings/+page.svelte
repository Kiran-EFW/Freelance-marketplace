<script lang="ts">
	import { IndianRupee, TrendingUp, ArrowLeft, CreditCard, Clock, CheckCircle } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Tabs from '$lib/components/ui/Tabs.svelte';

	let period = $state('weekly');
	const tabs = [
		{ id: 'daily', label: 'Daily' },
		{ id: 'weekly', label: 'Weekly' },
		{ id: 'monthly', label: 'Monthly' }
	];

	const totalEarnings = 148500;
	const pendingPayout = 12500;

	const weeklyData = [
		{ label: 'Mar 3', amount: 8500 },
		{ label: 'Mar 10', amount: 12000 },
		{ label: 'Mar 17', amount: 14500 }
	];

	const maxAmount = Math.max(...weeklyData.map((d) => d.amount));

	const payouts = [
		{ id: '1', amount: 35000, status: 'completed', bank: '4521', date: '2026-03-01', completedAt: '2026-03-02' },
		{ id: '2', amount: 28000, status: 'completed', bank: '4521', date: '2026-02-15', completedAt: '2026-02-16' },
		{ id: '3', amount: 42000, status: 'completed', bank: '4521', date: '2026-02-01', completedAt: '2026-02-02' },
		{ id: '4', amount: 12500, status: 'pending', bank: '4521', date: '2026-03-17', completedAt: null }
	];

	const statusBadge: Record<string, 'success' | 'warning' | 'info'> = {
		completed: 'success', processing: 'warning', pending: 'info'
	};
</script>

<svelte:head>
	<title>Earnings - Seva Provider</title>
</svelte:head>

<div class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
	<a href="/provider/dashboard" class="inline-flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700">
		<ArrowLeft class="h-4 w-4" />
		Back to Dashboard
	</a>

	<h1 class="mt-4 text-2xl font-bold text-gray-900 dark:text-white">Earnings</h1>

	<!-- Summary -->
	<div class="mt-6 grid gap-4 sm:grid-cols-3">
		<Card>
			<div class="flex items-center gap-2">
				<IndianRupee class="h-5 w-5 text-secondary-600" />
				<span class="text-sm text-gray-500 dark:text-gray-400">Total Earned</span>
			</div>
			<p class="mt-2 text-2xl font-bold text-gray-900 dark:text-white">Rs. {totalEarnings.toLocaleString()}</p>
		</Card>
		<Card>
			<div class="flex items-center gap-2">
				<Clock class="h-5 w-5 text-yellow-500" />
				<span class="text-sm text-gray-500 dark:text-gray-400">Pending Payout</span>
			</div>
			<p class="mt-2 text-2xl font-bold text-yellow-600 dark:text-yellow-400">Rs. {pendingPayout.toLocaleString()}</p>
		</Card>
		<Card>
			<div class="flex items-center gap-2">
				<CreditCard class="h-5 w-5 text-gray-500" />
				<span class="text-sm text-gray-500 dark:text-gray-400">Bank Account</span>
			</div>
			<p class="mt-2 text-lg font-bold text-gray-900 dark:text-white">**** **** **** 4521</p>
			<p class="text-xs text-gray-500 dark:text-gray-400">HDFC Bank</p>
		</Card>
	</div>

	<!-- Chart Section -->
	<Card class="mt-6">
		<div class="flex items-center justify-between">
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Earnings Overview</h2>
		</div>
		<Tabs tabs={tabs} bind:activeTab={period} class="mt-4" />
		<div class="mt-6">
			<!-- Simple bar chart -->
			<div class="flex items-end gap-4 h-48">
				{#each weeklyData as bar}
					{@const height = (bar.amount / maxAmount) * 100}
					<div class="flex flex-1 flex-col items-center gap-2">
						<span class="text-xs font-medium text-gray-700 dark:text-gray-300">Rs. {(bar.amount / 1000).toFixed(1)}k</span>
						<div class="w-full rounded-t-lg bg-primary-500 dark:bg-primary-600 transition-all" style="height: {height}%"></div>
						<span class="text-xs text-gray-500 dark:text-gray-400">{bar.label}</span>
					</div>
				{/each}
			</div>
		</div>
	</Card>

	<!-- Payout History -->
	<Card class="mt-6">
		<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Payout History</h2>
		<div class="mt-4 space-y-3">
			{#each payouts as payout}
				<div class="flex items-center gap-3 rounded-lg border border-gray-100 p-3 dark:border-gray-700">
					<div class="flex h-10 w-10 items-center justify-center rounded-full
						{payout.status === 'completed' ? 'bg-green-100 text-green-600 dark:bg-green-900/20' : 'bg-yellow-100 text-yellow-600 dark:bg-yellow-900/20'}">
						{#if payout.status === 'completed'}
							<CheckCircle class="h-5 w-5" />
						{:else}
							<Clock class="h-5 w-5" />
						{/if}
					</div>
					<div class="flex-1">
						<p class="text-sm font-medium text-gray-900 dark:text-white">Payout to ****{payout.bank}</p>
						<p class="text-xs text-gray-500 dark:text-gray-400">{payout.date}</p>
					</div>
					<div class="text-right">
						<p class="text-sm font-semibold text-gray-900 dark:text-white">Rs. {payout.amount.toLocaleString()}</p>
						<Badge variant={statusBadge[payout.status] || 'neutral'} size="sm">{payout.status}</Badge>
					</div>
				</div>
			{/each}
		</div>
	</Card>
</div>
