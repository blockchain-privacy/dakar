<template>
  <v-app>
    <v-app-bar app absolute>
      <router-link :to="{name: route.rootPage}">
        <v-img
            style="cursor:pointer"
            alt="Dakar Logo"
            class="shrink mr-2"
            contain
            src="./assets/dakar_dash.svg"
            transition="scale-transition"
            width="32">
        </v-img>
      </router-link>
      <router-link :to="{name: route.rootPage}" style="color: inherit; text-decoration: inherit">
        <v-toolbar-title class="mx-2 d-none d-sm-flex" style="cursor:pointer">
          {{ applicationName }}
        </v-toolbar-title>
      </router-link>
      <v-spacer></v-spacer>
      <QueryInput class="mx-4"/>
      <v-spacer></v-spacer>
      <v-btn icon :to="{name: route.heuristicsPage}" v-if="showTools">
        <v-icon>{{ icon.mdiToolbox }}</v-icon>
      </v-btn>
      <v-menu offset-y style="z-index: 99">
        <template v-slot:activator="{ on, attrs }">
          <v-btn
              icon
              v-bind="attrs"
              v-on="on">
            <v-icon>{{ icon.mdiAccount }}</v-icon>
          </v-btn>
        </template>
        <v-list nav dense>
          <v-list-item v-if="this.userData">
            <v-list-item-icon>
              <v-icon>{{ icon.mdiAccountCircle }}</v-icon>
            </v-list-item-icon>
            <v-list-item-title> {{ this.userData.email }}</v-list-item-title>
          </v-list-item>
          <v-divider v-if="this.userData"/>
          <v-list-item :to="{name: route.userProfilePage}" v-if="this.userData"
                       :disabled="isUserProfileDisabled">
            <v-list-item-icon>
              <v-icon>{{ icon.mdiCog }}</v-icon>
            </v-list-item-icon>
            <v-list-item-title>Settings</v-list-item-title>
          </v-list-item>
          <v-list-item :to="{ name: route.userLoginPage }"
                       v-if="!this.userData" :disabled="isUserLoginDisabled">
            <v-list-item-icon>
              <v-icon>{{ icon.mdiLogin }}</v-icon>
            </v-list-item-icon>
            <v-list-item-title>Login</v-list-item-title>
          </v-list-item>
          <v-list-item :to="{ name: route.userAdminPage }" :disabled="isUserAdminDisabled"
                       v-if="showUserAdmin">
            <v-list-item-icon>
              <v-icon>{{ icon.mdiAccountSupervisor }}</v-icon>
            </v-list-item-icon>
            <v-list-item-title>User Administration</v-list-item-title>
          </v-list-item>
          <v-list-item @click="logout" v-if="this.userData">
            <v-list-item-icon>
              <v-icon>{{ icon.mdiLogout }}</v-icon>
            </v-list-item-icon>
            <v-list-item-title>Logout</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </v-app-bar>
    <v-main>
      <MsgBox/>
      <transition name="component-fade" mode="out-in">
        <router-view/>
      </transition>
    </v-main>
  </v-app>
</template>

<script>
import {
  mdiAccount, mdiLogin, mdiLogout, mdiAccountSupervisor, mdiAccountCircle,
  mdiCog, mdiToolbox,
} from '@mdi/js';
import QueryInput from './components/QueryInput.vue';
import MsgBox from './components/notification/MsgBox.vue';
import * as Constants from './constants';
import '@fontsource/roboto';
import {
  doGet, getLocalUser, resetLocal, getLocalSettings, isAdminUser, isPrivilegedUser,
} from './utilities';
import { ROUTE_USER_LOGOUT, DEFAULT_SETTINGS } from './constants';

export default {
  name: 'App',
  components: {
    MsgBox,
    QueryInput,
  },
  data() {
    return {
      applicationName: Constants.APPLICATION_NAME,
      icon: {
        mdiAccount,
        mdiLogin,
        mdiLogout,
        mdiAccountSupervisor,
        mdiAccountCircle,
        mdiCog,
        mdiToolbox,
      },
      route: {
        userProfilePage: Constants.ROUTE_NAME_USER_PROFILE_PAGE,
        userAdminPage: Constants.ROUTE_NAME_USER_ADMIN_PAGE,
        userLoginPage: Constants.ROUTE_NAME_LOGIN_PAGE,
        heuristicsPage: Constants.ROUTE_NAME_USER_HEURISTIC_PAGE,
        rootPage: Constants.ROUTE_NAME_ENTRY_PAGE,
      },
      isUserAdminDisabled: false,
      isUserLoginDisabled: false,
      isUserProfileDisabled: false,
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
    showUserAdmin() {
      return isAdminUser(this.userData);
    },
    showTools() {
      return isPrivilegedUser(this.userData) || isAdminUser(this.userData);
    },
  },
  methods: {
    setErrorMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'error', temporary: true });
    },
    // goToPage should receive a page name from ./constants
    goToPage(pageName) {
      // only change route if not already on page
      if (this.$route.name !== pageName) this.$router.push({ name: pageName });
    },
    logout() {
      doGet(ROUTE_USER_LOGOUT, this.$router)
        .then((data) => {
          if (data.success === undefined) throw Error('error logging out');
          if (data.success === false) {
            throw Error(data.msg);
          }
          resetLocal();
          this.userData = null;
          this.settings = null;
          this.goToPage(Constants.ROUTE_NAME_LOGIN_PAGE);
        })
        .catch((error) => {
          this.setErrorMessage(error);
        });
    },
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
