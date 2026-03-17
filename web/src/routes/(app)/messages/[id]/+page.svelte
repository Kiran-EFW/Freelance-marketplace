<script lang="ts">
	import { page } from '$app/stores';
	import { ArrowLeft, Send, Loader2 } from 'lucide-svelte';
	import { t } from '$lib/i18n/index.svelte';
	import { messages as messagesApi } from '$lib/api/client';
	import { subscribe as authSubscribe, type AuthState } from '$lib/stores/auth';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import type { Conversation, ChatMessage } from '$lib/types';

	let conversationId = $derived($page.params.id);

	let conversation = $state<Conversation | null>(null);
	let chatMessages = $state<ChatMessage[]>([]);
	let loading = $state(true);
	let sending = $state(false);
	let error = $state<string | null>(null);
	let messageInput = $state('');
	let messagesContainer = $state<HTMLDivElement | null>(null);
	let pollInterval = $state<ReturnType<typeof setInterval> | null>(null);
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
		if (authState.user && conversationId) {
			loadConversation();
			startPolling();
		}

		return () => {
			stopPolling();
		};
	});

	function startPolling() {
		stopPolling();
		pollInterval = setInterval(() => {
			if (!loading && !sending) {
				pollMessages();
			}
		}, 5000);
	}

	function stopPolling() {
		if (pollInterval) {
			clearInterval(pollInterval);
			pollInterval = null;
		}
	}

	async function loadConversation() {
		loading = true;
		error = null;
		try {
			const response = await messagesApi.getConversation(conversationId);
			conversation = response.data.conversation;
			chatMessages = (response.data.messages ?? []).reverse();
			scrollToBottom();

			// Mark messages as read
			await messagesApi.markRead(conversationId);
		} catch (err) {
			error = t('common.error');
			console.error('Failed to load conversation:', err);
		} finally {
			loading = false;
		}
	}

	async function pollMessages() {
		try {
			const response = await messagesApi.getConversation(conversationId);
			const newMessages = (response.data.messages ?? []).reverse();

			// Only update if there are new messages
			if (newMessages.length !== chatMessages.length) {
				chatMessages = newMessages;
				scrollToBottom();
				await messagesApi.markRead(conversationId);
			}
		} catch {
			// Silent fail for polling
		}
	}

	async function sendMessage() {
		const content = messageInput.trim();
		if (!content || sending) return;

		sending = true;
		try {
			const response = await messagesApi.sendMessage(conversationId, content);
			chatMessages = [...chatMessages, response.data];
			messageInput = '';
			scrollToBottom();
		} catch (err) {
			console.error('Failed to send message:', err);
		} finally {
			sending = false;
		}
	}

	function handleKeyDown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			sendMessage();
		}
	}

	function scrollToBottom() {
		setTimeout(() => {
			if (messagesContainer) {
				messagesContainer.scrollTop = messagesContainer.scrollHeight;
			}
		}, 50);
	}

	function getOtherUserName(): string {
		if (!conversation || !authState.user) return 'User';
		if (conversation.other_user?.name) return conversation.other_user.name;
		const otherId = conversation.participant_1 === authState.user.id
			? conversation.participant_2
			: conversation.participant_1;
		return otherId.substring(0, 8);
	}

	function isSentByMe(msg: ChatMessage): boolean {
		return authState.user ? msg.sender_id === authState.user.id : false;
	}

	function shouldShowTimestamp(index: number): boolean {
		if (index === 0) return true;
		const curr = new Date(chatMessages[index].created_at);
		const prev = new Date(chatMessages[index - 1].created_at);
		// Show timestamp if more than 30 minutes between messages
		return curr.getTime() - prev.getTime() > 30 * 60 * 1000;
	}

	function formatMessageTime(dateStr: string): string {
		const date = new Date(dateStr);
		return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
	}

	function formatDateSeparator(dateStr: string): string {
		const date = new Date(dateStr);
		const now = new Date();
		const diff = now.getTime() - date.getTime();
		const dayMs = 86400000;

		if (diff < dayMs && date.getDate() === now.getDate()) {
			return t('messages.today');
		}

		const yesterday = new Date(now);
		yesterday.setDate(yesterday.getDate() - 1);
		if (date.getDate() === yesterday.getDate() && date.getMonth() === yesterday.getMonth()) {
			return t('messages.yesterday');
		}

		return date.toLocaleDateString([], { weekday: 'long', month: 'short', day: 'numeric' });
	}
</script>

<svelte:head>
	<title>{t('messages.conversation_with')} {getOtherUserName()} - Seva</title>
</svelte:head>

<div class="mx-auto flex h-[calc(100vh-4rem)] max-w-3xl flex-col px-0 sm:px-6 lg:px-8 sm:py-4">
	<!-- Header -->
	<div class="flex items-center gap-3 border-b border-gray-200 bg-white px-4 py-3 dark:border-gray-700 dark:bg-gray-800 sm:rounded-t-lg">
		<a href="/messages" class="rounded-lg p-1 text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700">
			<ArrowLeft class="h-5 w-5" />
		</a>
		<Avatar name={getOtherUserName()} size="sm" />
		<div class="flex-1">
			<h2 class="text-sm font-semibold text-gray-900 dark:text-white">{getOtherUserName()}</h2>
		</div>
	</div>

	<!-- Messages Area -->
	{#if loading}
		<div class="flex flex-1 flex-col items-center justify-center bg-gray-50 dark:bg-gray-900">
			<Loader2 class="h-8 w-8 animate-spin text-primary-600" />
			<p class="mt-2 text-sm text-gray-500 dark:text-gray-400">{t('messages.loading')}</p>
		</div>
	{:else if error}
		<div class="flex flex-1 flex-col items-center justify-center bg-gray-50 dark:bg-gray-900">
			<p class="text-sm text-red-600 dark:text-red-400">{error}</p>
			<button
				onclick={loadConversation}
				class="mt-2 text-sm font-medium text-primary-600 hover:text-primary-700"
			>
				{t('common.retry')}
			</button>
		</div>
	{:else}
		<div
			bind:this={messagesContainer}
			class="flex-1 overflow-y-auto bg-gray-50 p-4 dark:bg-gray-900"
		>
			{#if chatMessages.length === 0}
				<div class="flex h-full flex-col items-center justify-center">
					<p class="text-sm text-gray-500 dark:text-gray-400">{t('messages.no_messages')}</p>
				</div>
			{:else}
				<div class="space-y-1">
					{#each chatMessages as msg, index}
						<!-- Date separator -->
						{#if shouldShowTimestamp(index)}
							<div class="flex items-center justify-center py-3">
								<span class="rounded-full bg-gray-200 px-3 py-1 text-xs text-gray-600 dark:bg-gray-700 dark:text-gray-400">
									{formatDateSeparator(msg.created_at)}
								</span>
							</div>
						{/if}

						<!-- Message bubble -->
						<div class="flex {isSentByMe(msg) ? 'justify-end' : 'justify-start'}">
							<div class="flex max-w-[75%] flex-col {isSentByMe(msg) ? 'items-end' : 'items-start'}">
								<div
									class="rounded-2xl px-4 py-2 {isSentByMe(msg)
										? 'rounded-br-md bg-primary-600 text-white'
										: 'rounded-bl-md bg-white text-gray-900 shadow-sm dark:bg-gray-800 dark:text-white'}"
								>
									<p class="whitespace-pre-wrap break-words text-sm">{msg.content}</p>
								</div>
								<span class="mt-0.5 px-1 text-[10px] text-gray-400 dark:text-gray-500">
									{formatMessageTime(msg.created_at)}
									{#if isSentByMe(msg) && msg.is_read}
										<span class="ml-1">Read</span>
									{/if}
								</span>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	{/if}

	<!-- Input Area -->
	<div class="border-t border-gray-200 bg-white px-4 py-3 dark:border-gray-700 dark:bg-gray-800 sm:rounded-b-lg">
		<div class="flex items-end gap-2">
			<div class="relative flex-1">
				<textarea
					bind:value={messageInput}
					onkeydown={handleKeyDown}
					placeholder={t('messages.type_message')}
					rows="1"
					class="block w-full resize-none rounded-xl border border-gray-300 bg-gray-50 px-4 py-2.5 text-sm text-gray-900 placeholder-gray-400 focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-500 dark:focus:border-primary-500"
					style="max-height: 120px;"
				></textarea>
			</div>
			<button
				onclick={sendMessage}
				disabled={!messageInput.trim() || sending}
				class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-primary-600 text-white transition-colors hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
			>
				{#if sending}
					<Loader2 class="h-4 w-4 animate-spin" />
				{:else}
					<Send class="h-4 w-4" />
				{/if}
			</button>
		</div>
	</div>
</div>
