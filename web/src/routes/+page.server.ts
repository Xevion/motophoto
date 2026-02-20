import { apiFetch, type EventsResponse, type HealthResponse } from '$lib/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch }) => {
	const [eventsData, health] = await Promise.all([
		apiFetch<EventsResponse>('/api/v1/events', fetch),
		apiFetch<HealthResponse>('/api/health', fetch)
	]);

	return {
		events: eventsData.events,
		total: eventsData.total,
		backendStatus: health.status
	};
};
