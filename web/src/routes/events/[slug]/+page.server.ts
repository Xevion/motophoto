import { api, throwApiError } from '$lib/api';
import type { ItemResponse, EventResponse } from '$lib/types.gen';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ parent }) => {
	const { event } = await parent()

	return {
		event
	}
};
