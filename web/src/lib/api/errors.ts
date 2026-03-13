import type { ErrorResponse } from '$lib/types.gen';

/**
 * Structured error thrown when the backend returns a non-2xx HTTP response.
 * Preserves the status code and parsed error body so callers can branch
 * on specific conditions without string matching.
 */
export class ApiError extends Error {
	readonly name = 'ApiError' as const;

	constructor(
		readonly status: number,
		readonly statusText: string,
		readonly body: ErrorResponse | null,
	) {
		super(body?.error ?? `${status} ${statusText}`);
	}

	get isNotFound(): boolean {
		return this.status === 404;
	}
	get isUnauthorized(): boolean {
		return this.status === 401;
	}
	get isForbidden(): boolean {
		return this.status === 403;
	}
	get isConflict(): boolean {
		return this.status === 409;
	}
	get isRateLimited(): boolean {
		return this.status === 429;
	}
	get isServerError(): boolean {
		return this.status >= 500;
	}
}

/**
 * Thrown when the fetch call itself fails (network unreachable, DNS failure,
 * backend down). Distinct from ApiError so callers can handle connectivity
 * issues separately from HTTP error responses.
 */
export class NetworkError extends Error {
	readonly name = 'NetworkError' as const;

	constructor(message: string) {
		super(message);
	}
}
