<script lang="ts">
	import type { PageData } from "./$types";
	import { resolve } from "$app/paths";
	import { Badge } from "$lib/components/ui/badge/index.js";
	import { Button } from "$lib/components/ui/button/index.js";
	import ArrowLeft from "@lucide/svelte/icons/arrow-left";

	let { data }: { data: PageData } = $props();
</script>

<svelte:head>
	<title>{data.event.name} — MotoPhoto</title>
</svelte:head>

<div class="max-w-3xl space-y-6">
	<div>
		<Button variant="ghost" size="sm" href={resolve("/")} class="gap-1 text-muted-foreground hover:text-primary">
			<ArrowLeft class="h-4 w-4" />
			All events
		</Button>
	</div>

	<header class="space-y-2">
		<div class="text-xs font-semibold uppercase tracking-wider text-primary">
			{data.event.sport}
		</div>
		<h1 class="text-3xl font-bold tracking-tight">
			{data.event.name}
		</h1>
		<p class="text-muted-foreground">
			{data.event.location} &middot;
			{new Date(data.event.date).toLocaleDateString("en-US", {
				month: "long",
				day: "numeric",
				year: "numeric",
			})}
		</p>
	</header>

	<p class="text-muted-foreground leading-relaxed">
		{data.event.description}
	</p>

	<div class="flex gap-8">
		<div class="space-y-0.5">
			<div class="text-2xl font-bold">{data.event.photo_count.toLocaleString()}</div>
			<div class="text-xs text-muted-foreground">Photos</div>
		</div>
		<div class="space-y-0.5">
			<div class="text-2xl font-bold">{data.event.galleries}</div>
			<div class="text-xs text-muted-foreground">{data.event.galleries === 1 ? "Gallery" : "Galleries"}</div>
		</div>
	</div>

	<div class="flex flex-wrap gap-2">
		{#each data.event.tags as tag (tag)}
			<Badge variant="secondary">{tag}</Badge>
		{/each}
	</div>

	<div class="rounded-lg border border-dashed border-border p-12 text-center text-muted-foreground">
		<p>Gallery grid will go here once photo uploads are wired up.</p>
	</div>
</div>
