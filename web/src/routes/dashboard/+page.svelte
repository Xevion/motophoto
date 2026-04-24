<script lang="ts">
import { css, cx } from 'styled-system/css';
import Button from '$lib/components/ui/button.svelte';
import Badge from '$lib/components/ui/badge.svelte';
import UiSelect from '$lib/components/ui/select.svelte';
import { Plus, Pencil, Trash2, Archive, ArrowUpDown } from '@lucide/svelte';
import type { PageData } from './$types';
import type { EventResponse } from '$lib/types.gen';

const { data }: { data: PageData } = $props();

type SortKey = 'date-desc' | 'date-asc' | 'name-asc';
type StatusFilter = 'all' | 'draft' | 'published' | 'archived';

const sortItems = [
	{ value: 'date-desc', label: 'Newest first' },
	{ value: 'date-asc', label: 'Oldest first' },
	{ value: 'name-asc', label: 'Name A–Z' },
];

const statusItems = [
	{ value: 'all', label: 'All status' },
	{ value: 'draft', label: 'Draft' },
	{ value: 'published', label: 'Published' },
	{ value: 'archived', label: 'Archived' },
];

let sortValue = $state<string[]>(['date-desc']);
let statusValue = $state<string[]>(['all']);

let sortKey = $derived((sortValue[0] as SortKey) ?? 'date-desc');
let statusFilter = $derived((statusValue[0] as StatusFilter) ?? 'all');

function compareEvents(a: EventResponse, b: EventResponse): number {
	switch (sortKey) {
		case 'date-desc':
			return (b.date ?? '').localeCompare(a.date ?? '');
		case 'date-asc':
			return (a.date ?? '').localeCompare(b.date ?? '');
		case 'name-asc':
			return a.name.localeCompare(b.name);
		default:
			return 0;
	}
}

const filteredEvents = $derived(
	data.events
		.filter((e) => statusFilter === 'all' || e.status === statusFilter)
		.sort(compareEvents)
);

const pageTitle = css({
	fontSize: '2xl',
	fontWeight: 'bold',
	mb: '6',
});

const controlsSection = css({
	display: 'flex',
	gap: '4',
	mb: '6',
	alignItems: 'center',
	justifyContent: 'space-between',
});

const leftControls = css({
	display: 'flex',
	gap: '3',
	alignItems: 'center',
});

const rightControls = css({
	display: 'flex',
	gap: '2',
});

const selectTrigger = css({
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

const table = css({
	w: 'full',
	borderCollapse: 'collapse',
	borderRadius: 'lg',
	overflow: 'hidden',
	border: '1px solid',
	borderColor: 'border',
});

const th = css({
	bg: 'bg.muted',
	px: '4',
	py: '3',
	textAlign: 'left',
	fontSize: 'sm',
	fontWeight: 'semibold',
	color: 'fg.muted',
	borderBottomWidth: '1px',
	borderBottomColor: 'border',
});

const td = css({
	px: '4',
	py: '3',
	borderBottomWidth: '1px',
	borderBottomColor: 'border',
	fontSize: 'sm',
});

const tdLast = css({
	borderBottom: 'none',
});

const eventNameCell = css({
	fontWeight: 'medium',
});

const statusBadge = css({
	display: 'inline-flex',
	alignItems: 'center',
});

const actionsCell = css({
	display: 'flex',
	gap: '2',
	justifyContent: 'flex-end',
});

const emptyState = css({
	textAlign: 'center',
	py: '12',
	color: 'fg.muted',
});

const emptyStateTitle = css({
	fontSize: 'lg',
	fontWeight: 'semibold',
	mb: '2',
	color: 'fg',
});

const statusVariantMap = {
	draft: 'secondary',
	published: 'default',
	archived: 'outline',
} as const;
</script>

<div>
	<h1 class={pageTitle}>My Events</h1>

	<div class={controlsSection}>
		<div class={leftControls}>
			<UiSelect
				items={statusItems}
				value={statusValue}
				onValueChange={(v: string[]) => (statusValue = v)}
				triggerClass={selectTrigger}
			>
				{#snippet icon()}
					<ArrowUpDown />
				{/snippet}
			</UiSelect>

			<UiSelect
				items={sortItems}
				value={sortValue}
				onValueChange={(v: string[]) => (sortValue = v)}
				triggerClass={selectTrigger}
			>
				{#snippet icon()}
					<ArrowUpDown />
				{/snippet}
			</UiSelect>
		</div>

		<div class={rightControls}>
			<a href="/dashboard/events/new">
				<Button variant="solid" size="sm">
					<Plus />
					Create Event
				</Button>
			</a>
		</div>
	</div>

	{#if filteredEvents.length === 0}
		<div class={emptyState}>
			<p class={emptyStateTitle}>No events yet</p>
			<p>Create your first event to get started</p>
		</div>
	{:else}
		<table class={table}>
			<thead>
				<tr>
					<th class={th}>Event</th>
					<th class={th}>Date</th>
					<th class={th}>Photos</th>
					<th class={th}>Galleries</th>
					<th class={th}>Status</th>
					<th class={th} style="text-align: right;">Actions</th>
				</tr>
			</thead>
			<tbody>
				{#each filteredEvents as event (event.id)}
					<tr>
						<td class={cx(td, eventNameCell)}>
							<div>{event.name}</div>
							<div style="font-size: 0.875rem; color: var(--color-fg-muted);">
								{event.location || 'No location'}
							</div>
						</td>
						<td class={td}>
							{#if event.date}
								{new Date(event.date).toLocaleDateString('en-US', {
									month: 'short',
									day: 'numeric',
									year: 'numeric',
								})}
							{:else}
								—
							{/if}
						</td>
						<td class={td}>{event.photo_count.toLocaleString()}</td>
						<td class={td}>{event.galleries?.length ?? 0}</td>
						<td class={cx(td, statusBadge)}>
							<Badge
								variant={statusVariantMap[event.status as keyof typeof statusVariantMap] ||
									'secondary'}
							>
								{event.status}
							</Badge>
						</td>
						<td class={cx(td, actionsCell)}>
							<a href="/dashboard/events/{event.slug}/galleries">
								<Button variant="ghost" size="sm">View galleries</Button>
							</a>
							<a href="/dashboard/events/{event.slug}/edit">
								<Button variant="ghost" size="sm">
									<Pencil />
								</Button>
							</a>
							<form method="POST" action="?/delete" style="display: inline;">
								<input type="hidden" name="eventId" value={event.id} />
								<Button variant="ghost" size="sm" type="submit">
									<Trash2 />
								</Button>
							</form>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}
</div>
