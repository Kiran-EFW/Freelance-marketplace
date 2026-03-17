<script lang="ts">
	import { page } from '$app/stores';
	import { ArrowLeft, Star, MapPin, Clock, Shield, Phone, MessageSquare, Briefcase, Calendar, CheckCircle, IndianRupee } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import StarRating from '$lib/components/ui/StarRating.svelte';

	let providerId = $derived($page.params.id);

	const provider = {
		id: '1',
		name: 'Suresh Nair',
		bio: 'Experienced plumber with over 10 years of expertise. I specialize in residential plumbing, water heater installation, kitchen and bathroom renovations, and emergency leak repairs. Committed to quality work and customer satisfaction.',
		rating: 4.8,
		reviewCount: 156,
		completedJobs: 156,
		memberSince: '2025-07-20',
		responseTime: '30 minutes',
		hourlyRate: 500,
		isVerified: true,
		isOnline: true,
		trustScore: 92,
		skills: ['Plumbing', 'Pipe Fitting', 'Water Heater', 'Leak Repair', 'Bathroom Renovation', 'Kitchen Plumbing'],
		serviceArea: 'Koramangala, HSR Layout, BTM Layout, Jayanagar',
		languages: ['English', 'Hindi', 'Kannada'],
		workingHours: 'Mon-Sat, 8:00 AM - 7:00 PM'
	};

	const reviews = [
		{
			id: '1', customerName: 'Amit Verma', rating: 5, comment: 'Excellent work! Suresh fixed our kitchen plumbing quickly and professionally. Very punctual and clean work.',
			date: '2026-03-15', jobTitle: 'Kitchen plumbing repair'
		},
		{
			id: '2', customerName: 'Priya Menon', rating: 5, comment: 'Second time using Suresh for plumbing work. Always reliable and does quality work. Highly recommended.',
			date: '2026-03-10', jobTitle: 'Bathroom pipe repair'
		},
		{
			id: '3', customerName: 'Arjun Das', rating: 4, comment: 'Good work on the water heater installation. Arrived on time and completed within the estimated timeframe.',
			date: '2026-03-05', jobTitle: 'Water heater installation'
		},
		{
			id: '4', customerName: 'Meera Reddy', rating: 5, comment: 'Very knowledgeable and professional. Diagnosed the issue quickly and fixed it for a fair price.',
			date: '2026-02-28', jobTitle: 'Emergency leak repair'
		},
		{
			id: '5', customerName: 'Kiran Rao', rating: 4, comment: 'Solid work. Would have preferred more communication about the timeline, but the end result was great.',
			date: '2026-02-20', jobTitle: 'Pipe fitting'
		}
	];

	const completedPhotos = [
		{ id: '1', title: 'Kitchen plumbing renovation', description: 'Complete sink and pipe replacement' },
		{ id: '2', title: 'Water heater installation', description: 'New geyser installation with copper piping' },
		{ id: '3', title: 'Bathroom renovation', description: 'Full bathroom plumbing overhaul' },
		{ id: '4', title: 'Emergency pipe repair', description: 'Fixed burst pipe in basement' }
	];

	const ratingDist = { 5: 98, 4: 35, 3: 15, 2: 5, 1: 3 };
</script>

<svelte:head>
	<title>{provider.name} - Seva</title>
	<meta name="description" content="{provider.bio.substring(0, 160)}" />
</svelte:head>

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
					<Button variant="outline" class="w-full justify-center">
						<MessageSquare class="h-4 w-4" />
						Send Message
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
