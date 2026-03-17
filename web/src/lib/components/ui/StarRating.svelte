<script lang="ts">
	import { Star } from 'lucide-svelte';

	interface Props {
		rating?: number;
		editable?: boolean;
		onRate?: (rating: number) => void;
		size?: 'sm' | 'md' | 'lg';
		showValue?: boolean;
		class?: string;
	}

	let {
		rating = $bindable(0),
		editable = false,
		onRate,
		size = 'md',
		showValue = false,
		class: className = ''
	}: Props = $props();

	let hoverRating = $state(0);

	const sizeClasses = $derived({
		sm: 'h-3.5 w-3.5',
		md: 'h-5 w-5',
		lg: 'h-6 w-6'
	}[size]);

	function handleClick(star: number) {
		if (!editable) return;
		rating = star;
		onRate?.(star);
	}

	function handleMouseEnter(star: number) {
		if (!editable) return;
		hoverRating = star;
	}

	function handleMouseLeave() {
		if (!editable) return;
		hoverRating = 0;
	}

	const displayRating = $derived(hoverRating || rating);
</script>

<div class="inline-flex items-center gap-0.5 {className}" role={editable ? 'radiogroup' : 'img'} aria-label="Rating: {rating} out of 5">
	{#each [1, 2, 3, 4, 5] as star}
		{#if editable}
			<button
				type="button"
				onclick={() => handleClick(star)}
				onmouseenter={() => handleMouseEnter(star)}
				onmouseleave={handleMouseLeave}
				class="focus:outline-none {editable ? 'cursor-pointer' : ''}"
				aria-label="{star} star{star > 1 ? 's' : ''}"
			>
				<Star
					class="{sizeClasses} {star <= displayRating ? 'fill-yellow-400 text-yellow-400' : 'fill-none text-gray-300 dark:text-gray-600'} transition-colors"
				/>
			</button>
		{:else}
			<Star
				class="{sizeClasses} {star <= Math.round(displayRating) ? 'fill-yellow-400 text-yellow-400' : star - 0.5 <= displayRating ? 'fill-yellow-400/50 text-yellow-400' : 'fill-none text-gray-300 dark:text-gray-600'}"
			/>
		{/if}
	{/each}
	{#if showValue}
		<span class="ml-1 text-sm font-medium text-gray-700 dark:text-gray-300">{rating.toFixed(1)}</span>
	{/if}
</div>
