<script lang="ts">
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<svelte:head>
	<title>MotoPhoto — Event Photography Marketplace</title>
</svelte:head>

<main>
	<header>
		<h1>MotoPhoto</h1>
		<p class="tagline">Find your moment. Every event. Every angle.</p>
		<p class="status">
			Backend: <span class="badge" class:ok={data.backendStatus === 'ok'}>{data.backendStatus}</span>
		</p>
	</header>

	<section class="events">
		<h2>Upcoming Events ({data.total})</h2>
		<div class="grid">
			{#each data.events as event}
				<a href="/events/{event.id}" class="card">
					<div class="card-sport">{event.sport}</div>
					<h3>{event.name}</h3>
					<p class="meta">{event.location}</p>
					<p class="meta">{new Date(event.date).toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' })}</p>
					<p class="description">{event.description}</p>
					<div class="stats">
						<span>{event.photo_count.toLocaleString()} photos</span>
						<span>{event.galleries} {event.galleries === 1 ? 'gallery' : 'galleries'}</span>
					</div>
					<div class="tags">
						{#each event.tags as tag}
							<span class="tag">{tag}</span>
						{/each}
					</div>
				</a>
			{/each}
		</div>
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
		max-width: 960px;
		margin: 0 auto;
		padding: 2rem 1rem;
	}

	header {
		text-align: center;
		margin-bottom: 3rem;
	}

	h1 {
		font-size: 2.5rem;
		margin: 0;
		background: linear-gradient(135deg, #f97316, #ef4444);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	.tagline {
		color: #a3a3a3;
		font-size: 1.1rem;
		margin-top: 0.5rem;
	}

	.status {
		font-size: 0.85rem;
		color: #737373;
	}

	.badge {
		padding: 0.15rem 0.5rem;
		border-radius: 4px;
		background: #333;
		font-family: monospace;
	}

	.badge.ok {
		background: #14532d;
		color: #4ade80;
	}

	h2 {
		font-size: 1.4rem;
		margin-bottom: 1.5rem;
		color: #d4d4d4;
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 1.25rem;
	}

	.card {
		background: #171717;
		border: 1px solid #262626;
		border-radius: 8px;
		padding: 1.25rem;
		text-decoration: none;
		color: inherit;
		transition: border-color 0.15s;
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
	}

	.card:hover {
		border-color: #f97316;
	}

	.card-sport {
		font-size: 0.75rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: #f97316;
		font-weight: 600;
	}

	.card h3 {
		margin: 0;
		font-size: 1.1rem;
	}

	.meta {
		margin: 0;
		font-size: 0.85rem;
		color: #737373;
	}

	.description {
		margin: 0.25rem 0;
		font-size: 0.9rem;
		color: #a3a3a3;
		line-height: 1.4;
	}

	.stats {
		display: flex;
		gap: 1rem;
		font-size: 0.8rem;
		color: #a3a3a3;
		margin-top: 0.5rem;
	}

	.tags {
		display: flex;
		flex-wrap: wrap;
		gap: 0.4rem;
		margin-top: 0.5rem;
	}

	.tag {
		font-size: 0.7rem;
		padding: 0.15rem 0.5rem;
		background: #1e1e1e;
		border: 1px solid #333;
		border-radius: 99px;
		color: #a3a3a3;
	}
</style>
