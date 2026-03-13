/**
 * Maps terse backend error strings to user-facing messages. The backend stays
 * machine-readable; the frontend owns UX copy. Keys are exact matches against
 * the `error` field in the backend's JSON error responses.
 */
const ERROR_MESSAGES: Record<string, string> = {
	// auth
	'invalid credentials': 'Incorrect email or password.',
	'account is banned': 'This account has been suspended.',
	'email already exists': 'An account with this email already exists.',
	'login failed': 'Something went wrong during login. Please try again.',
	'logout failed': 'Something went wrong during logout. Please try again.',
	'failed to register': 'Something went wrong during registration. Please try again.',
	'invalid role': 'Invalid account type selected.',

	// generic
	'not found': 'The requested resource was not found.',
	'already exists': 'This resource already exists.',
	'invalid request body': 'The request could not be processed. Please check your input.',

	// proxy
	'Backend unavailable': 'The server is temporarily unavailable. Please try again shortly.',
	'Proxy error': 'Something went wrong reaching the server. Please try again.',
};

/** Default message when no mapping exists for the backend error. */
const FALLBACK = 'Something went wrong. Please try again.';

/**
 * Translates a backend error string into a user-facing message. Falls back
 * to a generic message if the error is unrecognized. Validation errors
 * (prefixed with "validation failed:") are passed through as-is since they
 * already contain field-specific details.
 */
export function userMessage(backendError: string | undefined): string {
	if (!backendError) return FALLBACK;
	if (backendError.startsWith('validation failed:')) return backendError;
	return ERROR_MESSAGES[backendError] ?? FALLBACK;
}
