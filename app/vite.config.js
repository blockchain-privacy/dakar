// Plugins
import vue from '@vitejs/plugin-vue';
import vuetify, {transformAssetUrls} from 'vite-plugin-vuetify';
import svgLoader from 'vite-svg-loader';

// Utilities
import {defineConfig} from 'vite';
import {fileURLToPath, URL} from 'node:url';

// https://vitejs.dev/config/
export default defineConfig({
	build: {
		target: ['firefox115'],
	},
	plugins: [
		vue({
			template: {transformAssetUrls},
		}),
		// https://github.com/vuetifyjs/vuetify-loader/tree/next/packages/vite-plugin
		vuetify({
			autoImport: true,
			styles: {
				configFile: 'src/styles/settings.scss',
			},
		}),
		svgLoader(),
	],
	define: {'process.env': {}},
	resolve: {
		alias: {
			'@': fileURLToPath(new URL('./src', import.meta.url)),
		},
		extensions: [
			'.js',
			'.json',
			'.jsx',
			'.mjs',
			'.ts',
			'.tsx',
			'.vue',
		],
	},
	server: {
		port: 3000,
		proxy: {
			'/api': {
				target: 'http://localhost:8081',
			},
			'/auth': {
				target: 'http://localhost:4433',
				changeOrigin: true,
				// Remove '/auth' prefix
				rewrite: path => path.replace(/^\/auth/, ''),
			},
			'/wikiapi': {
				target: 'http://localhost:4455',
				// Remove '/wiki' prefix
				changeOrigin: true,
				rewrite: path => path.replace(/^\/wikiapi/, ''),
			},
		},
	},
});
