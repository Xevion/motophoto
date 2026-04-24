import { api } from '$lib/api';
import type { ItemResponse, ListResponse, EventResponse } from '$lib/types.gen';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch }) => {
	try {
		// Get user's events (all events since we don't have a filter-by-owner endpoint yet)
		// The backend should ideally have a /api/v1/me/events endpoint
		const res = await api.get<ListResponse<EventResponse>>('/api/v1/events', { fetch });
		return { events: res.data || [] };
	} catch (err) {
		console.error('Failed to fetch events:', err);
		return { events: [] };
	}
};
