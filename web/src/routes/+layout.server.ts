import type { ItemResponse, UserResponse } from '$lib/types.gen';
import { api, ApiError } from '$lib/api';
import { getLogger } from '@logtape/logtape';
import type { LayoutServerLoad } from './$types';

const logger = getLogger(['ssr', 'auth']);

export const load: LayoutServerLoad = async ({ fetch }) => {
	try {
		const res = await api.get<ItemResponse<UserResponse>>('/api/v1/me', { fetch });
		return { user: res.data };
	} catch (err) {
		if (!(err instanceof ApiError && err.status === 401)) {
			logger.error('failed to fetch current user: {error}', {
				error: err instanceof Error ? err.message : String(err),
			});
		}
		return { user: null };
	}
};
