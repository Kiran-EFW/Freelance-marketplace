<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		variant?: 'success' | 'warning' | 'danger' | 'info' | 'neutral';
		size?: 'sm' | 'md';
		class?: string;
		children: Snippet;
	}

	let {
		variant = 'neutral',
		size = 'sm',
		class: className = '',
		children
	}: Props = $props();

	const variantClasses = $derived({
		success: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
		warning: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
		danger: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
		info: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
		neutral: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
	}[variant]);

	const sizeClasses = $derived({
		sm: 'px-2 py-0.5 text-xs',
		md: 'px-2.5 py-1 text-sm'
	}[size]);
</script>

<span class="inline-flex items-center rounded-full font-medium {variantClasses} {sizeClasses} {className}">
	{@render children()}
</span>
