<script lang="ts">
import { resolve } from '$app/paths';
import Badge from '$lib/components/ui/badge.svelte';
import { css } from 'styled-system/css';
import UiSelect from '$lib/components/ui/select.svelte';
import ArrowUpDown from '@lucide/svelte/icons/arrow-up-down';
import type { PageData } from './$types';
import type { EventResponse } from '$lib/types.gen';

const { data }: { data: PageData } = $props();

type SortKey = 'date-desc' | 'date-asc' | 'name-asc' | 'photos-desc';

const sortItems = [
	{ value: 'date-desc', label: 'Newest first' },
	{ value: 'date-asc', label: 'Oldest first' },
	{ value: 'name-asc', label: 'Name A\u2013Z' },
	{ value: 'photos-desc', label: 'Most photos' },
];

let sortValue = $state<string[]>(['date-desc']);
let sortKey = $derived((sortValue[0] as SortKey) ?? 'date-desc');

function compareEvents(a: EventResponse, b: EventResponse): number {
	switch (sortKey) {
		case 'date-desc':
			return (b.date ?? '').localeCompare(a.date ?? '');
		case 'date-asc':
			return (a.date ?? '').localeCompare(b.date ?? '');
		case 'name-asc':
			return a.name.localeCompare(b.name);
		case 'photos-desc':
			return b.photo_count - a.photo_count;
		default:
			return 0;
	}
}

const sortedEvents = $derived([...data.events].sort(compareEvents));

const page = css({
	display: 'flex',
	flexDirection: 'column',
	gap: '10',
});

const hero = css({
	textAlign: 'center',
	display: 'flex',
	flexDirection: 'column',
	gap: '3',
	py: '6',
});

const heroTitle = css({
	fontSize: '4xl',
	fontWeight: 'bold',
	letterSpacing: 'tight',
	background: 'linear-gradient(to right, {colors.orange.500}, {colors.orange.700})',
	backgroundClip: 'text',
	color: 'transparent',
});

const heroSub = css({
	color: 'fg.muted',
	fontSize: 'lg',
});

const sectionHeader = css({
	display: 'flex',
	alignItems: 'center',
	justifyContent: 'space-between',
	gap: '3',
});

const sectionLeft = css({
	display: 'flex',
	alignItems: 'baseline',
	gap: '3',
});

const sectionTitle = css({
	fontSize: 'xl',
	fontWeight: 'semibold',
});

const sectionCount = css({
	fontSize: 'sm',
	color: 'fg.muted',
});

const grid = css({
	display: 'grid',
	gridTemplateColumns: 'repeat(1, 1fr)',
	gap: '4',
	sm: { gridTemplateColumns: 'repeat(2, 1fr)' },
	lg: { gridTemplateColumns: 'repeat(3, 1fr)' },
});

const cardLink = css({
	display: 'block',
	textDecoration: 'none',
	borderRadius: 'xl',
	outline: 'none',
	_focusVisible: {
		outlineWidth: '2px',
		outlineStyle: 'solid',
		outlineColor: 'primary',
		outlineOffset: '2px',
	},
});

const card = css({
	h: 'full',
	bg: 'bg.subtle',
	borderWidth: '1px',
	borderColor: 'border',
	borderRadius: 'xl',
	p: '5',
	display: 'flex',
	flexDirection: 'column',
	gap: '3',
	transition: 'all',
	transitionDuration: '150ms',
	_hover: {
		borderColor: 'primary',
		bg: 'bg.muted',
	},
});

const sportLabel = css({
	fontSize: 'xs',
	fontWeight: 'semibold',
	textTransform: 'uppercase',
	letterSpacing: 'wider',
	color: 'primary',
});

const cardTitle = css({
	fontSize: 'md',
	fontWeight: 'semibold',
	lineHeight: 'snug',
	color: 'fg',
	mt: '1',
});

const cardLocation = css({
	fontSize: 'sm',
	color: 'fg.muted',
});

const cardDesc = css({
	fontSize: 'sm',
	color: 'fg.muted',
	lineHeight: 'relaxed',
	lineClamp: 2,
	flexGrow: '1',
});

const cardMeta = css({
	display: 'flex',
	gap: '4',
	fontSize: 'xs',
	color: 'fg.muted',
});

const tagRow = css({
	display: 'flex',
	flexWrap: 'wrap',
	gap: '1.5',
});

const eventListWrapper = css({
	mt: '4',
});

const sortTrigger = css({
	display: 'inline-flex',
	alignItems: 'center',
	gap: '1.5',
	borderRadius: 'md',
	fontSize: 'sm',
	fontWeight: 'medium',
	cursor: 'pointer',
	h: '8',
	px: '3',
	bg: 'transparent',
	color: 'fg.muted',
	borderWidth: '1px',
	borderColor: 'border',
	transition: 'all 150ms',
	_hover: { bg: 'bg.muted', color: 'fg' },
	_focusVisible: {
		outlineWidth: '2px',
		outlineColor: 'primary',
		outlineOffset: '2px',
		outlineStyle: 'solid',
	},
	'& svg': { pointerEvents: 'none', flexShrink: 0, width: '1em', height: '1em' },
});
</script>

<svelte:head>
  <title>MotoPhoto &mdash; Event Photography Marketplace</title>
</svelte:head>

<div class={page}>
  <section class={hero}>
    <h1 class={heroTitle}>MotoPhoto</h1>
    <p class={heroSub}>Find your moment. Every event. Every angle.</p>
  </section>

  <section>
    <div class={sectionHeader}>
      <div class={sectionLeft}>
        <h2 class={sectionTitle}>Upcoming Events</h2>
        <span class={sectionCount}>{data.events.length} events</span>
      </div>

      <UiSelect
        items={sortItems}
        value={sortValue}
        onValueChange={(v: string[]) => (sortValue = v)}
        placeholder="Sort"
        triggerClass={sortTrigger}
      >
        {#snippet icon()}
          <ArrowUpDown />
        {/snippet}
      </UiSelect>
    </div>

    <div class={eventListWrapper}>
      <div class={grid}>
        {#each sortedEvents as event (event.id)}
          <a href={resolve('/events/[id]', { id: String(event.id) })} class={cardLink}>
            <div class={card}>
              <div>
                <div class={sportLabel}>{event.sport}</div>
                <div class={cardTitle}>{event.name}</div>
                <div class={cardLocation}>{event.location}</div>
              </div>

              <p class={cardDesc}>{event.description}</p>

              <div class={cardMeta}>
                {#if event.date}
                  <span>
                    {new Date(event.date).toLocaleDateString('en-US', {
                      month: 'short',
                      day: 'numeric',
                      year: 'numeric',
                    })}
                  </span>
                {/if}
                <span>{event.photo_count.toLocaleString()} photos</span>
                <span>{event.galleries?.length ?? 0} {(event.galleries?.length ?? 0) === 1 ? 'gallery' : 'galleries'}</span>
              </div>

              <div class={tagRow}>
                {#each event.tags as tag (tag)}
                  <Badge variant="secondary">{tag}</Badge>
                {/each}
              </div>
            </div>
          </a>
        {/each}
      </div>
    </div>
  </section>
</div>
