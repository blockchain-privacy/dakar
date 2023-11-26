import App from './App.vue';

import {createApp} from 'vue';

import vuetify from './plugins/vuetify';
import router from '@/router';
import store from '@/state';
import oryConfig from './plugins/ory';
import dakarConfig from './plugins/dakarAPI';
import wikiapiConfig from './plugins/wikiAPI';

const app = createApp(App);

app.use(vuetify).use(router).use(store);

// Todo after switching fully to composition api, check if this can be removed
app.config.globalProperties.ory = oryConfig;
app.config.globalProperties.dakar = dakarConfig.setup(app.config.globalProperties);
app.config.globalProperties.wikiapi = wikiapiConfig.setup(app.config.globalProperties).default;

// Provide global variables here, so they can be later injected
app.provide('ory', oryConfig);
app.provide('dakar', dakarConfig.setup(app.config.globalProperties));
app.provide('wikiapi', wikiapiConfig.setup(app.config.globalProperties).default);

await router.isReady();
app.mount('#app');
