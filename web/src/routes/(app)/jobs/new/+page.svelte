<script lang="ts">
	import { MapPin, Calendar, Clock, IndianRupee, Camera, CreditCard, Crosshair, ArrowLeft } from 'lucide-svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import FileUpload from '$lib/components/ui/FileUpload.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import { toastSuccess, toastError } from '$lib/stores/toast';
	import { goto } from '$app/navigation';
	import api from '$lib/api/client';

	let title = $state('');
	let description = $state('');
	let topCategory = $state('');
	let subCategory = $state('');
	let postcode = $state('');
	let preferredDate = $state('');
	let timeSlot = $state('flexible');
	let budgetType = $state<'range' | 'quote'>('range');
	let budgetMin = $state(500);
	let budgetMax = $state(5000);
	let paymentMethod = $state('upi');
	let photos = $state<File[]>([]);
	let loading = $state(false);

	const topCategories = [
		{ value: 'plumbing', label: 'Plumbing' },
		{ value: 'electrical', label: 'Electrical' },
		{ value: 'cleaning', label: 'Cleaning' },
		{ value: 'gardening', label: 'Gardening' },
		{ value: 'painting', label: 'Painting' },
		{ value: 'carpentry', label: 'Carpentry' },
		{ value: 'hvac', label: 'HVAC' },
		{ value: 'moving', label: 'Moving' }
	];

	const subCategories: Record<string, { value: string; label: string }[]> = {
		plumbing: [
			{ value: 'leak_repair', label: 'Leak Repair' },
			{ value: 'pipe_installation', label: 'Pipe Installation' },
			{ value: 'drain_cleaning', label: 'Drain Cleaning' },
			{ value: 'water_heater', label: 'Water Heater' }
		],
		electrical: [
			{ value: 'wiring', label: 'Wiring' },
			{ value: 'switch_socket', label: 'Switch / Socket' },
			{ value: 'fan_installation', label: 'Fan Installation' },
			{ value: 'inverter', label: 'Inverter / UPS' }
		],
		cleaning: [
			{ value: 'deep_cleaning', label: 'Deep Cleaning' },
			{ value: 'regular', label: 'Regular Cleaning' },
			{ value: 'kitchen', label: 'Kitchen Cleaning' },
			{ value: 'bathroom', label: 'Bathroom Cleaning' }
		]
	};

	const timeSlots = [
		{ value: 'morning', label: 'Morning (8am - 12pm)' },
		{ value: 'afternoon', label: 'Afternoon (12pm - 4pm)' },
		{ value: 'evening', label: 'Evening (4pm - 8pm)' },
		{ value: 'flexible', label: 'Flexible' }
	];

	const availableSubCategories = $derived(subCategories[topCategory] || []);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!title.trim() || !topCategory || !postcode.trim()) {
			toastError('Please fill in all required fields');
			return;
		}
		loading = true;
		try {
			const jobData: any = {
				title: title.trim(),
				description: description.trim(),
				category_id: subCategory || topCategory,
				location: { postcode: postcode.trim() },
				preferred_time_slot: timeSlot
			};
			if (preferredDate) jobData.preferred_date = preferredDate;
			if (budgetType === 'range') {
				jobData.budget_min = budgetMin;
				jobData.budget_max = budgetMax;
			}

			const res = await api.jobs.create(jobData);

			// Upload photos if any
			if (photos.length > 0 && res.data?.id) {
				try {
					await api.jobs.uploadImages(res.data.id, photos);
				} catch {
					// Non-critical: job was created, photos failed
					toastError('Job posted but photo upload failed');
				}
			}

			toastSuccess('Job posted successfully! Providers will start sending quotes.');
			goto('/jobs');
		} catch (err) {
			toastError(err instanceof Error ? err.message : 'Failed to post job');
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Post a Job - Seva</title>
</svelte:head>

<div class="mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8">
	<a href="/jobs" class="inline-flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700">
		<ArrowLeft class="h-4 w-4" />
		Back to Jobs
	</a>

	<h1 class="mt-4 text-2xl font-bold text-gray-900 dark:text-white">Post a New Job</h1>
	<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">Describe what you need and get quotes from local providers.</p>

	<form onsubmit={handleSubmit} class="mt-8 space-y-6">
		<!-- Category -->
		<Card>
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Service Category</h2>
			<div class="mt-4 grid gap-4 sm:grid-cols-2">
				<Select options={topCategories} bind:value={topCategory} label="Category" placeholder="Select category" />
				{#if availableSubCategories.length > 0}
					<Select options={availableSubCategories} bind:value={subCategory} label="Sub-category" placeholder="Select sub-category" />
				{/if}
			</div>
		</Card>

		<!-- Title & Description -->
		<Card>
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Job Details</h2>
			<div class="mt-4 space-y-4">
				<Input label="Job Title" bind:value={title} required placeholder="e.g., Fix leaking kitchen tap" />
				<div>
					<label for="desc" class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">Description</label>
					<textarea
						id="desc"
						bind:value={description}
						rows="4"
						placeholder="Describe the issue or work needed in detail..."
						class="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400"
					></textarea>
				</div>
			</div>
		</Card>

		<!-- Location -->
		<Card>
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Location</h2>
			<div class="mt-4 flex gap-3">
				<div class="flex-1">
					<Input label="Postcode" bind:value={postcode} required placeholder="560001" icon={MapPin} />
				</div>
				<div class="flex items-end">
					<Button variant="outline" size="md" onclick={() => { postcode = '560001'; toastSuccess('Location detected'); }}>
						<Crosshair class="h-4 w-4" />
						Use my location
					</Button>
				</div>
			</div>
		</Card>

		<!-- Schedule -->
		<Card>
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Schedule</h2>
			<div class="mt-4 grid gap-4 sm:grid-cols-2">
				<Input label="Preferred Date" type="date" bind:value={preferredDate} icon={Calendar} />
				<Select options={timeSlots} bind:value={timeSlot} label="Time Preference" />
			</div>
		</Card>

		<!-- Budget -->
		<Card>
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Budget</h2>
			<div class="mt-4">
				<div class="flex gap-4">
					<label class="flex items-center gap-2">
						<input type="radio" bind:group={budgetType} value="range" class="accent-primary-600" />
						<span class="text-sm text-gray-700 dark:text-gray-300">Set a budget range</span>
					</label>
					<label class="flex items-center gap-2">
						<input type="radio" bind:group={budgetType} value="quote" class="accent-primary-600" />
						<span class="text-sm text-gray-700 dark:text-gray-300">Let providers quote</span>
					</label>
				</div>
				{#if budgetType === 'range'}
					<div class="mt-4 grid gap-4 sm:grid-cols-2">
						<Input label="Minimum (Rs.)" type="number" bind:value={budgetMin} icon={IndianRupee} />
						<Input label="Maximum (Rs.)" type="number" bind:value={budgetMax} icon={IndianRupee} />
					</div>
				{/if}
			</div>
		</Card>

		<!-- Photos -->
		<Card>
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Photos (optional)</h2>
			<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Add up to 5 photos to help providers understand the job.</p>
			<div class="mt-4">
				<FileUpload
					accept="image/*"
					multiple
					maxSize={5}
					onUpload={(files) => { photos = files; }}
				/>
			</div>
		</Card>

		<!-- Payment Method -->
		<Card>
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Payment Method</h2>
			<div class="mt-4 grid gap-3 sm:grid-cols-3">
				{#each [
					{ value: 'upi', label: 'UPI' },
					{ value: 'card', label: 'Card' },
					{ value: 'cash', label: 'Cash' }
				] as method}
					<button
						type="button"
						onclick={() => (paymentMethod = method.value)}
						class="flex items-center gap-2 rounded-lg border-2 px-4 py-3 text-sm font-medium transition
							{paymentMethod === method.value
								? 'border-primary-500 bg-primary-50 text-primary-700 dark:border-primary-400 dark:bg-primary-900/20 dark:text-primary-400'
								: 'border-gray-200 text-gray-600 hover:border-gray-300 dark:border-gray-600 dark:text-gray-400'}"
					>
						<CreditCard class="h-4 w-4" />
						{method.label}
					</button>
				{/each}
			</div>
		</Card>

		<!-- Submit -->
		<div class="flex justify-end gap-3">
			<Button variant="outline" href="/jobs">Cancel</Button>
			<Button type="submit" variant="primary" size="lg" {loading}>
				Post Job
			</Button>
		</div>
	</form>
</div>
