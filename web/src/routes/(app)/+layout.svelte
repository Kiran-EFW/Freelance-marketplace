<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { subscribe as authSubscribe, type AuthState } from '$lib/stores/auth';

	let { children } = $props();

	let authState = $state<AuthState>({
		user: null,
		loading: false,
		initialized: false,
		notificationCount: 0,
		pointsBalance: 0
	});

	$effect(() => {
		const unsub = authSubscribe((state) => {
			authState = state;
		});
		return unsub;
	});

	// Pages within (app) that are publicly accessible without login
	const publicPrefixes = ['/providers', '/map', '/jobs', '/seasonal-calendar'];

	function isPublicPath(pathname: string): boolean {
		// /jobs is public for browsing but /jobs/new requires auth
		if (pathname === '/jobs/new') return false;
		return publicPrefixes.some((prefix) => pathname === prefix || pathname.startsWith(prefix + '/'));
	}

	// Redirect to login if accessing a protected page while not authenticated
	$effect(() => {
		if (!authState.initialized || authState.loading) return;
		if (!authState.user && !isPublicPath($page.url.pathname)) {
			goto(`/login?redirect=${encodeURIComponent($page.url.pathname)}`);
		}
	});
</script>

{#if authState.loading && !authState.initialized}
	<div class="flex min-h-[60vh] items-center justify-center">
		<div class="h-8 w-8 animate-spin rounded-full border-4 border-primary-600 border-t-transparent"></div>
	</div>
{:else if authState.user || isPublicPath($page.url.pathname)}
	{@render children()}
{:else if authState.initialized}
	<!-- Redirecting to login... -->
	<div class="flex min-h-[60vh] items-center justify-center">
		<p class="text-gray-500">Redirecting to sign in...</p>
	</div>
{/if}
