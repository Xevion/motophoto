import { error as svelteError } from '@sveltejs/kit';
import { ApiError, NetworkError } from './errors';
import { userMessage } from './messages';

/**
 * Maps an ApiError or NetworkError to a SvelteKit error response. Use in load
 * function catch blocks to propagate the correct HTTP status to the client
 * instead of collapsing every failure into a generic 404 or 500.
 */
export function throwApiError(err: unknown): never {
	if (err instanceof ApiError) {
		svelteError(err.status, userMessage(err.body?.error));
	}
	if (err instanceof NetworkError) {
		svelteError(502, userMessage('Backend unavailable'));
	}
	svelteError(500, userMessage(undefined));
}
