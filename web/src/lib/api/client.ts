import type { ErrorResponse } from '$lib/types.gen';
import { ApiError, NetworkError } from './errors';

export interface ApiFetchOptions {
	method?: string;
	body?: unknown;
	params?: Record<string, string>;
	/** Pass SvelteKit's event `fetch` so cookies and headers propagate during SSR. */
	fetch?: typeof fetch;
}

/**
 * Typed fetch wrapper for the Go backend. Uses relative URLs routed through
 * the Vite proxy (dev) or SvelteKit's handle hook (production).
 *
 * Throws {@link ApiError} on non-2xx responses and {@link NetworkError} when
 * the request itself fails (backend unreachable, DNS, etc.).
 */
export async function apiFetch<T>(path: string, opts: ApiFetchOptions = {}): Promise<T> {
	const { method = 'GET', body, params, fetch: fetchFn = fetch } = opts;

	const url = params ? `${path}?${new URLSearchParams(params)}` : path;
	const headers: Record<string, string> = {};
	let rawBody: BodyInit | undefined;

	if (body !== undefined) {
		headers['Content-Type'] = 'application/json';
		rawBody = JSON.stringify(body);
	}

	let res: Response;
	try {
		res = await fetchFn(url, { method, headers, body: rawBody });
	} catch (err) {
		throw new NetworkError(err instanceof Error ? err.message : 'Network request failed');
	}

	if (!res.ok) {
		let errorBody: ErrorResponse | null = null;
		try {
			errorBody = (await res.json()) as ErrorResponse;
		} catch {
			// response body was not valid JSON
		}
		throw new ApiError(res.status, res.statusText, errorBody);
	}

	return res.json() as Promise<T>;
}

/** Convenience methods that fix the HTTP method for common operations. */
export const api = {
	get: <T>(path: string, opts?: Omit<ApiFetchOptions, 'method' | 'body'>) =>
		apiFetch<T>(path, { ...opts, method: 'GET' }),

	post: <T>(path: string, body: unknown, opts?: Omit<ApiFetchOptions, 'method' | 'body'>) =>
		apiFetch<T>(path, { ...opts, method: 'POST', body }),

	put: <T>(path: string, body: unknown, opts?: Omit<ApiFetchOptions, 'method' | 'body'>) =>
		apiFetch<T>(path, { ...opts, method: 'PUT', body }),

	patch: <T>(path: string, body: unknown, opts?: Omit<ApiFetchOptions, 'method' | 'body'>) =>
		apiFetch<T>(path, { ...opts, method: 'PATCH', body }),

	del: <T>(path: string, opts?: Omit<ApiFetchOptions, 'method' | 'body'>) =>
		apiFetch<T>(path, { ...opts, method: 'DELETE' }),
};
