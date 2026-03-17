<script lang="ts">
	import { onMount } from 'svelte';
	import { Users, Briefcase, IndianRupee, AlertTriangle, TrendingUp, TrendingDown, Shield, Clock, CheckCircle, ArrowRight, UserPlus, Star, Loader2 } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import api from '$lib/api/client';

	let loading = $state(true);
	let error = $state('');

	let stats = $state({
		totalUsers: 0,
		usersGrowth: 0,
		totalProviders: 0,
		providersGrowth: 0,
		totalJobs: 0,
		jobsGrowth: 0,
		revenue: 0,
		revenueGrowth: 0,
		openDisputes: 0,
		pendingKYC: 0,
		activeJobs: 0,
		completionRate: 0
	});

	let revenueChart = $state<{ month: string; amount: number }[]>([]);
	let pendingVerifications = $state<any[]>([]);
	let recentDisputes = $state<any[]>([]);
	let recentActivity = $state<any[]>([]);

	let maxRevenue = $derived(Math.max(...revenueChart.map((d) => d.amount), 1));

	onMount(async () => {
		try {
			const [statsRes, kycRes, disputesRes] = await Promise.all([
				api.admin.getStats(),
				api.admin.pendingKYC({ per_page: 4 }),
				api.disputes.list({ status: 'open', per_page: 3 })
			]);

			const s = statsRes.data;
			stats = {
				totalUsers: s.total_users || 0,
				usersGrowth: s.users_growth || 0,
				totalProviders: s.total_providers || 0,
				providersGrowth: s.providers_growth || 0,
				totalJobs: s.total_jobs || 0,
				jobsGrowth: s.jobs_growth || 0,
				revenue: s.revenue || 0,
				revenueGrowth: s.revenue_growth || 0,
				openDisputes: s.open_disputes || 0,
				pendingKYC: s.pending_kyc || 0,
				activeJobs: s.active_jobs || 0,
				completionRate: s.completion_rate || 0
			};

			revenueChart = (s.revenue_chart || s.revenue_trend || []).map((item: any) => ({
				month: item.month || item.label || '',
				amount: item.amount || item.revenue || 0
			}));

			recentActivity = (s.recent_activity || []).map((a: any) => {
				const iconMap: Record<string, any> = {
					user_registered: UserPlus,
					new_user: UserPlus,
					kyc_approved: Shield,
					dispute_resolved: CheckCircle,
					new_provider: UserPlus,
					review: Star,
					job_completed: CheckCircle
				};
				return {
					id: a.id,
					action: a.action || a.title || '',
					detail: a.detail || a.description || '',
					time: a.time || a.created_at || '',
					icon: iconMap[a.type] || CheckCircle
				};
			});

			pendingVerifications = (kycRes.data || []).map((k: any) => ({
				id: k.id,
				name: k.user?.name || k.provider_name || '',
				category: k.category?.name || k.category || '',
				submittedAt: k.created_at?.split('T')[0] || k.submitted_at || '',
				documents: k.documents_count || k.documents?.length || 0
			}));

			recentDisputes = (disputesRes.data || []).map((d: any) => ({
				id: d.id,
				title: d.title || d.job?.title || 'Dispute',
				customer: d.customer?.name || '',
				provider: d.provider?.user?.name || '',
				severity: d.severity || 'medium',
				createdAt: d.created_at?.split('T')[0] || ''
			}));
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load admin dashboard';
		} finally {
			loading = false;
		}
	});

	const severityBadge: Record<string, 'danger' | 'warning' | 'info'> = {
		high: 'danger', medium: 'warning', low: 'info'
	};
</script>

<svelte:head>
	<title>Admin Dashboard - Seva</title>
</svelte:head>

{#if loading}
<div class="flex items-center justify-center py-20">
	<Loader2 class="h-8 w-8 animate-spin text-primary-600" />
</div>
{:else if error}
<div class="px-6 py-8 lg:px-8">
	<div class="rounded-lg border border-red-200 bg-red-50 p-6 text-center dark:border-red-800 dark:bg-red-900/20">
		<p class="text-red-600 dark:text-red-400">{error}</p>
	</div>
</div>
{:else}
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
{/if}
