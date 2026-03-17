<script lang="ts">
	import { Search, MapPin, SlidersHorizontal, X } from 'lucide-svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import ProviderCard from '$lib/components/provider/ProviderCard.svelte';
	import Pagination from '$lib/components/ui/Pagination.svelte';

	let searchQuery = $state('');
	let categoryFilter = $state('');
	let locationFilter = $state('');
	let sortBy = $state('rating');
	let showFilters = $state(false);
	let currentPage = $state(1);

	const categories = [
		'Plumbing', 'Electrical', 'Cleaning', 'Gardening',
		'Painting', 'Moving', 'Carpentry', 'HVAC',
		'Pest Control', 'Appliance Repair'
	];

	const mockProviders = [
		{
			id: '1', name: 'Suresh Nair', rating: 4.8, reviewCount: 156,
			skills: ['Plumbing', 'Pipe Fitting', 'Water Heater'],
			hourlyRate: 500, distance: '1.2 km', isVerified: true,
			bio: 'Experienced plumber with 10+ years. Specializing in residential plumbing, water heater installation, and emergency repairs.',
			completedJobs: 156, responseTime: '30 min'
		},
		{
			id: '2', name: 'Lakshmi Bai', rating: 4.9, reviewCount: 210,
			skills: ['Deep Cleaning', 'Sanitization', 'Move-in/out'],
			hourlyRate: 400, distance: '2.1 km', isVerified: true,
			bio: 'Professional cleaning expert. Homes, offices, and post-construction cleanup.',
			completedJobs: 210, responseTime: '1 hour'
		},
		{
			id: '3', name: 'Deepak Kumar', rating: 4.5, reviewCount: 78,
			skills: ['Electrical', 'Wiring', 'Panel Installation'],
			hourlyRate: 550, distance: '3.5 km', isVerified: true,
			bio: 'Licensed electrician for all residential and commercial needs.',
			completedJobs: 78, responseTime: '2 hours'
		},
		{
			id: '4', name: 'Mohan Rao', rating: 4.7, reviewCount: 142,
			skills: ['Painting', 'Wall Finishing', 'Waterproofing'],
			hourlyRate: 450, distance: '1.8 km', isVerified: true,
			bio: 'Expert painter with premium finishes. Interior and exterior painting services.',
			completedJobs: 142, responseTime: '45 min'
		},
		{
			id: '5', name: 'Priya Sharma', rating: 4.6, reviewCount: 88,
			skills: ['Gardening', 'Landscaping', 'Tree Trimming'],
			hourlyRate: 350, distance: '4.2 km', isVerified: false,
			bio: 'Garden design and maintenance professional. Specializing in tropical plants.',
			completedJobs: 88, responseTime: '3 hours'
		},
		{
			id: '6', name: 'Kiran Rao', rating: 4.4, reviewCount: 65,
			skills: ['Carpentry', 'Furniture Repair', 'Custom Shelving'],
			hourlyRate: 600, distance: '2.8 km', isVerified: true,
			bio: 'Custom woodwork and furniture repairs. Fine carpentry with quality materials.',
			completedJobs: 65, responseTime: '1 hour'
		},
		{
			id: '7', name: 'Rajesh Iyer', rating: 4.3, reviewCount: 45,
			skills: ['HVAC', 'AC Servicing', 'Installation'],
			hourlyRate: 700, distance: '5.0 km', isVerified: true,
			bio: 'Air conditioning installation, servicing, and repair for all brands.',
			completedJobs: 45, responseTime: '2 hours'
		},
		{
			id: '8', name: 'Anita Gupta', rating: 4.8, reviewCount: 120,
			skills: ['Pest Control', 'Termite Treatment', 'Fumigation'],
			hourlyRate: 300, distance: '1.5 km', isVerified: true,
			bio: 'Eco-friendly pest control solutions. Licensed for all types of pest management.',
			completedJobs: 120, responseTime: '30 min'
		}
	];

	let filteredProviders = $derived(
		mockProviders
			.filter((p) => {
				const matchesSearch = !searchQuery ||
					p.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
					p.skills.some((s) => s.toLowerCase().includes(searchQuery.toLowerCase()));
				const matchesCategory = !categoryFilter ||
					p.skills.some((s) => s.toLowerCase().includes(categoryFilter.toLowerCase()));
				return matchesSearch && matchesCategory;
			})
			.sort((a, b) => {
				if (sortBy === 'rating') return b.rating - a.rating;
				if (sortBy === 'price_low') return a.hourlyRate - b.hourlyRate;
				if (sortBy === 'price_high') return b.hourlyRate - a.hourlyRate;
				if (sortBy === 'reviews') return b.reviewCount - a.reviewCount;
				return 0;
			})
	);

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

	<!-- Results Count -->
	<div class="mt-4 flex items-center justify-between">
		<p class="text-sm text-gray-500 dark:text-gray-400">{filteredProviders.length} providers found</p>
	</div>

	<!-- Provider Grid -->
	{#if filteredProviders.length > 0}
		<div class="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each filteredProviders as provider}
				<ProviderCard provider={{ id: provider.id, business_name: provider.name, rating_average: provider.rating, rating_count: provider.reviewCount, categories: provider.skills.map((s: string) => ({ id: '', name: s, slug: s.toLowerCase() })), hourly_rate: provider.hourlyRate, verification_status: provider.isVerified ? 'approved' : 'pending' } as any} distance={provider.distance} />
			{/each}
		</div>
		<div class="mt-8">
			<Pagination {currentPage} totalPages={3} onPageChange={(p) => (currentPage = p)} />
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
</div>
