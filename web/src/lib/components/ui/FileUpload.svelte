<script lang="ts">
	import { Upload, X, Image } from 'lucide-svelte';

	interface Props {
		accept?: string;
		multiple?: boolean;
		maxSize?: number; // in MB
		onUpload?: (files: File[]) => void;
		class?: string;
	}

	let {
		accept = 'image/*',
		multiple = false,
		maxSize = 5,
		onUpload,
		class: className = ''
	}: Props = $props();

	let files = $state<File[]>([]);
	let previews = $state<string[]>([]);
	let dragOver = $state(false);
	let error = $state('');
	let fileInput: HTMLInputElement | undefined = $state();

	function handleFiles(fileList: FileList | null) {
		if (!fileList) return;
		error = '';
		const newFiles: File[] = [];

		for (const file of Array.from(fileList)) {
			if (maxSize && file.size > maxSize * 1024 * 1024) {
				error = `File "${file.name}" exceeds ${maxSize}MB limit`;
				continue;
			}
			newFiles.push(file);
		}

		if (multiple) {
			files = [...files, ...newFiles];
		} else {
			files = newFiles.slice(0, 1);
			// Revoke old previews
			previews.forEach((p) => URL.revokeObjectURL(p));
			previews = [];
		}

		// Generate previews for images
		for (const file of newFiles) {
			if (file.type.startsWith('image/')) {
				previews = [...previews, URL.createObjectURL(file)];
			}
		}

		onUpload?.(files);
	}

	function removeFile(index: number) {
		if (previews[index]) {
			URL.revokeObjectURL(previews[index]);
		}
		files = files.filter((_, i) => i !== index);
		previews = previews.filter((_, i) => i !== index);
		onUpload?.(files);
	}

	function handleDrop(e: DragEvent) {
		e.preventDefault();
		dragOver = false;
		handleFiles(e.dataTransfer?.files ?? null);
	}

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		dragOver = true;
	}

	function handleDragLeave() {
		dragOver = false;
	}
</script>

<div class={className}>
	<!-- Drop zone -->
	<div
		class="relative rounded-lg border-2 border-dashed p-6 text-center transition-colors
			{dragOver
				? 'border-primary-500 bg-primary-50 dark:bg-primary-900/10'
				: 'border-gray-300 hover:border-gray-400 dark:border-gray-600 dark:hover:border-gray-500'}"
		ondrop={handleDrop}
		ondragover={handleDragOver}
		ondragleave={handleDragLeave}
		role="button"
		tabindex="0"
		onclick={() => fileInput?.click()}
		onkeydown={(e) => e.key === 'Enter' && fileInput?.click()}
	>
		<Upload class="mx-auto h-8 w-8 text-gray-400" />
		<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
			Drag and drop files here, or <span class="font-medium text-primary-600">browse</span>
		</p>
		<p class="mt-1 text-xs text-gray-500 dark:text-gray-500">
			{accept === 'image/*' ? 'PNG, JPG, GIF' : accept} up to {maxSize}MB
		</p>
		<input
			bind:this={fileInput}
			type="file"
			{accept}
			{multiple}
			class="hidden"
			onchange={(e) => handleFiles((e.target as HTMLInputElement).files)}
		/>
	</div>

	{#if error}
		<p class="mt-2 text-xs text-red-600 dark:text-red-400">{error}</p>
	{/if}

	<!-- Preview thumbnails -->
	{#if previews.length > 0}
		<div class="mt-3 flex flex-wrap gap-2">
			{#each previews as preview, i}
				<div class="group relative h-20 w-20 overflow-hidden rounded-lg border border-gray-200 dark:border-gray-700">
					<img src={preview} alt="Preview {i + 1}" class="h-full w-full object-cover" />
					<button
						onclick={() => removeFile(i)}
						class="absolute right-0.5 top-0.5 rounded-full bg-black/60 p-0.5 text-white opacity-0 transition-opacity group-hover:opacity-100"
						aria-label="Remove file"
					>
						<X class="h-3.5 w-3.5" />
					</button>
				</div>
			{/each}
		</div>
	{:else if files.length > 0}
		<div class="mt-3 space-y-1">
			{#each files as file, i}
				<div class="flex items-center gap-2 rounded-lg bg-gray-50 px-3 py-2 text-sm dark:bg-gray-700">
					<Image class="h-4 w-4 text-gray-400" />
					<span class="flex-1 truncate text-gray-700 dark:text-gray-300">{file.name}</span>
					<button onclick={() => removeFile(i)} class="text-gray-400 hover:text-red-500">
						<X class="h-4 w-4" />
					</button>
				</div>
			{/each}
		</div>
	{/if}
</div>
