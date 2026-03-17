<script lang="ts">
	import { goto } from '$app/navigation';
	import {
		Wrench, Sparkles, Scissors, Briefcase, Car, GraduationCap, HeartHandshake,
		PartyPopper, Truck, Monitor, TreePine, HardHat, Droplets, Zap, Hammer,
		Paintbrush, Construction, Home, Flame, Square, Lock, Bug, CookingPot, Bath,
		Sofa, Layers, Container, Building2, Heart, Palette, Dumbbell, Activity,
		Scale, Calculator, Ruler, FileText, TrendingUp, Bike, Circle, BookOpen,
		Music, Music2, Languages, Medal, Baby, Stethoscope, PawPrint,
		UtensilsCrossed, Flower2, Camera, Video, Speaker, Package, Armchair, Send,
		Trash2, Laptop, Smartphone, Eye, Wifi, Sun, Map as MapIcon,
		Search, MapPin, Loader2, ChevronLeft, ChevronRight, Check
	} from 'lucide-svelte';
	import VoiceInput from '$lib/components/ui/VoiceInput.svelte';
	import {
		categories,
		topLevelCategories,
		getSubcategories,
		type ServiceCategory
	} from '$lib/data/categories';
	import {
		detectLocation,
		type DetectedLocation
	} from '$lib/utils/geolocation';
	import { t, i18n } from '$lib/i18n/index.svelte';
	import { localeMap } from '$lib/i18n/locales';

	// -------------------------------------------------------------------------
	// Icon map: category icon name -> Lucide component
	// -------------------------------------------------------------------------

	const iconMap: Record<string, any> = {
		Wrench, Sparkles, Scissors, Briefcase, Car, GraduationCap, HeartHandshake,
		PartyPopper, Truck, Monitor, TreePine, HardHat, Droplets, Zap, Hammer,
		Paintbrush, Construction, Home, Flame, Square, Lock, Bug, CookingPot, Bath,
		Sofa, Layers, Container, Building2, Heart, Palette, Dumbbell, Activity,
		Scale, Calculator, Ruler, FileText, TrendingUp, Bike, Circle, BookOpen,
		Music, Music2, Languages, Medal, Baby, Stethoscope, PawPrint,
		UtensilsCrossed, Flower2, Camera, Video, Speaker, Package, Armchair, Send,
		Trash2, Laptop, Smartphone, Eye, Wifi, Sun, Map: MapIcon
	};

	// -------------------------------------------------------------------------
	// State
	// -------------------------------------------------------------------------

	let step = $state(1);
	let selectedCategory = $state<string | null>(null);
	let selectedSubcategory = $state<string | null>(null);
	let searchQuery = $state('');
	let location = $state<DetectedLocation | null>(null);
	let locationError = $state('');
	let locationLoading = $state(false);
	let manualPostcode = $state('');
	let showManualInput = $state(false);
	let showAllSubcategories = $state(false);

	// -------------------------------------------------------------------------
	// Derived data
	// -------------------------------------------------------------------------

	const speechCode = $derived(localeMap[i18n.locale]?.speechCode || 'en-US');

	/** Subcategories of the currently selected top-level category. */
	const subcategories = $derived<ServiceCategory[]>(
		selectedCategory ? getSubcategories(selectedCategory) : []
	);

	/** Filter top-level categories (and their subs) by search query. */
	const filteredTopLevel = $derived.by<ServiceCategory[]>(() => {
		const q = searchQuery.trim().toLowerCase();
		if (!q) return topLevelCategories;

		return topLevelCategories.filter((cat) => {
			// Match top-level name
			const topName = t(cat.translationKey).toLowerCase();
			if (topName.includes(q)) return true;

			// Match any subcategory name
			const subs = getSubcategories(cat.id);
			return subs.some((sub) =>
				t(sub.translationKey).toLowerCase().includes(q)
			);
		});
	});

	/** When searching, pre-filter subcategories too. */
	const filteredSubcategories = $derived.by<ServiceCategory[]>(() => {
		const q = searchQuery.trim().toLowerCase();
		if (!q) return subcategories;

		return subcategories.filter((sub) =>
			t(sub.translationKey).toLowerCase().includes(q)
		);
	});

	/**
	 * The category object that will be used for the final URL slug.
	 * Subcategory takes precedence; falls back to top-level.
	 */
	const resolvedCategory = $derived.by<ServiceCategory | undefined>(() => {
		if (selectedSubcategory) {
			return categories.find((c) => c.id === selectedSubcategory);
		}
		if (selectedCategory) {
			return categories.find((c) => c.id === selectedCategory);
		}
		return undefined;
	});

	// -------------------------------------------------------------------------
	// Step 1 handlers
	// -------------------------------------------------------------------------

	function selectTopLevel(catId: string) {
		selectedCategory = catId;
		selectedSubcategory = null;
		showAllSubcategories = false;

		const subs = getSubcategories(catId);
		if (subs.length === 0) {
			// No subcategories -- go straight to location step
			goToStep2();
		}
		// Otherwise show subcategory picker
	}

	function selectSubcategory(subId: string) {
		selectedSubcategory = subId;
		goToStep2();
	}

	function goBackToTopLevel() {
		selectedCategory = null;
		selectedSubcategory = null;
	}

	// -------------------------------------------------------------------------
	// Step 2 -- location
	// -------------------------------------------------------------------------

	async function goToStep2() {
		step = 2;
		locationError = '';
		showManualInput = false;

		// Auto-detect location
		locationLoading = true;
		try {
			location = await detectLocation();
		} catch (err: any) {
			locationError = err?.message || t('wizard.location_error');
			location = null;
		} finally {
			locationLoading = false;
		}
	}

	function useDetectedLocation() {
		goToStep3();
	}

	function showManual() {
		showManualInput = true;
	}

	function submitManualPostcode() {
		if (!manualPostcode.trim()) return;
		goToStep3();
	}

	// -------------------------------------------------------------------------
	// Step 3 -- redirect
	// -------------------------------------------------------------------------

	function goToStep3() {
		step = 3;

		const cat = resolvedCategory;
		if (!cat) return;

		const categorySlug = cat.slug;

		// Determine the redirect URL
		setTimeout(() => {
			if (location?.area && location?.city) {
				// SEO-friendly URL
				const citySlug = slugify(location.city);
				const areaSlug = slugify(location.area);
				goto(`/${categorySlug}/${citySlug}/${areaSlug}`);
			} else {
				const postcode = location?.postcode || manualPostcode.trim();
				goto(`/providers?category=${categorySlug}&postcode=${encodeURIComponent(postcode)}`);
			}
		}, 1200);
	}

	function goBack() {
		if (step === 2) {
			step = 1;
			location = null;
			locationError = '';
			showManualInput = false;
		}
	}

	// -------------------------------------------------------------------------
	// Helpers
	// -------------------------------------------------------------------------

	function slugify(text: string): string {
		return text
			.toLowerCase()
			.replace(/[^\w\s-]/g, '')
			.replace(/\s+/g, '-')
			.replace(/-+/g, '-')
			.trim();
	}

	function handleVoiceResult(text: string) {
		searchQuery = text;
	}
</script>

<!-- ====================================================================== -->
<!-- Service Wizard                                                         -->
<!-- ====================================================================== -->

<div class="mx-auto w-full max-w-3xl px-4 py-6">
	<!-- ------------------------------------------------------------------ -->
	<!-- Step Indicator                                                      -->
	<!-- ------------------------------------------------------------------ -->
	<div class="mb-8 flex items-center justify-center gap-3">
		{#each [1, 2, 3] as s}
			<div class="flex items-center gap-3">
				<!-- Circle -->
				<div
					class="flex h-9 w-9 items-center justify-center rounded-full text-sm font-semibold transition-all duration-300
						{step > s
							? 'bg-primary-600 text-white dark:bg-primary-500'
							: step === s
								? 'bg-primary-600 text-white shadow-lg ring-4 ring-primary-100 dark:bg-primary-500 dark:ring-primary-900/40'
								: 'border-2 border-gray-300 text-gray-400 dark:border-gray-600 dark:text-gray-500'}"
				>
					{#if step > s}
						<Check class="h-4 w-4" />
					{:else}
						{s}
					{/if}
				</div>

				<!-- Connector line (not after last) -->
				{#if s < 3}
					<div
						class="h-0.5 w-10 rounded-full transition-colors duration-300
							{step > s
								? 'bg-primary-600 dark:bg-primary-500'
								: 'bg-gray-200 dark:bg-gray-700'}"
					></div>
				{/if}
			</div>
		{/each}
	</div>

	<!-- ------------------------------------------------------------------ -->
	<!-- Step 1: Pick a Category                                            -->
	<!-- ------------------------------------------------------------------ -->
	{#if step === 1}
		<div class="wizard-fade-in">
			<!-- Heading -->
			<h2 class="mb-1 text-center text-2xl font-bold text-gray-900 dark:text-white">
				{t('wizard.pick_category')}
			</h2>
			<p class="mb-6 text-center text-sm text-gray-500 dark:text-gray-400">
				{t('wizard.pick_category_sub')}
			</p>

			<!-- Search + Voice Input -->
			<div class="mb-6 flex items-center gap-2">
				<div class="relative flex-1">
					<div class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
						<Search class="h-5 w-5" />
					</div>
					<input
						type="text"
						bind:value={searchQuery}
						placeholder={t('wizard.search_placeholder')}
						class="w-full rounded-xl border border-gray-300 bg-white py-3 pl-10 pr-4 text-sm text-gray-900 placeholder-gray-400 transition-colors focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-800 dark:text-white dark:placeholder-gray-400"
					/>
				</div>
				<VoiceInput
					lang={speechCode}
					onresult={handleVoiceResult}
					size="lg"
				/>
			</div>

			<!-- Subcategory view (when a top-level is selected) -->
			{#if selectedCategory}
				{@const parentCat = categories.find((c) => c.id === selectedCategory)}

				<!-- Back to top-level -->
				<button
					type="button"
					onclick={goBackToTopLevel}
					class="mb-4 inline-flex items-center gap-1.5 text-sm font-medium text-primary-600 transition-colors hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
				>
					<ChevronLeft class="h-4 w-4" />
					{t('wizard.back_to_categories')}
				</button>

				{#if parentCat}
					<h3 class="mb-4 text-lg font-semibold text-gray-800 dark:text-gray-200">
						{t(parentCat.translationKey)}
					</h3>
				{/if}

				<!-- Subcategory grid -->
				<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
					{#each filteredSubcategories as sub (sub.id)}
						{@const Icon = iconMap[sub.icon] || Wrench}
						<button
							type="button"
							onclick={() => selectSubcategory(sub.id)}
							class="group flex min-h-[60px] flex-col items-center justify-center gap-2 rounded-xl border border-gray-200 bg-white p-4 text-center transition-all duration-200 hover:border-primary-300 hover:shadow-md active:scale-[0.97] dark:border-gray-700 dark:bg-gray-800 dark:hover:border-primary-600"
						>
							<div class="flex h-10 w-10 items-center justify-center rounded-lg {sub.color} transition-transform duration-200 group-hover:scale-110">
								<Icon class="h-5 w-5" />
							</div>
							<span class="text-sm font-medium text-gray-700 dark:text-gray-300">
								{t(sub.translationKey)}
							</span>
						</button>
					{/each}
				</div>

				{#if filteredSubcategories.length === 0}
					<p class="py-8 text-center text-sm text-gray-500 dark:text-gray-400">
						{t('wizard.no_results')}
					</p>
				{/if}

			<!-- Top-level category grid -->
			{:else}
				<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
					{#each filteredTopLevel as cat (cat.id)}
						{@const Icon = iconMap[cat.icon] || Wrench}
						{@const subCount = getSubcategories(cat.id).length}
						<button
							type="button"
							onclick={() => selectTopLevel(cat.id)}
							class="group flex min-h-[80px] flex-col items-center justify-center gap-2.5 rounded-xl border border-gray-200 bg-white p-4 text-center transition-all duration-200 hover:border-primary-300 hover:shadow-md active:scale-[0.97] dark:border-gray-700 dark:bg-gray-800 dark:hover:border-primary-600"
						>
							<div class="flex h-12 w-12 items-center justify-center rounded-xl {cat.color} transition-transform duration-200 group-hover:scale-110">
								<Icon class="h-6 w-6" />
							</div>
							<span class="text-sm font-medium text-gray-700 dark:text-gray-300">
								{t(cat.translationKey)}
							</span>
						</button>
					{/each}
				</div>

				{#if filteredTopLevel.length === 0}
					<p class="py-8 text-center text-sm text-gray-500 dark:text-gray-400">
						{t('wizard.no_results')}
					</p>
				{/if}

				<!-- "See all services" expand -->
				{#if !searchQuery.trim() && filteredTopLevel.length > 0}
					<div class="mt-6 text-center">
						<button
							type="button"
							onclick={() => (showAllSubcategories = !showAllSubcategories)}
							class="inline-flex items-center gap-1 text-sm font-medium text-primary-600 transition-colors hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
						>
							{#if showAllSubcategories}
								{t('wizard.hide_all_services')}
							{:else}
								{t('wizard.see_all_services', { count: categories.filter((c) => c.parentId !== null).length })}
							{/if}
							<ChevronRight
								class="h-4 w-4 transition-transform duration-200 {showAllSubcategories ? 'rotate-90' : ''}"
							/>
						</button>
					</div>

					{#if showAllSubcategories}
						<div class="mt-4 space-y-6">
							{#each topLevelCategories as parent (parent.id)}
								{@const subs = getSubcategories(parent.id)}
								{#if subs.length > 0}
									<div>
										<h4 class="mb-2 text-sm font-semibold text-gray-600 dark:text-gray-400">
											{t(parent.translationKey)}
										</h4>
										<div class="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-4">
											{#each subs as sub (sub.id)}
												{@const SubIcon = iconMap[sub.icon] || Wrench}
												<button
													type="button"
													onclick={() => {
														selectedCategory = parent.id;
														selectSubcategory(sub.id);
													}}
													class="group flex min-h-[56px] items-center gap-3 rounded-lg border border-gray-200 bg-white px-3 py-2.5 text-left transition-all duration-200 hover:border-primary-300 hover:shadow-sm active:scale-[0.98] dark:border-gray-700 dark:bg-gray-800 dark:hover:border-primary-600"
												>
													<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg {sub.color}">
														<SubIcon class="h-4 w-4" />
													</div>
													<span class="text-sm text-gray-700 dark:text-gray-300">
														{t(sub.translationKey)}
													</span>
												</button>
											{/each}
										</div>
									</div>
								{/if}
							{/each}
						</div>
					{/if}
				{/if}
			{/if}
		</div>
	{/if}

	<!-- ------------------------------------------------------------------ -->
	<!-- Step 2: Confirm Location                                           -->
	<!-- ------------------------------------------------------------------ -->
	{#if step === 2}
		<div class="wizard-fade-in">
			<!-- Back button -->
			<button
				type="button"
				onclick={goBack}
				class="mb-6 inline-flex items-center gap-1.5 text-sm font-medium text-primary-600 transition-colors hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
			>
				<ChevronLeft class="h-4 w-4" />
				{t('wizard.back')}
			</button>

			<div class="mx-auto max-w-md text-center">
				<!-- Loading state -->
				{#if locationLoading}
					<div class="flex flex-col items-center gap-4 py-12">
						<div class="flex h-16 w-16 items-center justify-center rounded-full bg-primary-50 dark:bg-primary-900/20">
							<Loader2 class="h-8 w-8 animate-spin text-primary-600 dark:text-primary-400" />
						</div>
						<p class="text-lg font-medium text-gray-700 dark:text-gray-300">
							{t('wizard.detecting_location')}
						</p>
						<p class="text-sm text-gray-500 dark:text-gray-400">
							{t('wizard.detecting_location_sub')}
						</p>
					</div>

				<!-- Location detected -->
				{:else if location && !showManualInput}
					<div class="flex flex-col items-center gap-5 py-8">
						<div class="flex h-16 w-16 items-center justify-center rounded-full bg-green-50 dark:bg-green-900/20">
							<MapPin class="h-8 w-8 text-green-600 dark:text-green-400" />
						</div>

						<div>
							<p class="text-base text-gray-600 dark:text-gray-400">
								{t('wizard.detected_in')}
							</p>
							<p class="mt-1 text-xl font-bold text-gray-900 dark:text-white">
								{location.displayName}
							</p>
						</div>

						<div class="flex w-full flex-col gap-3 sm:flex-row sm:justify-center">
							<button
								type="button"
								onclick={useDetectedLocation}
								class="inline-flex items-center justify-center gap-2 rounded-xl bg-primary-600 px-8 py-3.5 text-base font-semibold text-white shadow-sm transition-colors hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 dark:bg-primary-500 dark:hover:bg-primary-600"
							>
								<MapPin class="h-5 w-5" />
								{t('wizard.use_location')}
							</button>
						</div>

						<button
							type="button"
							onclick={showManual}
							class="text-sm text-gray-500 underline transition-colors hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300"
						>
							{t('wizard.change_location')}
						</button>
					</div>

				<!-- Error or manual input -->
				{:else}
					<div class="flex flex-col items-center gap-5 py-8">
						{#if locationError && !showManualInput}
							<div class="flex h-16 w-16 items-center justify-center rounded-full bg-red-50 dark:bg-red-900/20">
								<MapPin class="h-8 w-8 text-red-500 dark:text-red-400" />
							</div>
							<p class="text-sm text-red-600 dark:text-red-400">
								{locationError}
							</p>
						{/if}

						<!-- Manual postcode input -->
						<div class="w-full max-w-sm">
							<label
								for="manual-postcode"
								class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
							>
								{t('wizard.enter_postcode')}
							</label>

							<div class="flex gap-2">
								<div class="relative flex-1">
									<div class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
										<MapPin class="h-5 w-5" />
									</div>
									<input
										id="manual-postcode"
										type="text"
										bind:value={manualPostcode}
										placeholder={t('wizard.postcode_placeholder')}
										onkeydown={(e) => { if (e.key === 'Enter') submitManualPostcode(); }}
										class="w-full rounded-xl border border-gray-300 bg-white py-3 pl-10 pr-4 text-sm text-gray-900 placeholder-gray-400 transition-colors focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-800 dark:text-white dark:placeholder-gray-400"
									/>
								</div>
								<button
									type="button"
									onclick={submitManualPostcode}
									disabled={!manualPostcode.trim()}
									class="inline-flex items-center justify-center gap-2 rounded-xl bg-primary-600 px-6 py-3 text-sm font-semibold text-white transition-colors hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-primary-500 dark:hover:bg-primary-600"
								>
									<ChevronRight class="h-4 w-4" />
									{t('wizard.continue')}
								</button>
							</div>
						</div>

						<!-- Retry detection link (only if error) -->
						{#if locationError}
							<button
								type="button"
								onclick={goToStep2}
								class="mt-2 text-sm text-primary-600 underline transition-colors hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
							>
								{t('wizard.retry_detection')}
							</button>
						{/if}
					</div>
				{/if}
			</div>
		</div>
	{/if}

	<!-- ------------------------------------------------------------------ -->
	<!-- Step 3: Redirect                                                   -->
	<!-- ------------------------------------------------------------------ -->
	{#if step === 3}
		<div class="wizard-fade-in">
			<div class="flex flex-col items-center gap-5 py-16">
				<div class="flex h-16 w-16 items-center justify-center rounded-full bg-primary-50 dark:bg-primary-900/20">
					<Loader2 class="h-8 w-8 animate-spin text-primary-600 dark:text-primary-400" />
				</div>
				<p class="text-lg font-medium text-gray-700 dark:text-gray-300">
					{t('wizard.finding_providers')}
				</p>
				{#if resolvedCategory}
					<p class="text-sm text-gray-500 dark:text-gray-400">
						{t(resolvedCategory.translationKey)}
						{#if location?.displayName}
							&middot; {location.displayName}
						{:else if manualPostcode}
							&middot; {manualPostcode}
						{/if}
					</p>
				{/if}
			</div>
		</div>
	{/if}
</div>

<style>
	@keyframes wizardFadeIn {
		from {
			opacity: 0;
			transform: translateY(8px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.wizard-fade-in {
		animation: wizardFadeIn 0.3s ease-out;
	}
</style>
