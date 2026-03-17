<script lang="ts">
	import { MapPin, Shield, Clock } from 'lucide-svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import StarRating from '$lib/components/ui/StarRating.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import type { ProviderProfile } from '$lib/types';

	interface Props {
		provider: ProviderProfile;
		distance?: string;
		onRequestQuote?: (provider: ProviderProfile) => void;
		class?: string;
	}

	let {
		provider,
		distance = '',
		onRequestQuote,
		class: className = ''
	}: Props = $props();

	const name = $derived(provider.user?.name || provider.business_name || 'Provider');
	const skills = $derived(provider.categories?.map((c) => c.name) || []);
	const priceDisplay = $derived(
		provider.hourly_rate ? `From Rs. ${provider.hourly_rate}/hr` : 'Request quote'
	);
</script>

<div class="rounded-xl border border-gray-200 bg-white p-5 transition-shadow hover:shadow-md dark:border-gray-700 dark:bg-gray-800 {className}">
	<div class="flex gap-4">
		<Avatar
			src={provider.user?.avatar_url}
			name={name}
			size="lg"
		/>

		<div class="flex-1 min-w-0">
			<div class="flex items-start justify-between">
				<div>
					<div class="flex items-center gap-2">
						<h3 class="font-semibold text-gray-900 dark:text-white truncate">{name}</h3>
						{#if provider.verification_status === 'approved'}
							<Shield class="h-4 w-4 shrink-0 text-secondary-500" />
						{/if}
					</div>
					{#if provider.bio}
						<p class="mt-0.5 text-sm text-gray-500 dark:text-gray-400 line-clamp-1">{provider.bio}</p>
					{/if}
				</div>
				<span class="shrink-0 text-sm font-semibold text-primary-600 dark:text-primary-400">{priceDisplay}</span>
			</div>

			<div class="mt-2 flex items-center gap-3 text-sm">
				<div class="flex items-center gap-1">
					<StarRating rating={provider.rating_average} size="sm" />
					<span class="font-medium text-gray-700 dark:text-gray-300">{provider.rating_average.toFixed(1)}</span>
					<span class="text-gray-400 dark:text-gray-500">({provider.rating_count})</span>
				</div>
				{#if distance}
					<div class="flex items-center gap-1 text-gray-500 dark:text-gray-400">
						<MapPin class="h-3.5 w-3.5" />
						<span>{distance}</span>
					</div>
				{/if}
				{#if provider.response_time_minutes}
					<div class="flex items-center gap-1 text-gray-500 dark:text-gray-400">
						<Clock class="h-3.5 w-3.5" />
						<span>{provider.response_time_minutes < 60 ? `${provider.response_time_minutes}m` : `${Math.round(provider.response_time_minutes / 60)}h`} reply</span>
					</div>
				{/if}
			</div>

			{#if skills.length > 0}
				<div class="mt-2.5 flex flex-wrap gap-1.5">
					{#each skills.slice(0, 4) as skill}
						<Badge variant="neutral" size="sm">{skill}</Badge>
					{/each}
					{#if skills.length > 4}
						<Badge variant="neutral" size="sm">+{skills.length - 4} more</Badge>
					{/if}
				</div>
			{/if}
		</div>
	</div>

	<div class="mt-4 flex gap-2 border-t border-gray-100 pt-4 dark:border-gray-700">
		<Button variant="outline" size="sm" href="/providers/{provider.id}">View Profile</Button>
		<Button variant="primary" size="sm" onclick={() => onRequestQuote?.(provider)}>Request Quote</Button>
	</div>
</div>
