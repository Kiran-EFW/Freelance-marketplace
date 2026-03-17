<script lang="ts">
	import { onMount } from 'svelte';
	import { MapPin, List, X, ChevronRight, Filter, Loader2 } from 'lucide-svelte';
	import MapView from '$lib/components/MapView.svelte';
	import { t } from '$lib/i18n/index.svelte';
	import api from '$lib/api/client';
	import type { ProviderProfile, Job, Category } from '$lib/types';

	interface MapMarker {
		id: string;
		lat: number;
		lng: number;
		title: string;
		category?: string;
		type: 'job' | 'provider';
		link?: string;
	}

	// State
	let markers = $state<MapMarker[]>([]);
	let loading = $state(false);
	let viewMode = $state<'providers' | 'jobs'>('providers');
	let selectedCategory = $state('');
	let searchRadius = $state(15);
	let categories = $state<Category[]>([]);
	let sidebarOpen = $state(true);
	let selectedMarkerId = $state<string | null>(null);
	let userLocation = $state({ lat: 12.9716, lng: 77.5946 });
	let locationDetected = $state(false);
	let mapComponent: MapView;

	// Derived
	let visibleMarkers = $derived(markers);

	async function loadCategories() {
		try {
			const res = await api.categories.list();
			categories = res.data;
		} catch {
			// Categories may not be available yet
			categories = [];
		}
	}

	async function loadMarkers() {
		loading = true;
		markers = [];

		try {
			if (viewMode === 'providers') {
				const params: Record<string, string | number | boolean | undefined> = {
					page: 1,
					per_page: 50,
					radius_km: searchRadius
				};
				if (selectedCategory) params.category = selectedCategory;
				if (locationDetected) {
					params.postcode = ''; // Use postcode-less search by geo if API supports it
				}

				const res = await api.providers.search(params);
				markers = (res.data || [])
					.filter((p: ProviderProfile) => {
						return p.service_areas?.some((sa) => sa.latitude && sa.longitude);
					})
					.map((p: ProviderProfile) => {
						const area = p.service_areas.find((sa) => sa.latitude && sa.longitude)!;
						return {
							id: p.id,
							lat: area.latitude,
							lng: area.longitude,
							title: p.business_name || p.user?.name || 'Provider',
							category: p.categories?.[0]?.name,
							type: 'provider' as const,
							link: `/providers/${p.id}`
						};
					});
			} else {
				const params: Record<string, string | number | boolean | undefined> = {
					page: 1,
					per_page: 50
				};
				if (selectedCategory) params.category = selectedCategory;

				const res = await api.jobs.list(params);
				markers = (res.data || [])
					.filter((j: Job) => j.location?.latitude && j.location?.longitude)
					.map((j: Job) => ({
						id: j.id,
						lat: j.location.latitude!,
						lng: j.location.longitude!,
						title: j.title,
						category: j.category?.name,
						type: 'job' as const,
						link: `/jobs/${j.id}`
					}));
			}
		} catch (err) {
			console.error('Failed to load map data:', err);
			markers = [];
		} finally {
			loading = false;
		}
	}

	function detectLocation() {
		if (!navigator.geolocation) return;

		navigator.geolocation.getCurrentPosition(
			(position) => {
				userLocation = {
					lat: position.coords.latitude,
					lng: position.coords.longitude
				};
				locationDetected = true;
				loadMarkers();
			},
			() => {
				// Use default Bangalore center on error
				locationDetected = false;
				loadMarkers();
			},
			{ enableHighAccuracy: true, timeout: 10000 }
		);
	}

	function handleMarkerClick(marker: MapMarker) {
		selectedMarkerId = marker.id;
		if (!sidebarOpen) sidebarOpen = true;
	}

	function handleListItemClick(marker: MapMarker) {
		selectedMarkerId = marker.id;
		mapComponent?.highlightMarker(marker.id);
	}

	function handleViewModeChange(mode: 'providers' | 'jobs') {
		viewMode = mode;
		selectedMarkerId = null;
		loadMarkers();
	}

	function handleCategoryChange(e: Event) {
		selectedCategory = (e.target as HTMLSelectElement).value;
		loadMarkers();
	}

	function handleRadiusChange(e: Event) {
		searchRadius = Number((e.target as HTMLInputElement).value);
	}

	function handleRadiusCommit() {
		loadMarkers();
	}

	onMount(() => {
		loadCategories();
		detectLocation();
	});
</script>

<svelte:head>
	<title>{t('map.title')} | Seva</title>
</svelte:head>

<div class="flex h-[calc(100vh-64px)] flex-col">
	<!-- Filter Bar -->
	<div class="flex flex-wrap items-center gap-3 border-b border-gray-200 bg-white px-4 py-3 dark:border-gray-800 dark:bg-gray-950">
		<!-- View Mode Toggle -->
		<div class="flex rounded-lg border border-gray-200 dark:border-gray-700">
			<button
				class="flex items-center gap-1.5 rounded-l-lg px-3 py-1.5 text-sm font-medium transition-colors {viewMode === 'providers'
					? 'bg-green-50 text-green-700 dark:bg-green-900/30 dark:text-green-400'
					: 'text-gray-600 hover:bg-gray-50 dark:text-gray-400 dark:hover:bg-gray-800'}"
				onclick={() => handleViewModeChange('providers')}
			>
				<MapPin class="h-4 w-4" />
				{t('map.show_providers')}
			</button>
			<button
				class="flex items-center gap-1.5 rounded-r-lg border-l border-gray-200 px-3 py-1.5 text-sm font-medium transition-colors dark:border-gray-700 {viewMode === 'jobs'
					? 'bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
					: 'text-gray-600 hover:bg-gray-50 dark:text-gray-400 dark:hover:bg-gray-800'}"
				onclick={() => handleViewModeChange('jobs')}
			>
				<Filter class="h-4 w-4" />
				{t('map.show_jobs')}
			</button>
		</div>

		<!-- Category Filter -->
		<select
			class="rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-sm text-gray-700 focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300"
			value={selectedCategory}
			onchange={handleCategoryChange}
		>
			<option value="">{t('map.filter_category')}</option>
			{#each categories as cat}
				<option value={cat.id}>{cat.name}</option>
			{/each}
		</select>

		<!-- Radius Slider -->
		<div class="flex items-center gap-2">
			<label class="text-sm text-gray-600 dark:text-gray-400" for="radius-slider">
				{t('map.filter_radius')}:
			</label>
			<input
				id="radius-slider"
				type="range"
				min="1"
				max="50"
				step="1"
				value={searchRadius}
				oninput={handleRadiusChange}
				onchange={handleRadiusCommit}
				class="h-1.5 w-24 cursor-pointer appearance-none rounded-full bg-gray-200 accent-primary-600 dark:bg-gray-700"
			/>
			<span class="min-w-[3rem] text-sm font-medium text-gray-700 dark:text-gray-300">
				{searchRadius} km
			</span>
		</div>

		<!-- Marker count -->
		<div class="ml-auto text-sm text-gray-500 dark:text-gray-400">
			{#if loading}
				<Loader2 class="inline h-4 w-4 animate-spin" />
			{:else}
				{markers.length} {viewMode === 'providers' ? t('map.show_providers').toLowerCase() : t('map.show_jobs').toLowerCase()}
			{/if}
		</div>

		<!-- Sidebar Toggle -->
		<button
			class="rounded-lg p-1.5 text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800"
			onclick={() => (sidebarOpen = !sidebarOpen)}
			aria-label="Toggle sidebar"
		>
			{#if sidebarOpen}
				<X class="h-5 w-5" />
			{:else}
				<List class="h-5 w-5" />
			{/if}
		</button>
	</div>

	<!-- Map + Sidebar -->
	<div class="relative flex flex-1 overflow-hidden">
		<!-- Map -->
		<div class="flex-1">
			<MapView
				bind:this={mapComponent}
				{markers}
				center={userLocation}
				zoom={13}
				onmarkerclick={handleMarkerClick}
			/>
		</div>

		<!-- Sidebar -->
		{#if sidebarOpen}
			<div class="w-80 shrink-0 overflow-y-auto border-l border-gray-200 bg-white dark:border-gray-800 dark:bg-gray-950 max-md:absolute max-md:right-0 max-md:top-0 max-md:z-10 max-md:h-full max-md:shadow-lg">
				<div class="sticky top-0 z-10 flex items-center justify-between border-b border-gray-100 bg-white px-4 py-3 dark:border-gray-800 dark:bg-gray-950">
					<h3 class="text-sm font-semibold text-gray-900 dark:text-white">
						{viewMode === 'providers' ? t('map.show_providers') : t('map.show_jobs')}
						<span class="ml-1 text-gray-400">({visibleMarkers.length})</span>
					</h3>
					<button
						class="rounded p-1 text-gray-400 hover:bg-gray-100 md:hidden dark:hover:bg-gray-800"
						onclick={() => (sidebarOpen = false)}
					>
						<X class="h-4 w-4" />
					</button>
				</div>

				{#if loading}
					<div class="flex items-center justify-center py-12">
						<Loader2 class="h-6 w-6 animate-spin text-primary-600" />
					</div>
				{:else if visibleMarkers.length === 0}
					<div class="px-4 py-12 text-center">
						<MapPin class="mx-auto h-10 w-10 text-gray-300 dark:text-gray-600" />
						<p class="mt-2 text-sm text-gray-500 dark:text-gray-400">{t('map.no_results')}</p>
					</div>
				{:else}
					<ul class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each visibleMarkers as marker (marker.id)}
							<li>
								<button
									class="flex w-full items-start gap-3 px-4 py-3 text-left transition-colors hover:bg-gray-50 dark:hover:bg-gray-900 {selectedMarkerId === marker.id
										? 'bg-primary-50 dark:bg-primary-900/20'
										: ''}"
									onclick={() => handleListItemClick(marker)}
								>
									<div
										class="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full {marker.type === 'job'
											? 'bg-blue-100 text-blue-600 dark:bg-blue-900/40 dark:text-blue-400'
											: 'bg-green-100 text-green-600 dark:bg-green-900/40 dark:text-green-400'}"
									>
										<MapPin class="h-4 w-4" />
									</div>
									<div class="min-w-0 flex-1">
										<p class="truncate text-sm font-medium text-gray-900 dark:text-white">
											{marker.title}
										</p>
										{#if marker.category}
											<p class="mt-0.5 truncate text-xs text-gray-500 dark:text-gray-400">
												{marker.category}
											</p>
										{/if}
									</div>
									<ChevronRight class="mt-1 h-4 w-4 shrink-0 text-gray-300 dark:text-gray-600" />
								</button>
							</li>
						{/each}
					</ul>
				{/if}
			</div>
		{/if}
	</div>
</div>
