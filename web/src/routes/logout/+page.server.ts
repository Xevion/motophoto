import type { Actions } from './$types';
import { redirect } from '@sveltejs/kit';
import { api } from '$lib/api';

export const actions: Actions = {
	default: async ({ fetch }) => {
		try {
			await api.post('/api/v1/auth/logout', {}, { fetch });
		} catch {
			// Ignore errors -- destroy session best-effort
		}
		redirect(303, '/login');
	},
};
