<script lang="ts">
import { resolve } from '$app/paths';
import Badge from '$lib/components/ui/badge.svelte';
import Button from '$lib/components/ui/button.svelte';
import ArrowLeft from '@lucide/svelte/icons/arrow-left';
import { css } from 'styled-system/css';
import type { PageData } from './$types';

const { data }: { data: PageData } = $props();

const event = data.event;

const wrapper = css({
	maxW: '3xl',
	display: 'flex',
	flexDirection: 'column',
	gap: '6',
});

const backRow = css({ mb: '1' });

const sportLabel = css({
	fontSize: 'xs',
	fontWeight: 'semibold',
	textTransform: 'uppercase',
	letterSpacing: 'wider',
	color: 'primary',
});

const title = css({
	fontSize: '3xl',
	fontWeight: 'bold',
	letterSpacing: 'tight',
	lineHeight: 'tight',
	mt: '1',
});

const meta = css({
	color: 'fg.muted',
	fontSize: 'sm',
	mt: '1',
});

const desc = css({
	color: 'fg.muted',
	lineHeight: 'relaxed',
});

const statsRow = css({
	display: 'flex',
	gap: '8',
});

const statBox = css({
	display: 'flex',
	flexDirection: 'column',
	gap: '0.5',
});

const statValue = css({
	fontSize: '2xl',
	fontWeight: 'bold',
});

const statLabel = css({
	fontSize: 'xs',
	color: 'fg.muted',
});

const tagRow = css({
	display: 'flex',
	flexWrap: 'wrap',
	gap: '2',
});

const galleryGrid = css({
	display: 'grid',
	gridTemplateColumns: 'repeat(2, 1fr)',
	gap: '2',
	sm: { gridTemplateColumns: 'repeat(3, 1fr)' },
	md: { gridTemplateColumns: 'repeat(4, 1fr)' },
});

const emptyState = css({
	color: 'fg.muted',
	padding: '4',
	border: '1px dashed',
	borderColor: 'border',
	borderRadius: 'lg',
});
</script>

<svelte:head>
	{#if event}
		<title>{event.name} — MotoPhoto</title>
	{/if}
</svelte:head>

{#if !event}
	<div class={emptyState}>
		Event not found.
	</div>
{:else}
	<div class={wrapper}>
		<div class={backRow}>
			<Button variant="ghost" size="sm" href={resolve('/')}>
				<ArrowLeft />
				All events
			</Button>
		</div>

		<header>
			<div class={sportLabel}>{event.sport}</div>
			<h1 class={title}>{event.name}</h1>

			<p class={meta}>
				{event.location}
				{event.date ? ' · ' : ''}
				{event.date
					? new Date(event.date).toLocaleDateString('en-US', {
							month: 'long',
							day: 'numeric',
							year: 'numeric'
						})
					: ''}
			</p>
		</header>

		{#if event.description}
			<p class={desc}>{event.description}</p>
		{/if}

		<div class={statsRow}>
			<div class={statBox}>
				<div class={statValue}>{event.photo_count ?? 0}</div>
				<div class={statLabel}>Photos</div>
			</div>

			<div class={statBox}>
				<div class={statValue}>{event.galleries?.length ?? 0}</div>
				<div class={statLabel}>
					{(event.galleries?.length ?? 0) === 1 ? 'Gallery' : 'Galleries'}
				</div>
			</div>
		</div>

		<div class={tagRow}>
			{#each event.tags ?? [] as tag (tag)}
				<Badge variant="secondary">{tag}</Badge>
			{/each}
		</div>

		<!-- GALLERIES -->
		<div>
			<h2>Galleries</h2>

			{#if event.galleries?.length}
				<div class={galleryGrid}>
					{#each event.galleries as gallery (gallery.id)}
						<a href={`/events/${event.slug}/galleries/${gallery.id}`}>
							<div>
								{#if gallery.cover_photo?.preview_url}
									<img
										src={gallery.cover_photo.preview_url}
										alt={gallery.name}
										style="width: 100%; aspect-ratio: 3/2; object-fit: cover; border-radius: 8px;"
									/>
								{:else}
									<div class={emptyState}>No image</div>
								{/if}
							</div>

							<p>{gallery.name}</p>
						</a>
					{/each}
				</div>
			{:else}
				<div class={emptyState}>
					No galleries in this event.
				</div>
			{/if}
		</div>
	</div>
{/if}
