<script lang="ts">
	import { onMount } from 'svelte';
	import { Bell, Briefcase, MessageSquare, Star, IndianRupee, AlertTriangle, CheckCircle, Award, Loader2 } from 'lucide-svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import { toastSuccess, toastError } from '$lib/stores/toast';
	import api from '$lib/api/client';

	const iconMap: Record<string, any> = {
		job: Briefcase, quote: MessageSquare, review: Star, payment: IndianRupee,
		dispute: AlertTriangle, system: Bell, points: Award, kyc: CheckCircle
	};

	let notifications = $state<any[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			const res = await api.notifications.list({ per_page: 20 });
			notifications = (res.data || []).map((n: any) => ({
				id: n.id,
				type: n.type || 'system',
				title: n.title || '',
				message: n.message || n.body || '',
				read: n.read ?? n.is_read ?? true,
				time: n.created_at || '',
				url: n.url || n.action_url || '/dashboard'
			}));
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load notifications';
		} finally {
			loading = false;
		}
	});

	const unreadCount = $derived(notifications.filter((n) => !n.read).length);

	async function markAllRead() {
		try {
			await api.notifications.markAllRead();
			notifications = notifications.map((n) => ({ ...n, read: true }));
			toastSuccess('All notifications marked as read');
		} catch (err) {
			toastError(err instanceof Error ? err.message : 'Failed to mark all as read');
		}
	}

	async function markRead(id: string) {
		try {
			await api.notifications.markRead(id);
			notifications = notifications.map((n) => n.id === id ? { ...n, read: true } : n);
		} catch {
			// Silently fail for individual mark-read
		}
	}
</script>

<svelte:head>
	<title>Notifications - Seva</title>
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
{/if}
