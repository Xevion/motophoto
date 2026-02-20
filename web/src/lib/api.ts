export type { Event } from './types.gen';
import type { Event } from './types.gen';

export interface EventsResponse {
	events: Event[];
	total: number;
}

export interface HealthResponse {
	status: string;
}

/**
 * Fetch wrapper using relative URLs. In dev, Vite proxies /api to the Go backend.
 * In production, hooks.server.ts proxies /api to the Go backend.
 * Both SSR and client-side requests go through the same path.
 */
export async function apiFetch<T>(path: string, fetchFn: typeof fetch = fetch): Promise<T> {
	const res = await fetchFn(path);
	if (!res.ok) {
		throw new Error(`API error: ${res.status} ${res.statusText}`);
	}
	return res.json() as Promise<T>;
}
