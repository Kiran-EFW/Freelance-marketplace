<script lang="ts">
	import { TrendingUp, TrendingDown, Users, Briefcase, IndianRupee, Star, MapPin, Clock } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Tabs from '$lib/components/ui/Tabs.svelte';

	let period = $state('monthly');
	const periodTabs = [
		{ id: 'weekly', label: 'Weekly' },
		{ id: 'monthly', label: 'Monthly' },
		{ id: 'yearly', label: 'Yearly' }
	];

	const kpiCards = [
		{ label: 'New Users', value: '1,842', change: 12.3, icon: Users, color: 'primary' },
		{ label: 'Jobs Created', value: '3,456', change: 15.6, icon: Briefcase, color: 'blue' },
		{ label: 'Revenue', value: 'Rs. 42.5L', change: 18.2, icon: IndianRupee, color: 'secondary' },
		{ label: 'Avg Rating', value: '4.6', change: 2.1, icon: Star, color: 'yellow' }
	];

	const userGrowth = [
		{ month: 'Oct', customers: 820, providers: 180 },
		{ month: 'Nov', customers: 950, providers: 210 },
		{ month: 'Dec', customers: 1100, providers: 240 },
		{ month: 'Jan', customers: 980, providers: 220 },
		{ month: 'Feb', customers: 1250, providers: 280 },
		{ month: 'Mar', customers: 1450, providers: 320 }
	];
	const maxUsers = Math.max(...userGrowth.map((d) => d.customers + d.providers));

	const jobsByCategory = [
		{ category: 'Plumbing', count: 4520, percentage: 28 },
		{ category: 'Electrical', count: 3200, percentage: 20 },
		{ category: 'Cleaning', count: 2850, percentage: 18 },
		{ category: 'Painting', count: 1950, percentage: 12 },
		{ category: 'Gardening', count: 1600, percentage: 10 },
		{ category: 'Carpentry', count: 1200, percentage: 8 },
		{ category: 'Other', count: 680, percentage: 4 }
	];

	const topProviders = [
		{ name: 'Lakshmi Bai', category: 'Cleaning', rating: 4.9, jobs: 210, revenue: 850000 },
		{ name: 'Suresh Nair', category: 'Plumbing', rating: 4.8, jobs: 156, revenue: 720000 },
		{ name: 'Mohan Rao', category: 'Electrical', rating: 4.7, jobs: 142, revenue: 680000 },
		{ name: 'Kiran Rao', category: 'Painting', rating: 4.5, jobs: 98, revenue: 450000 },
		{ name: 'Priya Sharma', category: 'Gardening', rating: 4.6, jobs: 88, revenue: 380000 }
	];

	const topCities = [
		{ city: 'Bangalore', jobs: 12500, revenue: 'Rs. 18.5L', growth: 22 },
		{ city: 'Mumbai', jobs: 8400, revenue: 'Rs. 12.8L', growth: 18 },
		{ city: 'Delhi', jobs: 6200, revenue: 'Rs. 9.2L', growth: 15 },
		{ city: 'Chennai', jobs: 4800, revenue: 'Rs. 7.1L', growth: 12 },
		{ city: 'Hyderabad', jobs: 3500, revenue: 'Rs. 5.2L', growth: 20 }
	];

	const platformMetrics = {
		avgJobValue: 3200,
		avgResponseTime: '2.4 hours',
		completionRate: 94.7,
		repeatCustomerRate: 42,
		avgProviderRating: 4.6,
		disputeRate: 2.1,
		customerSatisfaction: 91
	};

	const categoryColors = ['bg-primary-500', 'bg-blue-500', 'bg-secondary-500', 'bg-purple-500', 'bg-yellow-500', 'bg-pink-500', 'bg-gray-400'];
</script>

<svelte:head>
	<title>Analytics - Admin - Seva</title>
</svelte:head>

<div class="px-6 py-8 lg:px-8">
	<div class="flex flex-wrap items-center justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Analytics</h1>
			<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">Platform performance and insights.</p>
		</div>
		<Tabs tabs={periodTabs} bind:activeTab={period} />
	</div>

	<!-- KPI Cards -->
	<div class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
		{#each kpiCards as kpi}
			{@const Icon = kpi.icon}
			<Card>
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-3">
						<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-{kpi.color}-100 text-{kpi.color}-600 dark:bg-{kpi.color}-900/30">
							<Icon class="h-5 w-5" />
						</div>
						<span class="text-sm font-medium text-gray-600 dark:text-gray-400">{kpi.label}</span>
					</div>
					<div class="flex items-center gap-1 {kpi.change >= 0 ? 'text-secondary-600 dark:text-secondary-400' : 'text-red-600 dark:text-red-400'}">
						{#if kpi.change >= 0}
							<TrendingUp class="h-3 w-3" />
						{:else}
							<TrendingDown class="h-3 w-3" />
						{/if}
						<span class="text-xs">{kpi.change >= 0 ? '+' : ''}{kpi.change}%</span>
					</div>
				</div>
				<p class="mt-3 text-3xl font-bold text-gray-900 dark:text-white">{kpi.value}</p>
			</Card>
		{/each}
	</div>

	<!-- Charts Row -->
	<div class="mt-8 grid gap-6 lg:grid-cols-2">
		<!-- User Growth Chart -->
		<Card>
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">User Growth</h2>
			<div class="mt-6 flex items-end gap-3 h-48">
				{#each userGrowth as bar}
					{@const total = bar.customers + bar.providers}
					{@const height = (total / maxUsers) * 100}
					{@const providerHeight = (bar.providers / total) * 100}
					<div class="flex flex-1 flex-col items-center gap-2">
						<span class="text-xs font-medium text-gray-700 dark:text-gray-300">{total}</span>
						<div class="w-full rounded-t-lg overflow-hidden" style="height: {height}%">
							<div class="w-full bg-primary-500 dark:bg-primary-600" style="height: {100 - providerHeight}%"></div>
							<div class="w-full bg-secondary-500 dark:bg-secondary-600" style="height: {providerHeight}%"></div>
						</div>
						<span class="text-xs text-gray-500 dark:text-gray-400">{bar.month}</span>
					</div>
				{/each}
			</div>
			<div class="mt-4 flex items-center gap-6 text-xs text-gray-500 dark:text-gray-400">
				<div class="flex items-center gap-2">
					<div class="h-3 w-3 rounded bg-primary-500"></div>
					<span>Customers</span>
				</div>
				<div class="flex items-center gap-2">
					<div class="h-3 w-3 rounded bg-secondary-500"></div>
					<span>Providers</span>
				</div>
			</div>
		</Card>

		<!-- Jobs by Category -->
		<Card>
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Jobs by Category</h2>
			<div class="mt-6 space-y-3">
				{#each jobsByCategory as cat, i}
					<div class="flex items-center gap-3">
						<span class="w-20 text-sm text-gray-600 dark:text-gray-400 truncate">{cat.category}</span>
						<div class="flex-1 h-6 rounded-full bg-gray-100 dark:bg-gray-700 overflow-hidden">
							<div class="h-6 rounded-full {categoryColors[i] || 'bg-gray-400'} transition-all flex items-center px-2" style="width: {cat.percentage}%">
								{#if cat.percentage >= 10}
									<span class="text-xs font-medium text-white">{cat.percentage}%</span>
								{/if}
							</div>
						</div>
						<span class="w-12 text-right text-xs text-gray-500 dark:text-gray-400">{cat.count.toLocaleString()}</span>
					</div>
				{/each}
			</div>
		</Card>
	</div>

	<!-- Platform Metrics -->
	<Card class="mt-8">
		<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Platform Metrics</h2>
		<div class="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
			<div class="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
				<p class="text-sm text-gray-500 dark:text-gray-400">Avg Job Value</p>
				<p class="mt-1 text-xl font-bold text-gray-900 dark:text-white">Rs. {platformMetrics.avgJobValue.toLocaleString()}</p>
			</div>
			<div class="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
				<p class="text-sm text-gray-500 dark:text-gray-400">Avg Response Time</p>
				<p class="mt-1 text-xl font-bold text-gray-900 dark:text-white">{platformMetrics.avgResponseTime}</p>
			</div>
			<div class="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
				<p class="text-sm text-gray-500 dark:text-gray-400">Completion Rate</p>
				<p class="mt-1 text-xl font-bold text-secondary-600 dark:text-secondary-400">{platformMetrics.completionRate}%</p>
			</div>
			<div class="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
				<p class="text-sm text-gray-500 dark:text-gray-400">Repeat Customer Rate</p>
				<p class="mt-1 text-xl font-bold text-gray-900 dark:text-white">{platformMetrics.repeatCustomerRate}%</p>
			</div>
			<div class="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
				<p class="text-sm text-gray-500 dark:text-gray-400">Avg Provider Rating</p>
				<p class="mt-1 text-xl font-bold text-yellow-600 dark:text-yellow-400">{platformMetrics.avgProviderRating}/5</p>
			</div>
			<div class="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
				<p class="text-sm text-gray-500 dark:text-gray-400">Dispute Rate</p>
				<p class="mt-1 text-xl font-bold text-gray-900 dark:text-white">{platformMetrics.disputeRate}%</p>
			</div>
			<div class="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
				<p class="text-sm text-gray-500 dark:text-gray-400">Customer Satisfaction</p>
				<p class="mt-1 text-xl font-bold text-secondary-600 dark:text-secondary-400">{platformMetrics.customerSatisfaction}%</p>
			</div>
		</div>
	</Card>

	<!-- Bottom Row -->
	<div class="mt-8 grid gap-6 lg:grid-cols-2">
		<!-- Top Providers -->
		<Card>
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Top Providers</h2>
			<div class="mt-4 overflow-x-auto">
				<table class="w-full text-sm">
					<thead>
						<tr class="border-b border-gray-200 dark:border-gray-700">
							<th class="pb-2 text-left font-medium text-gray-500 dark:text-gray-400">Provider</th>
							<th class="pb-2 text-left font-medium text-gray-500 dark:text-gray-400">Category</th>
							<th class="pb-2 text-right font-medium text-gray-500 dark:text-gray-400">Rating</th>
							<th class="pb-2 text-right font-medium text-gray-500 dark:text-gray-400">Jobs</th>
							<th class="pb-2 text-right font-medium text-gray-500 dark:text-gray-400">Revenue</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-700">
						{#each topProviders as provider}
							<tr>
								<td class="py-2 font-medium text-gray-900 dark:text-white">{provider.name}</td>
								<td class="py-2 text-gray-500 dark:text-gray-400">{provider.category}</td>
								<td class="py-2 text-right text-yellow-600 dark:text-yellow-400">{provider.rating}</td>
								<td class="py-2 text-right text-gray-900 dark:text-white">{provider.jobs}</td>
								<td class="py-2 text-right text-gray-900 dark:text-white">Rs. {(provider.revenue / 1000).toFixed(0)}k</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</Card>

		<!-- Top Cities -->
		<Card>
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Top Cities</h2>
			<div class="mt-4 overflow-x-auto">
				<table class="w-full text-sm">
					<thead>
						<tr class="border-b border-gray-200 dark:border-gray-700">
							<th class="pb-2 text-left font-medium text-gray-500 dark:text-gray-400">City</th>
							<th class="pb-2 text-right font-medium text-gray-500 dark:text-gray-400">Jobs</th>
							<th class="pb-2 text-right font-medium text-gray-500 dark:text-gray-400">Revenue</th>
							<th class="pb-2 text-right font-medium text-gray-500 dark:text-gray-400">Growth</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-700">
						{#each topCities as city}
							<tr>
								<td class="py-2">
									<div class="flex items-center gap-2">
										<MapPin class="h-4 w-4 text-gray-400" />
										<span class="font-medium text-gray-900 dark:text-white">{city.city}</span>
									</div>
								</td>
								<td class="py-2 text-right text-gray-900 dark:text-white">{city.jobs.toLocaleString()}</td>
								<td class="py-2 text-right text-gray-900 dark:text-white">{city.revenue}</td>
								<td class="py-2 text-right">
									<span class="flex items-center justify-end gap-1 text-secondary-600 dark:text-secondary-400">
										<TrendingUp class="h-3 w-3" />
										+{city.growth}%
									</span>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</Card>
	</div>
</div>
