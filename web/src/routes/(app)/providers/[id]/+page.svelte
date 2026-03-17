<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { ArrowLeft, Star, MapPin, Clock, Shield, Phone, MessageSquare, Briefcase, Calendar, CheckCircle, IndianRupee, Loader2 } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import StarRating from '$lib/components/ui/StarRating.svelte';
	import api from '$lib/api/client';
	import { t } from '$lib/i18n/index.svelte';

	let providerId = $derived($page.params.id);
	let loading = $state(true);
	let error = $state('');

	let provider = $state<any>(null);
	let providerUserId = $state<string>('');
	let sendingMessage = $state(false);
	let reviews = $state<any[]>([]);
	let completedPhotos = $state<any[]>([]);
	let ratingDist = $state<Record<number, number>>({ 5: 0, 4: 0, 3: 0, 2: 0, 1: 0 });

	onMount(async () => {
		try {
			const [providerRes, reviewsRes] = await Promise.all([
				api.providers.get(providerId),
				api.reviews.listForProvider(providerId, { per_page: 10 }).catch(() => ({ data: [] }))
			]);

			const p = providerRes.data;
			providerUserId = p.user_id || p.id;
			provider = {
				id: p.id,
				name: p.user?.name || p.business_name || 'Provider',
				bio: p.bio || '',
				rating: p.rating_average || 0,
				reviewCount: p.rating_count || 0,
				completedJobs: p.completed_jobs_count || p.rating_count || 0,
				memberSince: p.created_at?.split('T')[0] || '',
				responseTime: p.response_time_minutes ? `${p.response_time_minutes} minutes` : 'N/A',
				hourlyRate: p.hourly_rate || 0,
				isVerified: p.verification_status === 'approved',
				isOnline: p.is_online ?? false,
				trustScore: p.trust_score || 0,
				skills: (p.categories || []).map((c: any) => c.name || c),
				serviceArea: (p.service_areas || []).join(', ') || 'N/A',
				languages: p.languages || [],
				workingHours: p.working_hours || 'N/A'
			};

			completedPhotos = (p.portfolio_images || []).map((img: any, i: number) => ({
				id: String(i),
				title: img.title || `Work ${i + 1}`,
				description: img.description || '',
				url: img.url || img
			}));

			reviews = (reviewsRes.data || []).map((r: any) => ({
				id: r.id,
				customerName: r.reviewer?.name || r.customer?.name || 'Customer',
				rating: r.rating,
				comment: r.comment || '',
				date: r.created_at?.split('T')[0] || '',
				jobTitle: r.job?.title || r.job_title || ''
			}));

			// Build rating distribution from reviews
			const dist: Record<number, number> = { 5: 0, 4: 0, 3: 0, 2: 0, 1: 0 };
			reviews.forEach((r: any) => {
				if (r.rating >= 1 && r.rating <= 5) dist[r.rating]++;
			});
			ratingDist = dist;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load provider profile';
		} finally {
			loading = false;
		}
	});

	async function handleSendMessage() {
		if (!providerUserId || sendingMessage) return;
		sendingMessage = true;
		try {
			const response = await api.messages.createConversation(providerUserId);
			const conversationId = response.data.id;
			goto(`/messages/${conversationId}`);
		} catch (err) {
			console.error('Failed to create conversation:', err);
			sendingMessage = false;
		}
	}
</script>

<svelte:head>
	<title>{provider ? `${provider.name} - Seva` : 'Provider - Seva'}</title>
	{#if provider}
		<meta name="description" content="{provider.bio.substring(0, 160)}" />
	{/if}
</svelte:head>

{#if loading}
<div class="flex items-center justify-center py-20">
	<Loader2 class="h-8 w-8 animate-spin text-primary-600" />
</div>
{:else if error || !provider}
<div class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
	<a href="/providers" class="inline-flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700">
		<ArrowLeft class="h-4 w-4" />
		Back to Providers
	</a>
	<div class="mt-6 rounded-lg border border-red-200 bg-red-50 p-6 text-center dark:border-red-800 dark:bg-red-900/20">
		<p class="text-red-600 dark:text-red-400">{error || 'Provider not found'}</p>
	</div>
</div>
{:else}

<div class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
	<a href="/providers" class="inline-flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700">
		<ArrowLeft class="h-4 w-4" />
		Back to Providers
	</a>

	<!-- Provider Profile Card -->
	<Card class="mt-6">
		<div class="flex flex-col gap-6 sm:flex-row">
			<div class="relative">
				<Avatar name={provider.name} size="lg" />
				{#if provider.isOnline}
					<div class="absolute -bottom-1 -right-1 h-4 w-4 rounded-full border-2 border-white bg-green-500 dark:border-gray-800"></div>
				{/if}
			</div>

			<div class="flex-1">
				<div class="flex items-center gap-2">
					<h1 class="text-2xl font-bold text-gray-900 dark:text-white">{provider.name}</h1>
					{#if provider.isVerified}
						<Shield class="h-5 w-5 text-secondary-500" />
					{/if}
					{#if provider.isOnline}
						<Badge variant="success" size="sm">Online</Badge>
					{/if}
				</div>

				<div class="mt-2 flex items-center gap-1">
					<StarRating rating={provider.rating} size="sm" />
					<span class="text-sm font-medium text-gray-900 dark:text-white">{provider.rating}</span>
					<span class="text-sm text-gray-500 dark:text-gray-400">({provider.reviewCount} reviews)</span>
				</div>

				<div class="mt-3 flex flex-wrap gap-4 text-sm text-gray-600 dark:text-gray-400">
					<span class="flex items-center gap-1">
						<MapPin class="h-4 w-4 text-gray-400" />
						{provider.serviceArea.split(',')[0]}
					</span>
					<span class="flex items-center gap-1">
						<Clock class="h-4 w-4 text-gray-400" />
						Responds in {provider.responseTime}
					</span>
					<span class="flex items-center gap-1">
						<CheckCircle class="h-4 w-4 text-gray-400" />
						{provider.completedJobs} jobs completed
					</span>
					<span class="flex items-center gap-1">
						<IndianRupee class="h-4 w-4 text-gray-400" />
						Rs. {provider.hourlyRate}/hour
					</span>
				</div>
			</div>

			<div class="flex flex-col gap-2 sm:items-end">
				<Button variant="primary" href="/jobs/new">
					<Briefcase class="h-4 w-4" />
					Request Quote
				</Button>
				<Button variant="outline">
					<Phone class="h-4 w-4" />
					Contact
				</Button>
			</div>
		</div>
	</Card>

	<div class="mt-8 grid gap-6 lg:grid-cols-3">
		<!-- Main Content -->
		<div class="space-y-6 lg:col-span-2">
			<!-- About -->
			<Card>
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">About</h2>
				<p class="mt-3 text-sm text-gray-600 dark:text-gray-400 leading-relaxed">{provider.bio}</p>
			</Card>

			<!-- Skills -->
			<Card>
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Skills & Services</h2>
				<div class="mt-3 flex flex-wrap gap-2">
					{#each provider.skills as skill}
						<Badge variant="info" size="sm">{skill}</Badge>
					{/each}
				</div>
			</Card>

			<!-- Portfolio -->
			<Card>
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Past Work</h2>
				<div class="mt-4 grid grid-cols-2 gap-3">
					{#each completedPhotos as photo}
						<div class="rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-gray-700 dark:bg-gray-800">
							<div class="flex h-24 items-center justify-center rounded bg-gray-200 dark:bg-gray-700">
								<Briefcase class="h-8 w-8 text-gray-400" />
							</div>
							<p class="mt-2 text-sm font-medium text-gray-900 dark:text-white">{photo.title}</p>
							<p class="text-xs text-gray-500 dark:text-gray-400">{photo.description}</p>
						</div>
					{/each}
				</div>
			</Card>

			<!-- Reviews -->
			<Card>
				<div class="flex items-center justify-between">
					<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Reviews ({provider.reviewCount})</h2>
				</div>

				<!-- Rating Summary -->
				<div class="mt-4 flex items-center gap-6 rounded-lg bg-gray-50 p-4 dark:bg-gray-800/50">
					<div class="text-center">
						<p class="text-4xl font-bold text-gray-900 dark:text-white">{provider.rating}</p>
						<StarRating rating={provider.rating} size="sm" />
						<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{provider.reviewCount} reviews</p>
					</div>
					<div class="flex-1 space-y-1.5">
						{#each [5, 4, 3, 2, 1] as stars}
							{@const count = ratingDist[stars as keyof typeof ratingDist] || 0}
							{@const pct = (count / provider.reviewCount) * 100}
							<div class="flex items-center gap-2 text-sm">
								<span class="w-3 text-gray-500">{stars}</span>
								<div class="flex-1 h-2 rounded-full bg-gray-200 dark:bg-gray-700">
									<div class="h-2 rounded-full bg-yellow-400" style="width: {pct}%"></div>
								</div>
								<span class="w-8 text-right text-xs text-gray-400">{count}</span>
							</div>
						{/each}
					</div>
				</div>

				<!-- Individual Reviews -->
				<div class="mt-4 space-y-4">
					{#each reviews as review}
						<div class="border-t border-gray-100 pt-4 first:border-0 first:pt-0 dark:border-gray-700">
							<div class="flex items-start gap-3">
								<Avatar name={review.customerName} size="sm" />
								<div class="flex-1">
									<div class="flex items-center gap-2">
										<span class="text-sm font-medium text-gray-900 dark:text-white">{review.customerName}</span>
										<StarRating rating={review.rating} size="sm" />
									</div>
									<p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{review.jobTitle} -- {review.date}</p>
									<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">{review.comment}</p>
								</div>
							</div>
						</div>
					{/each}
				</div>
			</Card>
		</div>

		<!-- Sidebar -->
		<div class="space-y-6">
			<!-- Trust Score -->
			<Card>
				<div class="flex items-center gap-2">
					<Shield class="h-5 w-5 text-secondary-500" />
					<h2 class="font-semibold text-gray-900 dark:text-white">Trust Score</h2>
				</div>
				<div class="mt-4 flex items-center justify-center">
					<div class="relative flex h-24 w-24 items-center justify-center rounded-full border-8 border-secondary-200 dark:border-secondary-800">
						<span class="text-2xl font-bold text-secondary-600 dark:text-secondary-400">{provider.trustScore}</span>
					</div>
				</div>
				<p class="mt-2 text-center text-sm text-gray-500 dark:text-gray-400">out of 100</p>
			</Card>

			<!-- Details -->
			<Card>
				<h2 class="font-semibold text-gray-900 dark:text-white">Details</h2>
				<div class="mt-4 space-y-3">
					<div>
						<p class="text-xs text-gray-500 dark:text-gray-400">Service Area</p>
						<p class="mt-0.5 text-sm text-gray-900 dark:text-white">{provider.serviceArea}</p>
					</div>
					<div>
						<p class="text-xs text-gray-500 dark:text-gray-400">Languages</p>
						<p class="mt-0.5 text-sm text-gray-900 dark:text-white">{provider.languages.join(', ')}</p>
					</div>
					<div>
						<p class="text-xs text-gray-500 dark:text-gray-400">Working Hours</p>
						<p class="mt-0.5 text-sm text-gray-900 dark:text-white">{provider.workingHours}</p>
					</div>
					<div>
						<p class="text-xs text-gray-500 dark:text-gray-400">Member Since</p>
						<p class="mt-0.5 text-sm text-gray-900 dark:text-white">{provider.memberSince}</p>
					</div>
					<div>
						<p class="text-xs text-gray-500 dark:text-gray-400">Rate</p>
						<p class="mt-0.5 text-sm font-semibold text-gray-900 dark:text-white">Rs. {provider.hourlyRate}/hour</p>
					</div>
				</div>
			</Card>

			<!-- Quick Actions -->
			<Card>
				<h2 class="font-semibold text-gray-900 dark:text-white">Quick Actions</h2>
				<div class="mt-4 space-y-2">
					<Button variant="primary" href="/jobs/new" class="w-full justify-center">
						<Briefcase class="h-4 w-4" />
						Request Quote
					</Button>
					<Button variant="outline" class="w-full justify-center" onclick={handleSendMessage} loading={sendingMessage}>
						<MessageSquare class="h-4 w-4" />
						{t('messages.start_conversation')}
					</Button>
					<Button variant="outline" class="w-full justify-center">
						<Phone class="h-4 w-4" />
						Call Provider
					</Button>
				</div>
			</Card>
		</div>
	</div>
</div>
{/if}
