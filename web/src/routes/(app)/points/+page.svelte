<script lang="ts">
	import { onMount } from 'svelte';
	import { Award, TrendingUp, Gift, Star, Trophy, ArrowUp, ArrowDown, Info, Loader2 } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import api from '$lib/api/client';

	let loading = $state(true);
	let error = $state('');

	let balance = $state(0);
	let currentLevel = $state({ level: 1, name: 'Bronze', minPoints: 0, maxPoints: 500 });
	let nextLevel = $state<{ level: number; name: string; minPoints: number; maxPoints: number } | null>({ level: 2, name: 'Silver', minPoints: 500, maxPoints: 2000 });
	let progress = $derived(nextLevel ? ((balance - currentLevel.minPoints) / (nextLevel.minPoints - currentLevel.minPoints)) * 100 : 100);

	let pointsHistory = $state<any[]>([]);
	let leaderboard = $state<any[]>([]);

	onMount(async () => {
		try {
			const [balanceRes, historyRes, levelRes, leaderboardRes] = await Promise.all([
				api.points.getBalance(),
				api.points.getHistory({ per_page: 10 }),
				api.points.getLevel(),
				api.points.getLeaderboard({ limit: 5 }).catch(() => ({ data: [] }))
			]);

			balance = balanceRes.data?.balance || 0;
			const level = balanceRes.data?.level || levelRes.data?.current;
			if (level) {
				currentLevel = {
					level: level.level || 1,
					name: level.name || 'Bronze',
					minPoints: level.min_points || 0,
					maxPoints: level.max_points || 500
				};
			}
			if (levelRes.data?.next) {
				nextLevel = {
					level: levelRes.data.next.level || 2,
					name: levelRes.data.next.name || 'Silver',
					minPoints: levelRes.data.next.min_points || 500,
					maxPoints: levelRes.data.next.max_points || 2000
				};
			} else {
				nextLevel = null;
			}

			pointsHistory = (historyRes.data || []).map((entry: any) => ({
				id: entry.id,
				points: entry.points || entry.amount || 0,
				type: entry.type || (entry.points > 0 ? 'earned' : 'spent'),
				reason: entry.reason || entry.description || '',
				date: entry.created_at?.split('T')[0] || ''
			}));

			leaderboard = (leaderboardRes.data || []).map((entry: any, i: number) => ({
				rank: entry.rank || i + 1,
				name: entry.user?.name || entry.name || 'User',
				points: entry.points || entry.total_points || 0,
				level: entry.level?.name || entry.level || ''
			}));
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load points data';
		} finally {
			loading = false;
		}
	});

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
				<span>{nextLevel ? `${nextLevel.name} (Level ${nextLevel.level})` : 'Max Level'}</span>
			</div>
			<div class="mt-2 h-3 rounded-full bg-white/20">
				<div class="h-3 rounded-full bg-white transition-all" style="width: {progress}%"></div>
			</div>
			<p class="mt-1 text-xs text-primary-200">{nextLevel ? `${nextLevel.minPoints - balance} points to reach ${nextLevel.name}` : 'You have reached the highest level!'}</p>
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
{/if}
