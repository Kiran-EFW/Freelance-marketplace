<script lang="ts">
	interface Props {
		src?: string;
		name?: string;
		size?: 'sm' | 'md' | 'lg' | 'xl';
		class?: string;
	}

	let {
		src = '',
		name = '',
		size = 'md',
		class: className = ''
	}: Props = $props();

	const sizeClasses = $derived({
		sm: 'h-8 w-8 text-xs',
		md: 'h-10 w-10 text-sm',
		lg: 'h-14 w-14 text-lg',
		xl: 'h-20 w-20 text-2xl'
	}[size]);

	const initials = $derived(
		name
			.split(' ')
			.map((n) => n[0])
			.join('')
			.toUpperCase()
			.slice(0, 2)
	);

	// Generate a consistent color based on name
	const colors = [
		'bg-primary-100 text-primary-700 dark:bg-primary-900/40 dark:text-primary-400',
		'bg-secondary-100 text-secondary-700 dark:bg-secondary-900/40 dark:text-secondary-400',
		'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-400',
		'bg-purple-100 text-purple-700 dark:bg-purple-900/40 dark:text-purple-400',
		'bg-pink-100 text-pink-700 dark:bg-pink-900/40 dark:text-pink-400',
		'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/40 dark:text-indigo-400'
	];

	const colorIndex = $derived(
		name.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0) % colors.length
	);
	const bgColor = $derived(colors[colorIndex]);

	let imgError = $state(false);
</script>

{#if src && !imgError}
	<img
		{src}
		alt={name || 'Avatar'}
		class="rounded-full object-cover {sizeClasses} {className}"
		onerror={() => (imgError = true)}
	/>
{:else}
	<div class="flex shrink-0 items-center justify-center rounded-full font-semibold {sizeClasses} {bgColor} {className}">
		{initials || '?'}
	</div>
{/if}
