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

	// Redirect unauthenticated users to login, non-providers to homepage
	$effect(() => {
		if (!authState.initialized || authState.loading) return;
		if (!authState.user) {
			goto(`/login?redirect=${encodeURIComponent($page.url.pathname)}`);
		} else if (authState.user.role !== 'provider' && authState.user.role !== 'admin') {
			goto('/');
		}
	});
</script>

{#if authState.loading && !authState.initialized}
	<div class="flex min-h-[60vh] items-center justify-center">
		<div class="h-8 w-8 animate-spin rounded-full border-4 border-primary-600 border-t-transparent"></div>
	</div>
{:else if authState.user && (authState.user.role === 'provider' || authState.user.role === 'admin')}
	{@render children()}
{:else if authState.initialized}
	<div class="flex min-h-[60vh] items-center justify-center">
		<p class="text-gray-500">Redirecting...</p>
	</div>
{/if}
