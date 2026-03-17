<script lang="ts">
	import { MapPin, Star, Shield, Clock, Users, CheckCircle, ChevronDown, ChevronUp, ArrowRight, Briefcase } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import StarRating from '$lib/components/ui/StarRating.svelte';

	let { data } = $props();

	let expandedFaq = $state<number | null>(null);

	function toggleFaq(index: number) {
		expandedFaq = expandedFaq === index ? null : index;
	}
</script>

<svelte:head>
	<title>{data.service} in {data.area}, {data.city} - Seva</title>
	<meta name="description" content="Find trusted {data.service.toLowerCase()} service providers in {data.area}, {data.city}. Compare quotes, read reviews, and book verified professionals on Seva." />
	<meta property="og:title" content="{data.service} in {data.area}, {data.city} - Seva" />
	<meta property="og:description" content="Find trusted {data.service.toLowerCase()} service providers in {data.area}, {data.city}." />
	<link rel="canonical" href="https://seva.app/{data.serviceSlug}/{data.citySlug}/{data.areaSlug}" />
</svelte:head>

<div class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
	<!-- Breadcrumb -->
	<nav class="text-sm text-gray-500 dark:text-gray-400">
		<ol class="flex items-center gap-2">
			<li><a href="/" class="hover:text-primary-600">Home</a></li>
			<li>/</li>
			<li><a href="/{data.serviceSlug}/{data.citySlug}/{data.areaSlug}" class="hover:text-primary-600">{data.service}</a></li>
			<li>/</li>
			<li>{data.city}</li>
			<li>/</li>
			<li class="text-gray-900 dark:text-white">{data.area}</li>
		</ol>
	</nav>

	<!-- Hero Section -->
	<div class="mt-6">
		<h1 class="text-3xl font-bold text-gray-900 dark:text-white lg:text-4xl">
			{data.service} Services in {data.area}, {data.city}
		</h1>
		<p class="mt-3 text-lg text-gray-600 dark:text-gray-400">
			Find and hire verified {data.service.toLowerCase()} professionals in {data.area}. Compare prices, read reviews, and book instantly.
		</p>
	</div>

	<!-- Stats Bar -->
	<div class="mt-8 grid grid-cols-2 gap-4 rounded-2xl bg-primary-50 p-6 dark:bg-primary-900/10 sm:grid-cols-4">
		<div class="text-center">
			<p class="text-2xl font-bold text-primary-700 dark:text-primary-400">{data.stats.totalProviders}</p>
			<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">Providers Available</p>
		</div>
		<div class="text-center">
			<p class="text-2xl font-bold text-primary-700 dark:text-primary-400">{data.stats.avgRating}</p>
			<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">Avg Rating</p>
		</div>
		<div class="text-center">
			<p class="text-2xl font-bold text-primary-700 dark:text-primary-400">{data.stats.completedJobs.toLocaleString()}</p>
			<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">Jobs Completed</p>
		</div>
		<div class="text-center">
			<p class="text-2xl font-bold text-primary-700 dark:text-primary-400">{data.stats.avgResponseTime}</p>
			<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">Avg Response</p>
		</div>
	</div>

	<!-- CTA -->
	<div class="mt-8 flex flex-col gap-4 rounded-2xl bg-gradient-to-r from-primary-600 to-primary-700 p-6 text-white sm:flex-row sm:items-center sm:justify-between dark:from-primary-700 dark:to-primary-800">
		<div>
			<h2 class="text-xl font-bold">Need {data.service.toLowerCase()} help in {data.area}?</h2>
			<p class="mt-1 text-primary-100">Post a job and get quotes from verified providers in minutes.</p>
		</div>
		<Button href="/jobs/new" variant="secondary" class="shrink-0">
			<Briefcase class="h-4 w-4" />
			Post a Job
		</Button>
	</div>

	<!-- Providers List -->
	<div class="mt-10">
		<h2 class="text-2xl font-bold text-gray-900 dark:text-white">
			Top {data.service} Providers in {data.area}
		</h2>
		<div class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each data.providers as provider}
				<Card hover>
					<div class="flex flex-col h-full">
						<div class="flex items-start gap-3">
							<Avatar name={provider.name} size="md" />
							<div class="flex-1 min-w-0">
								<div class="flex items-center gap-2">
									<h3 class="font-semibold text-gray-900 dark:text-white truncate">{provider.name}</h3>
									{#if provider.isVerified}
										<Shield class="h-4 w-4 shrink-0 text-secondary-500" />
									{/if}
								</div>
								<div class="mt-1 flex items-center gap-1">
									<StarRating rating={provider.rating} size="sm" />
									<span class="text-xs text-gray-500 dark:text-gray-400">({provider.reviewCount})</span>
								</div>
							</div>
						</div>

						<div class="mt-3 flex flex-wrap gap-1.5">
							{#each provider.skills.slice(0, 3) as skill}
								<Badge variant="neutral" size="sm">{skill}</Badge>
							{/each}
						</div>

						<div class="mt-3 flex items-center gap-4 text-sm text-gray-500 dark:text-gray-400">
							<span class="flex items-center gap-1">
								<MapPin class="h-3.5 w-3.5" />
								{provider.distance}
							</span>
							<span class="flex items-center gap-1">
								<CheckCircle class="h-3.5 w-3.5" />
								{provider.completedJobs} jobs
							</span>
						</div>

						<div class="mt-4 flex items-center justify-between border-t border-gray-100 pt-4 dark:border-gray-700">
							<div>
								<span class="text-lg font-bold text-gray-900 dark:text-white">Rs. {provider.hourlyRate}</span>
								<span class="text-sm text-gray-500 dark:text-gray-400">/hour</span>
							</div>
							<Button variant="primary" size="sm" href="/providers/{provider.id}">
								View Profile
							</Button>
						</div>
					</div>
				</Card>
			{/each}
		</div>

		<div class="mt-6 text-center">
			<Button variant="outline" href="/providers?service={data.serviceSlug}&area={data.areaSlug}">
				View All {data.service} Providers in {data.area}
				<ArrowRight class="h-4 w-4" />
			</Button>
		</div>
	</div>

	<!-- FAQs -->
	<div class="mt-16">
		<h2 class="text-2xl font-bold text-gray-900 dark:text-white">
			Frequently Asked Questions
		</h2>
		<div class="mt-6 space-y-3">
			{#each data.faqs as faq, i}
				<div class="rounded-xl border border-gray-200 dark:border-gray-700">
					<button
						onclick={() => toggleFaq(i)}
						class="flex w-full items-center justify-between px-6 py-4 text-left"
					>
						<h3 class="text-sm font-medium text-gray-900 dark:text-white pr-4">{faq.question}</h3>
						{#if expandedFaq === i}
							<ChevronUp class="h-5 w-5 shrink-0 text-gray-400" />
						{:else}
							<ChevronDown class="h-5 w-5 shrink-0 text-gray-400" />
						{/if}
					</button>
					{#if expandedFaq === i}
						<div class="border-t border-gray-200 px-6 py-4 dark:border-gray-700">
							<p class="text-sm text-gray-600 dark:text-gray-400">{faq.answer}</p>
						</div>
					{/if}
				</div>
			{/each}
		</div>
	</div>

	<!-- Related Links -->
	<div class="mt-16 grid gap-8 lg:grid-cols-2">
		<!-- Related Services -->
		<div>
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Other Services in {data.area}</h2>
			<div class="mt-4 flex flex-wrap gap-2">
				{#each data.relatedServices as svc}
					<a
						href="/{svc.slug}/{data.citySlug}/{data.areaSlug}"
						class="rounded-lg border border-gray-200 px-4 py-2 text-sm text-gray-700 transition hover:border-primary-300 hover:text-primary-600 dark:border-gray-700 dark:text-gray-300 dark:hover:border-primary-600 dark:hover:text-primary-400"
					>
						{svc.label}
					</a>
				{/each}
			</div>
		</div>

		<!-- Nearby Areas -->
		<div>
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">{data.service} in Nearby Areas</h2>
			<div class="mt-4 flex flex-wrap gap-2">
				{#each data.nearbyAreas as area}
					<a
						href="/{data.serviceSlug}/{data.citySlug}/{area.slug}"
						class="rounded-lg border border-gray-200 px-4 py-2 text-sm text-gray-700 transition hover:border-primary-300 hover:text-primary-600 dark:border-gray-700 dark:text-gray-300 dark:hover:border-primary-600 dark:hover:text-primary-400"
					>
						{area.label}
					</a>
				{/each}
			</div>
		</div>
	</div>

	<!-- Bottom CTA -->
	<div class="mt-16 rounded-2xl bg-gray-50 p-8 text-center dark:bg-gray-800">
		<h2 class="text-2xl font-bold text-gray-900 dark:text-white">
			Ready to hire a {data.service.toLowerCase()} professional?
		</h2>
		<p class="mt-2 text-gray-600 dark:text-gray-400">
			Join thousands of satisfied customers in {data.area}, {data.city}.
		</p>
		<div class="mt-6 flex justify-center gap-4">
			<Button variant="primary" href="/jobs/new">Post a Job</Button>
			<Button variant="outline" href="/register">Sign Up Free</Button>
		</div>
	</div>
</div>
