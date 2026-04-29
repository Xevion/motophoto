<script lang="ts">
import type { PageData } from './$types';

const { data }: { data: PageData } = $props();

const { event, gallery } = data;
</script>

<svelte:head>
	{#if gallery}
		<title>{gallery.name} — {event?.name ?? 'Gallery'}</title>
	{/if}
</svelte:head>

{#if !event || !gallery}
	<div class="p-6 text-gray-500">
		Gallery not found.
	</div>
{:else}
	<div class="p-6 max-w-5xl mx-auto">

		<!-- HEADER -->
		<h1 class="text-2xl font-bold">{gallery.name}</h1>
		<p class="text-gray-500">{event.name}</p>

		<!-- COVER PHOTO -->
		{#if gallery.cover_photo?.preview_url}
			<img
				src={gallery.cover_photo.preview_url}
				alt={gallery.name}
				class="w-full mt-4 rounded-lg object-cover"
			/>
		{/if}

		<!-- PHOTOS GRID -->
		{#if gallery.photos?.length}
			<div class="grid grid-cols-2 md:grid-cols-3 gap-3 mt-6">
				{#each gallery.photos as photo (photo.id)}
					<img
						src={photo.preview_url}
						alt="photo"
						class="w-full aspect-square object-cover rounded"
					/>
				{/each}
			</div>
		{:else}
			<p class="text-gray-500 mt-6">No photos in this gallery.</p>
		{/if}

	</div>
{/if}