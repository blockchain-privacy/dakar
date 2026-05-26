// Plugins
import {fileURLToPath} from 'node:url';
import vue from '@vitejs/plugin-vue';
import vuetify, {transformAssetUrls} from 'vite-plugin-vuetify';
import svgLoader from 'vite-svg-loader';
import {defineConfig} from 'vite';

// https://vitejs.dev/config/
export default defineConfig({
	build: {
		target: ['firefox128'],
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
			'@': fileURLToPath(new URL('src', import.meta.url)),
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
			'/dashdakar': {
				target: 'http://localhost:4455',
				changeOrigin: true,
			},
			'/btcdakar': {
				target: 'http://localhost:4455',
				changeOrigin: true,
			},
			'/wikiapi': {
				target: 'http://localhost:4455',
				changeOrigin: true,
				// Replace '/wikiapi' prefix with '/wiki'
				rewrite: path => path.replace(/^\/wikiapi/v, '/wiki'),
			},
			'/kratosadmin': {
				target: 'http://localhost:4455',
				changeOrigin: true,
			},
			'/auth': {
				target: 'http://localhost:4433',
				changeOrigin: true,
				// Remove '/auth' prefix
				rewrite: path => path.replace(/^\/auth/v, ''),
			},
			'/hydra': {
				target: 'http://localhost:4444',
				changeOrigin: true,
				// Remove '/hydra' prefix
				rewrite: path => path.replace(/^\/hydra/v, ''),
			},
			// Mcp auth discovery, see https://modelcontextprotocol.io/specification/draft/basic/authorization#authorization-server-metadata-discovery
			'/.well-known/oauth-authorization-server/dashmcp': {
				target: 'http://localhost:4444',
				changeOrigin: true,
				rewrite: path => path.replace(/(\/dashmcp\/?)$/v, ''),
			},
			'/.well-known/oauth-authorization-server/btcmcp': {
				target: 'http://localhost:4444',
				changeOrigin: true,
				rewrite: path => path.replace(/(\/btcmcp\/?)$/v, ''),
			},
			'/dashmcp': {
				target: 'http://localhost:4455',
				changeOrigin: true,
			},
			'/btcmcp': {
				target: 'http://localhost:4455',
				changeOrigin: true,
			},
		},
	},
});
