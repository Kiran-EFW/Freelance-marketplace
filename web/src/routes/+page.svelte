<script lang="ts">
	import { Search, MapPin, ArrowRight, Wrench, Zap, Sparkles, TreePine, Paintbrush, Truck, Hammer, Wind, CheckCircle, Users, Briefcase, Star, Quote } from 'lucide-svelte';

	let postcode = $state('');
	let category = $state('');

	const categories = [
		'Plumbing',
		'Electrical',
		'Cleaning',
		'Gardening',
		'Painting',
		'Moving',
		'Carpentry',
		'HVAC'
	];

	const categoryCards = [
		{ name: 'Plumbing', icon: Wrench, color: 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400' },
		{ name: 'Electrical', icon: Zap, color: 'bg-yellow-100 text-yellow-600 dark:bg-yellow-900/30 dark:text-yellow-400' },
		{ name: 'Cleaning', icon: Sparkles, color: 'bg-purple-100 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400' },
		{ name: 'Gardening', icon: TreePine, color: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400' },
		{ name: 'Painting', icon: Paintbrush, color: 'bg-pink-100 text-pink-600 dark:bg-pink-900/30 dark:text-pink-400' },
		{ name: 'Moving', icon: Truck, color: 'bg-indigo-100 text-indigo-600 dark:bg-indigo-900/30 dark:text-indigo-400' },
		{ name: 'Carpentry', icon: Hammer, color: 'bg-orange-100 text-orange-600 dark:bg-orange-900/30 dark:text-orange-400' },
		{ name: 'HVAC', icon: Wind, color: 'bg-cyan-100 text-cyan-600 dark:bg-cyan-900/30 dark:text-cyan-400' }
	];

	const stats = [
		{ value: '12,000+', label: 'Verified Providers' },
		{ value: '85,000+', label: 'Jobs Completed' },
		{ value: '150+', label: 'Cities Covered' },
		{ value: '4.8', label: 'Avg. Rating' }
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
</script>

<svelte:head>
	<title>Seva - Find Trusted Service Providers Near You</title>
	<meta name="description" content="Book, manage, and review home services with confidence. Connect with vetted professionals in your area. Plumbing, electrical, cleaning, gardening and more." />
</svelte:head>

<!-- Hero Section -->
<section class="relative overflow-hidden bg-gradient-to-br from-primary-50 to-primary-100 dark:from-gray-900 dark:to-gray-800">
	<div class="absolute inset-0 opacity-5" style="background-image: radial-gradient(circle at 25% 25%, var(--color-primary-500) 1px, transparent 1px); background-size: 40px 40px;"></div>
	<div class="relative mx-auto max-w-7xl px-4 py-20 sm:px-6 lg:px-8 lg:py-28">
		<div class="text-center">
			<h1 class="text-4xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-5xl lg:text-6xl">
				Find trusted service providers
				<span class="text-primary-600">near you</span>
			</h1>
			<p class="mx-auto mt-6 max-w-2xl text-lg text-gray-600 dark:text-gray-300">
				Book, manage, and review home services with confidence. Connect with vetted professionals in your area.
			</p>
		</div>

		<!-- Search Box -->
		<div class="mx-auto mt-10 max-w-2xl">
			<div class="flex flex-col gap-3 rounded-xl bg-white p-4 shadow-lg dark:bg-gray-800 sm:flex-row">
				<div class="relative flex-1">
					<MapPin class="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-gray-400" />
					<input
						type="text"
						bind:value={postcode}
						placeholder="Enter your postcode"
						class="w-full rounded-lg border border-gray-300 py-3 pl-10 pr-4 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400"
					/>
				</div>
				<div class="relative flex-1">
					<Search class="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-gray-400" />
					<select
						bind:value={category}
						class="w-full appearance-none rounded-lg border border-gray-300 py-3 pl-10 pr-4 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
					>
						<option value="">Select a category</option>
						{#each categories as cat}
							<option value={cat}>{cat}</option>
						{/each}
					</select>
				</div>
				<a
					href="/providers?postcode={postcode}&category={category}"
					class="flex items-center justify-center gap-2 rounded-lg bg-primary-600 px-6 py-3 text-sm font-medium text-white hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500/20"
				>
					Search
					<ArrowRight class="h-4 w-4" />
				</a>
			</div>
		</div>

		<!-- CTA Buttons -->
		<div class="mt-10 flex flex-col items-center justify-center gap-4 sm:flex-row">
			<a
				href="/register"
				class="rounded-lg bg-primary-600 px-8 py-3 text-sm font-medium text-white hover:bg-primary-700"
			>
				Get Started as Customer
			</a>
			<a
				href="/register?role=provider"
				class="rounded-lg border border-primary-600 px-8 py-3 text-sm font-medium text-primary-600 hover:bg-primary-50 dark:hover:bg-gray-800"
			>
				Join as a Provider
			</a>
		</div>
	</div>
</section>

<!-- Popular Categories -->
<section class="py-16 bg-white dark:bg-gray-950">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="text-center">
			<h2 class="text-3xl font-bold text-gray-900 dark:text-white">Popular Categories</h2>
			<p class="mt-3 text-gray-600 dark:text-gray-400">Browse our most requested service categories</p>
		</div>
		<div class="mt-10 grid grid-cols-2 gap-4 sm:grid-cols-4 lg:gap-6">
			{#each categoryCards as cat}
				{@const Icon = cat.icon}
				<a
					href="/providers?category={cat.name}"
					class="group flex flex-col items-center gap-3 rounded-xl border border-gray-200 bg-white p-6 transition-all hover:border-primary-300 hover:shadow-md dark:border-gray-700 dark:bg-gray-800 dark:hover:border-primary-600"
				>
					<div class="flex h-14 w-14 items-center justify-center rounded-xl {cat.color} transition-transform group-hover:scale-110">
						<Icon class="h-7 w-7" />
					</div>
					<span class="text-sm font-medium text-gray-900 dark:text-white">{cat.name}</span>
				</a>
			{/each}
		</div>
	</div>
</section>

<!-- How it Works -->
<section class="py-16 bg-gray-50 dark:bg-gray-900">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<h2 class="text-center text-3xl font-bold text-gray-900 dark:text-white">How Seva Works</h2>
		<p class="mt-3 text-center text-gray-600 dark:text-gray-400">Three simple steps to get the help you need</p>
		<div class="mt-12 grid gap-8 sm:grid-cols-3">
			<div class="relative text-center">
				<div class="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-primary-100 text-primary-600 dark:bg-primary-900/30">
					<Search class="h-8 w-8" />
				</div>
				<div class="absolute -right-4 top-8 hidden text-gray-300 dark:text-gray-600 sm:block">
					<ArrowRight class="h-6 w-6" />
				</div>
				<h3 class="mt-5 text-lg font-semibold text-gray-900 dark:text-white">1. Search</h3>
				<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
					Browse verified service providers in your area by category, rating, and availability.
				</p>
			</div>
			<div class="relative text-center">
				<div class="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-secondary-100 text-secondary-600 dark:bg-secondary-900/30">
					<CheckCircle class="h-8 w-8" />
				</div>
				<div class="absolute -right-4 top-8 hidden text-gray-300 dark:text-gray-600 sm:block">
					<ArrowRight class="h-6 w-6" />
				</div>
				<h3 class="mt-5 text-lg font-semibold text-gray-900 dark:text-white">2. Book</h3>
				<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
					Request quotes, compare providers, and book the right professional for your job.
				</p>
			</div>
			<div class="text-center">
				<div class="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-blue-100 text-blue-600 dark:bg-blue-900/30">
					<Star class="h-8 w-8" />
				</div>
				<h3 class="mt-5 text-lg font-semibold text-gray-900 dark:text-white">3. Done</h3>
				<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
					Pay securely, track progress, and leave a review once the job is complete.
				</p>
			</div>
		</div>
	</div>
</section>

<!-- Stats / Trusted By -->
<section class="py-16 bg-primary-600 dark:bg-primary-900">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<h2 class="text-center text-2xl font-bold text-white">Trusted by thousands across India</h2>
		<div class="mt-10 grid grid-cols-2 gap-8 lg:grid-cols-4">
			{#each stats as stat}
				<div class="text-center">
					<p class="text-3xl font-bold text-white sm:text-4xl">{stat.value}</p>
					<p class="mt-1 text-sm text-primary-100">{stat.label}</p>
				</div>
			{/each}
		</div>
	</div>
</section>

<!-- Testimonials -->
<section class="py-16 bg-white dark:bg-gray-950">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="text-center">
			<h2 class="text-3xl font-bold text-gray-900 dark:text-white">What Providers Say</h2>
			<p class="mt-3 text-gray-600 dark:text-gray-400">Hear from service providers who have grown with Seva</p>
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
		<h2 class="text-3xl font-bold text-white">Are you a service provider?</h2>
		<p class="mt-4 text-lg text-secondary-100">
			Join Seva today and connect with thousands of customers in your area. Build your reputation, grow your business, and earn more.
		</p>
		<div class="mt-8 flex flex-col items-center justify-center gap-4 sm:flex-row">
			<a
				href="/register?role=provider"
				class="rounded-lg bg-white px-8 py-3 text-sm font-medium text-secondary-700 shadow hover:bg-gray-50"
			>
				Join Seva as Provider
			</a>
			<a
				href="/providers"
				class="rounded-lg border border-white/30 px-8 py-3 text-sm font-medium text-white hover:bg-white/10"
			>
				Learn More
			</a>
		</div>
	</div>
</section>
