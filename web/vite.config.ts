import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';
import { jsonLogger } from './vite-plugin-json-logger';

export default defineConfig({
	plugins: [jsonLogger(), tailwindcss(), sveltekit()],
	server: {
		proxy: {
			'/api': {
				target: 'http://localhost:3001',
				changeOrigin: true,
			},
		},
	},
});
