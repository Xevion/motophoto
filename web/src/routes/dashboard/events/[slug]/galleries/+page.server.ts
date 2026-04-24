import { api } from '$lib/api';
import type { ItemResponse, EventResponse } from '$lib/types.gen';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch }) => {
	try {
		const res = await api.get<ItemResponse<EventResponse>>(`/api/v1/events/${params.slug}`, { fetch });
		return { event: res.data };
	} catch (err) {
		console.error('Failed to fetch event:', err);
		throw err;
	}
};
