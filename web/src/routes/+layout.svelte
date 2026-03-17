<script lang="ts">
	import '../app.css';
	import { Bell, Menu, X, ChevronDown, User, LayoutDashboard, Settings, LogOut, Briefcase, Search, Award } from 'lucide-svelte';
	import { subscribe as authSubscribe, logout, type AuthState } from '$lib/stores/auth';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import Toast from '$lib/components/ui/Toast.svelte';

	let { children } = $props();

	let authState = $state<AuthState>({
		user: null,
		loading: false,
		initialized: false,
		notificationCount: 0,
		pointsBalance: 0
	});

	let mobileMenuOpen = $state(false);
	let userDropdownOpen = $state(false);

	$effect(() => {
		const unsub = authSubscribe((state) => {
			authState = state;
		});
		return unsub;
	});

	const isLoggedIn = $derived(authState.user !== null);
	const isProviderUser = $derived(authState.user?.role === 'provider');
	const isAdminUser = $derived(authState.user?.role === 'admin');

	async function handleLogout() {
		userDropdownOpen = false;
		await logout();
		window.location.href = '/';
	}

	function closeDropdown(e: MouseEvent) {
		const target = e.target as HTMLElement;
		if (!target.closest('[data-dropdown]')) {
			userDropdownOpen = false;
		}
	}
</script>

<svelte:window onclick={closeDropdown} />

<div class="min-h-screen flex flex-col bg-white dark:bg-gray-950">
	<header class="sticky top-0 z-40 border-b border-gray-200 bg-white/95 backdrop-blur dark:border-gray-800 dark:bg-gray-950/95">
		<nav class="mx-auto flex max-w-7xl items-center justify-between px-4 py-3 sm:px-6 lg:px-8">
			<!-- Logo -->
			<a href="/" class="flex items-center gap-2">
				<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary-600 text-white font-bold text-sm">S</div>
				<span class="text-xl font-bold text-gray-900 dark:text-white">Seva</span>
			</a>

			<!-- Desktop Nav -->
			<div class="hidden items-center gap-6 md:flex">
				<a href="/providers" class="text-sm text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-100">
					Find Providers
				</a>
				<a href="/jobs" class="text-sm text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-100">
					Jobs
				</a>
				{#if isLoggedIn && isProviderUser}
					<a href="/provider/dashboard" class="text-sm text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-100">
						Provider Hub
					</a>
				{/if}
				{#if isLoggedIn && isAdminUser}
					<a href="/admin" class="text-sm text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-100">
						Admin
					</a>
				{/if}
			</div>

			<!-- Desktop Right -->
			<div class="hidden items-center gap-3 md:flex">
				{#if isLoggedIn}
					<!-- Notification Bell -->
					<a href="/notifications" class="relative rounded-lg p-2 text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800">
						<Bell class="h-5 w-5" />
						{#if authState.notificationCount > 0}
							<span class="absolute right-1 top-1 flex h-4 min-w-[1rem] items-center justify-center rounded-full bg-red-500 px-1 text-[10px] font-bold text-white">
								{authState.notificationCount > 99 ? '99+' : authState.notificationCount}
							</span>
						{/if}
					</a>

					<!-- User Dropdown -->
					<div class="relative" data-dropdown>
						<button
							onclick={() => (userDropdownOpen = !userDropdownOpen)}
							class="flex items-center gap-2 rounded-lg p-1.5 hover:bg-gray-100 dark:hover:bg-gray-800"
						>
							<Avatar
								src={authState.user?.avatar_url}
								name={authState.user?.name || ''}
								size="sm"
							/>
							<span class="text-sm font-medium text-gray-700 dark:text-gray-300">
								{authState.user?.name?.split(' ')[0] || 'Account'}
							</span>
							<ChevronDown class="h-4 w-4 text-gray-400" />
						</button>

						{#if userDropdownOpen}
							<div class="absolute right-0 mt-2 w-56 rounded-xl border border-gray-200 bg-white py-1 shadow-lg dark:border-gray-700 dark:bg-gray-800">
								<div class="border-b border-gray-100 px-4 py-3 dark:border-gray-700">
									<p class="text-sm font-medium text-gray-900 dark:text-white">{authState.user?.name}</p>
									<p class="text-xs text-gray-500 dark:text-gray-400">{authState.user?.phone}</p>
									{#if authState.pointsBalance > 0}
										<div class="mt-1 flex items-center gap-1 text-xs text-primary-600 dark:text-primary-400">
											<Award class="h-3 w-3" />
											<span>{authState.pointsBalance} points</span>
										</div>
									{/if}
								</div>
								<a href="/dashboard" class="flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 dark:text-gray-300 dark:hover:bg-gray-700" onclick={() => (userDropdownOpen = false)}>
									<LayoutDashboard class="h-4 w-4" />
									Dashboard
								</a>
								<a href="/profile" class="flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 dark:text-gray-300 dark:hover:bg-gray-700" onclick={() => (userDropdownOpen = false)}>
									<User class="h-4 w-4" />
									Profile
								</a>
								<a href="/points" class="flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 dark:text-gray-300 dark:hover:bg-gray-700" onclick={() => (userDropdownOpen = false)}>
									<Award class="h-4 w-4" />
									Points & Rewards
								</a>
								<div class="border-t border-gray-100 dark:border-gray-700">
									<button
										onclick={handleLogout}
										class="flex w-full items-center gap-3 px-4 py-2.5 text-sm text-red-600 hover:bg-gray-50 dark:text-red-400 dark:hover:bg-gray-700"
									>
										<LogOut class="h-4 w-4" />
										Sign Out
									</button>
								</div>
							</div>
						{/if}
					</div>
				{:else}
					<a
						href="/login"
						class="rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white hover:bg-primary-700"
					>
						Sign In
					</a>
				{/if}
			</div>

			<!-- Mobile Menu Button -->
			<button
				class="rounded-lg p-2 text-gray-500 hover:bg-gray-100 md:hidden dark:text-gray-400 dark:hover:bg-gray-800"
				onclick={() => (mobileMenuOpen = !mobileMenuOpen)}
				aria-label="Toggle menu"
			>
				{#if mobileMenuOpen}
					<X class="h-6 w-6" />
				{:else}
					<Menu class="h-6 w-6" />
				{/if}
			</button>
		</nav>

		<!-- Mobile Menu -->
		{#if mobileMenuOpen}
			<div class="border-t border-gray-200 bg-white pb-4 md:hidden dark:border-gray-700 dark:bg-gray-950">
				{#if isLoggedIn}
					<div class="border-b border-gray-100 px-4 py-3 dark:border-gray-700">
						<div class="flex items-center gap-3">
							<Avatar
								src={authState.user?.avatar_url}
								name={authState.user?.name || ''}
								size="md"
							/>
							<div>
								<p class="font-medium text-gray-900 dark:text-white">{authState.user?.name}</p>
								<p class="text-xs text-gray-500 dark:text-gray-400">{authState.user?.phone}</p>
							</div>
						</div>
					</div>
				{/if}
				<div class="space-y-1 px-4 pt-3">
					<a href="/providers" class="block rounded-lg px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800" onclick={() => (mobileMenuOpen = false)}>
						Find Providers
					</a>
					<a href="/jobs" class="block rounded-lg px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800" onclick={() => (mobileMenuOpen = false)}>
						Jobs
					</a>
					{#if isLoggedIn}
						<a href="/dashboard" class="block rounded-lg px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800" onclick={() => (mobileMenuOpen = false)}>
							Dashboard
						</a>
						<a href="/notifications" class="flex items-center justify-between rounded-lg px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800" onclick={() => (mobileMenuOpen = false)}>
							Notifications
							{#if authState.notificationCount > 0}
								<span class="rounded-full bg-red-500 px-2 py-0.5 text-[10px] font-bold text-white">{authState.notificationCount}</span>
							{/if}
						</a>
						<a href="/profile" class="block rounded-lg px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800" onclick={() => (mobileMenuOpen = false)}>
							Profile
						</a>
						<a href="/points" class="block rounded-lg px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800" onclick={() => (mobileMenuOpen = false)}>
							Points & Rewards
						</a>
						{#if isProviderUser}
							<a href="/provider/dashboard" class="block rounded-lg px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800" onclick={() => (mobileMenuOpen = false)}>
								Provider Hub
							</a>
						{/if}
						{#if isAdminUser}
							<a href="/admin" class="block rounded-lg px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800" onclick={() => (mobileMenuOpen = false)}>
								Admin Panel
							</a>
						{/if}
						<div class="border-t border-gray-100 pt-2 dark:border-gray-700">
							<button
								onclick={handleLogout}
								class="block w-full rounded-lg px-3 py-2 text-left text-sm text-red-600 hover:bg-gray-100 dark:text-red-400 dark:hover:bg-gray-800"
							>
								Sign Out
							</button>
						</div>
					{:else}
						<div class="border-t border-gray-100 pt-3 dark:border-gray-700">
							<a
								href="/login"
								class="block rounded-lg bg-primary-600 px-4 py-2.5 text-center text-sm font-medium text-white hover:bg-primary-700"
								onclick={() => (mobileMenuOpen = false)}
							>
								Sign In
							</a>
						</div>
					{/if}
				</div>
			</div>
		{/if}
	</header>

	<main class="flex-1">
		{@render children()}
	</main>

	<footer class="border-t border-gray-200 bg-gray-50 dark:border-gray-800 dark:bg-gray-900">
		<div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
			<div class="grid gap-8 sm:grid-cols-2 lg:grid-cols-4">
				<div>
					<div class="flex items-center gap-2">
						<div class="flex h-7 w-7 items-center justify-center rounded-lg bg-primary-600 text-white font-bold text-xs">S</div>
						<span class="text-lg font-bold text-gray-900 dark:text-white">Seva</span>
					</div>
					<p class="mt-3 text-sm text-gray-500 dark:text-gray-400">
						Connecting trusted service providers with customers across India.
					</p>
				</div>
				<div>
					<h3 class="text-sm font-semibold text-gray-900 dark:text-white">For Customers</h3>
					<ul class="mt-3 space-y-2 text-sm text-gray-500 dark:text-gray-400">
						<li><a href="/providers" class="hover:text-primary-600">Find Providers</a></li>
						<li><a href="/jobs/new" class="hover:text-primary-600">Post a Job</a></li>
						<li><a href="/register" class="hover:text-primary-600">Create Account</a></li>
					</ul>
				</div>
				<div>
					<h3 class="text-sm font-semibold text-gray-900 dark:text-white">For Providers</h3>
					<ul class="mt-3 space-y-2 text-sm text-gray-500 dark:text-gray-400">
						<li><a href="/register?role=provider" class="hover:text-primary-600">Join as Provider</a></li>
						<li><a href="/provider/dashboard" class="hover:text-primary-600">Provider Dashboard</a></li>
						<li><a href="/provider/earnings" class="hover:text-primary-600">Earnings</a></li>
					</ul>
				</div>
				<div>
					<h3 class="text-sm font-semibold text-gray-900 dark:text-white">Support</h3>
					<ul class="mt-3 space-y-2 text-sm text-gray-500 dark:text-gray-400">
						<li><a href="#" class="hover:text-primary-600">Help Center</a></li>
						<li><a href="#" class="hover:text-primary-600">Privacy Policy</a></li>
						<li><a href="#" class="hover:text-primary-600">Terms of Service</a></li>
					</ul>
				</div>
			</div>
			<div class="mt-8 border-t border-gray-200 pt-8 dark:border-gray-700">
				<p class="text-center text-sm text-gray-500 dark:text-gray-400">
					&copy; 2026 Seva. All rights reserved.
				</p>
			</div>
		</div>
	</footer>
</div>

<!-- Toast container -->
<Toast />
