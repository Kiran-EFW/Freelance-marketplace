<script lang="ts">
	import { X, CheckCircle, AlertTriangle, AlertCircle, Info } from 'lucide-svelte';
	import { subscribe as subscribeToasts, removeToast, type Toast } from '$lib/stores/toast';

	let toasts = $state<Toast[]>([]);

	$effect(() => {
		const unsub = subscribeToasts((t) => {
			toasts = t;
		});
		return unsub;
	});

	const iconMap = {
		success: CheckCircle,
		error: AlertCircle,
		warning: AlertTriangle,
		info: Info
	};

	const colorMap = {
		success: 'border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-900/20',
		error: 'border-red-200 bg-red-50 dark:border-red-800 dark:bg-red-900/20',
		warning: 'border-yellow-200 bg-yellow-50 dark:border-yellow-800 dark:bg-yellow-900/20',
		info: 'border-blue-200 bg-blue-50 dark:border-blue-800 dark:bg-blue-900/20'
	};

	const iconColorMap = {
		success: 'text-green-500',
		error: 'text-red-500',
		warning: 'text-yellow-500',
		info: 'text-blue-500'
	};
</script>

{#if toasts.length > 0}
	<div class="fixed right-4 top-4 z-[100] flex flex-col gap-2" aria-live="polite">
		{#each toasts as toast (toast.id)}
			{@const Icon = iconMap[toast.type]}
			<div class="flex w-80 items-start gap-3 rounded-lg border p-4 shadow-lg {colorMap[toast.type]}">
				<Icon class="mt-0.5 h-5 w-5 shrink-0 {iconColorMap[toast.type]}" />
				<div class="flex-1 min-w-0">
					{#if toast.title}
						<p class="text-sm font-semibold text-gray-900 dark:text-white">{toast.title}</p>
					{/if}
					<p class="text-sm text-gray-700 dark:text-gray-300">{toast.message}</p>
				</div>
				<button
					onclick={() => removeToast(toast.id)}
					class="shrink-0 rounded p-0.5 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
					aria-label="Dismiss"
				>
					<X class="h-4 w-4" />
				</button>
			</div>
		{/each}
	</div>
{/if}
