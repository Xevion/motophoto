import { api, type EventListResponse, type HealthResponse } from '$lib/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch }) => {
	try {
		const [eventsData, health] = await Promise.all([
			api.get<EventListResponse>('/api/v1/events', { fetch }),
			api.get<HealthResponse>('/api/health', { fetch }),
		]);

		return {
			events: eventsData.data,
			total: eventsData.data.length,
			backendStatus: health.status,
		};
	} catch {
		return {
			events: [],
			total: 0,
			backendStatus: 'unavailable',
		};
	}
};
