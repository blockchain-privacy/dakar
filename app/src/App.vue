<template>
  <v-app>
    <app-bar :minimize="isEntryPage" />
    <v-main>
      <div style="position: relative">
        <msg-box />
      </div>
      <router-view v-slot="{ Component }">
        <fade-transition>
          <component :is="Component" />
        </fade-transition>
      </router-view>
    </v-main>
  </v-app>
</template>

<script>
import MsgBox from './components/notification/MsgBox.vue';
import '@fontsource/roboto';
import {
	DEFAULT_SETTINGS, APPLICATION_NAME, ROUTE_NAME_ENTRY_PAGE,
} from './constants';
import AppBar from './components/appbar/AppBar.vue';
import {isSessionExpired} from './utilities';
import FadeTransition from '@/components/common/FadeTransition.vue';

export default {
	name: 'App',
	components: {
		FadeTransition,
		AppBar,
		MsgBox,
	},
	data() {
		return {
			applicationName: APPLICATION_NAME,
			route: {
				rootPage: ROUTE_NAME_ENTRY_PAGE,

			},
		};
	},
	computed: {
		settings: {
			get() {
				return this.$store.getters.getSettings;
			},
			set(value) {
				this.$store.dispatch('setSettings', value);
			},
		},
		session: {
			get() {
				return this.$store.getters.getSession;
			},
			set(value) {
				this.$store.dispatch('setSession', value);
			},
		},
		isEntryPage() {
			return this.$route.name === this.route.rootPage;
		},
	},
	mounted() {
		if (isSessionExpired(this.session)) {
			this.session = null;
		}
	},
	beforeMount() {
		this.setDarkTheme();
		window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', e => {
			this.persistDarkTheme(e.matches);
			this.setDarkTheme();
		});
	},
	methods: {
		persistDarkTheme(isDark) {
			const set = this.settings;
			set.dark = isDark;
			this.settings = set;
		},
		setDarkTheme() {
			// Create new settings object if necessary
			if (this.settings === null) {
				const defaultSettings = DEFAULT_SETTINGS;
				defaultSettings.dark = window.matchMedia('(prefers-color-scheme: dark)').matches;
				this.settings = defaultSettings;
			}

			this.$vuetify.theme.global.name = this.settings.dark ? 'dark' : 'light';
		},
	},
};
</script>

<style scoped>

</style>
