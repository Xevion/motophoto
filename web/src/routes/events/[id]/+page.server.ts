import { apiFetch } from '$lib/api';
import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { EventResponse } from '$lib/types.gen';

export const load: PageServerLoad = async ({ params, fetch }) => {
	try {
		const event = await apiFetch<EventResponse>(`/api/v1/events/${params.id}`, fetch);
		return { event };
	} catch {
		error(404, 'Event not found');
	}
};
