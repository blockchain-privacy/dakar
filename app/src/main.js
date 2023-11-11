import App from './App.vue';

import {createApp} from 'vue';

import vuetify from './plugins/vuetify';
import router from '@/router';
import store from '@/state';
import oryConfig from './plugins/ory';
import dakarConfig from './plugins/dakarAPI';
import wikiapiConfig from './plugins/wikiAPI';

// Todo: remove this function wrapper, when browser support "top-level await"
(
	async () => {
		const app = createApp(App);

		app.use(vuetify).use(router).use(store);
		app.config.globalProperties.ory = oryConfig;
		app.config.globalProperties.dakar = dakarConfig.setup(app.config.globalProperties);
		app.config.globalProperties.wikiapi = wikiapiConfig.setup(app.config.globalProperties).default;

		await router.isReady();
		app.mount('#app');
	}
)();

