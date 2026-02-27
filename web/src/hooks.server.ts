import { env } from '$env/dynamic/private';
import { initLogger } from '$lib/logger.server';
import { getLogger } from '@logtape/logtape';
import type { Handle, HandleFetch } from '@sveltejs/kit';

await initLogger();

const BACKEND_URL = env.BACKEND_URL ?? 'http://localhost:3001';

const proxyLogger = getLogger(['ssr', 'proxy']);

// Headers from the incoming request that are safe to forward to the Go backend.
// Railway's edge proxy controls X-Real-Ip (strips any client-provided value and
// sets it to the actual client IP) and appends to X-Forwarded-For, so both can
// be trusted when present. Forwarding only an explicit list prevents arbitrary
// client headers from reaching the backend if Railway's edge is bypassed.
const FORWARDED_REQUEST_HEADERS = [
	'x-railway-request-id',
	'x-request-id',
	'x-real-ip',
	'x-forwarded-for',
	'x-forwarded-proto',
	'content-type',
	'accept',
	'authorization',
	'cookie',
] as const;

/**
 * In dev, Vite's proxy handles /api forwarding.
 * In production, SvelteKit is the public-facing server and must forward
 * /api requests to the Go backend itself.
 */
export const handle: Handle = async ({ event, resolve }) => {
	// Capture the Railway request ID early so handleFetch can propagate it to
	// backend fetches made during SSR page loads.
	event.locals.requestId =
		event.request.headers.get('x-railway-request-id') ??
		event.request.headers.get('x-request-id') ??
		crypto.randomUUID();

	const { pathname } = event.url;

	if (pathname.startsWith('/api/')) {
		const targetUrl = `${BACKEND_URL}${pathname}${event.url.search}`;

		const headers = new Headers();
		for (const name of FORWARDED_REQUEST_HEADERS) {
			const value = event.request.headers.get(name);
			if (value !== null) headers.set(name, value);
		}

		try {
			const response = await fetch(targetUrl, {
				method: event.request.method,
				headers,
				body: event.request.body,
				// @ts-expect-error -- Bun supports duplex streaming
				duplex: 'half',
			});

			return new Response(response.body, {
				status: response.status,
				statusText: response.statusText,
				headers: response.headers,
			});
		} catch (err) {
			proxyLogger.error('{method} {path} -> backend unreachable', {
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

// Propagate the Railway request ID to backend fetches made during SSR load()
// calls so that Go backend log entries can be correlated with the originating
// page request. IP headers are intentionally NOT injected here -- the backend
// receives those from its actual network layer (Railway's edge), not from SSR.
export const handleFetch: HandleFetch = async ({ request, fetch, event }) => {
	const requestId = event.locals.requestId;
	if (requestId && !request.headers.has('x-request-id')) {
		const headers = new Headers(request.headers);
		headers.set('x-request-id', requestId);
		request = new Request(request, { headers });
	}
	return fetch(request);
};
