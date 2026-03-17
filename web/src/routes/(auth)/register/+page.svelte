<script lang="ts">
	import { User, Briefcase, Phone, CheckCircle, MapPin, ArrowLeft, ArrowRight } from 'lucide-svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import { toastError, toastSuccess } from '$lib/stores/toast';

	type Role = 'customer' | 'provider' | null;

	let step = $state(1);
	let selectedRole: Role = $state(null);
	let phone = $state('');
	let countryCode = $state('+91');
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

	const availableSkills = [
		'Plumbing', 'Electrical', 'Cleaning', 'Gardening',
		'Painting', 'Moving', 'Carpentry', 'HVAC',
		'Roofing', 'Pest Control', 'Appliance Repair', 'Landscaping'
	];

	$effect(() => {
		const role = $page.url.searchParams.get('role');
		if (role === 'provider') {
			selectedRole = 'provider';
			step = 2;
		}
	});

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
			await new Promise((r) => setTimeout(r, 1000));
			otpSent = true;
			startResendTimer();
			toastSuccess('OTP sent to ' + countryCode + phone);
			setTimeout(() => otpInputs[0]?.focus(), 100);
		} catch {
			toastError('Failed to send OTP');
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
			await new Promise((r) => setTimeout(r, 1000));
			otpVerified = true;
			step = 3;
			toastSuccess('Phone verified successfully');
		} catch {
			toastError('Invalid OTP');
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

	function toggleSkill(skill: string) {
		if (selectedSkills.includes(skill)) {
			selectedSkills = selectedSkills.filter((s) => s !== skill);
		} else {
			selectedSkills = [...selectedSkills, skill];
		}
	}

	async function handleRegister(e?: Event) {
		e?.preventDefault();
		if (!name.trim()) {
			toastError('Please enter your name');
			return;
		}
		loading = true;
		try {
			await new Promise((r) => setTimeout(r, 1500));
			step = 4;
			toastSuccess('Account created successfully!');
		} catch {
			toastError('Registration failed. Please try again.');
		} finally {
			loading = false;
		}
	}

	const stepLabels = ['Role', 'Verify', 'Details', 'Done'];
</script>

<svelte:head>
	<title>Register - Seva</title>
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
				<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Create your account</h1>
				<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">Choose how you want to use Seva.</p>

				<div class="mt-6 grid gap-4">
					<button
						onclick={() => selectRole('customer')}
						class="flex items-center gap-4 rounded-lg border-2 border-gray-200 p-5 text-left transition hover:border-primary-500 hover:shadow-sm dark:border-gray-600 dark:hover:border-primary-500"
					>
						<div class="flex h-14 w-14 items-center justify-center rounded-xl bg-primary-100 text-primary-600 dark:bg-primary-900/30">
							<User class="h-7 w-7" />
						</div>
						<div>
							<h3 class="font-semibold text-gray-900 dark:text-white">I need a service</h3>
							<p class="mt-0.5 text-sm text-gray-500 dark:text-gray-400">Find and book trusted providers near you</p>
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
							<h3 class="font-semibold text-gray-900 dark:text-white">I provide services</h3>
							<p class="mt-0.5 text-sm text-gray-500 dark:text-gray-400">List your services and grow your business</p>
						</div>
					</button>
				</div>

			<!-- Step 2: Phone Verification -->
			{:else if step === 2}
				<div class="flex items-center gap-2 mb-4">
					<button onclick={() => { step = 1; selectedRole = null; otpSent = false; otpDigits = ['','','','','','']; }} class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
						<ArrowLeft class="h-5 w-5" />
					</button>
					<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Verify your phone</h1>
				</div>
				<div class="flex items-center gap-2 rounded-lg bg-gray-50 px-3 py-2 text-sm dark:bg-gray-700">
					{#if selectedRole === 'customer'}
						<User class="h-4 w-4 text-primary-600" />
						<span class="text-gray-700 dark:text-gray-300">Registering as Customer</span>
					{:else}
						<Briefcase class="h-4 w-4 text-secondary-600" />
						<span class="text-gray-700 dark:text-gray-300">Registering as Provider</span>
					{/if}
				</div>

				{#if !otpSent}
					<form onsubmit={sendOtp} class="mt-6">
						<label for="reg-phone" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Phone number</label>
						<div class="mt-1 flex gap-2">
							<select
								bind:value={countryCode}
								class="rounded-lg border border-gray-300 px-3 py-3 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
							>
								<option value="+91">IN +91</option>
								<option value="+44">UK +44</option>
								<option value="+1">US +1</option>
							</select>
							<input
								id="reg-phone"
								type="tel"
								bind:value={phone}
								placeholder="9876543210"
								required
								class="flex-1 rounded-lg border border-gray-300 px-4 py-3 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
							/>
						</div>
						<Button type="submit" variant="primary" size="lg" {loading} disabled={!phone.trim()} class="mt-4 w-full">
							Send Verification Code
						</Button>
					</form>
				{:else}
					<form onsubmit={verifyOtp} class="mt-6">
						<p class="text-sm text-gray-600 dark:text-gray-400">Enter the code sent to {countryCode}{phone}</p>
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
							Verify
						</Button>
						<div class="mt-3 text-center">
							<button
								type="button"
								onclick={() => { if (resendTimer <= 0) sendOtp(); }}
								disabled={resendTimer > 0}
								class="text-sm {resendTimer > 0 ? 'text-gray-400' : 'text-primary-600 hover:text-primary-700'}"
							>
								{resendTimer > 0 ? `Resend in ${resendTimer}s` : 'Resend OTP'}
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
						{selectedRole === 'provider' ? 'Provider Details' : 'Your Details'}
					</h1>
				</div>

				<form onsubmit={handleRegister} class="space-y-4">
					<Input label="Full Name" bind:value={name} required placeholder="Your full name" />
					<Input label="Email (optional)" type="email" bind:value={email} placeholder="you@example.com" />
					<Input label="Postcode" bind:value={postcode} required placeholder="560001" icon={MapPin} />

					{#if selectedRole === 'provider'}
						<div>
							<label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">Skills / Services</label>
							<div class="flex flex-wrap gap-2">
								{#each availableSkills as skill}
									<button
										type="button"
										onclick={() => toggleSkill(skill)}
										class="rounded-full border px-3 py-1.5 text-xs font-medium transition
											{selectedSkills.includes(skill)
												? 'border-primary-500 bg-primary-50 text-primary-700 dark:border-primary-400 dark:bg-primary-900/20 dark:text-primary-400'
												: 'border-gray-300 text-gray-600 hover:border-gray-400 dark:border-gray-600 dark:text-gray-400'}"
									>
										{skill}
									</button>
								{/each}
							</div>
						</div>

						<div>
							<label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
								Service Radius: {serviceRadius} km
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
							<label for="bio" class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">Bio</label>
							<textarea
								id="bio"
								bind:value={bio}
								rows="3"
								placeholder="Tell customers about yourself and your experience..."
								class="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400"
							></textarea>
						</div>
					{/if}

					<Button type="submit" variant="primary" size="lg" {loading} class="w-full">
						Create Account
					</Button>
				</form>

			<!-- Step 4: Success -->
			{:else if step === 4}
				<div class="text-center py-8">
					<div class="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-secondary-100 dark:bg-secondary-900/30">
						<CheckCircle class="h-8 w-8 text-secondary-600" />
					</div>
					<h1 class="mt-4 text-2xl font-bold text-gray-900 dark:text-white">Welcome to Seva!</h1>
					<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
						Your account has been created successfully. You're ready to
						{selectedRole === 'provider' ? 'start accepting jobs' : 'find service providers'}.
					</p>
					<div class="mt-6">
						<Button variant="primary" size="lg" href="/dashboard" class="w-full">
							Go to Dashboard
						</Button>
					</div>
				</div>
			{/if}

			{#if step < 4}
				<p class="mt-6 text-center text-sm text-gray-600 dark:text-gray-400">
					Already have an account?
					<a href="/login" class="font-medium text-primary-600 hover:text-primary-700">Sign in</a>
				</p>
			{/if}
		</div>
	</div>
</div>
