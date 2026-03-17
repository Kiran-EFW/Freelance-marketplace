<script lang="ts">
	import { X } from 'lucide-svelte';
	import type { Snippet } from 'svelte';

	interface Props {
		open: boolean;
		title?: string;
		size?: 'sm' | 'md' | 'lg' | 'xl';
		onclose?: () => void;
		children: Snippet;
		footer?: Snippet;
	}

	let {
		open = $bindable(false),
		title = '',
		size = 'md',
		onclose,
		children,
		footer
	}: Props = $props();

	const sizeClasses = $derived({
		sm: 'max-w-sm',
		md: 'max-w-lg',
		lg: 'max-w-2xl',
		xl: 'max-w-4xl'
	}[size]);

	function close() {
		open = false;
		onclose?.();
	}

	function handleBackdrop(e: MouseEvent) {
		if (e.target === e.currentTarget) {
			close();
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			close();
		}
	}
</script>

<svelte:window onkeydown={open ? handleKeydown : undefined} />

{#if open}
	<!-- Backdrop -->
	<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto p-4"
		onclick={handleBackdrop}
		onkeydown={(e) => { if (e.key === 'Escape') close(); }}
		role="dialog"
		tabindex="-1"
		aria-modal="true"
		aria-labelledby={title ? 'modal-title' : undefined}
	>
		<!-- Overlay -->
		<div class="fixed inset-0 bg-black/50 transition-opacity" aria-hidden="true"></div>

		<!-- Panel -->
		<div class="relative w-full {sizeClasses} transform rounded-xl bg-white shadow-xl transition-all dark:bg-gray-800">
			<!-- Header -->
			{#if title}
				<div class="flex items-center justify-between border-b border-gray-200 px-6 py-4 dark:border-gray-700">
					<h2 id="modal-title" class="text-lg font-semibold text-gray-900 dark:text-white">{title}</h2>
					<button
						onclick={close}
						class="rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-700 dark:hover:text-gray-300"
					>
						<X class="h-5 w-5" />
					</button>
				</div>
			{/if}

			<!-- Content -->
			<div class="px-6 py-4">
				{@render children()}
			</div>

			<!-- Footer -->
			{#if footer}
				<div class="flex items-center justify-end gap-3 border-t border-gray-200 px-6 py-4 dark:border-gray-700">
					{@render footer()}
				</div>
			{/if}
		</div>
	</div>
{/if}
