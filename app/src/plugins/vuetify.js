import Vue from 'vue';
import Vuetify from 'vuetify/lib';

Vue.use(Vuetify);

const mq = window.matchMedia('(prefers-color-scheme: dark)')

const vuetify = new Vuetify({
    theme: {
        themes: {
            light: {
                primary: '#3f51b5',
                secondary: '#e8eaf6',
                accent: '#304ffe',
            },
            dark: {
                primary: '#3f51b5',
                secondary: '#e8eaf6',
                accent: '#304ffe',
            },
        },
        dark: mq.matches,
    },
});
mq.addEventListener('change', (e) => {
    vuetify.framework.theme.dark = e.matches
})

export default vuetify;

