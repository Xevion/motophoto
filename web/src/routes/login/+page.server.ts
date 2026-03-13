import { superValidate, message } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { loginSchema } from '$lib/schemas/auth';
import { fail, redirect } from '@sveltejs/kit';
import { api, ApiError, NetworkError } from '$lib/api';
import type { ItemResponse, UserResponse } from '$lib/types.gen';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	return {
		form: await superValidate(zod4(loginSchema)),
	};
};

export const actions: Actions = {
	default: async ({ request, fetch }) => {
		const form = await superValidate(request, zod4(loginSchema));

		if (!form.valid) {
			return fail(400, { form });
		}

		try {
			await api.post<ItemResponse<UserResponse>>(
				'/api/v1/auth/login',
				{ email: form.data.email, password: form.data.password },
				{ fetch },
			);
		} catch (err) {
			if (err instanceof NetworkError) {
				return message(form, 'Unable to reach the server. Please try again later.', {
					status: 503,
				});
			}
			if (err instanceof ApiError) {
				const msg = err.body?.error ?? 'Login failed. Please try again.';
				return message(form, msg, { status: err.status as 400 | 401 | 403 | 429 | 500 });
			}
			throw err;
		}

		redirect(303, '/');
	},
};
