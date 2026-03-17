<script lang="ts">
	import { Briefcase, Clock, Star, IndianRupee, ArrowRight, Bell, Plus, Search, Calendar, CheckCircle, MessageSquare } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';

	const stats = {
		activeJobs: 3,
		pendingQuotes: 5,
		totalReviews: 12,
		totalSpent: 45000
	};

	const recentJobs = [
		{
			id: '1', title: 'Fix kitchen plumbing', status: 'in_progress',
			category: 'Plumbing', provider: 'Suresh Nair', budget: 3500,
			createdAt: '2026-03-15', quotesCount: 4
		},
		{
			id: '2', title: 'Living room painting', status: 'quoted',
			category: 'Painting', provider: null, budget: 8000,
			createdAt: '2026-03-14', quotesCount: 6
		},
		{
			id: '3', title: 'AC servicing', status: 'completed',
			category: 'HVAC', provider: 'Rajesh Iyer', budget: 2000,
			createdAt: '2026-03-10', quotesCount: 3
		},
		{
			id: '4', title: 'Garden maintenance', status: 'posted',
			category: 'Gardening', provider: null, budget: 1500,
			createdAt: '2026-03-13', quotesCount: 2
		}
	];

	const recentNotifications = [
		{ id: '1', message: 'Suresh Nair sent you a quote for "Fix kitchen plumbing"', time: '10 min ago', read: false },
		{ id: '2', message: 'Your job "AC servicing" has been completed', time: '2 hours ago', read: false },
		{ id: '3', message: 'New review from Arjun Das', time: '5 hours ago', read: true },
		{ id: '4', message: 'Payment of Rs. 2,000 processed successfully', time: '1 day ago', read: true }
	];

	const statusBadge: Record<string, 'warning' | 'info' | 'success' | 'neutral' | 'danger'> = {
		posted: 'info',
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
