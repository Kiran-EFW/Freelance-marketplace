<script lang="ts">
	import {
		Search, ArrowRight, CheckCircle, Star, Quote,
		Wrench, Sparkles, Scissors, Briefcase, Car, GraduationCap,
		HeartHandshake, PartyPopper, Truck, Monitor, TreePine, HardHat
	} from 'lucide-svelte';
	import { t, initLocale } from '$lib/i18n/index.svelte';
	import ServiceWizard from '$lib/components/ServiceWizard.svelte';
	import { topLevelCategories } from '$lib/data/categories';

	const iconMap: Record<string, any> = {
		Wrench, Sparkles, Scissors, Briefcase, Car, GraduationCap, HeartHandshake, PartyPopper, Truck, Monitor, TreePine, HardHat
	};

	const stats = [
		{ value: '12,000+', labelKey: 'home.verified_providers' },
		{ value: '85,000+', labelKey: 'home.jobs_completed' },
		{ value: '150+', labelKey: 'home.cities_covered' },
		{ value: '4.8', labelKey: 'home.avg_rating' }
	];

	const testimonials = [
		{
			name: 'Rajesh Kumar',
			role: 'Plumber, Bangalore',
			rating: 5,
			text: 'Seva has transformed my business. I get regular work and the route planning feature saves me hours every week. My income has grown 3x since joining.'
		},
		{
			name: 'Priya Sharma',
			role: 'Electrician, Mumbai',
			rating: 5,
			text: 'The platform is easy to use and customers trust providers on Seva because of the verification process. I love the instant payment feature.'
		},
		{
			name: 'Anand Patel',
			role: 'House Cleaner, Delhi',
			rating: 4,
			text: 'I started as a solo cleaner and now I manage a team of 5. Seva helped me build a reputation through reviews and the points system keeps me motivated.'
		}
	];

	$effect(() => {
		if (typeof window !== 'undefined') {
			initLocale();
		}
	});
</script>

<svelte:head>
	<title>Seva - Find Trusted Service Providers Near You</title>
	<meta name="description" content="Book, manage, and review home services with confidence. Connect with vetted professionals in your area. Plumbing, electrical, cleaning, gardening and more." />
</svelte:head>

<!-- Hero Section with Service Wizard -->
<section class="relative overflow-hidden bg-gradient-to-br from-primary-50 to-primary-100 dark:from-gray-900 dark:to-gray-800">
	<div class="absolute inset-0 opacity-5" style="background-image: radial-gradient(circle at 25% 25%, var(--color-primary-500) 1px, transparent 1px); background-size: 40px 40px;"></div>
	<div class="relative mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8 lg:py-16">
		<div class="text-center mb-8">
			<h1 class="text-3xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-4xl lg:text-5xl">
				{t('home.hero_title')}
				<span class="text-primary-600">{t('home.hero_title_highlight')}</span>
			</h1>
			<p class="mx-auto mt-4 max-w-2xl text-base text-gray-600 dark:text-gray-300">
				{t('home.hero_subtitle')}
			</p>
		</div>

		<ServiceWizard />
	</div>
</section>

<!-- Popular Categories -->
<section class="py-16 bg-white dark:bg-gray-950">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="text-center">
			<h2 class="text-3xl font-bold text-gray-900 dark:text-white">{t('home.popular_categories')}</h2>
			<p class="mt-3 text-gray-600 dark:text-gray-400">{t('home.popular_categories_subtitle')}</p>
		</div>
		<div class="mt-10 grid grid-cols-3 gap-4 sm:grid-cols-4 lg:gap-6">
			{#each topLevelCategories as cat}
				{@const Icon = iconMap[cat.icon] || Wrench}
				<a
					href="/providers?category={cat.slug}"
					class="group flex flex-col items-center gap-3 rounded-xl border border-gray-200 bg-white p-6 transition-all hover:border-primary-300 hover:shadow-md dark:border-gray-700 dark:bg-gray-800 dark:hover:border-primary-600"
				>
					<div class="flex h-14 w-14 items-center justify-center rounded-xl {cat.color} transition-transform group-hover:scale-110">
						<Icon class="h-7 w-7" />
					</div>
					<span class="text-sm font-medium text-gray-900 dark:text-white text-center">{t(cat.translationKey)}</span>
				</a>
			{/each}
		</div>
	</div>
</section>

<!-- How it Works -->
<section class="py-16 bg-gray-50 dark:bg-gray-900">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<h2 class="text-center text-3xl font-bold text-gray-900 dark:text-white">{t('home.how_it_works')}</h2>
		<p class="mt-3 text-center text-gray-600 dark:text-gray-400">{t('home.how_it_works_subtitle')}</p>
		<div class="mt-12 grid gap-8 sm:grid-cols-3">
			<div class="relative text-center">
				<div class="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-primary-100 text-primary-600 dark:bg-primary-900/30">
					<Search class="h-8 w-8" />
				</div>
				<div class="absolute -right-4 top-8 hidden text-gray-300 dark:text-gray-600 sm:block">
					<ArrowRight class="h-6 w-6" />
				</div>
				<h3 class="mt-5 text-lg font-semibold text-gray-900 dark:text-white">1. {t('home.step_search')}</h3>
				<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
					{t('home.step_search_desc')}
				</p>
			</div>
			<div class="relative text-center">
				<div class="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-secondary-100 text-secondary-600 dark:bg-secondary-900/30">
					<CheckCircle class="h-8 w-8" />
				</div>
				<div class="absolute -right-4 top-8 hidden text-gray-300 dark:text-gray-600 sm:block">
					<ArrowRight class="h-6 w-6" />
				</div>
				<h3 class="mt-5 text-lg font-semibold text-gray-900 dark:text-white">2. {t('home.step_book')}</h3>
				<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
					{t('home.step_book_desc')}
				</p>
			</div>
			<div class="text-center">
				<div class="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-blue-100 text-blue-600 dark:bg-blue-900/30">
					<Star class="h-8 w-8" />
				</div>
				<h3 class="mt-5 text-lg font-semibold text-gray-900 dark:text-white">3. {t('home.step_done')}</h3>
				<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
					{t('home.step_done_desc')}
				</p>
			</div>
		</div>
	</div>
</section>

<!-- Stats / Trusted By -->
<section class="py-16 bg-primary-600 dark:bg-primary-900">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<h2 class="text-center text-2xl font-bold text-white">{t('home.trusted_by')}</h2>
		<div class="mt-10 grid grid-cols-2 gap-8 lg:grid-cols-4">
			{#each stats as stat}
				<div class="text-center">
					<p class="text-3xl font-bold text-white sm:text-4xl">{stat.value}</p>
					<p class="mt-1 text-sm text-primary-100">{t(stat.labelKey)}</p>
				</div>
			{/each}
		</div>
	</div>
</section>

<!-- Testimonials -->
<section class="py-16 bg-white dark:bg-gray-950">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="text-center">
			<h2 class="text-3xl font-bold text-gray-900 dark:text-white">{t('home.what_providers_say')}</h2>
			<p class="mt-3 text-gray-600 dark:text-gray-400">{t('home.what_providers_say_subtitle')}</p>
		</div>
		<div class="mt-10 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
			{#each testimonials as testimonial}
				<div class="rounded-xl border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
					<Quote class="h-8 w-8 text-primary-200 dark:text-primary-800" />
					<p class="mt-3 text-sm text-gray-600 dark:text-gray-400">{testimonial.text}</p>
					<div class="mt-4 flex items-center gap-1">
						{#each Array(5) as _, i}
							<Star class="h-4 w-4 {i < testimonial.rating ? 'fill-yellow-400 text-yellow-400' : 'fill-none text-gray-300 dark:text-gray-600'}" />
						{/each}
					</div>
					<div class="mt-3 border-t border-gray-100 pt-3 dark:border-gray-700">
						<p class="text-sm font-semibold text-gray-900 dark:text-white">{testimonial.name}</p>
						<p class="text-xs text-gray-500 dark:text-gray-400">{testimonial.role}</p>
					</div>
				</div>
			{/each}
		</div>
	</div>
</section>

<!-- CTA Banner -->
<section class="py-16 bg-gradient-to-r from-secondary-600 to-secondary-700 dark:from-secondary-800 dark:to-secondary-900">
	<div class="mx-auto max-w-4xl px-4 text-center sm:px-6 lg:px-8">
		<h2 class="text-3xl font-bold text-white">{t('home.are_you_provider')}</h2>
		<p class="mt-4 text-lg text-secondary-100">
			{t('home.provider_cta_subtitle')}
		</p>
		<div class="mt-8 flex flex-col items-center justify-center gap-4 sm:flex-row">
			<a
				href="/register?role=provider"
				class="rounded-lg bg-white px-8 py-3 text-sm font-medium text-secondary-700 shadow hover:bg-gray-50"
			>
				{t('home.join_seva_provider')}
			</a>
			<a
				href="/providers"
				class="rounded-lg border border-white/30 px-8 py-3 text-sm font-medium text-white hover:bg-white/10"
			>
				{t('home.learn_more')}
			</a>
		</div>
	</div>
</section>
