<script lang="ts">
	import { onMount } from 'svelte';
	import {
		MessageSquare,
		Phone,
		BarChart3,
		Settings,
		Clock,
		CheckCircle,
		XCircle,
		ArrowRight,
		Loader2,
		RefreshCw
	} from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import api from '$lib/api/client';

	let loading = $state(true);
	let error = $state('');
	let activeTab = $state<'overview' | 'sms' | 'ivr' | 'templates' | 'config'>('overview');

	// Stats
	let stats = $state({
		totalSMS: 0,
		smsToday: 0,
		totalCalls: 0,
		callsToday: 0,
		avgCallDuration: 0,
		successRate: 0,
		activeConversations: 0,
		topCategory: ''
	});

	// SMS conversations
	let smsConversations = $state<
		{
			phone: string;
			lastMessage: string;
			lastReply: string;
			timestamp: string;
			state: string;
			messageCount: number;
		}[]
	>([]);

	// IVR call logs
	let callLogs = $state<
		{
			callSid: string;
			from: string;
			duration: string;
			status: string;
			language: string;
			category: string;
			timestamp: string;
		}[]
	>([]);

	// SMS templates
	let templates = $state<
		{
			key: string;
			en: string;
			hi: string;
			kn: string;
		}[]
	>([
		{
			key: 'help',
			en: 'Seva SMS Commands:\nFIND <service> - Search providers\nBOOK <number> - Book a provider\nSTATUS - Check active jobs\nCANCEL <id> - Cancel a job\nRATE <1-5> - Rate last job\nHELP - Show this message',
			hi: 'Seva SMS:\nFIND <seva> - provider khojein\nBOOK <number> - booking karein\nSTATUS - active jobs dekhein\nCANCEL <id> - cancel karein\nRATE <1-5> - rating dein\nHELP - yeh message',
			kn: 'Seva SMS:\nFIND <seva> - provider hudukiri\nBOOK <number> - booking maadi\nSTATUS - active jobs nodi\nCANCEL <id> - cancel maadi\nRATE <1-5> - rating kodi\nHELP - ee message'
		},
		{
			key: 'no_results',
			en: "No providers found for '%s'. Try a different search term.",
			hi: "'%s' ke liye koi provider nahi mila. Kuch aur khojein.",
			kn: "'%s' ge yaaru provider sigalilla. Bere hesharu prayatnisi."
		},
		{
			key: 'booking_started',
			en: 'Booking with %s. They will contact you shortly.',
			hi: '%s ke saath booking ho gayi. Woh jaldi sampark karenge.',
			kn: '%s jote booking aayitu. Avaru bega samparki suttaare.'
		},
		{
			key: 'rating_confirmed',
			en: 'Thank you! You rated the service %d/5.',
			hi: 'Dhanyavaad! Aapne %d/5 rating di.',
			kn: 'Dhanyavaadagalu! Neevu %d/5 rating kottiddeera.'
		}
	]);

	// Configuration
	let config = $state({
		smsEnabled: true,
		ivrEnabled: true,
		defaultLanguage: 'en',
		maxSessionTTL: 30,
		twilioConfigured: false
	});

	let editingTemplate = $state<string | null>(null);

	onMount(async () => {
		try {
			// In production, fetch real data from the admin API.
			// For now, populate with sample data for the UI.
			stats = {
				totalSMS: 1247,
				smsToday: 43,
				totalCalls: 389,
				callsToday: 12,
				avgCallDuration: 145,
				successRate: 87,
				activeConversations: 8,
				topCategory: 'Coconut climbing'
			};

			smsConversations = [
				{
					phone: '+919876543210',
					lastMessage: 'FIND plumber',
					lastReply: 'Found 3 providers for plumber...',
					timestamp: '2 min ago',
					state: 'searching',
					messageCount: 4
				},
				{
					phone: '+919123456789',
					lastMessage: 'BOOK 2',
					lastReply: 'Booking with Rajesh Kumar. They will contact you shortly.',
					timestamp: '15 min ago',
					state: 'booking',
					messageCount: 6
				},
				{
					phone: '+918765432109',
					lastMessage: 'STATUS',
					lastReply: 'You have no active jobs.',
					timestamp: '1 hour ago',
					state: 'idle',
					messageCount: 2
				},
				{
					phone: '+917654321098',
					lastMessage: 'RATE 5',
					lastReply: 'Thank you! You rated the service 5/5.',
					timestamp: '2 hours ago',
					state: 'idle',
					messageCount: 8
				}
			];

			callLogs = [
				{
					callSid: 'CA123abc',
					from: '+919876543210',
					duration: '2:34',
					status: 'completed',
					language: 'English',
					category: 'Plumbing',
					timestamp: '10 min ago'
				},
				{
					callSid: 'CA456def',
					from: '+919123456789',
					duration: '1:15',
					status: 'completed',
					language: 'Hindi',
					category: 'Coconut climbing',
					timestamp: '30 min ago'
				},
				{
					callSid: 'CA789ghi',
					from: '+918765432109',
					duration: '0:45',
					status: 'no-answer',
					language: 'Kannada',
					category: '-',
					timestamp: '1 hour ago'
				}
			];

			config.twilioConfigured = true;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load SMS/IVR data';
		} finally {
			loading = false;
		}
	});

	function formatDuration(seconds: number): string {
		const min = Math.floor(seconds / 60);
		const sec = seconds % 60;
		return `${min}m ${sec}s`;
	}

	const stateBadge: Record<string, 'info' | 'warning' | 'success'> = {
		idle: 'info',
		searching: 'warning',
		booking: 'success',
		rating: 'info'
	};

	const statusBadge: Record<string, 'success' | 'danger' | 'warning' | 'info'> = {
		completed: 'success',
		'no-answer': 'danger',
		busy: 'warning',
		'in-progress': 'info'
	};
</script>

<svelte:head>
	<title>SMS/IVR Management - Seva Admin</title>
</svelte:head>

{#if loading}
	<div class="flex items-center justify-center py-20">
		<Loader2 class="h-8 w-8 animate-spin text-primary-600" />
	</div>
{:else if error}
	<div class="px-6 py-8 lg:px-8">
		<div
			class="rounded-lg border border-red-200 bg-red-50 p-6 text-center dark:border-red-800 dark:bg-red-900/20"
		>
			<p class="text-red-600 dark:text-red-400">{error}</p>
		</div>
	</div>
{:else}
	<div class="px-6 py-8 lg:px-8">
		<!-- Header -->
		<div class="flex flex-wrap items-center justify-between gap-4">
			<div>
				<h1 class="text-2xl font-bold text-gray-900 dark:text-white">SMS/IVR Management</h1>
				<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
					Manage basic phone interfaces for SMS and voice calls.
				</p>
			</div>
			<div class="flex items-center gap-3">
				{#if !config.twilioConfigured}
					<Badge variant="danger">Twilio not configured</Badge>
				{:else}
					<Badge variant="success">Twilio connected</Badge>
				{/if}
			</div>
		</div>

		<!-- Tabs -->
		<div class="mt-6 border-b border-gray-200 dark:border-gray-700">
			<nav class="-mb-px flex gap-6">
				{#each [
					{ id: 'overview', label: 'Overview', icon: BarChart3 },
					{ id: 'sms', label: 'SMS Conversations', icon: MessageSquare },
					{ id: 'ivr', label: 'IVR Call Logs', icon: Phone },
					{ id: 'templates', label: 'Templates', icon: MessageSquare },
					{ id: 'config', label: 'Configuration', icon: Settings }
				] as tab}
					{@const Icon = tab.icon}
					<button
						onclick={() => (activeTab = tab.id as typeof activeTab)}
						class="flex items-center gap-2 border-b-2 px-1 py-3 text-sm font-medium transition
							{activeTab === tab.id
							? 'border-primary-500 text-primary-600 dark:text-primary-400'
							: 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'}"
					>
						<Icon class="h-4 w-4" />
						{tab.label}
					</button>
				{/each}
			</nav>
		</div>

		<!-- Overview Tab -->
		{#if activeTab === 'overview'}
			<div class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
				<Card>
					<div class="flex items-center gap-3">
						<div
							class="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100 text-blue-600 dark:bg-blue-900/30"
						>
							<MessageSquare class="h-5 w-5" />
						</div>
						<div>
							<p class="text-sm text-gray-600 dark:text-gray-400">Total SMS</p>
							<p class="text-2xl font-bold text-gray-900 dark:text-white">
								{stats.totalSMS.toLocaleString()}
							</p>
						</div>
					</div>
					<p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
						{stats.smsToday} today
					</p>
				</Card>

				<Card>
					<div class="flex items-center gap-3">
						<div
							class="flex h-10 w-10 items-center justify-center rounded-lg bg-green-100 text-green-600 dark:bg-green-900/30"
						>
							<Phone class="h-5 w-5" />
						</div>
						<div>
							<p class="text-sm text-gray-600 dark:text-gray-400">Total Calls</p>
							<p class="text-2xl font-bold text-gray-900 dark:text-white">
								{stats.totalCalls.toLocaleString()}
							</p>
						</div>
					</div>
					<p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
						{stats.callsToday} today
					</p>
				</Card>

				<Card>
					<div class="flex items-center gap-3">
						<div
							class="flex h-10 w-10 items-center justify-center rounded-lg bg-purple-100 text-purple-600 dark:bg-purple-900/30"
						>
							<Clock class="h-5 w-5" />
						</div>
						<div>
							<p class="text-sm text-gray-600 dark:text-gray-400">Avg Call Duration</p>
							<p class="text-2xl font-bold text-gray-900 dark:text-white">
								{formatDuration(stats.avgCallDuration)}
							</p>
						</div>
					</div>
				</Card>

				<Card>
					<div class="flex items-center gap-3">
						<div
							class="flex h-10 w-10 items-center justify-center rounded-lg bg-secondary-100 text-secondary-600 dark:bg-secondary-900/30"
						>
							<CheckCircle class="h-5 w-5" />
						</div>
						<div>
							<p class="text-sm text-gray-600 dark:text-gray-400">Success Rate</p>
							<p class="text-2xl font-bold text-gray-900 dark:text-white">
								{stats.successRate}%
							</p>
						</div>
					</div>
					<p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
						{stats.activeConversations} active conversations
					</p>
				</Card>
			</div>

			<!-- Recent Activity -->
			<div class="mt-8 grid gap-6 lg:grid-cols-2">
				<Card>
					<div class="flex items-center justify-between">
						<h2 class="text-lg font-semibold text-gray-900 dark:text-white">
							Recent SMS Conversations
						</h2>
						<button
							onclick={() => (activeTab = 'sms')}
							class="flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700"
						>
							View all <ArrowRight class="h-4 w-4" />
						</button>
					</div>
					<div class="mt-4 space-y-3">
						{#each smsConversations.slice(0, 3) as conv}
							<div
								class="rounded-lg border border-gray-100 p-3 dark:border-gray-700"
							>
								<div class="flex items-center justify-between">
									<span class="text-sm font-medium text-gray-900 dark:text-white"
										>{conv.phone}</span
									>
									<Badge variant={stateBadge[conv.state] || 'info'} size="sm"
										>{conv.state}</Badge
									>
								</div>
								<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
									User: {conv.lastMessage}
								</p>
								<p class="text-xs text-gray-400 dark:text-gray-500">
									{conv.timestamp} -- {conv.messageCount} messages
								</p>
							</div>
						{/each}
					</div>
				</Card>

				<Card>
					<div class="flex items-center justify-between">
						<h2 class="text-lg font-semibold text-gray-900 dark:text-white">
							Recent IVR Calls
						</h2>
						<button
							onclick={() => (activeTab = 'ivr')}
							class="flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700"
						>
							View all <ArrowRight class="h-4 w-4" />
						</button>
					</div>
					<div class="mt-4 space-y-3">
						{#each callLogs.slice(0, 3) as call}
							<div
								class="rounded-lg border border-gray-100 p-3 dark:border-gray-700"
							>
								<div class="flex items-center justify-between">
									<span class="text-sm font-medium text-gray-900 dark:text-white"
										>{call.from}</span
									>
									<Badge variant={statusBadge[call.status] || 'info'} size="sm"
										>{call.status}</Badge
									>
								</div>
								<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
									{call.language} | {call.category} | {call.duration}
								</p>
								<p class="text-xs text-gray-400 dark:text-gray-500">
									{call.timestamp}
								</p>
							</div>
						{/each}
					</div>
				</Card>
			</div>
		{/if}

		<!-- SMS Tab -->
		{#if activeTab === 'sms'}
			<div class="mt-6">
				<Card>
					<h2 class="text-lg font-semibold text-gray-900 dark:text-white">
						SMS Conversations
					</h2>
					<div class="mt-4 overflow-x-auto">
						<table class="w-full text-left text-sm">
							<thead>
								<tr class="border-b border-gray-200 dark:border-gray-700">
									<th class="px-4 py-3 font-medium text-gray-600 dark:text-gray-400"
										>Phone</th
									>
									<th class="px-4 py-3 font-medium text-gray-600 dark:text-gray-400"
										>Last Message</th
									>
									<th class="px-4 py-3 font-medium text-gray-600 dark:text-gray-400"
										>Reply</th
									>
									<th class="px-4 py-3 font-medium text-gray-600 dark:text-gray-400"
										>State</th
									>
									<th class="px-4 py-3 font-medium text-gray-600 dark:text-gray-400"
										>Messages</th
									>
									<th class="px-4 py-3 font-medium text-gray-600 dark:text-gray-400"
										>Time</th
									>
								</tr>
							</thead>
							<tbody>
								{#each smsConversations as conv}
									<tr class="border-b border-gray-100 dark:border-gray-800">
										<td
											class="px-4 py-3 font-medium text-gray-900 dark:text-white"
											>{conv.phone}</td
										>
										<td class="px-4 py-3 text-gray-600 dark:text-gray-400"
											>{conv.lastMessage}</td
										>
										<td
											class="max-w-xs truncate px-4 py-3 text-gray-600 dark:text-gray-400"
											>{conv.lastReply}</td
										>
										<td class="px-4 py-3">
											<Badge variant={stateBadge[conv.state] || 'info'} size="sm"
												>{conv.state}</Badge
											>
										</td>
										<td class="px-4 py-3 text-gray-600 dark:text-gray-400"
											>{conv.messageCount}</td
										>
										<td class="px-4 py-3 text-gray-500 dark:text-gray-400"
											>{conv.timestamp}</td
										>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</Card>
			</div>
		{/if}

		<!-- IVR Tab -->
		{#if activeTab === 'ivr'}
			<div class="mt-6">
				<Card>
					<h2 class="text-lg font-semibold text-gray-900 dark:text-white">
						IVR Call Logs
					</h2>
					<div class="mt-4 overflow-x-auto">
						<table class="w-full text-left text-sm">
							<thead>
								<tr class="border-b border-gray-200 dark:border-gray-700">
									<th class="px-4 py-3 font-medium text-gray-600 dark:text-gray-400"
										>Call SID</th
									>
									<th class="px-4 py-3 font-medium text-gray-600 dark:text-gray-400"
										>From</th
									>
									<th class="px-4 py-3 font-medium text-gray-600 dark:text-gray-400"
										>Language</th
									>
									<th class="px-4 py-3 font-medium text-gray-600 dark:text-gray-400"
										>Category</th
									>
									<th class="px-4 py-3 font-medium text-gray-600 dark:text-gray-400"
										>Duration</th
									>
									<th class="px-4 py-3 font-medium text-gray-600 dark:text-gray-400"
										>Status</th
									>
									<th class="px-4 py-3 font-medium text-gray-600 dark:text-gray-400"
										>Time</th
									>
								</tr>
							</thead>
							<tbody>
								{#each callLogs as call}
									<tr class="border-b border-gray-100 dark:border-gray-800">
										<td
											class="px-4 py-3 font-mono text-xs text-gray-600 dark:text-gray-400"
											>{call.callSid}</td
										>
										<td
											class="px-4 py-3 font-medium text-gray-900 dark:text-white"
											>{call.from}</td
										>
										<td class="px-4 py-3 text-gray-600 dark:text-gray-400"
											>{call.language}</td
										>
										<td class="px-4 py-3 text-gray-600 dark:text-gray-400"
											>{call.category}</td
										>
										<td class="px-4 py-3 text-gray-600 dark:text-gray-400"
											>{call.duration}</td
										>
										<td class="px-4 py-3">
											<Badge
												variant={statusBadge[call.status] || 'info'}
												size="sm">{call.status}</Badge
											>
										</td>
										<td class="px-4 py-3 text-gray-500 dark:text-gray-400"
											>{call.timestamp}</td
										>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</Card>
			</div>
		{/if}

		<!-- Templates Tab -->
		{#if activeTab === 'templates'}
			<div class="mt-6 space-y-4">
				{#each templates as template}
					<Card>
						<div class="flex items-center justify-between">
							<h3 class="text-sm font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
								{template.key}
							</h3>
							<button
								onclick={() =>
									(editingTemplate =
										editingTemplate === template.key ? null : template.key)}
								class="text-sm text-primary-600 hover:text-primary-700"
							>
								{editingTemplate === template.key ? 'Close' : 'Edit'}
							</button>
						</div>
						<div class="mt-3 grid gap-4 md:grid-cols-3">
							<div>
								<p class="mb-1 text-xs font-medium text-gray-500 dark:text-gray-400">
									English
								</p>
								{#if editingTemplate === template.key}
									<textarea
										class="w-full rounded-lg border border-gray-300 p-2 text-sm dark:border-gray-600 dark:bg-gray-800 dark:text-white"
										rows="4"
										bind:value={template.en}
									></textarea>
								{:else}
									<p class="whitespace-pre-wrap text-sm text-gray-700 dark:text-gray-300">
										{template.en}
									</p>
								{/if}
							</div>
							<div>
								<p class="mb-1 text-xs font-medium text-gray-500 dark:text-gray-400">
									Hindi
								</p>
								{#if editingTemplate === template.key}
									<textarea
										class="w-full rounded-lg border border-gray-300 p-2 text-sm dark:border-gray-600 dark:bg-gray-800 dark:text-white"
										rows="4"
										bind:value={template.hi}
									></textarea>
								{:else}
									<p class="whitespace-pre-wrap text-sm text-gray-700 dark:text-gray-300">
										{template.hi}
									</p>
								{/if}
							</div>
							<div>
								<p class="mb-1 text-xs font-medium text-gray-500 dark:text-gray-400">
									Kannada
								</p>
								{#if editingTemplate === template.key}
									<textarea
										class="w-full rounded-lg border border-gray-300 p-2 text-sm dark:border-gray-600 dark:bg-gray-800 dark:text-white"
										rows="4"
										bind:value={template.kn}
									></textarea>
								{:else}
									<p class="whitespace-pre-wrap text-sm text-gray-700 dark:text-gray-300">
										{template.kn}
									</p>
								{/if}
							</div>
						</div>
						{#if editingTemplate === template.key}
							<div class="mt-3 flex justify-end">
								<Button size="sm">Save Template</Button>
							</div>
						{/if}
					</Card>
				{/each}
			</div>
		{/if}

		<!-- Configuration Tab -->
		{#if activeTab === 'config'}
			<div class="mt-6 max-w-2xl space-y-6">
				<Card>
					<h2 class="text-lg font-semibold text-gray-900 dark:text-white">
						SMS/IVR Configuration
					</h2>
					<div class="mt-6 space-y-6">
						<!-- Enable/Disable SMS -->
						<div class="flex items-center justify-between">
							<div>
								<p class="font-medium text-gray-900 dark:text-white">
									SMS Interface
								</p>
								<p class="text-sm text-gray-500 dark:text-gray-400">
									Allow users to interact via SMS commands.
								</p>
							</div>
							<label class="relative inline-flex cursor-pointer items-center">
								<input
									type="checkbox"
									bind:checked={config.smsEnabled}
									class="peer sr-only"
								/>
								<div
									class="peer h-6 w-11 rounded-full bg-gray-200 after:absolute after:start-[2px] after:top-[2px] after:h-5 after:w-5 after:rounded-full after:border after:border-gray-300 after:bg-white after:transition-all after:content-[''] peer-checked:bg-primary-600 peer-checked:after:translate-x-full peer-focus:outline-none dark:bg-gray-700"
								></div>
							</label>
						</div>

						<!-- Enable/Disable IVR -->
						<div class="flex items-center justify-between">
							<div>
								<p class="font-medium text-gray-900 dark:text-white">
									IVR (Voice) Interface
								</p>
								<p class="text-sm text-gray-500 dark:text-gray-400">
									Allow users to interact via phone calls with DTMF.
								</p>
							</div>
							<label class="relative inline-flex cursor-pointer items-center">
								<input
									type="checkbox"
									bind:checked={config.ivrEnabled}
									class="peer sr-only"
								/>
								<div
									class="peer h-6 w-11 rounded-full bg-gray-200 after:absolute after:start-[2px] after:top-[2px] after:h-5 after:w-5 after:rounded-full after:border after:border-gray-300 after:bg-white after:transition-all after:content-[''] peer-checked:bg-primary-600 peer-checked:after:translate-x-full peer-focus:outline-none dark:bg-gray-700"
								></div>
							</label>
						</div>

						<!-- Default Language -->
						<div>
							<label
								for="defaultLang"
								class="block font-medium text-gray-900 dark:text-white"
							>
								Default Language
							</label>
							<p class="text-sm text-gray-500 dark:text-gray-400">
								Language used when user preference is unknown.
							</p>
							<select
								id="defaultLang"
								bind:value={config.defaultLanguage}
								class="mt-2 w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm dark:border-gray-600 dark:bg-gray-800 dark:text-white"
							>
								<option value="en">English</option>
								<option value="hi">Hindi</option>
								<option value="kn">Kannada</option>
								<option value="ta">Tamil</option>
								<option value="te">Telugu</option>
								<option value="ml">Malayalam</option>
							</select>
						</div>

						<!-- Session TTL -->
						<div>
							<label
								for="sessionTTL"
								class="block font-medium text-gray-900 dark:text-white"
							>
								Session Timeout (minutes)
							</label>
							<p class="text-sm text-gray-500 dark:text-gray-400">
								How long an SMS conversation session stays active.
							</p>
							<input
								id="sessionTTL"
								type="number"
								bind:value={config.maxSessionTTL}
								min="5"
								max="120"
								class="mt-2 w-32 rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm dark:border-gray-600 dark:bg-gray-800 dark:text-white"
							/>
						</div>

						<div class="flex justify-end pt-4">
							<Button>Save Configuration</Button>
						</div>
					</div>
				</Card>
			</div>
		{/if}
	</div>
{/if}
