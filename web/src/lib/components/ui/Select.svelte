<script lang="ts">
	import { ChevronDown, Search, Check } from 'lucide-svelte';

	interface SelectOption {
		value: string;
		label: string;
		group?: string;
	}

	interface Props {
		options: SelectOption[];
		value?: string;
		label?: string;
		placeholder?: string;
		error?: string;
		searchable?: boolean;
		disabled?: boolean;
		class?: string;
		onchange?: (value: string) => void;
	}

	let {
		options,
		value = $bindable(''),
		label = '',
		placeholder = 'Select an option',
		error = '',
		searchable = false,
		disabled = false,
		class: className = '',
		onchange
	}: Props = $props();

	let open = $state(false);
	let searchQuery = $state('');
	let dropdownEl: HTMLDivElement | undefined = $state();
	const selectId = `select-${Math.random().toString(36).slice(2, 9)}`;

	const filteredOptions = $derived(
		searchable && searchQuery
			? options.filter((o) => o.label.toLowerCase().includes(searchQuery.toLowerCase()))
			: options
	);

	const selectedLabel = $derived(options.find((o) => o.value === value)?.label || '');

	function select(optionValue: string) {
		value = optionValue;
		open = false;
		searchQuery = '';
		onchange?.(optionValue);
	}

	function handleClickOutside(e: MouseEvent) {
		if (dropdownEl && !dropdownEl.contains(e.target as Node)) {
			open = false;
			searchQuery = '';
		}
	}

	$effect(() => {
		if (open) {
			document.addEventListener('click', handleClickOutside);
			return () => document.removeEventListener('click', handleClickOutside);
		}
	});
</script>

<div class="relative {className}" bind:this={dropdownEl}>
	{#if label}
		<span id="{selectId}-label" class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">{label}</span>
	{/if}

	<button
		type="button"
		{disabled}
		aria-labelledby={label ? `${selectId}-label` : undefined}
		onclick={() => { if (!disabled) open = !open; }}
		class="flex w-full items-center justify-between rounded-lg border px-4 py-2.5 text-left text-sm transition-colors
			{error
				? 'border-red-300 focus:border-red-500 focus:ring-red-500/20'
				: 'border-gray-300 focus:border-primary-500 focus:ring-primary-500/20 dark:border-gray-600'}
			{disabled ? 'cursor-not-allowed bg-gray-50 dark:bg-gray-800' : 'bg-white dark:bg-gray-700'}
			focus:outline-none focus:ring-2"
	>
		<span class="{selectedLabel ? 'text-gray-900 dark:text-white' : 'text-gray-400 dark:text-gray-500'}">
			{selectedLabel || placeholder}
		</span>
		<ChevronDown class="h-4 w-4 text-gray-400 transition-transform {open ? 'rotate-180' : ''}" />
	</button>

	{#if open}
		<div class="absolute z-20 mt-1 w-full rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-600 dark:bg-gray-700">
			{#if searchable}
				<div class="border-b border-gray-200 p-2 dark:border-gray-600">
					<div class="relative">
						<Search class="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
						<input
							type="text"
							bind:value={searchQuery}
							placeholder="Search..."
							class="w-full rounded border-0 bg-gray-50 py-1.5 pl-8 pr-3 text-sm focus:bg-white focus:outline-none focus:ring-1 focus:ring-primary-500 dark:bg-gray-600 dark:text-white dark:focus:bg-gray-500"
						/>
					</div>
				</div>
			{/if}

			<div class="max-h-60 overflow-y-auto py-1">
				{#if filteredOptions.length === 0}
					<p class="px-4 py-2 text-sm text-gray-500 dark:text-gray-400">No options found</p>
				{:else}
					{#each filteredOptions as option}
						<button
							type="button"
							onclick={() => select(option.value)}
							class="flex w-full items-center gap-2 px-4 py-2 text-left text-sm transition-colors
								{option.value === value
									? 'bg-primary-50 text-primary-700 dark:bg-primary-900/20 dark:text-primary-400'
									: 'text-gray-700 hover:bg-gray-50 dark:text-gray-300 dark:hover:bg-gray-600'}"
						>
							<span class="flex-1">{option.label}</span>
							{#if option.value === value}
								<Check class="h-4 w-4 text-primary-600" />
							{/if}
						</button>
					{/each}
				{/if}
			</div>
		</div>
	{/if}

	{#if error}
		<p class="mt-1.5 text-xs text-red-600 dark:text-red-400">{error}</p>
	{/if}
</div>
