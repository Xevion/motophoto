import { api, throwApiError } from '$lib/api';
import type { ItemResponse, EventResponse, GalleryResponse, ListResponse, PhotoResponse } from '$lib/types.gen';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch }) => {
	try {
		const [eventRes, galleryRes, photosRes] = await Promise.all([
			api.get<ItemResponse<EventResponse>>(`/api/v1/events/${params.id}`, { fetch }),
			api.get<ItemResponse<GalleryResponse>>(
				`/api/v1/events/${params.id}/galleries/${params.galleryId}`,
				{ fetch }
			),
			api.get<ListResponse<PhotoResponse>>(
				`/api/v1/events/${params.id}/galleries/${params.galleryId}/photos`,
				{ fetch }
			),
		]);

		return {
			event: eventRes.data,
			gallery: galleryRes.data,
			photos: photosRes.data,
		};
	} catch (err) {
		throwApiError(err);
	}
};
