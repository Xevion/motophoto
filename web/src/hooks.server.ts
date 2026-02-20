import { env } from '$env/dynamic/private';
import { initLogger } from '$lib/logger.server';
import { getLogger } from '@logtape/logtape';
import type { Handle } from '@sveltejs/kit';

await initLogger();

const BACKEND_URL = env.BACKEND_URL ?? 'http://localhost:3001';

const proxyLogger = getLogger(['ssr', 'proxy']);

/**
 * In dev, Vite's proxy handles /api forwarding.
 * In production, SvelteKit is the public-facing server and must forward
 * /api requests to the Go backend itself.
 */
export const handle: Handle = async ({ event, resolve }) => {
	const { pathname } = event.url;

	if (pathname.startsWith('/api/')) {
		const targetUrl = `${BACKEND_URL}${pathname}${event.url.search}`;

		const headers = new Headers(event.request.headers);
		headers.delete('host');

		try {
			const response = await fetch(targetUrl, {
				method: event.request.method,
				headers,
				body: event.request.body,
				// @ts-expect-error — Bun supports duplex streaming
				duplex: 'half',
			});

			return new Response(response.body, {
				status: response.status,
				statusText: response.statusText,
				headers: response.headers,
			});
		} catch (err) {
			proxyLogger.error('{method} {path} → backend unreachable', {
				method: event.request.method,
				path: pathname,
				error: err instanceof Error ? err.message : String(err),
			});
			return new Response(JSON.stringify({ error: 'Backend unavailable' }), {
				status: 502,
				headers: { 'Content-Type': 'application/json' },
			});
		}
	}

	return resolve(event);
};
