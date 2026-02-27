<script lang="ts">
import { resolve } from '$app/paths';
import Badge from '$lib/components/ui/badge.svelte';
import Button from '$lib/components/ui/button.svelte';
import ArrowLeft from '@lucide/svelte/icons/arrow-left';
import { css } from 'styled-system/css';
import type { PageData } from './$types';

const { data }: { data: PageData } = $props();

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

const skeletonPhoto = css({
	aspectRatio: '3/2',
	bg: 'bg.muted',
	borderRadius: 'lg',
	animation: 'pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite',
});
</script>

<svelte:head>
  <title>{data.event ? data.event.name : 'Event Not Found'} &mdash; MotoPhoto</title>
</svelte:head>

{#if data.event}
  <div class={wrapper}>
    <div class={backRow}>
      <Button variant="ghost" size="sm" href={resolve("/")}>
        <ArrowLeft />
        All events
      </Button>
    </div>

    <header>
      <div class={sportLabel}>{data.event.sport}</div>
      <h1 class={title}>{data.event.name}</h1>
      <p class={meta}>
        {data.event.location}{data.event.date ? ' \u00B7 ' : ''}
        {data.event.date ? new Date(data.event.date).toLocaleDateString('en-US', {
          month: 'long',
          day: 'numeric',
          year: 'numeric',
        }) : ''}
      </p>
    </header>

    <p class={desc}>{data.event.description}</p>

    <div class={statsRow}>
      <div class={statBox}>
        <div class={statValue}>{data.event.photo_count.toLocaleString()}</div>
        <div class={statLabel}>Photos</div>
      </div>
      <div class={statBox}>
        <div class={statValue}>{data.event.galleries?.length ?? 0}</div>
        <div class={statLabel}>{(data.event.galleries?.length ?? 0) === 1 ? 'Gallery' : 'Galleries'}</div>
      </div>
    </div>

    <div class={tagRow}>
      {#each data.event.tags as tag (tag)}
        <Badge variant="secondary">{tag}</Badge>
      {/each}
    </div>

    <!-- Skeleton gallery grid -->
    <div>
      <div class={galleryGrid}>
        {#each { length: 8 } as _, i (i)}
          <div class={skeletonPhoto}></div>
        {/each}
      </div>
    </div>
  </div>
{:else}
  <div class={css({ textAlign: 'center', py: '20', color: 'fg.muted' })}>
    <p class={css({ fontSize: '2xl', fontWeight: 'semibold', mb: '2' })}>Event not found</p>
    <p>That event doesn't exist or may have been removed.</p>
    <div class={css({ mt: '6' })}>
      <Button href={resolve("/")}>
        <ArrowLeft />
        Back to events
      </Button>
    </div>
  </div>
{/if}
