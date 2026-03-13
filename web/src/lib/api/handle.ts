import { error as svelteError } from '@sveltejs/kit';
import { ApiError, NetworkError } from './errors';

/**
 * Maps an ApiError or NetworkError to a SvelteKit error response. Use in load
 * function catch blocks to propagate the correct HTTP status to the client
 * instead of collapsing every failure into a generic 404 or 500.
 */
export function throwApiError(err: unknown, fallbackMessage = 'Something went wrong'): never {
	if (err instanceof ApiError) {
		svelteError(err.status, err.body?.error ?? err.statusText);
	}
	if (err instanceof NetworkError) {
		svelteError(502, 'Backend unavailable');
	}
	svelteError(500, fallbackMessage);
}
