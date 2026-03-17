<script lang="ts">
	import { MessageSquare, Plus, Loader2 } from 'lucide-svelte';
	import { t } from '$lib/i18n/index.svelte';
	import { messages as messagesApi } from '$lib/api/client';
	import { subscribe as authSubscribe, type AuthState } from '$lib/stores/auth';
	import { wsManager, type WSNewMessagePayload } from '$lib/api/websocket';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import type { Conversation } from '$lib/types';

	let conversations = $state<Conversation[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let onlineUsers = $state<Set<string>>(new Set());
	let authState = $state<AuthState>({
		user: null,
		loading: false,
		initialized: false,
		notificationCount: 0,
		pointsBalance: 0
	});

	$effect(() => {
		const unsub = authSubscribe((state) => {
			authState = state;
		});
		return unsub;
	});

	$effect(() => {
		if (authState.user) {
			loadConversations();

			// Subscribe to WebSocket new_message events to update conversation list
			const unsubNewMsg = wsManager.on<WSNewMessagePayload>('new_message', (payload) => {
				// Update conversation list when a new message arrives
				const convIdx = conversations.findIndex((c) => c.id === payload.conversation_id);
				if (convIdx >= 0) {
					const conv = { ...conversations[convIdx] };
					conv.last_message_at = payload.created_at;
					conv.last_message_preview = payload.content.length > 100
						? payload.content.substring(0, 100) + '...'
						: payload.content;
					conv.unread_count = (conv.unread_count || 0) + 1;
					// Move to top
					conversations = [conv, ...conversations.filter((_, i) => i !== convIdx)];
				} else {
					// New conversation, reload
					loadConversations();
				}
			});

			// Subscribe to online status
			const unsubOnline = wsManager.subscribeOnline((users) => {
				onlineUsers = users;
			});

			return () => {
				unsubNewMsg();
				unsubOnline();
			};
		}
	});

	async function loadConversations() {
		loading = true;
		error = null;
		try {
			const response = await messagesApi.listConversations();
			conversations = response.data ?? [];
		} catch (err) {
			error = t('common.error');
			console.error('Failed to load conversations:', err);
		} finally {
			loading = false;
		}
	}

	function getOtherParticipantId(conv: Conversation): string {
		if (!authState.user) return '';
		return conv.participant_1 === authState.user.id ? conv.participant_2 : conv.participant_1;
	}

	function isOtherUserOnline(conv: Conversation): boolean {
		const otherId = getOtherParticipantId(conv);
		return onlineUsers.has(otherId);
	}

	function formatTime(dateStr?: string): string {
		if (!dateStr) return '';
		const date = new Date(dateStr);
		const now = new Date();
		const diff = now.getTime() - date.getTime();
		const dayMs = 86400000;

		if (diff < dayMs && date.getDate() === now.getDate()) {
			return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
		}

		const yesterday = new Date(now);
		yesterday.setDate(yesterday.getDate() - 1);
		if (date.getDate() === yesterday.getDate() && date.getMonth() === yesterday.getMonth()) {
			return t('messages.yesterday');
		}

		return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
	}
</script>

<svelte:head>
	<title>{t('messages.title')} - Seva</title>
</svelte:head>

<div class="mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold text-gray-900 dark:text-white">{t('messages.title')}</h1>
		<a
			href="/providers"
			class="inline-flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white hover:bg-primary-700"
		>
			<Plus class="h-4 w-4" />
			{t('messages.new_message')}
		</a>
	</div>

	<!-- Content -->
	<div class="mt-6">
		{#if loading}
			<div class="flex flex-col items-center justify-center py-16">
				<Loader2 class="h-8 w-8 animate-spin text-primary-600" />
				<p class="mt-2 text-sm text-gray-500 dark:text-gray-400">{t('messages.loading')}</p>
			</div>
		{:else if error}
			<div class="rounded-lg border border-red-200 bg-red-50 p-6 text-center dark:border-red-800 dark:bg-red-900/20">
				<p class="text-sm text-red-600 dark:text-red-400">{error}</p>
				<button
					onclick={loadConversations}
					class="mt-2 text-sm font-medium text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300"
				>
					{t('common.retry')}
				</button>
			</div>
		{:else if conversations.length === 0}
			<div class="flex flex-col items-center justify-center rounded-lg border border-gray-200 bg-gray-50 py-16 dark:border-gray-700 dark:bg-gray-800/50">
				<MessageSquare class="h-12 w-12 text-gray-400" />
				<p class="mt-3 text-lg font-medium text-gray-900 dark:text-white">{t('messages.no_conversations')}</p>
				<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{t('messages.start_conversation')}</p>
				<a
					href="/providers"
					class="mt-4 inline-flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white hover:bg-primary-700"
				>
					{t('nav.find_providers')}
				</a>
			</div>
		{:else}
			<div class="divide-y divide-gray-200 rounded-lg border border-gray-200 bg-white dark:divide-gray-700 dark:border-gray-700 dark:bg-gray-800">
				{#each conversations as conv}
					{@const otherId = getOtherParticipantId(conv)}
					<a
						href="/messages/{conv.id}"
						class="flex items-center gap-4 p-4 transition-colors hover:bg-gray-50 dark:hover:bg-gray-700/50"
					>
						<div class="relative flex-shrink-0">
							<Avatar name={conv.other_user?.name || otherId.substring(0, 8)} size="md" />
							{#if isOtherUserOnline(conv)}
								<span class="absolute bottom-0 right-0 h-3 w-3 rounded-full border-2 border-white bg-green-500 dark:border-gray-800"></span>
							{/if}
						</div>
						<div class="min-w-0 flex-1">
							<div class="flex items-center justify-between">
								<p class="truncate text-sm font-semibold text-gray-900 dark:text-white">
									{conv.other_user?.name || 'User'}
								</p>
								<span class="flex-shrink-0 text-xs text-gray-500 dark:text-gray-400">
									{formatTime(conv.last_message_at)}
								</span>
							</div>
							<div class="mt-1 flex items-center justify-between">
								<p class="truncate text-sm text-gray-500 dark:text-gray-400">
									{conv.last_message_preview || t('messages.no_messages')}
								</p>
								{#if conv.unread_count && conv.unread_count > 0}
									<span class="ml-2 flex h-5 min-w-[1.25rem] flex-shrink-0 items-center justify-center rounded-full bg-primary-600 px-1.5 text-[10px] font-bold text-white">
										{conv.unread_count > 99 ? '99+' : conv.unread_count}
									</span>
								{/if}
							</div>
						</div>
					</a>
				{/each}
			</div>
		{/if}
	</div>
</div>
