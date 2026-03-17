<script lang="ts">
	import { Phone, KeyRound, ChevronDown } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import Button from '$lib/components/ui/Button.svelte';
	import { toastError, toastSuccess } from '$lib/stores/toast';
	import { jurisdictions, jurisdictionMap, type Jurisdiction } from '$lib/data/jurisdictions';
	import { t } from '$lib/i18n/index.svelte';

	let phone = $state('');
	let selectedJurisdiction = $state<Jurisdiction>(jurisdictionMap['IN']);
	let otpDigits = $state<string[]>(['', '', '', '', '', '']);
	let otpSent = $state(false);
	let loading = $state(false);
	let resendTimer = $state(0);
	let otpInputs: HTMLInputElement[] = $state([]);

	function selectJurisdiction(code: string) {
		const j = jurisdictionMap[code];
		if (j) {
			selectedJurisdiction = j;
		}
	}

	function startResendTimer() {
		resendTimer = 30;
		const interval = setInterval(() => {
			resendTimer -= 1;
			if (resendTimer <= 0) {
				clearInterval(interval);
			}
		}, 1000);
	}

	async function requestOtp(e?: Event) {
		e?.preventDefault();
		if (!phone.trim()) return;
		loading = true;
		try {
			// Mock: in real app, call requestOtp from auth store
			await new Promise((r) => setTimeout(r, 1000));
			otpSent = true;
			startResendTimer();
			toastSuccess(t('auth.otp_sent_to', { phone: selectedJurisdiction.phoneCode + phone }));
			// Focus first OTP input
			setTimeout(() => otpInputs[0]?.focus(), 100);
		} catch (err) {
			toastError(t('auth.otp_failed'));
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
			// Mock: in real app, call login from auth store
			await new Promise((r) => setTimeout(r, 1000));
			toastSuccess(t('auth.signed_in_success'));
			goto('/dashboard');
		} catch (err) {
			toastError(t('auth.invalid_otp_retry'));
		} finally {
			loading = false;
		}
	}

	function handleOtpInput(index: number, e: Event) {
		const input = e.target as HTMLInputElement;
		const value = input.value;

		if (value.length > 0) {
			otpDigits[index] = value.slice(-1);
			// Auto-focus next input
			if (index < 5) {
				otpInputs[index + 1]?.focus();
			}
		} else {
			otpDigits[index] = '';
		}

		// Auto-submit when all digits filled
		if (otpDigits.every((d) => d !== '')) {
			verifyOtp();
		}
	}

	function handleOtpKeydown(index: number, e: KeyboardEvent) {
		if (e.key === 'Backspace' && !otpDigits[index] && index > 0) {
			otpInputs[index - 1]?.focus();
		}
	}

	function handleOtpPaste(e: ClipboardEvent) {
		e.preventDefault();
		const text = e.clipboardData?.getData('text') || '';
		const digits = text.replace(/\D/g, '').slice(0, 6).split('');
		digits.forEach((d, i) => {
			otpDigits[i] = d;
		});
		if (digits.length > 0) {
			const focusIndex = Math.min(digits.length, 5);
			otpInputs[focusIndex]?.focus();
		}
		if (otpDigits.every((d) => d !== '')) {
			verifyOtp();
		}
	}

	async function resendOtp() {
		if (resendTimer > 0) return;
		await requestOtp();
	}
</script>

<svelte:head>
	<title>{t('auth.sign_in_title')} - Seva</title>
</svelte:head>

<div class="flex min-h-[calc(100vh-theme(spacing.32))] items-center justify-center px-4 py-12">
	<div class="w-full max-w-md">
		<div class="rounded-xl border border-gray-200 bg-white p-8 shadow-sm dark:border-gray-700 dark:bg-gray-800">
			<div class="text-center">
				<div class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-primary-100 dark:bg-primary-900/30">
					<Phone class="h-6 w-6 text-primary-600" />
				</div>
				<h1 class="mt-4 text-2xl font-bold text-gray-900 dark:text-white">{t('auth.sign_in_title')}</h1>
				<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
					{#if !otpSent}
						{t('auth.sign_in_subtitle')}
					{:else}
						{t('auth.sign_in_otp_subtitle', { phone: selectedJurisdiction.phoneCode + phone })}
					{/if}
				</p>
			</div>

			{#if !otpSent}
				<!-- Phone Number Step -->
				<form onsubmit={requestOtp} class="mt-6">
					<label for="phone" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
						{t('auth.phone_number')}
					</label>
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
						<div class="relative flex-1">
							<input
								id="phone"
								type="tel"
								bind:value={phone}
								placeholder={selectedJurisdiction.phonePlaceholder}
								maxlength={selectedJurisdiction.phoneMaxLength}
								required
								class="w-full rounded-lg border border-gray-300 px-4 py-3 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
							/>
						</div>
					</div>
					<Button
						type="submit"
						variant="primary"
						size="lg"
						{loading}
						disabled={!phone.trim()}
						class="mt-4 w-full"
					>
						{t('auth.send_otp')}
					</Button>
				</form>
			{:else}
				<!-- OTP Verification Step -->
				<form onsubmit={verifyOtp} class="mt-6">
					<label class="block text-sm font-medium text-gray-700 dark:text-gray-300">
						{t('auth.verification_code')}
					</label>
					<!-- OTP Digit Inputs -->
					<div class="mt-2 flex justify-center gap-2" onpaste={handleOtpPaste}>
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

					<Button
						type="submit"
						variant="primary"
						size="lg"
						{loading}
						disabled={otpDigits.some((d) => !d)}
						class="mt-4 w-full"
					>
						{t('auth.verify_sign_in')}
					</Button>

					<div class="mt-4 flex items-center justify-between">
						<button
							type="button"
							onclick={() => { otpSent = false; otpDigits = ['', '', '', '', '', '']; }}
							class="text-sm text-primary-600 hover:text-primary-700"
						>
							{t('auth.change_number')}
						</button>
						<button
							type="button"
							onclick={resendOtp}
							disabled={resendTimer > 0}
							class="text-sm {resendTimer > 0 ? 'text-gray-400 dark:text-gray-500' : 'text-primary-600 hover:text-primary-700'}"
						>
							{#if resendTimer > 0}
								{t('auth.resend_in', { seconds: resendTimer })}
							{:else}
								{t('auth.resend_otp')}
							{/if}
						</button>
					</div>
				</form>
			{/if}

			<p class="mt-6 text-center text-sm text-gray-600 dark:text-gray-400">
				{t('auth.dont_have_account')}
				<a href="/register" class="font-medium text-primary-600 hover:text-primary-700">{t('auth.register')}</a>
			</p>
		</div>
	</div>
</div>
