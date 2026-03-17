<script lang="ts">
	import { Mic, MicOff } from 'lucide-svelte';

	interface Props {
		lang?: string;
		onresult?: (text: string) => void;
		onerror?: (error: string) => void;
		class?: string;
		size?: 'sm' | 'md' | 'lg';
	}

	let {
		lang = 'en-US',
		onresult,
		onerror,
		class: className = '',
		size = 'md'
	}: Props = $props();

	let supported = $state(false);
	let listening = $state(false);
	let recognition: any = null;

	const buttonSizeClasses = $derived({
		sm: 'h-8 w-8',
		md: 'h-10 w-10',
		lg: 'h-12 w-12'
	}[size]);

	const iconSizeClasses = $derived({
		sm: 'h-4 w-4',
		md: 'h-5 w-5',
		lg: 'h-6 w-6'
	}[size]);

	$effect(() => {
		if (typeof window !== 'undefined') {
			const SpeechRecognition =
				(window as any).SpeechRecognition || (window as any).webkitSpeechRecognition;
			if (SpeechRecognition) {
				supported = true;
				recognition = new SpeechRecognition();
				recognition.interimResults = false;
				recognition.continuous = false;
			}
		}

		return () => {
			if (recognition && listening) {
				recognition.abort();
			}
		};
	});

	function toggle() {
		if (!recognition) return;

		if (listening) {
			recognition.stop();
			listening = false;
			return;
		}

		recognition.lang = lang;

		recognition.onresult = (event: any) => {
			const transcript = event.results[0][0].transcript;
			listening = false;
			onresult?.(transcript);
		};

		recognition.onerror = (event: any) => {
			listening = false;
			onerror?.(event.error);
		};

		recognition.onend = () => {
			listening = false;
		};

		recognition.start();
		listening = true;
	}
</script>

{#if supported}
	<div class="relative inline-flex items-center justify-center {className}">
		<!-- Pulsing ring when listening -->
		{#if listening}
			<span
				class="absolute inset-0 rounded-full bg-red-400 opacity-75 animate-ping dark:bg-red-500"
			></span>
		{/if}

		<button
			type="button"
			onclick={toggle}
			aria-label={listening ? 'Stop listening' : 'Start voice input'}
			class="relative inline-flex items-center justify-center rounded-full border transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2
				{buttonSizeClasses}
				{listening
					? 'border-red-300 bg-red-50 text-primary-600 hover:bg-red-100 focus:ring-red-500 dark:border-red-600 dark:bg-red-900/20 dark:text-primary-400 dark:hover:bg-red-900/30'
					: 'border-gray-300 bg-white text-gray-500 hover:bg-gray-50 hover:text-gray-700 focus:ring-primary-500 dark:border-gray-600 dark:bg-gray-800 dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-gray-300'}"
		>
			{#if listening}
				<MicOff class={iconSizeClasses} />
			{:else}
				<Mic class={iconSizeClasses} />
			{/if}
		</button>
	</div>
{/if}
