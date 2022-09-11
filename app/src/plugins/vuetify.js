import Vue from 'vue';
import Vuetify from 'vuetify/lib';

Vue.use(Vuetify);

const mq = window.matchMedia('(prefers-color-scheme: dark)');

const vuetify = new Vuetify({
  theme: {
    dark: mq.matches,
    themes: {
      light: {
        primary: '#1976d2',
      },
      dark: {
        primary: '#1976d2',
      },
    },
    options: {
      customProperties: true,
    },
  },
  icons: {
    iconfont: 'mdiSvg',
  },
});
mq.addEventListener('change', (e) => {
  vuetify.framework.theme.dark = e.matches;
});

export default vuetify;
