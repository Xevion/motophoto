import adapter from '@xevion/svelte-adapter-bun';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		adapter: adapter(),
		alias: {
			'styled-system': './styled-system/*',
		},
	},
};

export default config;
