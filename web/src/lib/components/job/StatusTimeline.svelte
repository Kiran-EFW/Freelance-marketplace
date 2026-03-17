<script lang="ts">
	import { Check, Circle } from 'lucide-svelte';
	import type { JobStatus } from '$lib/types';

	interface Props {
		currentStatus: JobStatus;
		class?: string;
	}

	let { currentStatus, class: className = '' }: Props = $props();

	const steps = [
		{ key: 'open', label: 'Posted' },
		{ key: 'quoted', label: 'Quoted' },
		{ key: 'accepted', label: 'Accepted' },
		{ key: 'in_progress', label: 'In Progress' },
		{ key: 'completed', label: 'Completed' },
	] as const;

	const statusOrder: Record<string, number> = {
		draft: -1,
		open: 0,
		quoted: 1,
		accepted: 2,
		in_progress: 3,
		completed: 4,
		cancelled: -2,
		disputed: -2
	};

	const currentIndex = $derived(statusOrder[currentStatus] ?? -1);

	function getStepState(stepIndex: number): 'completed' | 'current' | 'upcoming' {
		if (currentStatus === 'cancelled' || currentStatus === 'disputed') {
			return stepIndex === 0 ? 'completed' : 'upcoming';
		}
		if (stepIndex < currentIndex) return 'completed';
		if (stepIndex === currentIndex) return 'current';
		return 'upcoming';
	}
</script>

<div class="w-full {className}">
	{#if currentStatus === 'cancelled'}
		<div class="mb-3 rounded-lg bg-red-50 px-4 py-2 text-sm text-red-700 dark:bg-red-900/20 dark:text-red-400">
			This job has been cancelled.
		</div>
	{:else if currentStatus === 'disputed'}
		<div class="mb-3 rounded-lg bg-red-50 px-4 py-2 text-sm text-red-700 dark:bg-red-900/20 dark:text-red-400">
			This job is under dispute.
		</div>
	{/if}

	<!-- Mobile: vertical timeline -->
	<div class="sm:hidden">
		{#each steps as step, i}
			{@const state = getStepState(i)}
			<div class="flex gap-3 {i < steps.length - 1 ? 'pb-6' : ''}">
				<div class="flex flex-col items-center">
					{#if state === 'completed'}
						<div class="flex h-7 w-7 items-center justify-center rounded-full bg-primary-600 text-white">
							<Check class="h-4 w-4" />
						</div>
					{:else if state === 'current'}
						<div class="flex h-7 w-7 items-center justify-center rounded-full border-2 border-primary-600 bg-primary-50 dark:bg-primary-900/30">
							<div class="h-2.5 w-2.5 rounded-full bg-primary-600"></div>
						</div>
					{:else}
						<div class="flex h-7 w-7 items-center justify-center rounded-full border-2 border-gray-300 dark:border-gray-600">
							<Circle class="h-3 w-3 text-gray-300 dark:text-gray-600" />
						</div>
					{/if}
					{#if i < steps.length - 1}
						<div class="mt-1 w-0.5 flex-1 {state === 'completed' ? 'bg-primary-600' : 'bg-gray-200 dark:bg-gray-700'}"></div>
					{/if}
				</div>
				<div class="pt-0.5">
					<p class="text-sm font-medium {state === 'upcoming' ? 'text-gray-400 dark:text-gray-500' : 'text-gray-900 dark:text-white'}">
						{step.label}
					</p>
				</div>
			</div>
		{/each}
	</div>

	<!-- Desktop: horizontal timeline -->
	<div class="hidden sm:block">
		<div class="flex items-center justify-between">
			{#each steps as step, i}
				{@const state = getStepState(i)}
				<div class="flex flex-1 items-center {i === steps.length - 1 ? '' : ''}">
					<div class="flex flex-col items-center">
						{#if state === 'completed'}
							<div class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-600 text-white">
								<Check class="h-4 w-4" />
							</div>
						{:else if state === 'current'}
							<div class="flex h-8 w-8 items-center justify-center rounded-full border-2 border-primary-600 bg-primary-50 dark:bg-primary-900/30">
								<div class="h-2.5 w-2.5 rounded-full bg-primary-600"></div>
							</div>
						{:else}
							<div class="flex h-8 w-8 items-center justify-center rounded-full border-2 border-gray-300 dark:border-gray-600"></div>
						{/if}
						<span class="mt-2 text-xs font-medium {state === 'upcoming' ? 'text-gray-400 dark:text-gray-500' : 'text-gray-700 dark:text-gray-300'}">
							{step.label}
						</span>
					</div>
					{#if i < steps.length - 1}
						<div class="mx-2 h-0.5 flex-1 {state === 'completed' ? 'bg-primary-600' : 'bg-gray-200 dark:bg-gray-700'}"></div>
					{/if}
				</div>
			{/each}
		</div>
	</div>
</div>
