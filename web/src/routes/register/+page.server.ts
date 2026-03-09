import { superValidate, message } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { registerSchema } from '$lib/schemas/auth';
import { fail, redirect } from '@sveltejs/kit';
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

		const res = await fetch('/api/v1/auth/register', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				email: form.data.email,
				password: form.data.password,
				display_name: form.data.display_name,
				role: form.data.role,
			}),
		});

		if (!res.ok) {
			if (res.status === 409) {
				return message(form, 'An account with this email already exists.', { status: 409 });
			}
			return message(form, 'Something went wrong. Please try again.', { status: 500 });
		}

		redirect(303, '/');
	},
};
