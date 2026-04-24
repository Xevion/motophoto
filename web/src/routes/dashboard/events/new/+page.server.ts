import { fail } from '@sveltejs/kit';
import { superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { eventSchema } from '$lib/schemas/event';
import { api, throwApiError } from '$lib/api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	const form = await superValidate(zod4(eventSchema));
	return { form };
};

export const actions: Actions = {
	default: async ({ request, fetch }) => {
		const form = await superValidate(request, zod4(eventSchema));

		if (!form.valid) {
			return fail(400, { form });
		}

		try {
			await api.post('/api/v1/events', form.data, { fetch });
			return { form, success: true };
		} catch (err) {
			console.error('Failed to create event:', err);
			return fail(500, {
				form,
				error: err instanceof Error ? err.message : 'Failed to create event',
			});
		}
	},
};
