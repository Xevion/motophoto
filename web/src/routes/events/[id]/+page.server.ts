import { api, throwApiError } from '$lib/api';
import type { ItemResponse, EventResponse } from '$lib/types.gen';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch }) => {
	try {
		const { data: event } = await api.get<ItemResponse<EventResponse>>(`/api/v1/events/${params.id}`, { fetch });
		return { event };
	} catch (err) {
		throwApiError(err);
	}
};
