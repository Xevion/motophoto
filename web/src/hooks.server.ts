import { env } from '$env/dynamic/private';
import type { Handle } from '@sveltejs/kit';

const BACKEND_URL = env.BACKEND_URL ?? 'http://localhost:3001';

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
				duplex: 'half'
			});

			return new Response(response.body, {
				status: response.status,
				statusText: response.statusText,
				headers: response.headers
			});
		} catch {
			return new Response(JSON.stringify({ error: 'Backend unavailable' }), {
				status: 502,
				headers: { 'Content-Type': 'application/json' }
			});
		}
	}

	return resolve(event);
};
