<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type L from 'leaflet';

	export interface RouteMapStop {
		id: string;
		lat: number;
		lng: number;
		address: string;
		customerName: string;
		order: number;
	}

	let {
		stops = [],
		optimizedStops = [],
		editable = false,
		showOptimized = false,
		onoptimize,
		onstopclick
	}: {
		stops?: RouteMapStop[];
		optimizedStops?: RouteMapStop[];
		editable?: boolean;
		showOptimized?: boolean;
		onoptimize?: () => void;
		onstopclick?: (stop: RouteMapStop) => void;
	} = $props();

	let mapContainer: HTMLDivElement;
	let map: L.Map | null = null;
	let leaflet: typeof L;
	let markerLayer: L.LayerGroup | null = null;
	let originalPolyline: L.Polyline | null = null;
	let optimizedPolyline: L.Polyline | null = null;

	function createNumberedIcon(num: number, isOptimized: boolean): L.DivIcon {
		const bg = isOptimized ? '#16a34a' : '#3b82f6';
		const border = isOptimized ? '#15803d' : '#2563eb';
		return leaflet.divIcon({
			className: 'seva-route-marker',
			html: `<div style="
				width: 30px;
				height: 30px;
				background: ${bg};
				border: 2px solid ${border};
				border-radius: 50%;
				display: flex;
				align-items: center;
				justify-content: center;
				color: white;
				font-weight: 700;
				font-size: 13px;
				font-family: system-ui, sans-serif;
				box-shadow: 0 2px 6px rgba(0,0,0,0.3);
			">${num}</div>`,
			iconSize: [30, 30],
			iconAnchor: [15, 15],
			popupAnchor: [0, -18]
		});
	}

	function renderMap() {
		if (!map || !leaflet || !markerLayer) return;

		markerLayer.clearLayers();
		if (originalPolyline) {
			map.removeLayer(originalPolyline);
			originalPolyline = null;
		}
		if (optimizedPolyline) {
			map.removeLayer(optimizedPolyline);
			optimizedPolyline = null;
		}

		const displayStops = showOptimized && optimizedStops.length > 0 ? optimizedStops : stops;

		if (displayStops.length === 0) return;

		// Draw original route polyline (gray) when showing optimized comparison
		if (showOptimized && optimizedStops.length > 0 && stops.length > 1) {
			const originalLatLngs = stops.map((s) => leaflet.latLng(s.lat, s.lng));
			originalPolyline = leaflet.polyline(originalLatLngs, {
				color: '#9ca3af',
				weight: 3,
				opacity: 0.5,
				dashArray: '8, 8'
			}).addTo(map);
		}

		// Draw the active route polyline
		if (displayStops.length > 1) {
			const latLngs = displayStops.map((s) => leaflet.latLng(s.lat, s.lng));
			const color = showOptimized && optimizedStops.length > 0 ? '#16a34a' : '#3b82f6';
			optimizedPolyline = leaflet.polyline(latLngs, {
				color,
				weight: 4,
				opacity: 0.8
			}).addTo(map);
		}

		// Add numbered markers
		const isOpt = showOptimized && optimizedStops.length > 0;
		for (const stop of displayStops) {
			const icon = createNumberedIcon(stop.order, isOpt);
			const marker = leaflet.marker([stop.lat, stop.lng], { icon });

			const popupContent = `
				<div style="min-width: 180px; font-family: system-ui, sans-serif;">
					<p style="margin: 0 0 2px 0; font-weight: 600; font-size: 14px; color: #111827;">
						Stop ${stop.order}: ${stop.customerName}
					</p>
					<p style="margin: 0; font-size: 12px; color: #6b7280;">${stop.address}</p>
				</div>
			`;
			marker.bindPopup(popupContent);

			marker.on('click', () => {
				if (onstopclick) onstopclick(stop);
			});

			markerLayer.addLayer(marker);
		}

		// Fit bounds
		const bounds = leaflet.latLngBounds(displayStops.map((s) => [s.lat, s.lng]));
		map.fitBounds(bounds, { padding: [50, 50], maxZoom: 16 });
	}

	onMount(async () => {
		leaflet = (await import('leaflet')).default;

		// Import Leaflet CSS
		const linkEl = document.createElement('link');
		linkEl.rel = 'stylesheet';
		linkEl.href = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.css';
		linkEl.integrity = 'sha256-p4NxAoJBhIIN+hmNHrzRCf9tD/miZyoHS5obTRR9BMY=';
		linkEl.crossOrigin = '';
		document.head.appendChild(linkEl);

		const defaultCenter: [number, number] = [12.9716, 77.5946]; // Bangalore
		map = leaflet.map(mapContainer).setView(defaultCenter, 13);

		leaflet
			.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
				attribution:
					'&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
				maxZoom: 19
			})
			.addTo(map);

		markerLayer = leaflet.layerGroup().addTo(map);

		renderMap();

		setTimeout(() => map?.invalidateSize(), 100);
	});

	$effect(() => {
		if (map && leaflet && markerLayer) {
			// Track dependencies for reactivity
			void stops;
			void optimizedStops;
			void showOptimized;
			renderMap();
		}
	});

	onDestroy(() => {
		if (map) {
			map.remove();
			map = null;
		}
	});
</script>

<div class="relative">
	<div
		bind:this={mapContainer}
		class="h-full w-full min-h-[350px] rounded-xl"
		style="z-index: 0;"
	></div>

	{#if showOptimized && optimizedStops.length > 0}
		<div class="absolute bottom-3 left-3 z-10 flex items-center gap-4 rounded-lg bg-white/90 px-3 py-2 text-xs shadow-md backdrop-blur dark:bg-gray-800/90 dark:text-gray-200">
			<span class="flex items-center gap-1.5">
				<span class="inline-block h-2.5 w-5 rounded-sm" style="background: #9ca3af;"></span>
				Original
			</span>
			<span class="flex items-center gap-1.5">
				<span class="inline-block h-2.5 w-5 rounded-sm" style="background: #16a34a;"></span>
				Optimized
			</span>
		</div>
	{/if}

	{#if editable && onoptimize}
		<button
			onclick={onoptimize}
			class="absolute right-3 top-3 z-10 rounded-lg bg-primary-600 px-3 py-1.5 text-sm font-medium text-white shadow-md hover:bg-primary-700 transition-colors"
		>
			Optimize Route
		</button>
	{/if}
</div>

<style>
	:global(.seva-route-marker) {
		background: transparent !important;
		border: none !important;
	}
</style>
