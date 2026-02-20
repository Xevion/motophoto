# SvelteKit Frontend

The frontend is a SvelteKit 2 app using Svelte 5, managed with Bun. See [ARCHITECTURE.md](ARCHITECTURE.md) for overall system design and how the frontend connects to the Go backend.

## Project Structure

```
web/
├── src/
│   ├── routes/                      # File-based routing
│   │   ├── +page.svelte             # Home page — event listing grid
│   │   ├── +page.server.ts          # Home page data loader
│   │   └── events/
│   │       └── [id]/
│   │           ├── +page.svelte     # Event detail page
│   │           └── +page.server.ts  # Event detail data loader
│   ├── lib/
│   │   ├── api.ts                   # API client (apiFetch helper)
│   │   └── components/              # Shared components
│   └── app.html                     # HTML shell
├── static/                          # Static assets (favicon, etc.)
├── svelte.config.js                 # SvelteKit config (@xevion/svelte-adapter-bun)
├── vite.config.ts                   # Vite config (API proxy)
├── tsconfig.json
├── eslint.config.js
└── package.json
```

## Routing

SvelteKit uses file-based routing. Each route is a directory under `src/routes/` containing:

- **`+page.svelte`** — the page component (what renders)
- **`+page.server.ts`** — the server-side data loader (runs on the server, fetches data for the page)

| Route | URL | Description |
|-------|-----|-------------|
| `src/routes/` | `/` | Home page — lists all events |
| `src/routes/events/[id]/` | `/events/:id` | Event detail page |

### Adding a New Route

1. Create a directory under `src/routes/`:

```
src/routes/galleries/
├── +page.svelte          # Gallery listing page
└── +page.server.ts       # Load function
```

2. Write the load function in `+page.server.ts`:

```typescript
import { apiFetch } from '$lib/api';
import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
    const response = await apiFetch('/api/v1/galleries');
    if (!response.ok) {
        throw error(response.status, 'Failed to load galleries');
    }
    return { galleries: await response.json() };
};
```

3. Use the data in `+page.svelte`:

```svelte
<script lang="ts">
    let { data } = $props();
</script>

<h1>Galleries</h1>
{#each data.galleries as gallery}
    <p>{gallery.title}</p>
{/each}
```

## Data Fetching

### apiFetch

All API calls go through the `apiFetch` helper in `src/lib/api.ts`. It uses relative URLs (e.g., `/api/v1/events`), which are routed to the Go backend automatically:

- **In dev** — Vite's proxy forwards `/api/*` to Go on `:3001`
- **In production** — SvelteKit's `hooks.server.ts` reverse-proxies `/api/*` to Go on `:3001`

```typescript
import { apiFetch } from '$lib/api';

const response = await apiFetch('/api/v1/events');
const events = await response.json();
```

### Load Functions

Data fetching happens in `+page.server.ts` load functions, which run on the server during SSR and on navigation. They return an object that becomes `data` in the corresponding `+page.svelte`.

```typescript
// +page.server.ts
export const load: PageServerLoad = async ({ fetch }) => {
    const [eventsData, health] = await Promise.all([
        apiFetch<EventsResponse>('/api/v1/events', fetch),
        apiFetch<HealthResponse>('/api/health', fetch),
    ]);
    return {
        events: eventsData.events,
        total: eventsData.total,
        backendStatus: health.status,
    };
};
```

## Svelte 5

This project uses **Svelte 5** with runes — the new reactivity system. Key differences from Svelte 4:

| Svelte 4 | Svelte 5 | Notes |
|----------|----------|-------|
| `export let prop` | `let { prop } = $props()` | Props are destructured from `$props()` |
| `$:` reactive | `$derived()`, `$effect()` | Explicit reactivity primitives |
| `let count = 0` (reactive) | `let count = $state(0)` | State must be explicitly declared |

See the [Svelte 5 docs](https://svelte.dev/docs/svelte) for the full runes API.

## Configuration

### Vite Proxy (`vite.config.ts`)

The dev server proxies API requests to the Go backend:

```typescript
server: {
    proxy: {
        '/api': { target: 'http://localhost:3001', changeOrigin: true },
    }
}
```

### SvelteKit Adapter

Uses `@xevion/svelte-adapter-bun` for server-side rendering via the Bun runtime. In production, the SvelteKit SSR server runs as a separate process alongside the Go backend, orchestrated by `web/entrypoint.ts`.
