import App from './App.vue';

import {createApp} from 'vue';
import {createPinia} from 'pinia';

import vuetify from './plugins/vuetify';
import router from '@/router';
import oryConfig from './plugins/ory';
import dakarConfig from './plugins/dakarAPI';
import wikiapiConfig from './plugins/wikiAPI';

const pinia = createPinia();
const app = createApp(App);

app.use(pinia);
app.use(vuetify).use(router);

// Provide global variables here, so they can be later injected
app.provide('ory', oryConfig);
app.provide('dakar', dakarConfig.setup(app.config.globalProperties));
app.provide('wikiapi', wikiapiConfig.setup(app.config.globalProperties).default);

await router.isReady();
app.mount('#app');
