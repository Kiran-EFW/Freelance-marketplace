<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { recurring } from '$lib/api/client';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import { Calendar, Plus, Pause, Play, Trash2, Clock, RefreshCw } from 'lucide-svelte';

	// State
	let schedules = $state<any[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let showCreateModal = $state(false);
	let creating = $state(false);
	let actionLoading = $state<string | null>(null);

	// Create form state
	let formTitle = $state('');
	let formDescription = $state('');
	let formProviderId = $state('');
	let formCategoryId = $state('');
	let formFrequency = $state<'daily' | 'weekly' | 'biweekly' | 'monthly' | 'quarterly'>('weekly');
	let formDayOfWeek = $state<number | undefined>(undefined);
	let formDayOfMonth = $state<number | undefined>(undefined);
	let formPreferredTime = $state('09:00');
	let formAmount = $state(0);
	let formCurrency = $state('INR');
	let formMaxOccurrences = $state<number | undefined>(undefined);

	const frequencyOptions = [
		{ value: 'daily', label: t('recurring.frequency_daily') },
		{ value: 'weekly', label: t('recurring.frequency_weekly') },
		{ value: 'biweekly', label: t('recurring.frequency_biweekly') },
		{ value: 'monthly', label: t('recurring.frequency_monthly') },
		{ value: 'quarterly', label: t('recurring.frequency_quarterly') }
	];

	const dayOfWeekOptions = [
		{ value: 0, label: 'Sunday' },
		{ value: 1, label: 'Monday' },
		{ value: 2, label: 'Tuesday' },
		{ value: 3, label: 'Wednesday' },
		{ value: 4, label: 'Thursday' },
		{ value: 5, label: 'Friday' },
		{ value: 6, label: 'Saturday' }
	];

	function getStatusVariant(status: string): 'success' | 'warning' | 'danger' | 'neutral' {
		switch (status) {
			case 'active': return 'success';
			case 'paused': return 'warning';
			case 'cancelled': return 'danger';
			default: return 'neutral';
		}
	}

	function getStatusLabel(status: string): string {
		switch (status) {
			case 'active': return t('recurring.status_active');
			case 'paused': return t('recurring.status_paused');
			case 'cancelled': return t('recurring.status_cancelled');
			default: return status;
		}
	}

	function getFrequencyLabel(frequency: string): string {
		switch (frequency) {
			case 'daily': return t('recurring.frequency_daily');
			case 'weekly': return t('recurring.frequency_weekly');
			case 'biweekly': return t('recurring.frequency_biweekly');
			case 'monthly': return t('recurring.frequency_monthly');
			case 'quarterly': return t('recurring.frequency_quarterly');
			default: return frequency;
		}
	}

	function formatNextOccurrence(dateStr: string | null): string {
		if (!dateStr) return '--';
		const date = new Date(dateStr);
		const now = new Date();
		const diffMs = date.getTime() - now.getTime();

		if (diffMs < 0) return 'Overdue';

		const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
		const diffDays = Math.floor(diffHours / 24);

		if (diffDays > 0) {
			return t('recurring.countdown', { time: `${diffDays}d ${diffHours % 24}h` });
		}
		if (diffHours > 0) {
			return t('recurring.countdown', { time: `${diffHours}h` });
		}
		const diffMinutes = Math.floor(diffMs / (1000 * 60));
		return t('recurring.countdown', { time: `${diffMinutes}m` });
	}

	async function loadSchedules() {
		loading = true;
		error = null;
		try {
			const response = await recurring.list();
			schedules = response.data || [];
		} catch (err: any) {
			error = err.message || 'Failed to load schedules';
		} finally {
			loading = false;
		}
	}

	async function createSchedule() {
		creating = true;
		try {
			await recurring.create({
				provider_id: formProviderId,
				category_id: formCategoryId,
				title: formTitle,
				description: formDescription || undefined,
				frequency: formFrequency,
				day_of_week: formDayOfWeek,
				day_of_month: formDayOfMonth,
				preferred_time: formPreferredTime,
				amount: formAmount,
				currency: formCurrency,
				max_occurrences: formMaxOccurrences
			});
			showCreateModal = false;
			resetForm();
			await loadSchedules();
		} catch (err: any) {
			error = err.message || 'Failed to create schedule';
		} finally {
			creating = false;
		}
	}

	async function pauseSchedule(id: string) {
		actionLoading = id;
		try {
			await recurring.pause(id);
			await loadSchedules();
		} catch (err: any) {
			error = err.message;
		} finally {
			actionLoading = null;
		}
	}

	async function resumeSchedule(id: string) {
		actionLoading = id;
		try {
			await recurring.resume(id);
			await loadSchedules();
		} catch (err: any) {
			error = err.message;
		} finally {
			actionLoading = null;
		}
	}

	async function cancelSchedule(id: string) {
		if (!confirm(t('recurring.cancel_confirm'))) return;
		actionLoading = id;
		try {
			await recurring.cancel(id);
			await loadSchedules();
		} catch (err: any) {
			error = err.message;
		} finally {
			actionLoading = null;
		}
	}

	function resetForm() {
		formTitle = '';
		formDescription = '';
		formProviderId = '';
		formCategoryId = '';
		formFrequency = 'weekly';
		formDayOfWeek = undefined;
		formDayOfMonth = undefined;
		formPreferredTime = '09:00';
		formAmount = 0;
		formCurrency = 'INR';
		formMaxOccurrences = undefined;
	}

	$effect(() => {
		loadSchedules();
	});
</script>

<svelte:head>
	<title>{t('recurring.title')} | Seva</title>
</svelte:head>

<div class="mx-auto max-w-4xl space-y-6 p-4 sm:p-6">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">
				{t('recurring.title')}
			</h1>
			<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
				{t('recurring.subtitle')}
			</p>
		</div>
		<Button onclick={() => (showCreateModal = true)}>
			<Plus class="h-4 w-4" />
			{t('recurring.create')}
		</Button>
	</div>

	<!-- Error -->
	{#if error}
		<div class="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-800 dark:bg-red-900/20 dark:text-red-400">
			{error}
		</div>
	{/if}

	<!-- Loading -->
	{#if loading}
		<div class="flex items-center justify-center py-16">
			<RefreshCw class="h-6 w-6 animate-spin text-gray-400" />
		</div>
	{:else if schedules.length === 0}
		<!-- Empty state -->
		<Card>
			<div class="flex flex-col items-center justify-center py-12 text-center">
				<Calendar class="mb-4 h-12 w-12 text-gray-300 dark:text-gray-600" />
				<h3 class="text-lg font-medium text-gray-900 dark:text-white">
					{t('recurring.no_schedules')}
				</h3>
				<p class="mt-2 max-w-sm text-sm text-gray-500 dark:text-gray-400">
					{t('recurring.no_schedules_desc')}
				</p>
				<div class="mt-6">
					<Button onclick={() => (showCreateModal = true)}>
						<Plus class="h-4 w-4" />
						{t('recurring.create')}
					</Button>
				</div>
			</div>
		</Card>
	{:else}
		<!-- Schedule list -->
		<div class="space-y-4">
			{#each schedules as schedule (schedule.id)}
				<Card>
					<div class="flex items-start justify-between">
						<div class="flex-1">
							<div class="flex items-center gap-3">
								<h3 class="text-lg font-semibold text-gray-900 dark:text-white">
									{schedule.title}
								</h3>
								<Badge variant={getStatusVariant(schedule.status)}>
									{getStatusLabel(schedule.status)}
								</Badge>
							</div>

							{#if schedule.description}
								<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
									{schedule.description}
								</p>
							{/if}

							<div class="mt-3 flex flex-wrap items-center gap-4 text-sm text-gray-600 dark:text-gray-300">
								<span class="flex items-center gap-1">
									<RefreshCw class="h-4 w-4" />
									{getFrequencyLabel(schedule.frequency)}
								</span>
								<span class="flex items-center gap-1">
									<Clock class="h-4 w-4" />
									{schedule.preferred_time || '09:00'}
								</span>
								<span class="font-medium text-gray-900 dark:text-white">
									{schedule.currency} {schedule.amount?.toFixed(2)}
								</span>
							</div>

							<div class="mt-2 flex flex-wrap items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
								{#if schedule.next_occurrence && schedule.status === 'active'}
									<span>
										{t('recurring.next_occurrence')}: {formatNextOccurrence(schedule.next_occurrence)}
									</span>
								{/if}
								<span>
									{t('recurring.total_occurrences')}: {schedule.total_occurrences || 0}
								</span>
								{#if schedule.max_occurrences}
									<span>
										{t('recurring.max_occurrences')}: {schedule.max_occurrences}
									</span>
								{/if}
							</div>
						</div>

						<!-- Actions -->
						{#if schedule.status !== 'cancelled'}
							<div class="ml-4 flex items-center gap-2">
								{#if schedule.status === 'active'}
									<Button
										variant="outline"
										size="sm"
										loading={actionLoading === schedule.id}
										onclick={() => pauseSchedule(schedule.id)}
									>
										<Pause class="h-3.5 w-3.5" />
										{t('recurring.pause')}
									</Button>
								{:else if schedule.status === 'paused'}
									<Button
										variant="outline"
										size="sm"
										loading={actionLoading === schedule.id}
										onclick={() => resumeSchedule(schedule.id)}
									>
										<Play class="h-3.5 w-3.5" />
										{t('recurring.resume')}
									</Button>
								{/if}
								<Button
									variant="danger"
									size="sm"
									loading={actionLoading === schedule.id}
									onclick={() => cancelSchedule(schedule.id)}
								>
									<Trash2 class="h-3.5 w-3.5" />
									{t('recurring.cancel')}
								</Button>
							</div>
						{/if}
					</div>
				</Card>
			{/each}
		</div>
	{/if}
</div>

<!-- Create Schedule Modal -->
{#if showCreateModal}
	<Modal open={showCreateModal} title={t('recurring.create_title')} onclose={() => (showCreateModal = false)}>
		<form
			class="space-y-4"
			onsubmit={(e) => {
				e.preventDefault();
				createSchedule();
			}}
		>
			<Input
				label={t('recurring.title_label')}
				bind:value={formTitle}
				required
				placeholder="e.g. Weekly Garden Maintenance"
			/>

			<Input
				label={t('recurring.description')}
				bind:value={formDescription}
				placeholder="Optional description"
			/>

			<Input
				label={t('recurring.provider') + ' ID'}
				bind:value={formProviderId}
				required
				placeholder="Provider UUID"
			/>

			<Input
				label={t('recurring.category') + ' ID'}
				bind:value={formCategoryId}
				required
				placeholder="Category UUID"
			/>

			<div class="grid grid-cols-2 gap-4">
				<div>
					<label for="schedule-frequency" class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">
						{t('recurring.frequency')}
					</label>
					<select
						id="schedule-frequency"
						bind:value={formFrequency}
						class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm dark:border-gray-600 dark:bg-gray-800 dark:text-gray-200"
					>
						{#each frequencyOptions as option}
							<option value={option.value}>{option.label}</option>
						{/each}
					</select>
				</div>

				<Input
					label={t('recurring.preferred_time')}
					type="time"
					bind:value={formPreferredTime}
				/>
			</div>

			{#if formFrequency === 'weekly' || formFrequency === 'biweekly'}
				<div>
					<label for="schedule-day-of-week" class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">
						{t('recurring.day_of_week')}
					</label>
					<select
						id="schedule-day-of-week"
						bind:value={formDayOfWeek}
						class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm dark:border-gray-600 dark:bg-gray-800 dark:text-gray-200"
					>
						<option value={undefined}>Select day</option>
						{#each dayOfWeekOptions as option}
							<option value={option.value}>{option.label}</option>
						{/each}
					</select>
				</div>
			{/if}

			{#if formFrequency === 'monthly' || formFrequency === 'quarterly'}
				<Input
					label={t('recurring.day_of_month')}
					type="number"
					bind:value={formDayOfMonth}
					min={1}
					max={31}
					placeholder="1-31"
				/>
			{/if}

			<div class="grid grid-cols-2 gap-4">
				<Input
					label={t('recurring.amount')}
					type="number"
					bind:value={formAmount}
					required
					min={1}
					step={0.01}
				/>

				<Input
					label={t('recurring.max_occurrences')}
					type="number"
					bind:value={formMaxOccurrences}
					placeholder={t('recurring.max_occurrences_hint')}
					min={1}
				/>
			</div>

			<div class="flex justify-end gap-3 pt-2">
				<Button variant="outline" onclick={() => (showCreateModal = false)}>
					Cancel
				</Button>
				<Button type="submit" loading={creating}>
					{creating ? t('recurring.saving') : t('recurring.save')}
				</Button>
			</div>
		</form>
	</Modal>
{/if}
