<script lang="ts">
	import { onMount } from 'svelte';
	import { Search, BookOpen, Wrench, DollarSign, Shield, ChevronRight, TrendingUp, Loader2, Eye } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import api from '$lib/api/client';
	import { t } from '$lib/i18n/index.svelte';

	// State
	let loading = $state(true);
	let articles = $state<any[]>([]);
	let popularArticles = $state<any[]>([]);
	let searchQuery = $state('');
	let selectedCategory = $state('');
	let selectedAudience = $state('');
	let total = $state(0);

	// Category configuration
	const categoryConfig: Record<string, { label: string; icon: typeof BookOpen; variant: 'success' | 'warning' | 'danger' | 'info' | 'neutral' }> = {
		provider_guide: { label: 'Provider Guide', icon: TrendingUp, variant: 'info' },
		customer_tip: { label: 'Customer Tips', icon: BookOpen, variant: 'success' },
		maintenance: { label: 'Maintenance', icon: Wrench, variant: 'warning' },
		pricing: { label: 'Pricing', icon: DollarSign, variant: 'neutral' },
		legal: { label: 'Legal', icon: Shield, variant: 'danger' }
	};

	// Computed: filtered articles based on search
	let filteredArticles = $derived(
		searchQuery.trim()
			? articles.filter(
					(a) =>
						a.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
						a.summary.toLowerCase().includes(searchQuery.toLowerCase()) ||
						(a.tags || []).some((tag: string) => tag.toLowerCase().includes(searchQuery.toLowerCase()))
				)
			: articles
	);

	// Group articles by category
	let groupedArticles = $derived.by(() => {
		const groups: Record<string, any[]> = {};
		for (const article of filteredArticles) {
			const cat = article.category || 'general';
			if (!groups[cat]) groups[cat] = [];
			groups[cat].push(article);
		}
		return groups;
	});

	onMount(async () => {
		try {
			const [articlesRes, popularRes] = await Promise.all([
				api.content.list({
					audience: selectedAudience || undefined,
					category: selectedCategory || undefined,
					per_page: 50
				}),
				api.content.popular({ limit: 5 })
			]);

			articles = articlesRes.data || [];
			total = (articlesRes as any).meta?.total || articles.length;
			popularArticles = popularRes.data || [];
		} catch (err) {
			console.error('Failed to load articles:', err);
		} finally {
			loading = false;
		}
	});

	async function fetchArticles() {
		loading = true;
		try {
			const res = await api.content.list({
				audience: selectedAudience || undefined,
				category: selectedCategory || undefined,
				per_page: 50
			});
			articles = res.data || [];
			total = (res as any).meta?.total || articles.length;
		} catch (err) {
			console.error('Failed to load articles:', err);
		} finally {
			loading = false;
		}
	}

	function handleCategoryChange(category: string) {
		selectedCategory = selectedCategory === category ? '' : category;
		fetchArticles();
	}

	function handleAudienceChange(audience: string) {
		selectedAudience = selectedAudience === audience ? '' : audience;
		fetchArticles();
	}

	function getCategoryConfig(category: string) {
		return categoryConfig[category] || { label: category, icon: BookOpen, variant: 'neutral' as const };
	}

	function formatDate(dateStr: string) {
		if (!dateStr) return '';
		return new Date(dateStr).toLocaleDateString('en-IN', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}
</script>

<svelte:head>
	<title>{t('common.help') ?? 'Help Center'} - Seva</title>
	<meta name="description" content="Guides, tips, and educational content for Seva customers and service providers. Learn how to grow your business, prepare for services, and more." />
</svelte:head>

<div class="mx-auto max-w-6xl px-4 py-8 sm:px-6 lg:px-8">
	<!-- Header -->
	<div class="text-center">
		<h1 class="text-3xl font-bold text-gray-900 dark:text-white sm:text-4xl">
			{t('common.help') ?? 'Help Center'}
		</h1>
		<p class="mt-3 text-lg text-gray-600 dark:text-gray-400">
			Guides, tips, and resources to help you get the most out of Seva
		</p>
	</div>

	<!-- Search -->
	<div class="mx-auto mt-8 max-w-xl">
		<div class="relative">
			<Search class="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-gray-400" />
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Search articles..."
				class="w-full rounded-xl border border-gray-300 bg-white py-3 pl-10 pr-4 text-sm shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-800 dark:text-white"
			/>
		</div>
	</div>

	<!-- Audience Filter -->
	<div class="mt-6 flex flex-wrap justify-center gap-2">
		<button
			onclick={() => handleAudienceChange('')}
			class="rounded-full px-4 py-1.5 text-sm font-medium transition-colors {selectedAudience === '' ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600'}"
		>
			All
		</button>
		<button
			onclick={() => handleAudienceChange('customer')}
			class="rounded-full px-4 py-1.5 text-sm font-medium transition-colors {selectedAudience === 'customer' ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600'}"
		>
			For Customers
		</button>
		<button
			onclick={() => handleAudienceChange('provider')}
			class="rounded-full px-4 py-1.5 text-sm font-medium transition-colors {selectedAudience === 'provider' ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600'}"
		>
			For Providers
		</button>
	</div>

	<!-- Category Filter -->
	<div class="mt-4 flex flex-wrap justify-center gap-2">
		{#each Object.entries(categoryConfig) as [key, config]}
			<button
				onclick={() => handleCategoryChange(key)}
				class="rounded-full px-3 py-1 text-xs font-medium transition-colors {selectedCategory === key ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-400 dark:hover:bg-gray-600'}"
			>
				{config.label}
			</button>
		{/each}
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<Loader2 class="h-8 w-8 animate-spin text-primary-600" />
		</div>
	{:else}
		<div class="mt-10 grid gap-8 lg:grid-cols-3">
			<!-- Main Content -->
			<div class="space-y-8 lg:col-span-2">
				{#if filteredArticles.length === 0}
					<div class="rounded-xl border border-gray-200 bg-gray-50 p-8 text-center dark:border-gray-700 dark:bg-gray-800/50">
						<BookOpen class="mx-auto h-12 w-12 text-gray-400" />
						<p class="mt-3 text-gray-600 dark:text-gray-400">No articles found. Try a different search or filter.</p>
					</div>
				{:else}
					{#each Object.entries(groupedArticles) as [category, categoryArticles]}
						{@const config = getCategoryConfig(category)}
						<section>
							<div class="flex items-center gap-2 mb-4">
								<svelte:component this={config.icon} class="h-5 w-5 text-gray-500 dark:text-gray-400" />
								<h2 class="text-lg font-semibold text-gray-900 dark:text-white">{config.label}</h2>
								<Badge variant={config.variant} size="sm">{categoryArticles.length}</Badge>
							</div>

							<div class="space-y-3">
								{#each categoryArticles as article}
									<a
										href="/help/{article.slug}"
										class="group block rounded-xl border border-gray-200 bg-white p-4 transition-all hover:border-primary-300 hover:shadow-md dark:border-gray-700 dark:bg-gray-800 dark:hover:border-primary-600"
									>
										<div class="flex items-start justify-between gap-4">
											<div class="flex-1 min-w-0">
												<h3 class="font-medium text-gray-900 group-hover:text-primary-600 dark:text-white dark:group-hover:text-primary-400">
													{article.title}
												</h3>
												<p class="mt-1 text-sm text-gray-500 dark:text-gray-400 line-clamp-2">
													{article.summary}
												</p>
												<div class="mt-2 flex flex-wrap items-center gap-2">
													<Badge variant={article.audience === 'provider' ? 'info' : article.audience === 'customer' ? 'success' : 'neutral'} size="sm">
														{article.audience === 'provider' ? 'Provider' : article.audience === 'customer' ? 'Customer' : 'Everyone'}
													</Badge>
													{#each (article.tags || []).slice(0, 3) as tag}
														<span class="text-xs text-gray-400 dark:text-gray-500">#{tag}</span>
													{/each}
													{#if article.view_count > 0}
														<span class="flex items-center gap-1 text-xs text-gray-400 dark:text-gray-500">
															<Eye class="h-3 w-3" />
															{article.view_count}
														</span>
													{/if}
												</div>
											</div>
											<ChevronRight class="h-5 w-5 flex-shrink-0 text-gray-400 group-hover:text-primary-500 dark:text-gray-500" />
										</div>
									</a>
								{/each}
							</div>
						</section>
					{/each}
				{/if}
			</div>

			<!-- Sidebar -->
			<div class="space-y-6">
				<!-- Popular Articles -->
				{#if popularArticles.length > 0}
					<Card>
						<h2 class="flex items-center gap-2 font-semibold text-gray-900 dark:text-white">
							<TrendingUp class="h-5 w-5 text-primary-500" />
							Popular Articles
						</h2>
						<div class="mt-4 space-y-3">
							{#each popularArticles as article, i}
								<a
									href="/help/{article.slug}"
									class="group flex items-start gap-3 rounded-lg p-2 -mx-2 transition-colors hover:bg-gray-50 dark:hover:bg-gray-700/50"
								>
									<span class="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-primary-100 text-xs font-bold text-primary-700 dark:bg-primary-900/30 dark:text-primary-400">
										{i + 1}
									</span>
									<div class="flex-1 min-w-0">
										<p class="text-sm font-medium text-gray-900 group-hover:text-primary-600 dark:text-white dark:group-hover:text-primary-400 line-clamp-2">
											{article.title}
										</p>
										<p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400 flex items-center gap-1">
											<Eye class="h-3 w-3" />
											{article.view_count} views
										</p>
									</div>
								</a>
							{/each}
						</div>
					</Card>
				{/if}

				<!-- Quick Links -->
				<Card>
					<h2 class="font-semibold text-gray-900 dark:text-white">Quick Links</h2>
					<div class="mt-4 space-y-2">
						<a href="/help/how-to-grow-your-service-business-on-seva" class="block text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300">
							How to grow your business
						</a>
						<a href="/help/how-to-choose-the-right-service-provider" class="block text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300">
							Choosing the right provider
						</a>
						<a href="/help/understanding-sevas-escrow-payment-protection" class="block text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300">
							Payment protection
						</a>
						<a href="/help/how-sevas-trust-score-works" class="block text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300">
							Trust scores explained
						</a>
						<a href="/help/safety-tips-when-booking-home-services" class="block text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300">
							Safety tips
						</a>
					</div>
				</Card>

				<!-- Contact Support -->
				<Card>
					<h2 class="font-semibold text-gray-900 dark:text-white">Need more help?</h2>
					<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
						Cannot find what you are looking for? Our support team is here to help.
					</p>
					<div class="mt-4 space-y-2">
						<a
							href="mailto:support@seva.app"
							class="block rounded-lg bg-primary-600 px-4 py-2 text-center text-sm font-medium text-white hover:bg-primary-700"
						>
							Contact Support
						</a>
					</div>
				</Card>
			</div>
		</div>
	{/if}
</div>
