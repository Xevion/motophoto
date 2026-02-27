# SvelteKit Frontend

The frontend is a SvelteKit 2 app using Svelte 5, managed with Bun. See [ARCHITECTURE.md](ARCHITECTURE.md) for overall system design and how the frontend connects to the Go backend.

## Project Structure

```
web/
+-- src/
|   +-- routes/                      # File-based routing
|   |   +-- +layout.svelte           # Root layout (theme, global styles)
|   |   +-- +page.svelte             # Home page -- event listing grid
|   |   +-- +page.server.ts          # Home page data loader
|   |   +-- events/
|   |       +-- [id]/
|   |           +-- +page.svelte     # Event detail page
|   |           +-- +page.server.ts  # Event detail data loader
|   +-- lib/
|   |   +-- api.ts                   # API client (apiFetch helper)
|   |   +-- utils.ts                 # Shared utility functions
|   |   +-- index.ts                 # Barrel export
|   |   +-- types.gen.ts             # Auto-generated Go -> TypeScript bindings (tygo)
|   |   +-- logger.client.ts         # Client-side logger (logtape)
|   |   +-- logger.server.ts         # Server-side logger (logtape)
|   |   +-- stores/
|   |   |   +-- theme.svelte.ts      # Dark/light theme store ($state rune)
|   |   +-- components/
|   |   |   +-- theme-toggle.svelte  # Dark/light mode toggle
|   |   |   +-- ui/                  # Reusable UI primitives (button, badge, ...)
|   |   |       +-- button.svelte
|   |   |       +-- badge.svelte
|   |   |       +-- index.ts
|   |   +-- recipes/                 # PandaCSS recipes for styled components
|   |       +-- button.ts
|   |       +-- badge.ts
|   |       +-- toggle.ts
|   +-- app.html                     # HTML shell
+-- static/                          # Static assets (favicon, etc.)
+-- styled-system/                   # PandaCSS generated output (DO NOT EDIT)
+-- panda.config.ts                  # PandaCSS design system config
+-- svelte.config.js                 # SvelteKit config (@xevion/svelte-adapter-bun)
+-- vite.config.ts                   # Vite config (API proxy)
+-- tsconfig.json
+-- eslint.config.js
+-- package.json
```

## Routing

SvelteKit uses file-based routing. Each route is a directory under `src/routes/` containing:

- **`+page.svelte`** -- the page component (what renders)
- **`+page.server.ts`** -- the server-side data loader (runs on the server, fetches data for the page)

| Route | URL | Description |
|-------|-----|-------------|
| `src/routes/` | `/` | Home page -- lists all events |
| `src/routes/events/[id]/` | `/events/:id` | Event detail page |

### Adding a New Route

1. Create a directory under `src/routes/`:

```
src/routes/galleries/
+-- +page.svelte          # Gallery listing page
+-- +page.server.ts       # Load function
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

- **In dev** -- Vite's proxy forwards `/api/*` to Go on `:3001`
- **In production** -- SvelteKit's `hooks.server.ts` reverse-proxies `/api/*` to Go on `:3001`

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

This project uses **Svelte 5** with runes -- the new reactivity system. Key differences from Svelte 4:

| Svelte 4 | Svelte 5 | Notes |
|----------|----------|-------|
| `export let prop` | `let { prop } = $props()` | Props are destructured from `$props()` |
| `$:` reactive | `$derived()`, `$effect()` | Explicit reactivity primitives |
| `let count = 0` (reactive) | `let count = $state(0)` | State must be explicitly declared |

See the [Svelte 5 docs](https://svelte.dev/docs/svelte) for the full runes API.

## Key Libraries

### PandaCSS

[PandaCSS](https://panda-css.com) is the CSS-in-JS-at-build-time framework used for styling. It generates a `styled-system/` directory with utility functions and CSS variables -- never edit that directory.

Config lives in `panda.config.ts`. The design tokens there (colors, spacing, radii) are the single source of truth for the visual language.

Recipes in `src/lib/recipes/` define multi-variant component styles (e.g., `button` has `variant` and `size` props). Apply recipes in Svelte components:

```svelte
<script lang="ts">
    import { css } from 'styled-system/css';
    import { button } from '$lib/recipes/button';
</script>

<button class={button({ variant: 'solid', size: 'md' })}>Click me</button>
```

Run `panda codegen` (via `bun run prepare`) to regenerate `styled-system/` after changing `panda.config.ts` or recipes.

### Ark UI

[Ark UI](https://ark-ui.com) provides unstyled, accessible headless components (dialogs, menus, popovers, etc.). Pair them with PandaCSS recipes for styled, accessible primitives:

```svelte
<script lang="ts">
    import * as Dialog from '@ark-ui/svelte/dialog';
</script>

<Dialog.Root>
    <Dialog.Trigger>Open</Dialog.Trigger>
    <Dialog.Backdrop />
    <Dialog.Positioner>
        <Dialog.Content>...</Dialog.Content>
    </Dialog.Positioner>
</Dialog.Root>
```

### Forms (superforms + zod)

[sveltekit-superforms](https://superforms.rocks) handles form state, validation, and progressive enhancement. [Zod](https://zod.dev) defines the schema:

```typescript
// +page.server.ts
import { superValidate } from 'sveltekit-superforms';
import { zod } from 'sveltekit-superforms/adapters';
import { z } from 'zod';

const schema = z.object({ name: z.string().min(1) });

export const load = async () => ({ form: await superValidate(zod(schema)) });

export const actions = {
    default: async ({ request }) => {
        const form = await superValidate(request, zod(schema));
        if (!form.valid) return fail(400, { form });
        // handle valid data
        return { form };
    },
};
```

### PhotoSwipe

[PhotoSwipe](https://photoswipe.com) is a lightbox / gallery viewer for displaying full-resolution photos. Used on event and gallery pages where customers browse photos for purchase.

### Logging (logtape)

[LogTape](https://logtape.org) provides structured logging in both server and client contexts. Use the pre-configured loggers from `$lib/logger.server.ts` (server-side) and `$lib/logger.client.ts` (browser). Never use `console.log` directly.

### Icons (Lucide)

Icons come from [`@lucide/svelte`](https://lucide.dev). Import individual icons to keep bundle size small:

```svelte
<script lang="ts">
    import { Camera, Calendar } from '@lucide/svelte';
</script>

<Camera size={16} />
```

### date-fns

[date-fns](https://date-fns.org) handles date formatting and arithmetic. Prefer it over `Date` methods for locale-aware formatting:

```typescript
import { format, formatDistanceToNow } from 'date-fns';
format(new Date(event.date), 'MMMM d, yyyy');
```

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

### Formatting

Biome (`@biomejs/biome`) handles code formatting and fast linting. Run `bun run format` (which calls `biome format --write .`) to format. ESLint is still used for Svelte-specific rules (`eslint-plugin-svelte`) that Biome doesn't cover.
