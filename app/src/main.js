import Vue from 'vue';
import App from './App.vue';
import vuetify from './plugins/vuetify';
import router from './router';
import store from './state';
import ory from './plugins/ory';

Vue.config.productionTip = false;
Vue.use(ory);

new Vue({
  vuetify,
  store,
  router,
  render: (h) => h(App),
}).$mount('#app');
