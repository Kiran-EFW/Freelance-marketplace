<script lang="ts">
	import { onMount } from 'svelte';
	import { Camera, MapPin, Phone, Mail, Globe, Bell, Award, Shield, Loader2 } from 'lucide-svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import FileUpload from '$lib/components/ui/FileUpload.svelte';
	import { toastSuccess, toastError } from '$lib/stores/toast';
	import api from '$lib/api/client';
	import { getCurrentUser, setUser } from '$lib/stores/auth';

	let loading = $state(true);
	let error = $state('');

	let userId = $state('');
	let name = $state('');
	let email = $state('');
	let userPhone = $state('');
	let address = $state('');
	let postcode = $state('');
	let language = $state('en');
	let role = $state<'customer' | 'provider'>('customer');
	let pointsBalance = $state(0);
	let levelName = $state('Bronze');
	let avatarUrl = $state('');

	// Notification preferences
	let notifJobUpdates = $state(true);
	let notifQuotes = $state(true);
	let notifMarketing = $state(false);
	let notifSms = $state(true);

	// Provider-specific
	let bio = $state('');
	let hourlyRate = $state(0);
	let isOnline = $state(false);

	let saving = $state(false);

	const languages = [
		{ value: 'en', label: 'English' },
		{ value: 'hi', label: 'Hindi' },
		{ value: 'kn', label: 'Kannada' },
		{ value: 'ta', label: 'Tamil' },
		{ value: 'te', label: 'Telugu' },
		{ value: 'mr', label: 'Marathi' }
	];

	onMount(async () => {
		try {
			const res = await api.auth.me();
			const user = res.data;
			userId = user.id;
			name = user.name || '';
			email = user.email || '';
			userPhone = user.phone || '';
			address = user.address || '';
			postcode = user.postcode || '';
			language = user.language || 'en';
			role = (user.role as 'customer' | 'provider') || 'customer';
			pointsBalance = user.points_balance || 0;
			levelName = user.level?.name || 'Bronze';
			avatarUrl = user.avatar_url || '';

			if (user.notification_preferences) {
				notifJobUpdates = user.notification_preferences.job_updates ?? true;
				notifQuotes = user.notification_preferences.quotes ?? true;
				notifMarketing = user.notification_preferences.marketing ?? false;
				notifSms = user.notification_preferences.sms ?? true;
			}

			if (role === 'provider' && user.provider_profile) {
				bio = user.provider_profile.bio || '';
				hourlyRate = user.provider_profile.hourly_rate || 0;
				isOnline = user.provider_profile.is_online ?? false;
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load profile';
		} finally {
			loading = false;
		}
	});

	async function saveProfile(e: Event) {
		e.preventDefault();
		saving = true;
		try {
			const updateData: any = {
				name,
				email: email || undefined,
				address: address || undefined,
				postcode: postcode || undefined,
				language
			};

			const res = await api.users.update(userId, updateData);
			if (res.data) {
				setUser(res.data);
			}

			if (role === 'provider') {
				await api.providers.updateProfile({
					bio: bio || undefined,
					hourly_rate: hourlyRate || undefined
				} as any);
			}

			toastSuccess('Profile updated successfully');
		} catch (err) {
			toastError(err instanceof Error ? err.message : 'Failed to update profile');
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>Profile - Seva</title>
</svelte:head>

{#if loading}
<div class="flex items-center justify-center py-20">
	<Loader2 class="h-8 w-8 animate-spin text-primary-600" />
</div>
{:else if error}
<div class="mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8">
	<div class="rounded-lg border border-red-200 bg-red-50 p-6 text-center dark:border-red-800 dark:bg-red-900/20">
		<p class="text-red-600 dark:text-red-400">{error}</p>
	</div>
</div>
{:else}
<div class="mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8">
	<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Profile</h1>
	<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">Manage your account settings and preferences.</p>

	<form onsubmit={saveProfile} class="mt-8 space-y-6">
		<!-- Avatar Section -->
		<Card>
			<div class="flex items-center gap-6">
				<div class="relative">
					<Avatar src={avatarUrl} name={name} size="xl" />
					<button
						type="button"
						class="absolute bottom-0 right-0 flex h-8 w-8 items-center justify-center rounded-full bg-primary-600 text-white shadow hover:bg-primary-700"
						aria-label="Change avatar"
					>
						<Camera class="h-4 w-4" />
					</button>
				</div>
				<div>
					<h2 class="text-lg font-semibold text-gray-900 dark:text-white">{name}</h2>
					<p class="text-sm text-gray-500 dark:text-gray-400">{userPhone}</p>
					<div class="mt-2 flex items-center gap-2">
						<Badge variant="success">
							<Award class="mr-1 h-3 w-3" />
							{levelName} - {pointsBalance} pts
						</Badge>
						<Badge variant="info">
							<Shield class="mr-1 h-3 w-3" />
							Verified
						</Badge>
					</div>
				</div>
			</div>
		</Card>

		<!-- Personal Info -->
		<Card>
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Personal Information</h2>
			<div class="mt-4 grid gap-4 sm:grid-cols-2">
				<Input label="Full Name" bind:value={name} required />
				<Input label="Email" type="email" bind:value={email} icon={Mail} />
				<Input label="Phone" bind:value={userPhone} disabled icon={Phone} hint="Phone cannot be changed" />
				<Input label="Address" bind:value={address} icon={MapPin} />
				<Input label="Postcode" bind:value={postcode} />
				<div>
					<label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">Preferred Language</label>
					<select
						bind:value={language}
						class="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
					>
						{#each languages as lang}
							<option value={lang.value}>{lang.label}</option>
						{/each}
					</select>
				</div>
			</div>
		</Card>

		<!-- Provider-specific -->
		{#if role === 'provider'}
			<Card>
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Provider Settings</h2>
				<div class="mt-4 space-y-4">
					<div>
						<label for="bio-edit" class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">Bio</label>
						<textarea
							id="bio-edit"
							bind:value={bio}
							rows="3"
							class="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
						></textarea>
					</div>
					<Input label="Hourly Rate (Rs.)" type="number" bind:value={hourlyRate} />
					<div class="flex items-center justify-between">
						<div>
							<p class="text-sm font-medium text-gray-700 dark:text-gray-300">Online Status</p>
							<p class="text-xs text-gray-500 dark:text-gray-400">Set yourself as available for new jobs</p>
						</div>
						<button
							type="button"
							onclick={() => (isOnline = !isOnline)}
							class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors {isOnline ? 'bg-secondary-600' : 'bg-gray-300 dark:bg-gray-600'}"
						>
							<span class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform {isOnline ? 'translate-x-6' : 'translate-x-1'}"></span>
						</button>
					</div>
				</div>
			</Card>
		{/if}

		<!-- Notification Preferences -->
		<Card>
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Notification Preferences</h2>
			<div class="mt-4 space-y-4">
				{#each [
					{ label: 'Job updates', desc: 'Status changes, new quotes', bind: () => notifJobUpdates, toggle: () => (notifJobUpdates = !notifJobUpdates) },
					{ label: 'Quote notifications', desc: 'When you receive new quotes', bind: () => notifQuotes, toggle: () => (notifQuotes = !notifQuotes) },
					{ label: 'SMS notifications', desc: 'Receive SMS for important updates', bind: () => notifSms, toggle: () => (notifSms = !notifSms) },
					{ label: 'Marketing', desc: 'Promotions and new features', bind: () => notifMarketing, toggle: () => (notifMarketing = !notifMarketing) }
				] as pref}
					<div class="flex items-center justify-between">
						<div>
							<p class="text-sm font-medium text-gray-700 dark:text-gray-300">{pref.label}</p>
							<p class="text-xs text-gray-500 dark:text-gray-400">{pref.desc}</p>
						</div>
						<button
							type="button"
							onclick={pref.toggle}
							class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors {pref.bind() ? 'bg-primary-600' : 'bg-gray-300 dark:bg-gray-600'}"
						>
							<span class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform {pref.bind() ? 'translate-x-6' : 'translate-x-1'}"></span>
						</button>
					</div>
				{/each}
			</div>
		</Card>

		<div class="flex justify-end">
			<Button type="submit" variant="primary" loading={saving}>
				Save Changes
			</Button>
		</div>
	</form>
</div>
{/if}
