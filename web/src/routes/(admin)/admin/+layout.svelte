<script lang="ts">
	import { page } from '$app/stores';
	import {
		LayoutDashboard,
		Users,
		Briefcase,
		Shield,
		BarChart3,
		AlertTriangle,
		MessageSquare,
		Menu,
		X
	} from 'lucide-svelte';

	let { children } = $props();
	let mobileMenuOpen = $state(false);

	const navItems = [
		{ href: '/admin', label: 'Dashboard', icon: LayoutDashboard },
		{ href: '/admin/users', label: 'Users', icon: Users },
		{ href: '/admin/kyc', label: 'KYC Verification', icon: Shield },
		{ href: '/admin/disputes', label: 'Disputes', icon: AlertTriangle },
		{ href: '/admin/analytics', label: 'Analytics', icon: BarChart3 },
		{ href: '/admin/sms-ivr', label: 'SMS/IVR', icon: MessageSquare }
	];

	function isActive(href: string): boolean {
		if (href === '/admin') return $page.url.pathname === '/admin';
		return $page.url.pathname.startsWith(href);
	}
</script>

<div class="flex min-h-[calc(100vh-theme(spacing.32))]">
	<!-- Mobile menu toggle -->
	<div class="fixed bottom-4 right-4 z-50 lg:hidden">
		<button
			onclick={() => (mobileMenuOpen = !mobileMenuOpen)}
			class="flex h-12 w-12 items-center justify-center rounded-full bg-primary-600 text-white shadow-lg"
		>
			{#if mobileMenuOpen}
				<X class="h-5 w-5" />
			{:else}
				<Menu class="h-5 w-5" />
			{/if}
		</button>
	</div>

	<!-- Mobile sidebar overlay -->
	{#if mobileMenuOpen}
		<div class="fixed inset-0 z-40 lg:hidden">
			<button onclick={() => (mobileMenuOpen = false)} aria-label="Close sidebar menu" class="absolute inset-0 bg-black/50"></button>
			<aside class="absolute left-0 top-0 h-full w-64 bg-gray-50 dark:bg-gray-900">
				<div class="p-6">
					<h2 class="text-sm font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
						Admin Panel
					</h2>
				</div>
				<nav class="space-y-1 px-3">
					{#each navItems as item}
						{@const Icon = item.icon}
						{@const active = isActive(item.href)}
						<a
							href={item.href}
							onclick={() => (mobileMenuOpen = false)}
							class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition
								{active
									? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400'
									: 'text-gray-700 hover:bg-gray-200 dark:text-gray-300 dark:hover:bg-gray-800'}"
						>
							<Icon class="h-5 w-5" />
							{item.label}
						</a>
					{/each}
				</nav>
			</aside>
		</div>
	{/if}

	<!-- Desktop Sidebar -->
	<aside class="hidden w-64 shrink-0 border-r border-gray-200 bg-gray-50 dark:border-gray-700 dark:bg-gray-900 lg:block">
		<div class="p-6">
			<h2 class="text-sm font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
				Admin Panel
			</h2>
		</div>
		<nav class="space-y-1 px-3">
			{#each navItems as item}
				{@const Icon = item.icon}
				{@const active = isActive(item.href)}
				<a
					href={item.href}
					class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition
						{active
							? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400'
							: 'text-gray-700 hover:bg-gray-200 dark:text-gray-300 dark:hover:bg-gray-800'}"
				>
					<Icon class="h-5 w-5" />
					{item.label}
				</a>
			{/each}
		</nav>
	</aside>

	<!-- Main Content -->
	<div class="flex-1">
		{@render children()}
	</div>
</div>
