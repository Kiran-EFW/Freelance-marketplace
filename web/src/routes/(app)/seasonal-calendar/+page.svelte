<script lang="ts">
	import { ChevronLeft, ChevronRight, Leaf, Search, Sun, CloudRain, Snowflake, Sprout, Trees, Wheat, Apple, Coffee } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import { t } from '$lib/i18n/index.svelte';

	const months = [
		'January', 'February', 'March', 'April', 'May', 'June',
		'July', 'August', 'September', 'October', 'November', 'December'
	];

	let selectedMonth = $state(new Date().getMonth()); // 0-indexed
	let filterCrop = $state('');
	let searchQuery = $state('');

	// Crop icons mapping
	const cropIcons: Record<string, typeof Leaf> = {
		coconut: Trees,
		rice: Wheat,
		wheat: Wheat,
		sugarcane: Sprout,
		cotton: Leaf,
		arecanut: Trees,
		rubber: Trees,
		tea: Coffee,
		coffee: Coffee,
		banana: Leaf,
		mango: Apple,
	};

	// Season indicator per month
	function getSeason(month: number): { name: string; icon: typeof Sun; color: string } {
		if (month >= 2 && month <= 4) return { name: 'Summer', icon: Sun, color: 'text-orange-500' };
		if (month >= 5 && month <= 8) return { name: 'Monsoon', icon: CloudRain, color: 'text-blue-500' };
		if (month >= 9 && month <= 10) return { name: 'Autumn', icon: Leaf, color: 'text-amber-500' };
		return { name: 'Winter', icon: Snowflake, color: 'text-sky-400' };
	}

	// Crop catalog data (same as backend seed)
	interface PriceRange {
		min: number;
		max: number;
		currency: string;
	}

	interface WorkType {
		slug: string;
		name: string;
		pricing_model: string;
		typical_price: PriceRange;
	}

	interface CropEntry {
		slug: string;
		name: string;
		name_local?: string;
		work_types: WorkType[];
		seasonal_calendar: Record<string, string[]>;
	}

	const crops: CropEntry[] = [
		{
			slug: 'coconut', name: 'Coconut', name_local: 'Theng',
			work_types: [
				{ slug: 'tree_climbing', name: 'Tree Climbing & Harvesting', pricing_model: 'per_tree', typical_price: { min: 30, max: 80, currency: 'INR' } },
				{ slug: 'tree_pruning', name: 'Frond Cutting & Pruning', pricing_model: 'per_tree', typical_price: { min: 20, max: 50, currency: 'INR' } },
				{ slug: 'pest_treatment', name: 'Pest Treatment', pricing_model: 'per_tree', typical_price: { min: 50, max: 150, currency: 'INR' } },
				{ slug: 'fertilizing', name: 'Fertilizer Application', pricing_model: 'per_tree', typical_price: { min: 30, max: 80, currency: 'INR' } }
			],
			seasonal_calendar: { '1': ['tree_climbing', 'fertilizing'], '2': ['tree_climbing', 'pest_treatment'], '3': ['tree_climbing', 'tree_pruning'], '4': ['tree_climbing', 'fertilizing'], '5': ['tree_climbing'], '6': ['tree_climbing', 'pest_treatment'], '7': ['tree_climbing', 'tree_pruning'], '8': ['tree_climbing', 'fertilizing'], '9': ['tree_climbing'], '10': ['tree_climbing', 'pest_treatment'], '11': ['tree_climbing', 'fertilizing'], '12': ['tree_climbing', 'tree_pruning'] }
		},
		{
			slug: 'rice', name: 'Rice / Paddy', name_local: 'Dhaan',
			work_types: [
				{ slug: 'ploughing', name: 'Ploughing & Land Preparation', pricing_model: 'per_day', typical_price: { min: 1500, max: 3000, currency: 'INR' } },
				{ slug: 'transplanting', name: 'Transplanting', pricing_model: 'per_day', typical_price: { min: 500, max: 800, currency: 'INR' } },
				{ slug: 'harvesting', name: 'Harvesting & Threshing', pricing_model: 'per_day', typical_price: { min: 2000, max: 5000, currency: 'INR' } },
				{ slug: 'spraying', name: 'Pesticide Spraying', pricing_model: 'per_day', typical_price: { min: 800, max: 1500, currency: 'INR' } }
			],
			seasonal_calendar: { '1': [], '2': [], '3': [], '4': ['ploughing'], '5': ['ploughing', 'transplanting'], '6': ['transplanting', 'spraying'], '7': ['spraying'], '8': ['spraying'], '9': ['harvesting'], '10': ['harvesting'], '11': [], '12': [] }
		},
		{
			slug: 'wheat', name: 'Wheat', name_local: 'Gehun',
			work_types: [
				{ slug: 'ploughing', name: 'Ploughing & Land Preparation', pricing_model: 'per_day', typical_price: { min: 1500, max: 3000, currency: 'INR' } },
				{ slug: 'sowing', name: 'Sowing & Seed Drilling', pricing_model: 'per_day', typical_price: { min: 1000, max: 2000, currency: 'INR' } },
				{ slug: 'irrigation', name: 'Irrigation Management', pricing_model: 'per_day', typical_price: { min: 500, max: 1000, currency: 'INR' } },
				{ slug: 'spraying', name: 'Pesticide & Weed Spraying', pricing_model: 'per_day', typical_price: { min: 800, max: 1500, currency: 'INR' } },
				{ slug: 'harvesting', name: 'Harvesting & Threshing', pricing_model: 'per_day', typical_price: { min: 2000, max: 4000, currency: 'INR' } }
			],
			seasonal_calendar: { '1': ['irrigation', 'spraying'], '2': ['irrigation', 'spraying'], '3': ['harvesting'], '4': ['harvesting'], '5': [], '6': [], '7': [], '8': [], '9': [], '10': ['ploughing'], '11': ['ploughing', 'sowing'], '12': ['sowing', 'irrigation'] }
		},
		{
			slug: 'sugarcane', name: 'Sugarcane', name_local: 'Ganna',
			work_types: [
				{ slug: 'ploughing', name: 'Land Preparation & Ridging', pricing_model: 'per_day', typical_price: { min: 1500, max: 3500, currency: 'INR' } },
				{ slug: 'planting', name: 'Sett Planting', pricing_model: 'per_day', typical_price: { min: 600, max: 1200, currency: 'INR' } },
				{ slug: 'weeding', name: 'Weeding & Earthing Up', pricing_model: 'per_day', typical_price: { min: 400, max: 800, currency: 'INR' } },
				{ slug: 'spraying', name: 'Pesticide Spraying', pricing_model: 'per_day', typical_price: { min: 800, max: 1500, currency: 'INR' } },
				{ slug: 'harvesting', name: 'Harvesting & Loading', pricing_model: 'per_day', typical_price: { min: 2500, max: 5000, currency: 'INR' } }
			],
			seasonal_calendar: { '1': ['harvesting'], '2': ['ploughing', 'planting'], '3': ['planting', 'weeding'], '4': ['weeding', 'spraying'], '5': ['weeding', 'spraying'], '6': ['spraying'], '7': ['spraying'], '8': ['weeding'], '9': ['weeding'], '10': ['spraying'], '11': ['harvesting'], '12': ['harvesting'] }
		},
		{
			slug: 'cotton', name: 'Cotton', name_local: 'Kapas',
			work_types: [
				{ slug: 'ploughing', name: 'Land Preparation', pricing_model: 'per_day', typical_price: { min: 1500, max: 3000, currency: 'INR' } },
				{ slug: 'sowing', name: 'Sowing', pricing_model: 'per_day', typical_price: { min: 800, max: 1500, currency: 'INR' } },
				{ slug: 'spraying', name: 'Pesticide & Bollworm Spraying', pricing_model: 'per_day', typical_price: { min: 1000, max: 2000, currency: 'INR' } },
				{ slug: 'picking', name: 'Cotton Picking', pricing_model: 'per_day', typical_price: { min: 400, max: 700, currency: 'INR' } },
				{ slug: 'weeding', name: 'Weeding & Thinning', pricing_model: 'per_day', typical_price: { min: 400, max: 800, currency: 'INR' } }
			],
			seasonal_calendar: { '1': [], '2': [], '3': [], '4': ['ploughing'], '5': ['ploughing', 'sowing'], '6': ['sowing', 'weeding'], '7': ['weeding', 'spraying'], '8': ['spraying'], '9': ['spraying', 'picking'], '10': ['picking'], '11': ['picking'], '12': [] }
		},
		{
			slug: 'mango', name: 'Mango', name_local: 'Aam',
			work_types: [
				{ slug: 'pruning', name: 'Pruning & Dead Wood Removal', pricing_model: 'per_tree', typical_price: { min: 50, max: 200, currency: 'INR' } },
				{ slug: 'spraying', name: 'Pest & Fungal Spraying', pricing_model: 'per_tree', typical_price: { min: 40, max: 120, currency: 'INR' } },
				{ slug: 'harvesting', name: 'Fruit Harvesting', pricing_model: 'per_tree', typical_price: { min: 30, max: 100, currency: 'INR' } },
				{ slug: 'fertilizing', name: 'Fertilizer Application', pricing_model: 'per_tree', typical_price: { min: 30, max: 80, currency: 'INR' } }
			],
			seasonal_calendar: { '1': ['spraying'], '2': ['spraying'], '3': ['spraying', 'harvesting'], '4': ['harvesting'], '5': ['harvesting'], '6': ['harvesting', 'pruning'], '7': ['pruning'], '8': ['fertilizing'], '9': ['fertilizing'], '10': ['spraying'], '11': ['spraying'], '12': ['spraying'] }
		},
		{
			slug: 'banana', name: 'Banana', name_local: 'Kela',
			work_types: [
				{ slug: 'planting', name: 'Sucker Planting', pricing_model: 'per_day', typical_price: { min: 500, max: 1000, currency: 'INR' } },
				{ slug: 'deleafing', name: 'De-leafing & Propping', pricing_model: 'per_day', typical_price: { min: 400, max: 700, currency: 'INR' } },
				{ slug: 'spraying', name: 'Pest & Disease Spraying', pricing_model: 'per_day', typical_price: { min: 800, max: 1500, currency: 'INR' } },
				{ slug: 'harvesting', name: 'Bunch Harvesting', pricing_model: 'per_day', typical_price: { min: 600, max: 1200, currency: 'INR' } },
				{ slug: 'fertilizing', name: 'Fertilizer & Manure Application', pricing_model: 'per_day', typical_price: { min: 500, max: 900, currency: 'INR' } }
			],
			seasonal_calendar: { '1': ['deleafing', 'fertilizing'], '2': ['planting', 'fertilizing'], '3': ['planting', 'spraying'], '4': ['spraying', 'deleafing'], '5': ['spraying', 'deleafing'], '6': ['planting', 'spraying'], '7': ['spraying', 'fertilizing'], '8': ['deleafing', 'harvesting'], '9': ['harvesting'], '10': ['harvesting', 'fertilizing'], '11': ['deleafing', 'spraying'], '12': ['harvesting', 'deleafing'] }
		},
		{
			slug: 'coffee', name: 'Coffee', name_local: 'Kaapi',
			work_types: [
				{ slug: 'picking', name: 'Cherry Picking', pricing_model: 'per_day', typical_price: { min: 400, max: 700, currency: 'INR' } },
				{ slug: 'pruning', name: 'Shade Tree Pruning', pricing_model: 'per_day', typical_price: { min: 600, max: 1000, currency: 'INR' } },
				{ slug: 'spraying', name: 'Pest & Borer Spraying', pricing_model: 'per_day', typical_price: { min: 800, max: 1500, currency: 'INR' } },
				{ slug: 'processing', name: 'Pulping & Drying', pricing_model: 'per_day', typical_price: { min: 500, max: 900, currency: 'INR' } }
			],
			seasonal_calendar: { '1': ['picking', 'processing'], '2': ['processing', 'pruning'], '3': ['pruning'], '4': ['spraying'], '5': ['spraying'], '6': [], '7': [], '8': ['spraying'], '9': ['spraying'], '10': ['picking'], '11': ['picking'], '12': ['picking', 'processing'] }
		}
	];

	function getPricingLabel(model: string): string {
		const labels: Record<string, string> = {
			per_tree: t('crop_calendar.per_tree'),
			per_day: t('crop_calendar.per_day'),
			per_sqft: t('crop_calendar.per_sqft')
		};
		return labels[model] || model;
	}

	function formatPrice(price: PriceRange): string {
		return `${price.currency === 'INR' ? '\u20B9' : price.currency}${price.min}-${price.max}`;
	}

	function prevMonth() {
		selectedMonth = selectedMonth === 0 ? 11 : selectedMonth - 1;
	}

	function nextMonth() {
		selectedMonth = selectedMonth === 11 ? 0 : selectedMonth + 1;
	}

	const season = $derived(getSeason(selectedMonth));
	const monthKey = $derived(String(selectedMonth + 1)); // 1-indexed for calendar lookup
	const isCurrentMonth = $derived(selectedMonth === new Date().getMonth());

	const filteredCrops = $derived(
		crops
			.filter((c) => {
				if (filterCrop && c.slug !== filterCrop) return false;
				if (searchQuery) {
					const q = searchQuery.toLowerCase();
					return c.name.toLowerCase().includes(q) || c.slug.includes(q) ||
						c.work_types.some((wt) => wt.name.toLowerCase().includes(q));
				}
				return true;
			})
			.map((c) => {
				const inSeason = c.seasonal_calendar[monthKey] || [];
				return {
					...c,
					inSeasonWorkTypes: c.work_types.filter((wt) => inSeason.includes(wt.slug)),
					offSeasonWorkTypes: c.work_types.filter((wt) => !inSeason.includes(wt.slug)),
					hasWorkThisMonth: inSeason.length > 0
				};
			})
	);

	const cropsWithWork = $derived(filteredCrops.filter((c) => c.hasWorkThisMonth));
	const cropsWithoutWork = $derived(filteredCrops.filter((c) => !c.hasWorkThisMonth));
</script>

<svelte:head>
	<title>{t('crop_calendar.title')} - Seva</title>
</svelte:head>

<div class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
	<!-- Header -->
	<div class="text-center">
		<h1 class="text-2xl font-bold text-gray-900 dark:text-white">{t('crop_calendar.title')}</h1>
		<p class="mt-2 text-sm text-gray-500 dark:text-gray-400">{t('crop_calendar.subtitle')}</p>
	</div>

	<!-- Month Selector Carousel -->
	<div class="mt-8">
		<div class="flex items-center justify-center gap-4">
			<button
				onclick={prevMonth}
				class="rounded-full p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-800 dark:hover:text-gray-300"
			>
				<ChevronLeft class="h-5 w-5" />
			</button>

			<div class="flex items-center gap-2 overflow-x-auto py-2 scrollbar-hide">
				{#each months as month, i}
					<button
						onclick={() => (selectedMonth = i)}
						class="shrink-0 rounded-full px-4 py-2 text-sm font-medium transition-all {selectedMonth === i
							? 'bg-primary-600 text-white shadow-md'
							: i === new Date().getMonth()
								? 'bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400'
								: 'text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800'}"
					>
						{month.slice(0, 3)}
					</button>
				{/each}
			</div>

			<button
				onclick={nextMonth}
				class="rounded-full p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-800 dark:hover:text-gray-300"
			>
				<ChevronRight class="h-5 w-5" />
			</button>
		</div>

		<!-- Current month + season indicator -->
		<div class="mt-3 flex items-center justify-center gap-3">
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">
				{months[selectedMonth]}
				{#if isCurrentMonth}
					<Badge variant="info" size="sm">{t('crop_calendar.this_month')}</Badge>
				{/if}
			</h2>
			{#if season.icon}
				{@const SeasonIcon = season.icon}
				<span class="flex items-center gap-1 text-sm {season.color}">
					<SeasonIcon class="h-4 w-4" />
					{season.name}
				</span>
			{/if}
		</div>
	</div>

	<!-- Search & Filter -->
	<div class="mt-6 flex flex-col gap-3 sm:flex-row">
		<div class="relative flex-1">
			<Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Search crops or work types..."
				class="w-full rounded-lg border border-gray-300 py-2.5 pl-10 pr-4 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-800 dark:text-white"
			/>
		</div>
		<select
			bind:value={filterCrop}
			class="rounded-lg border border-gray-300 px-4 py-2.5 text-sm focus:border-primary-500 focus:outline-none dark:border-gray-600 dark:bg-gray-800 dark:text-white"
		>
			<option value="">All Crops</option>
			{#each crops as crop}
				<option value={crop.slug}>{crop.name}</option>
			{/each}
		</select>
	</div>

	<!-- In-Season Crops -->
	{#if cropsWithWork.length > 0}
		<div class="mt-8">
			<h3 class="flex items-center gap-2 text-base font-semibold text-green-700 dark:text-green-400">
				<Sprout class="h-5 w-5" />
				{t('crop_calendar.in_season')} ({cropsWithWork.length} crops)
			</h3>
			<div class="mt-4 space-y-4">
				{#each cropsWithWork as crop}
					{@const CropIcon = cropIcons[crop.slug] || Leaf}
					<Card>
						<div class="flex items-start gap-4">
							<div class="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400">
								<CropIcon class="h-6 w-6" />
							</div>
							<div class="flex-1 min-w-0">
								<div class="flex items-center gap-2">
									<h4 class="font-semibold text-gray-900 dark:text-white">{crop.name}</h4>
									{#if crop.name_local}
										<span class="text-xs text-gray-400">({crop.name_local})</span>
									{/if}
									<Badge variant="success" size="sm">{t('crop_calendar.in_season')}</Badge>
								</div>

								<!-- In-season work types -->
								<div class="mt-3 space-y-2">
									{#each crop.inSeasonWorkTypes as wt}
										<div class="flex flex-wrap items-center justify-between gap-2 rounded-lg border border-green-100 bg-green-50/50 p-3 dark:border-green-900/30 dark:bg-green-900/10">
											<div>
												<p class="text-sm font-medium text-gray-900 dark:text-white">{wt.name}</p>
												<p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
													{t('crop_calendar.typical_price')}: {formatPrice(wt.typical_price)} {getPricingLabel(wt.pricing_model)}
												</p>
											</div>
											<a
												href="/providers?category={crop.slug}_{wt.slug}"
												class="shrink-0 rounded-lg bg-primary-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-primary-700 transition-colors"
											>
												{t('crop_calendar.find_provider')}
											</a>
										</div>
									{/each}
								</div>

								<!-- Off-season work types (collapsed) -->
								{#if crop.offSeasonWorkTypes.length > 0}
									<div class="mt-2">
										<p class="text-xs text-gray-400 dark:text-gray-500">
											{t('crop_calendar.off_season')}: {crop.offSeasonWorkTypes.map((wt) => wt.name).join(', ')}
										</p>
									</div>
								{/if}
							</div>
						</div>
					</Card>
				{/each}
			</div>
		</div>
	{/if}

	<!-- Off-Season Crops -->
	{#if cropsWithoutWork.length > 0}
		<div class="mt-8">
			<h3 class="flex items-center gap-2 text-base font-semibold text-gray-400 dark:text-gray-500">
				<Leaf class="h-5 w-5" />
				{t('crop_calendar.off_season')} ({cropsWithoutWork.length} crops)
			</h3>
			<div class="mt-4 space-y-2">
				{#each cropsWithoutWork as crop}
					{@const CropIcon = cropIcons[crop.slug] || Leaf}
					<Card>
						<div class="flex items-center gap-4 opacity-60">
							<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-gray-100 text-gray-400 dark:bg-gray-800 dark:text-gray-500">
								<CropIcon class="h-5 w-5" />
							</div>
							<div class="flex-1 min-w-0">
								<div class="flex items-center gap-2">
									<h4 class="font-medium text-gray-700 dark:text-gray-400">{crop.name}</h4>
									<Badge variant="neutral" size="sm">{t('crop_calendar.off_season')}</Badge>
								</div>
								<p class="mt-0.5 text-xs text-gray-400 dark:text-gray-500">
									{crop.work_types.length} work types available in other months
								</p>
							</div>
						</div>
					</Card>
				{/each}
			</div>
		</div>
	{/if}

	<!-- No results -->
	{#if filteredCrops.length === 0}
		<div class="mt-12 text-center">
			<Leaf class="mx-auto h-12 w-12 text-gray-300 dark:text-gray-600" />
			<p class="mt-4 text-gray-500 dark:text-gray-400">No crops match your search.</p>
		</div>
	{/if}

	<!-- Calendar Overview (mini 12-month grid) -->
	<div class="mt-12">
		<h3 class="text-base font-semibold text-gray-900 dark:text-white">Year-Round Overview</h3>
		<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">Number of active crop activities per month</p>
		<div class="mt-4 grid grid-cols-6 gap-2 sm:grid-cols-12">
			{#each months as month, i}
				{@const monthActivities = crops.reduce((sum, c) => sum + (c.seasonal_calendar[String(i + 1)]?.length || 0), 0)}
				{@const intensity = Math.min(monthActivities / 20, 1)}
				<button
					onclick={() => (selectedMonth = i)}
					class="group relative rounded-lg p-2 text-center transition-all
						{selectedMonth === i ? 'ring-2 ring-primary-500 ring-offset-1' : ''}
						{i === new Date().getMonth() ? 'font-bold' : ''}
					"
					style="background-color: rgba(34, 197, 94, {0.1 + intensity * 0.5});"
				>
					<span class="block text-[10px] font-medium text-gray-700 dark:text-gray-300">{month.slice(0, 3)}</span>
					<span class="block text-lg font-bold text-gray-900 dark:text-white">{monthActivities}</span>
				</button>
			{/each}
		</div>
	</div>
</div>

<style>
	.scrollbar-hide::-webkit-scrollbar {
		display: none;
	}
	.scrollbar-hide {
		-ms-overflow-style: none;
		scrollbar-width: none;
	}
</style>
