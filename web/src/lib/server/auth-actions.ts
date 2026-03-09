/**
 * Parses the JSON error body from the Go backend.
 * Returns the error string if present, or a generic fallback.
 */
export async function parseBackendError(res: Response): Promise<string> {
	try {
		const body = (await res.json()) as { error?: string };
		if (typeof body.error === 'string' && body.error.length > 0) {
			return body.error;
		}
	} catch {
		// response body was not valid JSON
	}
	return 'Something went wrong. Please try again.';
}
