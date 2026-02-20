<script lang="ts">
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<svelte:head>
	<title>{data.event.name} — MotoPhoto</title>
</svelte:head>

<main>
	<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
	<a href="/" class="back">&larr; All events</a>

	<header>
		<span class="sport">{data.event.sport}</span>
		<h1>{data.event.name}</h1>
		<p class="meta">
			{data.event.location} &middot;
			{new Date(data.event.date).toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' })}
		</p>
	</header>

	<p class="description">{data.event.description}</p>

	<div class="stats">
		<div class="stat">
			<span class="stat-value">{data.event.photo_count.toLocaleString()}</span>
			<span class="stat-label">Photos</span>
		</div>
		<div class="stat">
			<span class="stat-value">{data.event.galleries}</span>
			<span class="stat-label">{data.event.galleries === 1 ? 'Gallery' : 'Galleries'}</span>
		</div>
	</div>

	<div class="tags">
		{#each data.event.tags as tag (tag)}
			<span class="tag">{tag}</span>
		{/each}
	</div>

	<section class="placeholder">
		<p>Gallery grid will go here once photo uploads are wired up.</p>
	</section>
</main>

<style>
	:global(body) {
		margin: 0;
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
		background: #0a0a0a;
		color: #e5e5e5;
	}

	main {
		max-width: 720px;
		margin: 0 auto;
		padding: 2rem 1rem;
	}

	.back {
		color: #737373;
		text-decoration: none;
		font-size: 0.9rem;
	}

	.back:hover {
		color: #f97316;
	}

	header {
		margin-top: 1.5rem;
	}

	.sport {
		font-size: 0.75rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: #f97316;
		font-weight: 600;
	}

	h1 {
		font-size: 2rem;
		margin: 0.25rem 0;
	}

	.meta {
		color: #737373;
		font-size: 0.95rem;
	}

	.description {
		color: #a3a3a3;
		line-height: 1.5;
		margin: 1.5rem 0;
	}

	.stats {
		display: flex;
		gap: 2rem;
		margin-bottom: 1.5rem;
	}

	.stat {
		display: flex;
		flex-direction: column;
	}

	.stat-value {
		font-size: 1.5rem;
		font-weight: 700;
	}

	.stat-label {
		font-size: 0.8rem;
		color: #737373;
	}

	.tags {
		display: flex;
		flex-wrap: wrap;
		gap: 0.4rem;
		margin-bottom: 2rem;
	}

	.tag {
		font-size: 0.75rem;
		padding: 0.2rem 0.6rem;
		background: #1e1e1e;
		border: 1px solid #333;
		border-radius: 99px;
		color: #a3a3a3;
	}

	.placeholder {
		padding: 3rem;
		text-align: center;
		border: 1px dashed #333;
		border-radius: 8px;
		color: #525252;
	}
</style>
