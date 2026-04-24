<script lang="ts">
	import { resolve } from '$app/paths';
	import Button from '$lib/components/ui/button.svelte';
	import TimeRangePicker from '$lib/components/gallery/time-range-picker.svelte';
	import ArrowLeft from '@lucide/svelte/icons/arrow-left';
	import { css } from 'styled-system/css';
	import type { PageData } from './$types';

	const { data }: { data: PageData } = $props();

	let filteredPhotos = $state(data.photos || []);
	let takenAfter = $state<string | null>(null);
	let takenBefore = $state<string | null>(null);

	function handleFilterChange(after: string | null, before: string | null) {
		takenAfter = after;
		takenBefore = before;
		updatePhotos();
	}

	async function updatePhotos() {
		const params = new URLSearchParams();
		if (takenAfter) params.set('taken_after', takenAfter);
		if (takenBefore) params.set('taken_before', takenBefore);

		const query = params.toString() ? `?${params.toString()}` : '';
		const response = await fetch(
			`/api/v1/events/${data.event.id}/galleries/${data.gallery.id}/photos${query}`
		);
		if (response.ok) {
			const result = await response.json();
			filteredPhotos = result.data || [];
		}
	}

	const wrapper = css({
		maxW: '7xl',
		display: 'flex',
		flexDirection: 'column',
		gap: '6',
	});

	const backRow = css({ mb: '1' });

	const header = css({
		display: 'flex',
		flexDirection: 'column',
		gap: '2',
	});

	const breadcrumb = css({
		fontSize: 'sm',
		color: 'fg.muted',
	});

	const title = css({
		fontSize: '2xl',
		fontWeight: 'bold',
		letterSpacing: 'tight',
		lineHeight: 'tight',
	});

	const description = css({
		color: 'fg.muted',
		fontSize: 'sm',
	});

	const photoGrid = css({
		display: 'grid',
		gridTemplateColumns: 'repeat(2, 1fr)',
		gap: '2',
		sm: { gridTemplateColumns: 'repeat(3, 1fr)' },
		md: { gridTemplateColumns: 'repeat(4, 1fr)' },
	});

	const photoCard = css({
		position: 'relative',
		aspectRatio: '1',
		bg: 'bg.muted',
		borderRadius: 'lg',
		overflow: 'hidden',
		cursor: 'pointer',
		transition: 'transform 0.2s',
		_hover: { transform: 'scale(1.05)' },
	});

	const photoImage = css({
		width: '100%',
		height: '100%',
		objectFit: 'cover',
	});

	const emptyState = css({
		textAlign: 'center',
		py: '12',
		color: 'fg.muted',
	});
</script>

<svelte:head>
	<title>{data.gallery.name} &mdash; {data.event.name} &mdash; MotoPhoto</title>
</svelte:head>

{#if data.event && data.gallery}
	<div class={wrapper}>
		<div class={backRow}>
			<Button
				variant="ghost"
				size="sm"
				href={resolve(`/events/${data.event.id}`)}
			>
				<ArrowLeft />
				Back to event
			</Button>
		</div>

		<div class={header}>
			<div class={breadcrumb}>
				{data.event.name}
			</div>
			<h1 class={title}>{data.gallery.name}</h1>
			{#if data.gallery.description}
				<p class={description}>{data.gallery.description}</p>
			{/if}
		</div>

		{#if data.gallery.earliest_photo_time && data.gallery.latest_photo_time}
			<TimeRangePicker
				minTime={data.gallery.earliest_photo_time}
				maxTime={data.gallery.latest_photo_time}
				onFilterChange={handleFilterChange}
			/>
		{/if}

		<div>
			{#if filteredPhotos.length > 0}
				<div class={photoGrid}>
					{#each filteredPhotos as photo (photo.id)}
						<div class={photoCard}>
							<img
								src={photo.preview_url}
								alt={photo.filename}
								class={photoImage}
								loading="lazy"
							/>
						</div>
					{/each}
				</div>
			{:else}
				<div class={emptyState}>
					<p>No photos found{takenAfter || takenBefore ? ' for this time range' : ''}.</p>
				</div>
			{/if}
		</div>
	</div>
{/if}
