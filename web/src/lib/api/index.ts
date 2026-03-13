import type { EventResponse, ListResponse } from '$lib/types.gen';

export { apiFetch, api } from './client';
export type { ApiFetchOptions } from './client';
export { ApiError, NetworkError } from './errors';
export { throwApiError } from './handle';

export type EventListResponse = ListResponse<EventResponse>;

export interface HealthResponse {
	status: string;
}
