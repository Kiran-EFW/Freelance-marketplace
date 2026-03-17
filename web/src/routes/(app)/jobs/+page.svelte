<script lang="ts">
	import { onMount } from 'svelte';
	import { Search, Filter, Plus, Loader2 } from 'lucide-svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import JobCard from '$lib/components/job/JobCard.svelte';
	import Pagination from '$lib/components/ui/Pagination.svelte';
	import Tabs from '$lib/components/ui/Tabs.svelte';
	import api from '$lib/api/client';

	let searchQuery = $state('');
	let statusFilter = $state('all');
	let currentPage = $state(1);
	let loading = $state(true);
	let error = $state('');
	let allJobs = $state<any[]>([]);
	let totalPages = $state(1);

	const statusTabs = [
		{ id: 'all', label: 'All' },
		{ id: 'open', label: 'Posted' },
		{ id: 'quoted', label: 'Quoted' },
		{ id: 'in_progress', label: 'In Progress' },
		{ id: 'completed', label: 'Completed' }
	];

	async function fetchJobs() {
		loading = true;
		error = '';
		try {
			const params: any = {
				page: currentPage,
				per_page: 10
			};
			if (statusFilter !== 'all') params.status = statusFilter;
			if (searchQuery.trim()) params.search = searchQuery.trim();

			const res = await api.jobs.list(params);
			allJobs = res.data || [];
			totalPages = res.meta?.total_pages || Math.ceil((res.meta?.total || 1) / 10);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load jobs';
		} finally {
			loading = false;
		}
	}

	// Fetch on mount and re-fetch when filters change
	$effect(() => {
		// Access reactive values to track them
		const _status = statusFilter;
		const _page = currentPage;
		const _search = searchQuery;
		fetchJobs();
	});

	let filteredJobs = $derived(allJobs);

	function getTimeAgo(dateStr: string): string {
		const diff = Date.now() - new Date(dateStr).getTime();
		const days = Math.floor(diff / (1000 * 60 * 60 * 24));
		if (days === 0) return 'Today';
		if (days === 1) return '1 day ago';
		if (days < 7) return `${days} days ago`;
		if (days < 30) return `${Math.floor(days / 7)} weeks ago`;
		return `${Math.floor(days / 30)} months ago`;
	}
</script>

<svelte:head>
	<title>Jobs - Seva</title>
</svelte:head>

<div class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
	<div class="flex flex-wrap items-center justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">My Jobs</h1>
			<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">Manage your service requests and track progress.</p>
		</div>
		<Button variant="primary" href="/jobs/new">
			<Plus class="h-4 w-4" />
			Post a Job
		</Button>
	</div>

	<!-- Filters -->
	<div class="mt-6 flex flex-col gap-4 sm:flex-row sm:items-center">
		<div class="relative flex-1">
			<Search class="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-gray-400" />
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Search jobs..."
				class="w-full rounded-lg border border-gray-300 py-2.5 pl-10 pr-4 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
			/>
		</div>
	</div>

	<!-- Status Tabs -->
	<div class="mt-4">
		<Tabs tabs={statusTabs} bind:activeTab={statusFilter} />
	</div>

	<!-- Results -->
	{#if loading}
		<div class="mt-8 flex items-center justify-center py-12">
			<Loader2 class="h-8 w-8 animate-spin text-primary-600" />
		</div>
	{:else if error}
		<div class="mt-4 rounded-lg border border-red-200 bg-red-50 p-6 text-center dark:border-red-800 dark:bg-red-900/20">
			<p class="text-red-600 dark:text-red-400">{error}</p>
		</div>
	{:else}
		<div class="mt-2 text-sm text-gray-500 dark:text-gray-400">{filteredJobs.length} jobs</div>

		<!-- Jobs List -->
		{#if filteredJobs.length > 0}
			<div class="mt-4 space-y-4">
				{#each filteredJobs as job}
					<JobCard {job} />
				{/each}
			</div>
			<div class="mt-8">
				<Pagination {currentPage} {totalPages} onPageChange={(p) => (currentPage = p)} />
			</div>
		{:else}
			<div class="mt-8 rounded-lg border border-gray-200 bg-white p-12 text-center dark:border-gray-700 dark:bg-gray-800">
				<Search class="mx-auto h-12 w-12 text-gray-300 dark:text-gray-600" />
				<h3 class="mt-4 text-lg font-medium text-gray-900 dark:text-white">No jobs found</h3>
				<p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
					{#if statusFilter !== 'all'}
						No jobs with status "{statusFilter}". Try a different filter.
					{:else}
						Post your first job to get started finding service providers.
					{/if}
				</p>
				<div class="mt-4">
					<Button variant="primary" href="/jobs/new">Post a Job</Button>
				</div>
			</div>
		{/if}
	{/if}
</div>
