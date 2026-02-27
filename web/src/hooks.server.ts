import { env } from '$env/dynamic/private';
import { initLogger } from '$lib/logger.server';
import { getLogger } from '@logtape/logtape';
import type { Handle, HandleFetch } from '@sveltejs/kit';

await initLogger();

const BACKEND_URL = env.BACKEND_URL ?? 'http://localhost:3001';

const proxyLogger = getLogger(['ssr', 'proxy']);

// Headers from the incoming request that are safe to forward to the Go backend.
// The proxy chain is: Client -> Cloudflare -> Fastly (Railway edge) -> SvelteKit -> Go.
// Cloudflare sets True-Client-IP and CF-Connecting-IP to the real client IP.
// Railway's X-Real-IP contains its immediate upstream (Cloudflare), not the client.
// Chi's RealIP middleware checks True-Client-IP first, so forwarding it gives
// the backend the actual client IP. Forwarding only an explicit list prevents
// arbitrary client headers from reaching the backend.
const FORWARDED_REQUEST_HEADERS = [
	'x-railway-request-id',
	'x-request-id',
	'true-client-ip',
	'cf-connecting-ip',
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

// Propagate the request ID and client IP headers to backend fetches during SSR.
// The Go backend's network peer is SvelteKit, not Railway's edge, so IP headers
// must be forwarded explicitly or the backend logs [::1] for every page load.
export const handleFetch: HandleFetch = async ({ request, fetch, event }) => {
	const headers = new Headers(request.headers);

	const requestId = event.locals.requestId;
	if (requestId && !headers.has('x-request-id')) {
		headers.set('x-request-id', requestId);
	}

	for (const name of ['true-client-ip', 'cf-connecting-ip', 'x-real-ip', 'x-forwarded-for'] as const) {
		const value = event.request.headers.get(name);
		if (value !== null && !headers.has(name)) {
			headers.set(name, value);
		}
	}

	return fetch(new Request(request, { headers }));
};
