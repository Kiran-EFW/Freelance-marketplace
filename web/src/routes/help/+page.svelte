<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
</script>

<svelte:head>
	<title>Help - Seva</title>
</svelte:head>

<div class="container mx-auto px-4 py-8 max-w-4xl">
	<h1 class="text-3xl font-bold mb-6">{t('common.help') ?? 'Help Center'}</h1>

	<div class="space-y-6">
		<section>
			<h2 class="text-xl font-semibold mb-3">Getting Started</h2>
			<div class="bg-white rounded-lg shadow p-4 space-y-3">
				<details class="border-b pb-3">
					<summary class="cursor-pointer font-medium">How do I find a service provider?</summary>
					<p class="mt-2 text-gray-600">Use the search on our homepage to find providers by service type and location. You can filter by rating, distance, and price.</p>
				</details>
				<details class="border-b pb-3">
					<summary class="cursor-pointer font-medium">How do I post a job?</summary>
					<p class="mt-2 text-gray-600">Click "Post a Job" from the navigation menu, fill in the details about the service you need, and providers in your area will send you quotes.</p>
				</details>
				<details class="border-b pb-3">
					<summary class="cursor-pointer font-medium">How do payments work?</summary>
					<p class="mt-2 text-gray-600">Payments are held in escrow until the job is completed to your satisfaction. You can pay online or in cash as agreed with the provider.</p>
				</details>
				<details>
					<summary class="cursor-pointer font-medium">How do I contact support?</summary>
					<p class="mt-2 text-gray-600">Email us at support@seva.app or use the in-app messaging system to reach our support team.</p>
				</details>
			</div>
		</section>
	</div>
</div>
