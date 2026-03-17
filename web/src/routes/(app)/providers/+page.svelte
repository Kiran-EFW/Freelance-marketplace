<script lang="ts">
	import { Search, MapPin, SlidersHorizontal, X, Loader2 } from 'lucide-svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import ProviderCard from '$lib/components/provider/ProviderCard.svelte';
	import Pagination from '$lib/components/ui/Pagination.svelte';
	import api from '$lib/api/client';

	let searchQuery = $state('');
	let categoryFilter = $state('');
	let locationFilter = $state('');
	let sortBy = $state('rating');
	let showFilters = $state(false);
	let currentPage = $state(1);
	let loading = $state(true);
	let error = $state('');
	let providersList = $state<any[]>([]);
	let totalPages = $state(1);

	const categories = [
		'Plumbing', 'Electrical', 'Cleaning', 'Gardening',
		'Painting', 'Moving', 'Carpentry', 'HVAC',
		'Pest Control', 'Appliance Repair'
	];

	async function fetchProviders() {
		loading = true;
		error = '';
		try {
			const params: any = {
				page: currentPage,
				per_page: 9,
				sort_by: sortBy
			};
			if (searchQuery.trim()) params.query = searchQuery.trim();
			if (categoryFilter) params.category = categoryFilter;
			if (locationFilter) params.postcode = locationFilter;

			const res = await api.providers.search(params);
			providersList = res.data || [];
			totalPages = res.meta?.total_pages || Math.ceil((res.meta?.total || 1) / 9);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load providers';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		const _q = searchQuery;
		const _c = categoryFilter;
		const _l = locationFilter;
		const _s = sortBy;
		const _p = currentPage;
		fetchProviders();
	});

	let filteredProviders = $derived(providersList);

	const activeFilters = $derived(
		[categoryFilter, locationFilter].filter(Boolean).length
	);
</script>

<svelte:head>
	<title>Find Providers - Seva</title>
</svelte:head>

<div class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
	<div>
		<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Find Service Providers</h1>
		<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
			Browse verified professionals in your area.
		</p>
	</div>

	<!-- Search & Filters -->
	<div class="mt-6 flex flex-col gap-4 sm:flex-row">
		<div class="relative flex-1">
			<Search class="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-gray-400" />
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Search providers or services..."
				class="w-full rounded-lg border border-gray-300 py-2.5 pl-10 pr-4 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
			/>
		</div>
		<div class="relative">
			<MapPin class="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-gray-400" />
			<input
				type="text"
				bind:value={locationFilter}
				placeholder="Postcode"
				class="w-full rounded-lg border border-gray-300 py-2.5 pl-10 pr-4 text-sm focus:border-primary-500 focus:outline-none sm:w-40 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
			/>
		</div>
		<select
			bind:value={categoryFilter}
			class="rounded-lg border border-gray-300 px-3 py-2.5 text-sm focus:border-primary-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-white"
		>
			<option value="">All Categories</option>
			{#each categories as cat}
				<option value={cat}>{cat}</option>
			{/each}
		</select>
		<select
			bind:value={sortBy}
			class="rounded-lg border border-gray-300 px-3 py-2.5 text-sm focus:border-primary-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-white"
		>
			<option value="rating">Highest Rated</option>
			<option value="reviews">Most Reviews</option>
			<option value="price_low">Price: Low to High</option>
			<option value="price_high">Price: High to Low</option>
		</select>
	</div>

	<!-- Active Filters -->
	{#if activeFilters > 0}
		<div class="mt-3 flex items-center gap-2">
			<span class="text-xs text-gray-500 dark:text-gray-400">Active filters:</span>
			{#if categoryFilter}
				<button onclick={() => (categoryFilter = '')} class="flex items-center gap-1 rounded-full bg-primary-100 px-3 py-1 text-xs text-primary-700 dark:bg-primary-900/30 dark:text-primary-400">
					{categoryFilter}
					<X class="h-3 w-3" />
				</button>
			{/if}
			{#if locationFilter}
				<button onclick={() => (locationFilter = '')} class="flex items-center gap-1 rounded-full bg-primary-100 px-3 py-1 text-xs text-primary-700 dark:bg-primary-900/30 dark:text-primary-400">
					{locationFilter}
					<X class="h-3 w-3" />
				</button>
			{/if}
		</div>
	{/if}

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
		<div class="mt-4 flex items-center justify-between">
			<p class="text-sm text-gray-500 dark:text-gray-400">{filteredProviders.length} providers found</p>
		</div>

		<!-- Provider Grid -->
		{#if filteredProviders.length > 0}
			<div class="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
				{#each filteredProviders as provider}
					<ProviderCard {provider} />
				{/each}
			</div>
			<div class="mt-8">
				<Pagination {currentPage} {totalPages} onPageChange={(p) => (currentPage = p)} />
			</div>
		{:else}
			<div class="mt-8 rounded-lg border border-gray-200 bg-white p-12 text-center dark:border-gray-700 dark:bg-gray-800">
				<Search class="mx-auto h-12 w-12 text-gray-300 dark:text-gray-600" />
				<h3 class="mt-4 text-lg font-medium text-gray-900 dark:text-white">No providers found</h3>
				<p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
					Try adjusting your search or filters to find providers.
				</p>
			</div>
		{/if}
	{/if}
</div>
