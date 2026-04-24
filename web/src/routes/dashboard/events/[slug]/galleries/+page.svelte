<script lang="ts">
import { css } from 'styled-system/css';
import Button from '$lib/components/ui/button.svelte';
import { Plus, Trash2, ArrowLeft, ChevronDown, ChevronUp } from '@lucide/svelte';
import type { PageData } from './$types';

const { data }: { data: PageData } = $props();

const pageContainer = css({
	display: 'flex',
	flexDirection: 'column',
	gap: '6',
});

const pageTitle = css({
	fontSize: '2xl',
	fontWeight: 'bold',
	display: 'flex',
	alignItems: 'center',
	gap: '2',
});

const breadcrumb = css({
	fontSize: 'sm',
	color: 'fg.muted',
});

const headerSection = css({
	display: 'flex',
	justifyContent: 'space-between',
	alignItems: 'center',
	gap: '4',
});

const galleryList = css({
	display: 'flex',
	flexDirection: 'column',
	gap: '2',
});

const galleryCard = css({
	bg: 'bg.subtle',
	borderWidth: '1px',
	borderColor: 'border',
	borderRadius: 'lg',
	p: '4',
	display: 'flex',
	justifyContent: 'space-between',
	alignItems: 'center',
	transition: 'all 150ms',
	_hover: { borderColor: 'primary', bg: 'bg.muted' },
});

const galleryInfo = css({
	flex: '1',
	display: 'flex',
	flexDirection: 'column',
	gap: '1',
});

const galleryName = css({
	fontWeight: 'semibold',
	fontSize: 'sm',
});

const galleryMeta = css({
	fontSize: 'xs',
	color: 'fg.muted',
});

const actions = css({
	display: 'flex',
	gap: '2',
	alignItems: 'center',
});

const emptyState = css({
	textAlign: 'center',
	py: '12',
	color: 'fg.muted',
	bg: 'bg.subtle',
	borderWidth: '1px',
	borderColor: 'border',
	borderRadius: 'lg',
	p: '6',
});

const emptyStateTitle = css({
	fontSize: 'lg',
	fontWeight: 'semibold',
	mb: '2',
	color: 'fg',
});
</script>

<div class={pageContainer}>
	<div class={pageTitle}>
		<a href="/dashboard">
			<Button variant="ghost" size="sm">
				<ArrowLeft />
			</Button>
		</a>
		<div>
			<div>{data.event.name}</div>
			<div class={breadcrumb}>Manage Galleries</div>
		</div>
	</div>

	<div class={headerSection}>
		<div>
			<h2 class={css({ fontSize: 'lg', fontWeight: 'semibold' })}>
				{data.event.galleries?.length ?? 0} {(data.event.galleries?.length ?? 0) === 1 ? 'Gallery' : 'Galleries'}
			</h2>
		</div>
		<a href="/dashboard/events/{data.event.slug}/galleries/new">
			<Button variant="solid" size="sm">
				<Plus />
				Create Gallery
			</Button>
		</a>
	</div>

	{#if data.event.galleries && data.event.galleries.length > 0}
		<div class={galleryList}>
			{#each data.event.galleries as gallery (gallery.id)}
				<div class={galleryCard}>
					<div class={galleryInfo}>
						<div class={galleryName}>{gallery.name}</div>
						{#if gallery.description}
							<div class={galleryMeta}>{gallery.description}</div>
						{/if}
						<div class={galleryMeta}>
							{gallery.photo_count.toLocaleString()} {gallery.photo_count === 1 ? 'photo' : 'photos'}
						</div>
					</div>
					<div class={actions}>
						<a href="/dashboard/events/{data.event.slug}/galleries/{gallery.slug}/edit">
							<Button variant="ghost" size="sm">Edit</Button>
						</a>
						<form method="POST" action="?/deleteGallery" style="display: inline;">
							<input type="hidden" name="galleryId" value={gallery.id} />
							<Button variant="ghost" size="sm" type="submit">
								<Trash2 />
							</Button>
						</form>
					</div>
				</div>
			{/each}
		</div>
	{:else}
		<div class={emptyState}>
			<p class={emptyStateTitle}>No galleries yet</p>
			<p>Create your first gallery to start uploading photos</p>
		</div>
	{/if}
</div>
