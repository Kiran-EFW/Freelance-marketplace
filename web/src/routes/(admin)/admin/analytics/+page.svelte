<script lang="ts">
	import { TrendingUp, TrendingDown, Users, Briefcase, IndianRupee, Star, MapPin, Clock, Loader2 } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Tabs from '$lib/components/ui/Tabs.svelte';
	import api from '$lib/api/client';

	let period = $state('monthly');
	let loading = $state(true);
	let error = $state('');
	const periodTabs = [
		{ id: 'weekly', label: 'Weekly' },
		{ id: 'monthly', label: 'Monthly' },
		{ id: 'yearly', label: 'Yearly' }
	];

	let kpiCards = $state([
		{ label: 'New Users', value: '0', change: 0, icon: Users, color: 'primary' },
		{ label: 'Jobs Created', value: '0', change: 0, icon: Briefcase, color: 'blue' },
		{ label: 'Revenue', value: 'Rs. 0', change: 0, icon: IndianRupee, color: 'secondary' },
		{ label: 'Avg Rating', value: '0', change: 0, icon: Star, color: 'yellow' }
	]);

	let userGrowth = $state<{ month: string; customers: number; providers: number }[]>([]);
	let maxUsers = $derived(Math.max(...userGrowth.map((d) => d.customers + d.providers), 1));

	let jobsByCategory = $state<{ category: string; count: number; percentage: number }[]>([]);

	let topProviders = $state<{ name: string; category: string; rating: number; jobs: number; revenue: number }[]>([]);

	let topCities = $state<{ city: string; jobs: number; revenue: string; growth: number }[]>([]);

	let platformMetrics = $state({
		avgJobValue: 0,
		avgResponseTime: '0',
		completionRate: 0,
		repeatCustomerRate: 0,
		avgProviderRating: 0,
		disputeRate: 0,
		customerSatisfaction: 0
	});

	async function fetchAnalytics() {
		loading = true;
		error = '';
		try {
			const res = await api.admin.getAnalytics({ metric: period });
			const data = res.data || {};

			// Map KPI cards
			if (data.kpi) {
				const kpi = data.kpi as any;
				kpiCards = [
					{ label: 'New Users', value: kpi.new_users?.toLocaleString() || '0', change: kpi.new_users_change || 0, icon: Users, color: 'primary' },
					{ label: 'Jobs Created', value: kpi.jobs_created?.toLocaleString() || '0', change: kpi.jobs_created_change || 0, icon: Briefcase, color: 'blue' },
					{ label: 'Revenue', value: kpi.revenue_formatted || `Rs. ${((kpi.revenue || 0) / 100000).toFixed(1)}L`, change: kpi.revenue_change || 0, icon: IndianRupee, color: 'secondary' },
					{ label: 'Avg Rating', value: String(kpi.avg_rating || 0), change: kpi.avg_rating_change || 0, icon: Star, color: 'yellow' }
				];
			}

			// Map user growth chart data
			if (data.user_growth && Array.isArray(data.user_growth)) {
				userGrowth = (data.user_growth as any[]).map((item) => ({
					month: item.month || item.label || '',
					customers: item.customers || 0,
					providers: item.providers || 0
				}));
			}

			// Map jobs by category
			if (data.jobs_by_category && Array.isArray(data.jobs_by_category)) {
				jobsByCategory = (data.jobs_by_category as any[]).map((item) => ({
					category: item.category || item.name || '',
					count: item.count || 0,
					percentage: item.percentage || 0
				}));
			}

			// Map top providers
			if (data.top_providers && Array.isArray(data.top_providers)) {
				topProviders = (data.top_providers as any[]).map((p) => ({
					name: p.name || '',
					category: p.category || '',
					rating: p.rating || 0,
					jobs: p.jobs || p.completed_jobs || 0,
					revenue: p.revenue || 0
				}));
			}

			// Map top cities
			if (data.top_cities && Array.isArray(data.top_cities)) {
				topCities = (data.top_cities as any[]).map((c) => ({
					city: c.city || c.name || '',
					jobs: c.jobs || 0,
					revenue: c.revenue_formatted || `Rs. ${((c.revenue || 0) / 100000).toFixed(1)}L`,
					growth: c.growth || 0
				}));
			}

			// Map platform metrics
			if (data.platform_metrics) {
				const pm = data.platform_metrics as any;
				platformMetrics = {
					avgJobValue: pm.avg_job_value || 0,
					avgResponseTime: pm.avg_response_time || '0',
					completionRate: pm.completion_rate || 0,
					repeatCustomerRate: pm.repeat_customer_rate || 0,
					avgProviderRating: pm.avg_provider_rating || 0,
					disputeRate: pm.dispute_rate || 0,
					customerSatisfaction: pm.customer_satisfaction || 0
				};
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load analytics';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		const _p = period;
		fetchAnalytics();
	});

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

	{#if loading}
		<div class="mt-8 flex items-center justify-center py-12">
			<Loader2 class="h-8 w-8 animate-spin text-primary-600" />
		</div>
	{:else if error}
		<div class="mt-4 rounded-lg border border-red-200 bg-red-50 p-6 text-center dark:border-red-800 dark:bg-red-900/20">
			<p class="text-red-600 dark:text-red-400">{error}</p>
		</div>
	{:else}
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
	{/if}
</div>
