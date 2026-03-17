<script lang="ts">
	import { Globe, Check, ChevronDown } from 'lucide-svelte';
	import { locales, getLocalesByRegion, type Locale } from '$lib/i18n/locales';
	import { currentLocale, setLocale } from '$lib/i18n/index.svelte';

	interface Props {
		class?: string;
		variant?: 'dropdown' | 'compact';
	}

	let { class: className = '', variant = 'dropdown' }: Props = $props();

	let open = $state(false);
	let containerEl: HTMLDivElement | undefined = $state();

	const localeCode = $derived(currentLocale.toUpperCase());

	const regionLabels: Record<string, string> = {
		global: 'Global',
		indian: 'Indian Languages',
		european: 'European Languages'
	};

	const regionOrder = ['global', 'indian', 'european'];

	const groupedLocales = $derived(getLocalesByRegion());

	function selectLocale(code: string) {
		setLocale(code);
		open = false;
	}

	function handleClickOutside(e: MouseEvent) {
		if (containerEl && !containerEl.contains(e.target as Node)) {
			open = false;
		}
	}

	$effect(() => {
		if (open) {
			document.addEventListener('click', handleClickOutside);
			return () => document.removeEventListener('click', handleClickOutside);
		}
	});
</script>

<div class="relative {className}" bind:this={containerEl}>
	<button
		type="button"
		onclick={() => (open = !open)}
		class="inline-flex items-center gap-1.5 rounded-lg px-2.5 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:text-gray-300 dark:hover:bg-gray-800"
		aria-expanded={open}
		aria-haspopup="listbox"
		aria-label="Select language"
	>
		<Globe class="h-4 w-4" />
		<span class="text-xs font-semibold">{localeCode}</span>
		{#if variant === 'dropdown'}
			<ChevronDown class="h-3.5 w-3.5 text-gray-400 transition-transform {open ? 'rotate-180' : ''}" />
		{/if}
	</button>

	{#if open}
		<div
			class="absolute right-0 z-50 mt-2 w-72 origin-top-right rounded-xl border border-gray-200 bg-white shadow-xl transition-all dark:border-gray-700 dark:bg-gray-800"
			role="listbox"
			aria-label="Available languages"
		>
			<div class="max-h-80 overflow-y-auto overscroll-contain py-1">
				{#each regionOrder as region}
					{@const regionLocales = groupedLocales[region]}
					{#if regionLocales && regionLocales.length > 0}
						<div class="px-3 pb-1 pt-3 first:pt-2">
							<p class="text-[11px] font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">
								{regionLabels[region]}
							</p>
						</div>

						{#each regionLocales as locale (locale.code)}
							<button
								type="button"
								onclick={() => selectLocale(locale.code)}
								class="flex w-full items-center gap-3 px-3 py-2 text-left text-sm transition-colors
									{locale.code === currentLocale
										? 'bg-primary-50 text-primary-700 dark:bg-primary-900/20 dark:text-primary-400'
										: 'text-gray-700 hover:bg-gray-50 dark:text-gray-300 dark:hover:bg-gray-700/50'}"
								role="option"
								aria-selected={locale.code === currentLocale}
							>
								<span class="flex flex-1 items-baseline gap-2">
									<span class="font-medium">{locale.nativeName}</span>
									<span class="text-xs text-gray-400 dark:text-gray-500">{locale.name}</span>
								</span>
								{#if locale.code === currentLocale}
									<Check class="h-4 w-4 shrink-0 text-primary-600 dark:text-primary-400" />
								{/if}
							</button>
						{/each}

						{#if region !== regionOrder[regionOrder.length - 1]}
							<div class="mx-3 my-1 border-t border-gray-100 dark:border-gray-700/50"></div>
						{/if}
					{/if}
				{/each}
			</div>
		</div>
	{/if}
</div>
