<script lang="ts">
	interface Props {
		type?: string;
		label?: string;
		placeholder?: string;
		error?: string;
		hint?: string;
		disabled?: boolean;
		icon?: any;
		value?: string | number;
		id?: string;
		name?: string;
		required?: boolean;
		maxlength?: number;
		class?: string;
		oninput?: (e: Event) => void;
		onchange?: (e: Event) => void;
	}

	let {
		type = 'text',
		label = '',
		placeholder = '',
		error = '',
		hint = '',
		disabled = false,
		icon: Icon,
		value = $bindable(''),
		id = '',
		name = '',
		required = false,
		maxlength,
		class: className = '',
		oninput,
		onchange
	}: Props = $props();

	let focused = $state(false);
	const inputId = $derived(id || `input-${label?.toLowerCase().replace(/\s+/g, '-') || Math.random().toString(36).slice(2)}`);
</script>

<div class="w-full {className}">
	{#if label}
		<label for={inputId} class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
			{label}
			{#if required}
				<span class="text-red-500">*</span>
			{/if}
		</label>
	{/if}

	<div class="relative">
		{#if Icon}
			<div class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
				<Icon class="h-5 w-5" />
			</div>
		{/if}

		<input
			id={inputId}
			{type}
			{name}
			{placeholder}
			{disabled}
			{required}
			{maxlength}
			bind:value
			{oninput}
			{onchange}
			onfocus={() => (focused = true)}
			onblur={() => (focused = false)}
			class="w-full rounded-lg border px-4 py-2.5 text-sm transition-colors focus:outline-none focus:ring-2
				{Icon ? 'pl-10' : ''}
				{error
					? 'border-red-300 text-red-900 placeholder-red-300 focus:border-red-500 focus:ring-red-500/20 dark:border-red-600 dark:text-red-400'
					: 'border-gray-300 text-gray-900 placeholder-gray-400 focus:border-primary-500 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400'}
				{disabled ? 'cursor-not-allowed bg-gray-50 dark:bg-gray-800' : 'bg-white dark:bg-gray-700'}"
		/>
	</div>

	{#if error}
		<p class="mt-1.5 text-xs text-red-600 dark:text-red-400">{error}</p>
	{:else if hint}
		<p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">{hint}</p>
	{/if}
</div>
