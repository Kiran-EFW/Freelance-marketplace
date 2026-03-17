<script lang="ts">
	import { onMount } from 'svelte';
	import { IndianRupee, Briefcase, CheckCircle, Shield, Star, Clock, ToggleLeft, ToggleRight, ArrowRight, Calendar, Loader2 } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import StarRating from '$lib/components/ui/StarRating.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import api from '$lib/api/client';
	import { toastError } from '$lib/stores/toast';

	let loading = $state(true);
	let error = $state('');
	let isOnline = $state(true);

	let earnings = $state({ today: 0, thisWeek: 0, thisMonth: 0 });
	let activeJobs = $state(0);
	let completedThisMonth = $state(0);
	let trustScore = $state(0);
	let ratingAvg = $state(0);
	let ratingCount = $state(0);
	let ratingDist = $state<Record<number, number>>({ 5: 0, 4: 0, 3: 0, 2: 0, 1: 0 });
	let upcomingSchedule = $state<any[]>([]);

	onMount(async () => {
		try {
			const [dashRes, scheduleRes] = await Promise.all([
				api.providers.getDashboard(),
				api.routes.getSchedule().catch(() => ({ data: [] }))
			]);

			const dash = dashRes.data;
			earnings = {
				today: dash.earnings_today || 0,
				thisWeek: dash.earnings_this_week || 0,
				thisMonth: dash.earnings_this_month || 0
			};
			activeJobs = dash.active_jobs || 0;
			completedThisMonth = dash.completed_this_month || 0;
			trustScore = dash.trust_score || 0;
			ratingAvg = dash.rating_average || 0;
			ratingCount = dash.rating_count || 0;
			isOnline = dash.is_online ?? true;
			if (dash.rating_distribution) {
				ratingDist = dash.rating_distribution;
			}

			upcomingSchedule = (scheduleRes.data || []).slice(0, 5).map((entry: any) => ({
				id: entry.id,
				type: entry.type || 'job',
				title: entry.title || '',
				customer: entry.customer?.name || entry.customer_name || '',
				time: entry.time || entry.start_time || '',
				date: entry.date || '',
				location: entry.postcode || entry.location || ''
			}));
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load dashboard data';
		} finally {
			loading = false;
		}
	});

	async function toggleOnline() {
		const newState = !isOnline;
		try {
			await api.providers.toggleOnline(newState);
			isOnline = newState;
		} catch (err) {
			toastError(err instanceof Error ? err.message : 'Failed to update status');
		}
	}
</script>

<svelte:head>
	<title>Provider Dashboard - Seva</title>
</svelte:head>

{#if loading}
<div class="flex items-center justify-center py-20">
	<Loader2 class="h-8 w-8 animate-spin text-primary-600" />
</div>
{:else if error}
<div class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
	<div class="rounded-lg border border-red-200 bg-red-50 p-6 text-center dark:border-red-800 dark:bg-red-900/20">
		<p class="text-red-600 dark:text-red-400">{error}</p>
	</div>
</div>
{:else}
<div class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
	<div class="flex flex-wrap items-center justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Provider Dashboard</h1>
			<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">Welcome back! Here's your overview.</p>
		</div>
		<button
			onclick={toggleOnline}
			class="flex items-center gap-2 rounded-lg border px-4 py-2 text-sm font-medium transition
				{isOnline
					? 'border-secondary-300 bg-secondary-50 text-secondary-700 dark:border-secondary-600 dark:bg-secondary-900/20 dark:text-secondary-400'
					: 'border-gray-300 bg-gray-50 text-gray-600 dark:border-gray-600 dark:bg-gray-800 dark:text-gray-400'}"
		>
			{#if isOnline}
				<ToggleRight class="h-5 w-5" />
				Online
			{:else}
				<ToggleLeft class="h-5 w-5" />
				Offline
			{/if}
		</button>
	</div>

	<!-- Earnings Cards -->
	<div class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
		<Card>
			<div class="flex items-center gap-3">
				<IndianRupee class="h-5 w-5 text-secondary-600" />
				<span class="text-sm font-medium text-gray-600 dark:text-gray-400">Today</span>
			</div>
			<p class="mt-2 text-3xl font-bold text-gray-900 dark:text-white">Rs. {earnings.today.toLocaleString()}</p>
		</Card>
		<Card>
			<div class="flex items-center gap-3">
				<IndianRupee class="h-5 w-5 text-secondary-600" />
				<span class="text-sm font-medium text-gray-600 dark:text-gray-400">This Week</span>
			</div>
			<p class="mt-2 text-3xl font-bold text-gray-900 dark:text-white">Rs. {earnings.thisWeek.toLocaleString()}</p>
		</Card>
		<Card>
			<div class="flex items-center gap-3">
				<Briefcase class="h-5 w-5 text-primary-600" />
				<span class="text-sm font-medium text-gray-600 dark:text-gray-400">Active Jobs</span>
			</div>
			<p class="mt-2 text-3xl font-bold text-gray-900 dark:text-white">{activeJobs}</p>
		</Card>
		<Card>
			<div class="flex items-center gap-3">
				<CheckCircle class="h-5 w-5 text-secondary-600" />
				<span class="text-sm font-medium text-gray-600 dark:text-gray-400">Completed (Month)</span>
			</div>
			<p class="mt-2 text-3xl font-bold text-gray-900 dark:text-white">{completedThisMonth}</p>
		</Card>
	</div>

	<div class="mt-8 grid gap-6 lg:grid-cols-3">
		<!-- Main Content -->
		<div class="space-y-6 lg:col-span-2">
			<!-- Upcoming Schedule -->
			<Card>
				<div class="flex items-center justify-between">
					<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Upcoming Schedule</h2>
					<a href="/provider/schedule" class="flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700">
						View all <ArrowRight class="h-4 w-4" />
					</a>
				</div>
				<div class="mt-4 space-y-3">
					{#each upcomingSchedule as entry}
						<div class="flex items-center gap-3 rounded-lg border border-gray-100 p-3 dark:border-gray-700">
							<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg
								{entry.type === 'job' ? 'bg-primary-100 text-primary-600 dark:bg-primary-900/30' : 'bg-blue-100 text-blue-600 dark:bg-blue-900/30'}">
								{#if entry.type === 'job'}
									<Briefcase class="h-5 w-5" />
								{:else}
									<Calendar class="h-5 w-5" />
								{/if}
							</div>
							<div class="flex-1 min-w-0">
								<p class="text-sm font-medium text-gray-900 dark:text-white truncate">{entry.title}</p>
								<p class="text-xs text-gray-500 dark:text-gray-400">{entry.customer} -- {entry.location}</p>
							</div>
							<div class="text-right">
								<p class="text-xs font-medium text-gray-700 dark:text-gray-300">{entry.time}</p>
								<p class="text-xs text-gray-500 dark:text-gray-400">{entry.date}</p>
							</div>
						</div>
					{/each}
				</div>
			</Card>

			<!-- Quick Actions -->
			<div class="grid gap-3 sm:grid-cols-3">
				<Button variant="primary" href="/provider/schedule" class="justify-center">
					<Calendar class="h-4 w-4" />
					View Schedule
				</Button>
				<Button variant="outline" href="/provider/routes" class="justify-center">
					Manage Routes
				</Button>
				<Button variant="outline" href="/provider/earnings" class="justify-center">
					View Earnings
				</Button>
			</div>
		</div>

		<!-- Sidebar -->
		<div class="space-y-6">
			<!-- Trust Score -->
			<Card>
				<div class="flex items-center gap-2">
					<Shield class="h-5 w-5 text-secondary-500" />
					<h2 class="font-semibold text-gray-900 dark:text-white">Trust Score</h2>
				</div>
				<div class="mt-4 flex items-center justify-center">
					<div class="relative flex h-28 w-28 items-center justify-center rounded-full border-8 border-secondary-200 dark:border-secondary-800">
						<span class="text-3xl font-bold text-secondary-600 dark:text-secondary-400">{trustScore}</span>
					</div>
				</div>
				<p class="mt-2 text-center text-sm text-gray-500 dark:text-gray-400">out of 100</p>
			</Card>

			<!-- Rating -->
			<Card>
				<div class="flex items-center gap-2">
					<Star class="h-5 w-5 text-yellow-400" />
					<h2 class="font-semibold text-gray-900 dark:text-white">Rating</h2>
				</div>
				<div class="mt-4 flex items-center gap-3">
					<span class="text-3xl font-bold text-gray-900 dark:text-white">{ratingAvg}</span>
					<div>
						<StarRating rating={ratingAvg} size="sm" />
						<p class="text-xs text-gray-500 dark:text-gray-400">{ratingCount} reviews</p>
					</div>
				</div>
				<div class="mt-4 space-y-1.5">
					{#each [5, 4, 3, 2, 1] as stars}
						{@const count = ratingDist[stars as keyof typeof ratingDist] || 0}
						{@const pct = (count / ratingCount) * 100}
						<div class="flex items-center gap-2 text-sm">
							<span class="w-3 text-gray-500">{stars}</span>
							<div class="flex-1 h-2 rounded-full bg-gray-200 dark:bg-gray-700">
								<div class="h-2 rounded-full bg-yellow-400" style="width: {pct}%"></div>
							</div>
							<span class="w-8 text-right text-xs text-gray-400">{count}</span>
						</div>
					{/each}
				</div>
			</Card>

			<!-- Earnings Summary -->
			<Card>
				<h2 class="font-semibold text-gray-900 dark:text-white">This Month</h2>
				<p class="mt-2 text-2xl font-bold text-secondary-600 dark:text-secondary-400">Rs. {earnings.thisMonth.toLocaleString()}</p>
				<a href="/provider/earnings" class="mt-2 flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700">
					View breakdown <ArrowRight class="h-4 w-4" />
				</a>
			</Card>
		</div>
	</div>
</div>
{/if}
