<script lang="ts">
import type { PageData } from './$types';
import { resolve } from '$app/paths';
import { Badge } from '$lib/components/ui/badge/index.js';
import * as Card from '$lib/components/ui/card/index.js';

let { data }: { data: PageData } = $props();
</script>

<svelte:head>
	<title>MotoPhoto — Event Photography Marketplace</title>
</svelte:head>

<div class="space-y-8">
	<section class="text-center space-y-3">
		<h1 class="text-4xl font-bold tracking-tight bg-gradient-to-r from-primary to-red-500 bg-clip-text text-transparent">
			MotoPhoto
		</h1>
		<p class="text-muted-foreground text-lg">
			Find your moment. Every event. Every angle.
		</p>
		<p class="text-sm text-muted-foreground">
			Backend:
			<Badge variant={data.backendStatus === "ok" ? "default" : "destructive"} class={data.backendStatus === "ok" ? "bg-green-900 text-green-400 hover:bg-green-900" : ""}>
				{data.backendStatus}
			</Badge>
		</p>
	</section>

	<section class="space-y-4">
		<h2 class="text-xl font-semibold">Upcoming Events ({data.total})</h2>

		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each data.events as event (event.id)}
				<a href={resolve(`/events/${event.id}`)} class="group block">
					<Card.Root class="h-full transition-colors hover:border-primary">
						<Card.Header class="pb-3">
							<div class="text-xs font-semibold uppercase tracking-wider text-primary">
								{event.sport}
							</div>
							<Card.Title class="text-base leading-snug">
								{event.name}
							</Card.Title>
							<Card.Description>
								{event.location}
							</Card.Description>
						</Card.Header>
						<Card.Content class="space-y-3 pt-0">
							<p class="text-sm text-muted-foreground">
								{new Date(event.date).toLocaleDateString("en-US", {
									month: "long",
									day: "numeric",
									year: "numeric",
								})}
							</p>
							<p class="text-sm text-muted-foreground line-clamp-2">
								{event.description}
							</p>
							<div class="flex gap-3 text-xs text-muted-foreground">
								<span>{event.photo_count.toLocaleString()} photos</span>
								<span>{event.galleries} {event.galleries === 1 ? "gallery" : "galleries"}</span>
							</div>
							<div class="flex flex-wrap gap-1.5">
								{#each event.tags as tag (tag)}
									<Badge variant="secondary" class="text-xs">{tag}</Badge>
								{/each}
							</div>
						</Card.Content>
					</Card.Root>
				</a>
			{/each}
		</div>
	</section>
</div>
