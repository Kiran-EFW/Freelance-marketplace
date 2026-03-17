<script lang="ts">
	import { ChevronLeft, ChevronRight } from 'lucide-svelte';

	interface Props {
		currentPage: number;
		totalPages: number;
		onPageChange: (page: number) => void;
		class?: string;
	}

	let {
		currentPage,
		totalPages,
		onPageChange,
		class: className = ''
	}: Props = $props();

	const pages = $derived(() => {
		const items: (number | string)[] = [];
		const delta = 2;
		const left = Math.max(2, currentPage - delta);
		const right = Math.min(totalPages - 1, currentPage + delta);

		items.push(1);
		if (left > 2) items.push('...');
		for (let i = left; i <= right; i++) {
			items.push(i);
		}
		if (right < totalPages - 1) items.push('...');
		if (totalPages > 1) items.push(totalPages);

		return items;
	});
</script>

{#if totalPages > 1}
	<nav class="flex items-center justify-center gap-1 {className}" aria-label="Pagination">
		<button
			onclick={() => onPageChange(currentPage - 1)}
			disabled={currentPage <= 1}
			class="inline-flex items-center rounded-lg px-2 py-2 text-sm text-gray-500 hover:bg-gray-100 disabled:cursor-not-allowed disabled:opacity-50 dark:text-gray-400 dark:hover:bg-gray-800"
			aria-label="Previous page"
		>
			<ChevronLeft class="h-4 w-4" />
		</button>

		{#each pages() as page}
			{#if page === '...'}
				<span class="px-2 py-2 text-sm text-gray-400 dark:text-gray-500">...</span>
			{:else}
				<button
					onclick={() => onPageChange(page as number)}
					class="inline-flex min-w-[2.25rem] items-center justify-center rounded-lg px-3 py-2 text-sm font-medium transition-colors
						{page === currentPage
							? 'bg-primary-600 text-white'
							: 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800'}"
					aria-current={page === currentPage ? 'page' : undefined}
				>
					{page}
				</button>
			{/if}
		{/each}

		<button
			onclick={() => onPageChange(currentPage + 1)}
			disabled={currentPage >= totalPages}
			class="inline-flex items-center rounded-lg px-2 py-2 text-sm text-gray-500 hover:bg-gray-100 disabled:cursor-not-allowed disabled:opacity-50 dark:text-gray-400 dark:hover:bg-gray-800"
			aria-label="Next page"
		>
			<ChevronRight class="h-4 w-4" />
		</button>
	</nav>
{/if}
