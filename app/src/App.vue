<template>
  <v-app>
    <!-- show custom nav bar on entry page -->
    <AppBar :minimize="isEntryPage"/>
    <v-main>
      <MsgBox/>
      <transition name="component-fade" mode="out-in">
        <router-view/>
      </transition>
    </v-main>
    <!-- show footer only on entry page -->
    <v-container v-if="isEntryPage"
                 class="footer">
      <v-row justify="center">
        <v-col md="2" class="text-center mx-1">
          <v-btn :to="{name: route.privacyPage}" plain small>
            Privacy Policy
          </v-btn>
        </v-col>
        <v-col md="2" class="text-center mx-1">
          <v-btn :to="{name: route.termsOfUsePage}" plain small>
            Terms of Use
          </v-btn>
        </v-col>
        <v-col md="2" class="text-center mx-1">
          <v-btn :to="{name: route.aboutPage}" plain small>
            About
          </v-btn>
        </v-col>
      </v-row>
    </v-container>
  </v-app>
</template>

<script>
import MsgBox from './components/notification/MsgBox.vue';
import '@fontsource/roboto';
import {
  DEFAULT_SETTINGS, APPLICATION_NAME, ROUTE_NAME_ENTRY_PAGE, ROUTE_NAME_ABOUT,
  ROUTE_NAME_TERMS_OF_USE, ROUTE_NAME_PRIVACY,
} from './constants';
import AppBar from './components/AppBar.vue';
import { isSessionExpired } from './utilities';

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
        aboutPage: ROUTE_NAME_ABOUT,
        termsOfUsePage: ROUTE_NAME_TERMS_OF_USE,
        privacyPage: ROUTE_NAME_PRIVACY,
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
  methods: {
    persistDarkTheme(isDark) {
      const set = this.settings;
      set.dark = isDark;
      this.settings = set;
    },
    setDarkTheme() {
      if (this.settings !== null) {
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
  mounted() {
    if (isSessionExpired(this.session)) {
      this.session = null;
    }
  },
  beforeMount() {
    // eslint-disable-next-line no-console
    console.info(`Branch: ${__BRANCH__}, commit: ${__COMMIT_HASH__}`);
    this.setDarkTheme();
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

.footer {
  left:0;
  right:0;
  bottom:0;
}
</style>
