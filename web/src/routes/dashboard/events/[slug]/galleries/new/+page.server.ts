import { fail } from '@sveltejs/kit';
import { superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { gallerySchema } from '$lib/schemas/event';
import { api } from '$lib/api';
import type { PageServerLoad, Actions } from './$types';

export const load: PageServerLoad = async ({ params, fetch }) => {
	const form = await superValidate(zod4(gallerySchema));
	return { form };
};

export const actions: Actions = {
	default: async ({ params, request, fetch }) => {
		const form = await superValidate(request, zod4(gallerySchema));

		if (!form.valid) {
			return fail(400, { form });
		}

		try {
			await api.post(`/api/v1/events/${params.slug}/galleries`, form.data, { fetch });
			// Redirect to galleries list
			return { form, success: true };
		} catch (err) {
			console.error('Failed to create gallery:', err);
			return fail(500, {
				form,
				error: err instanceof Error ? err.message : 'Failed to create gallery',
			});
		}
	},
};
