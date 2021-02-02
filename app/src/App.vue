<template>
  <v-app>
    <v-app-bar app absolute>
      <v-img
          @click="goToRoot()" style="cursor:pointer"
          alt="Dakar Logo"
          class="shrink mr-2"
          contain
          src="./assets/dakar_dash.svg"
          transition="scale-transition"
          width="32"/>
      <v-toolbar-title class="mx-2 d-none d-sm-flex" @click="goToRoot()" style="cursor:pointer">
        {{ applicationName }}
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <QueryInput class="mx-4"/>
      <v-spacer></v-spacer>
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
          <v-list-item @click="goToSettings" v-if="this.userData" :disabled="isUserProfileDisabled">
            <v-list-item-icon>
              <v-icon>{{ icon.mdiCog }}</v-icon>
            </v-list-item-icon>
            <v-list-item-title>Settings</v-list-item-title>
          </v-list-item>
          <v-list-item @click="goToLogin" v-if="!this.userData" :disabled="isUserLoginDisabled">
            <v-list-item-icon>
              <v-icon>{{ icon.mdiLogin }}</v-icon>
            </v-list-item-icon>
            <v-list-item-title>Login</v-list-item-title>
          </v-list-item>
          <v-list-item @click="goToUserAdministration" :disabled="isUserAdminDisabled"
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
  mdiCog,
} from '@mdi/js';
import QueryInput from './components/QueryInput.vue';
import MsgBox from './components/MsgBox.vue';
import * as Constants from './constants';
import '@fontsource/roboto';
import {
  doGet, getLocalUser, resetLocal, getLocalSettings,
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
    errMsg: {
      get() {
        return this.$store.getters.getErrorMsg;
      },
      set(value) {
        this.$store.dispatch('setErrorMsg', value);
      },
    },
    showUserAdmin() {
      return this.userData && this.userData.roles && this.userData.roles.some((d) => d.role_name === 'admin');
    },
  },
  methods: {
    goToRoot() {
      this.goToPage(Constants.ROUTE_NAME_ENTRY_PAGE);
    },
    goToLogin() {
      this.goToPage(Constants.ROUTE_NAME_LOGIN_PAGE);
    },
    goToSettings() {
      this.goToPage(Constants.ROUTE_NAME_USER_PROFILE_PAGE);
    },
    goToUserAdministration() {
      this.goToPage(Constants.ROUTE_NAME_USER_ADMIN_PAGE);
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
          this.goToLogin();
        })
        .catch((error) => {
          this.errMsg = error;
        });
    },
    checkRoute(routeName) {
      this.isUserLoginDisabled = false;
      this.isUserAdminDisabled = false;
      this.isUserProfileDisabled = false;

      switch (routeName) {
        case Constants.ROUTE_NAME_LOGIN_PAGE:
          this.isUserLoginDisabled = true;
          break;
        case Constants.ROUTE_NAME_USER_ADMIN_PAGE:
          this.isUserAdminDisabled = true;
          break;
        case Constants.ROUTE_NAME_USER_PROFILE_PAGE:
          this.isUserProfileDisabled = true;
          break;
        default:
            // nothing
      }
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

    this.checkRoute(this.$router.currentRoute.name);

    this.loadStorageData();
  },
  watch: {
    $route(to) {
      this.checkRoute(to.name);
    },
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
