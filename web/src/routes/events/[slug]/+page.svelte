<script lang="ts">
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

const galleriesSection = css({
	mt: '6',
});

const galleriesTitle = css({
	fontSize: 'lg',
	fontWeight: 'semibold',
	mb: '3',
});

const emptyState = css({
	textAlign: 'center',
	color: 'fg.muted',
	py: '8',
	fontSize: 'sm',
});

const galleryLink = css({
	display: 'block',
	p: '3',
	borderRadius: 'md',
	border: '1px solid',
	borderColor: 'border',
	_hover: { bg: 'bg.muted' },
});

const galleryName = css({
	fontSize: 'sm',
	fontWeight: 'semibold',
});

const galleryMeta = css({
	fontSize: 'xs',
	color: 'fg.muted',
	mt: '1',
});

</script>

<svelte:head>
  <title>{data.event.name} &mdash; MotoPhoto</title>
</svelte:head>

{#if data.event}
  <div class={wrapper}>
    <div class={backRow}>
      <Button variant="ghost" size="sm" href="/">
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

    <!-- Galleries Section -->
    <div class={galleriesSection}>
      <h2 class={galleriesTitle}>Galleries</h2>
      {#if data.event.galleries && data.event.galleries.length > 0}
        <div style="display: flex; flex-direction: column; gap: 2;">
          {#each data.event.galleries as gallery (gallery.id)}
            <a href="/events/{data.event.slug}/galleries/{gallery.slug}" class={galleryLink}>
              <div class={galleryName}>{gallery.name}</div>
              {#if gallery.description}
                <p style="font-size: 0.875rem; color: var(--color-fg-muted); margin-top: 0.5rem;">
                  {gallery.description}
                </p>
              {/if}
              <div class={galleryMeta}>
                {gallery.photo_count.toLocaleString()} {gallery.photo_count === 1 ? 'photo' : 'photos'}
              </div>
            </a>
          {/each}
        </div>
      {:else}
        <div class={emptyState}>
          No galleries in this event yet
        </div>
      {/if}
    </div>
  </div>
{:else}
  <div class={emptyState}>
    No events yet
  </div>
{/if}
