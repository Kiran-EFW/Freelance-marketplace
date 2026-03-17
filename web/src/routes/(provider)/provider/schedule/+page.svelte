<script lang="ts">
	import { ChevronLeft, ChevronRight, Briefcase, Calendar, MapPin, Clock, User, Loader2 } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import api from '$lib/api/client';

	let loading = $state(true);
	let error = $state('');

	const daysOfWeek = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
	let weekOffset = $state(0);

	const now = new Date();
	const dayOfWeek = now.getDay();
	const baseDate = new Date(now);
	baseDate.setDate(now.getDate() - ((dayOfWeek + 6) % 7)); // Monday of current week

	function getWeekDates(offset: number) {
		const dates: { day: string; date: number; month: string; full: string; isToday: boolean }[] = [];
		const today = new Date();
		for (let i = 0; i < 7; i++) {
			const d = new Date(baseDate);
			d.setDate(d.getDate() + i + offset * 7);
			dates.push({
				day: daysOfWeek[i],
				date: d.getDate(),
				month: d.toLocaleString('en-US', { month: 'short' }),
				full: d.toISOString().split('T')[0],
				isToday: d.toDateString() === today.toDateString()
			});
		}
		return dates;
	}

	let weekDates = $derived(getWeekDates(weekOffset));
	let weekLabel = $derived(() => {
		const first = weekDates[0];
		const last = weekDates[6];
		return `${first.month} ${first.date} - ${last.month} ${last.date}`;
	});

	type ScheduleEntry = {
		id: string;
		type: 'job' | 'route';
		title: string;
		customer: string;
		time: string;
		duration: number;
		location: string;
		postcode: string;
		status: 'scheduled' | 'in_progress' | 'completed';
		date: string;
	};

	let scheduleData = $state<ScheduleEntry[]>([]);

	async function fetchSchedule() {
		loading = true;
		error = '';
		try {
			const dates = getWeekDates(weekOffset);
			const from = dates[0].full;
			const to = dates[6].full;
			const res = await api.routes.getSchedule({ from, to });
			scheduleData = (res.data || []).map((entry: any) => ({
				id: entry.id,
				type: entry.type || 'job',
				title: entry.title || '',
				customer: entry.customer?.name || entry.customer_name || '',
				time: entry.time || entry.start_time || '',
				duration: entry.duration || entry.duration_minutes || 60,
				location: entry.location || entry.address || '',
				postcode: entry.postcode || '',
				status: entry.status || 'scheduled',
				date: entry.date || entry.scheduled_date || ''
			}));
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load schedule';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		const _offset = weekOffset;
		fetchSchedule();
	});

	function getEntriesForDate(dateStr: string): ScheduleEntry[] {
		return scheduleData.filter((e) => e.date === dateStr).sort((a, b) => a.time.localeCompare(b.time));
	}

	let selectedDate = $state(new Date().toISOString().split('T')[0]);
	let selectedEntries = $derived(getEntriesForDate(selectedDate));

	const totalJobsThisWeek = $derived(
		weekDates.reduce((sum, d) => sum + getEntriesForDate(d.full).length, 0)
	);

	const totalHoursThisWeek = $derived(
		weekDates.reduce((sum, d) => {
			const entries = getEntriesForDate(d.full);
			return sum + entries.reduce((s, e) => s + e.duration, 0);
		}, 0)
	);

	const statusColors: Record<string, 'success' | 'warning' | 'info'> = {
		completed: 'success',
		in_progress: 'warning',
		scheduled: 'info'
	};
</script>

<svelte:head>
	<title>Schedule - Seva Provider</title>
</svelte:head>

{#if loading}
<div class="flex items-center justify-center py-20">
	<Loader2 class="h-8 w-8 animate-spin text-primary-600" />
</div>
{:else if error}
<div class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
	<div class="rounded-lg border border-red-200 bg-red-50 p-6 text-center dark:border-red-800 dark:bg-red-900/20">
		<p class="text-red-600 dark:text-red-400">{error}</p>
	</div>
</div>
{:else}
<div class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
	<div class="flex flex-wrap items-center justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Schedule</h1>
			<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">Manage your weekly appointments and route visits.</p>
		</div>
		<div class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
			<span>{totalJobsThisWeek} tasks</span>
			<span class="text-gray-300 dark:text-gray-600">|</span>
			<span>{Math.floor(totalHoursThisWeek / 60)}h {totalHoursThisWeek % 60}m total</span>
		</div>
	</div>

	<!-- Week Navigation -->
	<Card class="mt-6">
		<div class="flex items-center justify-between">
			<button onclick={() => (weekOffset -= 1)} class="rounded-lg p-2 text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700">
				<ChevronLeft class="h-5 w-5" />
			</button>
			<div class="text-center">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">{weekLabel()}</h2>
				{#if weekOffset !== 0}
					<button onclick={() => (weekOffset = 0)} class="mt-0.5 text-xs text-primary-600 hover:text-primary-700">
						Go to current week
					</button>
				{/if}
			</div>
			<button onclick={() => (weekOffset += 1)} class="rounded-lg p-2 text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700">
				<ChevronRight class="h-5 w-5" />
			</button>
		</div>

		<!-- Day Grid -->
		<div class="mt-4 grid grid-cols-7 gap-1">
			{#each weekDates as day}
				{@const entries = getEntriesForDate(day.full)}
				<button
					onclick={() => (selectedDate = day.full)}
					class="flex flex-col items-center rounded-xl p-2 transition
						{selectedDate === day.full
							? 'bg-primary-600 text-white'
							: day.isToday
								? 'bg-primary-50 text-primary-700 dark:bg-primary-900/20 dark:text-primary-400'
								: 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700'}"
				>
					<span class="text-xs font-medium">{day.day}</span>
					<span class="mt-1 text-lg font-bold">{day.date}</span>
					{#if entries.length > 0}
						<div class="mt-1 flex gap-0.5">
							{#each entries.slice(0, 3) as entry}
								<div class="h-1.5 w-1.5 rounded-full
									{selectedDate === day.full
										? 'bg-white/70'
										: entry.type === 'job'
											? 'bg-primary-500'
											: 'bg-blue-500'}">
								</div>
							{/each}
							{#if entries.length > 3}
								<div class="h-1.5 w-1.5 rounded-full
									{selectedDate === day.full ? 'bg-white/50' : 'bg-gray-400'}">
								</div>
							{/if}
						</div>
					{/if}
				</button>
			{/each}
		</div>
	</Card>

	<!-- Selected Day Details -->
	<div class="mt-6">
		<h2 class="text-lg font-semibold text-gray-900 dark:text-white">
			{#each weekDates as day}
				{#if day.full === selectedDate}
					{day.day}, {day.month} {day.date}
					{#if day.isToday}
						<Badge variant="success" size="sm" class="ml-2">Today</Badge>
					{/if}
				{/if}
			{/each}
		</h2>

		{#if selectedEntries.length === 0}
			<Card class="mt-4">
				<div class="py-8 text-center">
					<Calendar class="mx-auto h-10 w-10 text-gray-300 dark:text-gray-600" />
					<p class="mt-3 text-sm text-gray-500 dark:text-gray-400">No appointments scheduled for this day.</p>
				</div>
			</Card>
		{:else}
			<div class="mt-4 space-y-3">
				{#each selectedEntries as entry}
					<div class="rounded-xl border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-800">
						<div class="flex items-start gap-3">
							<!-- Time Column -->
							<div class="flex flex-col items-center pt-0.5">
								<span class="text-sm font-semibold text-gray-900 dark:text-white">{entry.time}</span>
								<span class="mt-0.5 text-xs text-gray-400">{entry.duration}min</span>
							</div>

							<!-- Vertical line + icon -->
							<div class="flex flex-col items-center">
								<div class="flex h-10 w-10 items-center justify-center rounded-lg
									{entry.type === 'job'
										? 'bg-primary-100 text-primary-600 dark:bg-primary-900/30 dark:text-primary-400'
										: 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400'}">
									{#if entry.type === 'job'}
										<Briefcase class="h-5 w-5" />
									{:else}
										<MapPin class="h-5 w-5" />
									{/if}
								</div>
							</div>

							<!-- Content -->
							<div class="flex-1 min-w-0">
								<div class="flex items-center gap-2">
									<h3 class="text-sm font-medium text-gray-900 dark:text-white truncate">{entry.title}</h3>
									<Badge variant={statusColors[entry.status] || 'info'} size="sm">{entry.status}</Badge>
								</div>
								<div class="mt-1 flex flex-wrap items-center gap-3 text-xs text-gray-500 dark:text-gray-400">
									<span class="flex items-center gap-1">
										<User class="h-3 w-3" />
										{entry.customer}
									</span>
									<span class="flex items-center gap-1">
										<MapPin class="h-3 w-3" />
										{entry.location}
									</span>
									<span class="flex items-center gap-1">
										<Clock class="h-3 w-3" />
										{entry.duration} minutes
									</span>
								</div>
							</div>

							<!-- Action -->
							{#if entry.status === 'scheduled'}
								<Button variant="outline" size="sm">Start</Button>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Legend -->
	<div class="mt-6 flex items-center gap-6 text-xs text-gray-500 dark:text-gray-400">
		<div class="flex items-center gap-2">
			<div class="h-3 w-3 rounded bg-primary-500"></div>
			<span>One-off Job</span>
		</div>
		<div class="flex items-center gap-2">
			<div class="h-3 w-3 rounded bg-blue-500"></div>
			<span>Route Visit</span>
		</div>
	</div>
</div>
{/if}
