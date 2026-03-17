<script lang="ts">
	import { ArrowLeft, Check, Crown, Star, Zap, Building2, Clock, IndianRupee, XCircle } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import { t } from '$lib/i18n/index.svelte';

	// Current subscription state (would come from API in production)
	let currentPlan = $state('free');
	let loading = $state(false);
	let showCancelModal = $state(false);

	interface PlanDef {
		tier: string;
		name: string;
		price: number;
		leadsPerMonth: number;
		commissionDiscount: number;
		features: string[];
		description: string;
		popular?: boolean;
	}

	const plans: PlanDef[] = [
		{
			tier: 'free',
			name: t('subscription.free_plan'),
			price: 0,
			leadsPerMonth: 5,
			commissionDiscount: 0,
			features: ['5 ' + t('subscription.leads_per_month')],
			description: t('subscription.free_desc')
		},
		{
			tier: 'professional',
			name: t('subscription.pro_plan'),
			price: 299,
			leadsPerMonth: -1,
			commissionDiscount: 2,
			features: [
				t('subscription.unlimited_leads'),
				t('subscription.priority_search'),
				t('subscription.pro_badge'),
				t('subscription.analytics'),
				t('subscription.commission_discount', { percent: '2' })
			],
			description: t('subscription.pro_desc'),
			popular: true
		},
		{
			tier: 'enterprise',
			name: t('subscription.business_plan'),
			price: 999,
			leadsPerMonth: -1,
			commissionDiscount: 3,
			features: [
				t('subscription.unlimited_leads'),
				t('subscription.priority_search'),
				t('subscription.pro_badge'),
				t('subscription.analytics'),
				t('subscription.commission_discount', { percent: '3' }),
				'Team profiles',
				'Branded page',
				'Quote templates',
				'Invoice generation'
			],
			description: t('subscription.business_desc')
		}
	];

	// Usage stats (would come from API)
	const usageStats = {
		leadsUsed: 3,
		leadsTotal: 5,
		daysRemaining: 14,
		nextBillingDate: '2026-04-01'
	};

	// Billing history (would come from API)
	const billingHistory = [
		{ id: '1', date: '2026-03-01', amount: 0, tier: 'free', status: 'active' },
		{ id: '2', date: '2026-02-01', amount: 0, tier: 'free', status: 'expired' }
	];

	function getButtonLabel(planTier: string): string {
		if (planTier === currentPlan) return t('subscription.current_plan');
		const planOrder = ['free', 'professional', 'enterprise'];
		const currentIdx = planOrder.indexOf(currentPlan);
		const targetIdx = planOrder.indexOf(planTier);
		return targetIdx > currentIdx ? t('subscription.upgrade') : t('subscription.downgrade');
	}

	function getButtonVariant(planTier: string): 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger' {
		if (planTier === currentPlan) return 'outline';
		const planOrder = ['free', 'professional', 'enterprise'];
		const currentIdx = planOrder.indexOf(currentPlan);
		const targetIdx = planOrder.indexOf(planTier);
		return targetIdx > currentIdx ? 'primary' : 'outline';
	}

	async function handleSubscribe(tier: string) {
		if (tier === currentPlan) return;
		loading = true;
		try {
			// API call would go here
			// await fetch('/api/v1/subscriptions', { method: 'POST', body: JSON.stringify({ tier }) });
			currentPlan = tier;
		} catch (err) {
			console.error('Subscription failed:', err);
		} finally {
			loading = false;
		}
	}

	async function handleCancel() {
		loading = true;
		try {
			// API call would go here
			currentPlan = 'free';
			showCancelModal = false;
		} catch (err) {
			console.error('Cancellation failed:', err);
		} finally {
			loading = false;
		}
	}

	function getPlanIcon(tier: string) {
		switch (tier) {
			case 'professional': return Zap;
			case 'enterprise': return Building2;
			default: return Star;
		}
	}
</script>

<svelte:head>
	<title>{t('subscription.title')} - Seva Provider</title>
</svelte:head>

<div class="mx-auto max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
	<!-- Back navigation -->
	<a href="/provider/dashboard" class="inline-flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700">
		<ArrowLeft class="h-4 w-4" />
		Back to Dashboard
	</a>

	<h1 class="mt-4 text-2xl font-bold text-gray-900 dark:text-white">{t('subscription.title')}</h1>

	<!-- Current Plan Summary -->
	<Card class="mt-6">
		<div class="flex flex-wrap items-center justify-between gap-4">
			<div>
				<p class="text-sm text-gray-500 dark:text-gray-400">{t('subscription.current_plan')}</p>
				<div class="mt-1 flex items-center gap-2">
					<Crown class="h-5 w-5 text-yellow-500" />
					<span class="text-xl font-bold text-gray-900 dark:text-white">
						{plans.find((p) => p.tier === currentPlan)?.name || 'Free'}
					</span>
					<Badge variant="success">Active</Badge>
				</div>
			</div>
			<div class="text-right">
				{#if currentPlan !== 'free'}
					<p class="text-sm text-gray-500 dark:text-gray-400">Next billing: {usageStats.nextBillingDate}</p>
					<p class="text-xs text-gray-400 dark:text-gray-500">{usageStats.daysRemaining} days remaining</p>
				{/if}
			</div>
		</div>

		<!-- Usage bar for free plan -->
		{#if currentPlan === 'free'}
			<div class="mt-4">
				<div class="flex items-center justify-between text-sm">
					<span class="text-gray-600 dark:text-gray-400">Leads used this month</span>
					<span class="font-medium text-gray-900 dark:text-white">{usageStats.leadsUsed} / {usageStats.leadsTotal}</span>
				</div>
				<div class="mt-2 h-2 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-gray-700">
					<div
						class="h-full rounded-full bg-primary-500 transition-all"
						style="width: {(usageStats.leadsUsed / usageStats.leadsTotal) * 100}%"
					></div>
				</div>
			</div>
		{/if}
	</Card>

	<!-- Plan Comparison Cards -->
	<div class="mt-8 grid gap-6 md:grid-cols-3">
		{#each plans as plan}
			{@const Icon = getPlanIcon(plan.tier)}
			{@const isCurrent = plan.tier === currentPlan}
			<div
				class="relative flex flex-col rounded-xl border-2 p-6 transition-shadow hover:shadow-lg
					{isCurrent
						? 'border-primary-500 bg-primary-50/50 dark:border-primary-400 dark:bg-primary-900/10'
						: plan.popular
							? 'border-secondary-300 dark:border-secondary-600'
							: 'border-gray-200 dark:border-gray-700'}
					bg-white dark:bg-gray-800"
			>
				<!-- Popular badge -->
				{#if plan.popular}
					<div class="absolute -top-3 left-1/2 -translate-x-1/2">
						<Badge variant="warning" size="md">Most Popular</Badge>
					</div>
				{/if}

				<!-- Plan header -->
				<div class="text-center">
					<div class="mx-auto flex h-12 w-12 items-center justify-center rounded-full
						{isCurrent ? 'bg-primary-100 text-primary-600 dark:bg-primary-900/30 dark:text-primary-400' : 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'}">
						<Icon class="h-6 w-6" />
					</div>
					<h3 class="mt-3 text-lg font-bold text-gray-900 dark:text-white">{plan.name}</h3>
					<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{plan.description}</p>
				</div>

				<!-- Price -->
				<div class="mt-4 text-center">
					<div class="flex items-baseline justify-center gap-1">
						{#if plan.price === 0}
							<span class="text-3xl font-bold text-gray-900 dark:text-white">{t('subscription.free_plan')}</span>
						{:else}
							<IndianRupee class="h-5 w-5 text-gray-900 dark:text-white" />
							<span class="text-3xl font-bold text-gray-900 dark:text-white">{plan.price}</span>
							<span class="text-sm text-gray-500 dark:text-gray-400">{t('subscription.per_month')}</span>
						{/if}
					</div>
				</div>

				<!-- Features -->
				<ul class="mt-6 flex-1 space-y-3">
					{#each plan.features as feature}
						<li class="flex items-start gap-2 text-sm">
							<Check class="mt-0.5 h-4 w-4 flex-shrink-0 text-secondary-500" />
							<span class="text-gray-700 dark:text-gray-300">{feature}</span>
						</li>
					{/each}
				</ul>

				<!-- Action button -->
				<div class="mt-6">
					<Button
						variant={getButtonVariant(plan.tier)}
						class="w-full"
						disabled={isCurrent || loading}
						{loading}
						onclick={() => handleSubscribe(plan.tier)}
					>
						{getButtonLabel(plan.tier)}
					</Button>
				</div>
			</div>
		{/each}
	</div>

	<!-- Cancel subscription -->
	{#if currentPlan !== 'free'}
		<Card class="mt-8">
			<div class="flex flex-wrap items-center justify-between gap-4">
				<div>
					<h3 class="font-semibold text-gray-900 dark:text-white">{t('subscription.cancel')}</h3>
					<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
						You will lose access to premium features at the end of your billing period.
					</p>
				</div>
				<Button variant="danger" onclick={() => (showCancelModal = true)}>
					{t('subscription.cancel')}
				</Button>
			</div>
		</Card>
	{/if}

	<!-- Billing History -->
	<Card class="mt-8">
		<h2 class="text-lg font-semibold text-gray-900 dark:text-white">{t('subscription.billing_history')}</h2>
		<div class="mt-4 space-y-3">
			{#each billingHistory as entry}
				<div class="flex items-center gap-3 rounded-lg border border-gray-100 p-3 dark:border-gray-700">
					<div class="flex h-10 w-10 items-center justify-center rounded-full bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400">
						<Clock class="h-5 w-5" />
					</div>
					<div class="flex-1">
						<p class="text-sm font-medium text-gray-900 dark:text-white capitalize">{entry.tier} Plan</p>
						<p class="text-xs text-gray-500 dark:text-gray-400">{entry.date}</p>
					</div>
					<div class="text-right">
						<p class="text-sm font-semibold text-gray-900 dark:text-white">
							{#if entry.amount === 0}
								Free
							{:else}
								Rs. {entry.amount}
							{/if}
						</p>
						<Badge variant={entry.status === 'active' ? 'success' : 'neutral'} size="sm">{entry.status}</Badge>
					</div>
				</div>
			{/each}
		</div>
	</Card>
</div>

<!-- Cancel Confirmation Modal -->
{#if showCancelModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
		<div class="w-full max-w-md rounded-xl bg-white p-6 shadow-xl dark:bg-gray-800">
			<div class="flex items-center gap-3">
				<div class="flex h-10 w-10 items-center justify-center rounded-full bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400">
					<XCircle class="h-6 w-6" />
				</div>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white">{t('subscription.cancel')}?</h3>
			</div>
			<p class="mt-3 text-sm text-gray-600 dark:text-gray-400">
				Are you sure you want to cancel your subscription? You will be downgraded to the Free plan at the end of your current billing period.
			</p>
			<div class="mt-6 flex gap-3 justify-end">
				<Button variant="outline" onclick={() => (showCancelModal = false)}>
					Keep Plan
				</Button>
				<Button variant="danger" {loading} onclick={handleCancel}>
					{t('subscription.cancel')}
				</Button>
			</div>
		</div>
	</div>
{/if}
