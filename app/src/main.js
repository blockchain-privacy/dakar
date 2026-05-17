// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

import {createApp} from 'vue';
import {createPinia} from 'pinia';
import App from './App.vue';
import vuetify from './plugins/vuetify';
import './assets/main.css';
import oryConfig from './plugins/ory';
import dakarConfig from './plugins/dakarAPI';
import wikiapiConfig from './plugins/wikiAPI';
import kratosadminConfig from './plugins/kratosadmin';
import {router, setupStore} from '@/router';

const pinia = createPinia();
const app = createApp(App);

app.use(pinia);

// Must not be called before app.use(pinia) and not after app.use(vuetify).use(router)
setupStore();

app.use(vuetify).use(router);

// Provide global variables here, so they can be later injected
app.provide('ory', oryConfig);
app.provide('dashdakar', dakarConfig.setup(app.config.globalProperties, '/dashdakar'));
app.provide('btcdakar', dakarConfig.setup(app.config.globalProperties, '/btcdakar'));
app.provide('wikiapi', wikiapiConfig.setup(app.config.globalProperties).default);
app.provide('kratosadmin', kratosadminConfig.setup(app.config.globalProperties));

await router.isReady();
app.mount('#app');
