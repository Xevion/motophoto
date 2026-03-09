import { superValidate, message } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { loginSchema } from '$lib/schemas/auth';
import { fail, redirect } from '@sveltejs/kit';
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

		const res = await fetch('/api/v1/auth/login', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				email: form.data.email,
				password: form.data.password,
			}),
		});

		if (!res.ok) {
			if (res.status === 403) {
				return message(form, 'This account has been banned.', { status: 403 });
			}
			if (res.status === 401) {
				return message(form, 'Invalid email or password.', { status: 401 });
			}
			return message(form, 'Something went wrong. Please try again.', { status: 500 });
		}

		redirect(303, '/');
	},
};
