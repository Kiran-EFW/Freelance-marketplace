<script lang="ts">
	import { page } from '$app/stores';
	import { ArrowLeft, MapPin, User, TreePine, Calendar, ChevronUp, ChevronDown, Plus, Zap, Trash2, GripVertical } from 'lucide-svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import { toastSuccess } from '$lib/stores/toast';

	let routeId = $derived($page.params.id);
	let showAddStopModal = $state(false);
	let newStopCustomer = $state('');
	let newStopAddress = $state('');
	let newStopPostcode = $state('');
	let newStopTrees = $state(0);

	const route = {
		id: '1', name: 'Koramangala Route', description: 'Weekly coconut tree maintenance',
		recurrence: 'weekly', nextVisit: '2026-03-19', isActive: true
	};

	let stops = $state([
		{ id: 's1', order: 1, customerName: 'Amit Verma', address: '123 MG Road', postcode: '560001', treeCount: 3, lastVisit: '2026-03-12', nextVisit: '2026-03-19', duration: 30 },
		{ id: 's2', order: 2, customerName: 'Priya Menon', address: '45 Church Street', postcode: '560001', treeCount: 5, lastVisit: '2026-03-12', nextVisit: '2026-03-19', duration: 45 },
		{ id: 's3', order: 3, customerName: 'Arjun Das', address: '78 Brigade Road', postcode: '560002', treeCount: 2, lastVisit: '2026-03-12', nextVisit: '2026-03-19', duration: 20 },
		{ id: 's4', order: 4, customerName: 'Meera Reddy', address: '12 Hosur Road', postcode: '560002', treeCount: 4, lastVisit: '2026-03-12', nextVisit: '2026-03-19', duration: 35 },
		{ id: 's5', order: 5, customerName: 'Kiran Rao', address: '56 Sarjapur Road', postcode: '560003', treeCount: 6, lastVisit: '2026-03-12', nextVisit: '2026-03-19', duration: 50 }
	]);

	function moveStop(index: number, direction: 'up' | 'down') {
		const newIndex = direction === 'up' ? index - 1 : index + 1;
		if (newIndex < 0 || newIndex >= stops.length) return;
		const newStops = [...stops];
		[newStops[index], newStops[newIndex]] = [newStops[newIndex], newStops[index]];
		newStops.forEach((s, i) => (s.order = i + 1));
		stops = newStops;
	}

	function removeStop(id: string) {
		stops = stops.filter((s) => s.id !== id);
		stops.forEach((s, i) => (s.order = i + 1));
		toastSuccess('Stop removed');
	}

	function optimizeRoute() {
		toastSuccess('Route optimized! Stops reordered for shortest travel time.');
	}

	function addStop() {
		if (!newStopCustomer.trim() || !newStopAddress.trim()) return;
		showAddStopModal = false;
		toastSuccess('Stop added to route');
		newStopCustomer = '';
		newStopAddress = '';
		newStopPostcode = '';
		newStopTrees = 0;
	}

	const totalDuration = $derived(stops.reduce((sum, s) => sum + s.duration, 0));
	const totalTrees = $derived(stops.reduce((sum, s) => sum + s.treeCount, 0));
</script>

<svelte:head>
	<title>{route.name} - Seva Provider</title>
</svelte:head>

<div class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
	<a href="/provider/routes" class="inline-flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700">
		<ArrowLeft class="h-4 w-4" />
		Back to Routes
	</a>

	<div class="mt-4 flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">{route.name}</h1>
			<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{route.description}</p>
			<div class="mt-2 flex items-center gap-3 text-sm text-gray-500 dark:text-gray-400">
				<span>{stops.length} stops</span>
				<span>{totalTrees} trees</span>
				<span>~{Math.round(totalDuration / 60)}h {totalDuration % 60}m total</span>
				<Badge variant="info" size="sm">Next: {route.nextVisit}</Badge>
			</div>
		</div>
		<div class="flex gap-2">
			<Button variant="outline" onclick={optimizeRoute}>
				<Zap class="h-4 w-4" />
				Optimize Route
			</Button>
			<Button variant="primary" onclick={() => (showAddStopModal = true)}>
				<Plus class="h-4 w-4" />
				Add Stop
			</Button>
		</div>
	</div>

	<!-- Map Placeholder -->
	<Card class="mt-6" padding="none">
		<div class="flex h-48 items-center justify-center bg-gray-100 rounded-t-xl dark:bg-gray-800">
			<div class="text-center">
				<MapPin class="mx-auto h-8 w-8 text-gray-400" />
				<p class="mt-2 text-sm text-gray-500 dark:text-gray-400">Route map will be displayed here</p>
			</div>
		</div>
	</Card>

	<!-- Stops List -->
	<div class="mt-6">
		<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Stops ({stops.length})</h2>
		<div class="mt-4 space-y-2">
			{#each stops as stop, i}
				<div class="rounded-xl border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-800">
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

<!-- Add Stop Modal -->
<Modal bind:open={showAddStopModal} title="Add Stop" size="md">
	<div class="space-y-4">
		<Input label="Customer Name" bind:value={newStopCustomer} required placeholder="Customer name" />
		<Input label="Address" bind:value={newStopAddress} required placeholder="Full address" />
		<div class="grid gap-4 sm:grid-cols-2">
			<Input label="Postcode" bind:value={newStopPostcode} placeholder="560001" />
			<Input label="Tree Count" type="number" bind:value={newStopTrees} placeholder="0" />
		</div>
	</div>
	{#snippet footer()}
		<Button variant="outline" onclick={() => (showAddStopModal = false)}>Cancel</Button>
		<Button variant="primary" onclick={addStop}>Add Stop</Button>
	{/snippet}
</Modal>
