<script lang="ts">
	import { page } from '$app/stores';
	import { ArrowLeft, MapPin, User, TreePine, Calendar, ChevronUp, ChevronDown, Plus, Zap, Trash2, Clock, Navigation, Loader2 } from 'lucide-svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import RouteMap from '$lib/components/RouteMap.svelte';
	import { toastSuccess, toastError } from '$lib/stores/toast';
	import { t } from '$lib/i18n/index.svelte';
	import api from '$lib/api/client';

	let routeId = $derived($page.params.id);
	let loading = $state(true);
	let error = $state('');
	let showAddStopModal = $state(false);
	let newStopCustomer = $state('');
	let newStopAddress = $state('');
	let newStopPostcode = $state('');
	let newStopTrees = $state(0);
	let isOptimizing = $state(false);
	let showOptimized = $state(false);

	let route = $state({
		id: '', name: '', description: '',
		recurrence: 'weekly', nextVisit: '', isActive: true
	});

	type Stop = {
		id: string;
		order: number;
		customerName: string;
		address: string;
		postcode: string;
		lat: number;
		lng: number;
		treeCount: number;
		lastVisit: string;
		nextVisit: string;
		duration: number;
	};

	let stops = $state<Stop[]>([]);

	async function fetchRoute() {
		loading = true;
		error = '';
		try {
			const res = await api.routes.get(routeId);
			const data = res.data as any;
			route = {
				id: data.id,
				name: data.name || '',
				description: data.description || '',
				recurrence: data.recurrence || 'weekly',
				nextVisit: data.next_visit?.split('T')[0] || data.nextVisit || '',
				isActive: data.is_active !== undefined ? data.is_active : true
			};
			stops = (data.stops || []).map((s: any, i: number) => ({
				id: s.id,
				order: s.order || i + 1,
				customerName: s.customer_name || s.customerName || '',
				address: s.address || '',
				postcode: s.postcode || '',
				lat: s.lat || s.latitude || 0,
				lng: s.lng || s.longitude || 0,
				treeCount: s.tree_count || s.treeCount || 0,
				lastVisit: s.last_visit?.split('T')[0] || s.lastVisit || '',
				nextVisit: s.next_visit?.split('T')[0] || s.nextVisit || '',
				duration: s.duration || s.duration_minutes || 30
			}));
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load route';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		const _id = routeId;
		if (routeId) fetchRoute();
	});

	let optimizedStops = $state<Stop[]>([]);

	const mapStops = $derived(
		stops.map((s) => ({
			id: s.id,
			lat: s.lat,
			lng: s.lng,
			address: `${s.address}, ${s.postcode}`,
			customerName: s.customerName,
			order: s.order
		}))
	);

	const mapOptimizedStops = $derived(
		optimizedStops.map((s) => ({
			id: s.id,
			lat: s.lat,
			lng: s.lng,
			address: `${s.address}, ${s.postcode}`,
			customerName: s.customerName,
			order: s.order
		}))
	);

	// Calculate total distance using Haversine (same as backend)
	function haversineKM(lat1: number, lng1: number, lat2: number, lng2: number): number {
		const R = 6371;
		const dLat = ((lat2 - lat1) * Math.PI) / 180;
		const dLng = ((lng2 - lng1) * Math.PI) / 180;
		const a =
			Math.sin(dLat / 2) * Math.sin(dLat / 2) +
			Math.cos((lat1 * Math.PI) / 180) * Math.cos((lat2 * Math.PI) / 180) *
			Math.sin(dLng / 2) * Math.sin(dLng / 2);
		return R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
	}

	function calcTotalDistance(stopList: typeof stops): number {
		let total = 0;
		for (let i = 0; i < stopList.length - 1; i++) {
			total += haversineKM(stopList[i].lat, stopList[i].lng, stopList[i + 1].lat, stopList[i + 1].lng);
		}
		return total;
	}

	const totalDistance = $derived(calcTotalDistance(stops));
	const optimizedDistance = $derived(optimizedStops.length > 0 ? calcTotalDistance(optimizedStops) : 0);
	const distanceSaved = $derived(totalDistance - optimizedDistance);

	async function moveStop(index: number, direction: 'up' | 'down') {
		const newIndex = direction === 'up' ? index - 1 : index + 1;
		if (newIndex < 0 || newIndex >= stops.length) return;
		const newStops = [...stops];
		[newStops[index], newStops[newIndex]] = [newStops[newIndex], newStops[index]];
		newStops.forEach((s, i) => (s.order = i + 1));
		stops = newStops;
		showOptimized = false;
		optimizedStops = [];
		try {
			const stopIds = newStops.map((s) => s.id);
			await api.routes.reorderStops(routeId, stopIds);
		} catch (err) {
			// Revert on failure silently - next fetch will correct
		}
	}

	async function removeStop(id: string) {
		try {
			await api.routes.removeStop(routeId, id);
			stops = stops.filter((s) => s.id !== id);
			stops.forEach((s, i) => (s.order = i + 1));
			showOptimized = false;
			optimizedStops = [];
			toastSuccess(t('routes.remove_stop'));
		} catch (err) {
			toastError(err instanceof Error ? err.message : 'Failed to remove stop');
		}
	}

	async function optimizeRoute() {
		isOptimizing = true;
		try {
			const res = await api.routes.optimize(routeId);
			const data = res.data as any;
			if (data.stops && Array.isArray(data.stops)) {
				optimizedStops = data.stops.map((s: any, i: number) => ({
					id: s.id,
					order: s.order || i + 1,
					customerName: s.customer_name || s.customerName || '',
					address: s.address || '',
					postcode: s.postcode || '',
					lat: s.lat || s.latitude || 0,
					lng: s.lng || s.longitude || 0,
					treeCount: s.tree_count || s.treeCount || 0,
					lastVisit: s.last_visit?.split('T')[0] || s.lastVisit || '',
					nextVisit: s.next_visit?.split('T')[0] || s.nextVisit || '',
					duration: s.duration || s.duration_minutes || 30
				}));
				showOptimized = true;
				toastSuccess(t('routes.optimize') + '! ' + t('routes.total_distance') + ': ' + calcTotalDistance(optimizedStops).toFixed(1) + ' km');
			}
		} catch (err) {
			toastError(err instanceof Error ? err.message : 'Failed to optimize route');
		} finally {
			isOptimizing = false;
		}
	}

	async function applyOptimization() {
		if (optimizedStops.length > 0) {
			try {
				const stopIds = optimizedStops.map((s) => s.id);
				await api.routes.reorderStops(routeId, stopIds);
				stops = [...optimizedStops];
				optimizedStops = [];
				showOptimized = false;
			} catch (err) {
				toastError(err instanceof Error ? err.message : 'Failed to apply optimization');
			}
		}
	}

	async function addStop() {
		if (!newStopCustomer.trim() || !newStopAddress.trim()) return;
		try {
			const res = await api.routes.addStop(routeId, {
				customer_name: newStopCustomer.trim(),
				address: newStopAddress.trim(),
				postcode: newStopPostcode.trim(),
				tree_count: newStopTrees
			});
			const newStop = res.data as any;
			stops = [...stops, {
				id: newStop.id,
				order: stops.length + 1,
				customerName: newStop.customer_name || newStopCustomer,
				address: newStop.address || newStopAddress,
				postcode: newStop.postcode || newStopPostcode,
				lat: newStop.lat || newStop.latitude || 0,
				lng: newStop.lng || newStop.longitude || 0,
				treeCount: newStop.tree_count || newStopTrees,
				lastVisit: newStop.last_visit?.split('T')[0] || '',
				nextVisit: newStop.next_visit?.split('T')[0] || '',
				duration: newStop.duration || 30
			}];
			showAddStopModal = false;
			toastSuccess(t('routes.add_stop'));
			newStopCustomer = '';
			newStopAddress = '';
			newStopPostcode = '';
			newStopTrees = 0;
		} catch (err) {
			toastError(err instanceof Error ? err.message : 'Failed to add stop');
		}
	}

	const totalDuration = $derived(stops.reduce((sum, s) => sum + s.duration, 0));
	const totalTrees = $derived(stops.reduce((sum, s) => sum + s.treeCount, 0));
	// Estimate travel time at 20 km/h average speed for urban service routes
	const estimatedTravelMins = $derived(Math.round((totalDistance / 20) * 60));
	const totalEstimatedMins = $derived(totalDuration + estimatedTravelMins);
</script>

<svelte:head>
	<title>{route.name} - Seva Provider</title>
</svelte:head>

{#if loading}
<div class="flex items-center justify-center py-20">
	<Loader2 class="h-8 w-8 animate-spin text-primary-600" />
</div>
{:else if error}
<div class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
	<a href="/provider/routes" class="inline-flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700">
		<ArrowLeft class="h-4 w-4" />
		Back to Routes
	</a>
	<div class="mt-4 rounded-lg border border-red-200 bg-red-50 p-6 text-center dark:border-red-800 dark:bg-red-900/20">
		<p class="text-red-600 dark:text-red-400">{error}</p>
	</div>
</div>
{:else}
<div class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
	<a href="/provider/routes" class="inline-flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700">
		<ArrowLeft class="h-4 w-4" />
		Back to Routes
	</a>

	<div class="mt-4 flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">{route.name}</h1>
			<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{route.description}</p>
			<div class="mt-2 flex flex-wrap items-center gap-3 text-sm text-gray-500 dark:text-gray-400">
				<span>{stops.length} stops</span>
				<span>{totalTrees} trees</span>
				<Badge variant="info" size="sm">Next: {route.nextVisit}</Badge>
			</div>
		</div>
		<div class="flex gap-2">
			<Button variant="outline" onclick={optimizeRoute} disabled={isOptimizing || stops.length < 2}>
				<Zap class="h-4 w-4" />
				{isOptimizing ? 'Optimizing...' : t('routes.optimize')}
			</Button>
			<Button variant="primary" onclick={() => (showAddStopModal = true)}>
				<Plus class="h-4 w-4" />
				{t('routes.add_stop')}
			</Button>
		</div>
	</div>

	<!-- Distance & Time Stats -->
	<div class="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
		<Card padding="sm">
			<div class="text-center">
				<div class="flex items-center justify-center gap-1 text-gray-400">
					<Navigation class="h-4 w-4" />
				</div>
				<p class="mt-1 text-lg font-bold text-gray-900 dark:text-white">{totalDistance.toFixed(1)} km</p>
				<p class="text-xs text-gray-500 dark:text-gray-400">{t('routes.total_distance')}</p>
			</div>
		</Card>
		<Card padding="sm">
			<div class="text-center">
				<div class="flex items-center justify-center gap-1 text-gray-400">
					<Clock class="h-4 w-4" />
				</div>
				<p class="mt-1 text-lg font-bold text-gray-900 dark:text-white">{Math.floor(totalEstimatedMins / 60)}h {totalEstimatedMins % 60}m</p>
				<p class="text-xs text-gray-500 dark:text-gray-400">{t('routes.estimated_time')}</p>
			</div>
		</Card>
		<Card padding="sm">
			<div class="text-center">
				<div class="flex items-center justify-center gap-1 text-gray-400">
					<MapPin class="h-4 w-4" />
				</div>
				<p class="mt-1 text-lg font-bold text-gray-900 dark:text-white">{stops.length}</p>
				<p class="text-xs text-gray-500 dark:text-gray-400">Stops</p>
			</div>
		</Card>
		<Card padding="sm">
			<div class="text-center">
				<div class="flex items-center justify-center gap-1 text-gray-400">
					<TreePine class="h-4 w-4" />
				</div>
				<p class="mt-1 text-lg font-bold text-gray-900 dark:text-white">{totalTrees}</p>
				<p class="text-xs text-gray-500 dark:text-gray-400">Trees</p>
			</div>
		</Card>
	</div>

	<!-- Optimization Result Banner -->
	{#if showOptimized && optimizedStops.length > 0 && distanceSaved > 0.01}
		<div class="mt-4 rounded-xl border border-green-200 bg-green-50 p-4 dark:border-green-900 dark:bg-green-900/20">
			<div class="flex flex-wrap items-center justify-between gap-3">
				<div>
					<p class="text-sm font-semibold text-green-800 dark:text-green-300">
						Route optimized! Save {distanceSaved.toFixed(1)} km ({((distanceSaved / totalDistance) * 100).toFixed(0)}% shorter)
					</p>
					<p class="mt-0.5 text-xs text-green-600 dark:text-green-400">
						New distance: {optimizedDistance.toFixed(1)} km (was {totalDistance.toFixed(1)} km)
					</p>
				</div>
				<div class="flex gap-2">
					<Button variant="outline" size="sm" onclick={() => { showOptimized = false; optimizedStops = []; }}>
						Dismiss
					</Button>
					<Button variant="primary" size="sm" onclick={applyOptimization}>
						Apply Changes
					</Button>
				</div>
			</div>
		</div>
	{/if}

	<!-- Route Map -->
	<Card class="mt-6" padding="none">
		<div class="h-[400px] overflow-hidden rounded-xl">
			<RouteMap
				stops={mapStops}
				optimizedStops={mapOptimizedStops}
				{editable}
				{showOptimized}
				onoptimize={optimizeRoute}
			/>
		</div>
	</Card>

	<!-- Stops List -->
	<div class="mt-6">
		<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Stops ({stops.length})</h2>
		<div class="mt-4 space-y-2">
			{#each stops as stop, i}
				<div class="rounded-xl border border-gray-200 bg-white p-4 transition-all hover:shadow-sm dark:border-gray-700 dark:bg-gray-800">
					<div class="flex items-center gap-3">
						<div class="flex flex-col items-center gap-0.5">
							<button onclick={() => moveStop(i, 'up')} disabled={i === 0} class="text-gray-400 hover:text-gray-600 disabled:opacity-30 dark:hover:text-gray-300">
								<ChevronUp class="h-4 w-4" />
							</button>
							<span class="flex h-7 w-7 items-center justify-center rounded-full bg-primary-100 text-xs font-bold text-primary-700 dark:bg-primary-900/30 dark:text-primary-400">{stop.order}</span>
							<button onclick={() => moveStop(i, 'down')} disabled={i === stops.length - 1} class="text-gray-400 hover:text-gray-600 disabled:opacity-30 dark:hover:text-gray-300">
								<ChevronDown class="h-4 w-4" />
							</button>
						</div>
						<div class="flex-1 min-w-0">
							<div class="flex items-center gap-2">
								<User class="h-4 w-4 text-gray-400 shrink-0" />
								<h3 class="font-medium text-gray-900 dark:text-white">{stop.customerName}</h3>
							</div>
							<p class="mt-0.5 text-sm text-gray-500 dark:text-gray-400">{stop.address}, {stop.postcode}</p>
							<div class="mt-1.5 flex flex-wrap items-center gap-3 text-xs text-gray-500 dark:text-gray-400">
								{#if stop.treeCount}
									<span class="flex items-center gap-1">
										<TreePine class="h-3 w-3" />
										{stop.treeCount} trees
									</span>
								{/if}
								<span>~{stop.duration}min</span>
								<span class="flex items-center gap-1">
									<Calendar class="h-3 w-3" />
									Last: {stop.lastVisit}
								</span>
								{#if i < stops.length - 1}
									<span class="flex items-center gap-1 text-primary-500">
										<Navigation class="h-3 w-3" />
										{haversineKM(stop.lat, stop.lng, stops[i + 1].lat, stops[i + 1].lng).toFixed(1)} km to next
									</span>
								{/if}
							</div>
						</div>
						<button onclick={() => removeStop(stop.id)} class="rounded p-1.5 text-gray-400 hover:bg-red-50 hover:text-red-500 dark:hover:bg-red-900/20">
							<Trash2 class="h-4 w-4" />
						</button>
					</div>
				</div>
			{/each}
		</div>
	</div>
</div>
{/if}

<!-- Add Stop Modal -->
<Modal bind:open={showAddStopModal} title={t('routes.add_stop')} size="md">
	<div class="space-y-4">
		<Input label="Customer Name" bind:value={newStopCustomer} required placeholder="Customer name" />
		<Input label="Address" bind:value={newStopAddress} required placeholder="Full address" />
		<div class="grid gap-4 sm:grid-cols-2">
			<Input label="Postcode" bind:value={newStopPostcode} placeholder="560001" />
			<Input label="Tree Count" type="number" bind:value={newStopTrees} placeholder="0" />
		</div>
	</div>
	{#snippet footer()}
		<Button variant="outline" onclick={() => (showAddStopModal = false)}>{t('common.cancel')}</Button>
		<Button variant="primary" onclick={addStop}>{t('routes.add_stop')}</Button>
	{/snippet}
</Modal>
