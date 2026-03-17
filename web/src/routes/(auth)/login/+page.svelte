<script lang="ts">
	import { Phone, KeyRound, ChevronDown } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import Button from '$lib/components/ui/Button.svelte';
	import { toastError, toastSuccess } from '$lib/stores/toast';

	let phone = $state('');
	let countryCode = $state('+91');
	let otpDigits = $state<string[]>(['', '', '', '', '', '']);
	let otpSent = $state(false);
	let loading = $state(false);
	let resendTimer = $state(0);
	let otpInputs: HTMLInputElement[] = $state([]);

	const countryCodes = [
		{ code: '+91', country: 'IN' },
		{ code: '+44', country: 'UK' },
		{ code: '+1', country: 'US' },
		{ code: '+61', country: 'AU' },
		{ code: '+971', country: 'AE' }
	];

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
			toastSuccess('OTP sent to ' + countryCode + phone);
			// Focus first OTP input
			setTimeout(() => otpInputs[0]?.focus(), 100);
		} catch (err) {
			toastError('Failed to send OTP. Please try again.');
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
			toastSuccess('Signed in successfully');
			goto('/dashboard');
		} catch (err) {
			toastError('Invalid OTP. Please try again.');
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
	<title>Sign In - Seva</title>
</svelte:head>

<div class="flex min-h-[calc(100vh-theme(spacing.32))] items-center justify-center px-4 py-12">
	<div class="w-full max-w-md">
		<div class="rounded-xl border border-gray-200 bg-white p-8 shadow-sm dark:border-gray-700 dark:bg-gray-800">
			<div class="text-center">
				<div class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-primary-100 dark:bg-primary-900/30">
					<Phone class="h-6 w-6 text-primary-600" />
				</div>
				<h1 class="mt-4 text-2xl font-bold text-gray-900 dark:text-white">Sign in to Seva</h1>
				<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
					{#if !otpSent}
						Enter your phone number to receive a verification code.
					{:else}
						Enter the 6-digit code sent to {countryCode}{phone}
					{/if}
				</p>
			</div>

			{#if !otpSent}
				<!-- Phone Number Step -->
				<form onsubmit={requestOtp} class="mt-6">
					<label for="phone" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
						Phone number
					</label>
					<div class="mt-1 flex gap-2">
						<select
							bind:value={countryCode}
							class="rounded-lg border border-gray-300 px-3 py-3 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
						>
							{#each countryCodes as cc}
								<option value={cc.code}>{cc.country} {cc.code}</option>
							{/each}
						</select>
						<div class="relative flex-1">
							<input
								id="phone"
								type="tel"
								bind:value={phone}
								placeholder="9876543210"
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
						Send OTP
					</Button>
				</form>
			{:else}
				<!-- OTP Verification Step -->
				<form onsubmit={verifyOtp} class="mt-6">
					<label class="block text-sm font-medium text-gray-700 dark:text-gray-300">
						Verification code
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
						Verify & Sign In
					</Button>

					<div class="mt-4 flex items-center justify-between">
						<button
							type="button"
							onclick={() => { otpSent = false; otpDigits = ['', '', '', '', '', '']; }}
							class="text-sm text-primary-600 hover:text-primary-700"
						>
							Change number
						</button>
						<button
							type="button"
							onclick={resendOtp}
							disabled={resendTimer > 0}
							class="text-sm {resendTimer > 0 ? 'text-gray-400 dark:text-gray-500' : 'text-primary-600 hover:text-primary-700'}"
						>
							{#if resendTimer > 0}
								Resend in {resendTimer}s
							{:else}
								Resend OTP
							{/if}
						</button>
					</div>
				</form>
			{/if}

			<p class="mt-6 text-center text-sm text-gray-600 dark:text-gray-400">
				Don't have an account?
				<a href="/register" class="font-medium text-primary-600 hover:text-primary-700">Register</a>
			</p>
		</div>
	</div>
</div>
