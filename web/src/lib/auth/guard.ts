import { redirect, error } from '@sveltejs/kit';
import { base } from '$app/paths';
import type { UserResponse } from '$lib/types.gen';

// Frontend uses a role hierarchy (higher rank includes lower).
// The backend's RequireRole does exact match instead. This is intentional:
// the backend only gates photographer-write routes, while the frontend
// needs "at least this role" checks for UI guards.
const ROLE_RANK: Record<string, number> = {
	anonymous: 0,
	customer: 1,
	photographer: 2,
};

/**
 * Require the user to have at least the given role.
 * Defaults to 'customer' (any authenticated user).
 * Throws redirect to /login if not authenticated, or 403 if insufficient role.
 */
export function requireRole(
	user: UserResponse | null,
	minimum: keyof typeof ROLE_RANK = 'customer',
): asserts user is UserResponse {
	if (minimum === 'anonymous') return;
	// eslint-disable-next-line @typescript-eslint/only-throw-error -- SvelteKit redirect/error are designed to be thrown
	if (!user) throw redirect(303, `${base}/login`);
	const userRank = ROLE_RANK[user.role] ?? 0;
	const requiredRank = ROLE_RANK[minimum];
	// eslint-disable-next-line @typescript-eslint/only-throw-error -- SvelteKit error() is designed to be thrown
	if (userRank < requiredRank) throw error(403, 'Insufficient permissions');
}

/**
 * Redirect authenticated users away (e.g., from login/register pages).
 */
export function redirectIfAuthenticated(user: UserResponse | null, to = `${base}/`): void {
	// eslint-disable-next-line @typescript-eslint/only-throw-error -- SvelteKit redirect is designed to be thrown
	if (user) throw redirect(303, to);
}
