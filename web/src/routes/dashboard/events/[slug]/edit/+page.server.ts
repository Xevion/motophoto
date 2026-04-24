import { fail, redirect } from '@sveltejs/kit';
import { superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { eventSchema } from '$lib/schemas/event';
import { api, throwApiError } from '$lib/api';
import type { PageServerLoad, Actions } from './$types';
import type { ItemResponse, EventResponse } from '$lib/types.gen';

export const load: PageServerLoad = async ({ params, fetch, parent }) => {
	try {
		const res = await api.get<ItemResponse<EventResponse>>(`/api/v1/events/${params.slug}`, { fetch });
		const event = res.data;

		const form = await superValidate(
			{
				name: event.name,
				slug: event.slug,
				sport: event.sport,
				status: event.status as 'draft' | 'published' | 'archived',
				tags: event.tags,
				description: event.description,
				location: event.location,
				date: event.date,
			},
			zod4(eventSchema)
		);
		return { form, event };
	} catch (err) {
		throwApiError(err);
	}
};

export const actions: Actions = {
	default: async ({ params, request, fetch }) => {
		const form = await superValidate(request, zod4(eventSchema));

		if (!form.valid) {
			return fail(400, { form });
		}

		try {
			await api.patch(`/api/v1/events/${params.slug}`, form.data, { fetch });
			redirect(302, '/dashboard');
		} catch (err) {
			console.error('Failed to update event:', err);
			return fail(500, {
				form,
				error: err instanceof Error ? err.message : 'Failed to update event',
			});
		}
	},
};
