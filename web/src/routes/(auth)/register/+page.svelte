<script lang="ts">
	import { User, Briefcase, Phone, CheckCircle, MapPin, ArrowLeft, ArrowRight, CreditCard, Smartphone, Banknote, Wallet, Building2 } from 'lucide-svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import { toastError, toastSuccess } from '$lib/stores/toast';
	import { jurisdictions, jurisdictionMap, type Jurisdiction } from '$lib/data/jurisdictions';
	import { topLevelCategories, getSubcategories } from '$lib/data/categories';
	import { t } from '$lib/i18n/index.svelte';
	import { requestOtp as authRequestOtp, login as authLogin, register as authRegister } from '$lib/stores/auth';

	type Role = 'customer' | 'provider' | null;

	let step = $state(1);
	let selectedRole: Role = $state(null);
	let phone = $state('');
	let selectedJurisdiction = $state<Jurisdiction>(jurisdictionMap['IN']);
	let otpDigits = $state<string[]>(['', '', '', '', '', '']);
	let otpSent = $state(false);
	let otpVerified = $state(false);
	let resendTimer = $state(0);
	let loading = $state(false);
	let otpInputs: HTMLInputElement[] = $state([]);

	// Step 3 fields
	let name = $state('');
	let email = $state('');
	let postcode = $state('');
	// Provider-specific
	let bio = $state('');
	let selectedSkills = $state<string[]>([]);
	let serviceRadius = $state(10);
	let selectedPaymentMethod = $state('');

	// Set default payment method when jurisdiction changes
	$effect(() => {
		const defaultMethod = selectedJurisdiction.paymentMethods.find((m) => m.isDefault);
		if (defaultMethod) {
			selectedPaymentMethod = defaultMethod.id;
		}
	});

	$effect(() => {
		const role = $page.url.searchParams.get('role');
		if (role === 'provider') {
			selectedRole = 'provider';
			step = 2;
		}
	});

	function selectJurisdiction(code: string) {
		const j = jurisdictionMap[code];
		if (j) {
			selectedJurisdiction = j;
		}
	}

	function selectRole(role: Role) {
		selectedRole = role;
		step = 2;
	}

	function startResendTimer() {
		resendTimer = 30;
		const interval = setInterval(() => {
			resendTimer -= 1;
			if (resendTimer <= 0) clearInterval(interval);
		}, 1000);
	}

	async function sendOtp(e?: Event) {
		e?.preventDefault();
		if (!phone.trim()) return;
		loading = true;
		try {
			const fullPhone = selectedJurisdiction.phoneCode + phone;
			await authRequestOtp(fullPhone);
			otpSent = true;
			startResendTimer();
			toastSuccess(t('auth.otp_sent_to', { phone: fullPhone }));
			setTimeout(() => otpInputs[0]?.focus(), 100);
		} catch (err) {
			toastError(err instanceof Error ? err.message : t('auth.otp_sent_failed'));
		} finally {
			loading = false;
		}
	}

	async function verifyOtp(e?: Event) {
		e?.preventDefault();
		const otp = otpDigits.join('');
		if (otp.length !== 6) return;
		loading = true;
		try {
			const fullPhone = selectedJurisdiction.phoneCode + phone;
			await authLogin(fullPhone, otp);
			otpVerified = true;
			step = 3;
			toastSuccess(t('auth.phone_verified'));
		} catch (err) {
			toastError(err instanceof Error ? err.message : t('auth.invalid_otp'));
		} finally {
			loading = false;
		}
	}

	function handleOtpInput(index: number, e: Event) {
		const input = e.target as HTMLInputElement;
		const value = input.value;
		if (value.length > 0) {
			otpDigits[index] = value.slice(-1);
			if (index < 5) otpInputs[index + 1]?.focus();
		} else {
			otpDigits[index] = '';
		}
		if (otpDigits.every((d) => d !== '')) verifyOtp();
	}

	function handleOtpKeydown(index: number, e: KeyboardEvent) {
		if (e.key === 'Backspace' && !otpDigits[index] && index > 0) {
			otpInputs[index - 1]?.focus();
		}
	}

	function toggleSkill(skillId: string) {
		if (selectedSkills.includes(skillId)) {
			selectedSkills = selectedSkills.filter((s) => s !== skillId);
		} else {
			selectedSkills = [...selectedSkills, skillId];
		}
	}

	async function handleRegister(e?: Event) {
		e?.preventDefault();
		if (!name.trim()) {
			toastError(t('auth.please_enter_name'));
			return;
		}
		loading = true;
		try {
			const fullPhone = selectedJurisdiction.phoneCode + phone;
			await authRegister(
				name,
				fullPhone,
				selectedRole as 'customer' | 'provider',
				email || undefined,
				{
					postcode: postcode || undefined,
					bio: bio || undefined,
					categories: selectedSkills.length > 0 ? selectedSkills : undefined,
					service_radius_km: selectedRole === 'provider' ? serviceRadius : undefined
				}
			);
			step = 4;
			toastSuccess(t('auth.account_created'));
		} catch (err) {
			toastError(err instanceof Error ? err.message : t('auth.registration_failed'));
		} finally {
			loading = false;
		}
	}

	const stepLabels = $derived([
		t('auth.step_role'),
		t('auth.step_verify'),
		t('auth.step_details'),
		t('auth.step_done')
	]);

	// Payment method icon mapping
	function getPaymentIcon(iconName: string) {
		const icons: Record<string, any> = {
			CreditCard,
			Smartphone,
			Banknote,
			Wallet,
			Building2
		};
		return icons[iconName] || CreditCard;
	}
</script>

<svelte:head>
	<title>{t('auth.sign_up_title')} - Seva</title>
</svelte:head>

<div class="flex min-h-[calc(100vh-theme(spacing.32))] items-center justify-center px-4 py-12">
	<div class="w-full max-w-lg">
		<!-- Progress Indicator -->
		<div class="mb-8 flex items-center justify-center gap-2">
			{#each stepLabels as label, i}
				<div class="flex items-center gap-2">
					<div class="flex h-8 w-8 items-center justify-center rounded-full text-xs font-semibold
						{i + 1 < step
							? 'bg-primary-600 text-white'
							: i + 1 === step
								? 'border-2 border-primary-600 text-primary-600 dark:text-primary-400'
								: 'border-2 border-gray-300 text-gray-400 dark:border-gray-600'}">
						{#if i + 1 < step}
							<CheckCircle class="h-5 w-5" />
						{:else}
							{i + 1}
						{/if}
					</div>
					{#if i < stepLabels.length - 1}
						<div class="h-0.5 w-8 {i + 1 < step ? 'bg-primary-600' : 'bg-gray-200 dark:bg-gray-700'}"></div>
					{/if}
				</div>
			{/each}
		</div>

		<div class="rounded-xl border border-gray-200 bg-white p-8 shadow-sm dark:border-gray-700 dark:bg-gray-800">
			<!-- Step 1: Role Selection -->
			{#if step === 1}
				<h1 class="text-2xl font-bold text-gray-900 dark:text-white">{t('auth.sign_up_title')}</h1>
				<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">{t('auth.sign_up_subtitle')}</p>

				<div class="mt-6 grid gap-4">
					<button
						onclick={() => selectRole('customer')}
						class="flex items-center gap-4 rounded-lg border-2 border-gray-200 p-5 text-left transition hover:border-primary-500 hover:shadow-sm dark:border-gray-600 dark:hover:border-primary-500"
					>
						<div class="flex h-14 w-14 items-center justify-center rounded-xl bg-primary-100 text-primary-600 dark:bg-primary-900/30">
							<User class="h-7 w-7" />
						</div>
						<div>
							<h3 class="font-semibold text-gray-900 dark:text-white">{t('auth.i_need_service')}</h3>
							<p class="mt-0.5 text-sm text-gray-500 dark:text-gray-400">{t('auth.i_need_service_desc')}</p>
						</div>
					</button>
					<button
						onclick={() => selectRole('provider')}
						class="flex items-center gap-4 rounded-lg border-2 border-gray-200 p-5 text-left transition hover:border-secondary-500 hover:shadow-sm dark:border-gray-600 dark:hover:border-secondary-500"
					>
						<div class="flex h-14 w-14 items-center justify-center rounded-xl bg-secondary-100 text-secondary-600 dark:bg-secondary-900/30">
							<Briefcase class="h-7 w-7" />
						</div>
						<div>
							<h3 class="font-semibold text-gray-900 dark:text-white">{t('auth.i_provide_services')}</h3>
							<p class="mt-0.5 text-sm text-gray-500 dark:text-gray-400">{t('auth.i_provide_services_desc')}</p>
						</div>
					</button>
				</div>

			<!-- Step 2: Phone Verification -->
			{:else if step === 2}
				<div class="flex items-center gap-2 mb-4">
					<button onclick={() => { step = 1; selectedRole = null; otpSent = false; otpDigits = ['','','','','','']; }} class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
						<ArrowLeft class="h-5 w-5" />
					</button>
					<h1 class="text-2xl font-bold text-gray-900 dark:text-white">{t('auth.verify_phone')}</h1>
				</div>
				<div class="flex items-center gap-2 rounded-lg bg-gray-50 px-3 py-2 text-sm dark:bg-gray-700">
					{#if selectedRole === 'customer'}
						<User class="h-4 w-4 text-primary-600" />
						<span class="text-gray-700 dark:text-gray-300">{t('auth.registering_as_customer')}</span>
					{:else}
						<Briefcase class="h-4 w-4 text-secondary-600" />
						<span class="text-gray-700 dark:text-gray-300">{t('auth.registering_as_provider')}</span>
					{/if}
				</div>

				{#if !otpSent}
					<form onsubmit={sendOtp} class="mt-6">
						<label for="reg-phone" class="block text-sm font-medium text-gray-700 dark:text-gray-300">{t('auth.phone_number')}</label>
						<div class="mt-1 flex gap-2">
							<select
								value={selectedJurisdiction.code}
								onchange={(e) => selectJurisdiction((e.target as HTMLSelectElement).value)}
								class="rounded-lg border border-gray-300 px-3 py-3 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
							>
								{#each jurisdictions as j}
									<option value={j.code}>{j.flag} {j.code} {j.phoneCode}</option>
								{/each}
							</select>
							<input
								id="reg-phone"
								type="tel"
								bind:value={phone}
								placeholder={selectedJurisdiction.phonePlaceholder}
								maxlength={selectedJurisdiction.phoneMaxLength}
								required
								class="flex-1 rounded-lg border border-gray-300 px-4 py-3 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
							/>
						</div>
						<Button type="submit" variant="primary" size="lg" {loading} disabled={!phone.trim()} class="mt-4 w-full">
							{t('auth.send_verification_code')}
						</Button>
					</form>
				{:else}
					<form onsubmit={verifyOtp} class="mt-6">
						<p class="text-sm text-gray-600 dark:text-gray-400">{t('auth.otp_code_sent_to', { phone: selectedJurisdiction.phoneCode + phone })}</p>
						<div class="mt-3 flex justify-center gap-2">
							{#each otpDigits as digit, i}
								<input
									bind:this={otpInputs[i]}
									type="text"
									inputmode="numeric"
									maxlength="1"
									value={digit}
									oninput={(e) => handleOtpInput(i, e)}
									onkeydown={(e) => handleOtpKeydown(i, e)}
									class="h-12 w-12 rounded-lg border border-gray-300 text-center text-lg font-semibold focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
								/>
							{/each}
						</div>
						<Button type="submit" variant="primary" size="lg" {loading} disabled={otpDigits.some(d => !d)} class="mt-4 w-full">
							{t('auth.verify')}
						</Button>
						<div class="mt-3 text-center">
							<button
								type="button"
								onclick={() => { if (resendTimer <= 0) sendOtp(); }}
								disabled={resendTimer > 0}
								class="text-sm {resendTimer > 0 ? 'text-gray-400' : 'text-primary-600 hover:text-primary-700'}"
							>
								{resendTimer > 0 ? t('auth.resend_in', { seconds: resendTimer }) : t('auth.resend_otp')}
							</button>
						</div>
					</form>
				{/if}

			<!-- Step 3: Details -->
			{:else if step === 3}
				<div class="flex items-center gap-2 mb-4">
					<button onclick={() => { step = 2; }} class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
						<ArrowLeft class="h-5 w-5" />
					</button>
					<h1 class="text-2xl font-bold text-gray-900 dark:text-white">
						{selectedRole === 'provider' ? t('auth.provider_details') : t('auth.your_details')}
					</h1>
				</div>

				<form onsubmit={handleRegister} class="space-y-4">
					<Input label={t('auth.full_name')} bind:value={name} required placeholder={t('auth.your_full_name')} />
					<Input label={t('auth.email_optional')} type="email" bind:value={email} placeholder={t('auth.email_placeholder')} />
					<Input label={t('auth.postcode')} bind:value={postcode} required placeholder={selectedJurisdiction.postcodePlaceholder} icon={MapPin} />

					{#if selectedRole === 'provider'}
						<div>
							<label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">{t('auth.skills_services')}</label>
							<div class="space-y-3">
								{#each topLevelCategories as category}
									{@const subcategories = getSubcategories(category.id)}
									{#if subcategories.length > 0}
										<div>
											<p class="mb-1.5 text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
												{t(category.translationKey)}
											</p>
											<div class="flex flex-wrap gap-2">
												{#each subcategories as sub}
													<button
														type="button"
														onclick={() => toggleSkill(sub.id)}
														class="rounded-full border px-3 py-1.5 text-xs font-medium transition
															{selectedSkills.includes(sub.id)
																? 'border-primary-500 bg-primary-50 text-primary-700 dark:border-primary-400 dark:bg-primary-900/20 dark:text-primary-400'
																: 'border-gray-300 text-gray-600 hover:border-gray-400 dark:border-gray-600 dark:text-gray-400'}"
													>
														{t(sub.translationKey)}
													</button>
												{/each}
											</div>
										</div>
									{/if}
								{/each}
							</div>
						</div>

						<div>
							<label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
								{t('auth.service_radius_value', { radius: serviceRadius })}
							</label>
							<input
								type="range"
								bind:value={serviceRadius}
								min="1"
								max="50"
								class="w-full accent-primary-600"
							/>
							<div class="flex justify-between text-xs text-gray-400">
								<span>1 km</span>
								<span>50 km</span>
							</div>
						</div>

						<div>
							<label for="bio" class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">{t('auth.bio')}</label>
							<textarea
								id="bio"
								bind:value={bio}
								rows="3"
								placeholder={t('auth.bio_placeholder')}
								class="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400"
							></textarea>
						</div>
					{/if}

					<!-- Preferred Payment Method -->
					<div>
						<label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">{t('auth.preferred_payment')}</label>
						<div class="grid gap-2">
							{#each selectedJurisdiction.paymentMethods as method}
								{@const IconComponent = getPaymentIcon(method.icon)}
								<button
									type="button"
									onclick={() => selectedPaymentMethod = method.id}
									class="flex items-center gap-3 rounded-lg border-2 px-4 py-3 text-left text-sm transition
										{selectedPaymentMethod === method.id
											? 'border-primary-500 bg-primary-50 dark:border-primary-400 dark:bg-primary-900/20'
											: 'border-gray-200 hover:border-gray-300 dark:border-gray-600 dark:hover:border-gray-500'}"
								>
									<div class="flex h-8 w-8 items-center justify-center rounded-lg
										{selectedPaymentMethod === method.id
											? 'bg-primary-100 text-primary-600 dark:bg-primary-900/30 dark:text-primary-400'
											: 'bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400'}">
										<IconComponent class="h-4 w-4" />
									</div>
									<div class="flex-1">
										<p class="font-medium text-gray-900 dark:text-white">{method.name}</p>
										<p class="text-xs text-gray-500 dark:text-gray-400">{method.description}</p>
									</div>
									{#if selectedPaymentMethod === method.id}
										<CheckCircle class="h-5 w-5 text-primary-600 dark:text-primary-400" />
									{/if}
								</button>
							{/each}
						</div>
					</div>

					<Button type="submit" variant="primary" size="lg" {loading} class="w-full">
						{t('auth.create_account')}
					</Button>
				</form>

			<!-- Step 4: Success -->
			{:else if step === 4}
				<div class="text-center py-8">
					<div class="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-secondary-100 dark:bg-secondary-900/30">
						<CheckCircle class="h-8 w-8 text-secondary-600" />
					</div>
					<h1 class="mt-4 text-2xl font-bold text-gray-900 dark:text-white">{t('auth.welcome_to_seva')}</h1>
					<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
						{t('auth.account_created_success', { action: selectedRole === 'provider' ? t('auth.start_accepting_jobs') : t('auth.find_providers') })}
					</p>
					<div class="mt-6">
						<Button variant="primary" size="lg" href="/dashboard" class="w-full">
							{t('auth.go_to_dashboard')}
						</Button>
					</div>
				</div>
			{/if}

			{#if step < 4}
				<p class="mt-6 text-center text-sm text-gray-600 dark:text-gray-400">
					{t('auth.already_have_account')}
					<a href="/login" class="font-medium text-primary-600 hover:text-primary-700">{t('auth.sign_in')}</a>
				</p>
			{/if}
		</div>
	</div>
</div>
