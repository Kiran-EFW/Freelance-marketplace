<script lang="ts">
	import '../../app.css';
	import { goto } from '$app/navigation';
	import { subscribe as authSubscribe, logout, initAuth, type AuthState } from '$lib/stores/auth';
	import { initLocale, t } from '$lib/i18n/index.svelte';
	import Toast from '$lib/components/ui/Toast.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import { LogOut, ArrowLeft } from 'lucide-svelte';

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

	$effect(() => {
		if (typeof window !== 'undefined') {
			initLocale();
			initAuth();
		}
	});

	// Redirect non-admin users away
	$effect(() => {
		if (!authState.initialized || authState.loading) return;
		if (!authState.user) {
			goto('/login?redirect=/admin');
		} else if (authState.user.role !== 'admin') {
			goto('/');
		}
	});

	async function handleLogout() {
		await logout();
		window.location.href = '/';
	}

	const isAdmin = $derived(authState.user?.role === 'admin');
</script>

{#if authState.loading && !authState.initialized}
	<div class="flex min-h-screen items-center justify-center bg-gray-100 dark:bg-gray-950">
		<div class="h-8 w-8 animate-spin rounded-full border-4 border-primary-600 border-t-transparent"></div>
	</div>
{:else if isAdmin}
	<div class="min-h-screen flex flex-col bg-gray-100 dark:bg-gray-950">
		<!-- Admin-only header — completely separate from consumer UI -->
		<header class="sticky top-0 z-40 border-b border-gray-300 bg-gray-900 text-white dark:border-gray-700">
			<nav class="flex items-center justify-between px-4 py-2.5 sm:px-6">
				<div class="flex items-center gap-4">
					<a href="/" class="flex items-center gap-1 text-xs text-gray-400 hover:text-white" title="Back to main site">
						<ArrowLeft class="h-3.5 w-3.5" />
						<span class="hidden sm:inline">Main Site</span>
					</a>
					<div class="h-4 w-px bg-gray-600"></div>
					<div class="flex items-center gap-2">
						<div class="flex h-7 w-7 items-center justify-center rounded bg-primary-600 text-white font-bold text-xs">S</div>
						<span class="font-semibold text-sm">Seva Admin</span>
					</div>
				</div>

				<div class="flex items-center gap-3">
					<div class="flex items-center gap-2">
						<Avatar
							src={authState.user?.avatar_url}
							name={authState.user?.name || ''}
							size="sm"
						/>
						<span class="text-sm text-gray-300 hidden sm:inline">
							{authState.user?.name || 'Admin'}
						</span>
					</div>
					<button
						onclick={handleLogout}
						class="flex items-center gap-1.5 rounded px-2.5 py-1.5 text-xs text-gray-400 hover:bg-gray-800 hover:text-white"
					>
						<LogOut class="h-3.5 w-3.5" />
						<span class="hidden sm:inline">{t('common.sign_out')}</span>
					</button>
				</div>
			</nav>
		</header>

		<main class="flex-1">
			{@render children()}
		</main>

		<footer class="border-t border-gray-300 bg-gray-900 px-4 py-3 dark:border-gray-700">
			<p class="text-center text-xs text-gray-500">
				Seva Admin Panel &mdash; Internal use only
			</p>
		</footer>
	</div>
{:else if authState.initialized}
	<div class="flex min-h-screen items-center justify-center bg-gray-100 dark:bg-gray-950">
		<p class="text-gray-500">Redirecting...</p>
	</div>
{/if}

<Toast />
