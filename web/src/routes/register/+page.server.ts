import { superValidate, message } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { registerSchema } from '$lib/schemas/auth';
import { fail, redirect } from '@sveltejs/kit';
import { parseBackendError } from '$lib/server/auth-actions';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	return {
		form: await superValidate(zod4(registerSchema)),
	};
};

export const actions: Actions = {
	default: async ({ request, fetch }) => {
		const form = await superValidate(request, zod4(registerSchema));

		if (!form.valid) {
			return fail(400, { form });
		}

		let res: Response;
		try {
			res = await fetch('/api/v1/auth/register', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					email: form.data.email,
					password: form.data.password,
					display_name: form.data.display_name,
					role: form.data.role,
				}),
			});
		} catch {
			return message(form, 'Unable to reach the server. Please try again later.', {
				status: 503,
			});
		}

		if (res.ok) {
			redirect(303, '/');
		}

		if (res.status === 409) {
			return message(form, 'An account with this email already exists.', { status: 409 });
		}
		if (res.status === 429) {
			return message(form, 'Too many attempts. Please wait a moment and try again.', {
				status: 429,
			});
		}

		const errorMsg = await parseBackendError(res);
		return message(form, errorMsg, { status: res.status as 400 | 500 });
	},
};
