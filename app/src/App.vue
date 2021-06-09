<template>
  <v-app>
    <!-- show custom nav bar on entry page -->
    <AppBar :minimize="[route.rootPage].includes($route.name)"/>
    <v-main>
      <MsgBox/>
      <transition name="component-fade" mode="out-in">
        <router-view/>
      </transition>
    </v-main>
  </v-app>
</template>

<script>
import MsgBox from './components/notification/MsgBox.vue';
import '@fontsource/roboto';
import {
  getLocalUser, getLocalSettings,
} from './utilities';
import {
  DEFAULT_SETTINGS, APPLICATION_NAME, ROUTE_NAME_ENTRY_PAGE,
} from './constants';
import AppBar from './components/AppBar.vue';

export default {
  name: 'App',
  components: {
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
    userData: {
      get() {
        return this.$store.getters.getActiveUser;
      },
      set(value) {
        this.$store.dispatch('setActiveUser', value);
      },
    },
    settings: {
      get() {
        return this.$store.getters.getSettings;
      },
      set(value) {
        this.$store.dispatch('setSettings', value);
      },
    },
  },
  methods: {
    persistDarkTheme(isDark) {
      const set = this.settings;
      set.dark = isDark;
      this.settings = set;
    },
    loadStorageData() {
      // load user data from localStorage
      const localStorageUserData = getLocalUser();
      if (localStorageUserData !== null) {
        this.userData = localStorageUserData;
      }

      // load settings from localStorage
      const localStorageSettingsData = getLocalSettings();
      if (localStorageSettingsData !== null) {
        this.settings = localStorageSettingsData;
        // dark mode according to settings
        this.$vuetify.theme.dark = this.settings.dark;
      } else {
        const defaultSettings = DEFAULT_SETTINGS;
        defaultSettings.dark = this.$vuetify.theme.dark;
        this.settings = defaultSettings;
      }

      window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
        this.persistDarkTheme(e.matches);
      });
    },
  },
  beforeMount() {
    // eslint-disable-next-line no-console
    console.info(`Branch: ${__BRANCH__}, commit: ${__COMMIT_HASH__}`);

    this.loadStorageData();
  },
};
</script>

<style>
.component-fade-enter-active, .component-fade-leave-active {
  transition: opacity 0.2s ease;
}

.component-fade-enter, .component-fade-leave-to {
  opacity: 0;
}
</style>
