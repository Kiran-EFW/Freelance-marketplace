<script lang="ts">
	import { Users, Briefcase, IndianRupee, AlertTriangle, TrendingUp, TrendingDown, Shield, Clock, CheckCircle, ArrowRight, UserPlus, Star } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	const stats = {
		totalUsers: 12485,
		usersGrowth: 12.3,
		totalProviders: 3240,
		providersGrowth: 8.1,
		totalJobs: 28450,
		jobsGrowth: 15.6,
		revenue: 4250000,
		revenueGrowth: 18.2,
		openDisputes: 14,
		pendingKYC: 23,
		activeJobs: 342,
		completionRate: 94.7
	};

	const revenueChart = [
		{ month: 'Oct', amount: 2800000 },
		{ month: 'Nov', amount: 3100000 },
		{ month: 'Dec', amount: 3450000 },
		{ month: 'Jan', amount: 3200000 },
		{ month: 'Feb', amount: 3800000 },
		{ month: 'Mar', amount: 4250000 }
	];
	const maxRevenue = Math.max(...revenueChart.map((d) => d.amount));

	const pendingVerifications = [
		{ id: '1', name: 'Vikram Singh', category: 'Plumbing', submittedAt: '2026-03-16', documents: 3 },
		{ id: '2', name: 'Fatima Begum', category: 'Cleaning', submittedAt: '2026-03-15', documents: 4 },
		{ id: '3', name: 'Ravi Shankar', category: 'Electrical', submittedAt: '2026-03-15', documents: 2 },
		{ id: '4', name: 'Meena Devi', category: 'Gardening', submittedAt: '2026-03-14', documents: 3 }
	];

	const recentDisputes = [
		{ id: '1', title: 'Incomplete work complaint', customer: 'Amit Verma', provider: 'Suresh Nair', severity: 'high', createdAt: '2026-03-16' },
		{ id: '2', title: 'Payment not received', customer: 'Priya Menon', provider: 'Deepak Kumar', severity: 'medium', createdAt: '2026-03-15' },
		{ id: '3', title: 'Late arrival issue', customer: 'Arjun Das', provider: 'Lakshmi Bai', severity: 'low', createdAt: '2026-03-14' }
	];

	const recentActivity = [
		{ id: '1', action: 'New user registered', detail: 'Anita Sharma joined as customer', time: '5 min ago', icon: UserPlus },
		{ id: '2', action: 'KYC approved', detail: 'Mohan Rao verified as electrician', time: '12 min ago', icon: Shield },
		{ id: '3', action: 'Dispute resolved', detail: 'Case #D-1042 closed with refund', time: '25 min ago', icon: CheckCircle },
		{ id: '4', action: 'New provider', detail: 'Sanjay Patel registered for plumbing', time: '1 hour ago', icon: UserPlus },
		{ id: '5', action: '5-star review', detail: 'Priya rated Suresh Nair 5 stars', time: '2 hours ago', icon: Star },
		{ id: '6', action: 'Job completed', detail: 'Kitchen plumbing fix completed', time: '3 hours ago', icon: CheckCircle }
	];

	const severityBadge: Record<string, 'danger' | 'warning' | 'info'> = {
		high: 'danger', medium: 'warning', low: 'info'
	};
</script>

<svelte:head>
	<title>Admin Dashboard - Seva</title>
</svelte:head>

<div class="px-6 py-8 lg:px-8">
	<div class="flex flex-wrap items-center justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Admin Dashboard</h1>
			<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">Platform overview and management.</p>
		</div>
		<div class="flex items-center gap-3">
			{#if stats.pendingKYC > 0}
				<Badge variant="warning">{stats.pendingKYC} pending KYC</Badge>
			{/if}
			{#if stats.openDisputes > 0}
				<Badge variant="danger">{stats.openDisputes} open disputes</Badge>
			{/if}
		</div>
	</div>

	<!-- Stats Grid -->
	<div class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
		<Card>
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-3">
					<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-primary-100 text-primary-600 dark:bg-primary-900/30">
						<Users class="h-5 w-5" />
					</div>
					<span class="text-sm font-medium text-gray-600 dark:text-gray-400">Total Users</span>
				</div>
				<div class="flex items-center gap-1 text-secondary-600 dark:text-secondary-400">
					<TrendingUp class="h-3 w-3" />
					<span class="text-xs">+{stats.usersGrowth}%</span>
				</div>
			</div>
			<p class="mt-3 text-3xl font-bold text-gray-900 dark:text-white">{stats.totalUsers.toLocaleString()}</p>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{stats.totalProviders.toLocaleString()} providers</p>
		</Card>
		<Card>
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-3">
					<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100 text-blue-600 dark:bg-blue-900/30">
						<Briefcase class="h-5 w-5" />
					</div>
					<span class="text-sm font-medium text-gray-600 dark:text-gray-400">Total Jobs</span>
				</div>
				<div class="flex items-center gap-1 text-secondary-600 dark:text-secondary-400">
					<TrendingUp class="h-3 w-3" />
					<span class="text-xs">+{stats.jobsGrowth}%</span>
				</div>
			</div>
			<p class="mt-3 text-3xl font-bold text-gray-900 dark:text-white">{stats.totalJobs.toLocaleString()}</p>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{stats.activeJobs} active now</p>
		</Card>
		<Card>
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-3">
					<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-secondary-100 text-secondary-600 dark:bg-secondary-900/30">
						<IndianRupee class="h-5 w-5" />
					</div>
					<span class="text-sm font-medium text-gray-600 dark:text-gray-400">Revenue</span>
				</div>
				<div class="flex items-center gap-1 text-secondary-600 dark:text-secondary-400">
					<TrendingUp class="h-3 w-3" />
					<span class="text-xs">+{stats.revenueGrowth}%</span>
				</div>
			</div>
			<p class="mt-3 text-3xl font-bold text-gray-900 dark:text-white">Rs. {(stats.revenue / 100000).toFixed(1)}L</p>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">this month</p>
		</Card>
		<Card>
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-3">
					<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-red-100 text-red-600 dark:bg-red-900/30">
						<AlertTriangle class="h-5 w-5" />
					</div>
					<span class="text-sm font-medium text-gray-600 dark:text-gray-400">Disputes</span>
				</div>
			</div>
			<p class="mt-3 text-3xl font-bold text-gray-900 dark:text-white">{stats.openDisputes}</p>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{stats.completionRate}% completion rate</p>
		</Card>
	</div>

	<!-- Revenue Chart + Activity -->
	<div class="mt-8 grid gap-6 lg:grid-cols-3">
		<!-- Revenue Chart -->
		<Card class="lg:col-span-2">
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Revenue Trend</h2>
			<div class="mt-6 flex items-end gap-4 h-48">
				{#each revenueChart as bar}
					{@const height = (bar.amount / maxRevenue) * 100}
					<div class="flex flex-1 flex-col items-center gap-2">
						<span class="text-xs font-medium text-gray-700 dark:text-gray-300">
							Rs. {(bar.amount / 100000).toFixed(1)}L
						</span>
						<div class="w-full rounded-t-lg bg-primary-500 dark:bg-primary-600 transition-all" style="height: {height}%"></div>
						<span class="text-xs text-gray-500 dark:text-gray-400">{bar.month}</span>
					</div>
				{/each}
			</div>
		</Card>

		<!-- Recent Activity -->
		<Card>
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Activity Feed</h2>
			<div class="mt-4 space-y-4">
				{#each recentActivity as activity}
					{@const Icon = activity.icon}
					<div class="flex items-start gap-3">
						<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400">
							<Icon class="h-4 w-4" />
						</div>
						<div class="flex-1 min-w-0">
							<p class="text-sm font-medium text-gray-900 dark:text-white">{activity.action}</p>
							<p class="text-xs text-gray-500 dark:text-gray-400 truncate">{activity.detail}</p>
							<p class="text-xs text-gray-400 dark:text-gray-500">{activity.time}</p>
						</div>
					</div>
				{/each}
			</div>
		</Card>
	</div>

	<!-- Pending Actions -->
	<div class="mt-8 grid gap-6 lg:grid-cols-2">
		<!-- Pending Verifications -->
		<Card>
			<div class="flex items-center justify-between">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Pending Verifications</h2>
				<a href="/admin/kyc" class="flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700">
					View all <ArrowRight class="h-4 w-4" />
				</a>
			</div>
			<div class="mt-4 space-y-3">
				{#each pendingVerifications as item}
					<div class="flex items-center gap-3 rounded-lg border border-gray-100 p-3 dark:border-gray-700">
						<Avatar name={item.name} size="sm" />
						<div class="flex-1 min-w-0">
							<p class="text-sm font-medium text-gray-900 dark:text-white">{item.name}</p>
							<p class="text-xs text-gray-500 dark:text-gray-400">{item.category} -- {item.documents} documents</p>
						</div>
						<div class="text-right">
							<p class="text-xs text-gray-500 dark:text-gray-400">{item.submittedAt}</p>
							<Badge variant="warning" size="sm">Pending</Badge>
						</div>
					</div>
				{/each}
			</div>
		</Card>

		<!-- Recent Disputes -->
		<Card>
			<div class="flex items-center justify-between">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Recent Disputes</h2>
				<a href="/admin/disputes" class="flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700">
					View all <ArrowRight class="h-4 w-4" />
				</a>
			</div>
			<div class="mt-4 space-y-3">
				{#each recentDisputes as dispute}
					<div class="flex items-center gap-3 rounded-lg border border-gray-100 p-3 dark:border-gray-700">
						<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-red-100 text-red-500 dark:bg-red-900/20">
							<AlertTriangle class="h-5 w-5" />
						</div>
						<div class="flex-1 min-w-0">
							<p class="text-sm font-medium text-gray-900 dark:text-white truncate">{dispute.title}</p>
							<p class="text-xs text-gray-500 dark:text-gray-400">{dispute.customer} vs {dispute.provider}</p>
						</div>
						<div class="text-right">
							<Badge variant={severityBadge[dispute.severity] || 'info'} size="sm">{dispute.severity}</Badge>
							<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{dispute.createdAt}</p>
						</div>
					</div>
				{/each}
			</div>
		</Card>
	</div>
</div>
