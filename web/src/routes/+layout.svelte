<script lang="ts">
import '../app.css';
import favicon from '$lib/assets/favicon.svg';
import { resolve } from '$app/paths';
import { themeStore } from '$lib/stores/theme.svelte';
import ThemeToggle from '$lib/components/theme-toggle.svelte';
import type { Snippet } from 'svelte';

let { children }: { children: Snippet } = $props();

themeStore.init();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<div class="min-h-screen bg-background text-foreground">
	<header class="sticky top-0 z-50 w-full border-b border-border bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
		<div class="mx-auto flex h-14 max-w-5xl items-center justify-between px-4">
			<a href={resolve("/")} class="flex items-center gap-2 font-bold text-lg">
				<span class="bg-gradient-to-r from-primary to-red-500 bg-clip-text text-transparent">
					MotoPhoto
				</span>
			</a>
			<ThemeToggle />
		</div>
	</header>

	<main class="mx-auto max-w-5xl px-4 py-8">
		{@render children()}
	</main>
</div>
