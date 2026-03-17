<script lang="ts">
	import { ArrowLeft, BookOpen, Clock, Eye, Tag, User, ChevronRight } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import { t } from '$lib/i18n/index.svelte';

	let { data } = $props();

	let article = $derived(data.article);
	let relatedArticles = $derived(data.relatedArticles || []);
	let error = $derived(data.error);

	// Simple markdown-to-HTML converter for article body.
	// Handles headings, bold, italic, lists, links, checkboxes, tables, and paragraphs.
	function renderMarkdown(md: string): string {
		if (!md) return '';

		return md
			.split('\n\n')
			.map((block) => {
				// Headings
				if (block.startsWith('## ')) {
					return `<h2 class="mt-8 mb-3 text-xl font-semibold text-gray-900 dark:text-white">${block.slice(3)}</h2>`;
				}
				if (block.startsWith('### ')) {
					return `<h3 class="mt-6 mb-2 text-lg font-medium text-gray-900 dark:text-white">${block.slice(4)}</h3>`;
				}

				// Table (pipe-delimited)
				if (block.includes('|') && block.split('\n').length >= 3) {
					const rows = block.split('\n').filter((r) => r.trim() && !r.match(/^\|[\s-|]+\|$/));
					if (rows.length >= 2) {
						const headerCells = rows[0].split('|').filter((c) => c.trim());
						const bodyRows = rows.slice(1);

						let table = '<div class="overflow-x-auto my-4"><table class="min-w-full border border-gray-200 dark:border-gray-700 rounded-lg text-sm">';
						table += '<thead class="bg-gray-50 dark:bg-gray-800"><tr>';
						for (const cell of headerCells) {
							table += `<th class="px-4 py-2 text-left font-medium text-gray-700 dark:text-gray-300 border-b border-gray-200 dark:border-gray-700">${cell.trim()}</th>`;
						}
						table += '</tr></thead><tbody>';
						for (const row of bodyRows) {
							const cells = row.split('|').filter((c) => c.trim());
							table += '<tr class="border-b border-gray-100 dark:border-gray-700">';
							for (const cell of cells) {
								table += `<td class="px-4 py-2 text-gray-600 dark:text-gray-400">${cell.trim()}</td>`;
							}
							table += '</tr>';
						}
						table += '</tbody></table></div>';
						return table;
					}
				}

				// Unordered list or checkbox list
				if (block.match(/^[-*] /m)) {
					const items = block.split('\n').map((line) => {
						const trimmed = line.replace(/^[-*] /, '');
						// Checkbox
						if (trimmed.startsWith('[ ] ')) {
							return `<li class="flex items-start gap-2"><input type="checkbox" disabled class="mt-1 rounded border-gray-300 dark:border-gray-600" /><span>${formatInline(trimmed.slice(4))}</span></li>`;
						}
						if (trimmed.startsWith('[x] ')) {
							return `<li class="flex items-start gap-2"><input type="checkbox" checked disabled class="mt-1 rounded border-gray-300 dark:border-gray-600" /><span class="line-through text-gray-500">${formatInline(trimmed.slice(4))}</span></li>`;
						}
						return `<li>${formatInline(trimmed)}</li>`;
					});
					return `<ul class="my-3 space-y-1.5 pl-5 list-disc text-sm text-gray-600 dark:text-gray-400">${items.join('')}</ul>`;
				}

				// Ordered list
				if (block.match(/^\d+\. /m)) {
					const items = block.split('\n').map((line) => {
						const text = line.replace(/^\d+\. /, '');
						return `<li>${formatInline(text)}</li>`;
					});
					return `<ol class="my-3 space-y-1.5 pl-5 list-decimal text-sm text-gray-600 dark:text-gray-400">${items.join('')}</ol>`;
				}

				// Regular paragraph
				return `<p class="my-3 text-sm leading-relaxed text-gray-600 dark:text-gray-400">${formatInline(block)}</p>`;
			})
			.join('');
	}

	// Format inline markdown (bold, italic, links, code)
	function formatInline(text: string): string {
		return text
			.replace(/\*\*(.+?)\*\*/g, '<strong class="font-semibold text-gray-900 dark:text-white">$1</strong>')
			.replace(/\*(.+?)\*/g, '<em>$1</em>')
			.replace(/`(.+?)`/g, '<code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-mono dark:bg-gray-700">$1</code>')
			.replace(/\[(.+?)\]\((.+?)\)/g, '<a href="$2" class="text-primary-600 hover:text-primary-700 underline dark:text-primary-400">$1</a>');
	}

	function getCategoryLabel(category: string): string {
		const labels: Record<string, string> = {
			provider_guide: 'Provider Guide',
			customer_tip: 'Customer Tips',
			maintenance: 'Maintenance',
			pricing: 'Pricing',
			legal: 'Legal'
		};
		return labels[category] || category;
	}

	function getAudienceLabel(audience: string): string {
		if (audience === 'provider') return 'For Providers';
		if (audience === 'customer') return 'For Customers';
		return 'For Everyone';
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '';
		return new Date(dateStr).toLocaleDateString('en-IN', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}
</script>

<svelte:head>
	{#if article}
		<title>{article.title} - Seva Help Center</title>
		<meta name="description" content={article.summary} />
		<meta property="og:title" content={article.title} />
		<meta property="og:description" content={article.summary} />
		<meta property="og:type" content="article" />
	{:else}
		<title>Article Not Found - Seva Help Center</title>
	{/if}
</svelte:head>

<div class="mx-auto max-w-6xl px-4 py-8 sm:px-6 lg:px-8">
	<!-- Back link -->
	<a href="/help" class="inline-flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300">
		<ArrowLeft class="h-4 w-4" />
		Back to Help Center
	</a>

	{#if error || !article}
		<div class="mt-6 rounded-xl border border-red-200 bg-red-50 p-8 text-center dark:border-red-800 dark:bg-red-900/20">
			<BookOpen class="mx-auto h-12 w-12 text-red-400" />
			<p class="mt-3 text-red-600 dark:text-red-400">{error || 'Article not found'}</p>
			<a href="/help" class="mt-4 inline-block rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white hover:bg-primary-700">
				Browse all articles
			</a>
		</div>
	{:else}
		<div class="mt-6 grid gap-8 lg:grid-cols-3">
			<!-- Article Content -->
			<div class="lg:col-span-2">
				<article>
					<!-- Header -->
					<header>
						<div class="flex flex-wrap items-center gap-2">
							<Badge variant={article.audience === 'provider' ? 'info' : article.audience === 'customer' ? 'success' : 'neutral'} size="sm">
								{getAudienceLabel(article.audience)}
							</Badge>
							<Badge variant="neutral" size="sm">
								{getCategoryLabel(article.category)}
							</Badge>
						</div>

						<h1 class="mt-3 text-2xl font-bold text-gray-900 dark:text-white sm:text-3xl">
							{article.title}
						</h1>

						{#if article.summary}
							<p class="mt-3 text-base text-gray-600 dark:text-gray-400 leading-relaxed">
								{article.summary}
							</p>
						{/if}

						<div class="mt-4 flex flex-wrap items-center gap-4 text-sm text-gray-500 dark:text-gray-400">
							<span class="flex items-center gap-1">
								<User class="h-4 w-4" />
								{article.author_name}
							</span>
							<span class="flex items-center gap-1">
								<Clock class="h-4 w-4" />
								{formatDate(article.created_at)}
							</span>
							<span class="flex items-center gap-1">
								<Eye class="h-4 w-4" />
								{article.view_count} views
							</span>
						</div>

						{#if article.tags && article.tags.length > 0}
							<div class="mt-3 flex flex-wrap items-center gap-2">
								<Tag class="h-4 w-4 text-gray-400" />
								{#each article.tags as tag}
									<span class="rounded-full bg-gray-100 px-2.5 py-0.5 text-xs text-gray-600 dark:bg-gray-700 dark:text-gray-400">
										{tag}
									</span>
								{/each}
							</div>
						{/if}
					</header>

					<!-- Divider -->
					<hr class="my-6 border-gray-200 dark:border-gray-700" />

					<!-- Body -->
					<div class="prose-seva">
						{@html renderMarkdown(article.body)}
					</div>
				</article>
			</div>

			<!-- Sidebar -->
			<aside class="space-y-6">
				<!-- Related Articles -->
				{#if relatedArticles.length > 0}
					<Card>
						<h2 class="font-semibold text-gray-900 dark:text-white">Related Articles</h2>
						<div class="mt-4 space-y-3">
							{#each relatedArticles as related}
								<a
									href="/help/{related.slug}"
									class="group flex items-start gap-2 rounded-lg p-2 -mx-2 transition-colors hover:bg-gray-50 dark:hover:bg-gray-700/50"
								>
									<ChevronRight class="h-4 w-4 flex-shrink-0 mt-0.5 text-gray-400 group-hover:text-primary-500" />
									<div class="flex-1 min-w-0">
										<p class="text-sm font-medium text-gray-900 group-hover:text-primary-600 dark:text-white dark:group-hover:text-primary-400 line-clamp-2">
											{related.title}
										</p>
										<p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400 line-clamp-1">
											{related.summary}
										</p>
									</div>
								</a>
							{/each}
						</div>
					</Card>
				{/if}

				<!-- Article Info -->
				<Card>
					<h2 class="font-semibold text-gray-900 dark:text-white">About this article</h2>
					<div class="mt-4 space-y-3">
						<div>
							<p class="text-xs text-gray-500 dark:text-gray-400">Category</p>
							<p class="mt-0.5 text-sm text-gray-900 dark:text-white">{getCategoryLabel(article.category)}</p>
						</div>
						<div>
							<p class="text-xs text-gray-500 dark:text-gray-400">Audience</p>
							<p class="mt-0.5 text-sm text-gray-900 dark:text-white">{getAudienceLabel(article.audience)}</p>
						</div>
						<div>
							<p class="text-xs text-gray-500 dark:text-gray-400">Author</p>
							<p class="mt-0.5 text-sm text-gray-900 dark:text-white">{article.author_name}</p>
						</div>
						<div>
							<p class="text-xs text-gray-500 dark:text-gray-400">Last updated</p>
							<p class="mt-0.5 text-sm text-gray-900 dark:text-white">{formatDate(article.updated_at)}</p>
						</div>
					</div>
				</Card>

				<!-- Back to Help -->
				<Card>
					<h2 class="font-semibold text-gray-900 dark:text-white">Need more help?</h2>
					<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
						Browse our full collection of guides and tips.
					</p>
					<div class="mt-4 space-y-2">
						<a
							href="/help"
							class="block rounded-lg bg-gray-100 px-4 py-2 text-center text-sm font-medium text-gray-700 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
						>
							View all articles
						</a>
						<a
							href="mailto:support@seva.app"
							class="block rounded-lg bg-primary-600 px-4 py-2 text-center text-sm font-medium text-white hover:bg-primary-700"
						>
							Contact Support
						</a>
					</div>
				</Card>
			</aside>
		</div>
	{/if}
</div>
