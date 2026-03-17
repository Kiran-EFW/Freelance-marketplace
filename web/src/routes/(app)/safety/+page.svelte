<script lang="ts">
	import { onMount } from 'svelte';
	import { Shield, Phone, Plus, Trash2, AlertTriangle, MapPin, Info, Loader2, CheckCircle, XCircle } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import api from '$lib/api/client';
	import { t } from '$lib/i18n/index.svelte';

	let loading = $state(true);
	let error = $state('');

	// Data
	let contacts = $state<any[]>([]);
	let alerts = $state<any[]>([]);

	// SOS state
	let showSOSConfirm = $state(false);
	let triggeringSOS = $state(false);
	let sosTriggered = $state(false);
	let sosError = $state('');

	// Add contact modal
	let showAddContactModal = $state(false);
	let newContact = $state({ name: '', phone: '', relationship: '' });
	let addingContact = $state(false);

	// Location sharing
	let locationSharing = $state(false);
	let watchId = $state<number | null>(null);

	const safetyTips = [
		t('safety.tip_1'),
		t('safety.tip_2'),
		t('safety.tip_3'),
		t('safety.tip_4'),
		t('safety.tip_5')
	];

	const alertStatusBadge: Record<string, 'warning' | 'info' | 'success' | 'neutral' | 'danger'> = {
		active: 'danger',
		responded: 'warning',
		resolved: 'success',
		false_alarm: 'neutral'
	};

	onMount(async () => {
		await loadSafetyData();
	});

	async function loadSafetyData() {
		loading = true;
		error = '';
		try {
			const [contactsRes, alertsRes] = await Promise.all([
				api.safety.listEmergencyContacts(),
				api.safety.listAlerts({ limit: 10 })
			]);

			contacts = contactsRes.data || [];
			alerts = alertsRes.data || [];
		} catch (err) {
			error = err instanceof Error ? err.message : t('safety.error');
		} finally {
			loading = false;
		}
	}

	async function handleTriggerSOS() {
		triggeringSOS = true;
		sosError = '';

		try {
			// Get current position
			const position = await new Promise<GeolocationPosition>((resolve, reject) => {
				if (!navigator.geolocation) {
					reject(new Error('Geolocation is not supported'));
					return;
				}
				navigator.geolocation.getCurrentPosition(resolve, reject, {
					enableHighAccuracy: true,
					timeout: 10000
				});
			});

			await api.safety.triggerSOS({
				latitude: position.coords.latitude,
				longitude: position.coords.longitude,
				notes: 'SOS triggered from web app'
			});

			sosTriggered = true;
			showSOSConfirm = false;
			await loadSafetyData();
		} catch (err) {
			sosError = err instanceof Error ? err.message : t('safety.sos_failed');
		} finally {
			triggeringSOS = false;
		}
	}

	async function handleResolveAlert(alertId: string, status: string) {
		try {
			await api.safety.resolveSOS(alertId, { status });
			await loadSafetyData();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to resolve alert';
		}
	}

	async function handleAddContact() {
		addingContact = true;
		try {
			await api.safety.addEmergencyContact({
				name: newContact.name,
				phone: newContact.phone,
				relationship: newContact.relationship || undefined
			});
			showAddContactModal = false;
			newContact = { name: '', phone: '', relationship: '' };
			await loadSafetyData();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to add contact';
		} finally {
			addingContact = false;
		}
	}

	async function handleRemoveContact(contactId: string) {
		try {
			await api.safety.removeEmergencyContact(contactId);
			await loadSafetyData();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to remove contact';
		}
	}

	function toggleLocationSharing() {
		if (locationSharing && watchId !== null) {
			navigator.geolocation.clearWatch(watchId);
			watchId = null;
			locationSharing = false;
		} else {
			if (!navigator.geolocation) {
				error = 'Geolocation is not supported by your browser';
				return;
			}
			locationSharing = true;
			// Note: In production this would send to the API with a specific job_id
		}
	}
</script>

<svelte:head>
	<title>{t('safety.title')} - Seva</title>
</svelte:head>

{#if loading}
<div class="flex items-center justify-center py-20">
	<Loader2 class="h-8 w-8 animate-spin text-primary-600" />
</div>
{:else}
<div class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
	<!-- Header -->
	<div class="flex items-center gap-3">
		<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-red-100 text-red-600 dark:bg-red-900/30">
			<Shield class="h-6 w-6" />
		</div>
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">{t('safety.title')}</h1>
			<p class="text-sm text-gray-600 dark:text-gray-400">{t('safety.sos_description')}</p>
		</div>
	</div>

	{#if error}
	<div class="mt-4 rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-600 dark:border-red-800 dark:bg-red-900/20 dark:text-red-400">
		{error}
	</div>
	{/if}

	{#if sosTriggered}
	<div class="mt-4 rounded-lg border border-green-200 bg-green-50 p-4 text-sm text-green-700 dark:border-green-800 dark:bg-green-900/20 dark:text-green-400">
		<div class="flex items-center gap-2">
			<CheckCircle class="h-5 w-5" />
			{t('safety.sos_triggered')}
		</div>
	</div>
	{/if}

	<!-- SOS Button Section -->
	<Card class="mt-8">
		<div class="flex flex-col items-center py-6 text-center">
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">{t('safety.sos_title')}</h2>
			<p class="mt-2 max-w-md text-sm text-gray-500 dark:text-gray-400">{t('safety.sos_description')}</p>

			<button
				onclick={() => (showSOSConfirm = true)}
				class="mt-6 flex h-28 w-28 items-center justify-center rounded-full bg-red-600 text-3xl font-bold text-white shadow-lg shadow-red-200 transition-all hover:bg-red-700 hover:shadow-xl hover:shadow-red-300 active:scale-95 dark:shadow-red-900/40"
			>
				{t('safety.sos_button')}
			</button>

			{#if sosError}
			<p class="mt-4 text-sm text-red-600 dark:text-red-400">{sosError}</p>
			{/if}
		</div>
	</Card>

	<!-- Active Alerts -->
	{#if alerts.filter(a => a.status === 'active').length > 0}
	<div class="mt-8">
		<h2 class="text-lg font-semibold text-gray-900 dark:text-white">{t('safety.active_alerts')}</h2>
		<div class="mt-3 space-y-3">
			{#each alerts.filter(a => a.status === 'active') as alert}
			<Card>
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-3">
						<div class="flex h-10 w-10 items-center justify-center rounded-full bg-red-100 text-red-600 dark:bg-red-900/30">
							<AlertTriangle class="h-5 w-5" />
						</div>
						<div>
							<div class="flex items-center gap-2">
								<p class="text-sm font-medium text-gray-900 dark:text-white">SOS Alert</p>
								<Badge variant="danger" size="sm">Active</Badge>
							</div>
							<p class="text-xs text-gray-500 dark:text-gray-400">
								{new Date(alert.created_at).toLocaleString()}
								{#if alert.emergency_contacts_notified}
								<span class="ml-2 text-green-600">Contacts notified</span>
								{/if}
							</p>
						</div>
					</div>
					<div class="flex gap-2">
						<Button variant="primary" size="sm" onclick={() => handleResolveAlert(alert.id, 'resolved')}>
							<CheckCircle class="h-4 w-4" />
							{t('safety.resolve_alert')}
						</Button>
						<Button variant="outline" size="sm" onclick={() => handleResolveAlert(alert.id, 'false_alarm')}>
							{t('safety.false_alarm')}
						</Button>
					</div>
				</div>
			</Card>
			{/each}
		</div>
	</div>
	{/if}

	<!-- Emergency Contacts -->
	<div class="mt-8">
		<div class="flex items-center justify-between">
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">{t('safety.emergency_contacts')}</h2>
			<Button variant="primary" size="sm" onclick={() => (showAddContactModal = true)}>
				<Plus class="h-4 w-4" />
				{t('safety.add_contact')}
			</Button>
		</div>

		<div class="mt-3 space-y-3">
			{#if contacts.length === 0}
			<Card>
				<div class="py-6 text-center">
					<Phone class="mx-auto h-10 w-10 text-gray-400" />
					<p class="mt-3 text-sm text-gray-500 dark:text-gray-400">{t('safety.no_contacts')}</p>
				</div>
			</Card>
			{:else}
				{#each contacts as contact}
				<Card>
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-3">
							<div class="flex h-10 w-10 items-center justify-center rounded-full bg-blue-100 text-blue-600 dark:bg-blue-900/30">
								<Phone class="h-5 w-5" />
							</div>
							<div>
								<p class="text-sm font-medium text-gray-900 dark:text-white">{contact.name}</p>
								<div class="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
									<span>{contact.phone}</span>
									{#if contact.relationship}
									<span>({contact.relationship})</span>
									{/if}
								</div>
							</div>
						</div>
						<Button variant="ghost" size="sm" onclick={() => handleRemoveContact(contact.id)}>
							<Trash2 class="h-4 w-4 text-red-500" />
						</Button>
					</div>
				</Card>
				{/each}
			{/if}
		</div>
	</div>

	<!-- Location Sharing -->
	<Card class="mt-8">
		<div class="flex items-center justify-between">
			<div class="flex items-center gap-3">
				<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-green-100 text-green-600 dark:bg-green-900/30">
					<MapPin class="h-5 w-5" />
				</div>
				<div>
					<h2 class="text-sm font-semibold text-gray-900 dark:text-white">{t('safety.location_sharing')}</h2>
					<p class="text-xs text-gray-500 dark:text-gray-400">{t('safety.location_sharing_desc')}</p>
				</div>
			</div>
			<button
				aria-label={t('safety.location_sharing')}
				role="switch"
				aria-checked={locationSharing}
				onclick={toggleLocationSharing}
				class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors
					{locationSharing ? 'bg-green-600' : 'bg-gray-300 dark:bg-gray-600'}"
			>
				<span
					class="inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform
						{locationSharing ? 'translate-x-6' : 'translate-x-1'}"
				></span>
			</button>
		</div>
	</Card>

	<!-- Alert History -->
	{#if alerts.length > 0}
	<div class="mt-8">
		<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Alert History</h2>
		<div class="mt-3 space-y-2">
			{#each alerts as alert}
			<div class="flex items-center justify-between rounded-lg border border-gray-200 px-4 py-3 dark:border-gray-700">
				<div class="flex items-center gap-3">
					<AlertTriangle class="h-4 w-4 text-gray-400" />
					<div>
						<div class="flex items-center gap-2">
							<span class="text-sm text-gray-900 dark:text-white">SOS Alert</span>
							<Badge variant={alertStatusBadge[alert.status] || 'neutral'} size="sm">{alert.status.replace('_', ' ')}</Badge>
						</div>
						<span class="text-xs text-gray-500 dark:text-gray-400">{new Date(alert.created_at).toLocaleString()}</span>
					</div>
				</div>
			</div>
			{/each}
		</div>
	</div>
	{/if}

	<!-- Safety Tips -->
	<Card class="mt-8">
		<div class="flex items-center gap-2">
			<Info class="h-5 w-5 text-blue-500" />
			<h2 class="text-sm font-semibold text-gray-900 dark:text-white">{t('safety.safety_tips')}</h2>
		</div>
		<ul class="mt-3 space-y-2">
			{#each safetyTips as tip}
			<li class="flex items-start gap-2 text-sm text-gray-600 dark:text-gray-400">
				<CheckCircle class="mt-0.5 h-4 w-4 flex-shrink-0 text-green-500" />
				{tip}
			</li>
			{/each}
		</ul>
	</Card>
</div>
{/if}

<!-- SOS Confirmation Modal -->
<Modal bind:open={showSOSConfirm} title={t('safety.sos_title')} size="sm">
	<div class="text-center">
		<div class="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-red-100 text-red-600">
			<AlertTriangle class="h-8 w-8" />
		</div>
		<p class="mt-4 text-sm text-gray-600 dark:text-gray-400">{t('safety.sos_confirm')}</p>
	</div>
	{#snippet footer()}
		<Button variant="outline" onclick={() => (showSOSConfirm = false)}>{t('common.cancel')}</Button>
		<Button variant="danger" loading={triggeringSOS} onclick={handleTriggerSOS}>
			<AlertTriangle class="h-4 w-4" />
			{t('safety.sos_button')}
		</Button>
	{/snippet}
</Modal>

<!-- Add Emergency Contact Modal -->
<Modal bind:open={showAddContactModal} title={t('safety.add_contact')}>
	<div class="space-y-4">
		<Input label={t('safety.contact_name')} bind:value={newContact.name} required placeholder="e.g. John Doe" />
		<Input label={t('safety.contact_phone')} bind:value={newContact.phone} required placeholder="+91..." />
		<Input label={t('safety.relationship')} bind:value={newContact.relationship} placeholder="e.g. Spouse, Parent, Friend" />
	</div>
	{#snippet footer()}
		<Button variant="outline" onclick={() => (showAddContactModal = false)}>{t('common.cancel')}</Button>
		<Button variant="primary" loading={addingContact} onclick={handleAddContact}>{t('safety.add_contact')}</Button>
	{/snippet}
</Modal>
