import { api, throwApiError } from '$lib/api';
import type { ItemResponse, EventResponse } from '$lib/types.gen';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ params, fetch }) => {
	try {
		const { data: event } = await api.get<ItemResponse<EventResponse>>(
			`/api/v1/events/${params.slug}`,
			{ fetch }
		);

		return {
			event
		};
	} catch (err) {
		return {
			event: null
		};
	}
};