<script lang="ts">
	import { Award, TrendingUp, Gift, Star, Trophy, ArrowUp, ArrowDown, Info } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';

	const balance = 1250;
	const currentLevel = { level: 2, name: 'Silver', minPoints: 500, maxPoints: 2000 };
	const nextLevel = { level: 3, name: 'Gold', minPoints: 2000, maxPoints: 5000 };
	const progress = ((balance - currentLevel.minPoints) / (nextLevel.minPoints - currentLevel.minPoints)) * 100;

	const pointsHistory = [
		{ id: '1', points: 50, type: 'earned', reason: 'Completed job: Fix leaking tap', date: '2026-03-15' },
		{ id: '2', points: 25, type: 'earned', reason: 'Left a review', date: '2026-03-14' },
		{ id: '3', points: -100, type: 'spent', reason: 'Redeemed: Priority listing', date: '2026-03-10' },
		{ id: '4', points: 100, type: 'bonus', reason: 'Referral bonus: Invited Priya', date: '2026-03-05' },
		{ id: '5', points: 50, type: 'earned', reason: 'Completed job: Electrical repair', date: '2026-03-01' },
		{ id: '6', points: 75, type: 'earned', reason: '5-star review received', date: '2026-02-28' },
		{ id: '7', points: 50, type: 'earned', reason: 'Completed job: Deep cleaning', date: '2026-02-20' }
	];

	const leaderboard = [
		{ rank: 1, name: 'Suresh Nair', points: 4500, level: 'Gold' },
		{ rank: 2, name: 'Priya Sharma', points: 3800, level: 'Gold' },
		{ rank: 3, name: 'Ravi Kumar', points: 3200, level: 'Gold' },
		{ rank: 4, name: 'Deepak Sharma', points: 2800, level: 'Silver' },
		{ rank: 5, name: 'Anita Gupta', points: 2500, level: 'Silver' }
	];

	const earnWays = [
		{ action: 'Complete a job', points: '50 pts', icon: Star },
		{ action: 'Receive a 5-star review', points: '75 pts', icon: Star },
		{ action: 'Leave a review', points: '25 pts', icon: Star },
		{ action: 'Refer a friend', points: '100 pts', icon: Gift },
		{ action: 'Maintain 95%+ completion', points: '200 pts/month', icon: Trophy },
		{ action: 'First job of the day', points: '10 pts', icon: TrendingUp }
	];
</script>

<svelte:head>
	<title>Points & Rewards - Seva</title>
</svelte:head>

<div class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
	<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Points & Rewards</h1>

	<!-- Balance Card -->
	<div class="mt-6 rounded-2xl bg-gradient-to-r from-primary-600 to-primary-700 p-6 text-white dark:from-primary-700 dark:to-primary-800">
		<div class="flex items-center gap-3">
			<Award class="h-8 w-8" />
			<div>
				<p class="text-sm text-primary-100">Available Points</p>
				<p class="text-4xl font-bold">{balance.toLocaleString()}</p>
			</div>
		</div>
		<div class="mt-6">
			<div class="flex items-center justify-between text-sm">
				<span>{currentLevel.name} (Level {currentLevel.level})</span>
				<span>{nextLevel.name} (Level {nextLevel.level})</span>
			</div>
			<div class="mt-2 h-3 rounded-full bg-white/20">
				<div class="h-3 rounded-full bg-white transition-all" style="width: {progress}%"></div>
			</div>
			<p class="mt-1 text-xs text-primary-200">{nextLevel.minPoints - balance} points to reach {nextLevel.name}</p>
		</div>
	</div>

	<div class="mt-8 grid gap-6 lg:grid-cols-3">
		<!-- Points History -->
		<div class="lg:col-span-2">
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Points History</h2>
			<div class="mt-4 space-y-2">
				{#each pointsHistory as entry}
					<div class="flex items-center gap-3 rounded-lg border border-gray-200 bg-white p-3 dark:border-gray-700 dark:bg-gray-800">
						<div class="flex h-8 w-8 items-center justify-center rounded-full
							{entry.points > 0 ? 'bg-green-100 text-green-600 dark:bg-green-900/20' : 'bg-red-100 text-red-600 dark:bg-red-900/20'}">
							{#if entry.points > 0}
								<ArrowUp class="h-4 w-4" />
							{:else}
								<ArrowDown class="h-4 w-4" />
							{/if}
						</div>
						<div class="flex-1 min-w-0">
							<p class="text-sm text-gray-900 dark:text-white truncate">{entry.reason}</p>
							<p class="text-xs text-gray-500 dark:text-gray-400">{entry.date}</p>
						</div>
						<span class="text-sm font-semibold {entry.points > 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}">
							{entry.points > 0 ? '+' : ''}{entry.points}
						</span>
					</div>
				{/each}
			</div>
		</div>

		<!-- Sidebar -->
		<div class="space-y-6">
			<!-- Leaderboard -->
			<Card>
				<div class="flex items-center gap-2">
					<Trophy class="h-5 w-5 text-yellow-500" />
					<h2 class="font-semibold text-gray-900 dark:text-white">Leaderboard</h2>
				</div>
				<div class="mt-4 space-y-3">
					{#each leaderboard as entry}
						<div class="flex items-center gap-3">
							<span class="w-5 text-center text-sm font-bold {entry.rank <= 3 ? 'text-yellow-500' : 'text-gray-400'}">
								{entry.rank}
							</span>
							<Avatar name={entry.name} size="sm" />
							<div class="flex-1 min-w-0">
								<p class="text-sm font-medium text-gray-900 dark:text-white truncate">{entry.name}</p>
								<p class="text-xs text-gray-500 dark:text-gray-400">{entry.level}</p>
							</div>
							<span class="text-xs font-semibold text-gray-700 dark:text-gray-300">{entry.points.toLocaleString()}</span>
						</div>
					{/each}
				</div>
			</Card>

			<!-- How to Earn -->
			<Card>
				<div class="flex items-center gap-2">
					<Info class="h-5 w-5 text-primary-600" />
					<h2 class="font-semibold text-gray-900 dark:text-white">How to Earn</h2>
				</div>
				<div class="mt-4 space-y-3">
					{#each earnWays as way}
						{@const Icon = way.icon}
						<div class="flex items-center gap-3">
							<Icon class="h-4 w-4 text-gray-400 shrink-0" />
							<span class="flex-1 text-sm text-gray-600 dark:text-gray-400">{way.action}</span>
							<Badge variant="success" size="sm">{way.points}</Badge>
						</div>
					{/each}
				</div>
			</Card>
		</div>
	</div>
</div>
