import { api, type ItemResponse } from '$lib/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, parent, fetch }) => {
	const { event } = await parent();

	if (!event) {
		return {
			event: null,
			gallery: null
		};
	}

	const { data: gallery } = await api.get<ItemResponse<any>>(
		`/api/v1/events/${event.id}/galleries/${params.galleryId}`,
		{ fetch }
	);

	return {
		event,
		gallery
	};
};