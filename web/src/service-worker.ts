/// <reference types="@sveltejs/kit" />
/// <reference no-default-lib="true"/>
/// <reference lib="esnext" />
/// <reference lib="webworker" />

const sw = self as unknown as ServiceWorkerGlobalScope;

import { build, files, version } from '$service-worker';

const CACHE_NAME = `seva-cache-${version}`;

const ASSETS = [...build, ...files];

// Pages to pre-cache for offline use
const OFFLINE_PAGES = ['/', '/login', '/register'];

sw.addEventListener('install', (event) => {
	event.waitUntil(
		caches
			.open(CACHE_NAME)
			.then((cache) => cache.addAll(ASSETS))
			.then(() => sw.skipWaiting())
	);
});

sw.addEventListener('activate', (event) => {
	event.waitUntil(
		caches
			.keys()
			.then((keys) =>
				Promise.all(
					keys
						.filter((key) => key !== CACHE_NAME)
						.map((key) => caches.delete(key))
				)
			)
			.then(() => sw.clients.claim())
	);
});

sw.addEventListener('fetch', (event) => {
	const url = new URL(event.request.url);

	// Skip non-GET requests
	if (event.request.method !== 'GET') return;

	// Skip API requests - let them go to network
	if (url.pathname.startsWith('/api/')) return;

	// Skip Chrome extension requests
	if (url.protocol === 'chrome-extension:') return;

	// For build assets and static files, use cache-first strategy
	if (ASSETS.includes(url.pathname)) {
		event.respondWith(
			caches.match(event.request).then((cached) => {
				return cached || fetch(event.request);
			})
		);
		return;
	}

	// For navigation requests, use network-first with offline fallback
	if (event.request.mode === 'navigate') {
		event.respondWith(
			fetch(event.request)
				.then((response) => {
					// Cache successful navigation responses
					const responseClone = response.clone();
					caches.open(CACHE_NAME).then((cache) => {
						cache.put(event.request, responseClone);
					});
					return response;
				})
				.catch(() => {
					// Try to serve from cache when offline
					return caches.match(event.request).then((cached) => {
						if (cached) return cached;
						// Fallback to cached home page
						return caches.match('/') as Promise<Response>;
					});
				})
		);
		return;
	}

	// For other requests (images, fonts, etc.), use stale-while-revalidate
	event.respondWith(
		caches.match(event.request).then((cached) => {
			const fetchPromise = fetch(event.request)
				.then((response) => {
					// Only cache successful responses
					if (response.ok) {
						const responseClone = response.clone();
						caches.open(CACHE_NAME).then((cache) => {
							cache.put(event.request, responseClone);
						});
					}
					return response;
				})
				.catch(() => {
					// If fetch fails and no cache, return a fallback
					if (cached) return cached;
					return new Response('Offline', { status: 503 });
				});

			return cached || fetchPromise;
		})
	);
});

// Handle push notifications
sw.addEventListener('push', (event) => {
	if (!event.data) return;

	const data = event.data.json();
	const options: NotificationOptions = {
		body: data.body || 'You have a new notification',
		icon: '/favicon.png',
		badge: '/favicon.png',
		data: {
			url: data.url || '/'
		},
		tag: data.tag || 'seva-notification',
		requireInteraction: data.requireInteraction || false
	};

	event.waitUntil(sw.registration.showNotification(data.title || 'Seva', options));
});

// Handle notification clicks
sw.addEventListener('notificationclick', (event) => {
	event.notification.close();
	const url = event.notification.data?.url || '/';
	event.waitUntil(
		sw.clients.matchAll({ type: 'window' }).then((clients) => {
			// Focus existing window if available
			for (const client of clients) {
				if (client.url === url && 'focus' in client) {
					return client.focus();
				}
			}
			// Otherwise open new window
			return sw.clients.openWindow(url);
		})
	);
});
