import adapter from '@xevion/svelte-adapter-bun';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		adapter: adapter(),
	},
};

export default config;
