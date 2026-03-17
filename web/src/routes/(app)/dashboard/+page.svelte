<script lang="ts">
	import { onMount } from 'svelte';
	import { Briefcase, Clock, Star, IndianRupee, ArrowRight, Bell, Plus, Search, Calendar, CheckCircle, MessageSquare, Loader2 } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import api from '$lib/api/client';
	import { getCurrentUser } from '$lib/stores/auth';

	let loading = $state(true);
	let error = $state('');

	let stats = $state({
		activeJobs: 0,
		pendingQuotes: 0,
		totalReviews: 0,
		totalSpent: 0
	});

	let recentJobs = $state<any[]>([]);
	let recentNotifications = $state<any[]>([]);

	onMount(async () => {
		try {
			const [jobsRes, notifsRes] = await Promise.all([
				api.jobs.list({ per_page: 4 }),
				api.notifications.list({ per_page: 4 })
			]);

			recentJobs = (jobsRes.data || []).map((job: any) => ({
				id: job.id,
				title: job.title,
				status: job.status,
				category: job.category?.name || '',
				provider: job.provider?.user?.name || null,
				budget: job.budget_min || 0,
				createdAt: job.created_at?.split('T')[0] || '',
				quotesCount: job.quotes_count || 0
			}));

			recentNotifications = (notifsRes.data || []).map((n: any) => ({
				id: n.id,
				message: n.message || n.title || '',
				time: n.created_at || '',
				read: n.read ?? n.is_read ?? true
			}));

			// Compute stats from jobs data
			const allJobs = jobsRes.data || [];
			stats = {
				activeJobs: allJobs.filter((j: any) => j.status === 'in_progress' || j.status === 'accepted').length,
				pendingQuotes: allJobs.filter((j: any) => j.status === 'quoted' || j.status === 'open').length,
				totalReviews: 0,
				totalSpent: allJobs.reduce((sum: number, j: any) => sum + (j.budget_min || 0), 0)
			};
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load dashboard data';
		} finally {
			loading = false;
		}
	});

	const statusBadge: Record<string, 'warning' | 'info' | 'success' | 'neutral' | 'danger'> = {
		posted: 'info',
		open: 'info',
		quoted: 'warning',
		accepted: 'info',
		in_progress: 'warning',
		completed: 'success',
		cancelled: 'danger'
	};
</script>

<svelte:head>
	<title>Dashboard - Seva</title>
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
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Dashboard</h1>
			<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">Welcome back! Here's an overview of your activity.</p>
		</div>
		<Button variant="primary" href="/jobs/new">
			<Plus class="h-4 w-4" />
			Post a Job
		</Button>
	</div>

	<!-- Stats Grid -->
	<div class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
		<Card>
			<div class="flex items-center gap-3">
				<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-primary-100 text-primary-600 dark:bg-primary-900/30">
					<Briefcase class="h-5 w-5" />
				</div>
				<span class="text-sm font-medium text-gray-600 dark:text-gray-400">Active Jobs</span>
			</div>
			<p class="mt-2 text-3xl font-bold text-gray-900 dark:text-white">{stats.activeJobs}</p>
		</Card>
		<Card>
			<div class="flex items-center gap-3">
				<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-yellow-100 text-yellow-600 dark:bg-yellow-900/30">
					<Clock class="h-5 w-5" />
				</div>
				<span class="text-sm font-medium text-gray-600 dark:text-gray-400">Pending Quotes</span>
			</div>
			<p class="mt-2 text-3xl font-bold text-gray-900 dark:text-white">{stats.pendingQuotes}</p>
		</Card>
		<Card>
			<div class="flex items-center gap-3">
				<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-yellow-100 text-yellow-500 dark:bg-yellow-900/30">
					<Star class="h-5 w-5" />
				</div>
				<span class="text-sm font-medium text-gray-600 dark:text-gray-400">Reviews</span>
			</div>
			<p class="mt-2 text-3xl font-bold text-gray-900 dark:text-white">{stats.totalReviews}</p>
		</Card>
		<Card>
			<div class="flex items-center gap-3">
				<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-secondary-100 text-secondary-600 dark:bg-secondary-900/30">
					<IndianRupee class="h-5 w-5" />
				</div>
				<span class="text-sm font-medium text-gray-600 dark:text-gray-400">Total Spent</span>
			</div>
			<p class="mt-2 text-3xl font-bold text-gray-900 dark:text-white">Rs. {stats.totalSpent.toLocaleString()}</p>
		</Card>
	</div>

	<div class="mt-8 grid gap-6 lg:grid-cols-3">
		<!-- Recent Jobs -->
		<div class="lg:col-span-2">
			<Card>
				<div class="flex items-center justify-between">
					<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Recent Jobs</h2>
					<a href="/jobs" class="flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700">
						View all <ArrowRight class="h-4 w-4" />
					</a>
				</div>
				<div class="mt-4 space-y-3">
					{#each recentJobs as job}
						<a href="/jobs/{job.id}" class="block">
							<div class="flex items-center gap-3 rounded-lg border border-gray-100 p-3 transition hover:bg-gray-50 dark:border-gray-700 dark:hover:bg-gray-800/50">
								<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary-100 text-primary-600 dark:bg-primary-900/30">
									<Briefcase class="h-5 w-5" />
								</div>
								<div class="flex-1 min-w-0">
									<div class="flex items-center gap-2">
										<p class="text-sm font-medium text-gray-900 dark:text-white truncate">{job.title}</p>
										<Badge variant={statusBadge[job.status] || 'neutral'} size="sm">{job.status.replace('_', ' ')}</Badge>
									</div>
									<div class="mt-0.5 flex items-center gap-3 text-xs text-gray-500 dark:text-gray-400">
										<span>{job.category}</span>
										{#if job.provider}
											<span>by {job.provider}</span>
										{:else}
											<span>{job.quotesCount} quotes</span>
										{/if}
									</div>
								</div>
								<div class="text-right">
									<p class="text-sm font-semibold text-gray-900 dark:text-white">Rs. {job.budget.toLocaleString()}</p>
									<p class="text-xs text-gray-500 dark:text-gray-400">{job.createdAt}</p>
								</div>
							</div>
						</a>
					{/each}
				</div>
			</Card>

			<!-- Quick Actions -->
			<div class="mt-6 grid gap-3 sm:grid-cols-3">
				<Button variant="primary" href="/jobs/new" class="justify-center">
					<Plus class="h-4 w-4" />
					Post a Job
				</Button>
				<Button variant="outline" href="/providers" class="justify-center">
					<Search class="h-4 w-4" />
					Find Providers
				</Button>
				<Button variant="outline" href="/payments" class="justify-center">
					<IndianRupee class="h-4 w-4" />
					Payments
				</Button>
			</div>
		</div>

		<!-- Sidebar -->
		<div class="space-y-6">
			<!-- Notifications -->
			<Card>
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-2">
						<Bell class="h-5 w-5 text-gray-500" />
						<h2 class="font-semibold text-gray-900 dark:text-white">Notifications</h2>
					</div>
					<a href="/notifications" class="text-xs text-primary-600 hover:text-primary-700">View all</a>
				</div>
				<div class="mt-4 space-y-3">
					{#each recentNotifications as notif}
						<div class="flex items-start gap-2">
							{#if !notif.read}
								<div class="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-primary-500"></div>
							{:else}
								<div class="mt-1.5 h-2 w-2 shrink-0"></div>
							{/if}
							<div>
								<p class="text-sm text-gray-700 dark:text-gray-300 {!notif.read ? 'font-medium' : ''}">{notif.message}</p>
								<p class="text-xs text-gray-400 dark:text-gray-500">{notif.time}</p>
							</div>
						</div>
					{/each}
				</div>
			</Card>

			<!-- Quick Links -->
			<Card>
				<h2 class="font-semibold text-gray-900 dark:text-white">Quick Links</h2>
				<div class="mt-4 space-y-2">
					<a href="/profile" class="flex items-center gap-3 rounded-lg p-2 text-sm text-gray-600 transition hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700">
						<Avatar name="User" size="sm" />
						Edit Profile
					</a>
					<a href="/reviews" class="flex items-center gap-2 rounded-lg p-2 text-sm text-gray-600 transition hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700">
						<Star class="h-4 w-4" />
						My Reviews
					</a>
					<a href="/disputes" class="flex items-center gap-2 rounded-lg p-2 text-sm text-gray-600 transition hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700">
						<MessageSquare class="h-4 w-4" />
						Disputes
					</a>
					<a href="/points" class="flex items-center gap-2 rounded-lg p-2 text-sm text-gray-600 transition hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700">
						<CheckCircle class="h-4 w-4" />
						Points & Rewards
					</a>
				</div>
			</Card>
		</div>
	</div>
</div>
{/if}
