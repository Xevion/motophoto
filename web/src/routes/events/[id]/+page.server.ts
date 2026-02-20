import { apiFetch, type Event } from '$lib/api';
import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch }) => {
	try {
		const event = await apiFetch<Event>(`/api/v1/events/${params.id}`, fetch);
		return { event };
	} catch {
		error(404, 'Event not found');
	}
};
