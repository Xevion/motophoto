const API_BASE = 'http://localhost:8080';

export interface Event {
	id: number;
	name: string;
	sport: string;
	location: string;
	date: string;
	photo_count: number;
	galleries: number;
	description: string;
	tags: string[];
}

export interface EventsResponse {
	events: Event[];
	total: number;
}

export interface HealthResponse {
	status: string;
}

/** Fetch wrapper that hits the Go backend directly (for SSR server-side loads). */
export async function apiFetch<T>(path: string, fetchFn: typeof fetch = fetch): Promise<T> {
	const res = await fetchFn(`${API_BASE}${path}`);
	if (!res.ok) {
		throw new Error(`API error: ${res.status} ${res.statusText}`);
	}
	return res.json() as Promise<T>;
}
