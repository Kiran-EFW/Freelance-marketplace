<script lang="ts">
	import { Search, Filter, Plus } from 'lucide-svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import JobCard from '$lib/components/job/JobCard.svelte';
	import Pagination from '$lib/components/ui/Pagination.svelte';
	import Tabs from '$lib/components/ui/Tabs.svelte';

	let searchQuery = $state('');
	let statusFilter = $state('all');
	let currentPage = $state(1);

	const statusTabs = [
		{ id: 'all', label: 'All' },
		{ id: 'posted', label: 'Posted' },
		{ id: 'quoted', label: 'Quoted' },
		{ id: 'in_progress', label: 'In Progress' },
		{ id: 'completed', label: 'Completed' }
	];

	const mockJobs = [
		{
			id: '1', title: 'Fix kitchen plumbing', description: 'Leaking tap and clogged drain in kitchen. Need an experienced plumber for repair.',
			category: 'Plumbing', status: 'in_progress', budget: 3500,
			createdAt: '2026-03-15', quotesCount: 4,
			images: []
		},
		{
			id: '2', title: 'Living room painting', description: 'Full living room painting with premium paint. Approximately 400 sq ft area.',
			category: 'Painting', status: 'quoted', budget: 8000,
			createdAt: '2026-03-14', quotesCount: 6,
			images: []
		},
		{
			id: '3', title: 'AC servicing - 3 units', description: 'Annual maintenance service for 3 split AC units. Include gas refill if needed.',
			category: 'HVAC', status: 'completed', budget: 4500,
			createdAt: '2026-03-10', quotesCount: 3,
			images: []
		},
		{
			id: '4', title: 'Garden maintenance', description: 'Monthly garden maintenance - trimming, weeding, watering, and fertilizing.',
			category: 'Gardening', status: 'posted', budget: 1500,
			createdAt: '2026-03-13', quotesCount: 2,
			images: []
		},
		{
			id: '5', title: 'Bathroom renovation', description: 'Complete bathroom renovation including tiling, plumbing, and fixtures.',
			category: 'Plumbing', status: 'posted', budget: 25000,
			createdAt: '2026-03-12', quotesCount: 0,
			images: []
		},
		{
			id: '6', title: 'Electrical wiring check', description: 'Full house electrical inspection and wiring check. 3BHK apartment.',
			category: 'Electrical', status: 'quoted', budget: 3000,
			createdAt: '2026-03-11', quotesCount: 5,
			images: []
		},
		{
			id: '7', title: 'Deep cleaning - 2BHK', description: 'Move-in deep cleaning for a 2BHK apartment. Kitchen, bathrooms, and all rooms.',
			category: 'Cleaning', status: 'completed', budget: 2500,
			createdAt: '2026-03-08', quotesCount: 4,
			images: []
		},
		{
			id: '8', title: 'Custom bookshelf installation', description: 'Design and install a custom bookshelf in the study room. Need quality wood.',
			category: 'Carpentry', status: 'completed', budget: 12000,
			createdAt: '2026-03-05', quotesCount: 3,
			images: []
		}
	];

	let filteredJobs = $derived(
		mockJobs.filter((j) => {
			const matchesSearch = !searchQuery ||
				j.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
				j.category.toLowerCase().includes(searchQuery.toLowerCase());
			const matchesStatus = statusFilter === 'all' || j.status === statusFilter;
			return matchesSearch && matchesStatus;
		})
	);

	function getTimeAgo(dateStr: string): string {
		const diff = new Date('2026-03-17').getTime() - new Date(dateStr).getTime();
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
	<div class="mt-2 text-sm text-gray-500 dark:text-gray-400">{filteredJobs.length} jobs</div>

	<!-- Jobs List -->
	{#if filteredJobs.length > 0}
		<div class="mt-4 space-y-4">
			{#each filteredJobs as job}
				<JobCard job={{ id: job.id, title: job.title, description: '', status: job.status, category: { id: '', name: job.category, slug: job.category.toLowerCase() }, budget_min: job.budget, quotes_count: job.quotesCount, images: job.images, created_at: job.createdAt } as any} />
			{/each}
		</div>
		<div class="mt-8">
			<Pagination {currentPage} totalPages={2} onPageChange={(p) => (currentPage = p)} />
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
</div>
