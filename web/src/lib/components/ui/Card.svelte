<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		padding?: 'none' | 'sm' | 'md' | 'lg';
		hover?: boolean;
		href?: string;
		class?: string;
		children: Snippet;
	}

	let {
		padding = 'md',
		hover = false,
		href,
		class: className = '',
		children
	}: Props = $props();

	const paddingClasses = $derived({
		none: '',
		sm: 'p-4',
		md: 'p-6',
		lg: 'p-8'
	}[padding]);

	const baseClasses = $derived(
		`rounded-xl border border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-800 ${paddingClasses} ${hover ? 'transition-shadow hover:shadow-md cursor-pointer' : ''} ${className}`
	);
</script>

{#if href}
	<a {href} class="block {baseClasses}">
		{@render children()}
	</a>
{:else}
	<div class={baseClasses}>
		{@render children()}
	</div>
{/if}
