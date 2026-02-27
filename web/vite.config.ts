import devtoolsJson from 'vite-plugin-devtools-json';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import { jsonLogger } from './vite-plugin-json-logger';

export default defineConfig({
	clearScreen: false,
	plugins: [jsonLogger(), sveltekit(), devtoolsJson()],
	server: {
		fs: { allow: ['styled-system'] },
		proxy: {
			'/api': { target: 'http://localhost:3001', changeOrigin: true },
		},
	},
});
