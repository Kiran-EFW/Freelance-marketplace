<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { BarChart3, TrendingUp, MapPin, Clock, Users, Lightbulb, Loader2, IndianRupee, Star, CheckCircle, MessageSquare } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import { analytics } from '$lib/api/client';
	import { t } from '$lib/i18n/index.svelte';
	import { toastError } from '$lib/stores/toast';
	import type L from 'leaflet';

	type Period = '7d' | '30d' | '90d' | '12m';

	let loading = $state(true);
	let selectedPeriod = $state<Period>('12m');

	// Data stores
	let earningsHistory = $state<{ month: string; earnings: number; job_count: number }[]>([]);
	let demandPostcodes = $state<{ postcode: string; demand_count: number; lat: number; lng: number }[]>([]);
	let demandCategories = $state<{ category_id: string; category_slug: string; category_name: string; demand_count: number }[]>([]);
	let performance = $state<{ response_rate: number; completion_rate: number; avg_rating: number; total_reviews: number; total_earnings: number } | null>(null);
	let peakHours = $state<{ hour_of_day: number; demand_count: number }[]>([]);
	let competitors = $state<{ postcode: string; category_slug: string; category_name: string; provider_count: number }[]>([]);
	let insights = $state<{ type: string; title: string; message: string; impact: string }[]>([]);

	// Map
	let mapContainer = $state<HTMLDivElement | null>(null);
	let map: L.Map | null = null;
	let leaflet: typeof L | null = null;
	let circleLayer: L.LayerGroup | null = null;

	const periods: { value: Period; label: string }[] = [
		{ value: '7d', label: 'analytics.period_7d' },
		{ value: '30d', label: 'analytics.period_30d' },
		{ value: '90d', label: 'analytics.period_90d' },
		{ value: '12m', label: 'analytics.period_12m' }
	];

	onMount(async () => {
		await loadAllData();
	});

	onDestroy(() => {
		if (map) {
			map.remove();
			map = null;
		}
	});

	async function loadAllData() {
		loading = true;
		try {
			const [earningsRes, demandRes, perfRes, peakRes, compRes, insightsRes] = await Promise.all([
				analytics.getEarnings({ period: selectedPeriod }).catch(() => null),
				analytics.getDemand({ lat: 12.9716, lng: 77.5946, radius: 25 }).catch(() => null),
				analytics.getPerformance().catch(() => null),
				analytics.getPeakHours().catch(() => null),
				analytics.getCompetitors().catch(() => null),
				analytics.getInsights().catch(() => null)
			]);

			if (earningsRes?.data) {
				earningsHistory = earningsRes.data.history ?? [];
			}
			if (demandRes?.data) {
				demandPostcodes = demandRes.data.postcodes ?? [];
				demandCategories = demandRes.data.categories ?? [];
			}
			if (perfRes?.data) {
				performance = perfRes.data;
			}
			if (peakRes?.data) {
				peakHours = Array.isArray(peakRes.data) ? peakRes.data : [];
			}
			if (compRes?.data) {
				competitors = Array.isArray(compRes.data) ? compRes.data : [];
			}
			if (insightsRes?.data) {
				insights = Array.isArray(insightsRes.data) ? insightsRes.data : [];
			}

			// Initialize map after data is loaded
			await initMap();
		} catch (err) {
			toastError(t('analytics.error'));
			console.error('Failed to load analytics:', err);
		} finally {
			loading = false;
		}
	}

	async function switchPeriod(period: Period) {
		selectedPeriod = period;
		try {
			const res = await analytics.getEarnings({ period });
			if (res?.data) {
				earningsHistory = res.data.history ?? [];
			}
		} catch {
			// Silent fail
		}
	}

	async function initMap() {
		if (!mapContainer || typeof window === 'undefined') return;

		leaflet = (await import('leaflet')).default;

		// Import Leaflet CSS
		const existingLink = document.querySelector('link[href*="leaflet"]');
		if (!existingLink) {
			const linkEl = document.createElement('link');
			linkEl.rel = 'stylesheet';
			linkEl.href = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.css';
			linkEl.integrity = 'sha256-p4NxAoJBhIIN+hmNHrzRCf9tD/miZyoHS5obTRR9BMY=';
			linkEl.crossOrigin = '';
			document.head.appendChild(linkEl);
		}

		map = leaflet.map(mapContainer).setView([12.9716, 77.5946], 11);

		leaflet
			.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
				attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
				maxZoom: 19
			})
			.addTo(map);

		circleLayer = leaflet.layerGroup().addTo(map);
		renderHeatmapCircles();

		setTimeout(() => map?.invalidateSize(), 100);
	}

	function renderHeatmapCircles() {
		if (!map || !leaflet || !circleLayer) return;

		circleLayer.clearLayers();

		if (demandPostcodes.length === 0) return;

		const maxDemand = Math.max(...demandPostcodes.map((d) => d.demand_count));

		for (const d of demandPostcodes) {
			const intensity = d.demand_count / maxDemand;
			const color = getHeatColor(intensity);
			const radius = 300 + intensity * 700;

			const circle = leaflet.circleMarker([d.lat, d.lng], {
				radius: radius / 50,
				fillColor: color,
				fillOpacity: 0.6,
				color: color,
				weight: 2,
				opacity: 0.8
			});

			circle.bindPopup(`
				<div style="font-family: system-ui, sans-serif; min-width: 120px;">
					<p style="margin: 0 0 4px; font-weight: 600; font-size: 14px;">${d.postcode}</p>
					<p style="margin: 0; font-size: 12px; color: #6b7280;">${d.demand_count} job requests</p>
				</div>
			`);

			circleLayer.addLayer(circle);
		}

		// Fit bounds
		if (demandPostcodes.length > 0) {
			const bounds = leaflet.latLngBounds(demandPostcodes.map((d) => [d.lat, d.lng] as [number, number]));
			map.fitBounds(bounds, { padding: [40, 40], maxZoom: 13 });
		}
	}

	function getHeatColor(intensity: number): string {
		if (intensity < 0.33) return '#22c55e'; // green
		if (intensity < 0.66) return '#eab308'; // yellow
		return '#ef4444'; // red
	}

	// SVG earnings chart helpers
	function getEarningsChartPath(): string {
		if (earningsHistory.length < 2) return '';

		const maxEarnings = Math.max(...earningsHistory.map((e) => e.earnings), 1);
		const width = 600;
		const height = 200;
		const padding = 20;
		const chartW = width - padding * 2;
		const chartH = height - padding * 2;

		const points = earningsHistory.map((e, i) => {
			const x = padding + (i / (earningsHistory.length - 1)) * chartW;
			const y = height - padding - (e.earnings / maxEarnings) * chartH;
			return `${x},${y}`;
		});

		return `M ${points.join(' L ')}`;
	}

	function getEarningsChartArea(): string {
		if (earningsHistory.length < 2) return '';

		const maxEarnings = Math.max(...earningsHistory.map((e) => e.earnings), 1);
		const width = 600;
		const height = 200;
		const padding = 20;
		const chartW = width - padding * 2;
		const chartH = height - padding * 2;

		const points = earningsHistory.map((e, i) => {
			const x = padding + (i / (earningsHistory.length - 1)) * chartW;
			const y = height - padding - (e.earnings / maxEarnings) * chartH;
			return `${x},${y}`;
		});

		return `M ${padding},${height - padding} L ${points.join(' L ')} L ${padding + chartW},${height - padding} Z`;
	}

	function getEarningsChartLabels(): { x: number; label: string; earnings: string }[] {
		if (earningsHistory.length === 0) return [];

		const width = 600;
		const padding = 20;
		const chartW = width - padding * 2;

		return earningsHistory.map((e, i) => {
			const x = padding + (i / Math.max(earningsHistory.length - 1, 1)) * chartW;
			const date = new Date(e.month);
			return {
				x,
				label: date.toLocaleDateString(undefined, { month: 'short' }),
				earnings: `Rs. ${e.earnings.toLocaleString()}`
			};
		});
	}

	// Peak hours chart
	function getPeakHoursBars(): { x: number; height: number; hour: string; count: number; width: number }[] {
		if (peakHours.length === 0) return [];

		// Ensure all 24 hours
		const hourMap = new Map(peakHours.map((h) => [h.hour_of_day, h.demand_count]));
		const allHours = Array.from({ length: 24 }, (_, i) => ({
			hour: i,
			count: hourMap.get(i) || 0
		}));

		const maxCount = Math.max(...allHours.map((h) => h.count), 1);
		const width = 600;
		const padding = 30;
		const chartW = width - padding * 2;
		const barW = chartW / 24 - 2;
		const chartH = 160;

		return allHours.map((h) => ({
			x: padding + (h.hour / 24) * chartW + 1,
			height: (h.count / maxCount) * chartH,
			hour: `${h.hour}:00`,
			count: h.count,
			width: barW
		}));
	}

	function formatPercent(value: number): string {
		return `${Math.round(value * 100)}%`;
	}

	function getInsightIcon(type: string): string {
		switch (type) {
			case 'opportunity':
				return 'text-secondary-600 bg-secondary-100 dark:bg-secondary-900/30';
			case 'performance':
				return 'text-primary-600 bg-primary-100 dark:bg-primary-900/30';
			case 'trend':
				return 'text-blue-600 bg-blue-100 dark:bg-blue-900/30';
			default:
				return 'text-gray-600 bg-gray-100 dark:bg-gray-900/30';
		}
	}

	function getImpactBadgeClass(impact: string): string {
		switch (impact) {
			case 'high':
				return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400';
			case 'medium':
				return 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400';
			case 'low':
				return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400';
			default:
				return 'bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-400';
		}
	}

	$effect(() => {
		if (map && leaflet && circleLayer) {
			void demandPostcodes;
			renderHeatmapCircles();
		}
	});
</script>

<svelte:head>
	<title>{t('analytics.title')} - Seva</title>
</svelte:head>

{#if loading}
<div class="flex items-center justify-center py-20">
	<Loader2 class="h-8 w-8 animate-spin text-primary-600" />
	<span class="ml-3 text-sm text-gray-500 dark:text-gray-400">{t('analytics.loading')}</span>
</div>
{:else}
<div class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
	<!-- Header -->
	<div class="flex flex-wrap items-center justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">{t('analytics.title')}</h1>
			<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{t('analytics.subtitle')}</p>
		</div>

		<!-- Period selector tabs -->
		<div class="flex rounded-lg border border-gray-200 bg-gray-50 p-0.5 dark:border-gray-700 dark:bg-gray-800">
			{#each periods as period}
				<button
					onclick={() => switchPeriod(period.value)}
					class="rounded-md px-3 py-1.5 text-xs font-medium transition-colors
						{selectedPeriod === period.value
							? 'bg-white text-gray-900 shadow-sm dark:bg-gray-700 dark:text-white'
							: 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'}"
				>
					{t(period.label)}
				</button>
			{/each}
		</div>
	</div>

	<!-- Performance Metrics Cards -->
	{#if performance}
	<div class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
		<Card>
			<div class="flex items-center gap-3">
				<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-secondary-100 dark:bg-secondary-900/30">
					<MessageSquare class="h-5 w-5 text-secondary-600" />
				</div>
				<div>
					<p class="text-xs font-medium text-gray-500 dark:text-gray-400">{t('analytics.response_rate')}</p>
					<p class="text-xl font-bold text-gray-900 dark:text-white">{formatPercent(performance.response_rate)}</p>
				</div>
			</div>
		</Card>
		<Card>
			<div class="flex items-center gap-3">
				<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-primary-100 dark:bg-primary-900/30">
					<CheckCircle class="h-5 w-5 text-primary-600" />
				</div>
				<div>
					<p class="text-xs font-medium text-gray-500 dark:text-gray-400">{t('analytics.completion_rate')}</p>
					<p class="text-xl font-bold text-gray-900 dark:text-white">{formatPercent(performance.completion_rate)}</p>
				</div>
			</div>
		</Card>
		<Card>
			<div class="flex items-center gap-3">
				<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-yellow-100 dark:bg-yellow-900/30">
					<Star class="h-5 w-5 text-yellow-500" />
				</div>
				<div>
					<p class="text-xs font-medium text-gray-500 dark:text-gray-400">{t('analytics.avg_rating')}</p>
					<p class="text-xl font-bold text-gray-900 dark:text-white">{performance.avg_rating.toFixed(1)}</p>
				</div>
			</div>
		</Card>
		<Card>
			<div class="flex items-center gap-3">
				<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100 dark:bg-blue-900/30">
					<Users class="h-5 w-5 text-blue-600" />
				</div>
				<div>
					<p class="text-xs font-medium text-gray-500 dark:text-gray-400">{t('analytics.total_reviews')}</p>
					<p class="text-xl font-bold text-gray-900 dark:text-white">{performance.total_reviews}</p>
				</div>
			</div>
		</Card>
		<Card>
			<div class="flex items-center gap-3">
				<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-secondary-100 dark:bg-secondary-900/30">
					<IndianRupee class="h-5 w-5 text-secondary-600" />
				</div>
				<div>
					<p class="text-xs font-medium text-gray-500 dark:text-gray-400">{t('analytics.total_earnings')}</p>
					<p class="text-xl font-bold text-gray-900 dark:text-white">Rs. {performance.total_earnings.toLocaleString()}</p>
				</div>
			</div>
		</Card>
	</div>
	{/if}

	<div class="mt-8 grid gap-6 lg:grid-cols-2">
		<!-- Earnings Chart -->
		<Card>
			<div class="flex items-center gap-2">
				<TrendingUp class="h-5 w-5 text-secondary-600" />
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">{t('analytics.earnings_chart')}</h2>
			</div>
			<div class="mt-4">
				{#if earningsHistory.length > 1}
					<svg viewBox="0 0 600 240" class="w-full" preserveAspectRatio="xMidYMid meet">
						<!-- Grid lines -->
						{#each [0, 0.25, 0.5, 0.75, 1] as ratio}
							<line
								x1="20" y1={200 - ratio * 160}
								x2="580" y2={200 - ratio * 160}
								stroke="currentColor" class="text-gray-200 dark:text-gray-700"
								stroke-width="0.5" stroke-dasharray="4,4"
							/>
						{/each}

						<!-- Area fill -->
						<path d={getEarningsChartArea()} fill="url(#earningsGradient)" opacity="0.3" />

						<!-- Line -->
						<path d={getEarningsChartPath()} fill="none" stroke="#16a34a" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" />

						<!-- Data points -->
						{#each getEarningsChartLabels() as point, i}
							{@const maxE = Math.max(...earningsHistory.map(e => e.earnings), 1)}
							{@const y = 200 - (earningsHistory[i].earnings / maxE) * 160}
							<circle cx={point.x} cy={y} r="4" fill="#16a34a" stroke="white" stroke-width="2" />
						{/each}

						<!-- X-axis labels -->
						{#each getEarningsChartLabels() as point}
							<text x={point.x} y="225" text-anchor="middle" class="fill-gray-500 dark:fill-gray-400" font-size="10">{point.label}</text>
						{/each}

						<!-- Gradient definition -->
						<defs>
							<linearGradient id="earningsGradient" x1="0" y1="0" x2="0" y2="1">
								<stop offset="0%" stop-color="#16a34a" stop-opacity="0.4" />
								<stop offset="100%" stop-color="#16a34a" stop-opacity="0.05" />
							</linearGradient>
						</defs>
					</svg>
				{:else}
					<div class="flex h-48 items-center justify-center">
						<p class="text-sm text-gray-500 dark:text-gray-400">{t('analytics.no_earnings')}</p>
					</div>
				{/if}
			</div>
		</Card>

		<!-- Demand Heatmap -->
		<Card>
			<div class="flex items-center gap-2">
				<MapPin class="h-5 w-5 text-red-500" />
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">{t('analytics.demand_heatmap')}</h2>
			</div>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{t('analytics.demand_heatmap_desc')}</p>
			<div class="mt-4 h-64 overflow-hidden rounded-lg border border-gray-200 dark:border-gray-700">
				<div bind:this={mapContainer} class="h-full w-full" style="z-index: 0;"></div>
			</div>
			<!-- Legend -->
			<div class="mt-3 flex items-center justify-center gap-4 text-xs text-gray-500 dark:text-gray-400">
				<div class="flex items-center gap-1.5">
					<span class="h-3 w-3 rounded-full bg-green-500"></span>
					{t('analytics.low')}
				</div>
				<div class="flex items-center gap-1.5">
					<span class="h-3 w-3 rounded-full bg-yellow-500"></span>
					{t('analytics.medium')}
				</div>
				<div class="flex items-center gap-1.5">
					<span class="h-3 w-3 rounded-full bg-red-500"></span>
					{t('analytics.high')}
				</div>
			</div>
		</Card>

		<!-- Peak Hours Chart -->
		<Card>
			<div class="flex items-center gap-2">
				<Clock class="h-5 w-5 text-blue-600" />
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">{t('analytics.peak_hours')}</h2>
			</div>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{t('analytics.peak_hours_desc')}</p>
			<div class="mt-4">
				{#if peakHours.length > 0}
					<svg viewBox="0 0 600 220" class="w-full" preserveAspectRatio="xMidYMid meet">
						{#each getPeakHoursBars() as bar}
							<rect
								x={bar.x}
								y={180 - bar.height}
								width={bar.width}
								height={bar.height}
								rx="2"
								class="fill-primary-500 hover:fill-primary-600 transition-colors"
							>
								<title>{bar.hour} - {bar.count} requests</title>
							</rect>
						{/each}

						<!-- X-axis labels (show every 3 hours) -->
						{#each getPeakHoursBars() as bar, i}
							{#if i % 3 === 0}
								<text
									x={bar.x + bar.width / 2}
									y="200"
									text-anchor="middle"
									class="fill-gray-500 dark:fill-gray-400"
									font-size="9"
								>{bar.hour}</text>
							{/if}
						{/each}
					</svg>
				{:else}
					<div class="flex h-40 items-center justify-center">
						<p class="text-sm text-gray-500 dark:text-gray-400">No demand data available</p>
					</div>
				{/if}
			</div>
		</Card>

		<!-- Demand by Category -->
		<Card>
			<div class="flex items-center gap-2">
				<BarChart3 class="h-5 w-5 text-primary-600" />
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">{t('analytics.demand_by_category')}</h2>
			</div>
			<div class="mt-4 space-y-3">
				{#if demandCategories.length > 0}
					{@const maxCatDemand = Math.max(...demandCategories.map(c => c.demand_count), 1)}
					{#each demandCategories.slice(0, 8) as cat}
						<div class="flex items-center gap-3">
							<span class="w-28 truncate text-sm text-gray-700 dark:text-gray-300">{cat.category_name}</span>
							<div class="flex-1">
								<div class="h-5 rounded-full bg-gray-100 dark:bg-gray-700">
									<div
										class="h-5 rounded-full bg-primary-500 flex items-center justify-end pr-2"
										style="width: {Math.max((cat.demand_count / maxCatDemand) * 100, 8)}%"
									>
										<span class="text-[10px] font-semibold text-white">{cat.demand_count}</span>
									</div>
								</div>
							</div>
						</div>
					{/each}
				{:else}
					<p class="text-sm text-gray-500 dark:text-gray-400">No demand data available</p>
				{/if}
			</div>
		</Card>
	</div>

	<!-- Competitor Density Table -->
	<div class="mt-6">
		<Card>
			<div class="flex items-center gap-2">
				<Users class="h-5 w-5 text-orange-500" />
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">{t('analytics.competitors')}</h2>
			</div>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{t('analytics.competitors_desc')}</p>
			<div class="mt-4 overflow-x-auto">
				{#if competitors.length > 0}
					<table class="w-full text-sm">
						<thead>
							<tr class="border-b border-gray-200 dark:border-gray-700">
								<th class="py-2 pr-4 text-left font-medium text-gray-500 dark:text-gray-400">{t('analytics.postcode')}</th>
								<th class="py-2 pr-4 text-left font-medium text-gray-500 dark:text-gray-400">{t('analytics.category')}</th>
								<th class="py-2 text-right font-medium text-gray-500 dark:text-gray-400">{t('analytics.provider_count')}</th>
							</tr>
						</thead>
						<tbody>
							{#each competitors.slice(0, 15) as comp}
								<tr class="border-b border-gray-100 dark:border-gray-800">
									<td class="py-2 pr-4 font-medium text-gray-900 dark:text-white">{comp.postcode}</td>
									<td class="py-2 pr-4 text-gray-600 dark:text-gray-400">{comp.category_name}</td>
									<td class="py-2 text-right">
										<span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium
											{comp.provider_count <= 3
												? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
												: comp.provider_count <= 7
													? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
													: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'}">
											{comp.provider_count}
										</span>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				{:else}
					<p class="py-4 text-center text-sm text-gray-500 dark:text-gray-400">No competitor data available</p>
				{/if}
			</div>
		</Card>
	</div>

	<!-- AI Insights -->
	<div class="mt-6">
		<Card>
			<div class="flex items-center gap-2">
				<Lightbulb class="h-5 w-5 text-yellow-500" />
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">{t('analytics.insights')}</h2>
			</div>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{t('analytics.insights_desc')}</p>
			<div class="mt-4 space-y-3">
				{#if insights.length > 0}
					{#each insights as insight}
						<div class="flex gap-3 rounded-lg border border-gray-100 p-4 dark:border-gray-700">
							<div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg {getInsightIcon(insight.type)}">
								<Lightbulb class="h-5 w-5" />
							</div>
							<div class="flex-1 min-w-0">
								<div class="flex items-center gap-2">
									<h3 class="text-sm font-semibold text-gray-900 dark:text-white">{insight.title}</h3>
									<span class="rounded-full px-2 py-0.5 text-[10px] font-medium {getImpactBadgeClass(insight.impact)}">
										{insight.impact}
									</span>
								</div>
								<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{insight.message}</p>
							</div>
						</div>
					{/each}
				{:else}
					<p class="py-4 text-center text-sm text-gray-500 dark:text-gray-400">{t('analytics.no_insights')}</p>
				{/if}
			</div>
		</Card>
	</div>
</div>
{/if}

<style>
	:global(.seva-map-marker) {
		background: transparent !important;
		border: none !important;
	}
</style>
