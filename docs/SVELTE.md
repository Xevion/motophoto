# SvelteKit Frontend

The frontend is a SvelteKit 2 app using Svelte 5, managed with Bun.

## Getting Started

### Prerequisites

- **Bun** — package manager and runtime: [bun.sh](https://bun.sh)

### Setup

```bash
cd web
bun install
```

### Running

```bash
# From project root — starts both frontend and backend
task dev

# From web/ — starts only the frontend
bun run dev
```

The dev server runs on **port 5173**. API requests (`/api/*`, `/health`) are proxied to the Go backend on port 8080 via Vite's proxy config in `vite.config.ts`.

### Type Checking & Linting

```bash
bun run check         # svelte-check (type errors)
bun run lint          # ESLint

# Or from project root
task check            # Runs both go vet + svelte-check
task lint             # Runs both golangci-lint + eslint
```

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
├── svelte.config.js                 # SvelteKit config (adapter-auto)
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

All API calls go through the `apiFetch` helper in `src/lib/api.ts`. It wraps `fetch` and points to the Go backend:

- **In the browser** — requests go to Vite's proxy (same origin), which forwards to `:8080`
- **In SSR (load functions)** — requests go directly to `http://localhost:8080`

```typescript
import { apiFetch } from '$lib/api';

const response = await apiFetch('/api/v1/events');
const events = await response.json();
```

### Load Functions

Data fetching happens in `+page.server.ts` load functions, which run on the server during SSR and on navigation. They return an object that becomes `data` in the corresponding `+page.svelte`.

```typescript
// +page.server.ts
export const load: PageServerLoad = async () => {
    const [eventsRes, healthRes] = await Promise.all([
        apiFetch('/api/v1/events'),
        apiFetch('/health'),
    ]);
    return {
        events: await eventsRes.json(),
        health: await healthRes.json(),
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
        '/api': 'http://localhost:8080',
        '/health': 'http://localhost:8080',
    }
}
```

If you add new top-level API paths (beyond `/api`), add them here too.

### SvelteKit Adapter

Currently uses `adapter-auto`, which auto-detects the deployment platform. In production (Railway), the SvelteKit output is built as static files served by the Go binary — the adapter config may change as the deployment story evolves.

## Building

```bash
bun run build         # Build SvelteKit for production
task build-frontend   # Same thing, from project root
```

The build output goes to `web/build/` and is copied into the Docker image during the production build.
