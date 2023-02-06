import { sveltekit } from '@sveltejs/kit/vite';
import type { UserConfig } from 'vite';

const config: UserConfig = {
	plugins: [sveltekit()],
	server: {
		proxy: {
			'/api': {
				target: 'http://localhost:8010',
				changeOrigin: true,
				rewrite: (path) => path.replace(/^\/api/, '')
			},
			'/ws': {
				target: 'http://localhost:8010',
				ws: true,
			}
		}
	}
};

export default config;
