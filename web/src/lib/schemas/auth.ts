import { z } from 'zod';

export const loginSchema = z.object({
	email: z.string().email('Enter a valid email address'),
	password: z.string().min(1, 'Password is required'),
});

export const registerSchema = z
	.object({
		email: z.string().email('Enter a valid email address'),
		password: z.string().min(8, 'Password must be at least 8 characters'),
		confirm_password: z.string().min(1, 'Confirm your password'),
		display_name: z.string().min(1, 'Display name is required'),
		role: z.enum(['photographer', 'customer'], { error: 'Select a role' }),
	})
	.refine((data) => data.password === data.confirm_password, {
		message: 'Passwords do not match',
		path: ['confirm_password'],
	});

export type LoginSchema = typeof loginSchema;
export type RegisterSchema = typeof registerSchema;
