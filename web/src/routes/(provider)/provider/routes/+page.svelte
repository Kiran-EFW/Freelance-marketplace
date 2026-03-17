<script lang="ts">
	import { MapPin, Calendar, Plus, ChevronRight, Route } from 'lucide-svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import { toastSuccess } from '$lib/stores/toast';

	let showNewRouteModal = $state(false);
	let newRouteName = $state('');
	let newRouteDesc = $state('');
	let newRouteRecurrence = $state('weekly');

	const mockRoutes = [
		{
			id: '1', name: 'Koramangala Route', description: 'Weekly coconut tree maintenance', stops: 8,
			nextVisit: '2026-03-19', recurrence: 'weekly', isActive: true
		},
		{
			id: '2', name: 'Indiranagar Route', description: 'Bi-weekly garden maintenance', stops: 5,
			nextVisit: '2026-03-25', recurrence: 'biweekly', isActive: true
		},
		{
			id: '3', name: 'HSR Layout Route', description: 'Monthly tree trimming', stops: 12,
			nextVisit: '2026-04-01', recurrence: 'monthly', isActive: false
		}
	];

	function createRoute() {
		if (!newRouteName.trim()) return;
		showNewRouteModal = false;
		newRouteName = '';
		newRouteDesc = '';
		toastSuccess('Route created successfully');
	}
</script>

<svelte:head>
	<title>Routes - Seva Provider</title>
</svelte:head>

<div class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Routes</h1>
			<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">Manage your service routes and stops.</p>
		</div>
		<Button variant="primary" onclick={() => (showNewRouteModal = true)}>
			<Plus class="h-4 w-4" />
			New Route
		</Button>
	</div>

	<div class="mt-6 space-y-4">
		{#each mockRoutes as route}
			<a href="/provider/routes/{route.id}" class="block">
				<Card hover>
					<div class="flex items-center gap-4">
						<div class="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400">
							<Route class="h-6 w-6" />
						</div>
						<div class="flex-1 min-w-0">
							<div class="flex items-center gap-2">
								<h3 class="font-semibold text-gray-900 dark:text-white">{route.name}</h3>
								{#if route.isActive}
									<Badge variant="success" size="sm">Active</Badge>
								{:else}
									<Badge variant="neutral" size="sm">Inactive</Badge>
								{/if}
							</div>
							{#if route.description}
								<p class="mt-0.5 text-sm text-gray-500 dark:text-gray-400">{route.description}</p>
							{/if}
							<div class="mt-2 flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
								<span class="flex items-center gap-1">
									<MapPin class="h-3 w-3" />
									{route.stops} stops
								</span>
								<span class="flex items-center gap-1">
									<Calendar class="h-3 w-3" />
									Next: {route.nextVisit}
								</span>
								<span class="capitalize">{route.recurrence}</span>
							</div>
						</div>
						<ChevronRight class="h-5 w-5 text-gray-400 shrink-0" />
					</div>
				</Card>
			</a>
		{/each}
	</div>
</div>

<!-- New Route Modal -->
<Modal bind:open={showNewRouteModal} title="Create New Route" size="md">
	<div class="space-y-4">
		<Input label="Route Name" bind:value={newRouteName} required placeholder="e.g., Koramangala Route" />
		<div>
			<label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">Description</label>
			<textarea
				bind:value={newRouteDesc}
				rows="2"
				placeholder="What type of service is this route for?"
				class="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
			></textarea>
		</div>
		<div>
			<label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">Recurrence</label>
			<select
				bind:value={newRouteRecurrence}
				class="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm focus:border-primary-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-white"
			>
				<option value="daily">Daily</option>
				<option value="weekly">Weekly</option>
				<option value="biweekly">Bi-weekly</option>
				<option value="monthly">Monthly</option>
			</select>
		</div>
	</div>
	{#snippet footer()}
		<Button variant="outline" onclick={() => (showNewRouteModal = false)}>Cancel</Button>
		<Button variant="primary" onclick={createRoute} disabled={!newRouteName.trim()}>Create Route</Button>
	{/snippet}
</Modal>
