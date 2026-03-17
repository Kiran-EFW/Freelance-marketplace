<script lang="ts">
	import { Bell, Briefcase, MessageSquare, Star, IndianRupee, AlertTriangle, CheckCircle, Award } from 'lucide-svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import { toastSuccess } from '$lib/stores/toast';

	const iconMap: Record<string, any> = {
		job: Briefcase, quote: MessageSquare, review: Star, payment: IndianRupee,
		dispute: AlertTriangle, system: Bell, points: Award, kyc: CheckCircle
	};

	const mockNotifications = [
		{ id: '1', type: 'quote', title: 'New quote received', message: 'Suresh Nair quoted Rs. 800 for "Fix leaking kitchen tap"', read: false, time: '5 min ago', url: '/jobs/1' },
		{ id: '2', type: 'quote', title: 'New quote received', message: 'Ravi Kumar quoted Rs. 1,200 for "Fix leaking kitchen tap"', read: false, time: '1 hr ago', url: '/jobs/1' },
		{ id: '3', type: 'payment', title: 'Payment received', message: 'Rs. 800 payment confirmed for "Electrical wiring job"', read: false, time: '3 hrs ago', url: '/payments' },
		{ id: '4', type: 'review', title: 'New review', message: 'Priya Menon left a 5-star review on "Tap installation"', read: true, time: '1 day ago', url: '/reviews' },
		{ id: '5', type: 'job', title: 'Job completed', message: 'Your job "Deep cleaning" has been marked as completed', read: true, time: '2 days ago', url: '/jobs/3' },
		{ id: '6', type: 'points', title: 'Points earned', message: 'You earned 50 points for completing a job', read: true, time: '2 days ago', url: '/points' },
		{ id: '7', type: 'system', title: 'Welcome to Seva', message: 'Your account has been verified. Start posting jobs or finding providers!', read: true, time: '1 week ago', url: '/dashboard' },
		{ id: '8', type: 'dispute', title: 'Dispute resolved', message: 'Your dispute for "Painting job" has been resolved in your favor', read: true, time: '1 week ago', url: '/disputes/1' }
	];

	let notifications = $state(mockNotifications);

	const unreadCount = $derived(notifications.filter((n) => !n.read).length);

	function markAllRead() {
		notifications = notifications.map((n) => ({ ...n, read: true }));
		toastSuccess('All notifications marked as read');
	}

	function markRead(id: string) {
		notifications = notifications.map((n) => n.id === id ? { ...n, read: true } : n);
	}
</script>

<svelte:head>
	<title>Notifications - Seva</title>
</svelte:head>

<div class="mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Notifications</h1>
			<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
				{unreadCount > 0 ? `${unreadCount} unread notification${unreadCount > 1 ? 's' : ''}` : 'All caught up!'}
			</p>
		</div>
		{#if unreadCount > 0}
			<Button variant="outline" size="sm" onclick={markAllRead}>Mark all read</Button>
		{/if}
	</div>

	<div class="mt-6 space-y-2">
		{#each notifications as notif}
			{@const Icon = iconMap[notif.type] || Bell}
			<a
				href={notif.url}
				onclick={() => markRead(notif.id)}
				class="flex items-start gap-4 rounded-xl border p-4 transition-colors
					{notif.read
						? 'border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-800'
						: 'border-primary-200 bg-primary-50/50 dark:border-primary-800 dark:bg-primary-900/10'}"
			>
				<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full
					{notif.read ? 'bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400' : 'bg-primary-100 text-primary-600 dark:bg-primary-900/30 dark:text-primary-400'}">
					<Icon class="h-5 w-5" />
				</div>
				<div class="flex-1 min-w-0">
					<div class="flex items-start justify-between gap-2">
						<p class="text-sm font-medium {notif.read ? 'text-gray-700 dark:text-gray-300' : 'text-gray-900 dark:text-white'}">{notif.title}</p>
						<span class="shrink-0 text-xs text-gray-500 dark:text-gray-400">{notif.time}</span>
					</div>
					<p class="mt-0.5 text-sm text-gray-500 dark:text-gray-400">{notif.message}</p>
				</div>
				{#if !notif.read}
					<div class="mt-2 h-2 w-2 shrink-0 rounded-full bg-primary-600"></div>
				{/if}
			</a>
		{/each}
	</div>
</div>
