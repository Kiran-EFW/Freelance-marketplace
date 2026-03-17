<script lang="ts">
	interface Tab {
		id: string;
		label: string;
		count?: number;
	}

	interface Props {
		tabs: Tab[];
		activeTab?: string;
		onTabChange?: (tabId: string) => void;
		class?: string;
	}

	let {
		tabs,
		activeTab = $bindable(''),
		onTabChange,
		class: className = ''
	}: Props = $props();

	// Default to first tab if none selected
	$effect(() => {
		if (!activeTab && tabs.length > 0) {
			activeTab = tabs[0].id;
		}
	});

	function selectTab(tabId: string) {
		activeTab = tabId;
		onTabChange?.(tabId);
	}
</script>

<div class="border-b border-gray-200 dark:border-gray-700 {className}">
	<div class="-mb-px flex gap-4 overflow-x-auto" role="tablist">
		{#each tabs as tab}
			<button
				type="button"
				onclick={() => selectTab(tab.id)}
				role="tab"
				aria-selected={tab.id === activeTab}
				class="inline-flex shrink-0 items-center gap-2 border-b-2 px-1 py-3 text-sm font-medium transition-colors
					{tab.id === activeTab
						? 'border-primary-600 text-primary-600 dark:border-primary-500 dark:text-primary-400'
						: 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-gray-400 dark:hover:border-gray-600 dark:hover:text-gray-300'}"
			>
				{tab.label}
				{#if tab.count !== undefined}
					<span class="rounded-full px-2 py-0.5 text-xs
						{tab.id === activeTab
							? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400'
							: 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'}">
						{tab.count}
					</span>
				{/if}
			</button>
		{/each}
	</div>
</div>
