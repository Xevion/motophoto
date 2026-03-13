// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
import type { UserResponse } from '$lib/types.gen';

declare global {
	namespace App {
		// interface Error {}
		interface Locals {
			requestId: string;
		}
		interface PageData {
			user: UserResponse | null;
		}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
