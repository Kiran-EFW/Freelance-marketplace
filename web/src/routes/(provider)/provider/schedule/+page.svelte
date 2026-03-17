<script lang="ts">
	import { ChevronLeft, ChevronRight, Briefcase, Calendar, MapPin, Clock, User } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	const daysOfWeek = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
	let weekOffset = $state(0);

	const baseDate = new Date('2026-03-16');

	function getWeekDates(offset: number) {
		const dates: { day: string; date: number; month: string; full: string; isToday: boolean }[] = [];
		const today = new Date('2026-03-17');
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
		return `${first.month} ${first.date} - ${last.month} ${last.date}, 2026`;
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

	const scheduleData: ScheduleEntry[] = [
		{ id: '1', type: 'job', title: 'Fix kitchen plumbing', customer: 'Amit Verma', time: '09:00', duration: 60, location: '123 MG Road', postcode: '560001', status: 'completed', date: '2026-03-16' },
		{ id: '2', type: 'route', title: 'Koramangala Route - Stop 1', customer: 'Priya Menon', time: '11:00', duration: 45, location: '45 Church Street', postcode: '560001', status: 'completed', date: '2026-03-16' },
		{ id: '3', type: 'route', title: 'Koramangala Route - Stop 2', customer: 'Arjun Das', time: '12:00', duration: 30, location: '78 Brigade Road', postcode: '560002', status: 'completed', date: '2026-03-16' },
		{ id: '4', type: 'job', title: 'Install water heater', customer: 'Meera Reddy', time: '10:00', duration: 90, location: '12 Hosur Road', postcode: '560002', status: 'scheduled', date: '2026-03-17' },
		{ id: '5', type: 'route', title: 'Koramangala Route - Stop 3', customer: 'Kiran Rao', time: '14:00', duration: 50, location: '56 Sarjapur Road', postcode: '560003', status: 'scheduled', date: '2026-03-17' },
		{ id: '6', type: 'job', title: 'Bathroom pipe repair', customer: 'Deepak Sharma', time: '16:00', duration: 45, location: '89 Whitefield Main', postcode: '560066', status: 'scheduled', date: '2026-03-17' },
		{ id: '7', type: 'job', title: 'AC servicing', customer: 'Anita Gupta', time: '09:30', duration: 60, location: '34 Outer Ring Road', postcode: '560037', status: 'scheduled', date: '2026-03-18' },
		{ id: '8', type: 'route', title: 'Indiranagar Route - Stop 1', customer: 'Suresh Nair', time: '11:30', duration: 40, location: '12 100 Feet Road', postcode: '560038', status: 'scheduled', date: '2026-03-18' },
		{ id: '9', type: 'route', title: 'Indiranagar Route - Stop 2', customer: 'Lakshmi Bai', time: '13:00', duration: 35, location: '67 CMH Road', postcode: '560038', status: 'scheduled', date: '2026-03-18' },
		{ id: '10', type: 'job', title: 'Garden maintenance', customer: 'Rajesh Kumar', time: '08:00', duration: 120, location: '5 JP Nagar', postcode: '560078', status: 'scheduled', date: '2026-03-19' },
		{ id: '11', type: 'route', title: 'Koramangala Route - Stop 1', customer: 'Priya Menon', time: '14:00', duration: 45, location: '45 Church Street', postcode: '560001', status: 'scheduled', date: '2026-03-19' },
		{ id: '12', type: 'job', title: 'Electrical wiring check', customer: 'Vivek Iyer', time: '10:00', duration: 60, location: '23 Bannerghatta Road', postcode: '560076', status: 'scheduled', date: '2026-03-20' },
		{ id: '13', type: 'job', title: 'Deep cleaning service', customer: 'Neha Patil', time: '09:00', duration: 180, location: '8 Koramangala 4th Block', postcode: '560034', status: 'scheduled', date: '2026-03-21' },
	];

	function getEntriesForDate(dateStr: string): ScheduleEntry[] {
		return scheduleData.filter((e) => e.date === dateStr).sort((a, b) => a.time.localeCompare(b.time));
	}

	let selectedDate = $state('2026-03-17');
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
