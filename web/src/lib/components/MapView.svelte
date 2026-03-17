<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type L from 'leaflet';

	interface MapMarker {
		id: string;
		lat: number;
		lng: number;
		title: string;
		category?: string;
		type: 'job' | 'provider';
		link?: string;
	}

	let {
		markers = [],
		center = { lat: 12.9716, lng: 77.5946 },
		zoom = 12,
		onmarkerclick
	}: {
		markers?: MapMarker[];
		center?: { lat: number; lng: number };
		zoom?: number;
		onmarkerclick?: (marker: MapMarker) => void;
	} = $props();

	let mapContainer: HTMLDivElement;
	let map: L.Map | null = null;
	let markerLayer: L.LayerGroup | null = null;
	let leaflet: typeof L;

	// Track the currently highlighted marker
	let highlightedMarkerId = $state<string | null>(null);
	let markerInstances = new Map<string, L.Marker>();

	export function highlightMarker(id: string) {
		highlightedMarkerId = id;
		const instance = markerInstances.get(id);
		if (instance && map) {
			map.setView(instance.getLatLng(), Math.max(map.getZoom(), 15));
			instance.openPopup();
		}
	}

	export function fitAllMarkers() {
		if (!map || !markerLayer || markers.length === 0) return;
		const bounds = leaflet.latLngBounds(markers.map((m) => [m.lat, m.lng]));
		map.fitBounds(bounds, { padding: [40, 40] });
	}

	function createIcon(type: 'job' | 'provider', isHighlighted: boolean): L.DivIcon {
		const color = type === 'job' ? '#3b82f6' : '#22c55e';
		const size = isHighlighted ? 36 : 28;
		const borderWidth = isHighlighted ? 3 : 2;

		return leaflet.divIcon({
			className: 'seva-map-marker',
			html: `<div style="
				width: ${size}px;
				height: ${size}px;
				background: ${color};
				border: ${borderWidth}px solid white;
				border-radius: 50% 50% 50% 0;
				transform: rotate(-45deg);
				box-shadow: 0 2px 6px rgba(0,0,0,0.3);
				${isHighlighted ? 'filter: brightness(1.2); z-index: 1000;' : ''}
			"></div>`,
			iconSize: [size, size],
			iconAnchor: [size / 2, size],
			popupAnchor: [0, -size]
		});
	}

	function renderMarkers() {
		if (!map || !leaflet || !markerLayer) return;

		markerLayer.clearLayers();
		markerInstances.clear();

		for (const m of markers) {
			const isHighlighted = highlightedMarkerId === m.id;
			const icon = createIcon(m.type, isHighlighted);

			const marker = leaflet.marker([m.lat, m.lng], { icon });

			const popupContent = `
				<div style="min-width: 160px; font-family: system-ui, sans-serif;">
					<p style="margin: 0 0 4px 0; font-weight: 600; font-size: 14px; color: #111827;">${m.title}</p>
					${m.category ? `<p style="margin: 0 0 6px 0; font-size: 12px; color: #6b7280;">${m.category}</p>` : ''}
					<span style="
						display: inline-block;
						padding: 1px 8px;
						border-radius: 9999px;
						font-size: 11px;
						font-weight: 500;
						background: ${m.type === 'job' ? '#dbeafe' : '#dcfce7'};
						color: ${m.type === 'job' ? '#1d4ed8' : '#15803d'};
					">${m.type === 'job' ? 'Job' : 'Provider'}</span>
					${m.link ? `<br><a href="${m.link}" style="display: inline-block; margin-top: 6px; font-size: 12px; color: #4f46e5; text-decoration: none; font-weight: 500;">View Details &rarr;</a>` : ''}
				</div>
			`;

			marker.bindPopup(popupContent);

			marker.on('click', () => {
				highlightedMarkerId = m.id;
				if (onmarkerclick) onmarkerclick(m);
			});

			markerLayer.addLayer(marker);
			markerInstances.set(m.id, marker);
		}

		// Auto-fit bounds when markers change
		if (markers.length > 0) {
			const bounds = leaflet.latLngBounds(markers.map((m) => [m.lat, m.lng]));
			map.fitBounds(bounds, { padding: [40, 40], maxZoom: 15 });
		}
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

		map = leaflet.map(mapContainer).setView([center.lat, center.lng], zoom);

		leaflet
			.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
				attribution:
					'&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
				maxZoom: 19
			})
			.addTo(map);

		markerLayer = leaflet.layerGroup().addTo(map);

		renderMarkers();

		// Invalidate size after a tick so the map fills its container correctly
		setTimeout(() => map?.invalidateSize(), 100);
	});

	$effect(() => {
		// Re-render markers whenever `markers` or `highlightedMarkerId` changes
		if (map && leaflet && markerLayer) {
			// Read the dependencies to trigger reactivity
			void markers;
			void highlightedMarkerId;
			renderMarkers();
		}
	});

	$effect(() => {
		if (map) {
			map.setView([center.lat, center.lng], zoom);
		}
	});

	onDestroy(() => {
		if (map) {
			map.remove();
			map = null;
		}
	});
</script>

<div
	bind:this={mapContainer}
	class="h-full w-full min-h-[300px]"
	style="z-index: 0;"
></div>

<style>
	:global(.seva-map-marker) {
		background: transparent !important;
		border: none !important;
	}

	:global(.leaflet-popup-content-wrapper) {
		border-radius: 12px;
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
	}

	:global(.leaflet-popup-tip) {
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
	}
</style>
