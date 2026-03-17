<script lang="ts">
	import { page } from '$app/stores';
	import { ArrowLeft, AlertTriangle, Clock, CheckCircle, Upload, Send } from 'lucide-svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';

	let disputeId = $derived($page.params.id);
	let newMessage = $state('');

	const dispute = {
		id: '1', jobTitle: 'Painting bedroom', type: 'quality', status: 'open' as const,
		raisedBy: 'Amit Verma', against: 'Kumar Singh',
		description: 'The paint job was uneven and patches were left uncovered. Multiple areas have drip marks and the edges were not properly taped.',
		createdAt: '2026-03-12',
		evidence: [
			{ id: 'e1', description: 'Photo of uneven wall', fileType: 'image', uploadedBy: 'Amit Verma', date: '2026-03-12' },
			{ id: 'e2', description: 'Photo of drip marks on floor', fileType: 'image', uploadedBy: 'Amit Verma', date: '2026-03-12' }
		],
		messages: [
			{ id: 'm1', sender: 'Amit Verma', message: 'The paint job quality is very poor. There are visible patches and drip marks everywhere.', date: '2026-03-12 10:30', isInternal: false },
			{ id: 'm2', sender: 'Kumar Singh', message: 'I apologize for the issues. The paint quality provided was not the best. I am willing to redo the affected areas.', date: '2026-03-12 14:15', isInternal: false },
			{ id: 'm3', sender: 'System', message: 'A mediator has been assigned to this dispute.', date: '2026-03-13 09:00', isInternal: true }
		],
		timeline: [
			{ label: 'Dispute filed', date: '2026-03-12', status: 'completed' },
			{ label: 'Evidence submitted', date: '2026-03-12', status: 'completed' },
			{ label: 'Mediator assigned', date: '2026-03-13', status: 'completed' },
			{ label: 'Under review', date: '', status: 'current' },
			{ label: 'Resolution', date: '', status: 'pending' }
		]
	};
</script>

<svelte:head>
	<title>Dispute #{disputeId} - Seva</title>
</svelte:head>

<div class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
	<a href="/disputes" class="inline-flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700">
		<ArrowLeft class="h-4 w-4" />
		Back to Disputes
	</a>

	<!-- Header -->
	<div class="mt-6 flex items-start justify-between gap-4">
		<div>
			<div class="flex items-center gap-2">
				<h1 class="text-2xl font-bold text-gray-900 dark:text-white">{dispute.jobTitle}</h1>
				<Badge variant="danger">Open</Badge>
			</div>
			<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
				Dispute #{disputeId} -- Filed {dispute.createdAt} -- Type: {dispute.type}
			</p>
		</div>
	</div>

	<div class="mt-6 grid gap-6 lg:grid-cols-3">
		<!-- Main Content -->
		<div class="space-y-6 lg:col-span-2">
			<!-- Description -->
			<Card>
				<h2 class="font-semibold text-gray-900 dark:text-white">Description</h2>
				<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">{dispute.description}</p>
			</Card>

			<!-- Evidence -->
			<Card>
				<h2 class="font-semibold text-gray-900 dark:text-white">Evidence ({dispute.evidence.length})</h2>
				<div class="mt-3 space-y-2">
					{#each dispute.evidence as ev}
						<div class="flex items-center gap-3 rounded-lg bg-gray-50 p-3 dark:bg-gray-700">
							<div class="h-12 w-12 rounded-lg bg-gray-200 dark:bg-gray-600"></div>
							<div class="flex-1">
								<p class="text-sm font-medium text-gray-900 dark:text-white">{ev.description}</p>
								<p class="text-xs text-gray-500 dark:text-gray-400">Uploaded by {ev.uploadedBy} on {ev.date}</p>
							</div>
						</div>
					{/each}
				</div>
				<Button variant="outline" size="sm" class="mt-3">
					<Upload class="h-4 w-4" />
					Add Evidence
				</Button>
			</Card>

			<!-- Messages -->
			<Card>
				<h2 class="font-semibold text-gray-900 dark:text-white">Messages</h2>
				<div class="mt-3 space-y-4">
					{#each dispute.messages as msg}
						{#if msg.isInternal}
							<div class="text-center">
								<span class="rounded-full bg-gray-100 px-3 py-1 text-xs text-gray-500 dark:bg-gray-700 dark:text-gray-400">{msg.message}</span>
							</div>
						{:else}
							<div class="flex items-start gap-3">
								<Avatar name={msg.sender} size="sm" />
								<div class="flex-1 rounded-lg bg-gray-50 p-3 dark:bg-gray-700">
									<div class="flex items-center justify-between">
										<p class="text-sm font-medium text-gray-900 dark:text-white">{msg.sender}</p>
										<span class="text-xs text-gray-400">{msg.date}</span>
									</div>
									<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{msg.message}</p>
								</div>
							</div>
						{/if}
					{/each}
				</div>
				<div class="mt-4 flex gap-2">
					<input
						type="text"
						bind:value={newMessage}
						placeholder="Type your message..."
						class="flex-1 rounded-lg border border-gray-300 px-4 py-2 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
					/>
					<Button variant="primary" disabled={!newMessage.trim()}>
						<Send class="h-4 w-4" />
					</Button>
				</div>
			</Card>
		</div>

		<!-- Sidebar: Timeline -->
		<div>
			<Card>
				<h2 class="font-semibold text-gray-900 dark:text-white">Timeline</h2>
				<div class="mt-4 space-y-0">
					{#each dispute.timeline as step, i}
						<div class="flex gap-3 {i < dispute.timeline.length - 1 ? 'pb-6' : ''}">
							<div class="flex flex-col items-center">
								{#if step.status === 'completed'}
									<div class="flex h-6 w-6 items-center justify-center rounded-full bg-secondary-600 text-white">
										<CheckCircle class="h-4 w-4" />
									</div>
								{:else if step.status === 'current'}
									<div class="flex h-6 w-6 items-center justify-center rounded-full border-2 border-primary-600 bg-primary-50 dark:bg-primary-900/30">
										<Clock class="h-3 w-3 text-primary-600" />
									</div>
								{:else}
									<div class="h-6 w-6 rounded-full border-2 border-gray-300 dark:border-gray-600"></div>
								{/if}
								{#if i < dispute.timeline.length - 1}
									<div class="mt-1 w-0.5 flex-1 {step.status === 'completed' ? 'bg-secondary-600' : 'bg-gray-200 dark:bg-gray-700'}"></div>
								{/if}
							</div>
							<div>
								<p class="text-sm font-medium {step.status === 'pending' ? 'text-gray-400' : 'text-gray-900 dark:text-white'}">{step.label}</p>
								{#if step.date}
									<p class="text-xs text-gray-500 dark:text-gray-400">{step.date}</p>
								{/if}
							</div>
						</div>
					{/each}
				</div>
			</Card>

			<Card class="mt-4">
				<h2 class="font-semibold text-gray-900 dark:text-white">Parties</h2>
				<div class="mt-3 space-y-3">
					<div class="flex items-center gap-3">
						<Avatar name={dispute.raisedBy} size="sm" />
						<div>
							<p class="text-sm font-medium text-gray-900 dark:text-white">{dispute.raisedBy}</p>
							<p class="text-xs text-gray-500 dark:text-gray-400">Raised by</p>
						</div>
					</div>
					<div class="flex items-center gap-3">
						<Avatar name={dispute.against} size="sm" />
						<div>
							<p class="text-sm font-medium text-gray-900 dark:text-white">{dispute.against}</p>
							<p class="text-xs text-gray-500 dark:text-gray-400">Against</p>
						</div>
					</div>
				</div>
			</Card>
		</div>
	</div>
</div>
